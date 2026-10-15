package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/Habeebamoo/Prism/internal/queue"
	"github.com/Habeebamoo/Prism/internal/storage"
	"github.com/Habeebamoo/Prism/internal/transcode"
	"github.com/Habeebamoo/Prism/internal/worker"
)

func main() {
	// init env configs
	cfg := configs.Load()

	// init redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb, err := queue.NewRedisClient(ctx, cfg.RedisUrl)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// init storage
	r2Storage, err := storage.NewR2Storage(ctx, cfg)
	if err != nil {
		log.Println("[Worker] Failed to initialize storage bucket")
		os.Exit(1)
	}

	// init transcoder workerpool
	transcoder := transcode.NewTranscoder()
	consumer := worker.NewConsumer(cfg, rdb, transcoder, r2Storage)

	go consumer.Start(sigCtx)

	<-sigCtx.Done()
	log.Println("[Worker] Shutting down process.")

	consumer.Stop()
	log.Println("[Worker] Shutdown complete.")
}