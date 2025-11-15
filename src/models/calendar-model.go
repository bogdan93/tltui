package models

import (
	"fmt"
	"strings"
	"time"
	"time-logger-tui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CalendarModel struct {
	Width        int
	Height       int
	SelectedDate time.Time
	ViewMonth    int // Month being viewed (1-12)
	ViewYear     int // Year being viewed
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
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			// Move selection left one day
			m.SelectedDate = m.SelectedDate.AddDate(0, 0, -1)
			return m, nil

		case "right", "l":
			// Move selection right one day
			m.SelectedDate = m.SelectedDate.AddDate(0, 0, 1)
			return m, nil

		case "up", "k":
			// Move selection up one week
			m.SelectedDate = m.SelectedDate.AddDate(0, 0, -7)
			return m, nil

		case "down", "j":
			// Move selection down one week
			m.SelectedDate = m.SelectedDate.AddDate(0, 0, 7)
			return m, nil

		case "p":
			// Previous month
			m.ViewMonth--
			if m.ViewMonth < 1 {
				m.ViewMonth = 12
				m.ViewYear--
			}
			return m, nil

		case "n":
			// Next month
			m.ViewMonth++
			if m.ViewMonth > 12 {
				m.ViewMonth = 1
				m.ViewYear++
			}
			return m, nil

		case "r":
			// Reset to current month
			m.ResetToCurrentMonth()
			return m, nil
		}
	}

	return m, nil
}

// updateViewToSelectedDate updates the view month/year to match the selected date
func (m *CalendarModel) updateViewToSelectedDate() {
	m.ViewMonth = int(m.SelectedDate.Month())
	m.ViewYear = m.SelectedDate.Year()
}

func (m CalendarModel) View() string {
	var sb strings.Builder

	// Calculate cell width based on available width
	// Reserve space for padding and borders
	availableWidth := m.Width - 6 // Account for global padding
	if availableWidth < 70 {
		availableWidth = 70 // Minimum width
	}
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

	weekdays := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
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
	for week := 0; week < 6; week++ {
		var dayCells []string

		// Render each day in the week
		for day := 0; day < 7; day++ {
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

			// Format cell content with day number at top and space for workhours
			formattedContent := cellContent + "\n\n"
			dayCells = append(dayCells, cellStyle.Render(formattedContent))
		}

		// Join cells horizontally to create a week row
		weekRows = append(weekRows, lipgloss.JoinHorizontal(lipgloss.Top, dayCells...))
	}

	// Join all week rows vertically
	sb.WriteString(lipgloss.JoinVertical(lipgloss.Left, weekRows...))
	sb.WriteString("\n")

	// Add help text
	helpText := render.RenderHelpText("←/→: day", "↑/↓: week", "p/n: month", "r: today", "q: quit")
	sb.WriteString("\n")
	sb.WriteString(helpText)

	return sb.String()
}

// ResetToCurrentMonth resets the view to the current month
func (m *CalendarModel) ResetToCurrentMonth() {
	now := time.Now()
	m.ViewMonth = int(now.Month())
	m.ViewYear = now.Year()
}

// getCalendarGrid returns a 6x7 grid of dates for the calendar
// Includes days from previous and next months to fill the grid
func (m CalendarModel) getCalendarGrid() [6][7]time.Time {
	var grid [6][7]time.Time

	// First day of the viewing month
	firstDay := time.Date(m.ViewYear, time.Month(m.ViewMonth), 1, 0, 0, 0, 0, time.Local)

	// Find the Sunday before or on the first day of the month
	weekday := int(firstDay.Weekday()) // 0 = Sunday, 1 = Monday, etc.
	startDate := firstDay.AddDate(0, 0, -weekday)

	// Fill the grid
	currentDate := startDate
	for week := 0; week < 6; week++ {
		for day := 0; day < 7; day++ {
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
