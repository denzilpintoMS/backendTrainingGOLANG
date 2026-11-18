package handlers

import (
	"log"
	"math"
	"strings"

	"bank/internal/models"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransferRequest struct {
	ToAccountNumber string `json:"to_account_number" validate:"required"`
	Amount          int64  `json:"amount" validate:"required,gt=0"` // Amount in paise.
}

func Transfer(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uint)
		if !ok || userID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		var input TransferRequest
		if err := c.Bind().Body(&input); err != nil || input.Amount <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "provide a receiver account number and a positive integer amount in paise"})
		}
		input.ToAccountNumber = strings.TrimSpace(input.ToAccountNumber)
		if input.ToAccountNumber == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "receiver account number is required"})
		}

		tx := db.WithContext(c.Context()).Begin()
		if tx.Error != nil {
			return transferFailure(c, tx.Error)
		}
		defer tx.Rollback()

		// Lock both accounts in ID order, even for transfers in opposite directions.
		var accounts []models.Account
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? OR account_number = ?", userID, input.ToAccountNumber).
			Order("id").Find(&accounts).Error
		if err != nil {
			return transferFailure(c, err)
		}
		var sender, receiver *models.Account
		for i := range accounts {
			if accounts[i].UserID == userID {
				sender = &accounts[i]
			}
			if accounts[i].AccountNumber == input.ToAccountNumber {
				receiver = &accounts[i]
			}
		}
		if sender == nil || receiver == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "sender or receiver account not found"})
		}
		if sender.ID == receiver.ID {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot transfer to the same account"})
		}
		if sender.Status != "active" || receiver.Status != "active" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "both accounts must be active"})
		}
		if sender.Balance < input.Amount {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "insufficient funds"})
		}
		if receiver.Balance > math.MaxInt64-input.Amount {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "transfer exceeds the receiver's maximum balance"})
		}

		sender.Balance -= input.Amount
		receiver.Balance += input.Amount
		if err := tx.Model(sender).Update("balance", sender.Balance).Error; err != nil {
			return transferFailure(c, err)
		}
		if err := tx.Model(receiver).Update("balance", receiver.Balance).Error; err != nil {
			return transferFailure(c, err)
		}
		record := models.Transaction{Type: "transfer", Amount: input.Amount, FromAccountID: &sender.ID, ToAccountID: &receiver.ID}
		if err := tx.Create(&record).Error; err != nil {
			return transferFailure(c, err)
		}
		if err := tx.Commit().Error; err != nil {
			return transferFailure(c, err)
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"account_id":  sender.ID,
			"balance":     sender.Balance,
			"transaction": record,
		})
	}
}

func transferFailure(c fiber.Ctx, err error) error {
	log.Printf("transfer: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to transfer"})
}
