package workhour_details

import (
	"strings"
	"tltui/src/render"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkhourDetailsEditModal struct {
	EditingWorkhourDetailID int
	NameInput               textinput.Model
	ShortNameInput          textinput.Model
	IsWork                  bool
	FocusedInput            int // 0 = name, 1 = shortname, 2 = iswork
	ErrorMessage            string
}

type WorkhourDetailsEditedMsg struct {
	WorkhourDetailID int
	Name             string
	ShortName        string
	IsWork           bool
}

type WorkhourDetailsEditCanceledMsg struct{}

func NewWorkhourDetailsEditModal(workhourDetailID int, name string, shortName string, isWork bool) *WorkhourDetailsEditModal {
	nameInput := textinput.New()
	nameInput.Placeholder = "Name"
	nameInput.SetValue(name)
	nameInput.Focus()
	nameInput.CharLimit = 64
	nameInput.Width = 40

	shortNameInput := textinput.New()
	shortNameInput.Placeholder = "Short Name"
	shortNameInput.SetValue(shortName)
	shortNameInput.CharLimit = 20
	shortNameInput.Width = 40

	return &WorkhourDetailsEditModal{
		EditingWorkhourDetailID: workhourDetailID,
		NameInput:               nameInput,
		ShortNameInput:          shortNameInput,
		IsWork:                  isWork,
		FocusedInput:            0,
	}
}

func (m *WorkhourDetailsEditModal) Update(msg tea.Msg) (WorkhourDetailsEditModal, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.ErrorMessage = ""

			name := strings.TrimSpace(m.NameInput.Value())
			if name == "" {
				m.ErrorMessage = "Name is required"
				m.FocusedInput = 0
				m.updateInputFocus()
				return *m, nil
			}

			shortName := strings.TrimSpace(m.ShortNameInput.Value())
			if shortName == "" {
				m.FocusedInput = 0
				m.updateInputFocus()
				m.ErrorMessage = "Short name is required"
				return *m, nil
			}

			return *m, tea.Batch(
				dispatchWorkhourDetailsEditedMsg(m.EditingWorkhourDetailID, name, shortName, m.IsWork),
			)
		case "esc":
			return *m, tea.Batch(
				dispatchWorkhourDetailsEditCanceledMsg(),
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

	// Update text inputs when modal is shown
	if m.FocusedInput == 0 {
		m.NameInput, cmd = m.NameInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.FocusedInput == 1 {
		m.ShortNameInput, cmd = m.ShortNameInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	return *m, tea.Batch(cmds...)
}

func (m *WorkhourDetailsEditModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))

	sb.WriteString(titleStyle.Render("Edit Workhour Detail"))
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
	sb.WriteString(helpStyle.Render("Tab/Shift+Tab: navigate • Space: toggle • Enter: save • ESC: cancel"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func dispatchWorkhourDetailsEditedMsg(workhourDetailID int, name string, shortName string, isWork bool) tea.Cmd {
	return func() tea.Msg {
		return WorkhourDetailsEditedMsg{
			WorkhourDetailID: workhourDetailID,
			Name:             name,
			ShortName:        shortName,
			IsWork:           isWork,
		}
	}
}

func dispatchWorkhourDetailsEditCanceledMsg() tea.Cmd {
	return func() tea.Msg {
		return WorkhourDetailsEditCanceledMsg{}
	}
}

func (m *WorkhourDetailsEditModal) updateInputFocus() {
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
