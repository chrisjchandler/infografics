package demo

import (
	"encoding/json"
	"fmt"
	"time"

	"infograf/internal/model"
	"infograf/internal/scenes"
)

const replicationSourceRef = "fixture-db-replication"

func ReplicationSceneDefinition() model.SceneDefinition {
	return model.SceneDefinition{
		ID:          "scene_replication_demo",
		Title:       "Database replication health",
		Description: "Deterministic replication scene driven by canonical fixture data and bindings.",
		SourceRef:   replicationSourceRef,
		Bindings: []model.SceneBinding{
			{
				ID:         "replication-lag-separation",
				SourceRef:  replicationSourceRef,
				MetricName: "replication_lag_seconds",
				Interpretation: model.MetricSemantics{
					Type:               "gauge",
					PreferredDirection: "lower",
					Range:              model.NumericRange{Min: 0, Max: 120},
					Unit:               "s",
				},
				Target:    model.SceneTarget{ObjectID: "replica-db", Property: "x"},
				Mapping:   model.SceneMapping{Direction: "direct", Output: model.NumericRange{Min: 0.58, Max: 0.90}},
				Animation: model.AnimationSpec{DurationMS: 500, Easing: "ease-out"},
			},
			{
				ID:         "replication-lag-altitude",
				SourceRef:  replicationSourceRef,
				MetricName: "replication_lag_seconds",
				Interpretation: model.MetricSemantics{
					Type:               "gauge",
					PreferredDirection: "lower",
					Range:              model.NumericRange{Min: 0, Max: 120},
					Unit:               "s",
				},
				Target:    model.SceneTarget{ObjectID: "replica-db", Property: "y"},
				Mapping:   model.SceneMapping{Direction: "inverse", Output: model.NumericRange{Min: 0.30, Max: 0.76}},
				Animation: model.AnimationSpec{DurationMS: 500, Easing: "ease-out"},
			},
			{
				ID:         "replication-rate-packets",
				SourceRef:  replicationSourceRef,
				MetricName: "replication_rate_mb_s",
				Interpretation: model.MetricSemantics{
					Type:               "gauge",
					PreferredDirection: "higher",
					Range:              model.NumericRange{Min: 0, Max: 16},
					Unit:               "MB/s",
				},
				Target:    model.SceneTarget{ObjectID: "packet-stream", Property: "animation_speed"},
				Mapping:   model.SceneMapping{Direction: "direct", Output: model.NumericRange{Min: 0.2, Max: 1}},
				Animation: model.AnimationSpec{DurationMS: 400, Easing: "linear"},
			},
			{
				ID:         "write-rate-pulse",
				SourceRef:  replicationSourceRef,
				MetricName: "write_tps",
				Interpretation: model.MetricSemantics{
					Type:               "gauge",
					PreferredDirection: "higher",
					Range:              model.NumericRange{Min: 0, Max: 1500},
					Unit:               "TPS",
				},
				Target:    model.SceneTarget{ObjectID: "primary-db", Property: "scale"},
				Mapping:   model.SceneMapping{Direction: "direct", Output: model.NumericRange{Min: 0.96, Max: 1.08}},
				Animation: model.AnimationSpec{DurationMS: 400, Easing: "ease-out"},
			},
			{
				ID:         "last-commit-age-tilt",
				SourceRef:  replicationSourceRef,
				MetricName: "last_commit_age_seconds",
				Interpretation: model.MetricSemantics{
					Type:               "gauge",
					PreferredDirection: "lower",
					Range:              model.NumericRange{Min: 0, Max: 180},
					Unit:               "s",
				},
				Target:    model.SceneTarget{ObjectID: "replica-db", Property: "rotation"},
				Mapping:   model.SceneMapping{Direction: "direct", Output: model.NumericRange{Min: -4, Max: 8}},
				Animation: model.AnimationSpec{DurationMS: 500, Easing: "ease-out"},
			},
			{
				ID:         "error-opacity",
				SourceRef:  replicationSourceRef,
				MetricName: "errors_5m",
				Interpretation: model.MetricSemantics{
					Type:               "gauge",
					PreferredDirection: "lower",
					Range:              model.NumericRange{Min: 0, Max: 1},
					Unit:               "count",
				},
				Target:    model.SceneTarget{ObjectID: "incident-callout", Property: "opacity"},
				Mapping:   model.SceneMapping{Direction: "direct", Output: model.NumericRange{Min: 0, Max: 1}},
				Animation: model.AnimationSpec{DurationMS: 200, Easing: "ease-out"},
			},
		},
	}
}

