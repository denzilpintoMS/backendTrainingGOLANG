package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWT(t *testing.T) {
	secret := "test-signing-secret-at-least-32-bytes"
	validClaims := jwt.RegisteredClaims{
		Subject:   "42",
		Issuer:    "bank-api",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	sign := func(claims jwt.RegisteredClaims, method jwt.SigningMethod, key string) string {
		t.Helper()
		token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(key))
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	change := func(edit func(*jwt.RegisteredClaims)) string {
		claims := validClaims
		edit(&claims)
		return sign(claims, jwt.SigningMethodHS256, secret)
	}
	validToken := sign(validClaims, jwt.SigningMethodHS256, secret)
	tests := []struct {
		name   string
		header string
		status int
	}{
		{"valid", "Bearer " + validToken, http.StatusOK},
		{"lowercase bearer", "bearer " + validToken, http.StatusOK},
		{"missing header", "", http.StatusUnauthorized},
		{"missing token", "Bearer", http.StatusUnauthorized},
		{"wrong scheme", "Basic " + validToken, http.StatusUnauthorized},
		{"extra token", "Bearer " + validToken + " extra", http.StatusUnauthorized},
		{"malformed token", "Bearer invalid", http.StatusUnauthorized},
		{"wrong signature", "Bearer " + sign(validClaims, jwt.SigningMethodHS256, "different-signing-secret-at-least-32-bytes"), http.StatusUnauthorized},
		{"wrong algorithm", "Bearer " + sign(validClaims, jwt.SigningMethodHS384, secret), http.StatusUnauthorized},
		{"expired", "Bearer " + change(func(c *jwt.RegisteredClaims) { c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour)) }), http.StatusUnauthorized},
		{"missing expiry", "Bearer " + change(func(c *jwt.RegisteredClaims) { c.ExpiresAt = nil }), http.StatusUnauthorized},
		{"wrong issuer", "Bearer " + change(func(c *jwt.RegisteredClaims) { c.Issuer = "other-api" }), http.StatusUnauthorized},
		{"missing issuer", "Bearer " + change(func(c *jwt.RegisteredClaims) { c.Issuer = "" }), http.StatusUnauthorized},
		{"missing user id", "Bearer " + change(func(c *jwt.RegisteredClaims) { c.Subject = "" }), http.StatusUnauthorized},
		{"invalid user id", "Bearer " + change(func(c *jwt.RegisteredClaims) { c.Subject = "abc" }), http.StatusUnauthorized},
		{"zero user id", "Bearer " + change(func(c *jwt.RegisteredClaims) { c.Subject = "0" }), http.StatusUnauthorized},
		{"not yet valid", "Bearer " + change(func(c *jwt.RegisteredClaims) { c.NotBefore = jwt.NewNumericDate(time.Now().Add(time.Hour)) }), http.StatusUnauthorized},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/protected", JWT(secret), func(c fiber.Ctx) error {
				return c.JSON(fiber.Map{"user_id": c.Locals("user_id")})
			})
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set("Authorization", test.header)
			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != test.status {
				t.Fatalf("status = %d, want %d", response.StatusCode, test.status)
			}
			if test.status == http.StatusOK {
				var body struct {
					UserID uint `json:"user_id"`
				}
				if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.UserID != 42 {
					t.Fatalf("user id = %d, want 42", body.UserID)
				}
			}
		})
	}
}
