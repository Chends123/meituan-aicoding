package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"meituan-aicoding/backend/internal/api/handler"
	"meituan-aicoding/backend/internal/config"
)

func New(db *gorm.DB, cfg *config.Config) (*gin.Engine, error) {
	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	v1 := r.Group("/api/v1")
	{
		reviewHandler := handler.NewReviewHandler(db)
		aiHandler, err := handler.NewAIHandler(db, cfg.AI)
		if err != nil {
			return nil, err
		}

		v1.GET("/reviews", reviewHandler.List)
		v1.GET("/reviews/trends", reviewHandler.Trends)
		v1.GET("/ai/reviews/analyze/stream", aiHandler.AnalyzeReviewsStream)
		v1.GET("/ai/reviews/:id/reply/stream", aiHandler.ReplySuggestionStream)
	}

	return r, nil
}
