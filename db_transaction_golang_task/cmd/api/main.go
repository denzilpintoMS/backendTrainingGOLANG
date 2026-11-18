package main

import (
	"bank/internal/config"
	"bank/internal/database"
	"bank/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()
	app := fiber.New(fiber.Config{AppName: "Bank API"})
	app.Use(recover.New())
	app.Use(logger.New())
	routes.Register(app, db)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Printf("server: %v", err)
	}
}
