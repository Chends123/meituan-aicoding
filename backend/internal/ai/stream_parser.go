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
	summaryText, summaryComplete := sectionText(normalized, "SUMMARY")
	positiveText, positiveComplete := sectionText(normalized, "POSITIVE_KEYWORDS")
	negativeText, negativeComplete := sectionText(normalized, "NEGATIVE_KEYWORDS")
	scoreText, scoreComplete := sectionText(normalized, "SENTIMENT_SCORE")
	suggestionText, suggestionComplete := sectionText(normalized, "SUGGESTIONS")

	snapshot := StreamSnapshot{
		PositiveKeywords: parseBulletList(positiveText, positiveComplete),
		NegativeKeywords: parseBulletList(negativeText, negativeComplete),
		Suggestions:      parseBulletList(suggestionText, suggestionComplete),
	}
	if summaryComplete {
		snapshot.Summary = strings.TrimSpace(summaryText)
	}
	if scoreComplete {
		snapshot.SentimentScore = parseSentimentScore(scoreText)
	}
	return snapshot
}

func normalizeStream(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return trimmed
}

func sectionText(raw string, name string) (string, bool) {
	startTag := "[" + name + "]"
	endTag := "[/" + name + "]"
	start := strings.Index(raw, startTag)
	if start == -1 {
		return "", false
	}
	start += len(startTag)
	rest := raw[start:]
	end := strings.Index(rest, endTag)
	if end != -1 {
		return strings.TrimSpace(rest[:end]), true
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
	if next < len(rest) {
		return strings.TrimSpace(rest[:next]), true
	}
	return strings.TrimSpace(rest[:next]), false
}

func parseBulletList(raw string, complete bool) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	lines := strings.Split(raw, "\n")
	if !complete && len(lines) > 0 {
		lastLine := strings.TrimSpace(lines[len(lines)-1])
		if lastLine != "" {
			lines = lines[:len(lines)-1]
		}
	}
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		cleaned := strings.TrimSpace(line)
		cleaned = strings.TrimPrefix(cleaned, "- ")
		cleaned = strings.TrimPrefix(cleaned, "-")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned == "" || strings.Contains(cleaned, "[") {
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
		if line == "" || strings.Contains(line, "[") {
			continue
		}
		value, err := strconv.Atoi(line)
		if err == nil {
			return &value
		}
	}
	return nil
}
