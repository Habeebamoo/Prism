package queue

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	cfg *configs.Config
	client *redis.Client
	sem chan struct{}
}

func NewConsumer(cfg *configs.Config, client *redis.Client) *Consumer {
	return &Consumer{
		cfg: cfg, 
		client: client, 
		sem: make(chan struct{}, cfg.SemaphoreSize),
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
				Consumer: fmt.Sprintf("worker-node-%d", os.Getpid()),
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

						// process
						log.Printf("[WorkerPool] Picked up message %v\n", m.ID)
						time.Sleep(30*time.Second)
					}(msg)
				}
			}
		}
	}
}