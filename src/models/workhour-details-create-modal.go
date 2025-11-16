package models

import (
	"strings"
	"tltui/src/render"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkhourDetailsCreateModal struct {
	NameInput      textinput.Model
	ShortNameInput textinput.Model
	IsWork         bool
	FocusedInput   int // 0 = name, 1 = shortname, 2 = iswork
	ErrorMessage   string
}

type WorkhourDetailsCreatedMsg struct {
	Name      string
	ShortName string
	IsWork    bool
}

type WorkhourDetailsCreateCanceledMsg struct{}

func NewWorkhourDetailsCreateModal() *WorkhourDetailsCreateModal {
	nameInput := textinput.New()
	nameInput.Placeholder = "Name"
	nameInput.Focus()
	nameInput.CharLimit = 64
	nameInput.Width = 40

	shortNameInput := textinput.New()
	shortNameInput.Placeholder = "Short Name"
	shortNameInput.CharLimit = 20
	shortNameInput.Width = 40

	return &WorkhourDetailsCreateModal{
		NameInput:      nameInput,
		ShortNameInput: shortNameInput,
		IsWork:         true, // Default to true for work-related entries
		FocusedInput:   0,
	}
}

func (m *WorkhourDetailsCreateModal) Update(msg tea.Msg) (WorkhourDetailsCreateModal, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Validate inputs
			name := strings.TrimSpace(m.NameInput.Value())
			if name == "" {
				m.ErrorMessage = "Name is required"
				m.FocusedInput = 0
				m.updateInputFocus()
				return *m, nil
			}

			shortName := strings.TrimSpace(m.ShortNameInput.Value())
			if shortName == "" {
				m.FocusedInput = 1
				m.updateInputFocus()
				m.ErrorMessage = "Short name is required"
				return *m, nil
			}

			// Clear error and dispatch
			m.ErrorMessage = ""
			return *m, tea.Batch(
				dispatchWorkhourDetailsCreatedMsg(name, shortName, m.IsWork),
			)

		case "esc":
			return *m, tea.Batch(
				dispatchWorkhourDetailsCreateCanceledMsg(),
			)

		case "tab":
			m.FocusedInput = (m.FocusedInput + 1) % 3
			m.updateInputFocus()
			return *m, nil

		case "shift+tab":
			m.FocusedInput = (m.FocusedInput - 1 + 3) % 3
			m.updateInputFocus()
			return *m, nil

		case " ":
			// Toggle IsWork when focused on the checkbox
			if m.FocusedInput == 2 {
				m.IsWork = !m.IsWork
				return *m, nil
			}
		}
	}

	// Clear error when typing
	m.ErrorMessage = ""

	// Update text inputs
	if m.FocusedInput == 0 {
		m.NameInput, cmd = m.NameInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.FocusedInput == 1 {
		m.ShortNameInput, cmd = m.ShortNameInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	return *m, tea.Batch(cmds...)
}

func (m *WorkhourDetailsCreateModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))

	sb.WriteString(titleStyle.Render("Create New Workhour Detail"))
	sb.WriteString("\n\n")

	// Name input
	sb.WriteString(labelStyle.Render("Name:"))
	sb.WriteString("\n")
	sb.WriteString(m.NameInput.View())
	sb.WriteString("\n\n")

	// Short Name input
	sb.WriteString(labelStyle.Render("Short Name:"))
	sb.WriteString("\n")
	sb.WriteString(m.ShortNameInput.View())
	sb.WriteString("\n")

	// Hint for short name
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	sb.WriteString(hintStyle.Render("Displayed in calendar view • Use emoji only"))
	sb.WriteString("\n\n")

	// IsWork checkbox
	sb.WriteString(labelStyle.Render("Is Work:"))
	sb.WriteString("\n")
	checkboxStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))
	if m.FocusedInput == 2 {
		checkboxStyle = checkboxStyle.
			Bold(true).
			Foreground(lipgloss.Color("39"))
	}
	checkbox := "[ ]"
	if m.IsWork {
		checkbox = "[✓]"
	}
	sb.WriteString(checkboxStyle.Render(checkbox))

	// Hint text - subtle and inline (reuse hintStyle from above)
	hintStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	if m.FocusedInput == 2 {
		sb.WriteString(" " + labelStyle.Render("(press Space to toggle)"))
	} else if m.IsWork {
		// Only show hint when checkbox is checked
		sb.WriteString("  " + hintStyle.Render("(included in mail report)"))
	}
	sb.WriteString("\n\n")

	// Error message
	if m.ErrorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		sb.WriteString(errorStyle.Render("⚠ " + m.ErrorMessage))
		sb.WriteString("\n\n")
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
	sb.WriteString(helpStyle.Render("Tab/Shift+Tab: navigate • Space: toggle • Enter: create • ESC: cancel"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func dispatchWorkhourDetailsCreatedMsg(name string, shortName string, isWork bool) tea.Cmd {
	return func() tea.Msg {
		return WorkhourDetailsCreatedMsg{
			Name:      name,
			ShortName: shortName,
			IsWork:    isWork,
		}
	}
}

func dispatchWorkhourDetailsCreateCanceledMsg() tea.Cmd {
	return func() tea.Msg {
		return WorkhourDetailsCreateCanceledMsg{}
	}
}

func (m *WorkhourDetailsCreateModal) updateInputFocus() {
	if m.FocusedInput == 0 {
		m.NameInput.Focus()
		m.ShortNameInput.Blur()
	} else if m.FocusedInput == 1 {
		m.NameInput.Blur()
		m.ShortNameInput.Focus()
	} else {
		m.NameInput.Blur()
		m.ShortNameInput.Blur()
	}
}
