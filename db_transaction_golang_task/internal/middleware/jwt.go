package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWT validates a Bearer token and stores the authenticated user ID in Locals.
// Attach it to individual routes or groups when they need authentication.
func JWT(secret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := strings.Fields(c.Get("Authorization"))
		if len(secret) < 32 || len(header) != 2 || !strings.EqualFold(header[0], "Bearer") {
			return unauthorized(c)
		}

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(header[1], claims, func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuer("bank-api"))
		if err != nil || !token.Valid {
			return unauthorized(c)
		}

		userID, err := strconv.ParseUint(claims.Subject, 10, strconv.IntSize)
		if err != nil || userID == 0 {
			return unauthorized(c)
		}

		c.Locals("user_id", uint(userID))
		return c.Next()
	}
}

func unauthorized(c fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
}
