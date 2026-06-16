package storage

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/s-piazzano/cosmoria/internal/core"
)

type StorageBackend interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	DownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
	IsLocal() bool
}

func NewBackend(cfg *core.Config) StorageBackend {
	if cfg.S3AccessKey != "" {
		s3 := NewS3Client(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3Region, cfg.S3UseSSL)
		if err := s3.Ping(); err != nil {
			slog.Warn("S3 endpoint unreachable, falling back to local storage", "error", err)
		} else {
			slog.Info("using S3-compatible storage backend", "endpoint", cfg.S3Endpoint, "bucket", cfg.S3Bucket)
			return &S3Backend{client: s3}
		}
	} else {
		slog.Info("S3_ACCESS_KEY not set, using local storage backend")
	}

	path := cfg.StoragePath
	slog.Info("using local storage backend", "path", path)
	return NewLocalBackend(path)
}

type S3Backend struct {
	client *S3Client
}

func (b *S3Backend) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	return b.client.PutObject(key, reader, size, contentType)
}

func (b *S3Backend) DownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return b.client.PresignedGetURL(key, expiry)
}

func (b *S3Backend) Delete(ctx context.Context, key string) error {
	return b.client.DeleteObject(key)
}

func (b *S3Backend) Ping(ctx context.Context) error {
	return b.client.Ping()
}

func (b *S3Backend) IsLocal() bool { return false }
