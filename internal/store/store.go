package store

import (
	"context"

	"infograf/internal/model"
)

type RevisionFilter struct {
	EntityType string
	EntityID   string
	Limit      int
}

type Store interface {
	Close() error

	GetConfig(ctx context.Context) (model.AppConfig, error)
	PutConfig(ctx context.Context, cfg model.AppConfig) (model.AppConfig, error)

	ListScenes(ctx context.Context) ([]model.SceneDocument, error)
	CreateScene(ctx context.Context, scene model.SceneDocument) (model.SceneDocument, error)
	GetScene(ctx context.Context, id string) (model.SceneDocument, error)
	UpdateScene(ctx context.Context, id string, scene model.SceneDocument) (model.SceneDocument, error)

	ListRevisions(ctx context.Context, filter RevisionFilter) ([]model.RevisionRecord, error)
}