func ReplicationFixture() model.ReplicationFixture {
	points := []model.ReplicationFixturePoint{
		{Timestamp: mustParseRFC3339("2026-08-30T10:20:00Z"), ReplicationLagSeconds: 0.42, ReplicationRateMBS: 12.4, WriteTPS: 1254, LastCommitAgeSeconds: 2, Errors5M: 0},
		{Timestamp: mustParseRFC3339("2026-08-30T10:21:00Z"), ReplicationLagSeconds: 12.87, ReplicationRateMBS: 8.1, WriteTPS: 1020, LastCommitAgeSeconds: 22, Errors5M: 0},
		{Timestamp: mustParseRFC3339("2026-08-30T10:22:00Z"), ReplicationLagSeconds: 78.35, ReplicationRateMBS: 3.2, WriteTPS: 320, LastCommitAgeSeconds: 150, Errors5M: 1},
		{Timestamp: mustParseRFC3339("2026-08-30T10:23:00Z"), ReplicationLagSeconds: 55.12, ReplicationRateMBS: 5.7, WriteTPS: 540, LastCommitAgeSeconds: 95, Errors5M: 0},
		{Timestamp: mustParseRFC3339("2026-08-30T10:24:00Z"), ReplicationLagSeconds: 2.31, ReplicationRateMBS: 11.0, WriteTPS: 1201, LastCommitAgeSeconds: 3, Errors5M: 0},
		{Timestamp: mustParseRFC3339("2026-08-30T10:25:00Z"), ReplicationLagSeconds: 0.38, ReplicationRateMBS: 12.6, WriteTPS: 1310, LastCommitAgeSeconds: 2, Errors5M: 0},
	}
	return model.ReplicationFixture{
		SourceRef: replicationSourceRef,
		Points:    points,
		Events: []model.CanonicalEvent{
			{
				ID:    "INC-REPL-001",
				Type:  "replication_incident",
				Start: mustParseRFC3339("2026-08-30T10:22:00Z"),
				End:   mustParseRFC3339("2026-08-30T10:24:00Z"),
				Title: "Replication lag exceeded recovery threshold",
				Attributes: map[string]string{
					"ticket":   "INC-REPL-001",
					"severity": "SEV2",
				},
			},
		},
	}
}

func ReplicationDemo(at *time.Time, overrides model.ReplicationOverrides) (model.ReplicationDemoResponse, error) {
	fixture := ReplicationFixture()
	def := ReplicationSceneDefinition()
	if at == nil {
		last := fixture.Points[len(fixture.Points)-1].Timestamp
		at = &last
	}
	series := replicationSeries(fixture, overrides)
	state, err := scenes.EvaluateScene(def, series, fixture.Events, at.UTC())
	if err != nil {
		return model.ReplicationDemoResponse{}, err
	}
	return model.ReplicationDemoResponse{
		Scene:   def,
		Fixture: fixture,
		State:   state,
		Trace: []model.TraceStep{
			{Stage: "source connection", Classification: "correct", Detail: "Fixture dataset is loaded through a source-aware demo adapter."},
			{Stage: "source adapter", Classification: "correct", Detail: fmt.Sprintf("Fixture rows are converted into %d canonical metric series.", len(series))},
			{Stage: "sample representation", Classification: "correct", Detail: "Each metric uses timestamped canonical numeric points."},
			{Stage: "semantic interpretation", Classification: "correct", Detail: "Bindings carry metric semantics, preferred direction, and numeric ranges."},
			{Stage: "expression or rule evaluation", Classification: "incomplete", Detail: "Threshold parsing is not implemented yet; Task 00 only establishes deterministic metric evaluation and event activation."},
			{Stage: "scene binding", Classification: "correct", Detail: "Bindings map canonical metrics into normalized scene properties through one backend implementation path."},
			{Stage: "scene-state construction", Classification: "correct", Detail: "Scene state is reconstructed for the requested timestamp with active events and object properties."},
			{Stage: "x or y property mapping", Classification: "correct", Detail: "Replica x and y are derived from normalized lag values; packet speed, scale, rotation, and callout opacity are also data-bound."},
			{Stage: "frontend state transport", Classification: "correct", Detail: "The API returns scene definition, fixture metadata, trace steps, and computed scene state as JSON."},
			{Stage: "svg or dom update", Classification: "correct", Detail: "The frontend renders object transforms and labels directly from backend scene state."},
			{Stage: "animation interpolation", Classification: "correct", Detail: "The frontend uses CSS transitions between backend-provided endpoints rather than timer-driven decorative loops."},
		},
	}, nil
}

