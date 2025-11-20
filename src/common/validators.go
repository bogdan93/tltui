package common

import (
	"fmt"
	"regexp"
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

// MinLengthValidator validates that the value meets minimum length
func MinLengthValidator(fieldName string, minLength int) func(string) error {
	return func(value string) error {
		if len(value) < minLength {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must be at least %d characters", fieldName, minLength),
			}
		}
		return nil
	}
}

// LengthRangeValidator validates that the value is within a length range
func LengthRangeValidator(fieldName string, minLength, maxLength int) func(string) error {
	return func(value string) error {
		length := len(value)
		if length < minLength {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must be at least %d characters", fieldName, minLength),
			}
		}
		if length > maxLength {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must not exceed %d characters", fieldName, maxLength),
			}
		}
		return nil
	}
}

// RegexValidator validates that the value matches a regular expression
func RegexValidator(fieldName string, pattern string, errorMessage string) func(string) error {
	re := regexp.MustCompile(pattern)
	return func(value string) error {
		if !re.MatchString(value) {
			if errorMessage == "" {
				errorMessage = fmt.Sprintf("%s has invalid format", fieldName)
			}
			return &ValidationError{
				Field:   fieldName,
				Message: errorMessage,
			}
		}
		return nil
	}
}

// EmailValidator validates that the value looks like an email address
func EmailValidator(fieldName string) func(string) error {
	// Simple email regex - not RFC compliant but good enough for basic validation
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return RegexValidator(fieldName, pattern, fieldName+" must be a valid email address")
}

// NumericValidator validates that the value contains only digits
func NumericValidator(fieldName string) func(string) error {
	return RegexValidator(fieldName, `^\d+$`, fieldName+" must contain only numbers")
}

// AlphanumericValidator validates that the value contains only letters and numbers
func AlphanumericValidator(fieldName string) func(string) error {
	return RegexValidator(fieldName, `^[a-zA-Z0-9]+$`, fieldName+" must contain only letters and numbers")
}

// ChainValidators chains multiple validators together
// All validators must pass for validation to succeed
func ChainValidators(validators ...func(string) error) func(string) error {
	return func(value string) error {
		for _, validator := range validators {
			if err := validator(value); err != nil {
				return err
			}
		}
		return nil
	}
}

// OptionalValidator makes a validator optional (only runs if value is not empty)
func OptionalValidator(validator func(string) error) func(string) error {
	return func(value string) error {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil
		}
		return validator(value)
	}
}
