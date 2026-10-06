package ingest

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/Habeebamoo/Prism/internal/queue"
	"github.com/Habeebamoo/Prism/internal/transcode"
	"github.com/Habeebamoo/Prism/pkg/utils"
	"github.com/google/uuid"
)

type IngestHandler struct {
	ctx       context.Context
	cfg       *configs.Config
  producer  *queue.Producer
}

func NewIngestHandler(ctx context.Context, cfg *configs.Config, producer *queue.Producer) *IngestHandler {
	return &IngestHandler{ctx, cfg, producer}
}

func (i *IngestHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JsonResponse(w, http.StatusMethodNotAllowed, utils.Payload{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// limit file size
	r.Body = http.MaxBytesReader(w, r.Body, i.cfg.MaxUploadSize)
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
		return
	}

	fmt.Println("HIT")

	// create output directory
	outputDir := fmt.Sprintf("storage/processed/%s.mov", videoId)
	err = os.MkdirAll(filepath.Dir(fmt.Sprintf("%s", outputDir)), 0755)
	if err != nil {
		log.Println(err)
	}

	// publish job to queue
	payload := transcode.TranscodeJob{ VideoId: videoId, RawPath: rawFilePath, OutputDir: outputDir }

	err = i.producer.Publish(payload)
	if err != nil {
		log.Println(err.Error())

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