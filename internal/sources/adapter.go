package sources

import (
	"context"
	"fmt"
	"strings"

	"infograf/internal/model"
)

// Request is transport-neutral input to a source-aware adapter. Configuration
// remains separate from scene definitions so the same scene can consume any
// adapter that emits equivalent canonical series and events.
type Request struct {
	SourceRef string
	Payload   []byte
}

type Result struct {
	Series []model.CanonicalSeries `json:"series"`
	Events []model.CanonicalEvent  `json:"events"`
}

type Adapter interface {
	Kind() string
	Adapt(context.Context, Request) (Result, error)
}

type Registry struct {
	adapters map[string]Adapter
}

func NewRegistry(adapters ...Adapter) *Registry {
	r := &Registry{adapters: make(map[string]Adapter, len(adapters))}
	for _, adapter := range adapters {
		r.adapters[adapter.Kind()] = adapter
	}
	return r
}

func DefaultRegistry() *Registry {
	return NewRegistry(CSVAdapter{}, XLSXAdapter{}, PlaintextAdapter{}, PrometheusAdapter{}, OTLPAdapter{}, RESTJSONAdapter{})
}

func (r *Registry) Adapt(ctx context.Context, kind string, req Request) (Result, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	switch kind {
	case "csv/xlsx", "csv-xlsx":
		kind = "csv"
	case "otel", "otlp":
		kind = "opentelemetry"
	case "rest/json", "rest-json":
		kind = "rest"
	case "raw", "raw/plaintext":
		kind = "plaintext"
	}
	adapter, ok := r.adapters[kind]
	if !ok {
		return Result{}, fmt.Errorf("unsupported source kind %q", kind)
	}
	return adapter.Adapt(ctx, req)
}
