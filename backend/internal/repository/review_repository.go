package repository

import (
	"context"
	"strconv"
	"time"

	"gorm.io/gorm"

	"meituan-aicoding/backend/internal/model"
)

type ReviewRepository struct {
	db *gorm.DB
}

type TrendItem struct {
	Date        string  `json:"date"`
	AvgScore    float64 `json:"avg_score"`
	ReviewCount int64   `json:"review_count"`
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) List(ctx context.Context, tab string, cursorTime string, cursorID string, pageSize string) ([]model.Review, bool, map[string]interface{}, error) {
	size, err := strconv.Atoi(pageSize)
	if err != nil || size <= 0 {
		size = 10
	}

	query := r.db.WithContext(ctx).Model(&model.Review{})
	switch tab {
	case "positive":
		query = query.Where("score >= ?", 4)
	case "negative":
		query = query.Where("score <= ?", 2)
	}

	if cursorTime != "" && cursorID != "" {
		parsedTime, parseErr := time.Parse(time.RFC3339, cursorTime)
		if parseErr == nil {
			id, _ := strconv.ParseUint(cursorID, 10, 64)
			query = query.Where("(created_at < ?) OR (created_at = ? AND id < ?)", parsedTime, parsedTime, id)
		}
	}

	var items []model.Review
	if err := query.Order("created_at DESC").Order("id DESC").Limit(size + 1).Find(&items).Error; err != nil {
		return nil, false, nil, err
	}

	hasMore := len(items) > size
	if hasMore {
		items = items[:size]
	}

	nextCursor := map[string]interface{}(nil)
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = map[string]interface{}{
			"cursor_time": last.CreatedAt.Format(time.RFC3339),
			"cursor_id":   last.ID,
		}
	}

	return items, hasMore, nextCursor, nil
}

func (r *ReviewRepository) ListByTab(ctx context.Context, tab string) ([]model.Review, error) {
	query := r.db.WithContext(ctx).Model(&model.Review{})
	switch tab {
	case "positive":
		query = query.Where("score >= ?", 4)
	case "negative":
		query = query.Where("score <= ?", 2)
	}

	var items []model.Review
	err := query.Order("created_at DESC").Order("id DESC").Limit(200).Find(&items).Error
	return items, err
}

func (r *ReviewRepository) FindByID(ctx context.Context, id string) (*model.Review, error) {
	var review model.Review
	if err := r.db.WithContext(ctx).First(&review, id).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) Trends(ctx context.Context, days int) ([]TrendItem, error) {
	type row struct {
		Date        string  `gorm:"column:date"`
		AvgScore    float64 `gorm:"column:avg_score"`
		ReviewCount int64   `gorm:"column:review_count"`
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS date, ROUND(AVG(score), 2) AS avg_score, COUNT(*) AS review_count
			FROM reviews
			WHERE deleted_at IS NULL
			  AND created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
			GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')
			ORDER BY DATE_FORMAT(created_at, '%Y-%m-%d') ASC
		`, days-1).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]TrendItem, 0, len(rows))
	for _, item := range rows {
		result = append(result, TrendItem{
			Date:        item.Date,
			AvgScore:    item.AvgScore,
			ReviewCount: item.ReviewCount,
		})
	}
	return result, nil
}
