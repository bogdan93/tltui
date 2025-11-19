package workhour_details

import (
	"strings"
	"tltui/src/common"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkhourDetailsCreateModal struct {
	Form *common.MixedForm
}

type WorkhourDetailsCreatedMsg struct {
	Name      string
	ShortName string
	IsWork    bool
}

type WorkhourDetailsCreateCanceledMsg struct{}

func NewWorkhourDetailsCreateModal() *WorkhourDetailsCreateModal {
	nameField := common.NewRequiredFormField("Name", "Name", 40)
	shortNameField := common.NewRequiredFormField("Short Name", "Short Name", 40).
		WithCharLimit(20).
		WithHelpText("Displayed in calendar view - Use emoji only")

	isWorkCheckbox := common.NewFormCheckbox("Is Work", true).
		WithHelpText("included in mail report")

	form := common.NewMixedForm(&nameField, &shortNameField, isWorkCheckbox)

	return &WorkhourDetailsCreateModal{
		Form: form,
	}
}

func (m *WorkhourDetailsCreateModal) Update(msg tea.Msg) (WorkhourDetailsCreateModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if err := m.Form.Validate(); err != nil {
				return *m, nil
			}

			nameField := m.Form.GetField(0)
			shortNameField := m.Form.GetField(1)
			isWorkCheckbox := m.Form.GetCheckbox(2)

			name := strings.TrimSpace(nameField.Value())
			shortName := strings.TrimSpace(shortNameField.Value())
			isWork := isWorkCheckbox.Value

			return *m, tea.Batch(
				dispatchWorkhourDetailsCreatedMsg(name, shortName, isWork),
			)

		case "esc":
			return *m, tea.Batch(
				dispatchWorkhourDetailsCreateCanceledMsg(),
			)
		}
	}

	cmd := m.Form.Update(msg)
	return *m, cmd
}

func (m *WorkhourDetailsCreateModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	sb.WriteString(titleStyle.Render("Create New Workhour Detail"))
	sb.WriteString("\n\n")

	sb.WriteString(m.Form.View())

	sb.WriteString(render.RenderHelpText("Tab/Shift+Tab: navigate", "Space: toggle", "Enter: create", "ESC: cancel"))

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
