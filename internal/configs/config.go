package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	MaxUploadSize  int64
	RedisUrl       string
	StreamName     string
}

var Envs []string

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	maxUploadSizeStr := "MAX_UPLOAD_SIZE_BYTES"
	redisUrl := "REDIS_URL"
	streamName := "STREAM_NAME"

	// validate env variables
	Envs =  append(Envs, maxUploadSizeStr, redisUrl, streamName)
	for _, env := range Envs {
		if err := Check(env); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	maxUploadSize, _ := strconv.ParseInt(os.Getenv(maxUploadSizeStr), 10, 64)

	return &Config{
		MaxUploadSize: maxUploadSize,
		RedisUrl: os.Getenv(redisUrl),
		StreamName: os.Getenv(streamName),
	}
}

func Check(envName string) error {
	env := os.Getenv(envName)
	if env == "" {
		return fmt.Errorf(fmt.Sprintf("%v is missing", envName))
	}

	return nil
}