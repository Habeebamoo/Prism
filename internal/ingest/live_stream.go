package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/Habeebamoo/Prism/pkg/utils"
)

type StreamSession struct {
	ID            string
	OutputDir     string
	PipeWriter    *io.PipeWriter
	Cmd           *exec.Cmd
	StderrBuf     *bytes.Buffer
	ExpectedSeq   int
	LastActivity  time.Time
	mu            sync.Mutex
	doneCh        chan error
}

type StreamManager struct {
	baseTempDir    string
	sessions       map[string]*StreamSession
	mu             sync.RWMutex
	onComplete     func (videoId string, hlsDir string)
}

func NewStreamManager(baseTempDir string, onComplete func(videoId string, hlsDir string)) *StreamManager {
	sm := &StreamManager{
		baseTempDir:  baseTempDir,
		sessions:     make(map[string]*StreamSession),
		onComplete:   onComplete,
	}

	go sm.reapStaleSessions()
	return sm
}

func (sm *StreamManager) getOrCreateSession(videoId string, seq int) (*StreamSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[videoId]; exists {
		return session, nil
	}

	if seq != 0 {
		return nil, fmt.Errorf("Session %s not found and initial chunk seq=%d != 0", videoId, seq)
	}

	outDir := filepath.Join(sm.baseTempDir, videoId)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("Failed to create output dir: %w", err)
	}

	pr, pw := io.Pipe()
	var stderr bytes.Buffer

	// Transcode live WebM (VP8/VP9 + Opus) from stdin into 2s HLS segments
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-fflags", "+genpts",
		"-i", "pipe:0",
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-g", "30",
		"-sc_threshold", "0",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "44100",
		"-f", "hls",
		"-hls_time", "2",
		"-hls_list_size", "0",
		"-hls_playlist_type", "event",
		"-hls_segment_filename", filepath.Join(outDir, "seg_%03d.ts"),
		filepath.Join(outDir, "master.m3u8"),
	)
	cmd.Stdin = pr
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		_ = pw.Close()
		_ = pr.Close()
		return nil, fmt.Errorf("start ffmpeg: %w", err)
	}

	session := &StreamSession{
		ID:            videoId,
		OutputDir:     outDir,
		PipeWriter:    pw,
		Cmd:           cmd,
		StderrBuf:     &stderr,
		ExpectedSeq:   0,
		LastActivity:  time.Now(),
		doneCh:        make(chan error, 1),
	}

	go func() {
		session.doneCh <- cmd.Wait()
	}()

	sm.sessions[videoId] = session
	return session, nil
}

func (sm *StreamManager) HandleStreamChunk(w http.ResponseWriter, r *http.Request) {
	videoId := r.PathValue("id")
	if videoId == "" {
		utils.JsonResponse(w, 400, utils.Payload{ Success: false, Message: "Missing Video Id" })
		return
	}

	seq, err := strconv.Atoi(r.URL.Query().Get("seq"))
	if err != nil || seq < 0 {
		utils.JsonResponse(w, 400, utils.Payload{ Success: false, Message: "Invalid seq query parameter" })
		return
	}

	session, err := sm.getOrCreateSession(videoId, seq)
	if err != nil {
		utils.JsonResponse(w, 400, utils.Payload{ Success: false, Message: err.Error()} )
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if seq != session.ExpectedSeq {
		utils.JsonResponse(w, 409, utils.Payload{ 
			Success: false, 
			Message: fmt.Sprintf("out-of-order chunk: expected %d, got %d", session.ExpectedSeq, seq), 
		})
		return
	}

	if _, err := io.Copy(session.PipeWriter, r.Body); err != nil {
		utils.JsonResponse(w, 500, utils.Payload{ Success: false, Message: "Pipe write error" })
		return
	}

	session.ExpectedSeq++
	session.LastActivity = time.Now()
	w.WriteHeader(http.StatusAccepted)
}

func (sm *StreamManager) HandleStreamFinish(w http.ResponseWriter, r *http.Request) {
	videoID := r.PathValue("id")

	sm.mu.Lock()
	session, exists := sm.sessions[videoID]
	if exists {
		delete(sm.sessions, videoID)
	}
	sm.mu.Unlock()

	if !exists {
		utils.JsonResponse(w, 404, utils.Payload{ Success: false, Message: "Active stream session not found" })
		return
	}

	// Send EOF to FFmpeg stdin so it flushes final frames and appends #EXT-X-ENDLIST
	session.mu.Lock()
	_ = session.PipeWriter.Close()
	session.mu.Unlock()

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	select {
	case err := <-session.doneCh:
		if err != nil {
			log.Printf("ffmpeg error for %s: %v\nstderr: %s", videoID, err, session.StderrBuf.String())

			utils.JsonResponse(w, 500, utils.Payload{
				Success: false,
				Message: "Transcoding failed during finalization",
			})
			return
		}
	case <-ctx.Done():
		_ = session.Cmd.Process.Kill()
		utils.JsonResponse(w, 504, utils.Payload{ Success: false, Message: "FFmpeg finalization timed out" })
		return
	}

	// Upload generated HLS folder to Backblaze B2 asynchronously while serving locally right away
	if sm.onComplete != nil {
		go sm.onComplete(videoID, session.OutputDir)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"video_id": videoID,
		"status":   "READY",
	})
}

func (sm *StreamManager) reapStaleSessions() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var stale []*StreamSession

		sm.mu.Lock()
		for id, session := range sm.sessions {
			session.mu.Lock()
			inactive := time.Since(session.LastActivity) > 30*time.Second
			session.mu.Unlock()

			if inactive {
				stale = append(stale, session)
				delete(sm.sessions, id)
			}
		}
		sm.mu.Unlock()

		for _, s := range stale {
			_ = s.PipeWriter.Close()
			if s.Cmd.Process != nil {
				_ = s.Cmd.Process.Kill()
			}
			_ = os.RemoveAll(s.OutputDir)
		}
	}
}