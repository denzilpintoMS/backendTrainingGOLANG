package routes

import (
	"bank/internal/handlers"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func Register(app *fiber.App, db *gorm.DB) {
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Bank API is running"})
	})
	app.Get("/api/v1/health", handlers.Health(db))
	app.Post("/api/v1/auth/register", handlers.Register(db))
	app.Get("/api/v1/accounts/:id", handlers.AccountInfo(db))
}
