package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"bank/internal/middleware"
	"bank/internal/models"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func TestAccountInfo(t *testing.T) {
	db := testDatabase(t)
	user := models.User{Name: "Account Owner", Email: "owner@example.com", Password: "unused"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	account := models.Account{UserID: user.ID, AccountNumber: "owner-account", Balance: 1000, Status: "active"}
	if err := db.Create(&account).Error; err != nil {
		t.Fatal(err)
	}
	secret := "account-test-secret-at-least-32-bytes"
	makeToken := func(userID uint) string {
		t.Helper()
		claims := jwt.RegisteredClaims{Subject: strconv.FormatUint(uint64(userID), 10), Issuer: "bank-api", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	token := makeToken(user.ID)
	otherToken := makeToken(user.ID + 1)
	app := fiber.New()
	app.Get("/accounts/:id", middleware.JWT(secret), AccountInfo(db))
	id := strconv.FormatUint(uint64(account.ID), 10)
	for _, test := range []struct {
		name, id, token string
		status          int
	}{
		{"owner", id, token, http.StatusOK},
		{"missing token", id, "", http.StatusUnauthorized},
		{"invalid token", id, "invalid", http.StatusUnauthorized},
		{"other user", id, otherToken, http.StatusNotFound},
		{"missing account", "999999", token, http.StatusNotFound},
		{"invalid id", "abc", token, http.StatusBadRequest},
		{"zero id", "0", token, http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/accounts/"+test.id, nil)
			req.Header.Set("Authorization", "Bearer "+test.token)
			response, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != test.status {
				t.Fatalf("status = %d, want %d", response.StatusCode, test.status)
			}
			if test.status == http.StatusOK {
				var result models.Account
				if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
					t.Fatal(err)
				}
				if result.ID != account.ID || result.UserID != user.ID || result.Balance != 1000 {
					t.Fatalf("unexpected account: %+v", result)
				}
			}
		})
	}
	if err := db.Delete(&account).Error; err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/accounts/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	response, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("soft-deleted account status = %d, want 404", response.StatusCode)
	}
}
