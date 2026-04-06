package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"meituan-aicoding/backend/internal/config"
)

func New(cfg config.MySQLConfig) (*gorm.DB, error) {
	dsn := MustDSN(cfg)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func MustDSN(cfg config.MySQLConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}
