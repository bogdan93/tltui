package models

import (
	"strconv"
	"strings"
	"time-logger-tui/src/render"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProjectEditModal struct {
	EditModalContent string
	EditingProjectID int
	NameInput        textinput.Model
	OdooIDInput      textinput.Model
	FocusedInput     int // 0 = name, 1 = odoo
}

type ProjectEditedMsg struct {
	ProjectID int
	Name      string
	OdooID    int
}

type ProjectEditCanceledMsg struct{}

func NewProjectEditModal(projectID int, name string, odooID int) *ProjectEditModal {
	nameInput := textinput.New()
	nameInput.Placeholder = "Project Name"
	nameInput.SetValue(name)
	nameInput.Focus()
	nameInput.CharLimit = 64
	nameInput.Width = 40

	odooIDInput := textinput.New()
	odooIDInput.Placeholder = "Odoo ID"
	odooIDInput.SetValue(strconv.Itoa(odooID))
	odooIDInput.CharLimit = 10
	odooIDInput.Width = 40

	return &ProjectEditModal{
		EditingProjectID: projectID,
		NameInput:        nameInput,
		OdooIDInput:      odooIDInput,
		FocusedInput:     0,
	}
}

func (m *ProjectEditModal) Update(msg tea.Msg) (ProjectEditModal, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			odooID, err := strconv.Atoi(m.OdooIDInput.Value())
			if err != nil {
				odooID = 0
			}
			return *m, tea.Batch(
				dispatchEditedMsg(m.EditingProjectID, m.NameInput.Value(), odooID),
			)
		case "esc":
			return *m, tea.Batch(
				dispatchEditCanceledMsg(),
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

	// Update text inputs when modal is shown
	if m.FocusedInput == 0 {
		m.NameInput, cmd = m.NameInput.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.OdooIDInput, cmd = m.OdooIDInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	return *m, tea.Batch(cmds...)
}

func (m *ProjectEditModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))

	sb.WriteString(titleStyle.Render("Edit Project"))
	sb.WriteString("\n\n")

	// Name input
	sb.WriteString(labelStyle.Render("Name:"))
	sb.WriteString("\n")
	sb.WriteString(m.NameInput.View())
	sb.WriteString("\n\n")

	// Odoo ID input
	sb.WriteString(labelStyle.Render("Odoo ID:"))
	sb.WriteString("\n")
	sb.WriteString(m.OdooIDInput.View())
	sb.WriteString("\n\n")

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
	sb.WriteString(helpStyle.Render("Tab/Shift+Tab: navigate • Enter: save • ESC: cancel"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func dispatchEditedMsg(projectID int, name string, odooID int) tea.Cmd {
	return func() tea.Msg {
		return ProjectEditedMsg{
			ProjectID: projectID,
			Name:      name,
			OdooID:    odooID,
		}
	}
}

func dispatchEditCanceledMsg() tea.Cmd {
	return func() tea.Msg {
		return ProjectEditCanceledMsg{}
	}
}

func (m *ProjectEditModal) updateInputFocus() {
	if m.FocusedInput == 0 {
		m.NameInput.Focus()
		m.OdooIDInput.Blur()
	} else {
		m.NameInput.Blur()
		m.OdooIDInput.Focus()
	}
}
