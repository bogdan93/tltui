package common

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Form manages multiple form fields with focus navigation
type Form struct {
	Fields       []*FormField
	FocusedIndex int
	ErrorMessage string
}

// NewForm creates a new form with the given fields
func NewForm(fields ...*FormField) *Form {
	form := &Form{
		Fields:       fields,
		FocusedIndex: 0,
	}

	// Focus the first field
	if len(fields) > 0 {
		fields[0].Focus()
	}

	return form
}

// Update handles common form interactions (tab navigation)
func (f *Form) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			f.NextField()
			return nil
		case "shift+tab":
			f.PrevField()
			return nil
		}
	}

	// Update the focused field
	if f.FocusedIndex >= 0 && f.FocusedIndex < len(f.Fields) {
		cmd = f.Fields[f.FocusedIndex].Update(msg)
	}

	return cmd
}

// NextField moves focus to the next field
func (f *Form) NextField() {
	if len(f.Fields) == 0 {
		return
	}

	f.Fields[f.FocusedIndex].Blur()
	f.FocusedIndex = (f.FocusedIndex + 1) % len(f.Fields)
	f.Fields[f.FocusedIndex].Focus()
}

// PrevField moves focus to the previous field
func (f *Form) PrevField() {
	if len(f.Fields) == 0 {
		return
	}

	f.Fields[f.FocusedIndex].Blur()
	f.FocusedIndex = (f.FocusedIndex - 1 + len(f.Fields)) % len(f.Fields)
	f.Fields[f.FocusedIndex].Focus()
}

// FocusField focuses a specific field by index
func (f *Form) FocusField(index int) {
	if index < 0 || index >= len(f.Fields) {
		return
	}

	f.Fields[f.FocusedIndex].Blur()
	f.FocusedIndex = index
	f.Fields[f.FocusedIndex].Focus()
}

// Validate validates all fields in the form
func (f *Form) Validate() error {
	f.ErrorMessage = ""

	for i, field := range f.Fields {
		if err := field.Validate(); err != nil {
			f.ErrorMessage = err.Error()
			f.FocusField(i)
			return err
		}
	}

	return nil
}

// SetError sets an error message on the form
func (f *Form) SetError(message string) {
	f.ErrorMessage = message
}

// ClearError clears the error message
func (f *Form) ClearError() {
	f.ErrorMessage = ""
}

// GetValue returns the value of a field by index
func (f *Form) GetValue(index int) string {
	if index < 0 || index >= len(f.Fields) {
		return ""
	}
	return f.Fields[index].Value()
}

// View renders all fields with error message if present
func (f *Form) View() string {
	var output string

	for _, field := range f.Fields {
		output += field.View()
		output += "\n"
	}

	// Show error message if present
	if f.ErrorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		output += errorStyle.Render("⚠ " + f.ErrorMessage)
		output += "\n"
	}

	return output
}
