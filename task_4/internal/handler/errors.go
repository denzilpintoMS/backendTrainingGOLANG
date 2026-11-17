package handler

import (
	"github.com/gofiber/fiber/v3"
	"log"
)

func databaseError(c fiber.Ctx, err error) error {
	log.Printf("request failed: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
}
