package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Book struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"desc"`
}

type bookStore struct {
	mu     sync.Mutex
	books  map[int]Book
	nextID int
}

func (s *bookStore) createBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Name        string `json:"name"`
		Author      string `json:"author"`
		Description string `json:"desc"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.nextID++
	book := Book{ID: s.nextID, Name: input.Name, Author: strings.TrimSpace(input.Author), Description: input.Description}
	s.books[book.ID] = book
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(book); err != nil {
		log.Printf("encode book response: %v", err)
	}
}

func (s *bookStore) getBookbyID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	log.Printf("ID : %v\n", id)
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "atoi strings error", http.StatusInternalServerError)
		return
	}
	s.mu.Lock()
	book, ok := s.books[idInt]
	s.mu.Unlock()
	if !ok {
		http.Error(w, "book not found", http.StatusNotFound)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(book); err != nil {
			log.Printf("encode book response: %v", err)
		}
	}
}

func (s *bookStore) viewAllbooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	books := make([]Book, 0, len(s.books))
	for _, book := range s.books {
		books = append(books, book)
	}
	s.mu.Unlock()
	sort.Slice(books, func(i, j int) bool { return books[i].ID < books[j].ID })

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if len(books) == 0 {
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "no books available"}); err != nil {
			log.Printf("encode empty books response: %v", err)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(books); err != nil {
		log.Printf("encode book response: %v", err)
	}
}

func main() {
	store := &bookStore{books: make(map[int]Book)}
	mux := http.NewServeMux()
	mux.HandleFunc("/books", store.createBook)
	mux.HandleFunc("/books/{id}", store.getBookbyID)
	mux.HandleFunc("/books/all", store.viewAllbooks)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
