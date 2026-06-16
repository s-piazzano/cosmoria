package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ApiKey struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	KeyPrefix string    `json:"key_prefix"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateApiKeyResult struct {
	ApiKey   ApiKey  `json:"api_key"`
	PlainKey string  `json:"plain_key"`
}

type ApiKeyService struct {
	pool *pgxpool.Pool
}

func NewApiKeyService(pool *pgxpool.Pool) *ApiKeyService {
	return &ApiKeyService{pool: pool}
}

func generateApiKey() (plainKey, keyHash, prefix string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("apikey: generate: %w", err)
	}
	plainKey = "ck_" + hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(plainKey))
	keyHash = hex.EncodeToString(hash[:])
	prefix = keyHash[:8]
	return plainKey, keyHash, prefix, nil
}

func (s *ApiKeyService) Create(ctx context.Context, projectID, userID, name string) (*CreateApiKeyResult, error) {
	plainKey, keyHash, prefix, err := generateApiKey()
	if err != nil {
		return nil, err
	}

	var key ApiKey
	err = s.pool.QueryRow(ctx,
		`INSERT INTO api_keys (project_id, user_id, name, key_hash)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, project_id, user_id, name, created_at`,
		projectID, userID, name, keyHash,
	).Scan(&key.ID, &key.ProjectID, &key.UserID, &key.Name, &key.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("apikey: create: %w", err)
	}
	key.KeyPrefix = prefix

	return &CreateApiKeyResult{
		ApiKey:   key,
		PlainKey: plainKey,
	}, nil
}

func (s *ApiKeyService) Validate(ctx context.Context, key string) (*Claims, error) {
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	var userID, projectID string
	err := s.pool.QueryRow(ctx,
		`SELECT user_id, project_id FROM api_keys WHERE key_hash = $1`,
		keyHash,
	).Scan(&userID, &projectID)
	if err != nil {
		return nil, fmt.Errorf("apikey: invalid key")
	}

	return &Claims{
		UserID:    userID,
		ProjectID: projectID,
	}, nil
}

func (s *ApiKeyService) List(ctx context.Context, projectID string) ([]ApiKey, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, project_id, user_id, name, SUBSTRING(key_hash, 1, 8), created_at
		 FROM api_keys WHERE project_id = $1
		 ORDER BY created_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("apikey: list: %w", err)
	}
	defer rows.Close()

	var keys []ApiKey
	for rows.Next() {
		var k ApiKey
		if err := rows.Scan(&k.ID, &k.ProjectID, &k.UserID, &k.Name, &k.KeyPrefix, &k.CreatedAt); err != nil {
			return nil, fmt.Errorf("apikey: scan: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *ApiKeyService) Delete(ctx context.Context, id, projectID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM api_keys WHERE id = $1 AND project_id = $2`,
		id, projectID,
	)
	if err != nil {
		return fmt.Errorf("apikey: delete: %w", err)
	}
	return nil
}
