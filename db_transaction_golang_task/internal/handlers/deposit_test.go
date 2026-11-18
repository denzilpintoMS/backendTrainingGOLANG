package handlers

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"bank/internal/middleware"
	"bank/internal/models"
	"bank/internal/validator"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Use a separate schema so these integration checks never modify application rows.
func TestDeposit(t *testing.T) {
	db := testDatabase(t)

	user := models.User{Name: "Deposit Test", Email: "deposit@example.com", Password: "unused"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	account := models.Account{UserID: user.ID, AccountNumber: "test-account", Status: "active"}
	if err := db.Create(&account).Error; err != nil {
		t.Fatal(err)
	}
	otherUser := models.User{Name: "Other User", Email: "other@example.com", Password: "unused"}
	if err := db.Create(&otherUser).Error; err != nil {
		t.Fatal(err)
	}
	otherAccount := models.Account{UserID: otherUser.ID, AccountNumber: "other-account", Status: "active"}
	if err := db.Create(&otherAccount).Error; err != nil {
		t.Fatal(err)
	}

	secret := "deposit-test-secret-at-least-32-bytes"
	makeToken := func(id uint) string {
		t.Helper()
		claims := jwt.RegisteredClaims{Subject: strconv.FormatUint(uint64(id), 10), Issuer: "bank-api", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	token := makeToken(user.ID)
	app := fiber.New(fiber.Config{StructValidator: validator.New()})
	app.Post("/deposit", middleware.JWT(secret), Deposit(db))
	request := func(payload, token string) (int, error) {
		req := httptest.NewRequest(http.MethodPost, "/deposit", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response, err := app.Test(req)
		if err != nil {
			return 0, err
		}
		defer response.Body.Close()
		return response.StatusCode, nil
	}
	expect := func(t *testing.T, payload, token string, status int) {
		t.Helper()
		got, err := request(payload, token)
		if err != nil {
			t.Fatal(err)
		}
		if got != status {
			t.Fatalf("status = %d, want %d for %s", got, status, payload)
		}
	}
	checkBalance := func(t *testing.T, want int64) {
		t.Helper()
		if err := db.First(&account, account.ID).Error; err != nil {
			t.Fatal(err)
		}
		if account.Balance != want {
			t.Fatalf("balance = %d, want %d", account.Balance, want)
		}
	}

	t.Run("deposit uses authenticated account", func(t *testing.T) {
		payload, _ := json.Marshal(fiber.Map{"amount": 10050, "account_id": otherAccount.ID})
		expect(t, string(payload), token, http.StatusCreated)
		checkBalance(t, 10050)
		var record models.Transaction
		if err := db.First(&record).Error; err != nil {
			t.Fatal(err)
		}
		if record.Type != "deposit" || record.Amount != 10050 || record.FromAccountID != nil || record.ToAccountID == nil || *record.ToAccountID != account.ID {
			t.Fatalf("unexpected record: %+v", record)
		}
		if err := db.First(&otherAccount, otherAccount.ID).Error; err != nil {
			t.Fatal(err)
		}
		if otherAccount.Balance != 0 {
			t.Fatal("changed another user's account")
		}
	})
	t.Run("invalid requests", func(t *testing.T) {
		expect(t, `{"amount":1}`, "", http.StatusUnauthorized)
		expect(t, `{"amount":1}`, "invalid", http.StatusUnauthorized)
		for _, payload := range []string{`{}`, `{"amount":0}`, `{"amount":-1}`, `{"amount":1.5}`, `{"amount":"100"}`, `{"amount":9223372036854775808}`, `{`} {
			expect(t, payload, token, http.StatusBadRequest)
		}
		expect(t, `{"amount":1}`, makeToken(otherUser.ID+1), http.StatusNotFound)
		checkBalance(t, 10050)
	})
	t.Run("inactive and deleted accounts", func(t *testing.T) {
		for _, status := range []string{"frozen", "closed"} {
			if err := db.Model(&account).Update("status", status).Error; err != nil {
				t.Fatal(err)
			}
			expect(t, `{"amount":1}`, token, http.StatusConflict)
		}
		if err := db.Model(&account).Update("status", "active").Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Delete(&otherAccount).Error; err != nil {
			t.Fatal(err)
		}
		expect(t, `{"amount":1}`, makeToken(otherUser.ID), http.StatusNotFound)
		checkBalance(t, 10050)
	})
	t.Run("record failure rolls back balance", func(t *testing.T) {
		if err := db.Exec("ALTER TABLE transactions ADD CONSTRAINT reject_test_deposit CHECK (amount <> 12345)").Error; err != nil {
			t.Fatal(err)
		}
		expect(t, `{"amount":12345}`, token, http.StatusInternalServerError)
		checkBalance(t, 10050)
		var count int64
		if err := db.Model(&models.Transaction{}).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("record count = %d, want 1", count)
		}
	})
	t.Run("concurrent deposits", func(t *testing.T) {
		var wg sync.WaitGroup
		for range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				status, err := request(`{"amount":1}`, token)
				if err != nil || status != http.StatusCreated {
					t.Errorf("concurrent deposit: status=%d error=%v", status, err)
				}
			}()
		}
		wg.Wait()
		checkBalance(t, 10060)
		var count int64
		if err := db.Model(&models.Transaction{}).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 11 {
			t.Fatalf("record count = %d, want 11", count)
		}
	})
	t.Run("balance overflow", func(t *testing.T) {
		if err := db.Model(&account).Update("balance", int64(math.MaxInt64-1)).Error; err != nil {
			t.Fatal(err)
		}
		expect(t, `{"amount":2}`, token, http.StatusBadRequest)
		checkBalance(t, math.MaxInt64-1)
	})
}
