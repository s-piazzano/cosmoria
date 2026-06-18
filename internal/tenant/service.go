package tenant

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Tenant struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) CreateTenant(ctx context.Context, projectID, name string) (*Tenant, error) {
	var t Tenant
	err := s.pool.QueryRow(ctx,
		`INSERT INTO tenants (project_id, name) VALUES ($1, $2) RETURNING id, project_id, name, created_at`,
		projectID, name,
	).Scan(&t.ID, &t.ProjectID, &t.Name, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("tenant: create: %w", err)
	}
	return &t, nil
}

func (s *Service) ListTenants(ctx context.Context, projectID string) ([]Tenant, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, project_id, name, created_at FROM tenants WHERE project_id = $1 ORDER BY name`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("tenant: list: %w", err)
	}
	defer rows.Close()

	var tenants []Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Name, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("tenant: scan: %w", err)
		}
		tenants = append(tenants, t)
	}
	return tenants, nil
}

func (s *Service) GetTenant(ctx context.Context, tenantID, projectID string) (*Tenant, error) {
	var t Tenant
	err := s.pool.QueryRow(ctx,
		`SELECT id, project_id, name, created_at FROM tenants WHERE id = $1 AND project_id = $2`,
		tenantID, projectID,
	).Scan(&t.ID, &t.ProjectID, &t.Name, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("tenant: get: %w", err)
	}
	return &t, nil
}

func (s *Service) DeleteTenant(ctx context.Context, tenantID, projectID string) error {
	ct, err := s.pool.Exec(ctx,
		`DELETE FROM tenants WHERE id = $1 AND project_id = $2`,
		tenantID, projectID,
	)
	if err != nil {
		return fmt.Errorf("tenant: delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("tenant: delete: not found")
	}
	return nil
}

func (s *Service) HasAccess(ctx context.Context, userID, tenantID, projectID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_tenants WHERE user_id = $1 AND tenant_id = $2 AND project_id = $3)`,
		userID, tenantID, projectID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("tenant: has access: %w", err)
	}
	return exists, nil
}

func (s *Service) AssignUser(ctx context.Context, userID, tenantID, projectID string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO user_tenants (user_id, tenant_id, project_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		userID, tenantID, projectID,
	)
	if err != nil {
		return fmt.Errorf("tenant: assign user: %w", err)
	}
	return nil
}

func (s *Service) RemoveUser(ctx context.Context, userID, tenantID, projectID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM user_tenants WHERE user_id = $1 AND tenant_id = $2 AND project_id = $3`,
		userID, tenantID, projectID,
	)
	if err != nil {
		return fmt.Errorf("tenant: remove user: %w", err)
	}
	return nil
}

func (s *Service) IsMultitenancyEnabled(ctx context.Context, projectID string) (bool, error) {
	var enabled bool
	err := s.pool.QueryRow(ctx, `SELECT multitenancy_enabled FROM projects WHERE id = $1`, projectID).Scan(&enabled)
	if err != nil {
		return false, fmt.Errorf("tenant: check multitenancy: %w", err)
	}
	return enabled, nil
}