func ReplicationSceneDocument() (model.SceneDocument, error) {
	def := ReplicationSceneDefinition()
	payload, err := json.Marshal(def)
	if err != nil {
		return model.SceneDocument{}, err
	}
	return model.SceneDocument{
		ID:          def.ID,
		Slug:        "database-replication-health",
		Name:        def.Title,
		Kind:        "replication-demo",
		Status:      "published",
		Description: def.Description,
		Payload:     payload,
	}, nil
}

func replicationSeries(fixture model.ReplicationFixture, overrides model.ReplicationOverrides) []model.CanonicalSeries {
	lag := make([]model.CanonicalPoint, 0, len(fixture.Points))
	rate := make([]model.CanonicalPoint, 0, len(fixture.Points))
	write := make([]model.CanonicalPoint, 0, len(fixture.Points))
	commitAge := make([]model.CanonicalPoint, 0, len(fixture.Points))
	errors := make([]model.CanonicalPoint, 0, len(fixture.Points))

	for idx, point := range fixture.Points {
		isLast := idx == len(fixture.Points)-1
		lagValue := point.ReplicationLagSeconds
		rateValue := point.ReplicationRateMBS
		writeValue := point.WriteTPS
		commitValue := point.LastCommitAgeSeconds
		errorValue := point.Errors5M
		if isLast {
			if overrides.LagSeconds != nil {
				lagValue = *overrides.LagSeconds
			}
			if overrides.RateMBS != nil {
				rateValue = *overrides.RateMBS
			}
			if overrides.WriteTPS != nil {
				writeValue = *overrides.WriteTPS
			}
			if overrides.LastCommitAgeSeconds != nil {
				commitValue = *overrides.LastCommitAgeSeconds
			}
			if overrides.Errors5M != nil {
				errorValue = *overrides.Errors5M
			}
		}
		lag = append(lag, model.CanonicalPoint{Timestamp: point.Timestamp, Value: lagValue})
		rate = append(rate, model.CanonicalPoint{Timestamp: point.Timestamp, Value: rateValue})
		write = append(write, model.CanonicalPoint{Timestamp: point.Timestamp, Value: writeValue})
		commitAge = append(commitAge, model.CanonicalPoint{Timestamp: point.Timestamp, Value: commitValue})
		errors = append(errors, model.CanonicalPoint{Timestamp: point.Timestamp, Value: errorValue})
	}

	return []model.CanonicalSeries{
		{ID: "lag", SourceRef: fixture.SourceRef, MetricName: "replication_lag_seconds", Label: "Replication lag", Semantics: model.MetricSemantics{Type: "gauge", PreferredDirection: "lower", Range: model.NumericRange{Min: 0, Max: 120}, Unit: "s"}, Points: lag},
		{ID: "rate", SourceRef: fixture.SourceRef, MetricName: "replication_rate_mb_s", Label: "Replication rate", Semantics: model.MetricSemantics{Type: "gauge", PreferredDirection: "higher", Range: model.NumericRange{Min: 0, Max: 16}, Unit: "MB/s"}, Points: rate},
		{ID: "write", SourceRef: fixture.SourceRef, MetricName: "write_tps", Label: "Write rate", Semantics: model.MetricSemantics{Type: "gauge", PreferredDirection: "higher", Range: model.NumericRange{Min: 0, Max: 1500}, Unit: "TPS"}, Points: write},
		{ID: "commit_age", SourceRef: fixture.SourceRef, MetricName: "last_commit_age_seconds", Label: "Last commit age", Semantics: model.MetricSemantics{Type: "gauge", PreferredDirection: "lower", Range: model.NumericRange{Min: 0, Max: 180}, Unit: "s"}, Points: commitAge},
		{ID: "errors", SourceRef: fixture.SourceRef, MetricName: "errors_5m", Label: "Errors", Semantics: model.MetricSemantics{Type: "gauge", PreferredDirection: "lower", Range: model.NumericRange{Min: 0, Max: 1}, Unit: "count"}, Points: errors},
	}
}

func mustParseRFC3339(value string) time.Time {
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return ts
}
