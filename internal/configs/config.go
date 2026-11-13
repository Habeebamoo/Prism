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
	Env                string
	MaxUploadSize      int64
	RedisUrl           string
	PostgresUrl        string
	StreamName         string
	WorkerGroupName    string
	SemaphoreSize      int64
	B2Endpoint         string
	B2AccountId        string
	B2AccessKeyId      string
	B2SecretAccessKey  string
	B2BucketName       string
	B2Region           string
	JWTSecret          string
}

var Envs []string

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
	 	log.Println("No .env file found. Using system env")
	}

	port               :=  "PORT"
	env                :=  "ENV"
	maxUploadSizeStr   :=  "MAX_UPLOAD_SIZE_BYTES"
	redisUrl           :=  "REDIS_URL"
	postresUrl         :=  "POSTGRES_URL"
	streamName         :=  "STREAM_NAME"
	workerGroupName    :=  "WORKER_GROUP_NAME"
	semaphoreSizeStr   :=  "SEMAPHORE_SIZE"
	b2Endpoint         :=  "B2_ENDPOINT"
	b2AccessKeyId      :=  "B2_ACCESS_KEY_ID"
	b2SecretAccessKey  :=  "B2_SECRET_ACCESS_KEY"
	b2BucketName       :=  "B2_BUCKET_NAME"
	b2Region           :=  "B2_REGION"
	jwtSecret          :=  "JWT_SECRET"

	// validate env variables
	Envs =  append(Envs,
		env, 
		maxUploadSizeStr, 
		redisUrl, 
		postresUrl,
		streamName, 
		workerGroupName, 
		port,
		semaphoreSizeStr,
		b2Endpoint,
		b2AccessKeyId,
		b2SecretAccessKey,
		b2BucketName,
		b2Region,
		jwtSecret,
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
		Env:                os.Getenv(env),
		MaxUploadSize:      maxUploadSize,
		RedisUrl:           os.Getenv(redisUrl),
		PostgresUrl:        os.Getenv(postresUrl),
		StreamName:         os.Getenv(streamName),
		WorkerGroupName:    os.Getenv(workerGroupName),
		SemaphoreSize:      semaphorSize,
		B2Endpoint:         os.Getenv(b2Endpoint),
		B2AccessKeyId:      os.Getenv(b2AccessKeyId),
		B2SecretAccessKey:  os.Getenv(b2SecretAccessKey),
		B2BucketName:       os.Getenv(b2BucketName),
		B2Region:           os.Getenv(b2Region),
		JWTSecret:          os.Getenv(jwtSecret),
	}
}

func Check(envName string) error {
	env := os.Getenv(envName)
	if env == "" {
		return fmt.Errorf("%v is missing", envName)
	}

	return nil
}