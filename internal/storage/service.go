package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type File struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	TenantID   string    `json:"tenant_id"`
	Filename   string    `json:"filename"`
	S3Key      string    `json:"s3_key"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mime_type"`
	UploadedBy string    `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type Service struct {
	pool          *pgxpool.Pool
	backend       StorageBackend
	maxUploadSize int64
}

func NewService(pool *pgxpool.Pool, backend StorageBackend, maxUploadSize int64) *Service {
	return &Service{pool: pool, backend: backend, maxUploadSize: maxUploadSize}
}

func (s *Service) MaxUploadSize() int64 {
	return s.maxUploadSize
}

func (s *Service) StoragePath() string {
	if lb, ok := s.backend.(*LocalBackend); ok {
		return lb.basePath
	}
	return ""
}

func (s *Service) Upload(ctx context.Context, projectID, tenantID, filename, mimeType string, reader io.Reader, size int64, uploadedBy string) (*File, error) {
	suffix := generateFileID()
	safeName := sanitizeFilename(filename)
	key := fmt.Sprintf("%s/%s/%s-%s", projectID, tenantID, suffix, safeName)

	if err := s.backend.Upload(ctx, key, reader, size, mimeType); err != nil {
		return nil, fmt.Errorf("storage: upload: %w", err)
	}

	var f File
	err := s.pool.QueryRow(ctx,
		`INSERT INTO files (project_id, tenant_id, filename, s3_key, size, mime_type, uploaded_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, project_id, tenant_id, filename, s3_key, size, mime_type, uploaded_by, created_at`,
		projectID, tenantID, safeName, key, size, mimeType, uploadedBy,
	).Scan(&f.ID, &f.ProjectID, &f.TenantID, &f.Filename, &f.S3Key, &f.Size, &f.MimeType, &f.UploadedBy, &f.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("storage: insert metadata: %w", err)
	}

	return &f, nil
}

type FileWithURL struct {
	File
	PresignedURL string `json:"presigned_url"`
}

func (s *Service) GetByID(ctx context.Context, projectID, tenantID, id string) (*FileWithURL, error) {
	f, err := s.GetMeta(ctx, projectID, tenantID, id)
	if err != nil {
		return nil, err
	}

	url, err := s.backend.DownloadURL(ctx, f.S3Key, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("storage: get download url: %w", err)
	}

	if s.backend.IsLocal() {
		// For local storage, use the API download endpoint
		url = fmt.Sprintf("/api/projects/%s/tenants/%s/files/%s/download", projectID, tenantID, id)
	}

	return &FileWithURL{File: *f, PresignedURL: url}, nil
}

func (s *Service) GetMeta(ctx context.Context, projectID, tenantID, id string) (*File, error) {
	var f File
	err := s.pool.QueryRow(ctx,
		`SELECT id, project_id, tenant_id, filename, s3_key, size, mime_type, uploaded_by, created_at
		 FROM files WHERE id = $1 AND project_id = $2 AND tenant_id = $3`,
		id, projectID, tenantID,
	).Scan(&f.ID, &f.ProjectID, &f.TenantID, &f.Filename, &f.S3Key, &f.Size, &f.MimeType, &f.UploadedBy, &f.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("storage: get: %w", err)
	}
	return &f, nil
}

func (s *Service) List(ctx context.Context, projectID, tenantID string, cursor string, limit int) ([]File, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `SELECT id, project_id, tenant_id, filename, s3_key, size, mime_type, uploaded_by, created_at
	           FROM files WHERE project_id = $1 AND tenant_id = $2`
	args := []any{projectID, tenantID}

	if cursor != "" {
		query += ` AND created_at < $3 ORDER BY created_at DESC LIMIT $4`
		args = append(args, cursor, limit+1)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $3`
		args = append(args, limit+1)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("storage: list: %w", err)
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var f File
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.TenantID, &f.Filename, &f.S3Key, &f.Size, &f.MimeType, &f.UploadedBy, &f.CreatedAt); err != nil {
			return nil, "", fmt.Errorf("storage: scan: %w", err)
		}
		files = append(files, f)
	}

	var nextCursor string
	if len(files) > limit {
		nextCursor = files[limit].CreatedAt.Format(time.RFC3339Nano)
		files = files[:limit]
	}

	return files, nextCursor, nil
}

func (s *Service) Delete(ctx context.Context, projectID, tenantID, id string) error {
	var s3Key string
	err := s.pool.QueryRow(ctx,
		`DELETE FROM files WHERE id = $1 AND project_id = $2 AND tenant_id = $3
		 RETURNING s3_key`,
		id, projectID, tenantID,
	).Scan(&s3Key)
	if err != nil {
		return fmt.Errorf("storage: delete: %w", err)
	}

	if err := s.backend.Delete(ctx, s3Key); err != nil {
		return fmt.Errorf("storage: delete from backend: %w", err)
	}

	return nil
}
