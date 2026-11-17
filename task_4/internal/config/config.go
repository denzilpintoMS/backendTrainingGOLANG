package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   string
}

// Load reads the optional .env file; environment variables take precedence.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, errors.New("could not load .env; check its format and permissions")
	}
	cfg := Config{DatabaseURL: os.Getenv("DATABASE_URL"), Port: os.Getenv("PORT"), JWTSecret: os.Getenv("JWT_SECRET")}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required; copy .env.example to .env and configure PostgreSQL")
	}
	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	return cfg, nil
}
