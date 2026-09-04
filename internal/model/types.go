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
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
	Code   string `json:"code,omitempty"`
}

type AppConfig struct {
	ProductName   string        `json:"product_name"`
	WorkspaceName string        `json:"workspace_name"`
	Theme         string        `json:"theme"`
	Storage       StorageConfig `json:"storage"`
	AI            AIConfig      `json:"ai"`
	Revision      int64         `json:"revision"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type AIConfig struct {
	Provider              string `json:"provider"`
	AuthMethod            string `json:"auth_method"`
	OAuthEnabled          bool   `json:"oauth_enabled"`
	URL                   string `json:"url"`
	Model                 string `json:"model"`
	AuthToken             string `json:"auth_token,omitempty"`
	OAuthAuthorizationURL string `json:"oauth_authorization_url,omitempty"`
	OAuthTokenURL         string `json:"oauth_token_url,omitempty"`
	OAuthClientID         string `json:"oauth_client_id,omitempty"`
	OAuthScopes           string `json:"oauth_scopes,omitempty"`
	OAuthRedirectURI      string `json:"oauth_redirect_uri,omitempty"`
}

// StorageConfig keeps the control-plane datastore explicit and deployable.
// SQLite is the current backend; the shape leaves room for a future backend
// without coupling datasource adapters or scenes to persistence.
type StorageConfig struct {
	Backend       string `json:"backend"`
	Path          string `json:"path"`
	RetentionDays int    `json:"retention_days"`
	WAL           bool   `json:"wal"`
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
