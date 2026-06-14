package records

import "time"

type Record struct {
	ID           string         `json:"id"`
	ProjectID    string         `json:"project_id"`
	TenantID     string         `json:"tenant_id"`
	CollectionID string         `json:"collection_id"`
	Data         map[string]any `json:"data"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
