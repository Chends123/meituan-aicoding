package service

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
	"gorm.io/gorm"

	internalai "meituan-aicoding/backend/internal/ai"
	"meituan-aicoding/backend/internal/config"
	"meituan-aicoding/backend/internal/repository"
)

type AIService struct {
	repo             *repository.ReviewRepository
	cfg              config.AIConfig
	analysisRunner   *runner.Runner
	analysisSessions session.Service
	replyRunner      *runner.Runner
	replySessions    session.Service
}

func NewAIService(db *gorm.DB, cfg config.AIConfig) (*AIService, error) {
	svc := &AIService{
		repo: repository.NewReviewRepository(db),
		cfg:  cfg,
	}

	if strings.TrimSpace(cfg.GoogleAPIKey) != "" {
		analysisRunner, analysisSessions, err := internalai.NewAnalysisRunner(context.Background(), cfg)
		if err != nil {
			return nil, fmt.Errorf("init analysis runner: %w", err)
		}
		svc.analysisRunner = analysisRunner
		svc.analysisSessions = analysisSessions

		replyRunner, replySessions, err := internalai.NewReplyRunner(context.Background(), cfg)
		if err != nil {
			return nil, fmt.Errorf("init reply runner: %w", err)
		}
		svc.replyRunner = replyRunner
		svc.replySessions = replySessions
	}

	return svc, nil
}

func (s *AIService) StreamReviewAnalysis(c *gin.Context, tab string) error {
	if strings.TrimSpace(s.cfg.GoogleAPIKey) == "" {
		return fmt.Errorf("GOOGLE_API_KEY is not configured")
	}

	normalizedTab := normalizeTab(tab)
	reviews, err := s.repo.ListByTab(c.Request.Context(), normalizedTab)
	if err != nil {
		return err
	}
	if len(reviews) == 0 {
		writeSSEJSON(c, "meta", gin.H{"review_count": 0, "tab": normalizedTab, "source": "adk-go"})
		writeSSEJSON(c, "positive_keywords", gin.H{"content": []string{}})
		writeSSEJSON(c, "negative_keywords", gin.H{"content": []string{}})
		writeSSEJSON(c, "sentiment_score", gin.H{"content": 0})
		writeSSEJSON(c, "suggestions", gin.H{"content": []string{"当前没有可分析的评价数据。"}})
		writeSSEJSON(c, "summary", gin.H{"content": "当前暂无评价数据，暂时无法生成经营分析。"})
		writeSSEJSON(c, "done", gin.H{"success": true})
		return nil
	}

	createResp, err := s.analysisSessions.Create(c.Request.Context(), &session.CreateRequest{
		AppName:   "meituan-review-ai",
		UserID:    "merchant-user",
		SessionID: fmt.Sprintf("analysis-%d", time.Now().UnixNano()),
	})
	if err != nil {
		return err
	}

	prompt := internalai.BuildReviewAnalysisPrompt(reviews, normalizedTab)
	userMessage := genai.NewContentFromText(prompt, genai.RoleUser)
	writeSSEJSON(c, "meta", gin.H{"review_count": len(reviews), "tab": normalizedTab, "source": "adk-go-structured-stream"})

	var allText strings.Builder
	var emittedText string
	lastSnapshot := internalai.StreamSnapshot{}
	for event, runErr := range s.analysisRunner.Run(c.Request.Context(), "merchant-user", createResp.Session.ID(), userMessage, agent.RunConfig{
		StreamingMode: agent.StreamingModeSSE,
	}) {
		if runErr != nil {
			return runErr
		}
		currentText := collectEventText(event)
		if currentText == "" {
			continue
		}
		delta := incrementalText(emittedText, currentText)
		if delta == "" {
			continue
		}
		emittedText += delta
		allText.WriteString(delta)
		writeSSEJSON(c, "model_delta", gin.H{"content": delta})

		snapshot := internalai.ParseStructuredStream(allText.String())
		emitStructuredDiff(c, &lastSnapshot, snapshot)
		lastSnapshot = snapshot
	}

	writeSSEJSON(c, "done", gin.H{"success": true})
	return nil
}

func (s *AIService) StreamReplySuggestion(c *gin.Context, reviewID string) error {
	if strings.TrimSpace(s.cfg.GoogleAPIKey) == "" {
		return fmt.Errorf("GOOGLE_API_KEY is not configured")
	}

	review, err := s.repo.FindByID(c.Request.Context(), reviewID)
	if err != nil {
		return err
	}

	createResp, err := s.replySessions.Create(c.Request.Context(), &session.CreateRequest{
		AppName:   "meituan-review-ai",
		UserID:    "merchant-user",
		SessionID: fmt.Sprintf("reply-%s-%d", reviewID, time.Now().UnixNano()),
	})
	if err != nil {
		return err
	}

	prompt := internalai.BuildReplyPrompt(review)
	userMessage := genai.NewContentFromText(prompt, genai.RoleUser)
	var fullText strings.Builder
	var emittedText string
	for event, runErr := range s.replyRunner.Run(c.Request.Context(), "merchant-user", createResp.Session.ID(), userMessage, agent.RunConfig{
		StreamingMode: agent.StreamingModeSSE,
	}) {
		if runErr != nil {
			return runErr
		}
		currentText := collectEventText(event)
		if currentText == "" {
			continue
		}
		delta := incrementalText(emittedText, currentText)
		if delta == "" {
			continue
		}
		emittedText += delta
		fullText.WriteString(delta)
		writeSSEJSON(c, "reply_delta", gin.H{"content": delta})
	}
	writeSSEJSON(c, "done", gin.H{"success": true, "full_content": strings.TrimSpace(fullText.String())})
	return nil
}

func emitStructuredDiff(c *gin.Context, previous *internalai.StreamSnapshot, current internalai.StreamSnapshot) {
	if current.Summary != "" && current.Summary != previous.Summary {
		writeSSEJSON(c, "summary", gin.H{"content": current.Summary})
	}
	if len(current.PositiveKeywords) > 0 && !reflect.DeepEqual(current.PositiveKeywords, previous.PositiveKeywords) {
		writeSSEJSON(c, "positive_keywords", gin.H{"content": current.PositiveKeywords})
	}
	if len(current.NegativeKeywords) > 0 && !reflect.DeepEqual(current.NegativeKeywords, previous.NegativeKeywords) {
		writeSSEJSON(c, "negative_keywords", gin.H{"content": current.NegativeKeywords})
	}
	if current.SentimentScore != nil {
		if previous.SentimentScore == nil || *current.SentimentScore != *previous.SentimentScore {
			writeSSEJSON(c, "sentiment_score", gin.H{"content": *current.SentimentScore})
		}
	}
	if len(current.Suggestions) > 0 && !reflect.DeepEqual(current.Suggestions, previous.Suggestions) {
		writeSSEJSON(c, "suggestions", gin.H{"content": current.Suggestions})
	}
}

func collectEventText(event *session.Event) string {
	if event == nil || event.Content == nil {
		return ""
	}
	var builder strings.Builder
	for _, part := range event.Content.Parts {
		if part == nil || strings.TrimSpace(part.Text) == "" {
			continue
		}
		builder.WriteString(part.Text)
	}
	return builder.String()
}

func incrementalText(previous string, current string) string {
	if current == "" {
		return ""
	}
	if previous == "" {
		return current
	}
	if strings.HasPrefix(current, previous) {
		return current[len(previous):]
	}
	if current == previous {
		return ""
	}
	return current
}
