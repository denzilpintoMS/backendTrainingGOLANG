package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func request(t *testing.T, app *fiber.App, method, path, body string, status int, target any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
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
	app := newApp()
	var empty map[string]string
	request(t, app, http.MethodGet, "/books/all", "", http.StatusOK, &empty)
	if empty["message"] != "no books available" {
		t.Fatalf("unexpected empty response: %v", empty)
	}

	want := []Book{
		{ID: 1, Name: "First book", Author: "First author", Description: "First description"},
		{ID: 2, Name: "Second book"},
	}
	bodies := []string{
		`{"name":"  First book  ","author":" First author ","desc":"First description"}`,
		`{"name":"Second book"}`,
	}
	for i, body := range bodies {
		var created Book
		request(t, app, http.MethodPost, "/books", body, http.StatusCreated, &created)
		if created != want[i] {
			t.Fatalf("got %+v, want %+v", created, want[i])
		}
	}

	var got Book
	request(t, app, http.MethodGet, "/books/1", "", http.StatusOK, &got)
	if got != want[0] {
		t.Fatalf("got %+v, want %+v", got, want[0])
	}

	var books []Book
	request(t, app, http.MethodGet, "/books/all", "", http.StatusOK, &books)
	if !reflect.DeepEqual(books, want) {
		t.Fatalf("got %v, want %v", books, want)
	}

	// A new application starts with fresh, independent memory storage.
	request(t, newApp(), http.MethodGet, "/books/1", "", http.StatusNotFound, nil)
}

func TestInvalidBooks(t *testing.T) {
	for _, body := range []string{
		"", "{", "null", `[]`, `{"name":123}`, `{}`, `{"name":"  "}`,
		`{"name":"Book","unknown":true}`, `{"name":"Book"} {}`, `{"name":"Book"} trailing`,
	} {
		t.Run(body, func(t *testing.T) {
			app := newApp()
			var response map[string]string
			request(t, app, http.MethodPost, "/books", body, http.StatusBadRequest, &response)
			if response["error"] == "" {
				t.Fatal("expected an error message")
			}
			// A rejected request must not consume an ID or create a book.
			var book Book
			request(t, app, http.MethodPost, "/books", `{"name":"Valid"}`, http.StatusCreated, &book)
			if book.ID != 1 {
				t.Fatalf("invalid request consumed an ID: %d", book.ID)
			}
		})
	}
}

func TestBookIDValidation(t *testing.T) {
	app := newApp()
	for _, id := range []string{"abc", "0", "-1", "99999999999999999999999999999999999999"} {
		t.Run(id, func(t *testing.T) {
			request(t, app, http.MethodGet, "/books/"+id, "", http.StatusBadRequest, nil)
		})
	}
	request(t, app, http.MethodGet, "/books/42", "", http.StatusNotFound, nil)
}

func TestRouting(t *testing.T) {
	app := newApp()
	request(t, app, http.MethodPut, "/books", "", http.StatusMethodNotAllowed, nil)
	request(t, app, http.MethodPost, "/books/1", "", http.StatusMethodNotAllowed, nil)
	request(t, app, http.MethodGet, "/unknown", "", http.StatusNotFound, nil)
}
