package ingest

import (
	"net/http"

	"github.com/Habeebamoo/Prism/pkg/utils"
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

	// stream file bytes to disk
	if err := StreamFileToDisk(file); err != nil {
		utils.JsonResponse(w, 500, utils.Payload{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	utils.JsonResponse(w, 200, utils.Payload{
		Success: true,
		Message: "File received successfully",
	})
}