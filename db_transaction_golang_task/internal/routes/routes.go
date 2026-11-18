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
	accounts := app.Group("/api/v1/accounts", middleware.JWT(jwtSecret))
	accounts.Get("/:id", handlers.AccountInfo(db))
	accounts.Post("/deposit", handlers.Deposit(db))
	accounts.Post("/withdraw", handlers.Withdraw(db))
	accounts.Post("/transfer", handlers.Transfer(db))
}
