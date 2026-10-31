package storage

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type B2Client struct {
	cfg     *configs.Config
	client  *s3.Client
	bucket  string
}

func NewB2Client(cfg *configs.Config, ctx context.Context) (*B2Client, error) {
	config, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.B2AccessKeyId, cfg.B2SecretAccessKey, ""),		
		),
		config.WithRegion(cfg.B2Region),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to load B2 config: %w\n", err)
	}

	client := s3.NewFromConfig(config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.B2Endpoint)
		o.UsePathStyle = true
	})

	return &B2Client{
		cfg: cfg,
		client: client,
		bucket: cfg.B2BucketName,
	}, nil
}

func (b *B2Client) UploadDirectory(ctx context.Context, localDir, remotePrefix string) error {
	return filepath.Walk(localDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return func() error {
			relPath, err := filepath.Rel(localDir, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			key := filepath.ToSlash(filepath.Join(remotePrefix, relPath))

			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file path: %w", err)
			}
			defer file.Close()

			log.Printf("[B2Storage] Uploading %s -> %s\n", relPath, key)

			uploadCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			_, err = b.client.PutObject(uploadCtx, &s3.PutObjectInput{
				Bucket:        aws.String(b.bucket),
				Key:           aws.String(key),
				Body:          file,
				ContentLength: aws.Int64(info.Size()),
				ContentType:   aws.String(getContentType(path)),
			})
			if err != nil {
				return fmt.Errorf("failed to upload %s to B2: %w", key, err)
			}

			return nil
		}()
	})
}

func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
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

