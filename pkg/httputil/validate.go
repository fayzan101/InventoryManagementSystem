package httputil

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func Required(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: field, Message: "is required"}
	}
	return nil
}

func PositiveInt(field string, value int) error {
	if value <= 0 {
		return ValidationError{Field: field, Message: "must be greater than zero"}
	}
	return nil
}

func NonNegativeFloat(field string, value float64) error {
	if value < 0 {
		return ValidationError{Field: field, Message: "must not be negative"}
	}
	return nil
}

func OneOf(field, value string, allowed ...string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return ValidationError{Field: field, Message: fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", "))}
}

func MinLen(field, value string, min int) error {
	if len(strings.TrimSpace(value)) < min {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be at least %d characters", min)}
	}
	return nil
}
