package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
	"task_3/internal/auth"
	"task_3/internal/model"
	"task_3/internal/repository"
)

type AuthHandler struct {
	users  repository.UserRepository
	tokens *auth.Tokens
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func decodeAuthBody(c fiber.Ctx, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("expected one JSON object")
	}
	return nil
}

func normalizeEmail(email string) (string, bool) {
	email = strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(email)
	return email, err == nil && parsed.Address == email && len(email) <= 254
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeAuthBody(c, &input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid JSON body"})
	}
	input.Name = strings.TrimSpace(input.Name)
	email, valid := normalizeEmail(input.Email)
	if input.Name == "" || !valid {
		return c.Status(400).JSON(fiber.Map{"error": "name and a valid email are required"})
	}
	// bcrypt limits passwords to 72 bytes. Never trim or silently truncate them.
	if utf8.RuneCountInString(input.Password) < 8 || len(input.Password) > 72 {
		return c.Status(400).JSON(fiber.Map{"error": "password must be at least 8 characters and at most 72 bytes"})
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return databaseError(c, err)
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	user, err := h.users.CreateUser(ctx, model.User{Name: input.Name, Email: email, PasswordHash: string(hash)})
	if errors.Is(err, repository.ErrEmailInUse) {
		return c.Status(409).JSON(fiber.Map{"error": "email already registered"})
	}
	if err != nil {
		return databaseError(c, err)
	}
	return c.Status(201).JSON(fiber.Map{"user": user})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var input credentials
	if err := decodeAuthBody(c, &input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid JSON body"})
	}
	email, valid := normalizeEmail(input.Email)
	if !valid || input.Password == "" || len(input.Password) > 72 {
		return c.Status(400).JSON(fiber.Map{"error": "valid email and password are required"})
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	user, err := h.users.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return databaseError(c, err)
	}
	// Compare against a valid dummy hash when no credentials exist, too.
	hash := user.PasswordHash
	if hash == "" {
		hash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	}
	passwordErr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password))
	if err != nil || user.PasswordHash == "" || passwordErr != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid email or password"})
	}
	token, err := h.tokens.Issue(user.ID)
	if err != nil {
		return databaseError(c, err)
	}
	c.Set("Cache-Control", "no-store")
	return c.JSON(fiber.Map{"message": "login successful", "user": user, "access_token": token, "token_type": "Bearer", "expires_in": int(auth.TokenLifetime.Seconds())})
}

func NewAuth(users repository.UserRepository, tokens *auth.Tokens) *AuthHandler {
	return &AuthHandler{users: users, tokens: tokens}
}
