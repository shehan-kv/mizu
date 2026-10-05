package filestore

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Store struct {
	client *s3.Client
	bucket string
}

type S3Config struct {
	Region string
	Bucket string

	AccessKey string
	SecretKey string

	Endpoint string
}

func NewS3Store(ctx context.Context, cfg S3Config) (*S3Store, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("s3store: load config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// S3-compatible storage (MinIO, LocalStack, etc)
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Store{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (s *S3Store) Save(ctx context.Context, key string, reader io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   reader,
	})

	if err != nil {
		return fmt.Errorf("s3store.Save: put object: %w", err)
	}

	return nil
}

func (s *S3Store) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("s3store.Open: get object: %w", err)
	}

	return out.Body, nil
}

func (s *S3Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})

	if err != nil {
		return fmt.Errorf("s3store.Delete: delete object: %w", err)
	}

	return nil
}
