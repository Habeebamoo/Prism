package ingest

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
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

func StreamFileToDisk(file multipart.File) error {
	videoId := uuid.New().String()
	rawFilePath := filepath.Join("./storage/raw", fmt.Sprintf("%s.mp4", videoId))

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