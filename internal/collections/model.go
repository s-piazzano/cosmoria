package collections

import "time"

type Field struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type Schema struct {
	Fields []Field `json:"fields"`
}

type Collection struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	Schema    Schema    `json:"schema"`
	CreatedAt time.Time `json:"created_at"`
}
