package handlers

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"bank/internal/models"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required"`
}

func Login(db *gorm.DB, jwtSecret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		var input LoginRequest
		if err := c.Bind().Body(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "provide a valid email and password"})
		}
		email := strings.ToLower(strings.TrimSpace(input.Email))
		var user models.User
		err := db.WithContext(c.Context()).Where("email = ?", email).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid email or password"})
		}
		if err != nil {
			log.Printf("find login user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to log in"})
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid email or password"})
		}

		now := time.Now()
		claims := jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			Issuer:    "bank-api",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSecret))
		if err != nil {
			log.Printf("sign login token: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to log in"})
		}
		return c.JSON(fiber.Map{
			"access_token": token,
			"token_type":   "Bearer",
			"expires_in":   3600,
			"user":         user,
		})
	}
}
