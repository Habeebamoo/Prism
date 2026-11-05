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


	// routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/ingest", ingestHandler.Ingest)
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