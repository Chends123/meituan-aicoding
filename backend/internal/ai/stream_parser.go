package ai

import (
	"encoding/json"
	"errors"
	"strings"
)

type AnalysisOutput struct {
	PositiveKeywords []string `json:"positive_keywords"`
	NegativeKeywords []string `json:"negative_keywords"`
	SentimentScore   int      `json:"sentiment_score"`
	Suggestions      []string `json:"suggestions"`
	Summary          string   `json:"summary"`
}

func ExtractJSONObject(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start == -1 || end == -1 || end < start {
		return "", errors.New("no JSON object found in model output")
	}
	return trimmed[start : end+1], nil
}

func ParseAnalysisOutput(raw string) (*AnalysisOutput, error) {
	jsonText, err := ExtractJSONObject(raw)
	if err != nil {
		return nil, err
	}
	var out AnalysisOutput
	if err := json.Unmarshal([]byte(jsonText), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
