package configs

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url")
	}

	client := redis.NewClient(opts)

	// ping redis server
	if client.Ping(ctx).Err() != nil {
		return nil, fmt.Errorf("couldn't connect to redis")
	}

	log.Println("Redis Connected")
	return client, nil
}