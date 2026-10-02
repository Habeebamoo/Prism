package ingest

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Habeebamoo/Prism/pkg/utils"
	"github.com/google/uuid"
)

var MAX_UPLOAD_SIZE int64 = 5 * 1024 * 1024 // 5MB

func IngestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JsonResponse(w, http.StatusMethodNotAllowed, utils.Payload{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// limit file size
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	file, _, err := r.FormFile("video")

	// check for errors
	res, isValid := ErrorType(err) 
	if !isValid {
		utils.JsonResponse(w, res.Code, utils.Payload{
			Success: false,
			Message: res.Message,
		})
		return
	}
	defer file.Close()

	// validate mime type
	if err := ValidateMimeType(file); err != nil {
		utils.JsonResponse(w, 400, utils.Payload{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	videoId := uuid.New().String()
	rawFilePath := filepath.Join("./storage/raw", fmt.Sprintf("%s.mp4", videoId))

	// stream file bytes to disk
	if err := StreamFileToDisk(file, rawFilePath); err != nil {
		utils.JsonResponse(w, 500, utils.Payload{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// probe video
	_, err = ProbeVideo(rawFilePath)
	if err != nil {
		os.Remove(rawFilePath)

		utils.JsonResponse(w, 500, utils.Payload{
			Success: false,
			Message: err.Error(),
		})
	}

	utils.JsonResponse(w, 200, utils.Payload{
		Success: true,
		Message: "File received successfully",
	})
}