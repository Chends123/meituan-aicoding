package ai

import (
	"strconv"
	"strings"
)

type StreamSnapshot struct {
	Summary          string
	PositiveKeywords []string
	NegativeKeywords []string
	SentimentScore   *int
	Suggestions      []string
}

var orderedSections = []string{"SUMMARY", "POSITIVE_KEYWORDS", "NEGATIVE_KEYWORDS", "SENTIMENT_SCORE", "SUGGESTIONS"}

func ParseStructuredStream(raw string) StreamSnapshot {
	normalized := normalizeStream(raw)
	return StreamSnapshot{
		Summary:          sectionText(normalized, "SUMMARY"),
		PositiveKeywords: parseBulletList(sectionText(normalized, "POSITIVE_KEYWORDS")),
		NegativeKeywords: parseBulletList(sectionText(normalized, "NEGATIVE_KEYWORDS")),
		SentimentScore:   parseSentimentScore(sectionText(normalized, "SENTIMENT_SCORE")),
		Suggestions:      parseBulletList(sectionText(normalized, "SUGGESTIONS")),
	}
}

func normalizeStream(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return trimmed
}

func sectionText(raw string, name string) string {
	startTag := "[" + name + "]"
	endTag := "[/" + name + "]"
	start := strings.Index(raw, startTag)
	if start == -1 {
		return ""
	}
	start += len(startTag)
	rest := raw[start:]
	end := strings.Index(rest, endTag)
	if end != -1 {
		return strings.TrimSpace(rest[:end])
	}
	next := len(rest)
	for _, candidate := range orderedSections {
		if candidate == name {
			continue
		}
		idx := strings.Index(rest, "["+candidate+"]")
		if idx != -1 && idx < next {
			next = idx
		}
	}
	return strings.TrimSpace(rest[:next])
}

func parseBulletList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	lines := strings.Split(raw, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		cleaned := strings.TrimSpace(line)
		cleaned = strings.TrimPrefix(cleaned, "- ")
		cleaned = strings.TrimPrefix(cleaned, "-")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned == "" {
			continue
		}
		result = append(result, cleaned)
	}
	return result
}

func parseSentimentScore(raw string) *int {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return nil
	}
	for _, line := range strings.Split(cleaned, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		value, err := strconv.Atoi(line)
		if err == nil {
			return &value
		}
	}
	return nil
}
