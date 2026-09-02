package sources

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"infograf/internal/model"
)

// PrometheusAdapter accepts the standard text exposition format. It is
// deliberately transport-neutral: HTTP range queries and auth belong to the
// connector configuration layer, while this adapter only canonicalizes data.
type PrometheusAdapter struct{}

var promName = regexp.MustCompile(`^([a-zA-Z_:][a-zA-Z0-9_:]*)(?:\{([^}]*)\})?$`)

func (PrometheusAdapter) Kind() string { return "prometheus" }

func (PrometheusAdapter) Adapt(_ context.Context, req Request) (Result, error) {
	series := map[string]*model.CanonicalSeries{}
	order := []string{}
	scanner := bufio.NewScanner(strings.NewReader(string(req.Payload)))
	line := 0
	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		fields := strings.Fields(raw)
		if len(fields) < 2 {
			return Result{}, fmt.Errorf("prometheus line %d: expected metric and value", line)
		}
		match := promName.FindStringSubmatch(fields[0])
		if match == nil {
			return Result{}, fmt.Errorf("prometheus line %d: invalid metric %q", line, fields[0])
		}
		value, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return Result{}, fmt.Errorf("prometheus line %d value: %w", line, err)
		}
		stamp := ""
		if len(fields) > 2 {
			stamp = fields[2]
		}
		var ts = modelTimeNow()
		if stamp != "" {
			ms, err := strconv.ParseInt(stamp, 10, 64)
			if err != nil {
				return Result{}, fmt.Errorf("prometheus line %d timestamp: %w", line, err)
			}
			ts = unixMillis(ms)
		}
		metric := match[1]
		if match[2] != "" {
			metric += "{" + match[2] + "}"
		}
		item := series[metric]
		if item == nil {
			item = &model.CanonicalSeries{ID: metric, SourceRef: req.SourceRef, MetricName: metric, Label: metric, Semantics: model.MetricSemantics{Type: "gauge", PreferredDirection: "contextual", Unit: "value"}}
			series[metric] = item
			order = append(order, metric)
		}
		item.Points = append(item.Points, model.CanonicalPoint{Timestamp: ts, Value: value})
	}
	if err := scanner.Err(); err != nil {
		return Result{}, err
	}
	return finalizeSeries(series, order)
}
