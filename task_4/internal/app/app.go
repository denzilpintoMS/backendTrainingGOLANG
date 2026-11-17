package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"task_3/internal/config"
	"task_3/internal/repository"
	"task_3/internal/routes"
)

// New builds the HTTP application with the supplied storage dependencies.
func New(books repository.BookRepository, users repository.UserRepository) *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: 1 << 20})
	routes.Register(app, books, users)
	return app
}

// Run loads configuration and manages database and server lifetimes.
func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	books, err := repository.NewPostgres(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		return err
	}
	defer books.Close()

	app := New(books, books)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() { serveErr <- app.Listen(":" + cfg.Port) }()
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
