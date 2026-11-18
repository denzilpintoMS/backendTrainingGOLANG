package handlers

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"bank/internal/models"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// AccountInfo returns an account only when it belongs to the signed-in user.
func AccountInfo(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uint)
		if !ok || userID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		id, err := strconv.ParseUint(c.Params("id"), 10, strconv.IntSize)
		if err != nil || id == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "account id must be a positive integer",
			})
		}

		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		var account models.Account
		err = db.WithContext(ctx).Where("user_id = ?", userID).First(&account, uint(id)).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "account not found",
			})
		}
		if err != nil {
			log.Printf("get account %d: %v", id, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "unable to retrieve account",
			})
		}

		return c.JSON(account)
	}
}
