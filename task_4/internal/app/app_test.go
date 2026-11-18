package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	application "task_3/internal/app"
	"testing"
	"time"

	"task_3/internal/auth"
	"task_3/internal/model"
	"task_3/internal/repository"

	"github.com/gofiber/fiber/v3"
)

func request(t *testing.T, app *fiber.App, method, path, body string, status int, target any, authorization ...string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if len(authorization) > 0 {
		if authorization[0] != "" {
			req.Header.Set("Authorization", authorization[0])
		}
	} else if method == http.MethodPost && path == "/api/v1/books" {
		token, err := testTokens().Issue(1)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	// Bcrypt is deliberately expensive, especially under the race detector.
	res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != status {
		t.Fatalf("%s %s: got status %d, want %d", method, path, res.StatusCode, status)
	}
	if target != nil {
		if !strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
			t.Fatalf("unexpected content type: %s", res.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(res.Body).Decode(target); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBookRoutes(t *testing.T) {
	app := newTestApp()
	var empty map[string]string
	request(t, app, http.MethodGet, "/api/v1/books/all", "", http.StatusOK, &empty)
	if empty["message"] != "no books available" {
		t.Fatalf("unexpected empty response: %v", empty)
	}

	want := []model.Book{
		{ID: 1, Name: "First book", Author: "First author", Description: "First description"},
		{ID: 2, Name: "Second book"},
	}
	bodies := []string{
		`{"name":"  First book  ","author":" First author ","desc":"First description"}`,
		`{"name":"Second book"}`,
	}
	for i, body := range bodies {
		var created model.Book
		request(t, app, http.MethodPost, "/api/v1/books", body, http.StatusCreated, &created)
		if created != want[i] {
			t.Fatalf("got %+v, want %+v", created, want[i])
		}
	}

	var got model.Book
	request(t, app, http.MethodGet, "/api/v1/books/1", "", http.StatusOK, &got)
	if got != want[0] {
		t.Fatalf("got %+v, want %+v", got, want[0])
	}

	var books []model.Book
	request(t, app, http.MethodGet, "/api/v1/books/all", "", http.StatusOK, &books)
	if !reflect.DeepEqual(books, want) {
		t.Fatalf("got %v, want %v", books, want)
	}

	// Each test application gets an independent fake repository.
	request(t, newTestApp(), http.MethodGet, "/api/v1/books/1", "", http.StatusNotFound, nil)
}

func TestInvalidBooks(t *testing.T) {
	for _, body := range []string{
		"", "{", "null", `[]`, `{"name":123}`, `{}`, `{"name":"  "}`,
		`{"name":"Book","unknown":true}`, `{"name":"Book"} {}`, `{"name":"Book"} trailing`,
	} {
		t.Run(body, func(t *testing.T) {
			app := newTestApp()
			var response map[string]string
			request(t, app, http.MethodPost, "/api/v1/books", body, http.StatusBadRequest, &response)
			if response["error"] == "" {
				t.Fatal("expected an error message")
			}
			// A rejected request must not consume an ID or create a book.
			var book model.Book
			request(t, app, http.MethodPost, "/api/v1/books", `{"name":"Valid"}`, http.StatusCreated, &book)
			if book.ID != 1 {
				t.Fatalf("invalid request consumed an ID: %d", book.ID)
			}
		})
	}
}

func TestBookIDValidation(t *testing.T) {
	app := newTestApp()
	for _, id := range []string{"abc", "0", "-1", "99999999999999999999999999999999999999"} {
		t.Run(id, func(t *testing.T) {
			request(t, app, http.MethodGet, "/api/v1/books/"+id, "", http.StatusBadRequest, nil)
		})
	}
	request(t, app, http.MethodGet, "/api/v1/books/42", "", http.StatusNotFound, nil)
}

func TestRouting(t *testing.T) {
	app := newTestApp()
	request(t, app, http.MethodPut, "/api/v1/books", "", http.StatusMethodNotAllowed, nil)
	request(t, app, http.MethodPost, "/api/v1/books/1", "", http.StatusMethodNotAllowed, nil)
	request(t, app, http.MethodGet, "/unknown", "", http.StatusNotFound, nil)
}

// This fake keeps handler tests independent of a running PostgreSQL server.
type fakeBookRepository struct {
	books []model.Book
	err   error
}

func newTestApp() *fiber.App {
	return application.New(&fakeBookRepository{}, &fakeUserRepository{}, testTokens())
}

func (r *fakeBookRepository) Create(_ context.Context, book model.Book) (model.Book, error) {
	if r.err != nil {
		return model.Book{}, r.err
	}
	book.ID = len(r.books) + 1
	r.books = append(r.books, book)
	return book, nil
}

func (r *fakeBookRepository) GetByID(_ context.Context, id int) (model.Book, error) {
	if r.err != nil {
		return model.Book{}, r.err
	}
	for _, book := range r.books {
		if book.ID == id {
			return book, nil
		}
	}
	return model.Book{}, repository.ErrBookNotFound
}

func (r *fakeBookRepository) List(_ context.Context, startID, limit int) ([]model.Book, error) {
	if r.err != nil {
		return nil, r.err
	}
	books := make([]model.Book, 0)
	for _, book := range r.books {
		if book.ID >= startID {
			books = append(books, book)
			if len(books) == limit {
				break
			}
		}
	}
	return books, nil
}

func TestRepositoryErrors(t *testing.T) {
	app := application.New(&fakeBookRepository{err: errors.New("private database error")}, &fakeUserRepository{}, testTokens())
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/api/v1/books", `{"name":"Book"}`},
		{http.MethodGet, "/api/v1/books/1", ""},
		{http.MethodGet, "/api/v1/books/all", ""},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			var response map[string]string
			request(t, app, tc.method, tc.path, tc.body, http.StatusInternalServerError, &response)
			if response["error"] != "internal server error" {
				t.Fatalf("unexpected database error response: %v", response)
			}
		})
	}
}

func testTokens() *auth.Tokens {
	tokens, err := auth.New("mindstack")
	if err != nil {
		panic(err)
	}
	return tokens
}

func TestBookPagination(t *testing.T) {
	repo := &fakeBookRepository{}
	for _, id := range []int{1, 3, 4, 7, 9, 10, 12, 15, 18, 20, 22, 25} {
		repo.books = append(repo.books, model.Book{ID: id, Name: "Sample"})
	}
	app := application.New(repo, &fakeUserRepository{}, testTokens())
	for _, tc := range []struct {
		query string
		want  []model.Book
	}{
		{"", repo.books[:10]},
		{"?start_id=1&limit=10", repo.books[:10]},
		{"?start_id=5&limit=2", repo.books[3:5]},
		{"?start_id=21", repo.books[10:]},
		{"?limit=1", repo.books[:1]},
		{"?limit=100", repo.books},
	} {
		t.Run(tc.query, func(t *testing.T) {
			var got []model.Book
			request(t, app, http.MethodGet, "/api/v1/books/all"+tc.query, "", 200, &got)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
	var empty map[string]string
	request(t, app, http.MethodGet, "/api/v1/books/all?start_id=26", "", 200, &empty)
	if empty["message"] != "no books available" {
		t.Fatalf("unexpected empty response: %v", empty)
	}
	for _, query := range []string{
		"?start_id=0", "?start_id=-1", "?start_id=abc", "?start_id=1.5",
		"?start_id=9999999999999999999999999999999",
		"?limit=0", "?limit=-1", "?limit=101", "?limit=abc", "?limit=2.5",
	} {
		t.Run(query, func(t *testing.T) {
			request(t, app, http.MethodGet, "/api/v1/books/all"+query, "", 400, nil)
		})
	}
}
