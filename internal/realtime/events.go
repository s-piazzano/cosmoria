package realtime

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID         string          `json:"id,omitempty"`
	ProjectID  string          `json:"project_id"`
	TenantID   string          `json:"tenant_id"`
	Resource   string          `json:"resource"`
	Action     string          `json:"action"`
	ResourceID string          `json:"resource_id"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
}

type WSMessage struct {
	Type      string          `json:"type"`
	TenantID  string          `json:"tenant_id,omitempty"`
	Event     *Event          `json:"event,omitempty"`
	Message   string          `json:"message,omitempty"`
}
