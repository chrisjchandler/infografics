package app

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"infograf/internal/model"
	"infograf/internal/sources"
	"infograf/internal/store"
	sqlitestore "infograf/internal/store/sqlite"
)

//go:embed static/index.html static/assets
var staticIndex embed.FS

type Options struct {
	DataDir string
	DBPath  string
}

type Server struct {
	started time.Time
	mux     *http.ServeMux
	store   store.Store
}

func New(opts Options) (*Server, error) {
	dbPath := strings.TrimSpace(opts.DBPath)
	if dbPath == "" {
		dataDir := strings.TrimSpace(opts.DataDir)
		if dataDir == "" {
			dataDir = "data"
		}
		dbPath = filepath.Join(dataDir, "infograf.db")
	}
	st, err := sqlitestore.Open(dbPath)
	if err != nil {
		return nil, err
	}
	s := &Server{
		started: time.Now().UTC(),
		mux:     http.NewServeMux(),
		store:   st,
	}
	s.routes()
	return s, nil
}

func (s *Server) Close() error {
	if s.store == nil {
		return nil
	}
	return s.store.Close()
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) routes() {
	assets, err := fs.Sub(staticIndex, "static/assets")
	if err != nil {
		panic(fmt.Sprintf("failed to mount embedded assets: %v", err))
	}
	s.mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
	s.mux.HandleFunc("/api/v1/health", s.handleHealth)
	s.mux.HandleFunc("/api/v1/readyz", s.handleReady)
	s.mux.HandleFunc("/api/v1/version", s.handleVersion)
	s.mux.HandleFunc("/api/v1/sources/graphite", s.handleGraphiteOverview)
	s.mux.HandleFunc("/api/v1/sources/graphite/catalog", s.handleGraphiteCatalog)
	s.mux.HandleFunc("/api/v1/sources/graphite/scene", s.handleGraphiteScene)
	s.mux.HandleFunc("/api/v1/config", s.handleConfig)
	s.mux.HandleFunc("/api/v1/scenes", s.handleScenes)
	s.mux.HandleFunc("/api/v1/scenes/", s.handleSceneByID)
	s.mux.HandleFunc("/api/v1/revisions", s.handleRevisions)
	s.mux.HandleFunc("/api/v1/sources/adapt", s.handleAdaptSource)
	s.mux.HandleFunc("/api/v1/analysis", s.handleMetricAnalysis)
	s.mux.HandleFunc("/", s.handleIndex)
}

type adaptSourceRequest struct {
	Kind      string `json:"kind"`
	SourceRef string `json:"source_ref"`
	Payload   string `json:"payload"`
}

func (s *Server) handleAdaptSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input adaptSourceRequest
	if err := decodeJSON(r.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid source input", err.Error(), "bad_request")
		return
	}
	if strings.TrimSpace(input.SourceRef) == "" {
		writeError(w, http.StatusBadRequest, "invalid source input", "source_ref is required", "bad_request")
		return
	}
	result, err := sources.DefaultRegistry().Adapt(r.Context(), strings.ToLower(strings.TrimSpace(input.Kind)), sources.Request{
		SourceRef: input.SourceRef,
		Payload:   []byte(input.Payload),
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "source adaptation failed", err.Error(), "source_error")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(staticIndex, "static/index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read index: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, model.HealthResponse{
		Name:          "infograf",
		Status:        "ok",
		UptimeSeconds: int64(time.Since(s.started).Seconds()),
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable", "store not initialized", "not_ready")
		return
	}
	if _, err := s.store.GetConfig(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "store not ready", err.Error(), "not_ready")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ready": true})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    "infograf",
		"version": "0.1.0-dev",
		"go":      "standard-library-with-sqlite-backend",
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, err := s.store.GetConfig(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "config unavailable", err.Error(), "config_error")
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	case http.MethodPut:
		var cfg model.AppConfig
		if err := decodeJSON(r.Body, &cfg); err != nil {
			writeError(w, http.StatusBadRequest, "invalid config", err.Error(), "bad_request")
			return
		}
		updated, err := s.store.PutConfig(r.Context(), cfg)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update config", err.Error(), "config_error")
			return
		}
		writeJSON(w, http.StatusOK, updated)
	default:
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleScenes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		scenes, err := s.store.ListScenes(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list scenes", err.Error(), "scene_error")
			return
		}
		writeJSON(w, http.StatusOK, scenes)
	case http.MethodPost:
		var scene model.SceneDocument
		if err := decodeJSON(r.Body, &scene); err != nil {
			writeError(w, http.StatusBadRequest, "invalid scene", err.Error(), "bad_request")
			return
		}
		created, err := s.store.CreateScene(r.Context(), scene)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create scene", err.Error(), "scene_error")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSceneByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/scenes/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		scene, err := s.store.GetScene(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "scene not found", err.Error(), "not_found")
			return
		}
		writeJSON(w, http.StatusOK, scene)
	case http.MethodPatch:
		var scene model.SceneDocument
		if err := decodeJSON(r.Body, &scene); err != nil {
			writeError(w, http.StatusBadRequest, "invalid scene patch", err.Error(), "bad_request")
			return
		}
		updated, err := s.store.UpdateScene(r.Context(), id, scene)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update scene", err.Error(), "scene_error")
			return
		}
		writeJSON(w, http.StatusOK, updated)
	default:
		w.Header().Set("Allow", "GET, PATCH")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRevisions(w http.ResponseWriter, r *http.Request) {
	limit := 25
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}
	revs, err := s.store.ListRevisions(r.Context(), store.RevisionFilter{
		EntityType: r.URL.Query().Get("entity_type"),
		EntityID:   r.URL.Query().Get("entity_id"),
		Limit:      limit,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list revisions", err.Error(), "revision_error")
		return
	}
	writeJSON(w, http.StatusOK, revs)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, message, detail, code string) {
	writeJSON(w, status, model.ErrorResponse{
		Error:  message,
		Detail: detail,
		Code:   code,
	})
}

func decodeJSON(r io.Reader, dst any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func optionalTimeParam(r *http.Request, name string) (*time.Time, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return nil, nil
	}
	ts, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &ts, nil
}
