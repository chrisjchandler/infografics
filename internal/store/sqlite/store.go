package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"infograf/internal/demo"
	"infograf/internal/model"
	"infograf/internal/store"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure sqlite: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.seed(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS config (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			revision INTEGER NOT NULL,
			payload TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS scenes (
			id TEXT PRIMARY KEY,
			slug TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			kind TEXT NOT NULL,
			status TEXT NOT NULL,
			description TEXT NOT NULL,
			revision INTEGER NOT NULL,
			payload TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS revisions (
			id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			action TEXT NOT NULL,
			summary TEXT NOT NULL,
			payload TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_revisions_entity ON revisions(entity_type, entity_id, created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_scenes_updated ON scenes(updated_at DESC);`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

func (s *Store) seed(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM config`).Scan(&count); err != nil {
		return fmt.Errorf("count config: %w", err)
	}
	if count == 0 {
		cfg := model.AppConfig{
			ProductName:  "Infograf",
			WorkspaceName: "Crayons",
			Theme:        "dark",
			Revision:     1,
			UpdatedAt:    time.Now().UTC(),
		}
		if _, err := s.putConfigTx(ctx, cfg, false); err != nil {
			return err
		}
	}

	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM scenes`).Scan(&count); err != nil {
		return fmt.Errorf("count scenes: %w", err)
	}
	if count == 0 {
		scene, err := demoSceneDocument()
		if err != nil {
			return err
		}
		if _, err := s.createSceneTx(ctx, scene, false); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetConfig(ctx context.Context) (model.AppConfig, error) {
	row := s.db.QueryRowContext(ctx, `SELECT payload FROM config WHERE id = 1`)
	var payload string
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AppConfig{}, fmt.Errorf("config not found")
		}
		return model.AppConfig{}, err
	}
	var cfg model.AppConfig
	if err := json.Unmarshal([]byte(payload), &cfg); err != nil {
		return model.AppConfig{}, err
	}
	return cfg, nil
}

func (s *Store) PutConfig(ctx context.Context, cfg model.AppConfig) (model.AppConfig, error) {
	return s.putConfigTx(ctx, cfg, false)
}

func (s *Store) putConfigTx(ctx context.Context, cfg model.AppConfig, silent bool) (model.AppConfig, error) {
	current, _ := s.GetConfig(ctx)
	if strings.TrimSpace(cfg.ProductName) == "" {
		cfg.ProductName = "Infograf"
	}
	if strings.TrimSpace(cfg.WorkspaceName) == "" {
		cfg.WorkspaceName = "Crayons"
	}
	if strings.TrimSpace(cfg.Theme) == "" {
		cfg.Theme = "dark"
	}
	cfg.UpdatedAt = time.Now().UTC()
	if cfg.Revision == 0 {
		cfg.Revision = current.Revision + 1
	}
	payload, err := json.Marshal(cfg)
	if err != nil {
		return model.AppConfig{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.AppConfig{}, err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `INSERT INTO config(id, revision, payload, updated_at) VALUES(1, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET revision=excluded.revision, payload=excluded.payload, updated_at=excluded.updated_at`,
		cfg.Revision, string(payload), cfg.UpdatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return model.AppConfig{}, err
	}
	if !silent {
		if err := s.insertRevisionTx(ctx, tx, model.RevisionRecord{
			EntityType: "config",
			EntityID:   "config",
			Action:     "upsert",
			Summary:    "configuration updated",
			Payload:    payload,
			CreatedAt:  cfg.UpdatedAt,
		}); err != nil {
			return model.AppConfig{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return model.AppConfig{}, err
	}
	return cfg, nil
}

func (s *Store) ListScenes(ctx context.Context) ([]model.SceneDocument, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, slug, name, kind, status, description, revision, payload, created_at, updated_at FROM scenes ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scenes []model.SceneDocument
	for rows.Next() {
		var scene model.SceneDocument
		var payload string
		var createdAt, updatedAt string
		if err := rows.Scan(&scene.ID, &scene.Slug, &scene.Name, &scene.Kind, &scene.Status, &scene.Description, &scene.Revision, &payload, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		scene.Payload = json.RawMessage(payload)
		scene.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		scene.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, err
		}
		scenes = append(scenes, scene)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if scenes == nil {
		return []model.SceneDocument{}, nil
	}
	return scenes, nil
}

func (s *Store) CreateScene(ctx context.Context, scene model.SceneDocument) (model.SceneDocument, error) {
	return s.createSceneTx(ctx, scene, false)
}

func (s *Store) createSceneTx(ctx context.Context, scene model.SceneDocument, silent bool) (model.SceneDocument, error) {
	now := time.Now().UTC()
	if strings.TrimSpace(scene.ID) == "" {
		scene.ID = newID("scene")
	}
	if strings.TrimSpace(scene.Name) == "" {
		scene.Name = scene.Slug
	}
	if strings.TrimSpace(scene.Slug) == "" {
		scene.Slug = slugify(scene.Name)
	}
	if strings.TrimSpace(scene.Kind) == "" {
		scene.Kind = "scene"
	}
	if strings.TrimSpace(scene.Status) == "" {
		scene.Status = "draft"
	}
	if len(scene.Payload) == 0 {
		scene.Payload = json.RawMessage(`{}`)
	}
	scene.Revision = 1
	scene.CreatedAt = now
	scene.UpdatedAt = now
	payload, err := json.Marshal(scene)
	if err != nil {
		return model.SceneDocument{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.SceneDocument{}, err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `INSERT INTO scenes(id, slug, name, kind, status, description, revision, payload, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		scene.ID, scene.Slug, scene.Name, scene.Kind, scene.Status, scene.Description, scene.Revision, string(payload),
		scene.CreatedAt.Format(time.RFC3339Nano), scene.UpdatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return model.SceneDocument{}, err
	}
	if !silent {
		if err := s.insertRevisionTx(ctx, tx, model.RevisionRecord{
			EntityType: "scene",
			EntityID:   scene.ID,
			Action:     "create",
			Summary:    "scene created",
			Payload:    payload,
			CreatedAt:  scene.CreatedAt,
		}); err != nil {
			return model.SceneDocument{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return model.SceneDocument{}, err
	}
	return scene, nil
}

func (s *Store) GetScene(ctx context.Context, id string) (model.SceneDocument, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, slug, name, kind, status, description, revision, payload, created_at, updated_at FROM scenes WHERE id = ?`, id)
	var scene model.SceneDocument
	var payload string
	var createdAt, updatedAt string
	if err := row.Scan(&scene.ID, &scene.Slug, &scene.Name, &scene.Kind, &scene.Status, &scene.Description, &scene.Revision, &payload, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.SceneDocument{}, fmt.Errorf("scene %q not found", id)
		}
		return model.SceneDocument{}, err
	}
	scene.Payload = json.RawMessage(payload)
	var err error
	scene.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return model.SceneDocument{}, err
	}
	scene.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return model.SceneDocument{}, err
	}
	return scene, nil
}

func (s *Store) UpdateScene(ctx context.Context, id string, scene model.SceneDocument) (model.SceneDocument, error) {
	current, err := s.GetScene(ctx, id)
	if err != nil {
		return model.SceneDocument{}, err
	}
	if strings.TrimSpace(scene.Name) != "" {
		current.Name = scene.Name
	}
	if strings.TrimSpace(scene.Slug) != "" {
		current.Slug = scene.Slug
	}
	if strings.TrimSpace(scene.Kind) != "" {
		current.Kind = scene.Kind
	}
	if strings.TrimSpace(scene.Status) != "" {
		current.Status = scene.Status
	}
	current.Description = scene.Description
	if len(scene.Payload) > 0 {
		current.Payload = scene.Payload
	}
	current.Revision++
	current.UpdatedAt = time.Now().UTC()
	payload, err := json.Marshal(current)
	if err != nil {
		return model.SceneDocument{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.SceneDocument{}, err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `UPDATE scenes SET slug = ?, name = ?, kind = ?, status = ?, description = ?, revision = ?, payload = ?, updated_at = ? WHERE id = ?`,
		current.Slug, current.Name, current.Kind, current.Status, current.Description, current.Revision, string(payload), current.UpdatedAt.Format(time.RFC3339Nano), id)
	if err != nil {
		return model.SceneDocument{}, err
	}
	if err := s.insertRevisionTx(ctx, tx, model.RevisionRecord{
		EntityType: "scene",
		EntityID:   current.ID,
		Action:     "update",
		Summary:    "scene updated",
		Payload:    payload,
		CreatedAt:  current.UpdatedAt,
	}); err != nil {
		return model.SceneDocument{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.SceneDocument{}, err
	}
	return current, nil
}

func (s *Store) ListRevisions(ctx context.Context, filter store.RevisionFilter) ([]model.RevisionRecord, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	query := `SELECT id, entity_type, entity_id, action, summary, payload, created_at FROM revisions`
	args := []any{}
	clauses := []string{}
	if strings.TrimSpace(filter.EntityType) != "" {
		clauses = append(clauses, "entity_type = ?")
		args = append(args, filter.EntityType)
	}
	if strings.TrimSpace(filter.EntityID) != "" {
		clauses = append(clauses, "entity_id = ?")
		args = append(args, filter.EntityID)
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var revisions []model.RevisionRecord
	for rows.Next() {
		var rev model.RevisionRecord
		var payload string
		var createdAt string
		if err := rows.Scan(&rev.ID, &rev.EntityType, &rev.EntityID, &rev.Action, &rev.Summary, &payload, &createdAt); err != nil {
			return nil, err
		}
		rev.Payload = json.RawMessage(payload)
		rev.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		revisions = append(revisions, rev)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if revisions == nil {
		return []model.RevisionRecord{}, nil
	}
	return revisions, nil
}

func (s *Store) insertRevisionTx(ctx context.Context, tx *sql.Tx, rev model.RevisionRecord) error {
	if strings.TrimSpace(rev.ID) == "" {
		rev.ID = newID("rev")
	}
	if len(rev.Payload) == 0 {
		rev.Payload = json.RawMessage(`{}`)
	}
	if rev.CreatedAt.IsZero() {
		rev.CreatedAt = time.Now().UTC()
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO revisions(id, entity_type, entity_id, action, summary, payload, created_at)
		VALUES(?, ?, ?, ?, ?, ?, ?)`,
		rev.ID, rev.EntityType, rev.EntityID, rev.Action, rev.Summary, string(rev.Payload), rev.CreatedAt.Format(time.RFC3339Nano))
	return err
}

func demoSceneDocument() (model.SceneDocument, error) {
	return demo.ReplicationSceneDocument()
}

func newID(prefix string) string {
	return prefix + "_" + time.Now().UTC().Format("20060102T150405.000000000Z07")
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '/' || r == ':':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "scene"
	}
	return result
}

var _ store.Store = (*Store)(nil)
