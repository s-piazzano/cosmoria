package collections

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) CreateCollection(ctx context.Context, projectID, name string, schema Schema) (*Collection, error) {
	var c Collection
	err := s.pool.QueryRow(ctx,
		`INSERT INTO collections (project_id, name, schema) VALUES ($1, $2, $3)
		 RETURNING id, project_id, name, schema, created_at`,
		projectID, name, schema,
	).Scan(&c.ID, &c.ProjectID, &c.Name, &c.Schema, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("collections: create: %w", err)
	}
	return &c, nil
}

func (s *Service) ListCollections(ctx context.Context, projectID string) ([]Collection, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, project_id, name, schema, created_at FROM collections WHERE project_id = $1 ORDER BY name`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("collections: list: %w", err)
	}
	defer rows.Close()

	var collections []Collection
	for rows.Next() {
		var c Collection
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &c.Schema, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("collections: scan: %w", err)
		}
		collections = append(collections, c)
	}
	return collections, nil
}

func (s *Service) GetCollection(ctx context.Context, collectionID, projectID string) (*Collection, error) {
	var c Collection
	err := s.pool.QueryRow(ctx,
		`SELECT id, project_id, name, schema, created_at FROM collections WHERE id = $1 AND project_id = $2`,
		collectionID, projectID,
	).Scan(&c.ID, &c.ProjectID, &c.Name, &c.Schema, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("collections: get: %w", err)
	}
	return &c, nil
}

func (s *Service) UpdateCollectionSchema(ctx context.Context, collectionID, projectID string, schema Schema) (*Collection, error) {
	var c Collection
	err := s.pool.QueryRow(ctx,
		`UPDATE collections SET schema = $1 WHERE id = $2 AND project_id = $3
		 RETURNING id, project_id, name, schema, created_at`,
		schema, collectionID, projectID,
	).Scan(&c.ID, &c.ProjectID, &c.Name, &c.Schema, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("collections: update schema: %w", err)
	}
	return &c, nil
}

func (s *Service) DeleteCollection(ctx context.Context, collectionID, projectID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM collections WHERE id = $1 AND project_id = $2`,
		collectionID, projectID,
	)
	if err != nil {
		return fmt.Errorf("collections: delete: %w", err)
	}
	return nil
}

func ValidateData(data map[string]any, schema Schema) error {
	for _, f := range schema.Fields {
		val, ok := data[f.Name]
		if f.Required && (!ok || val == nil) {
			return fmt.Errorf("field %q is required", f.Name)
		}
		if ok && val != nil {
			if err := validateType(val, f.Type); err != nil {
				return fmt.Errorf("field %q: %w", f.Name, err)
			}
		}
	}
	return nil
}

func validateType(val any, typ string) error {
	switch typ {
	case "string":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("expected string, got %T", val)
		}
	case "number":
		switch val.(type) {
		case float64, json.Number:
			return nil
		}
		return fmt.Errorf("expected number, got %T", val)
	case "boolean":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", val)
		}
	default:
		return fmt.Errorf("unsupported type %q", typ)
	}
	return nil
}
