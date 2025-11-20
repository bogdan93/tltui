package calendar

import (
	"fmt"
	"strings"
	"time"
	"tltui/src/common"
	"tltui/src/domain"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CalendarModel struct {
	Width        int
	Height       int
	SelectedDate time.Time
	ViewMonth    int // Month being viewed (1-12)
	ViewYear     int // Year being viewed

	ActiveModal     CalendarModal               // Currently displayed modal
	ViewModalParent *WorkhoursViewModalWrapper // Saved view modal when CRUD modals are open
	ShowHelp        bool

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
		m.ActiveModal = nil
		return m, nil

	case ReportGeneratorModalClosedMsg:
		m.ActiveModal = nil
		return m, nil

	case ReportGeneratedMsg:
		m.ActiveModal = nil
		return m, nil

	case ReportGenerationFailedMsg:
		m.ActiveModal = nil
		return m, common.NotifyError("Failed to generate report", msg.Error)

	// View modal requests
	case WorkhoursViewModalCreateRequestedMsg:
		return m.handleWorkhourCreateRequest(msg)

	case WorkhoursViewModalEditRequestedMsg:
		return m.handleWorkhourEditRequest(msg)

	case WorkhoursViewModalDeleteRequestedMsg:
		return m.handleWorkhourDeleteRequest(msg)

	// Create modal
	case WorkhourCreateSubmittedMsg:
		return m.handleWorkhourCreated(msg)

	case WorkhourCreateCanceledMsg:
		// Restore view modal without refresh
		if m.ViewModalParent != nil {
			m.ActiveModal = m.ViewModalParent
			m.ViewModalParent = nil
		} else {
			m.ActiveModal = nil
		}
		return m, nil

	// Edit modal
	case WorkhourEditSubmittedMsg:
		return m.handleWorkhourEdited(msg)

	case WorkhourEditCanceledMsg:
		// Restore view modal without refresh
		if m.ViewModalParent != nil {
			m.ActiveModal = m.ViewModalParent
			m.ViewModalParent = nil
		} else {
			m.ActiveModal = nil
		}
		return m, nil

	// Delete modal
	case WorkhourDeleteConfirmedMsg:
		return m.handleWorkhourDeleted(msg)

	case WorkhourDeleteCanceledMsg:
		// Restore view modal without refresh
		if m.ViewModalParent != nil {
			m.ActiveModal = m.ViewModalParent
			m.ViewModalParent = nil
		} else {
			m.ActiveModal = nil
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

		if m.ActiveModal != nil {
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
			return m.handleYankWorkhours()

		case "p":
			return m.handlePasteWorkhours()

		case "d", "x":
			return m.handleDeleteWorkhours()

		case "g":
			return m.handleOpenReportGenerator()

		case "enter":
			return m.handleOpenDayView()
		}
	}

	// Delegate to active modal if any
	if m.ActiveModal != nil {
		updatedModal, cmd := m.ActiveModal.Update(msg)
		m.ActiveModal = updatedModal
		return m, cmd
	}

	return m, nil
}

func (m CalendarModel) View() string {
	if m.ShowHelp {
		return m.renderHelpModal()
	}

	if m.ActiveModal != nil {
		return m.ActiveModal.View(m.Width, m.Height)
	}

	var sb strings.Builder

	availableWidth := max(m.Width-6, 70)
	cellWidth := availableWidth / 7
	cellHeight := (m.Height - 8) / 7

	monthName := time.Month(m.ViewMonth).String()
	header := fmt.Sprintf("%s %d", monthName, m.ViewYear)
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Width(m.Width).
		Align(lipgloss.Center)

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
			isCoppiedDate := m.isSameDay(cellDay, m.YankedFromDate)

			var cellStyle lipgloss.Style

			baseStyle := lipgloss.NewStyle().
				Width(cellWidth).
				Height(cellHeight).
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
			} else if isCoppiedDate {
				cellStyle = baseStyle.
					Foreground(lipgloss.Color("114")).
					BorderForeground(lipgloss.Color("114"))
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
					project := m.getProjectByID(wh.ProjectID)

					if details != nil {
						hoursStr := fmt.Sprintf("%.1f", wh.Hours)
						if wh.Hours == float64(int(wh.Hours)) {
							hoursStr = fmt.Sprintf("%d", int(wh.Hours))
						}

						prefixLen := len(details.ShortName) + len(hoursStr)
						availableSpace := cellWidth - prefixLen - 2

						projectName := project.Name
						if len(projectName) > availableSpace {
							if availableSpace > 1 {
								projectName = projectName[:availableSpace-1] + "…"
							} else {
								projectName = "…"
							}
						}
						workhourLines = append(
							workhourLines,
							fmt.Sprintf("%s%sh %s", details.ShortName, hoursStr, projectName),
						)
					}
				}
			}

			// Truncate workhour lines if they exceed cell height
			// -1 for day number line
			maxLines := cellHeight - 1
			if len(workhourLines) > maxLines && maxLines > 0 {
				if maxLines > 1 {
					workhourLines = workhourLines[:maxLines-1]
					workhourLines = append(workhourLines, "…")
				} else {
					workhourLines = []string{"…"}
				}
			}

			workhourText := strings.Join(workhourLines, "\n")
			formattedContent := cellContent + "\n" + workhourText
			dayCells = append(dayCells, cellStyle.Render(formattedContent))
		}

		weekRows = append(weekRows, lipgloss.JoinHorizontal(lipgloss.Top, dayCells...))
	}

	sb.WriteString(lipgloss.JoinVertical(lipgloss.Left, weekRows...))
	sb.WriteString("\n")

	helpText := render.RenderHelpText("←/→: day", "↑/↓: week", "</>: month", "?: help")
	sb.WriteString("\n")
	sb.WriteString(helpText)

	return sb.String()
}

