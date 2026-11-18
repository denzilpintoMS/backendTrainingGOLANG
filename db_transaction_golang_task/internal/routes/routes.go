package routes

import "github.com/gofiber/fiber/v3"

func Register(app *fiber.App) {
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Bank API is running"})
	})
}
