// Package request centralizes request decoding, normalization and validation.
package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

var validate = newValidator()
var errJSON = errors.New("invalid JSON body")
var errQuery = errors.New("invalid query parameters")

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		return strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
	})
	// Preserve the API's mailbox parsing and bcrypt byte limit. The built-in
	// min tag counts characters, whereas bcrypt limits passwords by bytes.
	if err := v.RegisterValidation("request_email", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		parsed, err := mail.ParseAddress(value)
		return err == nil && parsed.Address == value
	}); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("maxbytes", func(fl validator.FieldLevel) bool {
		limit, err := strconv.Atoi(fl.Param())
		return err == nil && len(fl.Field().String()) <= limit
	}); err != nil {
		panic(err)
	}
	return v
}

// Decode accepts exactly one JSON object, normalizes it, then validates tags.
// Additional request types can implement normalize when normalization is needed.
func Decode(body []byte, target any) error {
	if trimmed := bytes.TrimSpace(body); len(trimmed) == 0 || trimmed[0] != '{' {
		return errJSON
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errJSON
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errJSON
	}
	if input, ok := target.(interface{ normalize() }); ok {
		input.normalize()
	}
	return validate.Struct(target)
}

// DecodeQuery binds query parameters and validates them with the shared validator.
func DecodeQuery(c fiber.Ctx, target any) error {
	if err := c.Bind().Query(target); err != nil {
		return errQuery
	}
	return validate.Struct(target)
}

// PositiveID parses a path parameter and applies the shared positive-ID rule.
func PositiveID(value string) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return id, validate.Var(id, "gt=0")
}

// Error returns field errors without exposing submitted values such as passwords.
func Error(c fiber.Ctx, err error) error {
	var fields validator.ValidationErrors
	if errors.As(err, &fields) {
		messages := make([]string, 0, len(fields))
		for _, field := range fields {
			var message string
			switch field.Tag() {
			case "required":
				message = "is required"
			case "request_email":
				message = "must be a valid email"
			case "min":
				message = "must contain at least " + field.Param() + " characters"
			case "maxbytes":
				message = "must be at most " + field.Param() + " bytes"
			default:
				message = "is invalid"
			}
			messages = append(messages, fmt.Sprintf("%s %s", field.Field(), message))
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": strings.Join(messages, "; ")})
	}
	if errors.Is(err, errJSON) || errors.Is(err, errQuery) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
}
