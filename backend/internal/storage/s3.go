package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-engine/backend/internal/config"
)

type S3Store struct {
	client *s3.Client
	bucket string
	region string
}

func NewS3Store(cfg config.Config) (*S3Store, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.S3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKey, cfg.S3SecretKey, "",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("storage: aws config: %w", err)
	}

	s3Opts := []func(*s3.Options){}
	if cfg.S3Endpoint != "" {
		endpoint := cfg.S3Endpoint
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true // required for MinIO / LocalStack
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)
	return &S3Store{client: client, bucket: cfg.S3Bucket, region: cfg.S3Region}, nil
}

func (s *S3Store) Upload(ctx context.Context, key, contentType string, r io.Reader) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        r,
	})
	if err != nil {
		return "", fmt.Errorf("storage: upload %s: %w", key, err)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key), nil
}

func (s *S3Store) Download(ctx context.Context, rawURL string) (io.ReadCloser, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("storage: invalid url: %w", err)
	}
	key := parsed.Path[1:] // strip leading /

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("storage: download %s: %w", key, err)
	}
	return resp.Body, nil
}

func (s *S3Store) Delete(ctx context.Context, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("storage: invalid url: %w", err)
	}
	key := parsed.Path[1:]

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("storage: delete %s: %w", key, err)
	}
	return nil
}
