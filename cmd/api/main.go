package main

import (
	"context"
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
	"github.com/Habeebamoo/Prism/pkg/utils"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := configs.Load()

	// init redis
	rdb, err := queue.NewRedisClient(ctx, cfg.RedisUrl)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// init postgres
	postgresDB := database.Init(cfg)

	producer := queue.NewProducer(cfg, rdb)

	// init bc client
	storageClient, err := storage.NewB2Client(cfg, ctx)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// init handlers
	authHandler := handlers.NewAuthHandler(cfg, postgresDB)
	videoHandler := handlers.NewVideoHandler(cfg, postgresDB, storageClient)
	ingestHandler := ingest.NewIngestHandler(cfg, producer, videoHandler)
	streamHandler := stream.NewStreamHandler(cfg, storageClient)
	streamMngr := ingest.NewStreamManager("./tmp/live", videoHandler, ingest.OnLiveStreamComplete(storageClient))

	// routes
	mux := http.NewServeMux()
	authMiddleware := middlewares.RequireAuth(cfg)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		utils.JsonResponse(w, 200, utils.Payload{ Success: true, Message: "System is live!!!" })
	})

	mux.HandleFunc("POST /api/v1/auth/request-otp", authHandler.HandleRequestOTP)
	mux.HandleFunc("POST /api/v1/auth/verify-otp", authHandler.HandleVerifyOTP)
	mux.Handle("POST /api/v1/auth/logout", authMiddleware(http.HandlerFunc(authHandler.HandleLogout)))
	mux.Handle("GET /api/v1/auth/me", authMiddleware(http.HandlerFunc(authHandler.HandleGetMe)))

	mux.Handle("GET /api/v1/me/videos", authMiddleware(http.HandlerFunc(videoHandler.HandleListMyVideos)))
	mux.Handle("DELETE /api/v1/videos/{id}", authMiddleware(http.HandlerFunc(videoHandler.HandleDeleteVideo)))

	mux.Handle("POST /api/v1/ingest", authMiddleware(http.HandlerFunc(ingestHandler.Ingest)))
	//mux.HandleFunc("POST /api/v1/ingest", ingestHandler.Ingest)
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