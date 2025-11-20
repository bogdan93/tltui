package common

import (
	"testing"
)

func TestMinLengthValidator(t *testing.T) {
	validator := MinLengthValidator("Username", 3)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid length", "abc", false},
		{"longer than min", "abcdef", false},
		{"too short", "ab", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MinLengthValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMaxLengthValidator(t *testing.T) {
	validator := MaxLengthValidator("Field", 10)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "hello", false},
		{"exact max", "1234567890", false},
		{"too long", "12345678901", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MaxLengthValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLengthRangeValidator(t *testing.T) {
	validator := LengthRangeValidator("Field", 3, 10)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "hello", false},
		{"exact min", "abc", false},
		{"exact max", "1234567890", false},
		{"too short", "ab", true},
		{"too long", "12345678901", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("LengthRangeValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChainValidators(t *testing.T) {
	validator := ChainValidators(
		MinLengthValidator("Field", 3),
		MaxLengthValidator("Field", 10),
	)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "hello", false},
		{"too short", "hi", true},
		{"too long", "this is way too long", true},
		{"exact min", "abc", false},
		{"exact max", "1234567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChainValidators() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailValidator(t *testing.T) {
	validator := EmailValidator("Email")

	tests := []struct {
		input   string
		wantErr bool
	}{
		{"test@example.com", false},
		{"user+tag@domain.co.uk", false},
		{"invalid", true},
		{"@example.com", true},
		{"test@", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailValidator(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestNumericValidator(t *testing.T) {
	validator := NumericValidator("Field")

	tests := []struct {
		input   string
		wantErr bool
	}{
		{"123", false},
		{"0", false},
		{"abc", true},
		{"12a", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NumericValidator(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestAlphanumericValidator(t *testing.T) {
	validator := AlphanumericValidator("Field")

	tests := []struct {
		input   string
		wantErr bool
	}{
		{"abc123", false},
		{"ABC", false},
		{"123", false},
		{"abc-123", true},
		{"hello world", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AlphanumericValidator(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestOptionalValidator(t *testing.T) {
	validator := OptionalValidator(MinLengthValidator("Field", 5))

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty is valid (optional)", "", false},
		{"whitespace is valid (optional)", "   ", false},
		{"valid length", "hello", false},
		{"too short", "hi", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("OptionalValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPositiveIntValidator(t *testing.T) {
	validator := PositiveIntValidator("Field")

	tests := []struct {
		input   string
		wantErr bool
	}{
		{"1", false},
		{"100", false},
		{"0", true},
		{"-1", true},
		{"abc", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PositiveIntValidator(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestPositiveFloatValidator(t *testing.T) {
	validator := PositiveFloatValidator("Field")

	tests := []struct {
		input   string
		wantErr bool
	}{
		{"1.0", false},
		{"8.5", false},
		{"0.1", false},
		{"0", true},
		{"-1.5", true},
		{"abc", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PositiveFloatValidator(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestRequiredStringValidator(t *testing.T) {
	validator := RequiredStringValidator("Field")

	tests := []struct {
		input   string
		wantErr bool
	}{
		{"hello", false},
		{"a", false},
		{"", true},
		{"   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequiredStringValidator(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
