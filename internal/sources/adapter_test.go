package sources

import (
	"context"
	"testing"
)

func TestCSVAndPlaintextProduceEquivalentCanonicalSeries(t *testing.T) {
	t.Parallel()
	registry := DefaultRegistry()
	csvResult, err := registry.Adapt(context.Background(), "csv", Request{SourceRef: "csv-demo", Payload: []byte("timestamp,availability\n2026-09-02T12:00:00Z,99.9\n2026-09-02T12:01:00Z,98.5\n")})
	if err != nil {
		t.Fatalf("adapt csv: %v", err)
	}
	textResult, err := registry.Adapt(context.Background(), "plaintext", Request{SourceRef: "text-demo", Payload: []byte("availability 99.9 1788350400\navailability 98.5 1788350460\n")})
	if err != nil {
		t.Fatalf("adapt plaintext: %v", err)
	}
	if len(csvResult.Series) != 1 || len(textResult.Series) != 1 {
		t.Fatalf("unexpected series counts")
	}
	for idx := range csvResult.Series[0].Points {
		a, b := csvResult.Series[0].Points[idx], textResult.Series[0].Points[idx]
		if !a.Timestamp.Equal(b.Timestamp) || a.Value != b.Value {
			t.Fatalf("point %d differs: %#v %#v", idx, a, b)
		}
	}
}

func TestCSVRejectsInvalidTimestamp(t *testing.T) {
	t.Parallel()
	_, err := (CSVAdapter{}).Adapt(context.Background(), Request{Payload: []byte("timestamp,value\nnot-time,3\n")})
	if err == nil {
		t.Fatal("expected invalid timestamp error")
	}
}

func TestPrometheusAdapterCanonicalizesTextExposition(t *testing.T) {
	t.Parallel()
	result, err := (PrometheusAdapter{}).Adapt(context.Background(), Request{SourceRef: "prom", Payload: []byte("# HELP cpu CPU\ncpu_usage{host=\"a\"} 42 1788350400000\n")})
	if err != nil {
		t.Fatalf("adapt prometheus: %v", err)
	}
	if len(result.Series) != 1 || result.Series[0].MetricName != `cpu_usage{host="a"}` || result.Series[0].Points[0].Value != 42 {
		t.Fatalf("unexpected prometheus result: %#v", result)
	}
}

func TestRESTAndOTLPAdaptersCanonicalizeCommonJSON(t *testing.T) {
	t.Parallel()
	rest, err := (RESTJSONAdapter{}).Adapt(context.Background(), Request{SourceRef: "rest", Payload: []byte(`{"data":[{"metric":"latency","value":12.5,"timestamp":"2026-09-02T12:00:00Z"}]}`)})
	if err != nil || len(rest.Series) != 1 || rest.Series[0].Points[0].Value != 12.5 {
		t.Fatalf("unexpected REST result: %#v %v", rest, err)
	}
	otlp, err := (OTLPAdapter{}).Adapt(context.Background(), Request{SourceRef: "otel", Payload: []byte(`{"resourceMetrics":[{"scopeMetrics":[{"metrics":[{"name":"latency","unit":"ms","gauge":{"dataPoints":[{"asDouble":12.5,"timeUnixNano":"1788350400000000000"}]}}]}]}]}`)})
	if err != nil || len(otlp.Series) != 1 || otlp.Series[0].Points[0].Value != 12.5 {
		t.Fatalf("unexpected OTLP result: %#v %v", otlp, err)
	}
}

func TestRegistryAcceptsConnectorAliases(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"csv/xlsx", "otel", "rest/json", "raw/plaintext"} {
		payload := []byte("timestamp,value\n2026-09-02T12:00:00Z,1\n")
		if kind == "otel" {
			payload = []byte(`{"resourceMetrics":[{"scopeMetrics":[{"metrics":[{"name":"x","gauge":{"dataPoints":[{"asDouble":1,"timeUnixNano":"1788350400000000000"}]}}]}]}]}`)
		}
		if kind == "rest/json" {
			payload = []byte(`{"metric":"x","value":1}`)
		}
		if kind == "raw/plaintext" {
			payload = []byte("x 1 1788350400\n")
		}
		if result, err := DefaultRegistry().Adapt(context.Background(), kind, Request{Payload: payload}); err != nil || len(result.Series) != 1 {
			t.Fatalf("alias %q: %#v %v", kind, result, err)
		}
	}
}
