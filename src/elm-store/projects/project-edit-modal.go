package projects

import (
	"strconv"
	"strings"
	"tltui/src/common"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProjectEditModal struct {
	EditingProjectID int
	Form             *common.MixedForm
}

type ProjectEditedMsg struct {
	ProjectID int
	Name      string
	OdooID    int
}

type ProjectEditCanceledMsg struct{}

func NewProjectEditModal(projectID int, name string, odooID int) *ProjectEditModal {
	nameField := common.NewRequiredFormField("Name", "Project Name", 40).
		WithInitialValue(name).
		WithValidator(common.ChainValidators(
			common.MinLengthValidator("Name", 2),
			common.MaxLengthValidator("Name", 50),
		))

	odooIDField := common.NewRequiredFormField("Odoo ID", "Odoo ID", 40).
		WithCharLimit(10).
		WithValidator(common.PositiveIntValidator("Odoo ID")).
		WithInitialValue(strconv.Itoa(odooID))

	form := common.NewMixedForm(&nameField, &odooIDField)

	return &ProjectEditModal{
		EditingProjectID: projectID,
		Form:             form,
	}
}

func (m *ProjectEditModal) Update(msg tea.Msg) (ProjectEditModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if err := m.Form.Validate(); err != nil {
				return *m, nil
			}

			name := strings.TrimSpace(m.Form.GetField(0).Value())
			odooIDStr := strings.TrimSpace(m.Form.GetField(1).Value())
			odooID, _ := strconv.Atoi(odooIDStr) 

			return *m, tea.Batch(
				dispatchEditedMsg(m.EditingProjectID, name, odooID),
			)

		case "esc":
			return *m, tea.Batch(
				dispatchEditCanceledMsg(),
			)
		}
	}

	cmd := m.Form.Update(msg)
	return *m, cmd
}

func (m *ProjectEditModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	sb.WriteString(titleStyle.Render("Edit Project"))
	sb.WriteString("\n\n")

	sb.WriteString(m.Form.View())

	sb.WriteString(render.RenderHelpText("Tab/Shift+Tab: navigate", "Enter: save", "ESC: cancel"))

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
