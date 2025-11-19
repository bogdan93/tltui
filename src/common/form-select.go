package common

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectOption represents a single option in a dropdown
type SelectOption struct {
	ID          int
	DisplayName string
	ExtraInfo   string // Optional additional info to display
}

// FormSelect represents a dropdown/list selection component
type FormSelect struct {
	Label         string
	Options       []SelectOption
	SelectedIndex int
	Focused       bool
	Required      bool
	LabelStyle    lipgloss.Style
	ValueStyle    lipgloss.Style
	FocusedStyle  lipgloss.Style
}

// NewFormSelect creates a new select dropdown
func NewFormSelect(label string, options []SelectOption) *FormSelect {
	selectedIndex := -1
	if len(options) > 0 {
		selectedIndex = 0
	}

	return &FormSelect{
		Label:         label,
		Options:       options,
		SelectedIndex: selectedIndex,
		Focused:       false,
		Required:      false,
		LabelStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("241")),
		ValueStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		FocusedStyle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
	}
}

// NewRequiredFormSelect creates a required select dropdown
func NewRequiredFormSelect(label string, options []SelectOption) *FormSelect {
	select_ := NewFormSelect(label, options)
	select_.Required = true
	return select_
}

// SetOptions updates the options list
func (s *FormSelect) SetOptions(options []SelectOption) {
	s.Options = options
	if s.SelectedIndex >= len(options) {
		if len(options) > 0 {
			s.SelectedIndex = 0
		} else {
			s.SelectedIndex = -1
		}
	}
}

// Focus focuses the select
func (s *FormSelect) Focus() tea.Cmd {
	s.Focused = true
	return nil
}

// Blur unfocuses the select
func (s *FormSelect) Blur() {
	s.Focused = false
}

// Update handles up/down arrow navigation
func (s *FormSelect) Update(msg tea.Msg) tea.Cmd {
	if !s.Focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			s.SelectPrevious()
		case "down":
			s.SelectNext()
		}
	}

	return nil
}

// SelectNext selects the next option
func (s *FormSelect) SelectNext() {
	if len(s.Options) == 0 {
		return
	}

	if s.SelectedIndex < 0 {
		s.SelectedIndex = 0
	} else {
		s.SelectedIndex = (s.SelectedIndex + 1) % len(s.Options)
	}
}

// SelectPrevious selects the previous option
func (s *FormSelect) SelectPrevious() {
	if len(s.Options) == 0 {
		return
	}

	if s.SelectedIndex < 0 {
		s.SelectedIndex = len(s.Options) - 1
	} else {
		s.SelectedIndex = (s.SelectedIndex - 1 + len(s.Options)) % len(s.Options)
	}
}

// GetSelectedID returns the ID of the selected option
func (s *FormSelect) GetSelectedID() int {
	if s.SelectedIndex < 0 || s.SelectedIndex >= len(s.Options) {
		return -1
	}
	return s.Options[s.SelectedIndex].ID
}

// GetSelectedOption returns the selected option
func (s *FormSelect) GetSelectedOption() *SelectOption {
	if s.SelectedIndex < 0 || s.SelectedIndex >= len(s.Options) {
		return nil
	}
	return &s.Options[s.SelectedIndex]
}

// Validate validates the selection
func (s *FormSelect) Validate() error {
	if s.Required && (s.SelectedIndex < 0 || s.SelectedIndex >= len(s.Options)) {
		return &ValidationError{Field: s.Label, Message: s.Label + " is required"}
	}
	return nil
}

// View renders the select dropdown
func (s *FormSelect) View() string {
	var output string

	// Label
	output += s.LabelStyle.Render(s.Label + ":")
	output += "\n"

	// Options list
	if len(s.Options) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		output += "  " + emptyStyle.Render("(no options available)")
		output += "\n"
	} else {
		for i, option := range s.Options {
			var style lipgloss.Style
			var prefix string

			if i == s.SelectedIndex {
				if s.Focused {
					style = s.FocusedStyle
					prefix = "▶ "
				} else {
					style = s.ValueStyle
					prefix = "• "
				}
			} else {
				style = s.ValueStyle
				prefix = "  "
			}

			// Display option
			displayText := option.DisplayName
			if option.ExtraInfo != "" {
				displayText = fmt.Sprintf("%s (%s)", option.DisplayName, option.ExtraInfo)
			}

			output += prefix + style.Render(displayText)
			output += "\n"
		}
	}

	return output
}
