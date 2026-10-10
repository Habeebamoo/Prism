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

	hasAudio := HasAudioTrack(ctx, job.RawPath)

	segmentPattern := filepath.Join(job.OutputDir, "%v", "segment_%03d.ts")
	variantPlaylistPattern := filepath.Join(job.OutputDir, "%v", "index.m3u8")

	args := []string{
		 	"-i", job.RawPath, 

			"-filter_complex",
			"[0:v]split=3[v1][v2][v3]; "+
										"[v1]scale=w=1920:h=1080:force_original_aspect_ratio=decrease[v1out]; "+
										"[v2]scale=w=1280:h=720:force_original_aspect_ratio=decrease[v2out]; "+
										"[v3]scale=w=854:h=480:force_original_aspect_ratio=decrease[v3out]",

			//  ---- 1080p
			"-map", "[v1out]",
			"-c:v:0", "libx264",
			"-b:v:0", "5000k",
			"-maxrate:v:0", "5250k",
			"-bufsize:v:0", "7500k",

			// ---- 720p
			"-map", "[v2out]",
			"-c:v:1", "libx264",
			"-b:v:1", "2800k",
			"-maxrate:v:1", "2996k",
			"-bufsize:v:1", "4200k",

			// ---- 480p
			"-map", "[v3out]",
			"-c:v:2", "libx264",
			"-b:v:2", "1400k",
			"-maxrate:v:2", "1498k",
			"-bufsize:v:2", "2100k",
	}

	var streamMap string
	if hasAudio {
		 args = append(args, 
						"-map", "a:0?", "-c:a:0", "aac", "-b:a:0", "192k",
						"-map", "a:0?", "-c:a:1", "aac", "-b:a:1", "128k",
						"-map", "a:0?", "-c:a:2", "aac", "-b:a:2", "96k",
		)

		streamMap = "v:0,a:0,name:1080p v:1,a:1,name:720p v:2,a:2,name:480p"
	} else {
		streamMap = "v:0,name:1080p v:1,name:720p v:2,name:480p"
	}

	// HLS flags
	args = append(args, 
				"-preset", "veryfast",
				"-g", "120",
				"-sc_threshold", "0",
				"-f", "hls",
				"-hls_time", "4",
				"-hls_playlist_type", "vod",
				"-hls_flags", "independent_segments",
				"-hls_segment_filename", segmentPattern,
				"-master_pl_name", "master.m3u8",
				"-var_stream_map", streamMap,
				variantPlaylistPattern,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("[WorkerPool] ffmpeg failed: %v", string(output))
	}

	return nil
}