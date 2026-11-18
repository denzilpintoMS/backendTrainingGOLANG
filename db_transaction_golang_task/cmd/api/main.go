package main

import (
	"bank/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"log"
)

func main() {
	app := fiber.New(fiber.Config{AppName: "Bank API"})
	app.Use(recover.New())
	app.Use(logger.New())
	routes.Register(app)
	if err := app.Listen(":3000"); err != nil {
		log.Printf("server: %v", err)
	}
}
