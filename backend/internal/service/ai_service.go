package service

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
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
	emitter := newSSEEmitter(c)
	writeSSEJSONWithEmitter(emitter, "meta", gin.H{"review_count": len(reviews), "tab": normalizedTab, "source": "adk-go-structured-stream"})
	stopStatus := startStreamingStatus(c.Request.Context(), emitter, "analysis")
	defer stopStatus()

	var allText strings.Builder
	var emittedText string
	lastSnapshot := internalai.StreamSnapshot{}
	textStreamer := newAnalysisTextStreamer(emitter, "model_delta")
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
		if err := textStreamer.Push(c.Request.Context(), delta); err != nil {
			return err
		}

		snapshot := internalai.ParseStructuredStream(allText.String())
		emitStructuredDiff(emitter, &lastSnapshot, snapshot)
		lastSnapshot = snapshot
	}
	if err := textStreamer.Flush(c.Request.Context()); err != nil {
		return err
	}

	writeSSEJSONWithEmitter(emitter, "done", gin.H{"success": true})
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
	emitter := newSSEEmitter(c)
	stopStatus := startStreamingStatus(c.Request.Context(), emitter, "reply")
	defer stopStatus()

	var fullText strings.Builder
	var emittedText string
	textStreamer := newReplyTextStreamer(emitter, "reply_delta")
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
		if err := textStreamer.Push(c.Request.Context(), delta); err != nil {
			return err
		}
	}
	if err := textStreamer.Flush(c.Request.Context()); err != nil {
		return err
	}

	writeSSEJSONWithEmitter(emitter, "done", gin.H{"success": true, "full_content": strings.TrimSpace(fullText.String())})
	return nil
}

func emitStructuredDiff(emitter *sseEmitter, previous *internalai.StreamSnapshot, current internalai.StreamSnapshot) {
	if current.Summary != "" && current.Summary != previous.Summary {
		writeSSEJSONWithEmitter(emitter, "summary", gin.H{"content": current.Summary})
	}
	if len(current.PositiveKeywords) > 0 && !reflect.DeepEqual(current.PositiveKeywords, previous.PositiveKeywords) {
		writeSSEJSONWithEmitter(emitter, "positive_keywords", gin.H{"content": current.PositiveKeywords})
	}
	if len(current.NegativeKeywords) > 0 && !reflect.DeepEqual(current.NegativeKeywords, previous.NegativeKeywords) {
		writeSSEJSONWithEmitter(emitter, "negative_keywords", gin.H{"content": current.NegativeKeywords})
	}
	if current.SentimentScore != nil {
		if previous.SentimentScore == nil || *current.SentimentScore != *previous.SentimentScore {
			writeSSEJSONWithEmitter(emitter, "sentiment_score", gin.H{"content": *current.SentimentScore})
		}
	}
	if len(current.Suggestions) > 0 && !reflect.DeepEqual(current.Suggestions, previous.Suggestions) {
		writeSSEJSONWithEmitter(emitter, "suggestions", gin.H{"content": current.Suggestions})
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

type sseEmitter struct {
	c  *gin.Context
	mu sync.Mutex
}

func newSSEEmitter(c *gin.Context) *sseEmitter {
	return &sseEmitter{c: c}
}

func (e *sseEmitter) Emit(event string, payload gin.H) error {
	if err := e.c.Request.Context().Err(); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.c.SSEvent(event, payload)
	e.c.Writer.Flush()
	return nil
}

func writeSSEJSONWithEmitter(emitter *sseEmitter, event string, payload gin.H) {
	if emitter == nil {
		return
	}
	_ = emitter.Emit(event, payload)
}

type semanticTextStreamer struct {
	emitter   *sseEmitter
	event     string
	interval  time.Duration
	pending   string
	splitFunc func(string) ([]string, string)
}

func newReplyTextStreamer(emitter *sseEmitter, event string) *semanticTextStreamer {
	return &semanticTextStreamer{
		emitter:   emitter,
		event:     event,
		interval:  120 * time.Millisecond,
		splitFunc: splitReplyText,
	}
}

func newAnalysisTextStreamer(emitter *sseEmitter, event string) *semanticTextStreamer {
	return &semanticTextStreamer{
		emitter:   emitter,
		event:     event,
		interval:  90 * time.Millisecond,
		splitFunc: splitAnalysisText,
	}
}

func (s *semanticTextStreamer) Push(ctx context.Context, text string) error {
	if text == "" {
		return nil
	}
	s.pending += text
	chunks, remaining := s.splitFunc(s.pending)
	s.pending = remaining
	return s.emitChunks(ctx, chunks)
}

func (s *semanticTextStreamer) Flush(ctx context.Context) error {
	remaining := strings.TrimSpace(s.pending)
	if remaining == "" {
		s.pending = ""
		return nil
	}
	chunks := []string{remaining}
	s.pending = ""
	return s.emitChunks(ctx, chunks)
}

func (s *semanticTextStreamer) emitChunks(ctx context.Context, chunks []string) error {
	for index, chunk := range chunks {
		if chunk == "" {
			continue
		}
		if err := s.emitter.Emit(s.event, gin.H{"content": chunk}); err != nil {
			return err
		}
		if index == len(chunks)-1 {
			continue
		}
		if err := sleepWithContext(ctx, s.interval); err != nil {
			return err
		}
	}
	return nil
}

func splitReplyText(text string) ([]string, string) {
	return splitByBoundaries(text, func(r rune) bool {
		switch r {
		case '。', '！', '？', '；', '\n', '!', '?', ';':
			return true
		default:
			return false
		}
	}, 8)
}

func splitAnalysisText(text string) ([]string, string) {
	return splitByBoundaries(text, func(r rune) bool {
		switch r {
		case '\n', '。', '！', '？', '；', '，', '、', ']', '!', '?', ';', ',':
			return true
		default:
			return false
		}
	}, 6)
}

func splitByBoundaries(text string, isBoundary func(rune) bool, minRunes int) ([]string, string) {
	runes := []rune(text)
	chunks := make([]string, 0)
	start := 0
	lastBoundary := -1
	for index, r := range runes {
		if isBoundary(r) {
			lastBoundary = index + 1
			if lastBoundary-start >= minRunes {
				chunks = append(chunks, string(runes[start:lastBoundary]))
				start = lastBoundary
				lastBoundary = -1
			}
		}
	}
	return chunks, string(runes[start:])
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func startStreamingStatus(ctx context.Context, emitter *sseEmitter, mode string) func() {
	messages := []string{"正在等待模型返回内容..."}
	switch mode {
	case "analysis":
		messages = []string{
			"正在补充分析细节...",
			"正在核对关键词和建议...",
			"正在整理最终表述...",
		}
	case "reply":
		messages = []string{
			"正在调整回复语气...",
			"正在补充问题回应...",
			"正在润色最终话术...",
		}
	}

	var once sync.Once
	stop := make(chan struct{})
	send := func(step int) {
		index := step % len(messages)
		_ = emitter.Emit("status", gin.H{
			"mode":    mode,
			"step":    step + 1,
			"message": messages[index],
		})
	}

	go func() {
		ticker := time.NewTicker(1800 * time.Millisecond)
		defer ticker.Stop()
		step := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case <-ticker.C:
				send(step)
				step++
			}
		}
	}()

	return func() {
		once.Do(func() {
			close(stop)
		})
	}
}
