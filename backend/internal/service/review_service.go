package service

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"meituan-aicoding/backend/internal/repository"
)

type ReviewService struct {
	repo *repository.ReviewRepository
}

func NewReviewService(db *gorm.DB) *ReviewService {
	return &ReviewService{repo: repository.NewReviewRepository(db)}
}

func (s *ReviewService) List(ctx context.Context, tab string, cursorTime string, cursorID string, pageSize string) (map[string]interface{}, error) {
	items, hasMore, nextCursor, err := s.repo.List(ctx, normalizeTab(tab), cursorTime, cursorID, pageSize)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"list":        items,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
		"tab":         normalizeTab(tab),
	}, nil
}

func (s *ReviewService) Trends(ctx context.Context, days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 7
	}
	series, err := s.repo.Trends(ctx, days)
	if err != nil {
		return nil, err
	}

	filled := fillTrendGaps(series, days)
	return map[string]interface{}{
		"days":   days,
		"series": filled,
	}, nil
}

func normalizeTab(tab string) string {
	switch strings.ToLower(strings.TrimSpace(tab)) {
	case "positive":
		return "positive"
	case "negative":
		return "negative"
	default:
		return "all"
	}
}

func fillTrendGaps(rows []repository.TrendItem, days int) []repository.TrendItem {
	if days <= 0 {
		days = 7
	}
	lookup := make(map[string]repository.TrendItem, len(rows))
	for _, item := range rows {
		lookup[item.Date] = item
	}

	result := make([]repository.TrendItem, 0, days)
	base := time.Now().In(time.Local)
	for i := days - 1; i >= 0; i-- {
		day := base.AddDate(0, 0, -i).Format("2006-01-02")
		if item, ok := lookup[day]; ok {
			result = append(result, item)
			continue
		}
		result = append(result, repository.TrendItem{
			Date:        day,
			AvgScore:    0,
			ReviewCount: 0,
		})
	}

	return result
}

func writeSSEJSON(c *gin.Context, event string, payload gin.H) {
	c.SSEvent(event, payload)
	c.Writer.Flush()
}
