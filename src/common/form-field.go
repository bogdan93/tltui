package common

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormField wraps a text input with label, validation, and common styling
// Implements FormElement interface
type FormField struct {
	Label       string
	Input       textinput.Model
	Required    bool
	Validator   func(string) error // Custom validation function
	HelpText    string             // Optional help text shown below input
	ErrorStyle  lipgloss.Style
	LabelStyle  lipgloss.Style
	Focused     bool
}

// NewFormField creates a new form field with default configuration
func NewFormField(label, placeholder string, width int) FormField {
	input := textinput.New()
	input.Placeholder = placeholder
	input.Width = width
	input.CharLimit = 64

	return FormField{
		Label:      label,
		Input:      input,
		Required:   false,
		ErrorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		LabelStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("241")),
		Focused:    false,
	}
}

// NewRequiredFormField creates a required form field
func NewRequiredFormField(label, placeholder string, width int) FormField {
	field := NewFormField(label, placeholder, width)
	field.Required = true
	return field
}

// WithCharLimit sets the character limit for the input
func (f FormField) WithCharLimit(limit int) FormField {
	f.Input.CharLimit = limit
	return f
}

// WithValidator sets a custom validation function
func (f FormField) WithValidator(validator func(string) error) FormField {
	f.Validator = validator
	return f
}

// WithHelpText sets help text displayed below the input
func (f FormField) WithHelpText(text string) FormField {
	f.HelpText = text
	return f
}

// WithInitialValue sets the initial value of the input
func (f FormField) WithInitialValue(value string) FormField {
	f.Input.SetValue(value)
	return f
}

// Focus focuses the input field
func (f *FormField) Focus() tea.Cmd {
	f.Focused = true
	return f.Input.Focus()
}

// Blur unfocuses the input field
func (f *FormField) Blur() {
	f.Focused = false
	f.Input.Blur()
}

// Update updates the input field
func (f *FormField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.Input, cmd = f.Input.Update(msg)
	return cmd
}

// Validate validates the field value
func (f FormField) Validate() error {
	value := f.Input.Value()

	// Check required
	if f.Required && value == "" {
		return &ValidationError{Field: f.Label, Message: f.Label + " is required"}
	}

	// Run custom validator if provided
	if f.Validator != nil {
		return f.Validator(value)
	}

	return nil
}

// Value returns the trimmed value of the input
func (f FormField) Value() string {
	return f.Input.Value()
}

// View renders the form field with label and optional help text
func (f FormField) View() string {
	var output string

	// Label
	output += f.LabelStyle.Render(f.Label + ":")
	output += "\n"

	// Input
	output += f.Input.View()
	output += "\n"

	// Help text
	if f.HelpText != "" {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		output += helpStyle.Render(f.HelpText)
		output += "\n"
	}

	return output
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
