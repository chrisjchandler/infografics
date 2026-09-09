package model

import "time"

type SourceConnectorOption struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type GraphiteMetricOption struct {
	Metric             string `json:"metric"`
	Namespace          string `json:"namespace"`
	Label              string `json:"label"`
	Unit               string `json:"unit"`
	PreferredDirection string `json:"preferred_direction"`
}

type ScenePresetOption struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	MetricSlots int    `json:"metric_slots"`
}

type TimelinePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Label     string    `json:"label"`
	Value     float64   `json:"value"`
}

type SceneAnnotation struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
	Severity  string    `json:"severity"`
}

type GraphitePeriodOption struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Window string `json:"window"`
}

type GraphiteCard struct {
	Label        string    `json:"label"`
	Metric       string    `json:"metric"`
	ValueDisplay string    `json:"value_display"`
	DeltaDisplay string    `json:"delta_display"`
	Detail       string    `json:"detail"`
	Status       string    `json:"status"`
	Progress     int       `json:"progress"`
	Live         bool      `json:"live"`
	Points       []float64 `json:"points,omitempty"`
}

type GraphiteOverview struct {
	Connected   bool                   `json:"connected"`
	SourceURL   string                 `json:"source_url"`
	SourceLabel string                 `json:"source_label"`
	Period      string                 `json:"period"`
	PeriodLabel string                 `json:"period_label"`
	Window      string                 `json:"window"`
	CapturedAt  time.Time              `json:"captured_at"`
	Cards       []GraphiteCard         `json:"cards"`
	Notes       []string               `json:"notes"`
	Periods     []GraphitePeriodOption `json:"periods"`
}

type GraphiteCatalogResponse struct {
	Connected   bool                    `json:"connected"`
	SourceURL   string                  `json:"source_url"`
	SourceLabel string                  `json:"source_label"`
	CapturedAt  time.Time               `json:"captured_at"`
	Metrics     []GraphiteMetricOption  `json:"metrics"`
	Periods     []GraphitePeriodOption  `json:"periods"`
	Presets     []ScenePresetOption     `json:"presets"`
	Connectors  []SourceConnectorOption `json:"connectors"`
}

type GraphiteSceneResponse struct {
	Connected        bool                    `json:"connected"`
	SourceURL        string                  `json:"source_url"`
	SourceLabel      string                  `json:"source_label"`
	Period           string                  `json:"period"`
	PeriodLabel      string                  `json:"period_label"`
	Window           string                  `json:"window"`
	SelectedPreset   string                  `json:"selected_preset"`
	SelectedMetrics  []string                `json:"selected_metrics"`
	AvailableMetrics []GraphiteMetricOption  `json:"available_metrics"`
	ScenePresets     []ScenePresetOption     `json:"scene_presets"`
	Connectors       []SourceConnectorOption `json:"connectors"`
	Timeline         []TimelinePoint         `json:"timeline"`
	Annotations      []SceneAnnotation       `json:"annotations"`
	Scene            SceneDefinition         `json:"scene"`
	State            SceneState              `json:"state"`
	Frames           []SceneState            `json:"frames"`
	Series           []CanonicalSeries       `json:"series"`
}
