package utils

import (
	"encoding/json"
	"net/http"
)

type Payload struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data string `json:"data,omitempty"`
}

func JsonResponse(w http.ResponseWriter, statusCode int, payload Payload) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}