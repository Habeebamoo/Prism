package utils

import (
	"errors"
	"net/http"
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