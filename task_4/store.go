package main

import (
	"sort"
	"sync"
)

type Book struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"desc"`
}

// bookStore keeps books in memory until the process exits.
type bookStore struct {
	mu     sync.RWMutex
	books  map[int]Book
	nextID int
}

func newBookStore() *bookStore {
	return &bookStore{books: make(map[int]Book)}
}

func (s *bookStore) create(book Book) Book {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	book.ID = s.nextID
	s.books[book.ID] = book
	return book
}

func (s *bookStore) get(id int) (Book, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	book, ok := s.books[id]
	return book, ok
}

func (s *bookStore) list() []Book {
	s.mu.RLock()
	books := make([]Book, 0, len(s.books))
	for _, book := range s.books {
		books = append(books, book)
	}
	s.mu.RUnlock()

	sort.Slice(books, func(i, j int) bool { return books[i].ID < books[j].ID })
	return books
}
