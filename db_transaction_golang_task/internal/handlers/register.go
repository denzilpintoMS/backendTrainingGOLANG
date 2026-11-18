package handlers

import (
	"errors"
	"log"
	"strings"

	"bank/internal/models"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=8"`
}

func Register(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		var input RegisterRequest
		if err := c.Bind().Body(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "provide a name, valid email, and password of at least 8 characters"})
		}
		input.Name = strings.TrimSpace(input.Name)
		input.Email = strings.ToLower(strings.TrimSpace(input.Email))
		if input.Name == "" || len(input.Password) > 72 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name must not be blank and password must not exceed 72 bytes"})
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("hash password: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to register user"})
		}
		user := models.User{Name: input.Name, Email: input.Email, Password: string(hash)}
		account := models.Account{AccountNumber: uuid.NewString(), Status: "active"}

		// Both records must be created together. Rollback is harmless after Commit.
		tx := db.WithContext(c.Context()).Begin()
		if tx.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to register user"})
		}
		defer tx.Rollback()
		if err := tx.Create(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email is already registered"})
			}
			log.Printf("create user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to register user"})
		}
		account.UserID = user.ID
		if err := tx.Create(&account).Error; err != nil {
			log.Printf("create account: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to register user"})
		}
		if err := tx.Commit().Error; err != nil {
			log.Printf("commit registration: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to register user"})
		}
		user.Account = &account
		return c.Status(fiber.StatusCreated).JSON(user)
	}
}
