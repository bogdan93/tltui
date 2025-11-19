package calendar

import (
	"fmt"
	"strings"
	"time"
	"tltui/src/domain"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkhourDeleteModal struct {
	Date            time.Time
	Workhour        domain.Workhour
	WorkhourDetails []domain.WorkhourDetails
	Projects        []domain.Project
}

type WorkhourDeleteConfirmedMsg struct {
	ID   int
	Date time.Time
}

type WorkhourDeleteCanceledMsg struct{}

func NewWorkhourDeleteModal(date time.Time, workhour domain.Workhour, workhourDetails []domain.WorkhourDetails, projects []domain.Project) *WorkhourDeleteModal {
	return &WorkhourDeleteModal{
		Date:            date,
		Workhour:        workhour,
		WorkhourDetails: workhourDetails,
		Projects:        projects,
	}
}

func (m *WorkhourDeleteModal) Update(msg tea.Msg) (WorkhourDeleteModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			return *m, tea.Batch(
				dispatchWorkhourDeleteConfirmedMsg(m.Workhour.ID, m.Date),
			)

		case "n", "N", "esc":
			return *m, tea.Batch(
				dispatchWorkhourDeleteCanceledMsg(),
			)
		}
	}

	return *m, nil
}

func (m *WorkhourDeleteModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")). // Red for delete
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")). // Orange warning
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	var detailsName string
	var detailsShortName string
	for _, wd := range m.WorkhourDetails {
		if wd.ID == m.Workhour.DetailsID {
			detailsName = wd.Name
			detailsShortName = wd.ShortName
			break
		}
	}

	var projectName string
	for _, p := range m.Projects {
		if p.ID == m.Workhour.ProjectID {
			projectName = p.Name
			break
		}
	}

	hoursStr := fmt.Sprintf("%.1f", m.Workhour.Hours)
	if m.Workhour.Hours == float64(int(m.Workhour.Hours)) {
		hoursStr = fmt.Sprintf("%d", int(m.Workhour.Hours))
	}

	sb.WriteString(titleStyle.Render("⚠ Delete Work Hours"))
	sb.WriteString("\n\n")

	sb.WriteString(warningStyle.Render("Are you sure you want to delete this entry?"))
	sb.WriteString("\n\n")

	dateStr := m.Date.Format("Monday, January 2, 2006")
	sb.WriteString(labelStyle.Render("Date: "))
	sb.WriteString(valueStyle.Render(dateStr))
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("Type: "))
	sb.WriteString(valueStyle.Render(fmt.Sprintf("%s %s", detailsShortName, detailsName)))
	sb.WriteString("\n\n")

	if projectName != "" {
		sb.WriteString(labelStyle.Render("Project: "))
		sb.WriteString(valueStyle.Render(projectName))
		sb.WriteString("\n\n")
	}

	sb.WriteString(labelStyle.Render("Hours: "))
	sb.WriteString(valueStyle.Render(fmt.Sprintf("%sh", hoursStr)))
	sb.WriteString("\n\n")

	warningStyle2 := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
	sb.WriteString(warningStyle2.Render("This action cannot be undone!"))
	sb.WriteString("\n\n")

	sb.WriteString(helpStyle.Render("Y/Enter: confirm delete • N/ESC: cancel"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func dispatchWorkhourDeleteConfirmedMsg(id int, date time.Time) tea.Cmd {
	return func() tea.Msg {
		return WorkhourDeleteConfirmedMsg{
			ID:   id,
			Date: date,
		}
	}
}

func dispatchWorkhourDeleteCanceledMsg() tea.Cmd {
	return func() tea.Msg {
		return WorkhourDeleteCanceledMsg{}
	}
}
