package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"task_3/internal/model"
	"task_3/internal/repository"

	"github.com/gofiber/fiber/v3"
)

type BookHandler struct {
	books repository.BookRepository
}

func (h *BookHandler) Create(c fiber.Ctx) error {
	var input struct {
		Name        string `json:"name"`
		Author      string `json:"author"`
		Description string `json:"desc"`
	}
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}
	// Require exactly one JSON value in the request body.
	if err := decoder.Decode(new(any)); err != io.EOF {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	book, err := h.books.Create(ctx, model.Book{
		Name:        input.Name,
		Author:      strings.TrimSpace(input.Author),
		Description: input.Description,
	})
	if err != nil {
		return databaseError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(book)
}

func (h *BookHandler) GetByID(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id must be a positive integer"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	book, err := h.books.GetByID(ctx, id)
	if errors.Is(err, repository.ErrBookNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "book not found"})
	}
	if err != nil {
		return databaseError(c, err)
	}
	return c.JSON(book)
}

func (h *BookHandler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	books, err := h.books.List(ctx)
	if err != nil {
		return databaseError(c, err)
	}
	if len(books) == 0 {
		return c.JSON(fiber.Map{"message": "no books available"})
	}
	return c.JSON(books)
}

func NewBook(books repository.BookRepository) *BookHandler {
	return &BookHandler{books: books}
}
