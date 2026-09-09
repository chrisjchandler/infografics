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
	if len(scenes) != 0 {
		t.Fatalf("new store contains synthetic scenes: %#v", scenes)
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
	if string(loaded.Payload) != `{"widgets":1}` {
		t.Fatalf("scene payload was wrapped instead of round-tripped: %s", loaded.Payload)
	}
	listed, err := st.ListScenes(ctx)
	if err != nil || len(listed) != 1 || string(listed[0].Payload) != `{"widgets":1}` {
		t.Fatalf("listed scene payload did not round-trip: %#v, %v", listed, err)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("db not created: %v", err)
	}
}

func TestOpenRepairsWrappedPayloadAndRemovesLegacyReplicationDemo(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "infograf.db")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	ctx := context.Background()
	created, err := st.CreateScene(ctx, model.SceneDocument{Name: "Real dashboard", Kind: "graphite-dashboard", Payload: json.RawMessage(`{"metrics":["real.metric"]}`)})
	if err != nil {
		t.Fatalf("create scene: %v", err)
	}
	wrapped, err := json.Marshal(created)
	if err != nil {
		t.Fatalf("marshal legacy wrapper: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE scenes SET payload = ? WHERE id = ?`, string(wrapped), created.ID); err != nil {
		t.Fatalf("install wrapped payload: %v", err)
	}
	stamp := created.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00")
	if _, err := st.db.ExecContext(ctx, `INSERT INTO scenes(id, slug, name, kind, status, description, revision, payload, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"scene_replication_demo", "database-replication-health", "Database replication health", "replication-demo", "published", "legacy dummy", 1, `{}`, stamp, stamp); err != nil {
		t.Fatalf("install legacy demo: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	reopened, err := Open(dbPath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	scenes, err := reopened.ListScenes(ctx)
	if err != nil {
		t.Fatalf("list repaired scenes: %v", err)
	}
	if len(scenes) != 1 || scenes[0].ID != created.ID {
		t.Fatalf("legacy demo was not removed safely: %#v", scenes)
	}
	if string(scenes[0].Payload) != `{"metrics":["real.metric"]}` {
		t.Fatalf("wrapped payload was not repaired: %s", scenes[0].Payload)
	}
}

func TestAIConfigurationPersistsAcrossReopen(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "infograf.db")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	ctx := context.Background()
	cfg, err := st.GetConfig(ctx)
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	cfg.AI = model.AIConfig{
		Provider:   "generic",
		AuthMethod: "api_token",
		Enabled:    true,
		URL:        "http://172.16.0.47:8080/v1/models",
		Model:      "llama-3.2-3b-instruct-q4_k_m",
		AuthToken:  "test-token",
	}
	written, err := st.PutConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("put config: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	reopened, err := Open(dbPath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	persisted, err := reopened.GetConfig(ctx)
	if err != nil {
		t.Fatalf("get reopened config: %v", err)
	}
	if persisted.Revision != written.Revision || persisted.AI != cfg.AI {
		t.Fatalf("AI config after reopen = %#v revision %d, want %#v revision %d", persisted.AI, persisted.Revision, cfg.AI, written.Revision)
	}
}
