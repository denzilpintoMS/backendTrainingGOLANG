package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestViewAllbooks(t *testing.T) {
	t.Run("empty store", func(t *testing.T) {
		store := &bookStore{books: make(map[int]Book)}
		response := httptest.NewRecorder()
		store.viewAllbooks(response, httptest.NewRequest(http.MethodGet, "/books/all", nil))
		if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected response: %d %v", response.Code, response.Header())
		}
		var body map[string]string
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body["message"] != "no books available" {
			t.Fatalf("unexpected body: %v", body)
		}
	})

	t.Run("all registered books", func(t *testing.T) {
		want := []Book{
			{ID: 1, Name: "First book", Description: "First description"},
			{ID: 2, Name: "Second book", Description: "Second description"},
		}
		store := &bookStore{books: map[int]Book{2: want[1], 1: want[0]}}
		response := httptest.NewRecorder()
		store.viewAllbooks(response, httptest.NewRequest(http.MethodGet, "/books/all", nil))
		if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected response: %d %v", response.Code, response.Header())
		}
		var got []Book
		if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})
}
