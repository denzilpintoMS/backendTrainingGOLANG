package handlers

import (
	"bytes"
	"encoding/json"
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

func TestWithdraw(t *testing.T) {
	db := testDatabase(t)
	user := models.User{Name: "Withdrawal Test", Email: "withdraw@example.com", Password: "unused"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	account := models.Account{UserID: user.ID, AccountNumber: "withdraw-test", Balance: 10000, Status: "active"}
	if err := db.Create(&account).Error; err != nil {
		t.Fatal(err)
	}

	secret := "withdraw-test-secret-at-least-32-bytes"
	claims := jwt.RegisteredClaims{Subject: strconv.FormatUint(uint64(user.ID), 10), Issuer: "bank-api", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New(fiber.Config{StructValidator: validator.New()})
	app.Post("/withdraw", middleware.JWT(secret), Withdraw(db))
	request := func(body, token string) (int, error) {
		req := httptest.NewRequest(http.MethodPost, "/withdraw", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		response, err := app.Test(req)
		if err != nil {
			return 0, err
		}
		defer response.Body.Close()
		return response.StatusCode, nil
	}
	expect := func(body, token string, want int) {
		t.Helper()
		status, err := request(body, token)
		if err != nil {
			t.Fatal(err)
		}
		if status != want {
			t.Fatalf("status = %d, want %d for %s", status, want, body)
		}
	}
	checkBalance := func(want int64) {
		t.Helper()
		if err := db.First(&account, account.ID).Error; err != nil {
			t.Fatal(err)
		}
		if account.Balance != want {
			t.Fatalf("balance = %d, want %d", account.Balance, want)
		}
	}

	// Successful withdrawal creates a record with only the source account.
	expect(`{"amount":2500}`, token, http.StatusCreated)
	checkBalance(7500)
	var record models.Transaction
	if err := db.First(&record).Error; err != nil {
		t.Fatal(err)
	}
	if record.Type != "withdrawal" || record.Amount != 2500 || record.FromAccountID == nil || *record.FromAccountID != account.ID || record.ToAccountID != nil {
		t.Fatalf("unexpected transaction: %+v", record)
	}

	// Rejected requests must not change the balance.
	expect(`{"amount":1}`, "", http.StatusUnauthorized)
	for _, body := range []string{`{}`, `{"amount":0}`, `{"amount":-1}`, `{"amount":1.5}`} {
		expect(body, token, http.StatusBadRequest)
	}
	expect(`{"amount":10000}`, token, http.StatusConflict)
	if err := db.Model(&account).Update("status", "frozen").Error; err != nil {
		t.Fatal(err)
	}
	expect(`{"amount":1}`, token, http.StatusConflict)
	if err := db.Model(&account).Update("status", "active").Error; err != nil {
		t.Fatal(err)
	}
	checkBalance(7500)

	// A failed history insert must roll back the balance deduction.
	if err := db.Exec("ALTER TABLE transactions ADD CONSTRAINT reject_test_withdrawal CHECK (amount <> 7000)").Error; err != nil {
		t.Fatal(err)
	}
	expect(`{"amount":7000}`, token, http.StatusInternalServerError)
	checkBalance(7500)

	// Only seven of ten simultaneous 1000-paise withdrawals can succeed.
	var wg sync.WaitGroup
	statuses := make(chan int, 10)
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, err := request(`{"amount":1000}`, token)
			if err != nil {
				t.Errorf("withdraw: %v", err)
			}
			statuses <- status
		}()
	}
	wg.Wait()
	close(statuses)
	successes := 0
	for status := range statuses {
		if status == http.StatusCreated {
			successes++
		} else if status != http.StatusConflict {
			t.Fatalf("unexpected status: %d", status)
		}
	}
	if successes != 7 {
		t.Fatalf("successful withdrawals = %d, want 7", successes)
	}
	checkBalance(500)

	// Withdrawing the exact balance must persist zero.
	expect(`{"amount":500}`, token, http.StatusCreated)
	checkBalance(0)
	var records []models.Transaction
	if err := db.Find(&records).Error; err != nil {
		t.Fatal(err)
	}
	if len(records) != 9 {
		body, _ := json.Marshal(records)
		t.Fatalf("expected 9 records, got %s", body)
	}

	if err := db.Delete(&account).Error; err != nil {
		t.Fatal(err)
	}
	expect(`{"amount":1}`, token, http.StatusNotFound)
}
