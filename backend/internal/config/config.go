package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App   AppConfig
	MySQL MySQLConfig
	AI    AIConfig
}

type AppConfig struct {
	Name string
	Port string
}

type MySQLConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type AIConfig struct {
	GoogleAPIKey string
	Model        string
}

func Load() (*Config, error) {
	loadEnvFiles()

	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "meituan-review-ai"),
			Port: getEnv("APP_PORT", "8081"),
		},
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "127.0.0.1"),
			Port:     getEnvAsInt("MYSQL_PORT", 3306),
			Database: getEnv("MYSQL_DATABASE", "meituan_review_ai"),
			Username: getEnv("MYSQL_USERNAME", "appuser"),
			Password: getEnv("MYSQL_PASSWORD", "app123456"),
		},
		AI: AIConfig{
			GoogleAPIKey: getEnv("GOOGLE_API_KEY", ""),
			Model:        getEnv("AI_MODEL", "gemini-2.5-flash"),
		},
	}

	if cfg.MySQL.Database == "" {
		return nil, fmt.Errorf("MYSQL_DATABASE is required")
	}

	return cfg, nil
}

func loadEnvFiles() {
	cwd, err := os.Getwd()
	if err != nil {
		_ = godotenv.Load()
		return
	}

	paths := []string{
		filepath.Join(cwd, ".env"),
		filepath.Join(cwd, "..", ".env"),
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Overload(path)
		}
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
