package models

import (
	"fmt"
	"strings"
	"time"
	"time-logger-tui/src/render"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ModalMode int

const (
	ModeView ModalMode = iota
	ModeCreate
	ModeEdit
	ModeDelete
)

type WorkhoursViewModal struct {
	Date            time.Time
	Workhours       []Workhour
	WorkhourDetails []WorkhourDetails
	Projects        []Project

	// Mode management
	Mode ModalMode

	// Create/Edit mode fields
	HoursInput             textinput.Model
	SelectedDetailsIndex   int // Index in WorkhourDetails array
	SelectedProjectIndex   int // Index in Projects array (0 = None)
	FocusedInput           int // 0=details, 1=project, 2=hours
	ErrorMessage           string
}

type WorkhoursViewModalClosedMsg struct{}

func NewWorkhoursViewModal(date time.Time, workhours []Workhour, workhourDetails []WorkhourDetails, projects []Project) *WorkhoursViewModal {
	hoursInput := textinput.New()
	hoursInput.Placeholder = "8.0"
	hoursInput.CharLimit = 5
	hoursInput.Width = 20

	return &WorkhoursViewModal{
		Date:            date,
		Workhours:       workhours,
		WorkhourDetails: workhourDetails,
		Projects:        projects,
		Mode:            ModeView,
		HoursInput:      hoursInput,
		SelectedDetailsIndex: 0,
		SelectedProjectIndex: -1, // -1 means no project selected (optional)
		FocusedInput:    0,
	}
}

func (m *WorkhoursViewModal) Update(msg tea.Msg) (WorkhoursViewModal, tea.Cmd) {
	// Route to mode-specific update handlers
	switch m.Mode {
	case ModeView:
		return m.updateViewMode(msg)
	case ModeCreate:
		return m.updateCreateMode(msg)
	case ModeEdit:
		return m.updateEditMode(msg)
	case ModeDelete:
		return m.updateDeleteMode(msg)
	}

	return *m, nil
}

func (m *WorkhoursViewModal) updateViewMode(msg tea.Msg) (WorkhoursViewModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "enter":
			// Close modal
			return *m, tea.Batch(
				dispatchWorkhoursViewModalClosedMsg(),
			)
		case "n":
			// Switch to create mode
			m.initializeCreateMode()
			return *m, nil
		}
	}

	return *m, nil
}

func (m *WorkhoursViewModal) updateCreateMode(msg tea.Msg) (WorkhoursViewModal, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Validate and submit
			return m.submitCreate()

		case "esc":
			// Cancel and return to view mode
			m.Mode = ModeView
			m.ErrorMessage = ""
			return *m, nil

		case "tab":
			m.FocusedInput = (m.FocusedInput + 1) % 3
			m.updateInputFocus()
			return *m, nil

		case "shift+tab":
			m.FocusedInput = (m.FocusedInput - 1 + 3) % 3
			m.updateInputFocus()
			return *m, nil

		case "up":
			if m.FocusedInput == 0 && len(m.WorkhourDetails) > 0 {
				m.SelectedDetailsIndex = (m.SelectedDetailsIndex - 1 + len(m.WorkhourDetails)) % len(m.WorkhourDetails)
			} else if m.FocusedInput == 1 && len(m.Projects) > 0 {
				if m.SelectedProjectIndex < 0 {
					m.SelectedProjectIndex = len(m.Projects) - 1
				} else {
					m.SelectedProjectIndex = (m.SelectedProjectIndex - 1 + len(m.Projects)) % len(m.Projects)
				}
			}
			return *m, nil

		case "down":
			if m.FocusedInput == 0 && len(m.WorkhourDetails) > 0 {
				m.SelectedDetailsIndex = (m.SelectedDetailsIndex + 1) % len(m.WorkhourDetails)
			} else if m.FocusedInput == 1 && len(m.Projects) > 0 {
				if m.SelectedProjectIndex < 0 {
					m.SelectedProjectIndex = 0
				} else {
					m.SelectedProjectIndex = (m.SelectedProjectIndex + 1) % len(m.Projects)
				}
			}
			return *m, nil
		}
	}

	// Update hours input if focused
	if m.FocusedInput == 2 {
		m.HoursInput, cmd = m.HoursInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return *m, tea.Batch(cmds...)
}

func (m *WorkhoursViewModal) updateEditMode(msg tea.Msg) (WorkhoursViewModal, tea.Cmd) {
	// TODO: Implement edit mode
	return *m, nil
}

func (m *WorkhoursViewModal) updateDeleteMode(msg tea.Msg) (WorkhoursViewModal, tea.Cmd) {
	// TODO: Implement delete mode
	return *m, nil
}

func (m *WorkhoursViewModal) submitCreate() (WorkhoursViewModal, tea.Cmd) {
	// Validate hours
	hoursStr := strings.TrimSpace(m.HoursInput.Value())
	if hoursStr == "" {
		m.ErrorMessage = "Hours is required"
		return *m, nil
	}

	var hours float64
	_, err := fmt.Sscanf(hoursStr, "%f", &hours)
	if err != nil || hours <= 0 {
		m.ErrorMessage = "Hours must be a positive number"
		return *m, nil
	}

	// Get selected details ID
	if len(m.WorkhourDetails) == 0 || m.SelectedDetailsIndex >= len(m.WorkhourDetails) {
		m.ErrorMessage = "Please select a workhour type"
		return *m, nil
	}
	detailsID := m.WorkhourDetails[m.SelectedDetailsIndex].ID

	// Validate project selection (required)
	if m.SelectedProjectIndex < 0 || m.SelectedProjectIndex >= len(m.Projects) {
		m.ErrorMessage = "Please select a project"
		return *m, nil
	}
	projectID := m.Projects[m.SelectedProjectIndex].ID

	// Clear error and dispatch creation message
	m.ErrorMessage = ""
	m.Mode = ModeView // Return to view mode after creation
	return *m, tea.Batch(
		dispatchWorkhourCreatedMsg(m.Date, detailsID, projectID, hours),
	)
}

