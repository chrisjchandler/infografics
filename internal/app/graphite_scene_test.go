package app

import (
	"testing"
	"time"

	"infograf/internal/model"
)

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

func TestFallbackGraphiteSeriesIsDeterministic(t *testing.T) {
	t.Parallel()

	specs := []graphiteMetricSpec{{Metric: "demo.metric"}}
	first := fallbackGraphiteSeries(specs)["demo.metric"]
	second := fallbackGraphiteSeries(specs)["demo.metric"]
	if len(first.Datapoints) != 9 || len(second.Datapoints) != 9 {
		t.Fatalf("unexpected fallback lengths: %d and %d", len(first.Datapoints), len(second.Datapoints))
	}
	if first.Datapoints[4][0] != second.Datapoints[4][0] || first.Datapoints[4][1] != second.Datapoints[4][1] {
		t.Fatalf("fallback changed between calls: %#v %#v", first, second)
	}
}
