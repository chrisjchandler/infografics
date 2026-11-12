package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"infograf/internal/model"
	"infograf/internal/scenes"
)

const (
	graphiteMetricCatalogTTL       = 5 * time.Minute
	graphiteMetricCatalogRetryWait = 30 * time.Second
	graphiteMetricRefreshTimeout   = 30 * time.Second
)

type graphiteMetricCatalogCache struct {
	mu          sync.Mutex
	sourceURL   string
	metrics     []model.GraphiteMetricOption
	fetchedAt   time.Time
	nextRefresh time.Time
	lastErr     error
	refreshing  bool
	ready       chan struct{}
}

type graphiteTreeItem struct {
	ID            string `json:"id"`
	Text          string `json:"text"`
	Leaf          int    `json:"leaf"`
	AllowChildren int    `json:"allowChildren"`
	Expandable    int    `json:"expandable"`
}

type graphiteScenePreset struct {
	Key         string
	Label       string
	Description string
	MetricSlots int
}

var graphiteScenePresets = []graphiteScenePreset{
	{Key: "sisyphus", Label: "Sisyphus", Description: "A climber and boulder follow the metric curve. Time is X. Value is Y.", MetricSlots: 6},
	{Key: "icarus", Label: "Icarus", Description: "A flyer rides altitude from the metric value and panics on sharp descents.", MetricSlots: 6},
	{Key: "cerberus", Label: "Cerberus", Description: "Three heads guard selected metrics while the graph gallery compares the full overlay set.", MetricSlots: 6},
}

func (s *Server) handleGraphiteCatalog(w http.ResponseWriter, r *http.Request) {
	catalog, cacheStatus := s.graphiteCatalog(r.Context())
	w.Header().Set("X-Infograf-Metric-Catalog", cacheStatus)
	writeJSON(w, http.StatusOK, catalog)
}

func (s *Server) handleGraphiteScene(w http.ResponseWriter, r *http.Request) {
	period := graphitePeriodFor(strings.TrimSpace(r.URL.Query().Get("period")))
	at, err := optionalTimeParam(r, "at")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid timestamp", err.Error(), "bad_request")
		return
	}
	metrics := selectedGraphiteMetrics(r)
	preset := graphitePresetFor(strings.TrimSpace(r.URL.Query().Get("preset")))
	resp, cacheStatus, err := s.graphiteScene(r.Context(), preset, period, metrics, at)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to build live Graphite scene", err.Error(), "graphite_error")
		return
	}
	w.Header().Set("X-Infograf-Metric-Catalog", cacheStatus)
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) graphiteCatalog(ctx context.Context) (model.GraphiteCatalogResponse, string) {
	sourceURL := graphiteBaseURL()
	metrics, cacheStatus, err := s.graphiteMetricOptions(ctx, sourceURL)
	connected := err == nil
	sourceLabel := "live Graphite"
	if err != nil {
		metrics = []model.GraphiteMetricOption{}
		sourceLabel = fmt.Sprintf("Graphite unavailable (%v)", err)
	}

	return model.GraphiteCatalogResponse{
		Connected:   connected,
		SourceURL:   sourceURL,
		SourceLabel: sourceLabel,
		CapturedAt:  time.Now().UTC(),
		Metrics:     metrics,
		Periods:     graphitePeriodOptions(),
		Presets:     graphiteScenePresetOptions(),
		Connectors:  connectorOptions(),
	}, cacheStatus
}

