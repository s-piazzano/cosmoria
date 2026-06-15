package audit

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Logger struct {
	pool *pgxpool.Pool
}

func NewLogger(pool *pgxpool.Pool) *Logger {
	return &Logger{pool: pool}
}

func (l *Logger) Log(ctx context.Context, projectID, userID, action, resource string, resourceID *string, details json.RawMessage, ipAddress string) {
	go func() {
		bgCtx := context.Background()
		_, err := l.pool.Exec(bgCtx,
			`INSERT INTO audit_logs (project_id, user_id, action, resource, resource_id, details, ip_address)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			projectID, userID, action, resource, resourceID, details, ipAddress,
		)
		if err != nil {
			slog.Error("audit: log failed", "error", err)
		}
	}()
}
