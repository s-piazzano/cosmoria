package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type LocalBackend struct {
	basePath string
}

func NewLocalBackend(basePath string) *LocalBackend {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		slog.Warn("cannot create storage directory, using /tmp", "error", err)
		basePath = "/tmp/cosmoria-files"
		os.MkdirAll(basePath, 0755)
	}
	return &LocalBackend{basePath: basePath}
}

func (b *LocalBackend) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	fullPath := filepath.Join(b.basePath, key)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("local: create dir: %w", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("local: create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("local: write file: %w", err)
	}
	return nil
}

func (b *LocalBackend) DownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// Return the key itself; the service layer constructs the API download URL
	return key, nil
}

func (b *LocalBackend) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(b.basePath, key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local: delete: %w", err)
	}
	return nil
}

func (b *LocalBackend) Ping(ctx context.Context) error { return nil }

func (b *LocalBackend) IsLocal() bool { return true }
