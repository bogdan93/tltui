package calendar

import (
	"fmt"
	"strings"
	"time"
	"tltui/src/domain"
	"tltui/src/domain/repository"
	"tltui/src/render"

	"github.com/charmbracelet/lipgloss"
)

// getWorkhoursForDate retrieves workhours for a specific date
func (m CalendarModel) getWorkhoursForDate(date time.Time) []domain.Workhour {
	workhours, err := repository.GetWorkhoursByDate(date)
	if err != nil {
		return []domain.Workhour{}
	}
	return workhours
}

// getWorkhourDetailsByID retrieves workhour details by ID
func (m CalendarModel) getWorkhourDetailsByID(id int) *domain.WorkhourDetails {
	details, err := repository.GetWorkhourDetailsByID(id)
	if err != nil {
		return nil
	}
	return details
}

// getProjectByID retrieves a project by ID
func (m CalendarModel) getProjectByID(id int) *domain.Project {
	project, err := repository.GetProjectByID(id)
	if err != nil {
		return nil
	}
	return project
}

// renderHelpModal renders the keyboard shortcuts help modal
func (m CalendarModel) renderHelpModal() string {
	var sb strings.Builder

	title := "Calendar Keyboard Shortcuts"
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center)

	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n\n")

	helpItems := [][]string{
		{"←/h, →/l", "Move selection left/right by one day"},
		{"↑/k, ↓/j", "Move selection up/down by one week"},
		{"<", "Previous month"},
		{">", "Next month"},
		{"r", "Reset to current month"},
		{"y", "Yank workhours from selected day"},
		{"p", "Paste yanked workhours to selected day"},
		{"d, x", "Delete all workhours from selected day"},
		{"g", "Generate report for current month"},
		{"enter", "View/edit workhours for selected day"},
		{"?", "Toggle this help"},
		{"q/esc", "Quit"},
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Width(15)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	for _, item := range helpItems {
		sb.WriteString(keyStyle.Render(item[0]))
		sb.WriteString("  ")
		sb.WriteString(descStyle.Render(item[1]))
		sb.WriteString("\n")
	}

	if len(m.YankedWorkhours) > 0 {
		sb.WriteString("\n")
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")).
			Italic(true)
		sb.WriteString(infoStyle.Render(fmt.Sprintf("📋 %d workhour(s) copied", len(m.YankedWorkhours))))
	}

	sb.WriteString("\n\n")
	sb.WriteString(render.RenderHelpText("?: close help"))

	return render.RenderSimpleModal(m.Width, m.Height, sb.String())
}
