package common

import (
	"strconv"
	"strings"
)

// PositiveIntValidator validates that the value is a positive integer
func PositiveIntValidator(fieldName string) func(string) error {
	return func(value string) error {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return &ValidationError{Field: fieldName, Message: fieldName + " is required"}
		}

		num, err := strconv.Atoi(trimmed)
		if err != nil {
			return &ValidationError{Field: fieldName, Message: fieldName + " must be a number"}
		}

		if num <= 0 {
			return &ValidationError{Field: fieldName, Message: fieldName + " must be a positive number"}
		}

		return nil
	}
}

// RequiredStringValidator validates that the value is not empty
func RequiredStringValidator(fieldName string) func(string) error {
	return func(value string) error {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return &ValidationError{Field: fieldName, Message: fieldName + " is required"}
		}
		return nil
	}
}

// MaxLengthValidator validates that the value doesn't exceed max length
func MaxLengthValidator(fieldName string, maxLength int) func(string) error {
	return func(value string) error {
		if len(value) > maxLength {
			return &ValidationError{
				Field:   fieldName,
				Message: fieldName + " must not exceed " + strconv.Itoa(maxLength) + " characters",
			}
		}
		return nil
	}
}

// PositiveFloatValidator validates that the value is a positive float
func PositiveFloatValidator(fieldName string) func(string) error {
	return func(value string) error {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return &ValidationError{Field: fieldName, Message: fieldName + " is required"}
		}

		var num float64
		_, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return &ValidationError{Field: fieldName, Message: fieldName + " must be a number"}
		}

		// Parse again to get the actual value
		num, _ = strconv.ParseFloat(trimmed, 64)
		if num <= 0 {
			return &ValidationError{Field: fieldName, Message: fieldName + " must be a positive number"}
		}

		return nil
	}
}
