package scenes

import (
	"testing"
	"time"

	"infograf/internal/model"
)

func TestNormalizeAndMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		value      float64
		wantMapped float64
	}{
		{name: "min", value: 0, wantMapped: 0.20},
		{name: "quarter", value: 25, wantMapped: 0.35},
		{name: "half", value: 50, wantMapped: 0.50},
		{name: "three_quarters", value: 75, wantMapped: 0.65},
		{name: "max", value: 100, wantMapped: 0.80},
		{name: "below_min", value: -10, wantMapped: 0.20},
		{name: "above_max", value: 120, wantMapped: 0.80},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			normalized, err := Normalize(tc.value, 0, 100)
			if err != nil {
				t.Fatalf("normalize: %v", err)
			}
			got := MapNormalized(normalized, 0.20, 0.80, false)
			if diff := got - tc.wantMapped; diff < -0.0001 || diff > 0.0001 {
				t.Fatalf("mapped = %.4f, want %.4f", got, tc.wantMapped)
			}
		})
	}
}

func TestNormalizeInverseAndInvalidRange(t *testing.T) {
	t.Parallel()

	normalized, err := Normalize(25, 0, 100)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if got := MapNormalized(normalized, 0.20, 0.80, true); got < 0.6499 || got > 0.6501 {
		t.Fatalf("inverse mapped = %.2f, want 0.65", got)
	}
	if _, err := Normalize(1, 5, 5); err == nil {
		t.Fatal("expected invalid range error")
	}
}

func TestEvaluateSceneDeterministicAndEvents(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 8, 30, 10, 20, 0, 0, time.UTC)
	def := model.SceneDefinition{
		ID:       "demo",
		SourceRef: "fixture",
		Bindings: []model.SceneBinding{
			{
				ID:         "replication-lag-separation",
				MetricName: "replication_lag_seconds",
				Interpretation: model.MetricSemantics{
					Type: "gauge",
					Range: model.NumericRange{Min: 0, Max: 120},
				},
				Target:  model.SceneTarget{ObjectID: "replica-db", Property: "x"},
				Mapping: model.SceneMapping{Output: model.NumericRange{Min: 0.60, Max: 0.90}},
			},
			{
				ID:         "error-opacity",
				MetricName: "errors_5m",
				Interpretation: model.MetricSemantics{
					Type: "gauge",
					Range: model.NumericRange{Min: 0, Max: 1},
				},
				Target:  model.SceneTarget{ObjectID: "incident-callout", Property: "opacity"},
				Mapping: model.SceneMapping{Output: model.NumericRange{Min: 0, Max: 1}},
			},
		},
	}
	series := []model.CanonicalSeries{
		{
			MetricName: "replication_lag_seconds",
			Points: []model.CanonicalPoint{
				{Timestamp: base, Value: 0.42},
				{Timestamp: base.Add(time.Minute), Value: 12.87},
				{Timestamp: base.Add(2 * time.Minute), Value: 78.35},
			},
		},
		{
			MetricName: "errors_5m",
			Points: []model.CanonicalPoint{
				{Timestamp: base, Value: 0},
				{Timestamp: base.Add(time.Minute), Value: 0},
				{Timestamp: base.Add(2 * time.Minute), Value: 1},
			},
		},
	}
	events := []model.CanonicalEvent{
		{
			ID:    "INC-REPL-001",
			Type:  "replication_incident",
			Start: base.Add(2 * time.Minute),
			End:   base.Add(4 * time.Minute),
			Title: "Replication lag exceeded recovery threshold",
		},
	}

	at := base.Add(2 * time.Minute)
	first, err := EvaluateScene(def, series, events, at)
	if err != nil {
		t.Fatalf("evaluate first: %v", err)
	}
	second, err := EvaluateScene(def, series, events, at)
	if err != nil {
		t.Fatalf("evaluate second: %v", err)
	}
	if first.Timestamp != second.Timestamp || len(first.Objects) != len(second.Objects) || len(first.ActiveEvents) != 1 {
		t.Fatalf("non-deterministic state: %#v %#v", first, second)
	}
	replicaState := ""
	for _, object := range first.Objects {
		if object.ID == "replica-db" {
			replicaState = object.State
		}
	}
	if replicaState == "" {
		t.Fatalf("expected replica state to be set: %#v", first.Objects)
	}
}
