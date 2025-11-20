package common

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FormField struct {
	Label      string
	Input      textinput.Model
	Required   bool
	Validator  func(string) error
	HelpText   string
	ErrorStyle lipgloss.Style
	LabelStyle lipgloss.Style
	Focused    bool
}

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

func NewRequiredFormField(label, placeholder string, width int) FormField {
	field := NewFormField(label, placeholder, width)
	field.Required = true
	return field
}

func (f FormField) WithCharLimit(limit int) FormField {
	f.Input.CharLimit = limit
	return f
}

func (f FormField) WithValidator(validator func(string) error) FormField {
	f.Validator = validator
	return f
}

func (f FormField) WithHelpText(text string) FormField {
	f.HelpText = text
	return f
}

func (f FormField) WithInitialValue(value string) FormField {
	f.Input.SetValue(value)
	return f
}

func (f *FormField) Focus() tea.Cmd {
	f.Focused = true
	return f.Input.Focus()
}

func (f *FormField) Blur() {
	f.Focused = false
	f.Input.Blur()
}

func (f *FormField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.Input, cmd = f.Input.Update(msg)
	return cmd
}

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

func (f FormField) Value() string {
	return f.Input.Value()
}

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

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
