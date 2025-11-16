package calendar

import (
	"fmt"
	"strings"
	"time"
	"tltui/src/common"
	"tltui/src/domain"
	"tltui/src/domain/repository"
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
	ShowHelp               bool

	YankedWorkhours []domain.Workhour
	YankedFromDate  time.Time 
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
		m.ReportGeneratorModal = nil
		return m, nil

	case ReportGenerationFailedMsg:
		m.ReportGeneratorModal = nil
		return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to generate report: %v", msg.Error))

	case WorkhourCreatedMsg:
		newWorkhour := domain.Workhour{
			Date:      msg.Date,
			DetailsID: msg.DetailsID,
			ProjectID: msg.ProjectID,
			Hours:     msg.Hours,
		}
		_, err := repository.CreateWorkhour(newWorkhour)
		if err != nil {
			return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to create workhour: %v", err))
		}

		if m.WorkhoursViewModal != nil {
			m.WorkhoursViewModal.Workhours = m.getWorkhoursForDate(m.WorkhoursViewModal.Date)
		}

		return m, nil

	case WorkhourEditedMsg:
		updatedWorkhour := domain.Workhour{
			Date:      msg.Date,
			DetailsID: msg.DetailsID,
			ProjectID: msg.ProjectID,
			Hours:     msg.Hours,
		}
		err := repository.UpdateWorkhour(msg.ID, updatedWorkhour)
		if err != nil {
			return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to update workhour: %v", err))
		}

		if m.WorkhoursViewModal != nil {
			m.WorkhoursViewModal.Workhours = m.getWorkhoursForDate(m.WorkhoursViewModal.Date)
		}

		return m, nil

	case WorkhourDeletedMsg:
		err := repository.DeleteWorkhour(msg.ID)
		if err != nil {
			return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to delete workhour: %v", err))
		}

		if m.WorkhoursViewModal != nil {
			m.WorkhoursViewModal.Workhours = m.getWorkhoursForDate(m.WorkhoursViewModal.Date)
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
		if m.ShowHelp {
			switch msg.String() {
			case "?", "esc", "q":
				m.ShowHelp = false
				return m, nil
			}
			return m, nil
		}

		if msg.String() == "?" {
			m.ShowHelp = true
			return m, nil
		}

		if m.WorkhoursViewModal != nil || m.ReportGeneratorModal != nil {
			break
		}

		switch msg.String() {
		case "left", "h":
			newDate := m.SelectedDate.AddDate(0, 0, -1)
			if m.isDateInVisibleGrid(newDate) {
				m.SelectedDate = newDate
			}
			return m, nil

		case "right", "l":
			newDate := m.SelectedDate.AddDate(0, 0, 1)
			if m.isDateInVisibleGrid(newDate) {
				m.SelectedDate = newDate
			}
			return m, nil

		case "up", "k":
			newDate := m.SelectedDate.AddDate(0, 0, -7)
			if m.isDateInVisibleGrid(newDate) {
				m.SelectedDate = newDate
			}
			return m, nil

		case "down", "j":
			newDate := m.SelectedDate.AddDate(0, 0, 7)
			if m.isDateInVisibleGrid(newDate) {
				m.SelectedDate = newDate
			}
			return m, nil

		case "<":
			m.SelectedDate = m.SelectedDate.AddDate(0, -1, 0)
			m.ViewMonth = int(m.SelectedDate.Month())
			m.ViewYear = m.SelectedDate.Year()
			return m, nil

		case ">":
			m.SelectedDate = m.SelectedDate.AddDate(0, 1, 0)
			m.ViewMonth = int(m.SelectedDate.Month())
			m.ViewYear = m.SelectedDate.Year()
			return m, nil

		case "r":
			m.ResetToCurrentMonth()
			return m, nil

		case "y":
			m.YankedWorkhours = m.getWorkhoursForDate(m.SelectedDate)
			m.YankedFromDate = m.SelectedDate
			return m, nil

		case "p":
			if len(m.YankedWorkhours) == 0 {
				return m, nil
			}

			err := repository.DeleteWorkhoursByDate(m.SelectedDate)
			if err != nil {
				return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to clear existing workhours: %v", err))
			}

			for _, wh := range m.YankedWorkhours {
				newWorkhour := domain.Workhour{
					Date:      m.SelectedDate,
					DetailsID: wh.DetailsID,
					ProjectID: wh.ProjectID,
					Hours:     wh.Hours,
				}
				_, err := repository.CreateWorkhour(newWorkhour)
				if err != nil {
					return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to paste workhour: %v", err))
				}
			}
			return m, nil

		case "d":
			err := repository.DeleteWorkhoursByDate(m.SelectedDate)
			if err != nil {
				return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to delete workhours: %v", err))
			}
			return m, nil

		case "g":
			if m.ReportGeneratorModal == nil {
				m.ReportGeneratorModal = NewReportGeneratorModal(m.ViewMonth, m.ViewYear)
				return m, nil
			}

		case "enter":
			if m.WorkhoursViewModal == nil {
				workhours := m.getWorkhoursForDate(m.SelectedDate)
				workhourDetails, _ := repository.GetAllWorkhourDetailsFromDB()
				projects, _ := repository.GetAllProjectsFromDB()
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
	if m.ShowHelp {
		return m.renderHelpModal()
	}

	if m.ReportGeneratorModal != nil {
		return m.ReportGeneratorModal.View(m.Width, m.Height)
	}

	if m.WorkhoursViewModal != nil {
		return m.WorkhoursViewModal.View(m.Width, m.Height)
	}

	var sb strings.Builder

	availableWidth := max( m.Width-6, 70)
	cellWidth := availableWidth / 7

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

	grid := m.getCalendarGrid()
	today := time.Now()

	var weekRows []string
	for week := range 6 {
		var dayCells []string

		for day := range 7 {
			cellDay := grid[week][day]

			var cellContent string
			if cellDay.IsZero() {
				cellContent = ""
			} else {
				dayNum := cellDay.Day()
				cellContent = fmt.Sprintf("%2d", dayNum)
			}

			isToday := m.isSameDay(cellDay, today)
			isSelected := m.isSameDay(cellDay, m.SelectedDate)
			isCurrentMonth := cellDay.Month() == time.Month(m.ViewMonth)

			var cellStyle lipgloss.Style
			baseStyle := lipgloss.NewStyle().
				Width(cellWidth).
				Height(3).
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				BorderRight(day < 6).
				BorderBottom(week < 5)

			if isSelected {
				cellStyle = baseStyle.
					Bold(true).
					Foreground(lipgloss.Color("229")).
					Background(lipgloss.Color("57"))
			} else if isToday {
				cellStyle = baseStyle.
					Bold(true).
					Foreground(lipgloss.Color("39"))
			} else if !isCurrentMonth {
				cellStyle = baseStyle.
					Foreground(lipgloss.Color("240"))
			} else {
				cellStyle = baseStyle.
					Foreground(lipgloss.Color("255"))
			}

			var workhourLines []string
			if !cellDay.IsZero() {
				workhours := m.getWorkhoursForDate(cellDay)
				for _, wh := range workhours {
					details := m.getWorkhourDetailsByID(wh.DetailsID)
					if details != nil {
						hoursStr := fmt.Sprintf("%.1f", wh.Hours)
						if wh.Hours == float64(int(wh.Hours)) {
							hoursStr = fmt.Sprintf("%d", int(wh.Hours))
						}
						workhourLines = append(workhourLines, fmt.Sprintf("%s %sh", details.ShortName, hoursStr))
					}
				}
			}

			workhourText := strings.Join(workhourLines, " ")
			formattedContent := cellContent + "\n" + workhourText
			dayCells = append(dayCells, cellStyle.Render(formattedContent))
		}

		weekRows = append(weekRows, lipgloss.JoinHorizontal(lipgloss.Top, dayCells...))
	}

	sb.WriteString(lipgloss.JoinVertical(lipgloss.Left, weekRows...))
	sb.WriteString("\n")

	helpItems := []string{"←/→: day", "↑/↓: week", "</>: month", "?: help"}
	helpText := render.RenderHelpText(helpItems...)
	sb.WriteString("\n")
	sb.WriteString(helpText)

	return sb.String()
}

func (m *CalendarModel) ResetToCurrentMonth() {
	now := time.Now()
	m.SelectedDate = now
	m.ViewMonth = int(now.Month())
	m.ViewYear = now.Year()
}

func (m CalendarModel) getCalendarGrid() [6][7]time.Time {
	var grid [6][7]time.Time

	firstDay := time.Date(m.ViewYear, time.Month(m.ViewMonth), 1, 0, 0, 0, 0, time.Local)

	weekday := int(firstDay.Weekday()) // 0 = Sunday, 1 = Monday, etc.
	daysFromMonday := (weekday + 6) % 7
	startDate := firstDay.AddDate(0, 0, -daysFromMonday)

	currentDate := startDate
	for week := range 6 {
		for day := range 7 {
			grid[week][day] = currentDate
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}

	return grid
}

func (m CalendarModel) isSameDay(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func (m CalendarModel) isDateInVisibleGrid(date time.Time) bool {
	grid := m.getCalendarGrid()

	normalizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)

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

func (m CalendarModel) getWorkhoursForDate(date time.Time) []domain.Workhour {
	workhours, err := repository.GetWorkhoursByDate(date)
	if err != nil {
		return []domain.Workhour{}
	}
	return workhours
}

func (m CalendarModel) getWorkhourDetailsByID(id int) *domain.WorkhourDetails {
	details, err := repository.GetWorkhourDetailsByID(id)
	if err != nil {
		return nil
	}
	return details
}

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
