package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Client interface {
	Upload(ctx context.Context, bucket, key string, body io.Reader) error
	Download(ctx context.Context, bucket, key string, writer io.WriterAt) error
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
	DeleteMultiple(ctx context.Context, bucket string, keys []string) error
	GeneratePresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error)
	ListObjects(ctx context.Context, bucket, prefix string) ([]string, error)
}

type client struct {
	s3Client   *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	bucket     string
}

func NewS3Client(region, accessKeyID, secretAccessKey, bucket string) (S3Client, error) {
	ctx := context.Background()
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	return &client{
		s3Client:   s3Client,
		uploader:   manager.NewUploader(s3Client),
		downloader: manager.NewDownloader(s3Client),
		bucket:     bucket,
	}, nil
}

func (c *client) Upload(ctx context.Context, bucket, key string, body io.Reader) error {
	_, err := c.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}
	return nil
}

func (c *client) Download(ctx context.Context, bucket, key string, writer io.WriterAt) error {
	_, err := c.downloader.Download(ctx, writer, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}
	return nil
}

func (c *client) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	return result.Body, nil
}

func (c *client) Delete(ctx context.Context, bucket, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	return nil
}

func (c *client) DeleteMultiple(ctx context.Context, bucket string, keys []string) error {
	objects := make([]s3Types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = s3Types.ObjectIdentifier{
			Key: aws.String(key),
		}
	}

	_, err := c.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3Types.Delete{
			Objects: objects,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete multiple objects from S3: %w", err)
	}
	return nil
}

func (c *client) GeneratePresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.s3Client)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedReq.URL, nil
}

func (c *client) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	result, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	keys := make([]string, 0, len(result.Contents))
	for _, obj := range result.Contents {
		keys = append(keys, *obj.Key)
	}

	return keys, nil
}
