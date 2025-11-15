package models

import (
	"fmt"
	"strings"
	"time"
	"time-logger-tui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkhoursViewModal struct {
	Date            time.Time
	Workhours       []Workhour
	WorkhourDetails []WorkhourDetails
	Projects        []Project
}

type WorkhoursViewModalClosedMsg struct{}

func NewWorkhoursViewModal(date time.Time, workhours []Workhour, workhourDetails []WorkhourDetails, projects []Project) *WorkhoursViewModal {
	return &WorkhoursViewModal{
		Date:            date,
		Workhours:       workhours,
		WorkhourDetails: workhourDetails,
		Projects:        projects,
	}
}

func (m *WorkhoursViewModal) Update(msg tea.Msg) (WorkhoursViewModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "enter":
			// Close modal
			return *m, tea.Batch(
				dispatchWorkhoursViewModalClosedMsg(),
			)
		}
	}

	return *m, nil
}

func (m *WorkhoursViewModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("141")). // Purple
		MarginBottom(1)

	dateStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")). // Cyan
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	totalStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("114")). // Green
		MarginTop(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		MarginTop(1)

	// Title
	sb.WriteString(titleStyle.Render("Work Hours"))
	sb.WriteString("\n")

	// Date
	dateStr := m.Date.Format("Monday, January 2, 2006")
	sb.WriteString(dateStyle.Render(dateStr))
	sb.WriteString("\n\n")

	// Check if there are any workhours
	if len(m.Workhours) == 0 {
		sb.WriteString(emptyStyle.Render("No work hours logged for this day."))
		sb.WriteString("\n\n")
		sb.WriteString(helpStyle.Render("ESC/Enter: close"))
		return render.RenderSimpleModal(Width, Height, sb.String())
	}

	// Display each workhour entry
	var totalWorkHours float64
	var totalNonWorkHours float64

	for i, wh := range m.Workhours {
		// Find the workhour details
		var details *WorkhourDetails
		for _, wd := range m.WorkhourDetails {
			if wd.ID == wh.DetailsID {
				details = &wd
				break
			}
		}

		// Find the project
		var project *Project
		if wh.ProjectID > 0 {
			for _, p := range m.Projects {
				if p.ID == wh.ProjectID {
					project = &p
					break
				}
			}
		}

		// Format hours
		hoursStr := fmt.Sprintf("%.1f", wh.Hours)
		if wh.Hours == float64(int(wh.Hours)) {
			hoursStr = fmt.Sprintf("%d", int(wh.Hours))
		}

		// Display entry
		if details != nil {
			sb.WriteString(labelStyle.Render(fmt.Sprintf("%s %s: ", details.ShortName, details.Name)))
			sb.WriteString(valueStyle.Render(fmt.Sprintf("%sh", hoursStr)))

			if project != nil {
				sb.WriteString(valueStyle.Render(fmt.Sprintf(" (%s)", project.Name)))
			}

			// Track totals
			if details.IsWork {
				totalWorkHours += wh.Hours
			} else {
				totalNonWorkHours += wh.Hours
			}
		} else {
			sb.WriteString(labelStyle.Render(fmt.Sprintf("Unknown: ")))
			sb.WriteString(valueStyle.Render(fmt.Sprintf("%sh", hoursStr)))
		}

		if i < len(m.Workhours)-1 {
			sb.WriteString("\n")
		}
	}

	// Display totals
	sb.WriteString("\n\n")
	sb.WriteString(labelStyle.Render("───────────────────────────"))
	sb.WriteString("\n")

	totalHours := totalWorkHours + totalNonWorkHours

	if totalWorkHours > 0 {
		workHoursStr := fmt.Sprintf("%.1f", totalWorkHours)
		if totalWorkHours == float64(int(totalWorkHours)) {
			workHoursStr = fmt.Sprintf("%d", int(totalWorkHours))
		}
		sb.WriteString(labelStyle.Render("Work hours: "))
		sb.WriteString(totalStyle.Render(fmt.Sprintf("%sh", workHoursStr)))
		sb.WriteString("\n")
	}

	if totalNonWorkHours > 0 {
		nonWorkHoursStr := fmt.Sprintf("%.1f", totalNonWorkHours)
		if totalNonWorkHours == float64(int(totalNonWorkHours)) {
			nonWorkHoursStr = fmt.Sprintf("%d", int(totalNonWorkHours))
		}
		sb.WriteString(labelStyle.Render("Non-work hours: "))
		sb.WriteString(valueStyle.Render(fmt.Sprintf("%sh", nonWorkHoursStr)))
		sb.WriteString("\n")
	}

	totalHoursStr := fmt.Sprintf("%.1f", totalHours)
	if totalHours == float64(int(totalHours)) {
		totalHoursStr = fmt.Sprintf("%d", int(totalHours))
	}
	sb.WriteString(labelStyle.Render("Total: "))
	sb.WriteString(totalStyle.Render(fmt.Sprintf("%sh", totalHoursStr)))

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("ESC/Enter: close"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func dispatchWorkhoursViewModalClosedMsg() tea.Cmd {
	return func() tea.Msg {
		return WorkhoursViewModalClosedMsg{}
	}
}
