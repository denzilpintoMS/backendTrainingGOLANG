package routes

import (
	"github.com/gofiber/fiber/v3"
	"task_3/internal/auth"
	"task_3/internal/handler"
	"task_3/internal/middleware"
	"task_3/internal/repository"
)

// Register mounts all API endpoints under the shared version prefix.
func Register(app *fiber.App, books repository.BookRepository, users repository.UserRepository, tokens *auth.Tokens) {
	book := handler.NewBook(books)
	authHandler := handler.NewAuth(users, tokens)
	api := app.Group("/api/v1")
	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/login", authHandler.Login)
	api.Post("/books", middleware.RequireJWT(tokens), book.Create)
	api.Get("/books/all", book.List)
	api.Get("/books/:id", book.GetByID)
}
