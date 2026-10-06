package transcode

import (
	"fmt"
	"os/exec"
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

func (t *Transcoder) Transcode(job TranscodeJob) error {
	cmd := exec.Command("ffmpeg", "-i", job.RawPath, job.OutputDir)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}