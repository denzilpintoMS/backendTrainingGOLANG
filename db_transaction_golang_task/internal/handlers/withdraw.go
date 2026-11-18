package handlers

import (
	"errors"
	"log"

	"bank/internal/models"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WithdrawRequest struct {
	Amount int64 `json:"amount" validate:"required,gt=0"` // Amount in paise.
}

func Withdraw(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uint)
		if !ok || userID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		var input WithdrawRequest
		if err := c.Bind().Body(&input); err != nil || input.Amount <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "amount must be a positive integer in paise"})
		}

		tx := db.WithContext(c.Context()).Begin()
		if tx.Error != nil {
			return withdrawFailure(c, tx.Error)
		}
		defer tx.Rollback()

		// Lock the account so simultaneous withdrawals cannot spend the same balance.
		var account models.Account
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&account).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "account not found"})
		}
		if err != nil {
			return withdrawFailure(c, err)
		}
		if account.Status != "active" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "account must be active to withdraw"})
		}
		if account.Balance < input.Amount {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "insufficient funds"})
		}

		account.Balance -= input.Amount
		if err := tx.Model(&account).Update("balance", account.Balance).Error; err != nil {
			return withdrawFailure(c, err)
		}
		record := models.Transaction{
			Type:          "withdrawal",
			Amount:        input.Amount,
			FromAccountID: &account.ID,
		}
		if err := tx.Create(&record).Error; err != nil {
			return withdrawFailure(c, err)
		}
		if err := tx.Commit().Error; err != nil {
			return withdrawFailure(c, err)
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"account_id":  account.ID,
			"balance":     account.Balance,
			"transaction": record,
		})
	}
}

func withdrawFailure(c fiber.Ctx, err error) error {
	log.Printf("withdraw: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to withdraw"})
}
