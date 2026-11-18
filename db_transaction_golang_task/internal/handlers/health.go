package handlers

import (
	"context"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"time"
)

func Health(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(ctx) != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "unavailable"})
		}
		return c.JSON(fiber.Map{"status": "ok", "database": "connected"})
	}
}
