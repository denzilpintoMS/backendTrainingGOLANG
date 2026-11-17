package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

func newApp() *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: 1 << 20})
	handler := &bookHandler{store: newBookStore()}

	app.Post("/books", handler.createBook)
	app.Get("/books/all", handler.listBooks)
	app.Get("/books/:id", handler.getBookByID)

	return app
}

func main() {
	log.Fatal(newApp().Listen(":8080"))
}
