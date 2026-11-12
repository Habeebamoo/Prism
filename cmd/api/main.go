package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/Habeebamoo/Prism/internal/ingest"
	"github.com/Habeebamoo/Prism/internal/middlewares"
	"github.com/Habeebamoo/Prism/internal/queue"
	"github.com/Habeebamoo/Prism/internal/storage"
	"github.com/Habeebamoo/Prism/internal/stream"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// init redis
	cfg := configs.Load()
	rdb, err := queue.NewRedisClient(ctx, cfg.RedisUrl)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// init message broker
	producer := queue.NewProducer(cfg, rdb)

	// inint b2 client
	storageClient, err := storage.NewB2Client(cfg, ctx)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// init handlers
	ingestHandler := ingest.NewIngestHandler(cfg, producer)
	streamHandler := stream.NewStreamHandler(cfg, storageClient)

	streamMngr := ingest.NewStreamManager("./tmp/live", func(videoId, hlsDir string) {
		remotePrefix := fmt.Sprintf("videos/%s", videoId)
		outputDir := fmt.Sprintf("./tmp/live/%s", videoId)

		log.Println("[LIVE STREAM] Uploading assets to B2")

		uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer uploadCancel()

		if err := storageClient.UploadDirectory(uploadCtx, outputDir, remotePrefix); err != nil {
			log.Printf("[LIVE STREAM] B2 upload failed for video: %s | err: %v\n", videoId, err)
			return
		}

		log.Printf("[LIVE STREAM] B2 upload finished for video %s\n", videoId)

		// clean up local disk
		os.RemoveAll(outputDir)
		log.Println("[LIVE STREAM] Cleaned up local workspace")
	})

	// routes
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/ingest", ingestHandler.Ingest)
	mux.HandleFunc("POST /api/v1/ingest/stream/{id}", streamMngr.HandleStreamChunk)
	mux.HandleFunc("POST /api/v1/ingest/stream/{id}/finish", streamMngr.HandleStreamFinish)
	mux.HandleFunc("GET /api/v1/stream/{videoId}/{filepath...}", streamHandler.ServeHTTP)

	handler := middlewares.CORS(mux)

	server := http.Server{
		Addr: ":"+cfg.Port,
		Handler: handler,
	}

	go func() {
		log.Printf("Ingest API running on port: %s\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("Server failed")
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server")

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutDownCtx); err != nil {
		log.Println("Failed to shutdown server")
		return
	}

	log.Println("Server exited cleanly")
}