func (s *Server) graphiteScene(ctx context.Context, preset graphiteScenePreset, period graphitePeriodSpec, selected []string, at *time.Time) (model.GraphiteSceneResponse, string, error) {
	sourceURL := graphiteBaseURL()
	availableMetrics, cacheStatus, discoverErr := s.graphiteMetricOptions(ctx, sourceURL)
	if discoverErr != nil && len(selected) == 0 {
		return model.GraphiteSceneResponse{}, cacheStatus, fmt.Errorf("discover Graphite metrics: %w", discoverErr)
	}
	selected = normalizeSelectedMetrics(selected, preset, availableMetrics)
	if len(selected) == 0 {
		return model.GraphiteSceneResponse{}, cacheStatus, fmt.Errorf("Graphite did not return any leaf metrics")
	}
	specs := metricOptionsToSpecs(selected, availableMetrics)

	seriesByMetric, err := fetchGraphiteSeries(ctx, sourceURL, specs, period.Window)
	if err != nil {
		return model.GraphiteSceneResponse{}, cacheStatus, fmt.Errorf("fetch Graphite series: %w", err)
	}
	if graphiteSeriesSetIncomplete(seriesByMetric, specs) {
		return model.GraphiteSceneResponse{}, cacheStatus, fmt.Errorf("Graphite returned no datapoints for one or more selected metrics")
	}

	canonicalSeries := make([]model.CanonicalSeries, 0, len(specs)+1)
	for _, spec := range specs {
		series := graphiteSeriesToCanonical(spec, seriesByMetric[spec.Metric], sourceURL)
		if len(series.Points) == 0 {
			return model.GraphiteSceneResponse{}, cacheStatus, fmt.Errorf("Graphite metric %q has no numeric datapoints", spec.Metric)
		}
		canonicalSeries = append(canonicalSeries, series)
	}

	primary := canonicalSeries[0]
	timeSeries := timelineProgressSeries(primary, sourceURL)
	canonicalSeries = append(canonicalSeries, timeSeries)
	if at == nil {
		last := primary.Points[len(primary.Points)-1].Timestamp
		at = &last
	}

	sceneDef := graphiteSceneDefinition(preset, canonicalSeries)
	sceneState, err := scenes.EvaluateScene(sceneDef, canonicalSeries, graphiteAnnotationsAsEvents(primary), at.UTC())
	if err != nil {
		return model.GraphiteSceneResponse{}, cacheStatus, err
	}
	frames := make([]model.SceneState, 0, len(primary.Points))
	events := graphiteAnnotationsAsEvents(primary)
	for _, point := range primary.Points {
		frameState, err := scenes.EvaluateScene(sceneDef, canonicalSeries, events, point.Timestamp)
		if err != nil {
			return model.GraphiteSceneResponse{}, cacheStatus, err
		}
		frames = append(frames, frameState)
	}

	return model.GraphiteSceneResponse{
		Connected:        true,
		SourceURL:        sourceURL,
		SourceLabel:      "live Graphite",
		Period:           period.Key,
		PeriodLabel:      period.Label,
		Window:           period.Window,
		SelectedPreset:   preset.Key,
		SelectedMetrics:  selected,
		AvailableMetrics: availableMetrics,
		ScenePresets:     graphiteScenePresetOptions(),
		Connectors:       connectorOptions(),
		Timeline:         timelinePoints(primary),
		Annotations:      graphiteAnnotations(primary),
		Scene:            sceneDef,
		State:            sceneState,
		Frames:           frames,
		Series:           canonicalSeries[:len(canonicalSeries)-1],
	}, cacheStatus, nil
}

