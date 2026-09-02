package scenes

import (
	"fmt"
	"math"
	"slices"
	"time"

	"infograf/internal/model"
)

func Normalize(value, min, max float64) (float64, error) {
	if min == max {
		return 0, fmt.Errorf("invalid range: min equals max")
	}
	normalized := (value - min) / (max - min)
	if normalized < 0 {
		return 0, nil
	}
	if normalized > 1 {
		return 1, nil
	}
	return normalized, nil
}

func MapNormalized(normalized, outMin, outMax float64, inverse bool) float64 {
	if inverse {
		normalized = 1 - normalized
	}
	return outMin + (outMax-outMin)*normalized
}

type seriesSample struct {
	value     float64
	previous  float64
	delta     float64
	rate      float64
	accel     float64
	volatility float64
}

func EvaluateScene(def model.SceneDefinition, series []model.CanonicalSeries, events []model.CanonicalEvent, at time.Time) (model.SceneState, error) {
	index := make(map[string]model.CanonicalSeries, len(series))
	for _, item := range series {
		index[item.MetricName] = item
	}

	objects := map[string]*model.SceneObjectState{}
	metrics := make([]model.MetricEvaluation, 0, len(def.Bindings))
	for _, binding := range def.Bindings {
		item, ok := index[binding.MetricName]
		if !ok {
			return model.SceneState{}, fmt.Errorf("series not found for metric %q", binding.MetricName)
		}
		sample, err := sampleAt(item.Points, at)
		if err != nil {
			return model.SceneState{}, fmt.Errorf("sample binding %q: %w", binding.ID, err)
		}
		normalized, err := Normalize(sample.value, binding.Interpretation.Range.Min, binding.Interpretation.Range.Max)
		if err != nil {
			return model.SceneState{}, fmt.Errorf("normalize binding %q: %w", binding.ID, err)
		}
		mapped := MapNormalized(normalized, binding.Mapping.Output.Min, binding.Mapping.Output.Max, binding.Mapping.Direction == "inverse")
		metric := model.MetricEvaluation{
			BindingID:     binding.ID,
			MetricName:    binding.MetricName,
			Value:         sample.value,
			PreviousValue: sample.previous,
			Delta:         sample.delta,
			RatePerSecond: sample.rate,
			Normalized:    normalized,
			MappedValue:   mapped,
			Increasing:    sample.delta > 0,
			Decreasing:    sample.delta < 0,
			Stable:        math.Abs(sample.delta) < 1e-9,
			Acceleration:  sample.accel,
			Volatility:    sample.volatility,
		}
		metrics = append(metrics, metric)

		obj := objects[binding.Target.ObjectID]
		if obj == nil {
			obj = &model.SceneObjectState{
				ID:         binding.Target.ObjectID,
				Scale:      1,
				Opacity:    1,
				Properties: map[string]float64{},
			}
			objects[binding.Target.ObjectID] = obj
		}
		applyMappedProperty(obj, binding.Target.Property, mapped, metric)
	}

	activeEvents := activeEventsAt(events, at)
	for _, obj := range objects {
		if obj.ID == "incident-callout" {
			obj.Opacity = 0
			obj.State = "idle"
			if len(activeEvents) > 0 {
				obj.Opacity = 1
				obj.State = "active"
				obj.Label = activeEvents[0].Title
			}
		}
		if obj.ID == "replica-db" {
			if hasErrorMetric(metrics) {
				obj.State = "warning"
			} else if len(activeEvents) > 0 {
				obj.State = "incident"
			} else if metricByBinding(metrics, "replication-lag-separation").MappedValue >= 0.82 {
				obj.State = "critical"
			} else if metricByBinding(metrics, "replication-lag-separation").MappedValue >= 0.68 {
				obj.State = "warning"
			} else {
				obj.State = "healthy"
			}
		}
		if obj.ID == "packet-stream" {
			obj.State = "flowing"
			if metricByBinding(metrics, "replication-rate-packets").Value <= 0 {
				obj.State = "stalled"
			}
		}
	}

	orderedObjects := make([]model.SceneObjectState, 0, len(objects))
	keys := make([]string, 0, len(objects))
	for key := range objects {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		orderedObjects = append(orderedObjects, *objects[key])
	}

	return model.SceneState{
		SceneID:      def.ID,
		Timestamp:    at.UTC(),
		Objects:      orderedObjects,
		Metrics:      metrics,
		ActiveEvents: activeEvents,
	}, nil
}

