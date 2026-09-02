package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"infograf/internal/model"
)

const graphiteDefaultURL = "http://172.16.0.11:8080"

var graphiteHTTPClient = &http.Client{Timeout: 6 * time.Second}

type graphiteMetricSpec struct {
	Label            string
	Metric           string
	Unit             string
	Detail           string
	HigherIsBetter   bool
	HealthyThreshold float64
	WarningThreshold float64
	ProgressCap      float64
}

type graphiteSeries struct {
	Target     string  `json:"target"`
	Datapoints [][]any `json:"datapoints"`
}

type graphitePeriodSpec struct {
	Key    string
	Label  string
	Window string
}

var graphitePeriods = []graphitePeriodSpec{
	{Key: "1h", Label: "Last hour", Window: "-1h"},
	{Key: "6h", Label: "Last 6 hours", Window: "-6h"},
	{Key: "24h", Label: "Last 24 hours", Window: "-24h"},
	{Key: "7d", Label: "Last 7 days", Window: "-7d"},
}

func (s *Server) handleDemoGraphite(w http.ResponseWriter, r *http.Request) {
	period := graphitePeriodFor(strings.TrimSpace(r.URL.Query().Get("period")))
	writeJSON(w, http.StatusOK, graphiteOverview(r.Context(), period))
}

func graphiteOverview(ctx context.Context, period graphitePeriodSpec) model.GraphiteOverview {
	sourceURL := graphiteBaseURL()
	specs := graphiteMetricSpecs()
	seriesByMetric, err := fetchGraphiteSeries(ctx, sourceURL, specs, period.Window)
	if err != nil {
		return fallbackGraphiteOverview(sourceURL, period, err)
	}

	cards := make([]model.GraphiteCard, 0, len(specs))
	for _, spec := range specs {
		series := seriesByMetric[spec.Metric]
		cards = append(cards, buildGraphiteCard(spec, series, true))
	}

	return model.GraphiteOverview{
		Connected:   true,
		SourceURL:   sourceURL,
		SourceLabel: "live Graphite",
		Period:      period.Key,
		PeriodLabel: period.Label,
		Window:      period.Window,
		CapturedAt:  time.Now().UTC(),
		Cards:       cards,
		Periods:     graphitePeriodOptions(),
		Notes: []string{
			"The scene reads the live source through a period window so the story can shift between quick glances and long arcs.",
			"Longer windows are useful for comedy because the hill, cloud, or rope gets its shape from the trend.",
		},
	}
}

func fallbackGraphiteOverview(sourceURL string, period graphitePeriodSpec, err error) model.GraphiteOverview {
	specs := graphiteMetricSpecs()
	seed := []float64{47.2, 61.4, 7.8, 3.9, 0.1}
	cards := make([]model.GraphiteCard, 0, len(specs))
	for i, spec := range specs {
		cards = append(cards, buildFallbackCard(spec, seed[i%len(seed)]))
	}
	msg := "demo fallback values"
	if err != nil {
		msg = fmt.Sprintf("demo fallback values (%v)", err)
	}
	return model.GraphiteOverview{
		Connected:   false,
		SourceURL:   sourceURL,
		SourceLabel: msg,
		Period:      period.Key,
		PeriodLabel: period.Label,
		Window:      period.Window,
		CapturedAt:  time.Now().UTC(),
		Cards:       cards,
		Periods:     graphitePeriodOptions(),
		Notes: []string{
			"The Graphite host was unreachable, so the scene falls back to seeded values.",
			"Switching the period still works even in fallback mode, which is useful for demo continuity.",
		},
	}
}

