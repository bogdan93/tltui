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

type WorkhourCreateModal struct {
	Date time.Time
	Form *common.MixedForm
}

type WorkhourCreateSubmittedMsg struct {
	Date      time.Time
	DetailsID int
	ProjectID int
	Hours     float64
}

type WorkhourCreateCanceledMsg struct{}

func NewWorkhourCreateModal(date time.Time, workhourDetails []domain.WorkhourDetails, projects []domain.Project) *WorkhourCreateModal {
	// Build activity/details options
	detailsOptions := make([]common.SelectOption, len(workhourDetails))
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
	}

	// Build project options
	projectOptions := make([]common.SelectOption, len(projects))
	for i, p := range projects {
		projectOptions[i] = common.SelectOption{
			ID:          p.ID,
			DisplayName: p.Name,
			ExtraInfo:   fmt.Sprintf("Odoo: %d", p.OdooID),
		}
	}

	// Create form elements
	detailsSelect := common.NewRequiredFormSelect("Type", detailsOptions)
	projectSelect := common.NewRequiredFormSelect("Project", projectOptions)
	hoursField := common.NewRequiredFormField("Hours", "8.0", 20).
		WithCharLimit(5).
		WithValidator(common.PositiveFloatValidator("Hours"))

	// Create form
	form := common.NewMixedForm(detailsSelect, projectSelect, &hoursField)

	return &WorkhourCreateModal{
		Date: date,
		Form: form,
	}
}

func (m *WorkhourCreateModal) Update(msg tea.Msg) (WorkhourCreateModal, tea.Cmd) {
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
				dispatchWorkhourCreateSubmittedMsg(m.Date, detailsID, projectID, hours),
			)

		case "esc":
			return *m, tea.Batch(
				dispatchWorkhourCreateCanceledMsg(),
			)
		}
	}

	cmd := m.Form.Update(msg)
	return *m, cmd
}

func (m *WorkhourCreateModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")).
		MarginBottom(1)

	dateStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginBottom(1)

	sb.WriteString(titleStyle.Render("Add Work Hours"))
	sb.WriteString("\n")

	dateStr := m.Date.Format("Monday, January 2, 2006")
	sb.WriteString(dateStyle.Render(dateStr))
	sb.WriteString("\n\n")

	sb.WriteString(m.Form.View())

	sb.WriteString(render.RenderHelpText("Tab: next", "↑/↓: select", "Enter: save", "ESC: cancel"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func dispatchWorkhourCreateSubmittedMsg(date time.Time, detailsID int, projectID int, hours float64) tea.Cmd {
	return func() tea.Msg {
		return WorkhourCreateSubmittedMsg{
			Date:      date,
			DetailsID: detailsID,
			ProjectID: projectID,
			Hours:     hours,
		}
	}
}

func dispatchWorkhourCreateCanceledMsg() tea.Cmd {
	return func() tea.Msg {
		return WorkhourCreateCanceledMsg{}
	}
}
