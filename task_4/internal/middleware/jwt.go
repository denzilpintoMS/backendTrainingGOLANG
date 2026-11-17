package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"task_3/internal/auth"
)

// RequireJWT accepts only signed, unexpired bearer access tokens.
func RequireJWT(tokens *auth.Tokens) fiber.Handler {
	return func(c fiber.Ctx) error {
		parts := strings.Fields(c.Get("Authorization"))
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if id, err := tokens.Verify(parts[1]); err == nil {
				c.Locals("user_id", id)
				return c.Next()
			}
		}
		c.Set("WWW-Authenticate", "Bearer")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "valid bearer token required"})
	}
}
