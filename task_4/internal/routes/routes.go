package routes

import (
	"github.com/gofiber/fiber/v3"
	"task_3/internal/handler"
	"task_3/internal/repository"
)

// Register mounts all API endpoints under the shared version prefix.
func Register(app *fiber.App, books repository.BookRepository, users repository.UserRepository) {
	book := handler.NewBook(books)
	auth := handler.NewAuth(users)
	api := app.Group("/api/v1")
	api.Post("/auth/register", auth.Register)
	api.Post("/auth/login", auth.Login)
	api.Post("/books", book.Create)
	api.Get("/books/all", book.List)
	api.Get("/books/:id", book.GetByID)
}
