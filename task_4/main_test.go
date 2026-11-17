package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"task_3/internal/model"
	"task_3/internal/repository"

	"github.com/gofiber/fiber/v3"
)

func request(t *testing.T, app *fiber.App, method, path, body string, status int, target any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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
	return newApp(&fakeBookRepository{}, &fakeUserRepository{})
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

func (r *fakeBookRepository) List(_ context.Context) ([]model.Book, error) {
	return r.books, r.err
}

func TestRepositoryErrors(t *testing.T) {
	app := newApp(&fakeBookRepository{err: errors.New("private database error")}, &fakeUserRepository{})
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
