package sources

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"infograf/internal/model"
)

type CSVAdapter struct{}

func (CSVAdapter) Kind() string { return "csv" }

func (CSVAdapter) Adapt(_ context.Context, req Request) (Result, error) {
	reader := csv.NewReader(bytes.NewReader(req.Payload))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return Result{}, fmt.Errorf("parse csv: %w", err)
	}
	if len(records) < 2 || len(records[0]) < 2 {
		return Result{}, fmt.Errorf("csv requires a timestamp column and at least one metric column")
	}
	header := records[0]
	series := make([]model.CanonicalSeries, len(header)-1)
	for idx, name := range header[1:] {
		name = strings.TrimSpace(name)
		if name == "" {
			return Result{}, fmt.Errorf("metric column %d has no name", idx+2)
		}
		series[idx] = model.CanonicalSeries{
			ID: name, SourceRef: req.SourceRef, MetricName: name, Label: name,
			Semantics: model.MetricSemantics{Type: "gauge", PreferredDirection: "contextual", Unit: "value"},
		}
	}
	for rowIndex, record := range records[1:] {
		if len(record) != len(header) {
			return Result{}, fmt.Errorf("row %d has %d columns, want %d", rowIndex+2, len(record), len(header))
		}
		ts, err := parseTimestamp(record[0])
		if err != nil {
			return Result{}, fmt.Errorf("row %d timestamp: %w", rowIndex+2, err)
		}
		for metricIndex, raw := range record[1:] {
			value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
			if err != nil {
				return Result{}, fmt.Errorf("row %d metric %q: %w", rowIndex+2, header[metricIndex+1], err)
			}
			series[metricIndex].Points = append(series[metricIndex].Points, model.CanonicalPoint{Timestamp: ts, Value: value})
		}
	}
	for idx := range series {
		series[idx].Semantics.Range = valueRange(series[idx].Points)
	}
	return Result{Series: series, Events: []model.CanonicalEvent{}}, nil
}

func parseTimestamp(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if ts, err := time.Parse(time.RFC3339, raw); err == nil {
		return ts.UTC(), nil
	}
	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected RFC3339 or Unix seconds, got %q", raw)
	}
	return time.Unix(seconds, 0).UTC(), nil
}

func valueRange(points []model.CanonicalPoint) model.NumericRange {
	min, max := points[0].Value, points[0].Value
	for _, point := range points[1:] {
		if point.Value < min {
			min = point.Value
		}
		if point.Value > max {
			max = point.Value
		}
	}
	if min == max {
		if min == 0 {
			return model.NumericRange{Min: 0, Max: 1}
		}
		padding := min * 0.1
		if padding < 0 {
			padding = -padding
		}
		return model.NumericRange{Min: min - padding, Max: max + padding}
	}
	return model.NumericRange{Min: min, Max: max}
}
