package objectstorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

type s3Storage struct {
	client *minio.Client
	config S3Config
}

func NewS3Storage(config S3Config) (ObjectStorage, error) {
	if strings.TrimSpace(config.Bucket) == "" {
		return nil, errors.New("object storage bucket is required")
	}
	endpoint := strings.TrimSpace(config.Endpoint)
	if endpoint == "" {
		return nil, errors.New("object storage endpoint is required")
	}

	secure := strings.HasPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("create s3 client: %w", err)
	}

	exists, err := client.BucketExists(context.Background(), config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket exists: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(context.Background(), config.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}

	return &s3Storage{client: client, config: config}, nil
}

func (s *s3Storage) Descriptor() ObjectStorageDescriptor {
	return ObjectStorageDescriptor{
		Provider: "s3",
		Bucket:   s.config.Bucket,
	}
}

func (s *s3Storage) ObjectExists(ctx context.Context, key string) (bool, error) {
	if strings.TrimSpace(key) == "" {
		return false, errors.New("object key is required")
	}
	_, err := s.client.StatObject(ctx, s.config.Bucket, key, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}
	if minio.ToErrorResponse(err).Code == "NoSuchKey" {
		return false, nil
	}
	return false, fmt.Errorf("stat object: %w", err)
}

func (s *s3Storage) PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("object key is required")
	}
	if reader == nil {
		return errors.New("object reader is nil")
	}
	if size < -1 {
		return errors.New("object size is invalid")
	}

	if _, err := s.client.PutObject(ctx, s.config.Bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	}); err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	return nil
}

func (s *s3Storage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("object key is required")
	}
	obj, err := s.client.GetObject(ctx, s.config.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	return obj, nil
}

func (s *s3Storage) DeleteObject(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("object key is required")
	}
	return s.client.RemoveObject(ctx, s.config.Bucket, key, minio.RemoveObjectOptions{})
}

func (s *s3Storage) PresignedGetURL(ctx context.Context, key string, expire time.Duration) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("object key is required")
	}
	u, err := s.client.PresignedGetObject(ctx, s.config.Bucket, key, expire, url.Values{})
	if err != nil {
		return "", fmt.Errorf("presigned get object: %w", err)
	}
	return u.String(), nil
}
