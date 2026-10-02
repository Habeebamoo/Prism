package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/redis/go-redis/v9"
)

type Producer struct {
	cfg    *configs.Config
	client *redis.Client
}

type TranscodeJob struct {
	VideoId    string  `json:"video_id"`
	RawPath    string  `json:"raw_path"`
	OutputDir  string  `json:"output_dir"`
}

func NewProducer(cfg *configs.Config, client *redis.Client) *Producer {
	return &Producer{cfg, client}
}

func (p *Producer) Publish(job TranscodeJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("Failed to convert payload to JSON")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: p.cfg.StreamName,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]interface{}{
			"data": string(payload),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("redis XAdd failed on stream '%s': %w", p.cfg.StreamName, err)
	}

	return nil
}


