package main

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type bookHandler struct {
	store *bookStore
}

func (h *bookHandler) createBook(c fiber.Ctx) error {
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

	book := h.store.create(Book{
		Name:        input.Name,
		Author:      strings.TrimSpace(input.Author),
		Description: input.Description,
	})
	return c.Status(fiber.StatusCreated).JSON(book)
}

func (h *bookHandler) getBookByID(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id must be a positive integer"})
	}

	book, ok := h.store.get(id)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "book not found"})
	}
	return c.JSON(book)
}

func (h *bookHandler) listBooks(c fiber.Ctx) error {
	books := h.store.list()
	if len(books) == 0 {
		return c.JSON(fiber.Map{"message": "no books available"})
	}
	return c.JSON(books)
}
