package routes

import (
	"bank/internal/handlers"
	"bank/internal/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func Register(app *fiber.App, db *gorm.DB, jwtSecret string) {
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Bank API is running"})
	})
	app.Get("/api/v1/health", handlers.Health(db))
	app.Post("/api/v1/auth/register", handlers.Register(db))
	app.Post("/api/v1/auth/login", handlers.Login(db, jwtSecret))
	app.Get("/api/v1/accounts/:id", handlers.AccountInfo(db))
	app.Post("/api/v1/accounts/deposit", middleware.JWT(jwtSecret), handlers.Deposit(db))
	app.Post("/api/v1/accounts/withdraw", middleware.JWT(jwtSecret), handlers.Withdraw(db))
}
