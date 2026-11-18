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

func TestTransfer(t *testing.T) {
	db := testDatabase(t)
	secret := "transfer-test-secret-at-least-32-bytes"
	accounts := make([]models.Account, 2)
	tokens := make([]string, 2)
	for i := range accounts {
		user := models.User{Name: "Transfer Test", Email: "transfer" + strconv.Itoa(i) + "@example.com", Password: "unused"}
		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
		accounts[i] = models.Account{UserID: user.ID, AccountNumber: "transfer-account-" + strconv.Itoa(i), Balance: 10000, Status: "active"}
		if err := db.Create(&accounts[i]).Error; err != nil {
			t.Fatal(err)
		}
		claims := jwt.RegisteredClaims{Subject: strconv.FormatUint(uint64(user.ID), 10), Issuer: "bank-api", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
		if err != nil {
			t.Fatal(err)
		}
		tokens[i] = token
	}
	app := fiber.New(fiber.Config{StructValidator: validator.New()})
	app.Post("/transfer", middleware.JWT(secret), Transfer(db))
	request := func(body, token string) (int, error) {
		req := httptest.NewRequest(http.MethodPost, "/transfer", bytes.NewBufferString(body))
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
	body := func(receiver string, amount int64) string {
		data, _ := json.Marshal(fiber.Map{"to_account_number": receiver, "amount": amount})
		return string(data)
	}
	checkBalances := func(first, second int64) {
		t.Helper()
		for i, want := range []int64{first, second} {
			if err := db.First(&accounts[i], accounts[i].ID).Error; err != nil {
				t.Fatal(err)
			}
			if accounts[i].Balance != want {
				t.Fatalf("account %d balance = %d, want %d", i, accounts[i].Balance, want)
			}
		}
	}

	// Transfer 2500 paise, recording the source and destination accounts.
	expect(body(accounts[1].AccountNumber, 2500), tokens[0], http.StatusCreated)
	checkBalances(7500, 12500)
	var record models.Transaction
	if err := db.First(&record).Error; err != nil {
		t.Fatal(err)
	}
	if record.Type != "transfer" || record.Amount != 2500 || record.FromAccountID == nil || *record.FromAccountID != accounts[0].ID || record.ToAccountID == nil || *record.ToAccountID != accounts[1].ID {
		t.Fatalf("unexpected record: %+v", record)
	}

	// Invalid requests and insufficient funds must leave both accounts unchanged.
	expect(body(accounts[1].AccountNumber, 1), "", http.StatusUnauthorized)
	expect(body(accounts[1].AccountNumber, 10000), tokens[0], http.StatusConflict)
	expect(body(accounts[0].AccountNumber, 1), tokens[0], http.StatusBadRequest)
	expect(body("missing-account", 1), tokens[0], http.StatusNotFound)
	for _, payload := range []string{`{}`, `{"to_account_number":"transfer-account-1","amount":0}`, `{"to_account_number":"transfer-account-1","amount":-1}`, `{"to_account_number":"transfer-account-1","amount":1.5}`, `{"to_account_number":"   ","amount":1}`} {
		expect(payload, tokens[0], http.StatusBadRequest)
	}
	for i := range accounts {
		if err := db.Model(&accounts[i]).Update("status", "frozen").Error; err != nil {
			t.Fatal(err)
		}
		expect(body(accounts[1].AccountNumber, 1), tokens[0], http.StatusConflict)
		if err := db.Model(&accounts[i]).Update("status", "active").Error; err != nil {
			t.Fatal(err)
		}
	}
	checkBalances(7500, 12500)

	// A failed history insert must roll back both balance updates.
	if err := db.Exec("ALTER TABLE transactions ADD CONSTRAINT reject_test_transfer CHECK (amount <> 7000)").Error; err != nil {
		t.Fatal(err)
	}
	expect(body(accounts[1].AccountNumber, 7000), tokens[0], http.StatusInternalServerError)
	checkBalances(7500, 12500)
	var count int64
	if err := db.Model(&models.Transaction{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("record count = %d, want 1", count)
	}

	// A recipient's balance must not overflow.
	if err := db.Model(&accounts[1]).Update("balance", int64(math.MaxInt64)).Error; err != nil {
		t.Fatal(err)
	}
	expect(body(accounts[1].AccountNumber, 1), tokens[0], http.StatusBadRequest)
	checkBalances(7500, math.MaxInt64)
	if err := db.Model(&accounts[1]).Update("balance", int64(12500)).Error; err != nil {
		t.Fatal(err)
	}

	// Concurrent transfers cannot spend more than the sender has.
	var wg sync.WaitGroup
	statuses := make(chan int, 10)
	payload := body(accounts[1].AccountNumber, 1000)
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, err := request(payload, tokens[0])
			if err != nil {
				t.Errorf("transfer: %v", err)
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
		t.Fatalf("successful transfers = %d, want 7", successes)
	}
	checkBalances(500, 19500)

	// Opposite-direction transfers lock accounts in the same order and finish.
	for i := range accounts {
		if err := db.Model(&accounts[i]).Update("balance", int64(10000)).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := range accounts {
		payload := body(accounts[1-i].AccountNumber, 100)
		token := tokens[i]
		for range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				status, err := request(payload, token)
				if err != nil || status != http.StatusCreated {
					t.Errorf("opposite transfer: status=%d error=%v", status, err)
				}
			}()
		}
	}
	wg.Wait()
	checkBalances(10000, 10000)

	// The exact available balance can be transferred, leaving zero.
	expect(body(accounts[1].AccountNumber, 10000), tokens[0], http.StatusCreated)
	checkBalances(0, 20000)
	if err := db.Model(&models.Transaction{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 29 {
		t.Fatalf("record count = %d, want 29", count)
	}
	if err := db.Delete(&accounts[1]).Error; err != nil {
		t.Fatal(err)
	}
	expect(body(accounts[1].AccountNumber, 1), tokens[0], http.StatusNotFound)
}
