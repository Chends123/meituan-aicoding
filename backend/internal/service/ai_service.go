package service

import (
	"context"
	"fmt"
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
	writeSSEJSON(c, "meta", gin.H{"review_count": len(reviews), "tab": normalizedTab, "source": "adk-go"})

	var partialBuilder strings.Builder
	var finalText string
	for event, runErr := range s.analysisRunner.Run(c.Request.Context(), "merchant-user", createResp.Session.ID(), userMessage, agent.RunConfig{
		StreamingMode: agent.StreamingModeSSE,
	}) {
		if runErr != nil {
			return runErr
		}
		chunk := collectEventText(event)
		if chunk == "" {
			continue
		}
		if event.Partial {
			partialBuilder.WriteString(chunk)
			writeSSEJSON(c, "model_delta", gin.H{"content": chunk})
			continue
		}
		finalText = chunk
	}

	rawOutput := strings.TrimSpace(finalText)
	if rawOutput == "" {
		rawOutput = strings.TrimSpace(partialBuilder.String())
	}
	parsed, err := internalai.ParseAnalysisOutput(rawOutput)
	if err != nil {
		return fmt.Errorf("parse model analysis output failed: %w; raw=%s", err, rawOutput)
	}

	writeSSEJSON(c, "positive_keywords", gin.H{"content": parsed.PositiveKeywords})
	writeSSEJSON(c, "negative_keywords", gin.H{"content": parsed.NegativeKeywords})
	writeSSEJSON(c, "sentiment_score", gin.H{"content": parsed.SentimentScore})
	writeSSEJSON(c, "suggestions", gin.H{"content": parsed.Suggestions})
	writeSSEJSON(c, "summary", gin.H{"content": parsed.Summary})
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
	var partialBuilder strings.Builder
	var finalText string
	for event, runErr := range s.replyRunner.Run(c.Request.Context(), "merchant-user", createResp.Session.ID(), userMessage, agent.RunConfig{
		StreamingMode: agent.StreamingModeSSE,
	}) {
		if runErr != nil {
			return runErr
		}
		chunk := collectEventText(event)
		if chunk == "" {
			continue
		}
		if event.Partial {
			partialBuilder.WriteString(chunk)
			writeSSEJSON(c, "reply_delta", gin.H{"content": chunk})
			continue
		}
		finalText = chunk
	}

	// 优先使用流式累积内容；若模型未流式输出则用最终完整文本补发一次 delta
	fullContent := strings.TrimSpace(partialBuilder.String())
	if fullContent == "" {
		fullContent = strings.TrimSpace(finalText)
		if fullContent != "" {
			writeSSEJSON(c, "reply_delta", gin.H{"content": fullContent})
		}
	}
	writeSSEJSON(c, "done", gin.H{"success": true, "full_content": fullContent})
	return nil
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
