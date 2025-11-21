package calendar

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"tltui/src/common"
	"tltui/src/domain"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkhourEditModal struct {
	WorkhourID int
	Date       time.Time
	Form       *common.MixedForm
}

type WorkhourEditSubmittedMsg struct {
	WorkhourID int
	Date       time.Time
	DetailsID  int
	ProjectID  int
	Hours      float64
}

type WorkhourEditCanceledMsg struct{}

func NewWorkhourEditModal(
	workhourID int,
	date time.Time,
	currentDetailsID int,
	currentProjectID int,
	currentHours float64,
	workhourDetails []domain.WorkhourDetails,
	projects []domain.Project,
) *WorkhourEditModal {
	detailsOptions := make([]common.SelectOption, len(workhourDetails))
	selectedDetailsIndex := 0
	for i, d := range workhourDetails {
		workType := "work"
		if !d.IsWork {
			workType = "non-work"
		}
		detailsOptions[i] = common.SelectOption{
			ID:          d.ID,
			DisplayName: fmt.Sprintf("%s %s", d.ShortName, d.Name),
			ExtraInfo:   workType,
		}
		if d.ID == currentDetailsID {
			selectedDetailsIndex = i
		}
	}

	projectOptions := make([]common.SelectOption, len(projects))
	selectedProjectIndex := 0
	for i, p := range projects {
		projectOptions[i] = common.SelectOption{
			ID:          p.ID,
			DisplayName: p.Name,
			ExtraInfo:   fmt.Sprintf("Odoo: %d", p.OdooID),
		}
		if p.ID == currentProjectID {
			selectedProjectIndex = i
		}
	}

	detailsSelect := common.NewRequiredFormSelect("Type", detailsOptions)
	detailsSelect.SelectedIndex = selectedDetailsIndex

	projectSelect := common.NewRequiredFormSelect("Project", projectOptions)
	projectSelect.SelectedIndex = selectedProjectIndex

	hoursField := common.NewRequiredFormField("Hours", fmt.Sprintf("%.1f", currentHours), 20).
		WithCharLimit(5).
		WithValidator(common.PositiveFloatValidator("Hours"))
  hoursField.Input.SetValue(fmt.Sprintf("%.1f", currentHours))

	// Create form
	form := common.NewMixedForm(detailsSelect, projectSelect, &hoursField)

	return &WorkhourEditModal{
		WorkhourID: workhourID,
		Date:       date,
		Form:       form,
	}
}

func (m *WorkhourEditModal) Update(msg tea.Msg) (WorkhourEditModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if err := m.Form.Validate(); err != nil {
				return *m, nil
			}

			detailsID := m.Form.GetSelect(0).GetSelectedID()
			projectID := m.Form.GetSelect(1).GetSelectedID()
			hoursStr := strings.TrimSpace(m.Form.GetField(2).Value())
			hours, _ := strconv.ParseFloat(hoursStr, 64) // Already validated

			return *m, tea.Batch(
				dispatchWorkhourEditSubmittedMsg(m.WorkhourID, m.Date, detailsID, projectID, hours),
			)

		case "esc":
			return *m, tea.Batch(
				dispatchWorkhourEditCanceledMsg(),
			)
		}
	}

	cmd := m.Form.Update(msg)
	return *m, cmd
}

func (m *WorkhourEditModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")).
		MarginBottom(1)

	dateStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginBottom(1)

	sb.WriteString(titleStyle.Render("Edit Work Hours"))
	sb.WriteString("\n")

	dateStr := m.Date.Format("Monday, January 2, 2006")
	sb.WriteString(dateStyle.Render(dateStr))
	sb.WriteString("\n\n")

	sb.WriteString(m.Form.View())

	sb.WriteString(render.RenderHelpText("Tab: next", "↑/↓: select", "Enter: save", "ESC: cancel"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func dispatchWorkhourEditSubmittedMsg(workhourID int, date time.Time, detailsID int, projectID int, hours float64) tea.Cmd {
	return func() tea.Msg {
		return WorkhourEditSubmittedMsg{
			WorkhourID: workhourID,
			Date:       date,
			DetailsID:  detailsID,
			ProjectID:  projectID,
			Hours:      hours,
		}
	}
}

func dispatchWorkhourEditCanceledMsg() tea.Cmd {
	return func() tea.Msg {
		return WorkhourEditCanceledMsg{}
	}
}
