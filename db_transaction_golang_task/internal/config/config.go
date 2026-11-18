package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	return Config{
		Port:        env("PORT", "3000"),
		DatabaseURL: env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/bank_db?sslmode=disable"),
	}, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
