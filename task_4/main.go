package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task_3/internal/repository"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func newApp(books repository.BookRepository, users repository.UserRepository) *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: 1 << 20})
	handler := &bookHandler{books: books}
	auth := &authHandler{users: users}
	api := app.Group("/api/v1")
	api.Post("/auth/register", auth.register)
	api.Post("/auth/login", auth.login)

	api.Post("/books", handler.createBook)
	api.Get("/books/all", handler.listBooks)
	api.Get("/books/:id", handler.getBookByID)

	return app
}

func run() error {
	// Real environment variables take precedence over the optional local file.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return errors.New("could not load .env; check its format and permissions")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required; copy .env.example to .env and configure PostgreSQL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	books, err := repository.NewPostgres(ctx, databaseURL)
	cancel()
	if err != nil {
		return err
	}
	defer books.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	app := newApp(books, books)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() { serveErr <- app.Listen(":" + port) }()
	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			return fmt.Errorf("shut down server: %w", err)
		}
		return <-serveErr
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
