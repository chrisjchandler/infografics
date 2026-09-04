package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"infograf/internal/model"
	"infograf/internal/sources"
)

func TestAdaptSourceSupportsEquivalentSimpleInputs(t *testing.T) {
	t.Parallel()
	server, err := New(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	defer server.Close()
	for _, body := range []string{
		`{"kind":"csv","source_ref":"upload","payload":"timestamp,value\n2026-09-02T12:00:00Z,42\n"}`,
		`{"kind":"plaintext","source_ref":"raw","payload":"value 42 1788350400\n"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sources/adapt", bytes.NewBufferString(body))
		res := httptest.NewRecorder()
		server.mux.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("adapt source status = %d, body=%s", res.Code, res.Body.String())
		}
		var result sources.Result
		if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
			t.Fatalf("decode source result: %v", err)
		}
		if len(result.Series) != 1 || result.Series[0].Points[0].Value != 42 {
			t.Fatalf("unexpected canonical result: %#v", result)
		}
	}
}

func TestMetricAnalysisUsesDeterministicLocalProvider(t *testing.T) {
	t.Setenv("INFOGRAF_AI_URL", "")
	server, err := New(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	defer server.Close()
	body := `{"series":{"metric_name":"queue_depth","semantics":{"unit":"jobs"},"points":[{"timestamp":"2026-09-02T12:00:00Z","value":2},{"timestamp":"2026-09-02T12:01:00Z","value":7}]}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analysis", bytes.NewBufferString(body))
	res := httptest.NewRecorder()
	server.mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("analysis status = %d, body=%s", res.Code, res.Body.String())
	}
	var result metricAnalysisResponse
	if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode analysis: %v", err)
	}
	if result.Provider != "local" || result.Summary == "" {
		t.Fatalf("unexpected analysis: %#v", result)
	}
}

func TestGenericInferenceEndpointNormalizesModelDiscoveryURL(t *testing.T) {
	for input, want := range map[string]string{
		"http://foo.bar/v1/models": "http://foo.bar/v1/chat/completions",
		"http://foo.bar/v1":        "http://foo.bar/v1/chat/completions",
		"http://foo.bar":           "http://foo.bar/v1/chat/completions",
	} {
		if got := genericInferenceEndpoint(input); got != want {
			t.Fatalf("endpoint %q = %q, want %q", input, got, want)
		}
	}
}

func TestServerAPIFlow(t *testing.T) {
	t.Parallel()

	srv, err := New(Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	defer srv.Close()

	ts := httptest.NewServer(srv.mux)
	defer ts.Close()

	var health model.HealthResponse
	requestJSON(t, ts.URL+"/api/v1/health", http.MethodGet, nil, &health, http.StatusOK)
	if health.Status != "ok" || health.Name != "infograf" {
		t.Fatalf("unexpected health: %#v", health)
	}

	var ready map[string]any
	requestJSON(t, ts.URL+"/api/v1/readyz", http.MethodGet, nil, &ready, http.StatusOK)
	if ready["ready"] != true {
		t.Fatalf("unexpected ready: %#v", ready)
	}

	var cfg model.AppConfig
	requestJSON(t, ts.URL+"/api/v1/config", http.MethodGet, nil, &cfg, http.StatusOK)
	if cfg.ProductName != "Infografics" || cfg.WorkspaceName != "Operations" {
		t.Fatalf("unexpected seed config: %#v", cfg)
	}

	updatedCfg := model.AppConfig{ProductName: "Infografics", WorkspaceName: "Launchpad", Theme: "dark"}
	requestJSON(t, ts.URL+"/api/v1/config", http.MethodPut, updatedCfg, &cfg, http.StatusOK)
	if cfg.WorkspaceName != "Launchpad" || cfg.Revision <= 1 {
		t.Fatalf("config not updated: %#v", cfg)
	}

	createdSceneReq := model.SceneDocument{
		Name:        "Ops tower",
		Kind:        "dashboard",
		Description: "ops overview",
		Payload:     json.RawMessage(`{"widgets":2}`),
	}
	var created model.SceneDocument
	requestJSON(t, ts.URL+"/api/v1/scenes", http.MethodPost, createdSceneReq, &created, http.StatusCreated)
	if created.ID == "" || created.Revision != 1 {
		t.Fatalf("unexpected created scene: %#v", created)
	}

	var loaded model.SceneDocument
	requestJSON(t, ts.URL+"/api/v1/scenes/"+created.ID, http.MethodGet, nil, &loaded, http.StatusOK)
	if loaded.ID != created.ID || loaded.Name != "Ops tower" {
		t.Fatalf("unexpected loaded scene: %#v", loaded)
	}

	patchReq := model.SceneDocument{Name: "Ops tower v2", Status: "published"}
	requestJSON(t, ts.URL+"/api/v1/scenes/"+created.ID, http.MethodPatch, patchReq, &loaded, http.StatusOK)
	if loaded.Revision != 2 || loaded.Name != "Ops tower v2" || loaded.Status != "published" {
		t.Fatalf("unexpected patched scene: %#v", loaded)
	}

	var revs []model.RevisionRecord
	requestJSON(t, ts.URL+"/api/v1/revisions?entity_type=scene&entity_id="+created.ID, http.MethodGet, nil, &revs, http.StatusOK)
	if len(revs) < 2 {
		t.Fatalf("expected revisions, got %#v", revs)
	}

	var demo model.ReplicationDemoResponse
	requestJSON(t, ts.URL+"/api/v1/demo/replication?at=2026-08-30T10:22:00Z", http.MethodGet, nil, &demo, http.StatusOK)
	if demo.Scene.ID == "" || len(demo.State.Objects) == 0 || len(demo.Trace) == 0 {
		t.Fatalf("unexpected demo payload: %#v", demo)
	}
	if demo.State.Timestamp.Format(time.RFC3339) != "2026-08-30T10:22:00Z" {
		t.Fatalf("unexpected demo timestamp: %s", demo.State.Timestamp.Format(time.RFC3339))
	}

}

func requestJSON[T any](t *testing.T, url, method string, body any, out *T, wantStatus int) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d, body=%s", resp.StatusCode, string(data))
	}
	if out == nil {
		return
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
