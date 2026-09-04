package sqlite

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"infograf/internal/model"
)

func TestOpenSeedAndRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "infograf.db")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	cfg, err := st.GetConfig(ctx)
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if cfg.ProductName != "Infografics" {
		t.Fatalf("unexpected product name: %#v", cfg)
	}
	if cfg.Storage.Backend != "sqlite" || cfg.Storage.Path == "" || !cfg.Storage.WAL {
		t.Fatalf("unexpected storage defaults: %#v", cfg.Storage)
	}
	cfg.Storage.RetentionDays = 7
	cfg.Storage.Path = "data/custom.db"
	if _, err := st.PutConfig(ctx, cfg); err != nil {
		t.Fatalf("update config: %v", err)
	}
	updated, err := st.GetConfig(ctx)
	if err != nil || updated.Storage.RetentionDays != 7 || updated.Storage.Path != "data/custom.db" {
		t.Fatalf("storage round trip: %#v %v", updated.Storage, err)
	}
	cfg.AI.Provider, cfg.AI.AuthMethod, cfg.AI.Enabled = "generic", "api_token", true
	cfg.AI.URL, cfg.AI.Model, cfg.AI.AuthToken = "http://foo.bar/v1/models", "llama-3.2-3b-instruct-q4_k_m", "openclaw-local"
	if _, err := st.PutConfig(ctx, cfg); err != nil {
		t.Fatalf("generic AI config: %v", err)
	}
	aiCfg, err := st.GetConfig(ctx)
	if err != nil || aiCfg.AI.Provider != "generic" || aiCfg.AI.Model != "llama-3.2-3b-instruct-q4_k_m" {
		t.Fatalf("AI config round trip: %#v %v", aiCfg.AI, err)
	}
	scenes, err := st.ListScenes(ctx)
	if err != nil {
		t.Fatalf("list scenes: %v", err)
	}
	if len(scenes) == 0 {
		t.Fatal("expected seeded scene")
	}
	created, err := st.CreateScene(ctx, model.SceneDocument{
		Name:        "Ops overview",
		Kind:        "dashboard",
		Description: "test scene",
		Payload:     json.RawMessage(`{"widgets":1}`),
	})
	if err != nil {
		t.Fatalf("create scene: %v", err)
	}
	if created.ID == "" || created.Revision != 1 {
		t.Fatalf("bad created scene: %#v", created)
	}
	loaded, err := st.GetScene(ctx, created.ID)
	if err != nil {
		t.Fatalf("get scene: %v", err)
	}
	if loaded.Name != "Ops overview" {
		t.Fatalf("unexpected scene: %#v", loaded)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("db not created: %v", err)
	}
}
