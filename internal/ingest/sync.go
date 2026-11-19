package ingest

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Habeebamoo/Prism/internal/storage"
)

func OnLiveStreamComplete(storageClient *storage.B2Client) func(videoId, hlsDir string) {
	return func(videoId, hlsDir string) {
		remotePrefix := fmt.Sprintf("videos/%s", videoId)
		outputDir := hlsDir
		if outputDir == "" {
			outputDir = fmt.Sprintf("./tmp/live/%s", videoId)
		}

		log.Println("[LIVE STREAM] Uploading assets to B2")

		uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer uploadCancel()

		if err := storageClient.UploadDirectory(uploadCtx, outputDir, remotePrefix); err != nil {
			log.Printf("[LIVE STREAM] B2 upload failed for video: %s | err: %v\n", videoId, err)
			return
		}

		log.Printf("[LIVE STREAM] B2 upload finished for video %s\n", videoId)

		if err := os.RemoveAll(outputDir); err != nil {
			log.Printf("[LIVE STREAM] Failed to clean up workspace %s: %v\n", outputDir, err)
			return
		}
		log.Println("[LIVE STREAM] Cleaned up local workspace")
	}
}