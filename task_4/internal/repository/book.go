package repository

import (
	"context"
	"errors"

	"task_3/internal/model"
)

var ErrBookNotFound = errors.New("book not found")

// BookRepository describes storage operations independently of HTTP handlers.
type BookRepository interface {
	Create(ctx context.Context, book model.Book) (model.Book, error)
	GetByID(ctx context.Context, id int) (model.Book, error)
	List(ctx context.Context) ([]model.Book, error)
}
