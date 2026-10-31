package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	MaxUploadSize      int64
	RedisUrl           string
	StreamName         string
	WorkerGroupName    string
	SemaphoreSize      int64
	B2Endpoint         string
	B2AccountId        string
	B2AccessKeyId      string
	B2SecretAccessKey  string
	B2BucketName       string
	B2Region           string
}

var Envs []string

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
	 	log.Println("No .env file found. Using system env")
	}

	port               :=  "PORT"
	maxUploadSizeStr   :=  "MAX_UPLOAD_SIZE_BYTES"
	redisUrl           :=  "REDIS_URL"
	streamName         :=  "STREAM_NAME"
	workerGroupName    :=  "WORKER_GROUP_NAME"
	semaphoreSizeStr   :=  "SEMAPHORE_SIZE"
	b2Endpoint         :=  "B2_ENDPOINT"
	b2AccessKeyId      :=  "B2_ACCESS_KEY_ID"
	b2SecretAccessKey  :=  "B2_SECRET_ACCESS_KEY"
	b2BucketName       :=  "B2_BUCKET_NAME"
	b2Region           :=  "B2_REGION"

	// validate env variables
	Envs =  append(Envs, 
		maxUploadSizeStr, 
		redisUrl, 
		streamName, 
		workerGroupName, 
		port,
		semaphoreSizeStr,
		b2Endpoint,
		b2AccessKeyId,
		b2SecretAccessKey,
		b2BucketName,
		b2Region,
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
		Port:               os.Getenv(port),
		MaxUploadSize:      maxUploadSize,
		RedisUrl:           os.Getenv(redisUrl),
		StreamName:         os.Getenv(streamName),
		WorkerGroupName:    os.Getenv(workerGroupName),
		SemaphoreSize:      semaphorSize,
		B2Endpoint:         os.Getenv(b2Endpoint),
		B2AccessKeyId:      os.Getenv(b2AccessKeyId),
		B2SecretAccessKey:  os.Getenv(b2SecretAccessKey),
		B2BucketName:       os.Getenv(b2BucketName),
		B2Region:           os.Getenv(b2Region),
	}
}

func Check(envName string) error {
	env := os.Getenv(envName)
	if env == "" {
		return fmt.Errorf("%v is missing", envName)
	}

	return nil
}