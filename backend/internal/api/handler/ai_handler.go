package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"meituan-aicoding/backend/internal/config"
	"meituan-aicoding/backend/internal/service"
)

type AIHandler struct {
	aiService *service.AIService
}

func NewAIHandler(db *gorm.DB, cfg config.AIConfig) (*AIHandler, error) {
	svc, err := service.NewAIService(db, cfg)
	if err != nil {
		return nil, err
	}
	return &AIHandler{aiService: svc}, nil
}

func (h *AIHandler) AnalyzeReviewsStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	if err := h.aiService.StreamReviewAnalysis(c, c.Query("tab")); err != nil {
		c.SSEvent("error", gin.H{"message": err.Error()})
		c.Writer.Flush()
	}
}

func (h *AIHandler) ReplySuggestionStream(c *gin.Context) {
	idStr := c.Param("id")
	if _, err := strconv.ParseUint(idStr, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid review id"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	if err := h.aiService.StreamReplySuggestion(c, idStr); err != nil {
		c.SSEvent("error", gin.H{"message": err.Error()})
		c.Writer.Flush()
	}
}
