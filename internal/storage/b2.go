package storage

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Habeebamoo/Prism/internal/configs"
	"github.com/Habeebamoo/Prism/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
				ContentType:   aws.String(utils.GetContentType(path)),
			})
			if err != nil {
				return fmt.Errorf("failed to upload %s to B2: %w", key, err)
			}

			return nil
		}()
	})
}

func (b *B2Client) GetObject(ctx context.Context, key string) (*s3.GetObjectOutput, error) {
	return b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key: aws.String(key),
	})
}

func (b *B2Client) DeleteVideoFolder(ctx context.Context, videoId string) error {
	cleanId := strings.TrimSpace(videoId)
	if cleanId == "" || cleanId == "." || cleanId == "/" {
		return fmt.Errorf("refusing to delete empty or root prefix")
	}

	prefix := fmt.Sprintf("videos/%s/", cleanId)
	paginator := s3.NewListObjectsV2Paginator(b.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(b.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list b2 objects for %s: %w", prefix, err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		objects := make([]types.ObjectIdentifier, 0, len(page.Contents))
		for _, obj := range page.Contents {
			objects = append(objects, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		_, err = b.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(b.bucket),
			Delete: &types.Delete{
				Objects: objects,
				Quiet: aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("batch delete b2 objects for %s: %w", prefix, err)
		}
	}

	return nil
}

