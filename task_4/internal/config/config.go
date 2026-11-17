package config

import (
	"errors"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
}

// Load reads the optional .env file; environment variables take precedence.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, errors.New("could not load .env; check its format and permissions")
	}
	cfg := Config{DatabaseURL: os.Getenv("DATABASE_URL"), Port: os.Getenv("PORT")}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required; copy .env.example to .env and configure PostgreSQL")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	return cfg, nil
}
