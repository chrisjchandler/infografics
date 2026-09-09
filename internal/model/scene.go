package model

import "time"

type NumericRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type CanonicalPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type MetricSemantics struct {
	Type               string       `json:"type"`
	PreferredDirection string       `json:"preferred_direction"`
	Range              NumericRange `json:"range"`
	Unit               string       `json:"unit"`
}

type CanonicalSeries struct {
	ID         string           `json:"id"`
	SourceRef  string           `json:"source_ref"`
	MetricName string           `json:"metric_name"`
	Label      string           `json:"label"`
	Semantics  MetricSemantics  `json:"semantics"`
	Points     []CanonicalPoint `json:"points"`
}

type CanonicalEvent struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Start      time.Time         `json:"start"`
	End        time.Time         `json:"end"`
	Title      string            `json:"title"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type SceneTarget struct {
	ObjectID string `json:"object"`
	Property string `json:"property"`
}

type SceneMapping struct {
	Direction string       `json:"direction"`
	Output    NumericRange `json:"output"`
}

type AnimationSpec struct {
	DurationMS int    `json:"duration_ms"`
	Easing     string `json:"easing"`
}

type SceneBinding struct {
	ID             string          `json:"id"`
	SourceRef      string          `json:"source_ref"`
	MetricName     string          `json:"metric_name"`
	Interpretation MetricSemantics `json:"interpretation"`
	Target         SceneTarget     `json:"target"`
	Mapping        SceneMapping    `json:"mapping"`
	Animation      AnimationSpec   `json:"animation"`
}

type SceneDefinition struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	SourceRef   string         `json:"source_ref"`
	Bindings    []SceneBinding `json:"bindings"`
}

type MetricEvaluation struct {
	BindingID     string  `json:"binding_id"`
	MetricName    string  `json:"metric_name"`
	Value         float64 `json:"value"`
	PreviousValue float64 `json:"previous_value"`
	Delta         float64 `json:"delta"`
	RatePerSecond float64 `json:"rate_per_second"`
	Normalized    float64 `json:"normalized"`
	MappedValue   float64 `json:"mapped_value"`
	Increasing    bool    `json:"increasing"`
	Decreasing    bool    `json:"decreasing"`
	Stable        bool    `json:"stable"`
	Acceleration  float64 `json:"acceleration"`
	Volatility    float64 `json:"volatility"`
}

type SceneObjectState struct {
	ID             string             `json:"id"`
	X              float64            `json:"x"`
	Y              float64            `json:"y"`
	Scale          float64            `json:"scale"`
	Rotation       float64            `json:"rotation"`
	Opacity        float64            `json:"opacity"`
	PathProgress   float64            `json:"path_progress"`
	AnimationSpeed float64            `json:"animation_speed"`
	Label          string             `json:"label,omitempty"`
	State          string             `json:"state,omitempty"`
	Properties     map[string]float64 `json:"properties,omitempty"`
}

type SceneState struct {
	SceneID      string             `json:"scene_id"`
	Timestamp    time.Time          `json:"timestamp"`
	Objects      []SceneObjectState `json:"objects"`
	Metrics      []MetricEvaluation `json:"metrics"`
	ActiveEvents []CanonicalEvent   `json:"active_events"`
}
