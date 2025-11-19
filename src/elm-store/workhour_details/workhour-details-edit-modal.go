package workhour_details

import (
	"strings"
	"tltui/src/common"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkhourDetailsEditModal struct {
	EditingWorkhourDetailID int
	Form                    *common.MixedForm
}

type WorkhourDetailsEditedMsg struct {
	WorkhourDetailID int
	Name             string
	ShortName        string
	IsWork           bool
}

type WorkhourDetailsEditCanceledMsg struct{}

func NewWorkhourDetailsEditModal(workhourDetailID int, name string, shortName string, isWork bool) *WorkhourDetailsEditModal {
	nameField := common.NewRequiredFormField("Name", "Name", 40).
		WithInitialValue(name)

	shortNameField := common.NewRequiredFormField("Short Name", "Short Name", 40).
		WithCharLimit(20).
		WithHelpText("Displayed in calendar view - Use emoji only").
		WithInitialValue(shortName)

	isWorkCheckbox := common.NewFormCheckbox("Is Work", isWork).
		WithHelpText("included in mail report")

	form := common.NewMixedForm(&nameField, &shortNameField, isWorkCheckbox)

	return &WorkhourDetailsEditModal{
		EditingWorkhourDetailID: workhourDetailID,
		Form:                    form,
	}
}

func (m *WorkhourDetailsEditModal) Update(msg tea.Msg) (WorkhourDetailsEditModal, tea.Cmd) {
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
				dispatchWorkhourDetailsEditedMsg(m.EditingWorkhourDetailID, name, shortName, isWork),
			)

		case "esc":
			return *m, tea.Batch(
				dispatchWorkhourDetailsEditCanceledMsg(),
			)
		}
	}

	cmd := m.Form.Update(msg)
	return *m, cmd
}

func (m *WorkhourDetailsEditModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	sb.WriteString(titleStyle.Render("Edit Workhour Detail"))
	sb.WriteString("\n\n")

	sb.WriteString(m.Form.View())

	sb.WriteString(render.RenderHelpText("Tab/Shift+Tab: navigate", "Space: toggle", "Enter: save", "ESC: cancel"))

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