func (m *WorkhoursViewModal) View(Width, Height int) string {
	// Route to mode-specific view renderers
	switch m.Mode {
	case ModeView:
		return m.viewModeView(Width, Height)
	case ModeCreate:
		return m.viewCreateMode(Width, Height)
	case ModeEdit:
		return m.viewEditMode(Width, Height)
	case ModeDelete:
		return m.viewDeleteMode(Width, Height)
	}
	return ""
}

func (m *WorkhoursViewModal) viewModeView(Width, Height int) string {
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
		sb.WriteString(helpStyle.Render("n: new • ESC/Enter: close"))
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
	sb.WriteString(helpStyle.Render("n: new • ESC/Enter: close"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func (m *WorkhoursViewModal) viewCreateMode(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")). // Blue
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

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	focusedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))

	// Title
	sb.WriteString(titleStyle.Render("Add Work Hours"))
	sb.WriteString("\n")

	// Date (read-only)
	dateStr := m.Date.Format("Monday, January 2, 2006")
	sb.WriteString(dateStyle.Render(dateStr))
	sb.WriteString("\n\n")

	// WorkhourDetails selection
	sb.WriteString(labelStyle.Render("Type:"))
	sb.WriteString("\n")
	for i, details := range m.WorkhourDetails {
		style := valueStyle
		prefix := "  "
		if i == m.SelectedDetailsIndex {
			if m.FocusedInput == 0 {
				style = focusedStyle
				prefix = "▶ "
			} else {
				prefix = "• "
			}
		}
		workType := "work"
		if !details.IsWork {
			workType = "non-work"
		}
		sb.WriteString(prefix + style.Render(fmt.Sprintf("%s %s (%s)", details.ShortName, details.Name, workType)))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Project selection
	sb.WriteString(labelStyle.Render("Project:"))
	sb.WriteString("\n")

	if len(m.Projects) == 0 {
		// No projects available - show hint
		hintStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		sb.WriteString(hintStyle.Render("  No projects available • Create one in tab 2"))
		sb.WriteString("\n")
	} else {
		// Project list
		for i, project := range m.Projects {
			style := valueStyle
			prefix := "  "
			if i == m.SelectedProjectIndex {
				if m.FocusedInput == 1 {
					style = focusedStyle
					prefix = "▶ "
				} else {
					prefix = "• "
				}
			}
			sb.WriteString(prefix + style.Render(fmt.Sprintf("%s (Odoo: %d)", project.Name, project.OdooID)))
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	// Hours input
	sb.WriteString(labelStyle.Render("Hours:"))
	sb.WriteString("\n")
	if m.FocusedInput == 2 {
		sb.WriteString("▶ ")
	} else {
		sb.WriteString("  ")
	}
	sb.WriteString(m.HoursInput.View())
	sb.WriteString("\n\n")

	// Error message (always reserve space to prevent shifting)
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
	if m.ErrorMessage != "" {
		sb.WriteString(errorStyle.Render("⚠ " + m.ErrorMessage))
	} else {
		// Empty space to maintain consistent height
		sb.WriteString(" ")
	}
	sb.WriteString("\n\n")

	sb.WriteString(helpStyle.Render("Tab: next • ↑/↓: select • Enter: save • ESC: cancel"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func (m *WorkhoursViewModal) viewEditMode(Width, Height int) string {
	// TODO: Implement edit mode view
	return ""
}

func (m *WorkhoursViewModal) viewDeleteMode(Width, Height int) string {
	// TODO: Implement delete mode view
	return ""
}

func dispatchWorkhoursViewModalClosedMsg() tea.Cmd {
	return func() tea.Msg {
		return WorkhoursViewModalClosedMsg{}
	}
}

type WorkhourCreatedMsg struct {
	Date      time.Time
	DetailsID int
	ProjectID int
	Hours     float64
}

func dispatchWorkhourCreatedMsg(date time.Time, detailsID int, projectID int, hours float64) tea.Cmd {
	return func() tea.Msg {
		return WorkhourCreatedMsg{
			Date:      date,
			DetailsID: detailsID,
			ProjectID: projectID,
			Hours:     hours,
		}
	}
}

// Helper method to initialize/reset create mode
func (m *WorkhoursViewModal) initializeCreateMode() {
	m.Mode = ModeCreate
	m.HoursInput.SetValue("")
	m.HoursInput.Focus()
	m.SelectedDetailsIndex = 0
	// Preselect first project if available
	if len(m.Projects) > 0 {
		m.SelectedProjectIndex = 0
	} else {
		m.SelectedProjectIndex = -1
	}
	m.FocusedInput = 0
	m.ErrorMessage = ""
	m.updateInputFocus()
}

// Helper method to update input focus
func (m *WorkhoursViewModal) updateInputFocus() {
	if m.FocusedInput == 2 {
		m.HoursInput.Focus()
	} else {
		m.HoursInput.Blur()
	}
}
