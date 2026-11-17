package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"task_3/internal/model"
	"task_3/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run against PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	// Use a unique schema so tests never modify the application's books table.
	schemaName := "books_test_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	quotedSchema := pgx.Identifier{schemaName}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := admin.Exec(cleanupCtx, "DROP SCHEMA "+quotedSchema+" CASCADE"); err != nil {
			t.Errorf("clean up test schema: %v", err)
		}
	}()
	u, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", schemaName)
	u.RawQuery = query.Encode()
	testURL := u.String()
	// Exercise migration from the original SQL schema as well as fresh tables.
	if _, err := admin.Exec(ctx, `CREATE TABLE `+quotedSchema+`.books (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		name TEXT NOT NULL CHECK (length(trim(name)) > 0),
		author TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT ''
	)`); err != nil {
		t.Fatal(err)
	}

	books, err := repository.NewPostgres(ctx, testURL)
	if err != nil {
		t.Fatal(err)
	}
	// Capturing the variable also closes the replacement pool after reopening.
	defer func() { books.Close() }()
	app := newApp(books, books)
	var empty map[string]string
	request(t, app, http.MethodGet, "/api/v1/books/all", "", http.StatusOK, &empty)
	if empty["message"] != "no books available" {
		t.Fatalf("unexpected empty response: %v", empty)
	}
	want := []model.Book{
		{ID: 1, Name: "First book", Author: "An author", Description: "An author's description"},
		{ID: 2, Name: "Robert'); DROP TABLE books;--"},
	}
	for i, body := range []string{
		`{"name":" First book ","author":" An author ","desc":"An author's description"}`,
		`{"name":"Robert'); DROP TABLE books;--"}`,
	} {
		var book model.Book
		request(t, app, http.MethodPost, "/api/v1/books", body, http.StatusCreated, &book)
		if book != want[i] {
			t.Fatalf("got %+v, want %+v", book, want[i])
		}
	}

	books.Close()
	reopened, err := repository.NewPostgres(ctx, testURL)
	if err != nil {
		t.Fatal(err)
	}
	books = reopened
	app = newApp(books, books)
	var got model.Book
	request(t, app, http.MethodGet, "/api/v1/books/1", "", http.StatusOK, &got)
	if got != want[0] {
		t.Fatalf("book did not persist: got %+v, want %+v", got, want[0])
	}
	var list []model.Book
	request(t, app, http.MethodGet, "/api/v1/books/all", "", http.StatusOK, &list)
	if !reflect.DeepEqual(list, want) {
		t.Fatalf("got %v, want %v", list, want)
	}
	stored, err := books.GetByID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if stored.CreatedAt.IsZero() || stored.UpdatedAt.IsZero() {
		t.Fatalf("GORM did not populate book timestamps: %+v", stored)
	}
	request(t, app, http.MethodGet, "/api/v1/books/99", "", http.StatusNotFound, nil)

	// The User model must create its table and enforce its database constraints.
	insertUser := `INSERT INTO ` + quotedSchema + `.users (name, email) VALUES ($1, $2)`
	if _, err := admin.Exec(ctx, insertUser, "Alice", "alice@example.com"); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, email, code string }{
		{"Another Alice", "ALICE@example.com", "23505"},
		{" ", "valid@example.com", "23514"},
		{"Bob", " ", "23514"},
	} {
		_, err := admin.Exec(ctx, insertUser, tc.name, tc.email)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != tc.code {
			t.Fatalf("expected constraint error %s, got %v", tc.code, err)
		}
	}

	// Exercise registration, duplicate handling, and bcrypt login through GORM.
	request(t, app, http.MethodPost, "/api/v1/auth/register", `{"name":"Reader","email":"reader@example.com","password":"secret123"}`, http.StatusCreated, nil)
	request(t, app, http.MethodPost, "/api/v1/auth/register", `{"name":"Reader","email":"READER@example.com","password":"secret123"}`, http.StatusConflict, nil)
	request(t, app, http.MethodPost, "/api/v1/auth/login", `{"email":"READER@example.com","password":"secret123"}`, http.StatusOK, nil)
	request(t, app, http.MethodPost, "/api/v1/auth/login", `{"email":"reader@example.com","password":"wrong"}`, http.StatusUnauthorized, nil)

	// A canceled operation must return an error rather than wait indefinitely.
	canceledCtx, cancelOperation := context.WithCancel(context.Background())
	cancelOperation()
	if _, err := books.List(canceledCtx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