func (s *Server) graphiteMetricOptions(ctx context.Context, sourceURL string) ([]model.GraphiteMetricOption, string, error) {
	cache := &s.graphiteMetricCatalog
	for {
		now := time.Now()
		cache.mu.Lock()
		if cache.sourceURL != sourceURL {
			cache.sourceURL = sourceURL
			cache.metrics = nil
			cache.fetchedAt = time.Time{}
			cache.nextRefresh = time.Time{}
			cache.lastErr = nil
		}
		if len(cache.metrics) > 0 && now.Sub(cache.fetchedAt) < graphiteMetricCatalogTTL {
			metrics := cloneGraphiteMetricOptions(cache.metrics)
			cache.mu.Unlock()
			return metrics, "hit", nil
		}
		if len(cache.metrics) > 0 {
			metrics := cloneGraphiteMetricOptions(cache.metrics)
			if !cache.refreshing && !now.Before(cache.nextRefresh) {
				cache.refreshing = true
				cache.ready = make(chan struct{})
				go s.refreshGraphiteMetricOptions(sourceURL)
			}
			cache.mu.Unlock()
			return metrics, "stale", nil
		}
		if cache.refreshing {
			ready := cache.ready
			cache.mu.Unlock()
			select {
			case <-ctx.Done():
				return nil, "miss", ctx.Err()
			case <-ready:
				continue
			}
		}
		if cache.lastErr != nil && now.Before(cache.nextRefresh) {
			err := cache.lastErr
			cache.mu.Unlock()
			return nil, "error", err
		}
		cache.refreshing = true
		cache.ready = make(chan struct{})
		cache.mu.Unlock()

		metrics, err := discoverGraphiteMetricOptions(ctx, sourceURL)
		s.finishGraphiteMetricRefresh(sourceURL, metrics, err)
		if err != nil {
			return nil, "miss", err
		}
		return cloneGraphiteMetricOptions(metrics), "miss", nil
	}
}

func (s *Server) refreshGraphiteMetricOptions(sourceURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), graphiteMetricRefreshTimeout)
	defer cancel()
	metrics, err := discoverGraphiteMetricOptions(ctx, sourceURL)
	s.finishGraphiteMetricRefresh(sourceURL, metrics, err)
}

func (s *Server) finishGraphiteMetricRefresh(sourceURL string, metrics []model.GraphiteMetricOption, err error) {
	cache := &s.graphiteMetricCatalog
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.sourceURL == sourceURL {
		cache.lastErr = err
		if err == nil {
			cache.metrics = cloneGraphiteMetricOptions(metrics)
			cache.fetchedAt = time.Now()
			cache.nextRefresh = cache.fetchedAt.Add(graphiteMetricCatalogTTL)
		} else {
			cache.nextRefresh = time.Now().Add(graphiteMetricCatalogRetryWait)
		}
	}
	cache.refreshing = false
	if cache.ready != nil {
		close(cache.ready)
		cache.ready = nil
	}
}

func cloneGraphiteMetricOptions(metrics []model.GraphiteMetricOption) []model.GraphiteMetricOption {
	return append([]model.GraphiteMetricOption(nil), metrics...)
}

func graphiteSeriesSetIncomplete(series map[string]graphiteSeries, specs []graphiteMetricSpec) bool {
	for _, spec := range specs {
		if len(series[spec.Metric].Datapoints) == 0 {
			return true
		}
	}
	return false
}

