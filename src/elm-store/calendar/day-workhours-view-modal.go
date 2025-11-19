package calendar

import (
	"fmt"
	"strings"
	"time"
	"tltui/src/domain"
	"tltui/src/domain/repository"
	"tltui/src/render"

	"github.com/charmbracelet/bubbles/textinput"
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

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	focusedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

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

type ModalMode int

const (
	ModeView ModalMode = iota
	ModeCreate
	ModeEdit
	ModeDelete
)

type WorkhoursViewModal struct {
	Date            time.Time
	Workhours       []domain.Workhour
	WorkhourDetails []domain.WorkhourDetails
	Projects        []domain.Project

	Mode ModalMode

	SelectedWorkhourIndex int // Index in Workhours array for edit/delete

	HoursInput           textinput.Model
	SelectedDetailsIndex int // Index in domain.WorkhourDetails array
	SelectedProjectIndex int // Index in Projects array (-1 = None)
	FocusedInput         int // 0=details, 1=project, 2=hours
	ErrorMessage         string
}

type WorkhoursViewModalClosedMsg struct{}

func NewWorkhoursViewModal(date time.Time, workhours []domain.Workhour, workhourDetails []domain.WorkhourDetails, projects []domain.Project) *WorkhoursViewModal {
	hoursInput := textinput.New()
	hoursInput.Placeholder = "8.0"
	hoursInput.CharLimit = 5
	hoursInput.Width = 20

	return &WorkhoursViewModal{
		Date:                  date,
		Workhours:             workhours,
		WorkhourDetails:       workhourDetails,
		Projects:              projects,
		Mode:                  ModeView,
		HoursInput:            hoursInput,
		SelectedWorkhourIndex: 0,
		SelectedDetailsIndex:  0,
		SelectedProjectIndex:  -1, // -1 means no project selected (optional)
		FocusedInput:          0,
	}
}

func (m *WorkhoursViewModal) Update(msg tea.Msg) (WorkhoursViewModal, tea.Cmd) {
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
		case "esc", "q":
			return *m, tea.Batch(
				dispatchWorkhoursViewModalClosedMsg(),
			)
		case "n":
			m.initializeCreateMode()
			return *m, nil
		case "e", "enter":
			if len(m.Workhours) > 0 {
				m.initializeEditMode()
				return *m, nil
			}
			return *m, tea.Batch(
				dispatchWorkhoursViewModalClosedMsg(),
			)
		case "d":
			if len(m.Workhours) > 0 {
				m.Mode = ModeDelete
				return *m, nil
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

func (m *WorkhoursViewModal) updateCreateMode(msg tea.Msg) (WorkhoursViewModal, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m.submitCreate()

		case "esc":
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

	if m.FocusedInput == 2 {
		m.HoursInput, cmd = m.HoursInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return *m, tea.Batch(cmds...)
}

func (m *WorkhoursViewModal) updateEditMode(msg tea.Msg) (WorkhoursViewModal, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m.submitEdit()

		case "esc":
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

	if m.FocusedInput == 2 {
		m.HoursInput, cmd = m.HoursInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return *m, tea.Batch(cmds...)
}

func (m *WorkhoursViewModal) updateDeleteMode(msg tea.Msg) (WorkhoursViewModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			if m.SelectedWorkhourIndex >= 0 && m.SelectedWorkhourIndex < len(m.Workhours) {
				workhourID := m.Workhours[m.SelectedWorkhourIndex].ID
				m.Mode = ModeView
				return *m, tea.Batch(
					dispatchWorkhourDeletedMsg(workhourID, m.Date),
				)
			}
			m.Mode = ModeView
			return *m, nil

		case "n", "N", "esc":
			m.Mode = ModeView
			return *m, nil
		}
	}

	return *m, nil
}

func (m *WorkhoursViewModal) submitEdit() (WorkhoursViewModal, tea.Cmd) {
	hoursStr := strings.TrimSpace(m.HoursInput.Value())
	if hoursStr == "" {
		m.ErrorMessage = "Hours is required"
		m.FocusedInput = 2
		m.updateInputFocus()
		return *m, nil
	}

	var hours float64
	_, err := fmt.Sscanf(hoursStr, "%f", &hours)
	if err != nil || hours <= 0 {
		m.ErrorMessage = "Hours must be a positive number"
		m.FocusedInput = 2
		m.updateInputFocus()
		return *m, nil
	}

	if len(m.WorkhourDetails) == 0 || m.SelectedDetailsIndex >= len(m.WorkhourDetails) {
		m.ErrorMessage = "Please select a workhour type"
		m.FocusedInput = 0
		return *m, nil
	}
	detailsID := m.WorkhourDetails[m.SelectedDetailsIndex].ID

	if m.SelectedProjectIndex < 0 || m.SelectedProjectIndex >= len(m.Projects) {
		m.ErrorMessage = "Please select a project"
		m.FocusedInput = 1
		return *m, nil
	}
	projectID := m.Projects[m.SelectedProjectIndex].ID

	if m.SelectedWorkhourIndex < 0 || m.SelectedWorkhourIndex >= len(m.Workhours) {
		m.ErrorMessage = "Invalid workhour selection"
		m.FocusedInput = 0
		return *m, nil
	}

	workhourID := m.Workhours[m.SelectedWorkhourIndex].ID

	m.Mode = ModeView
	return *m, tea.Batch(
		dispatchWorkhourEditedMsg(workhourID, m.Date, detailsID, projectID, hours),
	)
}

func (m *WorkhoursViewModal) submitCreate() (WorkhoursViewModal, tea.Cmd) {
	hoursStr := strings.TrimSpace(m.HoursInput.Value())
	if hoursStr == "" {
		m.ErrorMessage = "Hours is required"
		m.FocusedInput = 2
		m.updateInputFocus()
		return *m, nil
	}

	var hours float64
	_, err := fmt.Sscanf(hoursStr, "%f", &hours)
	if err != nil || hours <= 0 {
		m.ErrorMessage = "Hours must be a positive number"
		m.FocusedInput = 2
		m.updateInputFocus()
		return *m, nil
	}

	if len(m.WorkhourDetails) == 0 || m.SelectedDetailsIndex >= len(m.WorkhourDetails) {
		m.ErrorMessage = "Please select a workhour type"
		m.FocusedInput = 0
		return *m, nil
	}
	detailsID := m.WorkhourDetails[m.SelectedDetailsIndex].ID

	if m.SelectedProjectIndex < 0 || m.SelectedProjectIndex >= len(m.Projects) {
		m.ErrorMessage = "Please select a project"
		m.FocusedInput = 1
		return *m, nil
	}
	projectID := m.Projects[m.SelectedProjectIndex].ID

	m.Mode = ModeView
	return *m, tea.Batch(
		dispatchWorkhourCreatedMsg(m.Date, detailsID, projectID, hours),
	)
}

func (m *WorkhoursViewModal) View(Width, Height int) string {
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

func (m *WorkhoursViewModal) viewCreateMode(Width, Height int) string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("Add Work Hours"))
	sb.WriteString("\n")

	dateStr := m.Date.Format("Monday, January 2, 2006")
	sb.WriteString(dateStyle.Render(dateStr))
	sb.WriteString("\n\n")

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

	sb.WriteString(labelStyle.Render("Project:"))
	sb.WriteString("\n")

	if len(m.Projects) == 0 {
		sb.WriteString(render.RenderHelpText("No projects available", "Create one in tab 2"))
		sb.WriteString("\n")
	} else {
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

	sb.WriteString(labelStyle.Render("Hours:"))
	sb.WriteString("\n")
	if m.FocusedInput == 2 {
		sb.WriteString("▶ ")
	} else {
		sb.WriteString("  ")
	}
	sb.WriteString(m.HoursInput.View())
	sb.WriteString("\n\n")

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
	if m.ErrorMessage != "" {
		sb.WriteString(errorStyle.Render("⚠ " + m.ErrorMessage))
	} else {
		sb.WriteString(" ")
	}
	sb.WriteString("\n\n")

	sb.WriteString(render.RenderHelpText("Tab: next",
		"↑/↓: select", "Enter: save", "ESC: cancel"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func (m *WorkhoursViewModal) viewEditMode(Width, Height int) string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("Edit Work Hours"))
	sb.WriteString("\n")

	dateStr := m.Date.Format("Monday, January 2, 2006")
	sb.WriteString(dateStyle.Render(dateStr))
	sb.WriteString("\n\n")

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

	sb.WriteString(labelStyle.Render("Project:"))
	sb.WriteString("\n")

	if len(m.Projects) == 0 {
		sb.WriteString(render.RenderHelpText("No projects available",  "Create one in tab 2"))
		sb.WriteString("\n")
	} else {
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

	sb.WriteString(labelStyle.Render("Hours:"))
	sb.WriteString("\n")
	if m.FocusedInput == 2 {
		sb.WriteString("▶ ")
	} else {
		sb.WriteString("  ")
	}
	sb.WriteString(m.HoursInput.View())
	sb.WriteString("\n\n")

	if m.ErrorMessage != "" {
		sb.WriteString(errorStyle.Render("⚠ " + m.ErrorMessage))
	} else {
		sb.WriteString(" ")
	}
	sb.WriteString("\n\n")

	sb.WriteString(render.RenderHelpText("Tab: next",
		"↑/↓: select",
		"Enter: save",
		"ESC: cancel"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func (m *WorkhoursViewModal) viewDeleteMode(Width, Height int) string {
	if m.SelectedWorkhourIndex < 0 || m.SelectedWorkhourIndex >= len(m.Workhours) {
		return ""
	}

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

	wh := m.Workhours[m.SelectedWorkhourIndex]

	var detailsName string
	var detailsShortName string
	for _, wd := range m.WorkhourDetails {
		if wd.ID == wh.DetailsID {
			detailsName = wd.Name
			detailsShortName = wd.ShortName
			break
		}
	}

	var projectName string
	for _, p := range m.Projects {
		if p.ID == wh.ProjectID {
			projectName = p.Name
			break
		}
	}

	hoursStr := fmt.Sprintf("%.1f", wh.Hours)
	if wh.Hours == float64(int(wh.Hours)) {
		hoursStr = fmt.Sprintf("%d", int(wh.Hours))
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

	// domain.Project
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

type WorkhourEditedMsg struct {
	ID        int // Database ID of the workhour
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

func dispatchWorkhourEditedMsg(id int, date time.Time, detailsID int, projectID int, hours float64) tea.Cmd {
	return func() tea.Msg {
		return WorkhourEditedMsg{
			ID:        id,
			Date:      date,
			DetailsID: detailsID,
			ProjectID: projectID,
			Hours:     hours,
		}
	}
}

type WorkhourDeletedMsg struct {
	ID   int
	Date time.Time
}

func dispatchWorkhourDeletedMsg(id int, date time.Time) tea.Cmd {
	return func() tea.Msg {
		return WorkhourDeletedMsg{
			ID:   id,
			Date: date,
		}
	}
}

func (m *WorkhoursViewModal) initializeCreateMode() {
	m.Mode = ModeCreate
	m.HoursInput.SetValue("")
	m.HoursInput.Focus()
	m.SelectedDetailsIndex = 0

	if projects, err := repository.GetAllProjectsFromDB(); err == nil {
		m.Projects = projects
	}
	if details, err := repository.GetAllWorkhourDetailsFromDB(); err == nil {
		m.WorkhourDetails = details
	}

	if len(m.Projects) > 0 {
		m.SelectedProjectIndex = 0
	} else {
		m.SelectedProjectIndex = -1
	}
	m.FocusedInput = 0
	m.ErrorMessage = ""
	m.updateInputFocus()
}

func (m *WorkhoursViewModal) initializeEditMode() {
	if len(m.Workhours) == 0 || m.SelectedWorkhourIndex >= len(m.Workhours) {
		return
	}

	m.Mode = ModeEdit
	wh := m.Workhours[m.SelectedWorkhourIndex]

	if projects, err := repository.GetAllProjectsFromDB(); err == nil {
		m.Projects = projects
	}
	if details, err := repository.GetAllWorkhourDetailsFromDB(); err == nil {
		m.WorkhourDetails = details
	}

	hoursStr := fmt.Sprintf("%.1f", wh.Hours)
	if wh.Hours == float64(int(wh.Hours)) {
		hoursStr = fmt.Sprintf("%d", int(wh.Hours))
	}
	m.HoursInput.SetValue(hoursStr)
	m.HoursInput.Focus()

	m.SelectedDetailsIndex = 0
	for i, details := range m.WorkhourDetails {
		if details.ID == wh.DetailsID {
			m.SelectedDetailsIndex = i
			break
		}
	}

	m.SelectedProjectIndex = -1
	for i, project := range m.Projects {
		if project.ID == wh.ProjectID {
			m.SelectedProjectIndex = i
			break
		}
	}

	m.FocusedInput = 0
	m.ErrorMessage = ""
	m.updateInputFocus()
}

func (m *WorkhoursViewModal) updateInputFocus() {
	if m.FocusedInput == 2 {
		m.HoursInput.Focus()
	} else {
		m.HoursInput.Blur()
	}
}
