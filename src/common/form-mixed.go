package common

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormElement represents any focusable form element
type FormElement interface {
	Focus() tea.Cmd
	Blur()
	Update(msg tea.Msg) tea.Cmd
	View() string
}

// MixedForm manages multiple form elements (fields, checkboxes, etc.) with focus navigation
type MixedForm struct {
	Elements     []FormElement
	FocusedIndex int
	ErrorMessage string
}

// NewMixedForm creates a new form with the given elements
func NewMixedForm(elements ...FormElement) *MixedForm {
	form := &MixedForm{
		Elements:     elements,
		FocusedIndex: 0,
	}

	// Focus the first element
	if len(elements) > 0 {
		elements[0].Focus()
	}

	return form
}

// Update handles common form interactions (tab navigation)
func (f *MixedForm) Update(msg tea.Msg) tea.Cmd {
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

	// Update the focused element
	if f.FocusedIndex >= 0 && f.FocusedIndex < len(f.Elements) {
		cmd = f.Elements[f.FocusedIndex].Update(msg)
	}

	return cmd
}

// NextField moves focus to the next element
func (f *MixedForm) NextField() {
	if len(f.Elements) == 0 {
		return
	}

	f.Elements[f.FocusedIndex].Blur()
	f.FocusedIndex = (f.FocusedIndex + 1) % len(f.Elements)
	f.Elements[f.FocusedIndex].Focus()
}

// PrevField moves focus to the previous element
func (f *MixedForm) PrevField() {
	if len(f.Elements) == 0 {
		return
	}

	f.Elements[f.FocusedIndex].Blur()
	f.FocusedIndex = (f.FocusedIndex - 1 + len(f.Elements)) % len(f.Elements)
	f.Elements[f.FocusedIndex].Focus()
}

// FocusField focuses a specific element by index
func (f *MixedForm) FocusField(index int) {
	if index < 0 || index >= len(f.Elements) {
		return
	}

	f.Elements[f.FocusedIndex].Blur()
	f.FocusedIndex = index
	f.Elements[f.FocusedIndex].Focus()
}

// SetError sets an error message on the form
func (f *MixedForm) SetError(message string) {
	f.ErrorMessage = message
}

// ClearError clears the error message
func (f *MixedForm) ClearError() {
	f.ErrorMessage = ""
}

// GetField returns a FormField by index (if the element is a FormField)
func (f *MixedForm) GetField(index int) *FormField {
	if index < 0 || index >= len(f.Elements) {
		return nil
	}
	if field, ok := f.Elements[index].(*FormField); ok {
		return field
	}
	return nil
}

// GetCheckbox returns a FormCheckbox by index (if the element is a FormCheckbox)
func (f *MixedForm) GetCheckbox(index int) *FormCheckbox {
	if index < 0 || index >= len(f.Elements) {
		return nil
	}
	if checkbox, ok := f.Elements[index].(*FormCheckbox); ok {
		return checkbox
	}
	return nil
}

// Validate validates all FormField elements in the form
func (f *MixedForm) Validate() error {
	f.ErrorMessage = ""

	for i, element := range f.Elements {
		// Only validate FormField elements
		if field, ok := element.(*FormField); ok {
			if err := field.Validate(); err != nil {
				f.ErrorMessage = err.Error()
				f.FocusField(i)
				return err
			}
		}
	}

	return nil
}

// View renders all elements with error message if present
func (f *MixedForm) View() string {
	var output string

	for _, element := range f.Elements {
		output += element.View()
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
