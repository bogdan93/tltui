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

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")).
			MarginBottom(1)

	dateStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("241"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	totalStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("114")).
			MarginTop(1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))
)

type WorkhoursViewModal struct {
	Date            time.Time
	Workhours       []domain.Workhour
	WorkhourDetails []domain.WorkhourDetails
	Projects        []domain.Project

	SelectedWorkhourIndex int // Index in Workhours array for selection
}

type WorkhoursViewModalClosedMsg struct{}

type WorkhoursViewModalCreateRequestedMsg struct {
	Date time.Time
}

type WorkhoursViewModalEditRequestedMsg struct {
	WorkhourID int
	Date       time.Time
}

type WorkhoursViewModalDeleteRequestedMsg struct {
	WorkhourID int
	Date       time.Time
}

func NewWorkhoursViewModal(date time.Time, workhours []domain.Workhour, workhourDetails []domain.WorkhourDetails, projects []domain.Project) *WorkhoursViewModal {
	return &WorkhoursViewModal{
		Date:                  date,
		Workhours:             workhours,
		WorkhourDetails:       workhourDetails,
		Projects:              projects,
		SelectedWorkhourIndex: 0,
	}
}

func (m *WorkhoursViewModal) Update(msg tea.Msg) (WorkhoursViewModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return *m, tea.Batch(
				dispatchWorkhoursViewModalClosedMsg(),
			)
		case "n":
			return *m, tea.Batch(
				dispatchWorkhoursViewModalCreateRequestedMsg(m.Date),
			)
		case "e", "enter":
			if len(m.Workhours) > 0 && m.SelectedWorkhourIndex >= 0 && m.SelectedWorkhourIndex < len(m.Workhours) {
				workhourID := m.Workhours[m.SelectedWorkhourIndex].ID
				return *m, tea.Batch(
					dispatchWorkhoursViewModalEditRequestedMsg(workhourID, m.Date),
				)
			}
			return *m, tea.Batch(
				dispatchWorkhoursViewModalClosedMsg(),
			)
		case "d":
			if len(m.Workhours) > 0 && m.SelectedWorkhourIndex >= 0 && m.SelectedWorkhourIndex < len(m.Workhours) {
				workhourID := m.Workhours[m.SelectedWorkhourIndex].ID
				return *m, tea.Batch(
					dispatchWorkhoursViewModalDeleteRequestedMsg(workhourID, m.Date),
				)
			}
		case "up", "k":
			if len(m.Workhours) > 0 {
				m.SelectedWorkhourIndex = (m.SelectedWorkhourIndex - 1 + len(m.Workhours)) % len(m.Workhours)
			}
			return *m, nil
		case "down", "j":
			if len(m.Workhours) > 0 {
				m.SelectedWorkhourIndex = (m.SelectedWorkhourIndex + 1) % len(m.Workhours)
			}
			return *m, nil
		}
	}

	return *m, nil
}

func (m *WorkhoursViewModal) View(Width, Height int) string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("Work Hours"))
	sb.WriteString("\n")

	dateStr := m.Date.Format("Monday, January 2, 2006")
	sb.WriteString(dateStyle.Render(dateStr))
	sb.WriteString("\n\n")

	if len(m.Workhours) == 0 {
		sb.WriteString(emptyStyle.Render("No work hours logged for this day."))
		sb.WriteString("\n\n")
		sb.WriteString(render.RenderHelpText("n: new", "ESC/Enter: close"))
		return render.RenderSimpleModal(Width, Height, sb.String())
	}

	var totalWorkHours float64
	var totalNonWorkHours float64

	for i, wh := range m.Workhours {
		prefix := "  "
		if i == m.SelectedWorkhourIndex {
			prefix = "▶ "
		}
		sb.WriteString(prefix)

		var details *domain.WorkhourDetails
		for _, wd := range m.WorkhourDetails {
			if wd.ID == wh.DetailsID {
				details = &wd
				break
			}
		}

		var project *domain.Project
		if wh.ProjectID > 0 {
			for _, p := range m.Projects {
				if p.ID == wh.ProjectID {
					project = &p
					break
				}
			}
		}

		hoursStr := fmt.Sprintf("%.1f", wh.Hours)
		if wh.Hours == float64(int(wh.Hours)) {
			hoursStr = fmt.Sprintf("%d", int(wh.Hours))
		}

		entryLabelStyle := labelStyle
		entryValueStyle := valueStyle
		if i == m.SelectedWorkhourIndex {
			entryLabelStyle = selectedStyle
			entryValueStyle = selectedStyle
		}

		if details != nil {
			sb.WriteString(entryLabelStyle.Render(fmt.Sprintf("%s %s: ", details.ShortName, details.Name)))
			sb.WriteString(entryValueStyle.Render(fmt.Sprintf("%sh", hoursStr)))

			if project != nil {
				sb.WriteString(entryValueStyle.Render(fmt.Sprintf(" (%s)", project.Name)))
			}

			if details.IsWork {
				totalWorkHours += wh.Hours
			} else {
				totalNonWorkHours += wh.Hours
			}
		} else {
			sb.WriteString(entryLabelStyle.Render("Unknown: "))
			sb.WriteString(entryValueStyle.Render(fmt.Sprintf("%sh", hoursStr)))
		}

		if i < len(m.Workhours)-1 {
			sb.WriteString("\n")
		}
	}

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

	sb.WriteString(render.RenderHelpText(
		"n: new",
		"e/Enter: edit",
		"d: delete",
		"↑/↓: select",
		"ESC: close"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func dispatchWorkhoursViewModalClosedMsg() tea.Cmd {
	return func() tea.Msg {
		return WorkhoursViewModalClosedMsg{}
	}
}

func dispatchWorkhoursViewModalCreateRequestedMsg(date time.Time) tea.Cmd {
	return func() tea.Msg {
		return WorkhoursViewModalCreateRequestedMsg{Date: date}
	}
}

func dispatchWorkhoursViewModalEditRequestedMsg(workhourID int, date time.Time) tea.Cmd {
	return func() tea.Msg {
		return WorkhoursViewModalEditRequestedMsg{
			WorkhourID: workhourID,
			Date:       date,
		}
	}
}

func dispatchWorkhoursViewModalDeleteRequestedMsg(workhourID int, date time.Time) tea.Cmd {
	return func() tea.Msg {
		return WorkhoursViewModalDeleteRequestedMsg{
			WorkhourID: workhourID,
			Date:       date,
		}
	}
}
