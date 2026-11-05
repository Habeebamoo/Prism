package stream

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/Habeebamoo/Prism/internal/storage"
	"github.com/Habeebamoo/Prism/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type StreamHandler struct {
	cfg    *configs.Config
	client *storage.B2Client
}

func NewStreamHandler(cfg *configs.Config, client *storage.B2Client) *StreamHandler {
	return &StreamHandler{ cfg, client }
}

func (s *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/v1/stream/")
	parts := strings.SplitN(trimmed, "/", 3)

	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		utils.JsonResponse(w, 400, utils.Payload{ Success: false, Message: "Invalid URL Format" })
		return
	}

	videoId := parts[0]
	relativePath := filepath.Clean(filepath.Join(parts[1:]...))

	if strings.HasPrefix(relativePath, "..") || strings.HasPrefix(relativePath, "/") {
		utils.JsonResponse(w, 403, utils.Payload{ Success: false, Message: "Forbidden path traversal attempt" })
		return
	}

	objectKey := fmt.Sprintf("videos/%s/%s", videoId, relativePath)

	output, err := s.client.GetObject(r.Context(), objectKey)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			utils.JsonResponse(w, 404, utils.Payload{ Success: false, Message: "Segment or Playlist not found" })
			return
		}

		utils.JsonResponse(w, 500, utils.Payload{ Success: false, Message: "Failed to stream segment" })
		return
	}
	defer output.Body.Close()

	w.Header().Set("Content-Type", utils.GetContentType(relativePath))
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Cache headers
	if strings.HasSuffix(relativePath, ".m3u8") {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}

	if output.ContentLength != nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", *output.ContentLength))
	}

	_, _ = io.Copy(w, output.Body)
}