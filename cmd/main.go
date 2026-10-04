package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var consumerWg sync.WaitGroup

	producer := queue.NewProducer(cfg, rdb)
	consumer := queue.NewConsumer(cfg, rdb)

	// start transcode worker pools
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		consumer.Start(sigCtx)
	}()

	// init handler
	ingestHandler := ingest.NewIngestHandler(ctx, cfg, producer)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ingest", ingestHandler.Ingest)

	handler := middlewares.CORS(mux)

	server := http.Server{
		Addr: ":"+cfg.Port,
		Handler: handler,
	}

	go func() {
		log.Println("Ingestion API is running on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("Server failed")
		}
	}()

	// block
	<-sigCtx.Done()
	consumerWg.Wait()

	log.Println("Shutting down server")

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutDownCtx); err != nil {
		log.Println("Failed to shut down server")
	}

	log.Println("Server exited cleanly")
}