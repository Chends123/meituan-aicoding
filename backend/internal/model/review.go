package model

import (
	"time"

	"gorm.io/gorm"
)

type Review struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"size:64;not null" json:"username"`
	Score     int            `gorm:"not null" json:"score"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Review) TableName() string {
	return "reviews"
}
