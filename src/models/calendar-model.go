package models

import (
	"fmt"
	"strings"
	"time"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CalendarModel struct {
	Width           int
	Height          int
	SelectedDate    time.Time
	ViewMonth       int // Month being viewed (1-12)
	ViewYear        int // Year being viewed

	WorkhoursViewModal     *WorkhoursViewModal
	ReportGeneratorModal   *ReportGeneratorModal
	ShowHelp               bool // Show help modal

	// Copy/Paste clipboard
	YankedWorkhours []Workhour // Workhours copied from a day
	YankedFromDate  time.Time  // The date they were copied from (for feedback)
}

func NewCalendarModel() CalendarModel {
	now := time.Now()
	return CalendarModel{
		SelectedDate: now,
		ViewMonth:    int(now.Month()),
		ViewYear:     now.Year(),
	}
}

func (m CalendarModel) Init() tea.Cmd {
	return nil
}

func (m CalendarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case WorkhoursViewModalClosedMsg:
		m.WorkhoursViewModal = nil
		return m, nil

	case ReportGeneratorModalClosedMsg:
		m.ReportGeneratorModal = nil
		return m, nil

	case ReportGeneratedMsg:
		// Report generated successfully, close modal
		m.ReportGeneratorModal = nil
		return m, nil

	case ReportGenerationFailedMsg:
		// Report generation failed, close modal and show error
		m.ReportGeneratorModal = nil
		return m, DispatchErrorNotification(fmt.Sprintf("Failed to generate report: %v", msg.Error))

	case WorkhourCreatedMsg:
		// Add new workhour to the database
		newWorkhour := Workhour{
			Date:      msg.Date,
			DetailsID: msg.DetailsID,
			ProjectID: msg.ProjectID,
			Hours:     msg.Hours,
		}
		_, err := CreateWorkhour(newWorkhour)
		if err != nil {
			return m, DispatchErrorNotification(fmt.Sprintf("Failed to create workhour: %v", err))
		}

		// Update the modal's workhours to show the new entry
		if m.WorkhoursViewModal != nil {
			m.WorkhoursViewModal.Workhours = m.getWorkhoursForDate(m.WorkhoursViewModal.Date)
		}

		return m, nil

	case WorkhourEditedMsg:
		// Update the workhour in the database
		updatedWorkhour := Workhour{
			Date:      msg.Date,
			DetailsID: msg.DetailsID,
			ProjectID: msg.ProjectID,
			Hours:     msg.Hours,
		}
		err := UpdateWorkhour(msg.ID, updatedWorkhour)
		if err != nil {
			return m, DispatchErrorNotification(fmt.Sprintf("Failed to update workhour: %v", err))
		}

		// Update the modal's workhours to show the updated entry
		if m.WorkhoursViewModal != nil {
			m.WorkhoursViewModal.Workhours = m.getWorkhoursForDate(m.WorkhoursViewModal.Date)
		}

		return m, nil

	case WorkhourDeletedMsg:
		// Delete the workhour from the database
		err := DeleteWorkhour(msg.ID)
		if err != nil {
			return m, DispatchErrorNotification(fmt.Sprintf("Failed to delete workhour: %v", err))
		}

		// Update the modal's workhours to show the updated list
		if m.WorkhoursViewModal != nil {
			m.WorkhoursViewModal.Workhours = m.getWorkhoursForDate(m.WorkhoursViewModal.Date)
			// Adjust selected index if needed
			if m.WorkhoursViewModal.SelectedWorkhourIndex >= len(m.WorkhoursViewModal.Workhours) && len(m.WorkhoursViewModal.Workhours) > 0 {
				m.WorkhoursViewModal.SelectedWorkhourIndex = len(m.WorkhoursViewModal.Workhours) - 1
			}
		}

		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle help modal
		if m.ShowHelp {
			switch msg.String() {
			case "?", "esc", "q":
				m.ShowHelp = false
				return m, nil
			}
			return m, nil
		}

		// Handle help modal toggle when not open
		if msg.String() == "?" {
			m.ShowHelp = true
			return m, nil
		}

		// Don't handle navigation keys if modal is open
		if m.WorkhoursViewModal != nil || m.ReportGeneratorModal != nil {
			break
		}

		switch msg.String() {
		case "left", "h":
			// Move selection left one day (clamped to visible grid)
			newDate := m.SelectedDate.AddDate(0, 0, -1)
			if m.isDateInVisibleGrid(newDate) {
				m.SelectedDate = newDate
			}
			return m, nil

		case "right", "l":
			// Move selection right one day (clamped to visible grid)
			newDate := m.SelectedDate.AddDate(0, 0, 1)
			if m.isDateInVisibleGrid(newDate) {
				m.SelectedDate = newDate
			}
			return m, nil

		case "up", "k":
			// Move selection up one week (clamped to visible grid)
			newDate := m.SelectedDate.AddDate(0, 0, -7)
			if m.isDateInVisibleGrid(newDate) {
				m.SelectedDate = newDate
			}
			return m, nil

		case "down", "j":
			// Move selection down one week (clamped to visible grid)
			newDate := m.SelectedDate.AddDate(0, 0, 7)
			if m.isDateInVisibleGrid(newDate) {
				m.SelectedDate = newDate
			}
			return m, nil

		case "<":
			// Previous month - move selected date and update view
			m.SelectedDate = m.SelectedDate.AddDate(0, -1, 0)
			m.ViewMonth = int(m.SelectedDate.Month())
			m.ViewYear = m.SelectedDate.Year()
			return m, nil

		case ">":
			// Next month - move selected date and update view
			m.SelectedDate = m.SelectedDate.AddDate(0, 1, 0)
			m.ViewMonth = int(m.SelectedDate.Month())
			m.ViewYear = m.SelectedDate.Year()
			return m, nil

		case "r":
			// Reset to current month
			m.ResetToCurrentMonth()
			return m, nil

		case "y":
			// Yank (copy) workhours from selected date
			m.YankedWorkhours = m.getWorkhoursForDate(m.SelectedDate)
			m.YankedFromDate = m.SelectedDate
			return m, nil

		case "p":
			// Paste copied workhours to selected date (replaces existing)
			if len(m.YankedWorkhours) == 0 {
				return m, nil
			}

			// First, delete all existing workhours for the selected date from database
			err := DeleteWorkhoursByDate(m.SelectedDate)
			if err != nil {
				return m, DispatchErrorNotification(fmt.Sprintf("Failed to clear existing workhours: %v", err))
			}

			// Then add the copied workhours with the new date to database
			for _, wh := range m.YankedWorkhours {
				newWorkhour := Workhour{
					Date:      m.SelectedDate,
					DetailsID: wh.DetailsID,
					ProjectID: wh.ProjectID,
					Hours:     wh.Hours,
				}
				_, err := CreateWorkhour(newWorkhour)
				if err != nil {
					return m, DispatchErrorNotification(fmt.Sprintf("Failed to paste workhour: %v", err))
				}
			}
			return m, nil

		case "d":
			// Delete all workhours for selected date
			err := DeleteWorkhoursByDate(m.SelectedDate)
			if err != nil {
				return m, DispatchErrorNotification(fmt.Sprintf("Failed to delete workhours: %v", err))
			}
			return m, nil

		case "g":
			// Open report generator modal
			if m.ReportGeneratorModal == nil {
				m.ReportGeneratorModal = NewReportGeneratorModal(m.ViewMonth, m.ViewYear)
				return m, nil
			}

		case "enter":
			// Open workhours view modal for selected date
			if m.WorkhoursViewModal == nil {
				workhours := m.getWorkhoursForDate(m.SelectedDate)
				workhourDetails, _ := GetAllWorkhourDetailsFromDB()
				projects, _ := GetAllProjectsFromDB()
				m.WorkhoursViewModal = NewWorkhoursViewModal(
					m.SelectedDate,
					workhours,
					workhourDetails,
					projects,
				)
				return m, nil
			}
		}
	}

	// Route messages to modals if open
	if m.ReportGeneratorModal != nil {
		updatedModal, cmd := m.ReportGeneratorModal.Update(msg)
		m.ReportGeneratorModal = &updatedModal
		return m, cmd
	}

	if m.WorkhoursViewModal != nil {
		_, cmd := m.WorkhoursViewModal.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m CalendarModel) View() string {
	// Show help modal if open
	if m.ShowHelp {
		return m.renderHelpModal()
	}

	// Show modals if open (replaces calendar view)
	if m.ReportGeneratorModal != nil {
		return m.ReportGeneratorModal.View(m.Width, m.Height)
	}

	if m.WorkhoursViewModal != nil {
		return m.WorkhoursViewModal.View(m.Width, m.Height)
	}

	var sb strings.Builder

	// Calculate cell width based on available width
	// Reserve space for padding and borders
	availableWidth := max( m.Width-6, 70)
	cellWidth := availableWidth / 7

	// Render month and year header
	monthName := time.Month(m.ViewMonth).String()
	header := fmt.Sprintf("%s %d", monthName, m.ViewYear)
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Width(availableWidth).
		Align(lipgloss.Center).
		MarginBottom(1)
	sb.WriteString(headerStyle.Render(header))
	sb.WriteString("\n")

	// Render weekday headers
	weekdayStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241")).
		Width(cellWidth).
		Align(lipgloss.Center).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("240"))

	weekdays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	var headerCells []string
	for i, wd := range weekdays {
		style := weekdayStyle
		if i < 6 {
			style = style.BorderRight(true)
		}
		headerCells = append(headerCells, style.Render(wd))
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	sb.WriteString("\n")

	// Get calendar grid
	grid := m.getCalendarGrid()
	today := time.Now()

	// Render calendar grid
	var weekRows []string
	for week := range 6 {
		var dayCells []string

		// Render each day in the week
		for day := range 7 {
			cellDay := grid[week][day]

			var cellContent string
			if cellDay.IsZero() {
				cellContent = ""
			} else {
				dayNum := cellDay.Day()
				cellContent = fmt.Sprintf("%2d", dayNum)
			}

			// Determine style based on selection and today
			isToday := m.isSameDay(cellDay, today)
			isSelected := m.isSameDay(cellDay, m.SelectedDate)
			isCurrentMonth := cellDay.Month() == time.Month(m.ViewMonth)

			var cellStyle lipgloss.Style
			// Always set all borders for consistent sizing
			baseStyle := lipgloss.NewStyle().
				Width(cellWidth).
				Height(3).
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				BorderRight(day < 6).
				BorderBottom(week < 5)

			if isSelected {
				// Selected day - bold with background
				cellStyle = baseStyle.
					Bold(true).
					Foreground(lipgloss.Color("229")).
					Background(lipgloss.Color("57"))
			} else if isToday {
				// Today - bold with border
				cellStyle = baseStyle.
					Bold(true).
					Foreground(lipgloss.Color("39"))
			} else if !isCurrentMonth {
				// Days from prev/next month - dimmed
				cellStyle = baseStyle.
					Foreground(lipgloss.Color("240"))
			} else {
				// Regular day
				cellStyle = baseStyle.
					Foreground(lipgloss.Color("255"))
			}

			// Get workhours for this day and format them
			var workhourLines []string
			if !cellDay.IsZero() {
				workhours := m.getWorkhoursForDate(cellDay)
				for _, wh := range workhours {
					details := m.getWorkhourDetailsByID(wh.DetailsID)
					if details != nil {
						// Format hours nicely (remove .0 for whole numbers)
						hoursStr := fmt.Sprintf("%.1f", wh.Hours)
						if wh.Hours == float64(int(wh.Hours)) {
							hoursStr = fmt.Sprintf("%d", int(wh.Hours))
						}
						workhourLines = append(workhourLines, fmt.Sprintf("%s %sh", details.ShortName, hoursStr))
					}
				}
			}

			// Format cell content with day number at top and workhours below
			workhourText := strings.Join(workhourLines, " ")
			formattedContent := cellContent + "\n" + workhourText
			dayCells = append(dayCells, cellStyle.Render(formattedContent))
		}

		// Join cells horizontally to create a week row
		weekRows = append(weekRows, lipgloss.JoinHorizontal(lipgloss.Top, dayCells...))
	}

	// Join all week rows vertically
	sb.WriteString(lipgloss.JoinVertical(lipgloss.Left, weekRows...))
	sb.WriteString("\n")

	// Add minimal help text
	helpItems := []string{"←/→: day", "↑/↓: week", "</>: month", "?: help"}
	helpText := render.RenderHelpText(helpItems...)
	sb.WriteString("\n")
	sb.WriteString(helpText)

	return sb.String()
}

// ResetToCurrentMonth resets the view to the current month and today's date
func (m *CalendarModel) ResetToCurrentMonth() {
	now := time.Now()
	m.SelectedDate = now
	m.ViewMonth = int(now.Month())
	m.ViewYear = now.Year()
}

// getCalendarGrid returns a 6x7 grid of dates for the calendar
// Includes days from previous and next months to fill the grid
func (m CalendarModel) getCalendarGrid() [6][7]time.Time {
	var grid [6][7]time.Time

	// First day of the viewing month
	firstDay := time.Date(m.ViewYear, time.Month(m.ViewMonth), 1, 0, 0, 0, 0, time.Local)

	// Find the Monday before or on the first day of the month
	weekday := int(firstDay.Weekday()) // 0 = Sunday, 1 = Monday, etc.
	// Adjust so Monday = 0, Tuesday = 1, ..., Sunday = 6
	daysFromMonday := (weekday + 6) % 7
	startDate := firstDay.AddDate(0, 0, -daysFromMonday)

	// Fill the grid
	currentDate := startDate
	for week := range 6 {
		for day := range 7 {
			grid[week][day] = currentDate
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}

	return grid
}

// isSameDay checks if two dates are the same day
func (m CalendarModel) isSameDay(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// isDateInVisibleGrid checks if a date is visible in the current calendar grid
func (m CalendarModel) isDateInVisibleGrid(date time.Time) bool {
	grid := m.getCalendarGrid()

	// Normalize all dates to midnight for comparison
	normalizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)

	// Check if date matches any cell in the grid
	for week := 0; week < 6; week++ {
		for day := 0; day < 7; day++ {
			gridDate := grid[week][day]
			if m.isSameDay(normalizedDate, gridDate) {
				return true
			}
		}
	}

	return false
}

// getWorkhoursForDate returns all workhours for a specific date from the database
func (m CalendarModel) getWorkhoursForDate(date time.Time) []Workhour {
	workhours, err := GetWorkhoursByDate(date)
	if err != nil {
		// Return empty slice on error (could log this in the future)
		return []Workhour{}
	}
	return workhours
}

// getWorkhourDetailsByID returns the WorkhourDetails for a given ID from the database
func (m CalendarModel) getWorkhourDetailsByID(id int) *WorkhourDetails {
	details, err := GetWorkhourDetailsByID(id)
	if err != nil {
		return nil
	}
	return details
}

// renderHelpModal renders the help modal with keyboard shortcuts
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
		{"d", "Delete all workhours from selected day"},
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
