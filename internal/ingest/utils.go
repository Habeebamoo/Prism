package ingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type ErrorResponse struct {
	Message string
	Code    int    
}

func ErrorType(err error) (ErrorResponse, bool) {
	if err != nil {
		var errMaxBytes *http.MaxBytesError

		if errors.Is(err, http.ErrMissingFile) {
			return ErrorResponse{
				Message: "File is missing, 'video' field is required.",
				Code:  http.StatusRequestEntityTooLarge,
			}, false

		} else if errors.As(err, &errMaxBytes) {
			return ErrorResponse{
				Message: "File is too large. Maximum 5MB allowed.",
				Code:    http.StatusRequestEntityTooLarge,
			}, false

		}
	}

	return ErrorResponse{}, true
}

func ValidateMimeType(file multipart.File) error {
	// sniff first 512 bytes
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		return fmt.Errorf("Unable to read file headers")
	}

	// reset file offset
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("Failed to reset file stream")
	}

	// validate content type
	contentType := http.DetectContentType(buf)
	if contentType != "video/mp4" && contentType != "video/quicktime" && contentType != "video/webm" {
		return fmt.Errorf("Invalid file type. Only MP4, MOV, and WebM are allowed.")
	}

	return nil
}

func StreamFileToDisk(file multipart.File, rawFilePath string) error {
	// create storage directory
	if err := os.MkdirAll(filepath.Dir(rawFilePath), 0755); err != nil {
		return fmt.Errorf("Failed to initialize storage directory: %v", err)
	}

	// create destination file
	dst, err := os.Create(rawFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create destination file: %v", err)
	}
	defer dst.Close()

	// stream file bytes to disk
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(rawFilePath)
		return fmt.Errorf("Failed to write file to disk: %v", err)
	}

	return nil
}

func ProbeVideo(filePath string) (*VideoMetadata, error) {
	cmd := exec.Command("ffprobe",
				"-v", "quiet",
				"-print_format", "json",
				"-show_format",
				"-show_streams",
				"-select_streams", "v:0",
				filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe execution failed: %v", err)
	}

	var probe FFProbeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return  nil, fmt.Errorf("Failed to parse ffprobe output: %v", err)
	}


	if len(probe.Streams) == 0 {
		return nil, fmt.Errorf("No video streams found in the file")
	}

	stream := probe.Streams[0]

	if stream.Width <= 0 || stream.Height <= 0 {
		return nil, fmt.Errorf("Invalid video dimensions")
	}

	duration, _ := strconv.ParseFloat(probe.Format.Duration, 64)
	
	return &VideoMetadata{
		Width: stream.Width,
		Height: stream.Height,
		Duration: duration,
		Codec: stream.CodecName,
	}, nil
}