package common

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormCheckbox represents a checkbox form field
type FormCheckbox struct {
	Label      string
	Value      bool
	Focused    bool
	HelpText   string
	LabelStyle lipgloss.Style
}

// NewFormCheckbox creates a new checkbox field
func NewFormCheckbox(label string, initialValue bool) *FormCheckbox {
	return &FormCheckbox{
		Label:      label,
		Value:      initialValue,
		Focused:    false,
		LabelStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("241")),
	}
}

// WithHelpText sets help text for the checkbox
func (c *FormCheckbox) WithHelpText(text string) *FormCheckbox {
	c.HelpText = text
	return c
}

// Focus focuses the checkbox
func (c *FormCheckbox) Focus() tea.Cmd {
	c.Focused = true
	return nil
}

// Blur unfocuses the checkbox
func (c *FormCheckbox) Blur() {
	c.Focused = false
}

// Update handles toggle on space key
func (c *FormCheckbox) Update(msg tea.Msg) tea.Cmd {
	if !c.Focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == " " {
			c.Value = !c.Value
		}
	}

	return nil
}

// Toggle toggles the checkbox value
func (c *FormCheckbox) Toggle() {
	c.Value = !c.Value
}

// View renders the checkbox
func (c *FormCheckbox) View() string {
	var output string

	// Label
	output += c.LabelStyle.Render(c.Label + ":")
	output += "\n"

	// Checkbox
	checkboxStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	if c.Focused {
		checkboxStyle = checkboxStyle.Bold(true).Foreground(lipgloss.Color("39"))
	}

	checkbox := "[ ]"
	if c.Value {
		checkbox = "[✓]"
	}
	output += checkboxStyle.Render(checkbox)

	// Help text
	if c.HelpText != "" {
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		if c.Focused {
			output += " " + c.LabelStyle.Render("(press Space to toggle)")
		} else {
			output += "  " + hintStyle.Render("(" + c.HelpText + ")")
		}
	}

	output += "\n"

	return output
}
