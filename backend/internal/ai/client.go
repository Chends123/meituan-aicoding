package ai

import (
	"context"

	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	"meituan-aicoding/backend/internal/config"
)

func NewAnalysisRunner(ctx context.Context, cfg config.AIConfig) (*runner.Runner, session.Service, error) {
	modelInstance, err := gemini.NewModel(ctx, cfg.Model, &genai.ClientConfig{
		APIKey: cfg.GoogleAPIKey,
	})
	if err != nil {
		return nil, nil, err
	}

	agentInstance, err := llmagent.New(llmagent.Config{
		Name:        "review_analysis_agent",
		Description: "Analyze merchant reviews and produce structured JSON output.",
		Model:       modelInstance,
		Instruction: ReviewAnalysisInstruction(),
		GenerateContentConfig: &genai.GenerateContentConfig{
			Temperature: genai.Ptr(float32(0.2)),
		},
	})
	if err != nil {
		return nil, nil, err
	}

	sessionService := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:           "meituan-review-ai",
		Agent:             agentInstance,
		SessionService:    sessionService,
		AutoCreateSession: true,
	})
	if err != nil {
		return nil, nil, err
	}
	return r, sessionService, nil
}

func NewReplyRunner(ctx context.Context, cfg config.AIConfig) (*runner.Runner, session.Service, error) {
	modelInstance, err := gemini.NewModel(ctx, cfg.Model, &genai.ClientConfig{
		APIKey: cfg.GoogleAPIKey,
	})
	if err != nil {
		return nil, nil, err
	}

	agentInstance, err := llmagent.New(llmagent.Config{
		Name:        "review_reply_agent",
		Description: "Generate merchant replies for low-score reviews.",
		Model:       modelInstance,
		Instruction: ReviewReplyInstruction(),
		GenerateContentConfig: &genai.GenerateContentConfig{
			Temperature: genai.Ptr(float32(0.4)),
		},
	})
	if err != nil {
		return nil, nil, err
	}

	sessionService := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:           "meituan-review-ai",
		Agent:             agentInstance,
		SessionService:    sessionService,
		AutoCreateSession: true,
	})
	if err != nil {
		return nil, nil, err
	}
	return r, sessionService, nil
}
