package main

import (
	"log"

	"meituan-aicoding/backend/internal/api/router"
	"meituan-aicoding/backend/internal/config"
	"meituan-aicoding/backend/internal/pkg/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	database, err := db.New(cfg.MySQL)
	if err != nil {
		log.Fatalf("init db failed: %v", err)
	}

	engine, err := router.New(database, cfg)
	if err != nil {
		log.Fatalf("init router failed: %v", err)
	}

	if err := engine.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}
