package objectstorage

import (
	"context"
	"io"
	"time"
)

type ObjectStorageDescriptor struct {
	Provider string
	Bucket   string
}

type ObjectStorage interface {
	Descriptor() ObjectStorageDescriptor
	ObjectExists(ctx context.Context, key string) (bool, error)
	PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, key string) error
	PresignedGetURL(ctx context.Context, key string, expire time.Duration) (string, error)
}
