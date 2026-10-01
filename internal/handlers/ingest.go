package handlers

import (
	"errors"
	"fmt"
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

	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)

	file, fileHeader, err := r.FormFile("video")
	if err != nil {
		var errMaxBytes *http.MaxBytesError

		if errors.Is(err, http.ErrMissingFile) {
			utils.JsonResponse(w, 400, utils.Payload{ 
				Success: false, 
				Message: "File missing, 'video' is required",
			})
		} else if errors.As(err, &errMaxBytes) {
			utils.JsonResponse(w, http.StatusRequestEntityTooLarge, utils.Payload{
				Success: false,
				Message: "File is too large. Maximum 5MB allowed.",
			})
		}

		return
	}

	defer file.Close()

	fmt.Printf("Received file: %s\n", fileHeader.Header)

	utils.JsonResponse(w, 200, utils.Payload{
		Success: true,
		Message: "File received successfully",
	})
}