func graphiteBaseURL() string {
	if v := strings.TrimSpace(os.Getenv("INFOGRAF_GRAPHITE_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return graphiteDefaultURL
}

func graphitePeriodFor(raw string) graphitePeriodSpec {
	raw = strings.TrimSpace(strings.ToLower(raw))
	for _, period := range graphitePeriods {
		if period.Key == raw {
			return period
		}
	}
	return graphitePeriods[2]
}

func graphitePeriodOptions() []model.GraphitePeriodOption {
	out := make([]model.GraphitePeriodOption, 0, len(graphitePeriods))
	for _, period := range graphitePeriods {
		out = append(out, model.GraphitePeriodOption{Key: period.Key, Label: period.Label, Window: period.Window})
	}
	return out
}

func graphiteMetricSpecs() []graphiteMetricSpec {
	return []graphiteMetricSpec{
		{
			Label:            "CPU active",
			Metric:           "servers.Console-1.cpu.usage_active",
			Unit:             "%",
			Detail:           "Console-1 over the selected period",
			HigherIsBetter:   false,
			HealthyThreshold: 55,
			WarningThreshold: 80,
			ProgressCap:      100,
		},
		{
			Label:            "Disk used",
			Metric:           "servers.Console-1.disk.mmcblk0p2.used_percent",
			Unit:             "%",
			Detail:           "System volume usage",
			HigherIsBetter:   false,
			HealthyThreshold: 70,
			WarningThreshold: 85,
			ProgressCap:      100,
		},
		{
			Label:            "Metrics received",
			Metric:           "stats.statsd.metrics_received",
			Unit:             "req/s",
			Detail:           "StatsD ingest activity",
			HigherIsBetter:   true,
			HealthyThreshold: 5,
			WarningThreshold: 2,
			ProgressCap:      10,
		},
		{
			Label:            "HTTP 200 rate",
			Metric:           "stats.response.200",
			Unit:             "req/s",
			Detail:           "Successful response throughput",
			HigherIsBetter:   true,
			HealthyThreshold: 1,
			WarningThreshold: 0.25,
			ProgressCap:      10,
		},
		{
			Label:            "HTTP 404 rate",
			Metric:           "stats.response.404",
			Unit:             "req/s",
			Detail:           "Client-side misses and unknown routes",
			HigherIsBetter:   false,
			HealthyThreshold: 0.5,
			WarningThreshold: 1.5,
			ProgressCap:      5,
		},
	}
}

func fetchGraphiteSeries(ctx context.Context, baseURL string, specs []graphiteMetricSpec, window string) (map[string]graphiteSeries, error) {
	values := url.Values{}
	values.Set("format", "json")
	values.Set("from", window)
	for _, spec := range specs {
		values.Add("target", spec.Metric)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/render?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := graphiteHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graphite render returned %s", resp.Status)
	}

	var raw []graphiteSeries
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make(map[string]graphiteSeries, len(raw))
	for _, series := range raw {
		out[series.Target] = series
	}
	return out, nil
}

func buildGraphiteCard(spec graphiteMetricSpec, series graphiteSeries, live bool) model.GraphiteCard {
	value, prev, samples, ok := summarizeGraphiteSeries(series.Datapoints)
	if !ok {
		return model.GraphiteCard{
			Label:        spec.Label,
			Metric:       spec.Metric,
			ValueDisplay: "n/a",
			DeltaDisplay: "no data",
			Detail:       spec.Detail,
			Status:       "neutral",
			Progress:     0,
			Live:         live,
		}
	}

	status := graphiteStatus(spec, value)
	progress := int((value / spec.ProgressCap) * 100)
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	return model.GraphiteCard{
		Label:        spec.Label,
		Metric:       spec.Metric,
		ValueDisplay: formatGraphiteValue(spec.Unit, value),
		DeltaDisplay: formatGraphiteDelta(spec.Unit, value, prev),
		Detail:       fmt.Sprintf("%s · %d datapoints", spec.Detail, len(samples)),
		Status:       status,
		Progress:     progress,
		Live:         live,
		Points:       samples,
	}
}

func buildFallbackCard(spec graphiteMetricSpec, value float64) model.GraphiteCard {
	status := graphiteStatus(spec, value)
	progress := int((value / spec.ProgressCap) * 100)
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	return model.GraphiteCard{
		Label:        spec.Label,
		Metric:       spec.Metric,
		ValueDisplay: formatGraphiteValue(spec.Unit, value),
		DeltaDisplay: "demo seed",
		Detail:       fmt.Sprintf("%s · seeded fallback", spec.Detail),
		Status:       status,
		Progress:     progress,
		Live:         false,
		Points:       []float64{value},
	}
}

func summarizeGraphiteSeries(datapoints [][]any) (value float64, prev float64, samples []float64, ok bool) {
	values := make([]float64, 0, len(datapoints))
	for _, dp := range datapoints {
		if len(dp) == 0 || dp[0] == nil {
			continue
		}
		switch v := dp[0].(type) {
		case float64:
			values = append(values, v)
		case int:
			values = append(values, float64(v))
		case json.Number:
			if parsed, err := v.Float64(); err == nil {
				values = append(values, parsed)
			}
		}
	}
	if len(values) == 0 {
		return 0, 0, nil, false
	}
	value = values[len(values)-1]
	if len(values) > 1 {
		prev = values[len(values)-2]
	}
	return value, prev, sampleSeries(values, 22), true
}

func sampleSeries(values []float64, maxPoints int) []float64 {
	if len(values) <= maxPoints {
		return append([]float64(nil), values...)
	}
	if maxPoints < 2 {
		return []float64{values[len(values)-1]}
	}
	out := make([]float64, 0, maxPoints)
	stride := float64(len(values)-1) / float64(maxPoints-1)
	for i := 0; i < maxPoints; i++ {
		idx := int(math.Round(float64(i) * stride))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(values) {
			idx = len(values) - 1
		}
		out = append(out, values[idx])
	}
	return out
}

func graphiteStatus(spec graphiteMetricSpec, value float64) string {
	if spec.HigherIsBetter {
		switch {
		case value >= spec.HealthyThreshold:
			return "healthy"
		case value >= spec.WarningThreshold:
			return "warning"
		default:
			return "critical"
		}
	}
	switch {
	case value <= spec.HealthyThreshold:
		return "healthy"
	case value <= spec.WarningThreshold:
		return "warning"
	default:
		return "critical"
	}
}

func formatGraphiteValue(unit string, value float64) string {
	switch unit {
	case "%":
		return fmt.Sprintf("%.1f%%", value)
	case "req/s":
		return fmt.Sprintf("%.1f req/s", value)
	default:
		return fmt.Sprintf("%.1f %s", value, unit)
	}
}

func formatGraphiteDelta(unit string, value float64, prev float64) string {
	delta := value - prev
	if prev == 0 && value == 0 {
		return "flat"
	}
	if delta == 0 {
		return "flat"
	}
	if delta > 0 {
		return fmt.Sprintf("▲ +%.1f %s", delta, deltaUnit(unit))
	}
	return fmt.Sprintf("▼ %.1f %s", delta, deltaUnit(unit))
}

func deltaUnit(unit string) string {
	switch unit {
	case "%":
		return "pts"
	case "req/s":
		return "rps"
	default:
		return unit
	}
}
