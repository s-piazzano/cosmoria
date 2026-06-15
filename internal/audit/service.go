package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditEntry struct {
	ID         string           `json:"id"`
	ProjectID  string           `json:"project_id"`
	UserID     string           `json:"user_id"`
	Action     string           `json:"action"`
	Resource   string           `json:"resource"`
	ResourceID *string          `json:"resource_id,omitempty"`
	Details    json.RawMessage  `json:"details,omitempty"`
	IPAddress  string           `json:"ip_address"`
	CreatedAt  time.Time        `json:"created_at"`
}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) List(ctx context.Context, projectID string, cursor string, limit int) ([]AuditEntry, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `SELECT id, project_id, user_id, action, resource, resource_id, details, ip_address, created_at
	           FROM audit_logs WHERE project_id = $1`
	args := []any{projectID}

	if cursor != "" {
		query += ` AND created_at < $2 ORDER BY created_at DESC LIMIT $3`
		args = append(args, cursor, limit+1)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $2`
		args = append(args, limit+1)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("audit: list: %w", err)
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.UserID, &e.Action, &e.Resource, &e.ResourceID, &e.Details, &e.IPAddress, &e.CreatedAt); err != nil {
			return nil, "", fmt.Errorf("audit: scan: %w", err)
		}
		entries = append(entries, e)
	}

	var nextCursor string
	if len(entries) > limit {
		nextCursor = entries[limit].CreatedAt.Format(time.RFC3339Nano)
		entries = entries[:limit]
	}

	return entries, nextCursor, nil
}
