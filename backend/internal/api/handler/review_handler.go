package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"meituan-aicoding/backend/internal/api/dto"
	"meituan-aicoding/backend/internal/service"
)

type ReviewHandler struct {
	reviewService *service.ReviewService
}

func NewReviewHandler(db *gorm.DB) *ReviewHandler {
	return &ReviewHandler{reviewService: service.NewReviewService(db)}
}

func (h *ReviewHandler) List(c *gin.Context) {
	var q dto.ReviewListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	cursorID := ""
	if q.CursorID > 0 {
		cursorID = strconv.FormatUint(q.CursorID, 10)
	}

	resp, err := h.reviewService.List(c.Request.Context(), q.Tab, q.CursorTime, cursorID, strconv.Itoa(pageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ReviewHandler) Trends(c *gin.Context) {
	days, err := strconv.Atoi(c.DefaultQuery("days", "7"))
	if err != nil || days <= 0 {
		days = 7
	}
	resp, err := h.reviewService.Trends(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
