package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"task_3/internal/model"

	"gorm.io/gorm"
)

// Postgres owns the GORM database and its underlying connection pool.
// HTTP handlers only depend on BookRepository.
type Postgres struct {
	db   *gorm.DB
	pool *sql.DB
}

var _ BookRepository = (*Postgres)(nil)

// NewPostgres opens PostgreSQL and migrates the development models. The
// database itself must already exist. Use versioned migrations for deployment.
func NewPostgres(ctx context.Context, databaseURL string) (*Postgres, error) {
	db, pool, err := openDatabase(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.WithContext(ctx).AutoMigrate(&model.User{}, &model.Book{}); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate models: %w", err)
	}
	return &Postgres{db: db, pool: pool}, nil
}

func (r *Postgres) Close() {
	r.pool.Close()
}

func (r *Postgres) Create(ctx context.Context, book model.Book) (model.Book, error) {
	if err := r.db.WithContext(ctx).Create(&book).Error; err != nil {
		return model.Book{}, fmt.Errorf("create book: %w", err)
	}
	return book, nil
}

func (r *Postgres) GetByID(ctx context.Context, id int) (model.Book, error) {
	var book model.Book
	err := r.db.WithContext(ctx).First(&book, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Book{}, ErrBookNotFound
	}
	if err != nil {
		return model.Book{}, fmt.Errorf("get book: %w", err)
	}
	return book, nil
}

func (r *Postgres) List(ctx context.Context) ([]model.Book, error) {
	books := make([]model.Book, 0)
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&books).Error; err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	return books, nil
}
