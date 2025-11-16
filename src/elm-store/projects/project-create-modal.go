package projects

import (
	"strconv"
	"strings"
	"tltui/src/render"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProjectCreateModal struct {
	NameInput    textinput.Model
	OdooIDInput  textinput.Model
	FocusedInput int // 0 = name, 1 = odoo
	ErrorMessage string
}

type ProjectCreatedMsg struct {
	Name   string
	OdooID int
}

type ProjectCreateCanceledMsg struct{}

func NewProjectCreateModal() *ProjectCreateModal {
	nameInput := textinput.New()
	nameInput.Placeholder = "domain.Project Name"
	nameInput.Focus()
	nameInput.CharLimit = 64
	nameInput.Width = 40

	odooIDInput := textinput.New()
	odooIDInput.Placeholder = "Odoo ID"
	odooIDInput.CharLimit = 10
	odooIDInput.Width = 40

	return &ProjectCreateModal{
		NameInput:    nameInput,
		OdooIDInput:  odooIDInput,
		FocusedInput: 0,
	}
}

func (m *ProjectCreateModal) Update(msg tea.Msg) (ProjectCreateModal, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.ErrorMessage = ""

			name := strings.TrimSpace(m.NameInput.Value())
			if name == "" {
				m.ErrorMessage = "domain.Project name is required"
				m.FocusedInput = 0
				m.updateInputFocus()
				return *m, nil
			}

			odooIDStr := strings.TrimSpace(m.OdooIDInput.Value())
			if odooIDStr == "" {
				m.ErrorMessage = "Odoo ID is required"
				m.FocusedInput = 1
				m.updateInputFocus()
				return *m, nil
			}

			odooID, err := strconv.Atoi(odooIDStr)
			if err != nil || odooID <= 0 {
				m.ErrorMessage = "Odoo ID must be a positive number"
				m.FocusedInput = 1
				m.updateInputFocus()
				return *m, nil
			}

			return *m, tea.Batch(
				dispatchCreatedMsg(name, odooID),
			)

		case "esc":
			return *m, tea.Batch(
				dispatchCreateCanceledMsg(),
			)

		case "tab":
			m.FocusedInput = (m.FocusedInput + 1) % 2
			m.updateInputFocus()
			return *m, nil

		case "shift+tab":
			m.FocusedInput = (m.FocusedInput - 1 + 2) % 2
			m.updateInputFocus()
			return *m, nil
		}
	}

	if m.FocusedInput == 0 {
		m.NameInput, cmd = m.NameInput.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.OdooIDInput, cmd = m.OdooIDInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	return *m, tea.Batch(cmds...)
}

func (m *ProjectCreateModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))

	sb.WriteString(titleStyle.Render("Create New domain.Project"))
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("Name:"))
	sb.WriteString("\n")
	sb.WriteString(m.NameInput.View())
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("Odoo ID:"))
	sb.WriteString("\n")
	sb.WriteString(m.OdooIDInput.View())
	sb.WriteString("\n\n")

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
	sb.WriteString(helpStyle.Render("Tab/Shift+Tab: navigate • Enter: create • ESC: cancel"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func dispatchCreatedMsg(name string, odooID int) tea.Cmd {
	return func() tea.Msg {
		return ProjectCreatedMsg{
			Name:   name,
			OdooID: odooID,
		}
	}
}

func dispatchCreateCanceledMsg() tea.Cmd {
	return func() tea.Msg {
		return ProjectCreateCanceledMsg{}
	}
}

func (m *ProjectCreateModal) updateInputFocus() {
	if m.FocusedInput == 0 {
		m.NameInput.Focus()
		m.OdooIDInput.Blur()
	} else {
		m.NameInput.Blur()
		m.OdooIDInput.Focus()
	}
}
