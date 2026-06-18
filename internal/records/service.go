package records

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-piazzano/cosmoria/internal/collections"
)

type Service struct {
	pool  *pgxpool.Pool
	colls *collections.Service
}

func NewService(pool *pgxpool.Pool, colls *collections.Service) *Service {
	return &Service{pool: pool, colls: colls}
}

func (s *Service) CreateRecord(ctx context.Context, projectID string, tenantID *string, collectionID string, data map[string]any) (*Record, error) {
	coll, err := s.colls.GetCollection(ctx, collectionID, projectID)
	if err != nil {
		return nil, fmt.Errorf("records: %w", err)
	}

	if err := collections.ValidateData(data, coll.Schema); err != nil {
		return nil, fmt.Errorf("records: validate: %w", err)
	}

	var r Record
	err = s.pool.QueryRow(ctx,
		`INSERT INTO records (project_id, tenant_id, collection_id, data)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, project_id, tenant_id, collection_id, data, created_at, updated_at`,
		projectID, tenantID, collectionID, data,
	).Scan(&r.ID, &r.ProjectID, &r.TenantID, &r.CollectionID, &r.Data, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("records: create: %w", err)
	}
	return &r, nil
}

func (s *Service) ListRecords(ctx context.Context, projectID string, tenantID *string, collectionID string, cursor string, limit int) ([]Record, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `SELECT id, project_id, tenant_id, collection_id, data, created_at, updated_at
	           FROM records
	           WHERE project_id = $1 AND collection_id = $2`
	args := []any{projectID, collectionID}
	argIdx := 3
	if tenantID != nil {
		query += ` AND tenant_id = $3`
		args = append(args, *tenantID)
		argIdx = 4
	} else {
		query += ` AND tenant_id IS NULL`
	}

	if cursor != "" {
		query += fmt.Sprintf(` AND id > $%d ORDER BY created_at, id LIMIT $%d`, argIdx, argIdx+1)
		args = append(args, cursor, limit+1)
	} else {
		query += fmt.Sprintf(` ORDER BY created_at, id LIMIT $%d`, argIdx)
		args = append(args, limit+1)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("records: list: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.ID, &r.ProjectID, &r.TenantID, &r.CollectionID, &r.Data, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, "", fmt.Errorf("records: scan: %w", err)
		}
		records = append(records, r)
	}

	var nextCursor string
	if len(records) > limit {
		nextCursor = records[limit].ID
		records = records[:limit]
	}

	return records, nextCursor, nil
}

func (s *Service) GetRecord(ctx context.Context, recordID, projectID string, tenantID *string) (*Record, error) {
	var r Record
	query := `SELECT id, project_id, tenant_id, collection_id, data, created_at, updated_at
	           FROM records WHERE id = $1 AND project_id = $2`
	args := []any{recordID, projectID}
	if tenantID != nil {
		query += ` AND tenant_id = $3`
		args = append(args, *tenantID)
	} else {
		query += ` AND tenant_id IS NULL`
	}
	err := s.pool.QueryRow(ctx, query, args...).Scan(&r.ID, &r.ProjectID, &r.TenantID, &r.CollectionID, &r.Data, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("records: get: %w", err)
	}
	return &r, nil
}

func (s *Service) UpdateRecord(ctx context.Context, recordID, projectID string, tenantID *string, data map[string]any) (*Record, error) {
	existing, err := s.GetRecord(ctx, recordID, projectID, tenantID)
	if err != nil {
		return nil, err
	}

	coll, err := s.colls.GetCollection(ctx, existing.CollectionID, projectID)
	if err != nil {
		return nil, fmt.Errorf("records: %w", err)
	}

	if err := collections.ValidateData(data, coll.Schema); err != nil {
		return nil, fmt.Errorf("records: validate: %w", err)
	}

	var r Record
	query := `UPDATE records SET data = $1, updated_at = now()
	           WHERE id = $2 AND project_id = $3`
	args := []any{data, recordID, projectID}
	if tenantID != nil {
		query += ` AND tenant_id = $4`
		args = append(args, *tenantID)
	} else {
		query += ` AND tenant_id IS NULL`
	}
	query += ` RETURNING id, project_id, tenant_id, collection_id, data, created_at, updated_at`
	err = s.pool.QueryRow(ctx, query, args...).Scan(&r.ID, &r.ProjectID, &r.TenantID, &r.CollectionID, &r.Data, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("records: update: %w", err)
	}
	return &r, nil
}

func (s *Service) DeleteRecord(ctx context.Context, recordID, projectID string, tenantID *string) error {
	query := `DELETE FROM records WHERE id = $1 AND project_id = $2`
	args := []any{recordID, projectID}
	if tenantID != nil {
		query += ` AND tenant_id = $3`
		args = append(args, *tenantID)
	} else {
		query += ` AND tenant_id IS NULL`
	}
	_, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("records: delete: %w", err)
	}
	return nil
}
