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
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// init redis
	cfg := configs.Load()
	rdb, err := configs.NewRedisClient(ctx, cfg.RedisUrl)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// init message broker
	producer := queue.NewProducer(cfg, rdb)

	// init handler
	ingestHandler := ingest.NewIngestHandler(cfg, producer)

	// init file server
	processedDir := "./storage/processed"
	fileServer := http.FileServer(http.Dir(processedDir))

	mux := http.NewServeMux()

	// routes
	mux.HandleFunc("POST /api/v1/ingest", ingestHandler.Ingest)
	mux.Handle("GET /api/v1/stream", http.StripPrefix("/api/v1/stream/", fileServer))

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