package queue

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/Habeebamoo/Prism/internal/transcode"
	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	cfg         *configs.Config
	client      *redis.Client
	sem         chan struct{}
	transcoder  *transcode.Transcoder
}

func NewConsumer(cfg *configs.Config, client *redis.Client, transcoder *transcode.Transcoder) *Consumer {
	return &Consumer{
		cfg: cfg, 
		client: client, 
		sem: make(chan struct{}, cfg.SemaphoreSize),
		transcoder: transcoder,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	// create workers
	err := c.client.XGroupCreateMkStream(ctx, c.cfg.StreamName, c.cfg.WorkerGroupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("[WorkerPool] Failed to create consumer group: %v\n", err)
	}
	log.Println("[WorkerPool] Workers listening on stream...")

	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			log.Println("[WorkerPool] Waiting for active transcoders to finish")

			wg.Wait()

			log.Println("[WorkerPool] All workers completed")
			return
		default:
			streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group: c.cfg.WorkerGroupName,
				Consumer: "worker-node",
				Streams: []string{c.cfg.StreamName, ">"},
				Count: 10,
				Block: 5*time.Second,
			}).Result()

			// error check
			if err != nil {
				if err == redis.Nil || ctx.Err() != nil {
					continue
				}

				log.Printf("[WorkerPool] stream read error: %v\n", err)
				time.Sleep(1*time.Second)
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					c.sem <- struct{}{} // occupy slot
					wg.Add(1)

					go func(m redis.XMessage) {
						defer func() { <-c.sem } () // release slot when done
						defer wg.Done()

						log.Printf("[WorkerPool] Picked up message %v\n", m.ID)

						// extract job
						raw := m.Values["data"].(string)
						var payload transcode.TranscodeJob

						if json.Unmarshal([]byte(raw), &payload) != nil {
							log.Println("[WorkerPool] Failed to extract job")
							return
						}

						// transcode job
						transcodeCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
						defer cancel()

						err := c.transcoder.Transcode(transcodeCtx, payload)
						if err != nil {
							log.Println("[WorkerPool] Failed to transcode job")
							return
						}

						log.Println("[WorkerPool] Job Transcoded Successfully")

						if c.Ack(ctx, m.ID) != nil {
							log.Println("[WorkerPool] Failed to ack job")
							return
						}
					}(msg)
				}
			}
		}
	}
}

func (c *Consumer) Ack(ctx context.Context, id string) error {
	return c.client.XAck(ctx, c.cfg.StreamName, c.cfg.WorkerGroupName, id).Err()
}