package model

import (
	"encoding/json"
	"time"
)

type HealthResponse struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Detail  string `json:"detail,omitempty"`
	Code    string `json:"code,omitempty"`
}

type AppConfig struct {
	ProductName   string    `json:"product_name"`
	WorkspaceName string    `json:"workspace_name"`
	Theme         string    `json:"theme"`
	Revision      int64     `json:"revision"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SceneDocument struct {
	ID          string          `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Kind        string          `json:"kind"`
	Status      string          `json:"status"`
	Description string          `json:"description"`
	Revision    int64           `json:"revision"`
	Payload     json.RawMessage `json:"payload,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type RevisionRecord struct {
	ID         string          `json:"id"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Action     string          `json:"action"`
	Summary    string          `json:"summary"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
