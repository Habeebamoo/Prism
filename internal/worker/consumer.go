package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/Habeebamoo/Prism/internal/storage"
	"github.com/Habeebamoo/Prism/internal/transcode"
	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	cfg         *configs.Config
	client      *redis.Client
	sem         chan struct{}
	wg          sync.WaitGroup
	transcoder  *transcode.Transcoder
	r2Storage   *storage.R2Storage
}

func NewConsumer(cfg *configs.Config, client *redis.Client, transcoder *transcode.Transcoder, r2Storage *storage.R2Storage) *Consumer {
	return &Consumer{
		cfg: cfg, 
		client: client, 
		sem: make(chan struct{}, cfg.SemaphoreSize),
		transcoder: transcoder,
		r2Storage: r2Storage,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	// create workers
	err := c.client.XGroupCreateMkStream(ctx, c.cfg.StreamName, c.cfg.WorkerGroupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("[WorkerPool] Failed to create consumer group: %v\n", err)
	}
	log.Println("[WorkerPool] Workers listening on stream...")

	for {
		select {
		case <-ctx.Done():
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
					c.wg.Add(1)

					go func(m redis.XMessage) {
						defer func() { <-c.sem } () // release slot when done
						defer c.wg.Done()

						// extract job
						raw := m.Values["data"].(string)
						var payload transcode.TranscodeJob

						if json.Unmarshal([]byte(raw), &payload) != nil {
							log.Println("[WorkerPool] Failed to extract job")
							return
						}

						log.Printf("[WorkerPool] Picked up video: (%s)\n", payload.VideoId)

						// transcode job
						transcodeCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
						defer cancel()

						log.Printf("[WorkerPool] Starting transcode for video: (%s)\n", payload.VideoId)

						err := c.transcoder.Transcode(transcodeCtx, payload)
						if err != nil {
							log.Println(err.Error())
							return
						}

						log.Printf("[WorkerPool] Job Transcoded Successfully: (%s)\n", payload.VideoId)

						// upload transcoded files to r2 bucket
						// remotePrefix := fmt.Sprintf("videos/%s", payload.VideoId)
						// log.Println("[WorkerPool] Uploading assets to Cloudflare R2")

						// if err := c.r2Storage.UploadDirectory(ctx, payload.RawPath, remotePrefix); err != nil {
						// 	log.Printf("[WorkerPool] R2 upload failed for video: %s | err: %v\n", payload.VideoId, err)
						// 	return
						// }

						// // clean up local disk
						// os.RemoveAll(payload.OutputDir)
						// os.RemoveAll(payload.RawPath)

						// ack job
						ackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						if c.Ack(ackCtx, m.ID) != nil {
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

func (c *Consumer) Stop() {
	log.Println("[WorkerPool] Waiting for active transcoders to finish")
	c.wg.Wait()
}