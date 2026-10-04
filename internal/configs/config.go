package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	MaxUploadSize    int64
	RedisUrl         string
	StreamName       string
	WorkerGroupName  string
	SemaphoreSize    int64
}

var Envs []string

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	port := "PORT"
	maxUploadSizeStr := "MAX_UPLOAD_SIZE_BYTES"
	redisUrl := "REDIS_URL"
	streamName := "STREAM_NAME"
	workerGroupName := "WORKER_GROUP_NAME"
	semaphoreSizeStr := "SEMAPHORE_SIZE"

	// validate env variables
	Envs =  append(Envs, 
		maxUploadSizeStr, 
		redisUrl, 
		streamName, 
		workerGroupName, 
		port,
		semaphoreSizeStr,
	)
	for _, env := range Envs {
		if err := Check(env); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	// conver string env -> int
	maxUploadSize, _ := strconv.ParseInt(os.Getenv(maxUploadSizeStr), 10, 64)
	semaphorSize, _ := strconv.ParseInt(os.Getenv(semaphoreSizeStr), 10, 64)

	return &Config{
		Port: os.Getenv(port),
		MaxUploadSize: maxUploadSize,
		RedisUrl: os.Getenv(redisUrl),
		StreamName: os.Getenv(streamName),
		WorkerGroupName: os.Getenv(workerGroupName),
		SemaphoreSize: semaphorSize,
	}
}

func Check(envName string) error {
	env := os.Getenv(envName)
	if env == "" {
		return fmt.Errorf(fmt.Sprintf("%v is missing", envName))
	}

	return nil
}