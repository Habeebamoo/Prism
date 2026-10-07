package transcode

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Transcoder struct {}

type TranscodeJob struct {
	VideoId    string  `json:"video_id"`
	RawPath    string  `json:"raw_path"`
	OutputDir  string  `json:"output_dir"`
}

func NewTranscoder() *Transcoder {
	return &Transcoder{}
}

func (t *Transcoder) Transcode(ctx context.Context, job TranscodeJob) error {
	// create directory
	if os.MkdirAll(job.OutputDir, 0755) != nil {
		return fmt.Errorf("[WorkerPool] Failed to create directory")
	}

	masterPlaylist := filepath.Join(job.OutputDir, "master.m3u8")
	segmentPattern := filepath.Join(job.OutputDir, "segment_%03d.ts")

	cmd := exec.CommandContext(ctx, "ffmpeg",
		 	"-i", job.RawPath, 
		 	"-c:v", "libx264",
		 	"-crf", "23",
		 	"-preset", "veryfast",
		 	"-c:a", "aac",
			"-b:a", "128k",
			"-f", "hls",
			"-hls_time", "4",
			"-hls_playlist_type", "vod",
			"-hls_segment_filename", segmentPattern,
			masterPlaylist,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[WorkerPool] ffmpeg failed: %v", err.Error())
	}

	fmt.Println(string(output))
	return nil
}