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
	Width           int
	Height          int
	SelectedDate    time.Time
	ViewMonth       int // Month being viewed (1-12)
	ViewYear        int // Year being viewed
	Workhours       []Workhour
	WorkhourDetails []WorkhourDetails
	Projects        []Project

	WorkhoursViewModal *WorkhoursViewModal

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

	case WorkhourCreatedMsg:
		// Add new workhour to the list
		newWorkhour := Workhour{
			Date:      msg.Date,
			DetailsID: msg.DetailsID,
			ProjectID: msg.ProjectID,
			Hours:     msg.Hours,
		}
		m.Workhours = append(m.Workhours, newWorkhour)

		// Update the modal's workhours to show the new entry
		if m.WorkhoursViewModal != nil {
			m.WorkhoursViewModal.Workhours = m.getWorkhoursForDate(m.WorkhoursViewModal.Date)
		}

		return m, nil

	case WorkhourEditedMsg:
		// Find and update the workhour
		dayWorkhours := m.getWorkhoursForDate(msg.Date)
		if msg.Index >= 0 && msg.Index < len(dayWorkhours) {
			// Find this workhour in the main array and update it
			targetWh := dayWorkhours[msg.Index]
			for i := range m.Workhours {
				if m.Workhours[i].Date.Equal(targetWh.Date) &&
					m.Workhours[i].DetailsID == targetWh.DetailsID &&
					m.Workhours[i].ProjectID == targetWh.ProjectID &&
					m.Workhours[i].Hours == targetWh.Hours {
					// Update the workhour
					m.Workhours[i].DetailsID = msg.DetailsID
					m.Workhours[i].ProjectID = msg.ProjectID
					m.Workhours[i].Hours = msg.Hours
					break
				}
			}

			// Update the modal's workhours to show the updated entry
			if m.WorkhoursViewModal != nil {
				m.WorkhoursViewModal.Workhours = m.getWorkhoursForDate(m.WorkhoursViewModal.Date)
			}
		}

		return m, nil

	case WorkhourDeletedMsg:
		// Find and delete the workhour
		dayWorkhours := m.getWorkhoursForDate(msg.Date)
		if msg.Index >= 0 && msg.Index < len(dayWorkhours) {
			// Find this workhour in the main array and delete it
			targetWh := dayWorkhours[msg.Index]
			for i := range m.Workhours {
				if m.Workhours[i].Date.Equal(targetWh.Date) &&
					m.Workhours[i].DetailsID == targetWh.DetailsID &&
					m.Workhours[i].ProjectID == targetWh.ProjectID &&
					m.Workhours[i].Hours == targetWh.Hours {
					// Delete the workhour by removing it from the slice
					m.Workhours = append(m.Workhours[:i], m.Workhours[i+1:]...)
					break
				}
			}

			// Update the modal's workhours to show the updated list
			if m.WorkhoursViewModal != nil {
				m.WorkhoursViewModal.Workhours = m.getWorkhoursForDate(m.WorkhoursViewModal.Date)
				// Adjust selected index if needed
				if m.WorkhoursViewModal.SelectedWorkhourIndex >= len(m.WorkhoursViewModal.Workhours) && len(m.WorkhoursViewModal.Workhours) > 0 {
					m.WorkhoursViewModal.SelectedWorkhourIndex = len(m.WorkhoursViewModal.Workhours) - 1
				}
			}
		}

		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Don't handle navigation keys if modal is open
		if m.WorkhoursViewModal != nil {
			break
		}

		switch msg.String() {
		case "left", "h":
			// Move selection left one day
			m.SelectedDate = m.SelectedDate.AddDate(0, 0, -1)
			// Update view if moved to different month
			m.ViewMonth = int(m.SelectedDate.Month())
			m.ViewYear = m.SelectedDate.Year()
			return m, nil

		case "right", "l":
			// Move selection right one day
			m.SelectedDate = m.SelectedDate.AddDate(0, 0, 1)
			// Update view if moved to different month
			m.ViewMonth = int(m.SelectedDate.Month())
			m.ViewYear = m.SelectedDate.Year()
			return m, nil

		case "up", "k":
			// Move selection up one week
			m.SelectedDate = m.SelectedDate.AddDate(0, 0, -7)
			// Update view if moved to different month
			m.ViewMonth = int(m.SelectedDate.Month())
			m.ViewYear = m.SelectedDate.Year()
			return m, nil

		case "down", "j":
			// Move selection down one week
			m.SelectedDate = m.SelectedDate.AddDate(0, 0, 7)
			// Update view if moved to different month
			m.ViewMonth = int(m.SelectedDate.Month())
			m.ViewYear = m.SelectedDate.Year()
			return m, nil

		case "p":
			// Previous month - move selected date and update view
			m.SelectedDate = m.SelectedDate.AddDate(0, -1, 0)
			m.ViewMonth = int(m.SelectedDate.Month())
			m.ViewYear = m.SelectedDate.Year()
			return m, nil

		case "n":
			// Next month - move selected date and update view
			m.SelectedDate = m.SelectedDate.AddDate(0, 1, 0)
			m.ViewMonth = int(m.SelectedDate.Month())
			m.ViewYear = m.SelectedDate.Year()
			return m, nil

		case "r":
			// Reset to current month
			m.ResetToCurrentMonth()
			return m, nil

		case "c":
			// Copy workhours from selected date
			m.YankedWorkhours = m.getWorkhoursForDate(m.SelectedDate)
			m.YankedFromDate = m.SelectedDate
			return m, nil

		case "v":
			// Paste copied workhours to selected date
			if len(m.YankedWorkhours) > 0 {
				// First, remove all existing workhours for the selected date
				var filteredWorkhours []Workhour
				for _, wh := range m.Workhours {
					if !m.isSameDay(wh.Date, m.SelectedDate) {
						filteredWorkhours = append(filteredWorkhours, wh)
					}
				}
				m.Workhours = filteredWorkhours

				// Then add the copied workhours with the new date
				for _, wh := range m.YankedWorkhours {
					newWorkhour := Workhour{
						Date:      m.SelectedDate,
						DetailsID: wh.DetailsID,
						ProjectID: wh.ProjectID,
						Hours:     wh.Hours,
					}
					m.Workhours = append(m.Workhours, newWorkhour)
				}
			}
			return m, nil

		case "enter":
			// Open workhours view modal for selected date
			if m.WorkhoursViewModal == nil {
				workhours := m.getWorkhoursForDate(m.SelectedDate)
				m.WorkhoursViewModal = NewWorkhoursViewModal(
					m.SelectedDate,
					workhours,
					m.WorkhourDetails,
					m.Projects,
				)
				return m, nil
			}
		}
	}

	// Route messages to modal if open
	if m.WorkhoursViewModal != nil {
		_, cmd := m.WorkhoursViewModal.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m CalendarModel) View() string {
	// Show modal if open (replaces calendar view)
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
			isCopied := len(m.YankedWorkhours) > 0 && m.isSameDay(cellDay, m.YankedFromDate)
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
			} else if isCopied {
				// Copied day - green border to indicate clipboard source
				cellStyle = baseStyle.
					Bold(true).
					Foreground(lipgloss.Color("114")). // Green
					BorderStyle(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("114"))
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

	// Add help text
	var helpItems []string
	if len(m.YankedWorkhours) > 0 {
		// Show paste hint when workhours are copied
		copyCount := fmt.Sprintf("%d copied", len(m.YankedWorkhours))
		helpItems = []string{"←/→: day", "↑/↓: week", "p/n: month", "c: copy", "v: paste", copyCount, "q: quit"}
	} else {
		helpItems = []string{"←/→: day", "↑/↓: week", "p/n: month", "c: copy", "q: quit"}
	}
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

// getWorkhoursForDate returns all workhours for a specific date
func (m CalendarModel) getWorkhoursForDate(date time.Time) []Workhour {
	var result []Workhour
	for _, wh := range m.Workhours {
		if m.isSameDay(wh.Date, date) {
			result = append(result, wh)
		}
	}
	return result
}

// getWorkhourDetailsByID returns the WorkhourDetails for a given ID
func (m CalendarModel) getWorkhourDetailsByID(id int) *WorkhourDetails {
	for i := range m.WorkhourDetails {
		if m.WorkhourDetails[i].ID == id {
			return &m.WorkhourDetails[i]
		}
	}
	return nil
}
