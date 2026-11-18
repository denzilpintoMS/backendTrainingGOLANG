package main

import (
	"bank/internal/config"
	"bank/internal/database"
	"bank/internal/routes"
	"bank/internal/validator"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	app := fiber.New(fiber.Config{
		AppName:         "Bank API",
		StructValidator: validator.New(),
	})
	app.Use(recover.New())
	app.Use(logger.New())
	routes.Register(app, db, cfg.JWTSecret)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	go func() {
		<-stop
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Printf("server: %v", err)
	}
}
