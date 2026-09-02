package sources

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"infograf/internal/model"
)

// PlaintextAdapter accepts one sample per line: metric.name value timestamp.
type PlaintextAdapter struct{}

func (PlaintextAdapter) Kind() string { return "plaintext" }

func (PlaintextAdapter) Adapt(_ context.Context, req Request) (Result, error) {
	byMetric := map[string]*model.CanonicalSeries{}
	order := make([]string, 0)
	scanner := bufio.NewScanner(bytes.NewReader(req.Payload))
	line := 0
	for scanner.Scan() {
		line++
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		if len(fields) != 3 {
			return Result{}, fmt.Errorf("line %d: expected metric value timestamp", line)
		}
		value, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return Result{}, fmt.Errorf("line %d value: %w", line, err)
		}
		ts, err := parseTimestamp(fields[2])
		if err != nil {
			return Result{}, fmt.Errorf("line %d timestamp: %w", line, err)
		}
		item := byMetric[fields[0]]
		if item == nil {
			item = &model.CanonicalSeries{ID: fields[0], SourceRef: req.SourceRef, MetricName: fields[0], Label: fields[0], Semantics: model.MetricSemantics{Type: "gauge", PreferredDirection: "contextual", Unit: "value"}}
			byMetric[fields[0]] = item
			order = append(order, fields[0])
		}
		item.Points = append(item.Points, model.CanonicalPoint{Timestamp: ts, Value: value})
	}
	if err := scanner.Err(); err != nil {
		return Result{}, fmt.Errorf("read plaintext: %w", err)
	}
	if len(order) == 0 {
		return Result{}, fmt.Errorf("plaintext source contains no samples")
	}
	series := make([]model.CanonicalSeries, 0, len(order))
	for _, metric := range order {
		item := byMetric[metric]
		item.Semantics.Range = valueRange(item.Points)
		series = append(series, *item)
	}
	return Result{Series: series, Events: []model.CanonicalEvent{}}, nil
}
