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
	"github.com/Habeebamoo/Prism/internal/database"
	"github.com/Habeebamoo/Prism/internal/handlers"
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

	// init postgres
	postgresDB := database.Init(cfg)

	// init message broker
	producer := queue.NewProducer(cfg, rdb)

	// inint b2 client
	storageClient, err := storage.NewB2Client(cfg, ctx)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// init handlers
	authHandler   := handlers.NewAuthHandler(cfg, postgresDB)
	videoHandler  := handlers.NewVideoHandler(cfg, postgresDB, storageClient)
	ingestHandler := ingest.NewIngestHandler(cfg, producer, videoHandler)
	streamHandler := stream.NewStreamHandler(cfg, storageClient)

	streamMngr := ingest.NewStreamManager("./tmp/live", videoHandler, func(videoId, hlsDir string) {
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
	authMiddleware := middlewares.RequireAuth(cfg)

	mux.HandleFunc("POST /api/v1/auth/request-otp", authHandler.HandleRequestOTP)
	mux.HandleFunc("POST /api/v1/auth/verify-otp", authHandler.HandleVerifyOTP)
	mux.Handle("POST /api/v1/auth/logout", authMiddleware(http.HandlerFunc(authHandler.HandleLogout)))
	mux.Handle("GET /api/v1/auth/me", authMiddleware(http.HandlerFunc(authHandler.HandleGetMe)))

	mux.Handle("GET /api/v1/me/videos", authMiddleware(http.HandlerFunc(videoHandler.HandleListMyVideos)))
	mux.Handle("DELETE /api/v1/videos/{id}", authMiddleware(http.HandlerFunc(videoHandler.HandleDeleteVideo)))

	mux.Handle("POST /api/v1/ingest", authMiddleware(http.HandlerFunc(ingestHandler.Ingest)))
	mux.Handle("POST /api/v1/ingest/stream/{id}", authMiddleware(http.HandlerFunc(streamMngr.HandleStreamChunk)))
	mux.Handle("POST /api/v1/ingest/stream/{id}/finish", authMiddleware(http.HandlerFunc(streamMngr.HandleStreamFinish)))
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