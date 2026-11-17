package app_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestBookJWTAuthorization(t *testing.T) {
	app := newTestApp()
	// Both reads stay public, even when the caller has no valid credentials.
	request(t, app, http.MethodGet, "/api/v1/books/all", "", 200, nil, "")
	request(t, app, http.MethodGet, "/api/v1/books/1", "", 404, nil, "Bearer invalid")
	for _, header := range []string{"", "Bearer", "Basic abc", "Bearer invalid", "Bearer a b"} {
		request(t, app, http.MethodPost, "/api/v1/books", `{"name":"Blocked"}`, 401, nil, header)
	}
	now := time.Now()
	for _, tc := range []struct {
		name, key, subject, issuer string
		method                     jwt.SigningMethod
		expiry                     *jwt.NumericDate
	}{
		{"expired", "mindstack", "1", "books-api", jwt.SigningMethodHS256, jwt.NewNumericDate(now.Add(-time.Hour))},
		{"wrong key", "another-secret", "1", "books-api", jwt.SigningMethodHS256, jwt.NewNumericDate(now.Add(time.Hour))},
		{"wrong algorithm", "mindstack", "1", "books-api", jwt.SigningMethodHS384, jwt.NewNumericDate(now.Add(time.Hour))},
		{"missing expiry", "mindstack", "1", "books-api", jwt.SigningMethodHS256, nil},
		{"invalid subject", "mindstack", "0", "books-api", jwt.SigningMethodHS256, jwt.NewNumericDate(now.Add(time.Hour))},
		{"wrong issuer", "mindstack", "1", "other-app", jwt.SigningMethodHS256, jwt.NewNumericDate(now.Add(time.Hour))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token, err := jwt.NewWithClaims(tc.method, jwt.RegisteredClaims{Subject: tc.subject, Issuer: tc.issuer, ExpiresAt: tc.expiry}).SignedString([]byte(tc.key))
			if err != nil {
				t.Fatal(err)
			}
			request(t, app, http.MethodPost, "/api/v1/books", `{"name":"Blocked"}`, 401, nil, "Bearer "+token)
		})
	}
	var empty map[string]string
	request(t, app, http.MethodGet, "/api/v1/books/all", "", 200, &empty)
	if empty["message"] != "no books available" {
		t.Fatal("unauthorized request created a book")
	}
	token, err := testTokens().Issue(1)
	if err != nil {
		t.Fatal(err)
	}
	request(t, app, http.MethodPost, "/api/v1/books", `{"name":"Allowed"}`, 201, nil, "Bearer "+token)
	request(t, app, http.MethodGet, "/api/v1/books/1", "", 200, nil, "")
}
