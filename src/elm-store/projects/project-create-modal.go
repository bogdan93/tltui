package projects

import (
	"strconv"
	"strings"
	"tltui/src/common"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProjectCreateModal struct {
	Form *common.Form
}

type ProjectCreatedMsg struct {
	Name   string
	OdooID int
}

type ProjectCreateCanceledMsg struct{}

func NewProjectCreateModal() *ProjectCreateModal {
	nameField := common.NewRequiredFormField("Name", "Project Name", 40)
	odooIDField := common.NewRequiredFormField("Odoo ID", "Odoo ID", 40).
		WithCharLimit(10).
		WithValidator(common.PositiveIntValidator("Odoo ID"))

	form := common.NewForm(&nameField, &odooIDField)

	return &ProjectCreateModal{
		Form: form,
	}
}

func (m *ProjectCreateModal) Update(msg tea.Msg) (ProjectCreateModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if err := m.Form.Validate(); err != nil {
				return *m, nil
			}

			name := strings.TrimSpace(m.Form.GetValue(0))
			odooIDStr := strings.TrimSpace(m.Form.GetValue(1))
			odooID, _ := strconv.Atoi(odooIDStr) // Already validated

			return *m, tea.Batch(
				dispatchCreatedMsg(name, odooID),
			)

		case "esc":
			return *m, tea.Batch(
				dispatchCreateCanceledMsg(),
			)
		}
	}

	cmd := m.Form.Update(msg)
	return *m, cmd
}

func (m *ProjectCreateModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	sb.WriteString(titleStyle.Render("Create New Project"))
	sb.WriteString("\n\n")

	sb.WriteString(m.Form.View())

	sb.WriteString(render.RenderHelpText("Tab/Shift+Tab: navigate", "Enter: create", "ESC: cancel"))

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
