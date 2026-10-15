package storage

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Storage struct {
	client *s3.Client
	bucket string
}

func NewR2Storage(ctx context.Context, localCfg *configs.Config) (*R2Storage, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(localCfg.R2AccessKeyId, localCfg.R2SecretAccessKey, ""),
		),
		config.WithRegion("auto"),
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to load AWS config: %w", err)
	}

	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", localCfg.R2AccountId)

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2Endpoint)
	})

	return &R2Storage{ client: client, bucket: localCfg.R2BucketName  }, nil
}

func (r *R2Storage) UploadDirectory(ctx context.Context, localDir, remotePrefix string) error {
	return filepath.Walk(localDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w\n", err)
		}

		key := filepath.ToSlash(filepath.Join(remotePrefix, relPath))

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w\n", path, err)
		}
		defer file.Close()

		_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(r.bucket),
			Key: aws.String(key),
			Body: file,
			ContentType: aws.String(getContentType(path)),
		})
		if err != nil {
			return fmt.Errorf("failed to upload %s to R2: %w\n", key, err)
		}

		return nil
	})
}

func getContentType(path string) string {
	ext :=  strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".m3u8":
		return "application/x-mpegURL"
	case ".ts":
		return "video/MP2T"
	default:
		if t := mime.TypeByExtension(ext); t != "" {
			return t
		}
		return "application/octet-stream"
	}
}