func applyMappedProperty(obj *model.SceneObjectState, property string, mapped float64, metric model.MetricEvaluation) {
	switch property {
	case "x":
		obj.X = mapped
	case "y":
		obj.Y = mapped
	case "scale":
		obj.Scale = mapped
	case "rotation":
		obj.Rotation = mapped
	case "opacity":
		obj.Opacity = mapped
	case "path_progress":
		obj.PathProgress = mapped
	case "animation_speed":
		obj.AnimationSpeed = mapped
	default:
		obj.Properties[property] = mapped
	}
	if property == "label_text" {
		obj.Label = fmt.Sprintf("%.2f", metric.Value)
	}
}

func sampleAt(points []model.CanonicalPoint, at time.Time) (seriesSample, error) {
	if len(points) == 0 {
		return seriesSample{}, fmt.Errorf("series has no points")
	}
	if at.Before(points[0].Timestamp) {
		return sampleFromWindow(points, 0, 0), nil
	}
	last := len(points) - 1
	if at.After(points[last].Timestamp) {
		return sampleFromWindow(points, last, last), nil
	}
	for idx := 1; idx < len(points); idx++ {
		prev := points[idx-1]
		next := points[idx]
		if at.Equal(next.Timestamp) {
			return sampleFromWindow(points, idx-1, idx), nil
		}
		if at.After(prev.Timestamp) && at.Before(next.Timestamp) {
			span := next.Timestamp.Sub(prev.Timestamp).Seconds()
			if span <= 0 {
				return sampleFromWindow(points, idx-1, idx), nil
			}
			progress := at.Sub(prev.Timestamp).Seconds() / span
			value := prev.Value + (next.Value-prev.Value)*progress
			delta := value - prev.Value
			rate := (next.Value - prev.Value) / span
			accel := accelerationAt(points, idx)
			volatility := volatilityAround(points, idx)
			return seriesSample{
				value:      value,
				previous:   prev.Value,
				delta:      delta,
				rate:       rate,
				accel:      accel,
				volatility: volatility,
			}, nil
		}
	}
	return sampleFromWindow(points, 0, 0), nil
}

func sampleFromWindow(points []model.CanonicalPoint, prevIdx, currIdx int) seriesSample {
	current := points[currIdx]
	previous := current
	if prevIdx >= 0 && prevIdx < len(points) {
		previous = points[prevIdx]
	}
	seconds := current.Timestamp.Sub(previous.Timestamp).Seconds()
	rate := 0.0
	if seconds > 0 {
		rate = (current.Value - previous.Value) / seconds
	}
	return seriesSample{
		value:      current.Value,
		previous:   previous.Value,
		delta:      current.Value - previous.Value,
		rate:       rate,
		accel:      accelerationAt(points, currIdx),
		volatility: volatilityAround(points, currIdx),
	}
}

func accelerationAt(points []model.CanonicalPoint, idx int) float64 {
	if idx < 2 || idx >= len(points) {
		return 0
	}
	a := rateBetween(points[idx-2], points[idx-1])
	b := rateBetween(points[idx-1], points[idx])
	return b - a
}

func volatilityAround(points []model.CanonicalPoint, idx int) float64 {
	start := max(0, idx-2)
	end := min(len(points)-1, idx+2)
	if end-start < 1 {
		return 0
	}
	values := make([]float64, 0, end-start+1)
	for i := start; i <= end; i++ {
		values = append(values, points[i].Value)
	}
	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))
	variance := 0.0
	for _, value := range values {
		diff := value - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}

func rateBetween(a, b model.CanonicalPoint) float64 {
	seconds := b.Timestamp.Sub(a.Timestamp).Seconds()
	if seconds <= 0 {
		return 0
	}
	return (b.Value - a.Value) / seconds
}

func activeEventsAt(events []model.CanonicalEvent, at time.Time) []model.CanonicalEvent {
	active := make([]model.CanonicalEvent, 0, len(events))
	for _, event := range events {
		if (at.Equal(event.Start) || at.After(event.Start)) && (at.Equal(event.End) || at.Before(event.End)) {
			active = append(active, event)
		}
	}
	return active
}

func hasErrorMetric(metrics []model.MetricEvaluation) bool {
	return metricByBinding(metrics, "error-opacity").Value > 0
}

func metricByBinding(metrics []model.MetricEvaluation, bindingID string) model.MetricEvaluation {
	for _, metric := range metrics {
		if metric.BindingID == bindingID {
			return metric
		}
	}
	return model.MetricEvaluation{}
}
