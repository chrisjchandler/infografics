package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"infograf/internal/model"
)

// XLSXAdapter is the dependency-free spreadsheet template. Upload handlers
// may provide the selected worksheet as CSV bytes; native ZIP/XLSX decoding is
// intentionally kept behind the connector boundary for a future optional
// parser, while the canonical contract is already stable.
type XLSXAdapter struct{}

func (XLSXAdapter) Kind() string { return "xlsx" }
func (XLSXAdapter) Adapt(ctx context.Context, req Request) (Result, error) {
	if strings.HasPrefix(string(req.Payload), "PK\x03\x04") {
		return Result{}, fmt.Errorf("xlsx workbook requires worksheet extraction before adaptation")
	}
	return (CSVAdapter{}).Adapt(ctx, req)
}

type OTLPAdapter struct{}

func (OTLPAdapter) Kind() string { return "opentelemetry" }
func (OTLPAdapter) Adapt(_ context.Context, req Request) (Result, error) {
	var root map[string]any
	if err := json.Unmarshal(req.Payload, &root); err != nil {
		return Result{}, fmt.Errorf("parse OTLP JSON: %w", err)
	}
	series := map[string]*model.CanonicalSeries{}
	order := []string{}
	walkOTLPMetrics(root, req.SourceRef, series, &order)
	if len(order) == 0 {
		return Result{}, fmt.Errorf("OTLP payload contains no metric data points")
	}
	return finalizeSeries(series, order)
}

func walkOTLPMetrics(node any, source string, series map[string]*model.CanonicalSeries, order *[]string) {
	switch value := node.(type) {
	case map[string]any:
		metricName, _ := value["name"].(string)
		unit, _ := value["unit"].(string)
		for _, key := range []string{"dataPoints", "gauge", "sum", "histogram", "exponentialHistogram"} {
			if child, ok := value[key]; ok {
				walkOTLPPoints(child, metricName, unit, source, series, order)
			}
		}
		// Data-point branches were already consumed above. Descend only through
		// OTLP container branches to avoid duplicate fallback series.
		for _, key := range []string{"resourceMetrics", "scopeMetrics", "metrics", "resource", "scope"} {
			if child, ok := value[key]; ok {
				walkOTLPMetrics(child, source, series, order)
			}
		}
	case []any:
		for _, child := range value {
			walkOTLPMetrics(child, source, series, order)
		}
	}
}

func walkOTLPPoints(node any, metricName, unit, source string, series map[string]*model.CanonicalSeries, order *[]string) {
	if metricName == "" {
		metricName = "otel.metric"
	}
	switch value := node.(type) {
	case map[string]any:
		raw, ok := value["value"].(float64)
		if !ok {
			raw, ok = value["asDouble"].(float64)
		}
		if !ok {
			if integer, exists := value["asInt"].(float64); exists {
				raw, ok = integer, true
			}
		}
		if ok {
			ts := modelTimeNow()
			if nano, ok := value["timeUnixNano"].(string); ok {
				if n, err := parseInt64(nano); err == nil {
					ts = time.Unix(0, n).UTC()
				}
			}
			item := series[metricName]
			if item == nil {
				item = &model.CanonicalSeries{ID: metricName, SourceRef: source, MetricName: metricName, Label: metricName, Semantics: model.MetricSemantics{Type: "gauge", PreferredDirection: "contextual", Unit: unit}}
				series[metricName] = item
				*order = append(*order, metricName)
			}
			item.Points = append(item.Points, model.CanonicalPoint{Timestamp: ts, Value: raw})
		}
		for _, child := range value {
			walkOTLPPoints(child, metricName, unit, source, series, order)
		}
	case []any:
		for _, child := range value {
			walkOTLPPoints(child, metricName, unit, source, series, order)
		}
	}
}

type RESTJSONAdapter struct{}

func (RESTJSONAdapter) Kind() string { return "rest" }
func (RESTJSONAdapter) Adapt(_ context.Context, req Request) (Result, error) {
	var root any
	if err := json.Unmarshal(req.Payload, &root); err != nil {
		return Result{}, fmt.Errorf("parse REST JSON: %w", err)
	}
	series := map[string]*model.CanonicalSeries{}
	order := []string{}
	walkREST(root, req.SourceRef, series, &order)
	if len(order) == 0 {
		return Result{}, fmt.Errorf("REST JSON contains no {metric,value,timestamp} samples")
	}
	return finalizeSeries(series, order)
}

func walkREST(node any, source string, series map[string]*model.CanonicalSeries, order *[]string) {
	switch value := node.(type) {
	case map[string]any:
		metric, mok := value["metric"].(string)
		if !mok {
			metric, mok = value["name"].(string)
		}
		raw, vok := value["value"].(float64)
		if mok && vok {
			ts := timestampFromAny(value["timestamp"])
			item := series[metric]
			if item == nil {
				item = &model.CanonicalSeries{ID: metric, SourceRef: source, MetricName: metric, Label: metric, Semantics: model.MetricSemantics{Type: "gauge", PreferredDirection: "contextual", Unit: "value"}}
				series[metric] = item
				*order = append(*order, metric)
			}
			item.Points = append(item.Points, model.CanonicalPoint{Timestamp: ts, Value: raw})
		}
		for _, child := range value {
			walkREST(child, source, series, order)
		}
	case []any:
		for _, child := range value {
			walkREST(child, source, series, order)
		}
	}
}

func finalizeSeries(series map[string]*model.CanonicalSeries, order []string) (Result, error) {
	out := make([]model.CanonicalSeries, 0, len(order))
	for _, key := range order {
		item := series[key]
		if len(item.Points) > 0 {
			item.Semantics.Range = valueRange(item.Points)
			out = append(out, *item)
		}
	}
	if len(out) == 0 {
		return Result{}, fmt.Errorf("source contains no usable points")
	}
	return Result{Series: out, Events: []model.CanonicalEvent{}}, nil
}

func modelTimeNow() time.Time { return time.Unix(0, 0).UTC() }
func unixMillis(ms int64) time.Time {
	return time.Unix(ms/1000, (ms%1000)*int64(time.Millisecond)).UTC()
}
func parseInt64(raw string) (int64, error) { var n int64; _, err := fmt.Sscan(raw, &n); return n, err }

func timestampFromAny(raw any) time.Time {
	switch value := raw.(type) {
	case string:
		if parsed, err := parseTimestamp(value); err == nil {
			return parsed
		}
		if n, err := parseInt64(value); err == nil {
			return unixMillis(n)
		}
	case float64:
		if value > 1e14 {
			return time.Unix(0, int64(value)).UTC()
		}
		if value > 1e11 {
			return unixMillis(int64(value))
		}
		return time.Unix(int64(value), 0).UTC()
	}
	return modelTimeNow()
}