func selectedGraphiteMetrics(r *http.Request) []string {
	keys := []string{"metric", "metric_b", "metric_c", "metric_d", "metric_e", "metric_f"}
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		if value := strings.TrimSpace(r.URL.Query().Get(key)); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func graphitePresetFor(raw string) graphiteScenePreset {
	for _, preset := range graphiteScenePresets {
		if preset.Key == strings.ToLower(strings.TrimSpace(raw)) {
			return preset
		}
	}
	return graphiteScenePresets[0]
}

func graphiteScenePresetOptions() []model.ScenePresetOption {
	out := make([]model.ScenePresetOption, 0, len(graphiteScenePresets))
	for _, preset := range graphiteScenePresets {
		out = append(out, model.ScenePresetOption{
			Key:         preset.Key,
			Label:       preset.Label,
			Description: preset.Description,
			MetricSlots: preset.MetricSlots,
		})
	}
	return out
}

func connectorOptions() []model.SourceConnectorOption {
	return []model.SourceConnectorOption{
		{Key: "graphite", Label: "Graphite", Status: "available", Description: "Live render and metric discovery are wired today."},
		{Key: "prometheus", Label: "Prometheus", Status: "available", Description: "Text exposition is normalized into canonical timestamped series."},
		{Key: "csv", Label: "CSV", Status: "available", Description: "RFC4180 uploads with metric, value, and timestamp columns."},
		{Key: "xlsx", Label: "XLSX worksheet", Status: "available", Description: "Worksheet-shaped CSV bytes share the CSV canonical contract."},
		{Key: "opentelemetry", Label: "OpenTelemetry", Status: "available", Description: "OTLP JSON gauges, sums, histograms, and data points."},
		{Key: "rest", Label: "REST / JSON", Status: "available", Description: "Nested metric/name + value API payloads map to canonical series."},
		{Key: "plaintext", Label: "Raw plaintext", Status: "available", Description: "metric value timestamp lines for simple probes and scripts."},
	}
}

func discoverGraphiteMetricOptions(ctx context.Context, sourceURL string) ([]model.GraphiteMetricOption, error) {
	if metrics, err := fetchGraphiteMetricIndex(ctx, sourceURL); err == nil && len(metrics) > 0 {
		out := make([]model.GraphiteMetricOption, 0, len(metrics))
		seen := make(map[string]bool, len(metrics))
		for _, metric := range metrics {
			metric = strings.TrimSpace(metric)
			if metric == "" || seen[metric] {
				continue
			}
			seen[metric] = true
			out = append(out, graphiteMetricOption(metric))
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
		return out, nil
	}

	queries := []string{"*"}
	visited := map[string]bool{}
	seen := map[string]model.GraphiteMetricOption{}
	for len(queries) > 0 {
		query := queries[0]
		queries = queries[1:]
		if visited[query] {
			continue
		}
		visited[query] = true
		items, err := fetchGraphiteMetricTree(ctx, sourceURL, query)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			metric := strings.TrimSpace(item.ID)
			if metric == "" {
				continue
			}
			if item.Leaf != 0 {
				option := graphiteMetricOption(metric)
				seen[option.Metric] = option
			}
			if item.Leaf == 0 || item.AllowChildren != 0 || item.Expandable != 0 {
				queries = append(queries, metric+".*")
			}
		}
	}

	out := make([]model.GraphiteMetricOption, 0, len(seen))
	for _, option := range seen {
		out = append(out, option)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out, nil
}

func fetchGraphiteMetricIndex(ctx context.Context, sourceURL string) ([]string, error) {
	path := fmt.Sprintf("%s/metrics/index.json", sourceURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := graphiteHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graphite metric index returned %s", resp.Status)
	}
	var metrics []string
	if err := decodeJSON(resp.Body, &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}

func fetchGraphiteMetricTree(ctx context.Context, sourceURL, query string) ([]graphiteTreeItem, error) {
	path := fmt.Sprintf("%s/metrics/find?format=treejson&query=%s", sourceURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := graphiteHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("graphite metric tree returned %s", resp.Status)
	}
	var items []graphiteTreeItem
	if err := decodeJSON(resp.Body, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func normalizeSelectedMetrics(selected []string, preset graphiteScenePreset, available []model.GraphiteMetricOption) []string {
	seen := map[string]bool{}
	out := make([]string, 0, preset.MetricSlots)
	for _, metric := range selected {
		metric = strings.TrimSpace(metric)
		if metric == "" || seen[metric] {
			continue
		}
		out = append(out, metric)
		seen[metric] = true
		if len(out) == preset.MetricSlots {
			break
		}
	}
	if len(out) == 0 && len(available) > 0 {
		out = append(out, available[0].Metric)
	}
	return out
}

func metricOptionsToSpecs(metrics []string, available []model.GraphiteMetricOption) []graphiteMetricSpec {
	index := map[string]model.GraphiteMetricOption{}
	for _, option := range available {
		index[option.Metric] = option
	}
	out := make([]graphiteMetricSpec, 0, len(metrics))
	for _, metric := range metrics {
		option, ok := index[metric]
		if !ok {
			option = graphiteMetricOption(metric)
		}
		out = append(out, graphiteMetricSpec{
			Label:            option.Label,
			Metric:           option.Metric,
			Unit:             option.Unit,
			Detail:           "Live Graphite series",
			HigherIsBetter:   option.PreferredDirection != "lower",
			HealthyThreshold: 0,
			WarningThreshold: 0,
			ProgressCap:      1,
		})
	}
	return out
}

func graphiteMetricOption(metric string) model.GraphiteMetricOption {
	label := metric
	if parts := strings.Split(metric, "."); len(parts) > 0 {
		label = strings.ReplaceAll(parts[len(parts)-1], "_", " ")
		label = strings.Title(label)
	}
	unit := "value"
	preferred := "higher"
	if strings.Contains(metric, "percent") || strings.Contains(metric, "usage_") {
		unit = "%"
	}
	if strings.Contains(metric, "idle") || strings.Contains(metric, "free") {
		preferred = "higher"
	} else if strings.Contains(metric, "used") || strings.Contains(metric, "404") || strings.Contains(metric, "400") || strings.Contains(metric, "error") {
		preferred = "lower"
	}
	if strings.Contains(metric, "metrics_received") || strings.Contains(metric, "packets_received") || strings.Contains(metric, "response.") {
		unit = "req/s"
	}
	return model.GraphiteMetricOption{
		Metric:             metric,
		Namespace:          graphiteMetricNamespace(metric),
		Label:              label,
		Unit:               unit,
		PreferredDirection: preferred,
	}
}

func graphiteMetricNamespace(metric string) string {
	metric = strings.TrimSpace(metric)
	if split := strings.LastIndex(metric, "."); split > 0 {
		return metric[:split]
	}
	return ""
}

func preferredDirection(spec graphiteMetricSpec) string {
	if spec.HigherIsBetter {
		return "higher"
	}
	return "lower"
}

func graphiteSeriesToCanonical(spec graphiteMetricSpec, series graphiteSeries, sourceURL string) model.CanonicalSeries {
	points := make([]model.CanonicalPoint, 0, len(series.Datapoints))
	values := make([]float64, 0, len(series.Datapoints))
	for _, dp := range series.Datapoints {
		if len(dp) < 2 || dp[0] == nil {
			continue
		}
		value, ok := toFloat(dp[0])
		if !ok {
			continue
		}
		seconds, ok := toFloat(dp[1])
		if !ok {
			continue
		}
		ts := time.Unix(int64(seconds), 0).UTC()
		points = append(points, model.CanonicalPoint{Timestamp: ts, Value: value})
		values = append(values, value)
	}
	min, max := rangeForValues(values)
	return model.CanonicalSeries{
		ID:         strings.ReplaceAll(spec.Metric, ".", "_"),
		SourceRef:  sourceURL,
		MetricName: spec.Metric,
		Label:      spec.Label,
		Semantics: model.MetricSemantics{
			Type:               "gauge",
			PreferredDirection: preferredDirection(spec),
			Range:              model.NumericRange{Min: min, Max: max},
			Unit:               spec.Unit,
		},
		Points: points,
	}
}

func timelineProgressSeries(primary model.CanonicalSeries, sourceURL string) model.CanonicalSeries {
	points := make([]model.CanonicalPoint, 0, len(primary.Points))
	if len(primary.Points) == 1 {
		points = append(points, model.CanonicalPoint{Timestamp: primary.Points[0].Timestamp, Value: 100})
	} else if len(primary.Points) > 1 {
		start := primary.Points[0].Timestamp
		end := primary.Points[len(primary.Points)-1].Timestamp
		span := end.Sub(start).Seconds()
		for idx, point := range primary.Points {
			progress := 0.0
			if span > 0 {
				progress = point.Timestamp.Sub(start).Seconds() / span * 100
			} else {
				// Invalid or unordered timestamps still produce deterministic output;
				// upstream validation can report the malformed series.
				progress = float64(idx) / float64(len(primary.Points)-1) * 100
			}
			points = append(points, model.CanonicalPoint{Timestamp: point.Timestamp, Value: progress})
		}
	}
	return model.CanonicalSeries{
		ID:         "timeline_progress",
		SourceRef:  sourceURL,
		MetricName: "timeline_progress",
		Label:      "Timeline progress",
		Semantics: model.MetricSemantics{
			Type:               "progress",
			PreferredDirection: "higher",
			Range:              model.NumericRange{Min: 0, Max: 100},
			Unit:               "%",
		},
		Points: points,
	}
}

func graphiteSceneDefinition(preset graphiteScenePreset, series []model.CanonicalSeries) model.SceneDefinition {
	main := series[0]
	timeMetric := "timeline_progress"
	description := preset.Description + " The right rail shows annotations and the scene can be scrubbed or played over the source timestamps."
	bindings := []model.SceneBinding{
		{
			ID:         preset.Key + "-timeline-x",
			SourceRef:  main.SourceRef,
			MetricName: timeMetric,
			Interpretation: model.MetricSemantics{
				Type:               "progress",
				PreferredDirection: "higher",
				Range:              model.NumericRange{Min: 0, Max: 100},
				Unit:               "%",
			},
			Target:    model.SceneTarget{ObjectID: "focus", Property: "x"},
			Mapping:   model.SceneMapping{Direction: "direct", Output: model.NumericRange{Min: 0.1, Max: 0.9}},
			Animation: model.AnimationSpec{DurationMS: 280, Easing: "linear"},
		},
		{
			ID:             preset.Key + "-value-y",
			SourceRef:      main.SourceRef,
			MetricName:     main.MetricName,
			Interpretation: main.Semantics,
			Target:         model.SceneTarget{ObjectID: "focus", Property: "y"},
			Mapping:        model.SceneMapping{Direction: "inverse", Output: model.NumericRange{Min: 0.14, Max: 0.86}},
			Animation:      model.AnimationSpec{DurationMS: 420, Easing: "ease-out"},
		},
		{
			ID:             preset.Key + "-value-scale",
			SourceRef:      main.SourceRef,
			MetricName:     main.MetricName,
			Interpretation: main.Semantics,
			Target:         model.SceneTarget{ObjectID: "focus", Property: "scale"},
			Mapping:        model.SceneMapping{Direction: "direct", Output: model.NumericRange{Min: 0.88, Max: 1.22}},
			Animation:      model.AnimationSpec{DurationMS: 420, Easing: "ease-out"},
		},
	}

	switch preset.Key {
	case "icarus":
		bindings = append(bindings, model.SceneBinding{
			ID:             "icarus-tilt",
			SourceRef:      main.SourceRef,
			MetricName:     main.MetricName,
			Interpretation: main.Semantics,
			Target:         model.SceneTarget{ObjectID: "focus", Property: "rotation"},
			Mapping:        model.SceneMapping{Direction: "direct", Output: model.NumericRange{Min: -26, Max: 12}},
			Animation:      model.AnimationSpec{DurationMS: 240, Easing: "ease-out"},
		})
	case "cerberus":
		for idx := 0; idx < len(series)-1 && idx < 3; idx++ {
			item := series[idx]
			objectID := fmt.Sprintf("head-%d", idx+1)
			bindings = append(bindings,
				model.SceneBinding{
					ID:             objectID + "-y",
					SourceRef:      item.SourceRef,
					MetricName:     item.MetricName,
					Interpretation: item.Semantics,
					Target:         model.SceneTarget{ObjectID: objectID, Property: "y"},
					Mapping:        model.SceneMapping{Direction: "inverse", Output: model.NumericRange{Min: 0.18, Max: 0.86}},
					Animation:      model.AnimationSpec{DurationMS: 400, Easing: "ease-out"},
				},
				model.SceneBinding{
					ID:             objectID + "-scale",
					SourceRef:      item.SourceRef,
					MetricName:     item.MetricName,
					Interpretation: item.Semantics,
					Target:         model.SceneTarget{ObjectID: objectID, Property: "scale"},
					Mapping:        model.SceneMapping{Direction: "direct", Output: model.NumericRange{Min: 0.84, Max: 1.16}},
					Animation:      model.AnimationSpec{DurationMS: 400, Easing: "ease-out"},
				},
			)
		}
	}

	return model.SceneDefinition{
		ID:          "scene_graphite_" + preset.Key,
		Title:       preset.Label + " live metric playback",
		Description: description,
		SourceRef:   main.SourceRef,
		Bindings:    bindings,
	}
}

func timelinePoints(series model.CanonicalSeries) []model.TimelinePoint {
	out := make([]model.TimelinePoint, 0, len(series.Points))
	for _, point := range series.Points {
		out = append(out, model.TimelinePoint{
			Timestamp: point.Timestamp,
			Label:     point.Timestamp.Format("15:04"),
			Value:     point.Value,
		})
	}
	return out
}

func graphiteAnnotations(series model.CanonicalSeries) []model.SceneAnnotation {
	if len(series.Points) == 0 {
		return nil
	}
	maxIdx, minIdx := 0, 0
	steepIdx := 0
	steepDelta := 0.0
	for idx := 1; idx < len(series.Points); idx++ {
		if series.Points[idx].Value > series.Points[maxIdx].Value {
			maxIdx = idx
		}
		if series.Points[idx].Value < series.Points[minIdx].Value {
			minIdx = idx
		}
		delta := series.Points[idx].Value - series.Points[idx-1].Value
		if abs(delta) > abs(steepDelta) {
			steepDelta = delta
			steepIdx = idx
		}
	}
	current := series.Points[len(series.Points)-1]
	annotations := []model.SceneAnnotation{
		{ID: "current", Timestamp: current.Timestamp, Title: "Current sample", Detail: fmt.Sprintf("%.2f %s at %s", current.Value, series.Semantics.Unit, current.Timestamp.Format(time.RFC3339)), Severity: "info"},
		{ID: "peak", Timestamp: series.Points[maxIdx].Timestamp, Title: "Peak value", Detail: fmt.Sprintf("%.2f %s", series.Points[maxIdx].Value, series.Semantics.Unit), Severity: "warning"},
		{ID: "trough", Timestamp: series.Points[minIdx].Timestamp, Title: "Lowest value", Detail: fmt.Sprintf("%.2f %s", series.Points[minIdx].Value, series.Semantics.Unit), Severity: "healthy"},
	}
	if steepIdx > 0 {
		severity := "info"
		title := "Sharp rise"
		if steepDelta < 0 {
			title = "Sharp drop"
			severity = "critical"
		}
		annotations = append(annotations, model.SceneAnnotation{
			ID:        "slope",
			Timestamp: series.Points[steepIdx].Timestamp,
			Title:     title,
			Detail:    fmt.Sprintf("%+.2f %s between %s and %s", steepDelta, series.Semantics.Unit, series.Points[steepIdx-1].Timestamp.Format("15:04"), series.Points[steepIdx].Timestamp.Format("15:04")),
			Severity:  severity,
		})
	}
	sort.Slice(annotations, func(i, j int) bool { return annotations[i].Timestamp.Before(annotations[j].Timestamp) })
	return annotations
}

func graphiteAnnotationsAsEvents(series model.CanonicalSeries) []model.CanonicalEvent {
	annotations := graphiteAnnotations(series)
	out := make([]model.CanonicalEvent, 0, len(annotations))
	for _, note := range annotations {
		out = append(out, model.CanonicalEvent{
			ID:    note.ID,
			Type:  "annotation",
			Start: note.Timestamp,
			End:   note.Timestamp.Add(time.Minute),
			Title: note.Title,
			Attributes: map[string]string{
				"detail":   note.Detail,
				"severity": note.Severity,
			},
		})
	}
	return out
}

func rangeForValues(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 1
	}
	min, max := values[0], values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	if min == max {
		if min == 0 {
			return 0, 1
		}
		padding := abs(min) * 0.1
		if padding == 0 {
			padding = 1
		}
		return min - padding, max + padding
	}
	padding := (max - min) * 0.08
	return min - padding, max + padding
}

func toFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	}
	return 0, false
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
