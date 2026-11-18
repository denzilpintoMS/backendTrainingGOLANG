package handler

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
	"task_3/internal/auth"
	"task_3/internal/model"
	"task_3/internal/repository"
	"task_3/internal/request"
)

type AuthHandler struct {
	users  repository.UserRepository
	tokens *auth.Tokens
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var input request.Register
	if err := request.Decode(c.Body(), &input); err != nil {
		return request.Error(c, err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return databaseError(c, err)
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	user, err := h.users.CreateUser(ctx, model.User{Name: input.Name, Email: input.Email, PasswordHash: string(hash)})
	if errors.Is(err, repository.ErrEmailInUse) {
		return c.Status(409).JSON(fiber.Map{"error": "email already registered"})
	}
	if err != nil {
		return databaseError(c, err)
	}
	return c.Status(201).JSON(fiber.Map{"user": user})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var input request.Login
	if err := request.Decode(c.Body(), &input); err != nil {
		return request.Error(c, err)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	user, err := h.users.GetUserByEmail(ctx, input.Email)
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
