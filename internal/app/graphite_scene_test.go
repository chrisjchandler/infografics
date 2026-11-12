package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"infograf/internal/model"
)

func TestDiscoverGraphiteMetricOptionsPrefersBulkIndex(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/metrics/index.json" {
			t.Errorf("unexpected fallback request: %s", r.URL.String())
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode([]string{"servers.b.cpu.idle", "servers.a.cpu.used"})
	}))
	defer server.Close()

	options, err := discoverGraphiteMetricOptions(t.Context(), server.URL)
	if err != nil {
		t.Fatalf("discover metrics: %v", err)
	}
	if requests.Load() != 1 || len(options) != 2 {
		t.Fatalf("requests = %d, options = %#v", requests.Load(), options)
	}
}

func TestGraphiteMetricCatalogCachesDiscovery(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	graphite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_ = json.NewEncoder(w).Encode([]string{"servers.nas.cpu.idle"})
	}))
	defer graphite.Close()

	server := &Server{}
	first, firstStatus, err := server.graphiteMetricOptions(context.Background(), graphite.URL)
	if err != nil {
		t.Fatalf("first discovery: %v", err)
	}
	second, secondStatus, err := server.graphiteMetricOptions(context.Background(), graphite.URL)
	if err != nil {
		t.Fatalf("cached discovery: %v", err)
	}
	if requests.Load() != 1 || firstStatus != "miss" || secondStatus != "hit" || !reflect.DeepEqual(first, second) {
		t.Fatalf("requests=%d statuses=%q/%q first=%#v second=%#v", requests.Load(), firstStatus, secondStatus, first, second)
	}
}

func TestTimelineProgressSeries(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	primary := model.CanonicalSeries{
		MetricName: "metric",
		Points: []model.CanonicalPoint{
			{Timestamp: base, Value: 10},
			{Timestamp: base.Add(time.Minute), Value: 20},
			{Timestamp: base.Add(2 * time.Minute), Value: 30},
		},
	}

	progress := timelineProgressSeries(primary, "graphite")
	if len(progress.Points) != 3 {
		t.Fatalf("got %d points, want 3", len(progress.Points))
	}
	if progress.Points[0].Value != 0 || progress.Points[1].Value != 50 || progress.Points[2].Value != 100 {
		t.Fatalf("unexpected progress points: %#v", progress.Points)
	}
}

func TestTimelineProgressSeriesUsesElapsedTime(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	primary := model.CanonicalSeries{
		MetricName: "metric",
		Points: []model.CanonicalPoint{
			{Timestamp: base, Value: 10},
			{Timestamp: base.Add(time.Minute), Value: 20},
			{Timestamp: base.Add(4 * time.Minute), Value: 30},
		},
	}

	progress := timelineProgressSeries(primary, "graphite")
	if got := progress.Points[1].Value; got != 25 {
		t.Fatalf("irregular timestamp progress = %v, want 25", got)
	}
}

func TestRangeForValuesAddsPadding(t *testing.T) {
	t.Parallel()

	min, max := rangeForValues([]float64{42, 42, 42})
	if !(min < 42 && max > 42) {
		t.Fatalf("expected padding around identical values, got min=%v max=%v", min, max)
	}

	min, max = rangeForValues([]float64{10, 20, 30})
	if !(min < 10 && max > 30) {
		t.Fatalf("expected padding around range, got min=%v max=%v", min, max)
	}
}

func TestNormalizeSelectedMetricsHonorsPresetSlots(t *testing.T) {
	t.Parallel()

	available := []model.GraphiteMetricOption{
		{Metric: "a"},
		{Metric: "b"},
		{Metric: "c"},
	}
	got := normalizeSelectedMetrics([]string{"a", "a", "b", "c"}, graphitePresetFor("cerberus"), available)
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("unexpected metrics: %#v", got)
	}
}

func TestDiscoverGraphiteMetricOptionsTraversesEveryBranch(t *testing.T) {
	t.Parallel()

	responses := map[string][]graphiteTreeItem{
		"*":                           {{ID: "servers", AllowChildren: 1}, {ID: "stats", Expandable: 1}},
		"servers.*":                   {{ID: "servers.Console-1", AllowChildren: 1}, {ID: "servers.Console-2", AllowChildren: 1}},
		"stats.*":                     {{ID: "stats.response", AllowChildren: 1}},
		"servers.Console-1.*":         {{ID: "servers.Console-1.cpu", AllowChildren: 1}},
		"servers.Console-2.*":         {{ID: "servers.Console-2.network", AllowChildren: 1}},
		"stats.response.*":            {{ID: "stats.response.200", Leaf: 1}},
		"servers.Console-1.cpu.*":     {{ID: "servers.Console-1.cpu.usage_active", Leaf: 1}},
		"servers.Console-2.network.*": {{ID: "servers.Console-2.network.eth0", AllowChildren: 1}},
		"servers.Console-2.network.eth0.*": {
			{ID: "servers.Console-2.network.eth0.bytes_received", Leaf: 1},
		},
	}
	requested := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, _ := url.QueryUnescape(r.URL.Query().Get("query"))
		requested = append(requested, query)
		if err := json.NewEncoder(w).Encode(responses[query]); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer server.Close()

	options, err := discoverGraphiteMetricOptions(t.Context(), server.URL)
	if err != nil {
		t.Fatalf("discover metrics: %v", err)
	}
	got := make([]string, 0, len(options))
	for _, option := range options {
		got = append(got, option.Metric)
	}
	want := []string{"stats.response.200", "servers.Console-2.network.eth0.bytes_received", "servers.Console-1.cpu.usage_active"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("metrics = %#v, want %#v", got, want)
	}
	if !containsString(requested, "servers.Console-2.network.eth0.*") {
		t.Fatalf("network branch was not traversed; queries: %#v", requested)
	}
}

func TestGraphiteMetricOptionIncludesNamespaceForManualMetric(t *testing.T) {
	t.Parallel()
	option := graphiteMetricOption("servers.Console-2.network.eth0.bytes_received")
	if option.Namespace != "servers.Console-2.network.eth0" {
		t.Fatalf("namespace = %q", option.Namespace)
	}
}

func TestNormalizeSelectedMetricsDoesNotFillManualSelection(t *testing.T) {
	t.Parallel()
	available := []model.GraphiteMetricOption{{Metric: "catalog.a"}, {Metric: "catalog.b"}}
	got := normalizeSelectedMetrics([]string{"manual.network.bytes"}, graphitePresetFor("sisyphus"), available)
	if !reflect.DeepEqual(got, []string{"manual.network.bytes"}) {
		t.Fatalf("manual selection changed: %#v", got)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
