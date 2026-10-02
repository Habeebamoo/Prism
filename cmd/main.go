package main

import (
	"context"
	"log"
	"net/http"
	"os"
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

	// init handlers
	producer := queue.NewProducer(cfg, rdb)
	ingestHandler := ingest.NewIngestHandler(ctx, cfg, producer)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ingest", ingestHandler.Ingest)

	handler := middlewares.CORS(mux)

	log.Println("Ingestion API is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}