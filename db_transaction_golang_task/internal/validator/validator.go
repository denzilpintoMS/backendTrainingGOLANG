package validator

import "github.com/go-playground/validator/v10"

// Validator validates request structs using their validate tags.
// Create it once and reuse it across requests.
type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

// Validate implements Fiber's StructValidator interface.
func (v *Validator) Validate(value any) error {
	return v.validate.Struct(value)
}
