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
	if cfg.ProductName != "Infograf" {
		t.Fatalf("unexpected product name: %#v", cfg)
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
