package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"infograf/internal/model"
)

type metricAnalysisRequest struct {
	Series model.CanonicalSeries `json:"series"`
}
type metricAnalysisResponse struct {
	Summary  string `json:"summary"`
	Provider string `json:"provider"`
	Model    string `json:"model,omitempty"`
}

func (s *Server) handleMetricAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input metricAnalysisRequest
	if err := decodeJSON(r.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis input", err.Error(), "bad_request")
		return
	}
	if strings.TrimSpace(input.Series.MetricName) == "" || len(input.Series.Points) == 0 {
		writeError(w, http.StatusBadRequest, "invalid analysis input", "series.metric_name and series.points are required", "bad_request")
		return
	}
	result, err := analyzeMetric(r.Context(), input.Series)
	if err != nil {
		writeError(w, http.StatusBadGateway, "analysis provider failed", err.Error(), "analysis_error")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func analyzeMetric(ctx context.Context, series model.CanonicalSeries) (metricAnalysisResponse, error) {
	modelName := strings.TrimSpace(os.Getenv("INFOGRAF_AI_MODEL"))
	if modelName == "" {
		modelName = "llama3.2:3b"
	}
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("INFOGRAF_AI_URL")), "/")
	if baseURL == "" {
		return metricAnalysisResponse{Summary: localMetricSummary(series), Provider: "local", Model: "deterministic"}, nil
	}
	prompt := metricPrompt(series)
	body, _ := json.Marshal(map[string]any{"model": modelName, "prompt": prompt, "stream": false})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return metricAnalysisResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return metricAnalysisResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return metricAnalysisResponse{}, fmt.Errorf("AI endpoint returned %s", resp.Status)
	}
	var output struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return metricAnalysisResponse{}, err
	}
	if strings.TrimSpace(output.Response) == "" {
		return metricAnalysisResponse{}, fmt.Errorf("AI endpoint returned an empty response")
	}
	return metricAnalysisResponse{Summary: strings.TrimSpace(output.Response), Provider: "ollama", Model: modelName}, nil
}

func metricPrompt(series model.CanonicalSeries) string {
	limit := len(series.Points)
	if limit > 24 {
		limit = 24
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Analyze metric %s (%s) using these timestamp/value samples. Be concise: state range, direction/trend, volatility, and one operational implication. Do not invent causes.\n", series.MetricName, series.Semantics.Unit)
	for _, point := range series.Points[len(series.Points)-limit:] {
		fmt.Fprintf(&b, "%s %.6g\n", point.Timestamp.UTC().Format(time.RFC3339), point.Value)
	}
	return b.String()
}

func localMetricSummary(series model.CanonicalSeries) string {
	values := make([]float64, 0, len(series.Points))
	for _, point := range series.Points {
		values = append(values, point.Value)
	}
	min, max := values[0], values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	delta := values[len(values)-1] - values[0]
	direction := "stable"
	if delta > 0 {
		direction = "rising"
	}
	if delta < 0 {
		direction = "falling"
	}
	unit := strings.TrimSpace(series.Semantics.Unit)
	if unit == "" {
		unit = "value"
	}
	return fmt.Sprintf("%s spans %.3g–%.3g %s across %d samples and is %s overall (net change %+.3g). This deterministic local summary is ready for an AI provider when INFOGRAF_AI_URL is configured.", series.MetricName, min, max, unit, len(values), direction, delta)
}
