package calendar

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"tltui/src/domain"
	"tltui/src/domain/repository"
	generator "tltui/src/elm-store/calendar/report-generator"
	"tltui/src/render"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ReportType int

const (
	ReportTypeOdooCSV ReportType = iota
	ReportTypeMailReport
)

type ReportGeneratorModal struct {
	SelectedReportType int
	ReportTypes        []string
	Generating         bool
	ErrorMessage       string
	ViewMonth          int // Month to generate report for
	ViewYear           int // Year to generate report for

	ShowingInputForm   bool                       // True when showing From/To company inputs
	FromCompanyInput   textinput.Model            // "From Company" text input
	ToCompanyInput     textinput.Model            // "To Company" text input
	InvoiceNameInput   textinput.Model            // "Invoice Name" text input
	SignatureImagePath string                     // Path to signature image file
	FocusedInput       int                        // 0 = FromCompany, 1 = ToCompany, 2 = InvoiceName, 3 = SignatureImage, 4+ = checkbox items
	PreviewStats       *generator.WorkhourStats   // Cached stats for preview display
	SelectedItems      map[string]map[string]bool // project -> activity -> selected
	FocusedItemIndex   int                        // Index of focused checkbox item in the flattened list
}

type ReportGeneratorModalClosedMsg struct{}
type ReportGeneratedMsg struct {
	FilePath string
}
type ReportGenerationFailedMsg struct {
	Error error
}

func NewReportGeneratorModal(viewMonth, viewYear int) *ReportGeneratorModal {
	fromCompanyInput := textinput.New()
	fromCompanyInput.Placeholder = "From Company"
	fromCompanyInput.CharLimit = 64
	fromCompanyInput.Width = 40

	toCompanyInput := textinput.New()
	toCompanyInput.Placeholder = "To Company"
	toCompanyInput.CharLimit = 64
	toCompanyInput.Width = 40

	invoiceNameInput := textinput.New()
	invoiceNameInput.Placeholder = "Invoice Name"
	invoiceNameInput.CharLimit = 64
	invoiceNameInput.Width = 40

	return &ReportGeneratorModal{
		SelectedReportType: 0,
		ReportTypes:        []string{"Odoo CSV", "Mail Report"},
		Generating:         false,
		ViewMonth:          viewMonth,
		ViewYear:           viewYear,
		ShowingInputForm:   false,
		FromCompanyInput:   fromCompanyInput,
		ToCompanyInput:     toCompanyInput,
		InvoiceNameInput:   invoiceNameInput,
		FocusedInput:       0,
		PreviewStats:       nil,
		SelectedItems:      make(map[string]map[string]bool),
		FocusedItemIndex:   -1,
	}
}

func (m ReportGeneratorModal) Init() tea.Cmd {
	return nil
}

func (m ReportGeneratorModal) Update(msg tea.Msg) (ReportGeneratorModal, tea.Cmd) {
	if m.Generating {
		return m, nil
	}

	if m.ShowingInputForm {
		return m.handleInputForm(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, dispatchReportGeneratorModalClosedMsg()

		case "up", "k":
			if m.SelectedReportType > 0 {
				m.SelectedReportType--
			}
			return m, nil

		case "down", "j":
			if m.SelectedReportType < len(m.ReportTypes)-1 {
				m.SelectedReportType++
			}
			return m, nil

		case "enter":
			if m.ReportTypes[m.SelectedReportType] == "Mail Report" {
				m.ShowingInputForm = true
				m.FocusedInput = 0
				m.FocusedItemIndex = -1
				m.PreviewStats = m.calculatePreviewStats()
				m.initializeSelectedItems()
				m.updateInputFocus()
				return m, nil
			}
			m.Generating = true
			return m, m.generateReport()

		case "o", "O":
			m.SelectedReportType = 0
			m.Generating = true
			return m, m.generateReport()

		case "m", "M":
			m.SelectedReportType = 1
			m.ShowingInputForm = true
			m.FocusedInput = 0
			m.FocusedItemIndex = -1
			m.PreviewStats = m.calculatePreviewStats()
			m.initializeSelectedItems()
			m.updateInputFocus()
			return m, nil
		}
	}

	return m, nil
}

func (m *ReportGeneratorModal) initializeSelectedItems() {
	if m.PreviewStats == nil {
		return
	}

	// Get workhour details to check IsWork property
	workhourDetails, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		return
	}

	detailsMap := make(map[string]domain.WorkhourDetails)
	for _, wd := range workhourDetails {
		detailsMap[wd.Name] = wd
	}

	m.SelectedItems = make(map[string]map[string]bool)

	for projectName, activities := range m.PreviewStats.ProjectActivityHours {
		m.SelectedItems[projectName] = make(map[string]bool)
		for activityName := range activities {
			// Only preselect if IsWork is true
			if detail, exists := detailsMap[activityName]; exists && detail.IsWork {
				m.SelectedItems[projectName][activityName] = true
			} else {
				m.SelectedItems[projectName][activityName] = false
			}
		}
	}
}

func (m ReportGeneratorModal) calculatePreviewStats() *generator.WorkhourStats {
	startDate := time.Date(m.ViewYear, time.Month(m.ViewMonth), 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(m.ViewYear, time.Month(m.ViewMonth+1), 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)

	workhours, err := repository.GetWorkhoursByDateRange(startDate, endDate)
	if err != nil {
		return nil
	}

	workhourDetails, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		return nil
	}

	projects, err := repository.GetAllProjectsFromDB()
	if err != nil {
		return nil
	}

	detailsMap := make(map[int]domain.WorkhourDetails)
	for _, wd := range workhourDetails {
		detailsMap[wd.ID] = wd
	}

	projectsMap := make(map[int]domain.Project)
	for _, p := range projects {
		projectsMap[p.ID] = p
	}

	stats := generator.CalculateWorkhourStats(workhours, detailsMap, projectsMap)
	return &stats
}

func (m ReportGeneratorModal) handleInputForm(msg tea.Msg) (ReportGeneratorModal, tea.Cmd) {
	totalCheckboxItems := m.getTotalCheckboxItems()
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.ErrorMessage = ""

			fromCompany := strings.TrimSpace(m.FromCompanyInput.Value())
			if fromCompany == "" {
				m.ErrorMessage = "From Company is required"
				m.FocusedInput = 0
				m.updateInputFocus()
				return m, nil
			}

			toCompany := strings.TrimSpace(m.ToCompanyInput.Value())
			if toCompany == "" {
				m.ErrorMessage = "To Company is required"
				m.FocusedInput = 1
				m.updateInputFocus()
				return m, nil
			}

			invoiceName := strings.TrimSpace(m.InvoiceNameInput.Value())
			if invoiceName == "" {
				m.ErrorMessage = "Invoice Name is required"
				m.FocusedInput = 2
				m.updateInputFocus()
				return m, nil
			}

			m.ShowingInputForm = false
			m.Generating = true
			return m, m.generateReport()

		case "esc":
			m.ShowingInputForm = false
			m.FromCompanyInput.SetValue("")
			m.ToCompanyInput.SetValue("")
			m.InvoiceNameInput.SetValue("")
			m.SignatureImagePath = ""
			m.FocusedInput = 0
			m.FocusedItemIndex = -1
			m.ErrorMessage = ""
			return m, nil

		case "tab", "down", "j":
			if m.FocusedInput < 3 {
				m.FocusedInput++
				m.updateInputFocus()
			} else if m.FocusedInput == 3 {
				if totalCheckboxItems > 0 {
					m.FocusedInput = 4
					m.FocusedItemIndex = 0
				}
			} else {
				if m.FocusedItemIndex < totalCheckboxItems-1 {
					m.FocusedItemIndex++
				}
			}
			return m, nil

		case "shift+tab", "up", "k":
			if m.FocusedInput == 4 && m.FocusedItemIndex > 0 {
				m.FocusedItemIndex--
			} else if m.FocusedInput == 4 && m.FocusedItemIndex == 0 {
				m.FocusedInput = 3
				m.FocusedItemIndex = -1
				m.updateInputFocus()
			} else if m.FocusedInput > 0 {
				m.FocusedInput--
				m.updateInputFocus()
			}
			return m, nil

		case " ":
			if m.FocusedInput == 4 && m.FocusedItemIndex >= 0 {
				m.toggleCheckboxAtIndex(m.FocusedItemIndex)
				return m, nil
			}

		case "s":
			if m.FocusedInput == 3 {
				return m, generator.OpenImageFileDialog()
			}
		}

	case generator.SignatureImageSelectedMsg:
		if msg.ImagePath != "" {
			m.SignatureImagePath = msg.ImagePath
		}
		return m, nil
	}

	if m.FocusedInput == 0 {
		m.FromCompanyInput, cmd = m.FromCompanyInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.FocusedInput == 1 {
		m.ToCompanyInput, cmd = m.ToCompanyInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.FocusedInput == 2 {
		m.InvoiceNameInput, cmd = m.InvoiceNameInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *ReportGeneratorModal) updateInputFocus() {
	if m.FocusedInput == 0 {
		m.FromCompanyInput.Focus()
		m.ToCompanyInput.Blur()
		m.InvoiceNameInput.Blur()
	} else if m.FocusedInput == 1 {
		m.FromCompanyInput.Blur()
		m.ToCompanyInput.Focus()
		m.InvoiceNameInput.Blur()
	} else if m.FocusedInput == 2 {
		m.FromCompanyInput.Blur()
		m.ToCompanyInput.Blur()
		m.InvoiceNameInput.Focus()
	} else {
		m.FromCompanyInput.Blur()
		m.ToCompanyInput.Blur()
		m.InvoiceNameInput.Blur()
	}
}

func (m ReportGeneratorModal) getTotalCheckboxItems() int {
	if m.PreviewStats == nil {
		return 0
	}

	count := 0
	for _, activities := range m.PreviewStats.ProjectActivityHours {
		count += len(activities)
	}
	return count
}

func (m *ReportGeneratorModal) toggleCheckboxAtIndex(index int) {
	if m.PreviewStats == nil {
		return
	}

	type projectHour struct {
		name  string
		hours float64
	}
	projects := make([]projectHour, 0, len(m.PreviewStats.ProjectHours))
	for name, hours := range m.PreviewStats.ProjectHours {
		projects = append(projects, projectHour{name, hours})
	}
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].hours > projects[j].hours
	})

	currentIndex := 0
	for _, p := range projects {
		if activities, ok := m.PreviewStats.ProjectActivityHours[p.name]; ok {
			type activityHour struct {
				name  string
				hours float64
			}
			activityList := make([]activityHour, 0, len(activities))
			for name, hours := range activities {
				activityList = append(activityList, activityHour{name, hours})
			}
			sort.Slice(activityList, func(i, j int) bool {
				return activityList[i].hours > activityList[j].hours
			})

			for _, a := range activityList {
				if currentIndex == index {
					if m.SelectedItems[p.name] == nil {
						m.SelectedItems[p.name] = make(map[string]bool)
					}
					m.SelectedItems[p.name][a.name] = !m.SelectedItems[p.name][a.name]
					return
				}
				currentIndex++
			}
		}
	}
}

func (m ReportGeneratorModal) View(width, height int) string {
	if m.ShowingInputForm {
		return m.renderInputForm(width, height)
	}

	var sb strings.Builder

	monthName := time.Month(m.ViewMonth).String()
	title := fmt.Sprintf("Generate Report - %s %d", monthName, m.ViewYear)
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center)

	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n\n")

	if m.Generating {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("⏳ Generating report..."))
		sb.WriteString("\n")
	} else if m.ErrorMessage != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		sb.WriteString(errorStyle.Render("⚠ " + m.ErrorMessage))
		sb.WriteString("\n\n")
	} else {
		sb.WriteString("Select report type:\n\n")

		highlightStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))

		for i, reportType := range m.ReportTypes {
			prefix := "  "
			style := lipgloss.NewStyle()

			if i == m.SelectedReportType {
				prefix = "▶ "
				style = style.Bold(true).Foreground(lipgloss.Color("39"))
			}

			if len(reportType) > 0 {
				firstLetter := highlightStyle.Render(string(reportType[0]))
				rest := reportType[1:]
				sb.WriteString(prefix + firstLetter + style.Render(rest) + "\n")
			} else {
				sb.WriteString(prefix + style.Render(reportType) + "\n")
			}
		}

		sb.WriteString("\n")
		helpItems := []string{"↑/↓: select", "o/m: quick select", "enter: generate", "esc/q: cancel"}
		sb.WriteString(render.RenderHelpText(helpItems...))
	}

	return render.RenderSimpleModal(width, height, sb.String())
}

func (m ReportGeneratorModal) renderInputForm(width, height int) string {
	monthName := time.Month(m.ViewMonth).String()

	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center)

	sb.WriteString(titleStyle.Render(fmt.Sprintf("Mail Report - %s %d", monthName, m.ViewYear)))
	sb.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))

	sb.WriteString(labelStyle.Render("From Company:"))
	sb.WriteString("\n")
	sb.WriteString(m.FromCompanyInput.View())
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("To Company:"))
	sb.WriteString("\n")
	sb.WriteString(m.ToCompanyInput.View())
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("Invoice Name:"))
	sb.WriteString("\n")
	sb.WriteString(m.InvoiceNameInput.View())
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("Signature Image:"))
	sb.WriteString("\n")

	isFocused := m.FocusedInput == 3

	if m.SignatureImagePath != "" {
		imageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
		if isFocused {
			imageStyle = imageStyle.Bold(true)
		}
		sb.WriteString(imageStyle.Render("✓ " + filepath.Base(m.SignatureImagePath)))
	} else {
		noImageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		if isFocused {
			noImageStyle = noImageStyle.Bold(true)
		}
		sb.WriteString(noImageStyle.Render("No image selected"))
	}
	sb.WriteString("\n")

	if isFocused {
		buttonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		sb.WriteString(buttonStyle.Render("[Press 's' to select image]"))
	}
	sb.WriteString("\n\n")

	if m.ErrorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		sb.WriteString(errorStyle.Render("⚠ " + m.ErrorMessage))
		sb.WriteString("\n\n")
	}

	if m.PreviewStats != nil {
		sb.WriteString(m.renderStatsPreview())
		sb.WriteString("\n")
	}

	helpItems := []string{"↑/↓/j/k: navigate", "space: toggle", "s: select image", "enter: generate", "esc: cancel"}
	sb.WriteString(render.RenderHelpText(helpItems...))

	return render.RenderSimpleModal(width, height, sb.String())
}

func (m ReportGeneratorModal) renderStatsPreview() string {
	if m.PreviewStats == nil {
		return ""
	}

	var sb strings.Builder
	stats := m.PreviewStats

	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	projectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	activityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hoursStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	daysStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	focusedDaysStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("240"))

	sb.WriteString(sectionStyle.Render("Summary"))
	sb.WriteString("\n")
	sb.WriteString(activityStyle.Render(fmt.Sprintf("  Total Hours: %s", hoursStyle.Render(fmt.Sprintf("%.1f", stats.TotalHours)))))
	sb.WriteString("\n")
	sb.WriteString(activityStyle.Render(fmt.Sprintf("  Days Worked: %s", hoursStyle.Render(fmt.Sprintf("%d", stats.TotalDays)))))
	sb.WriteString("\n\n")

	if len(stats.ProjectActivityHours) > 0 {
		sb.WriteString(sectionStyle.Render("Hours by Project (Space to toggle)"))
		sb.WriteString("\n")

		type projectHour struct {
			name  string
			hours float64
		}
		projects := make([]projectHour, 0, len(stats.ProjectHours))
		for name, hours := range stats.ProjectHours {
			projects = append(projects, projectHour{name, hours})
		}
		sort.Slice(projects, func(i, j int) bool {
			return projects[i].hours > projects[j].hours
		})

		currentIndex := 0
		for _, p := range projects {
			projectDays := p.hours / 8.0
			sb.WriteString(projectStyle.Render(fmt.Sprintf("  %s", p.name)))
			sb.WriteString(hoursStyle.Render(fmt.Sprintf(" %.1fh", p.hours)))
			sb.WriteString(daysStyle.Render(fmt.Sprintf("(%.1fd)", projectDays)))
			sb.WriteString("\n")

			if activities, ok := stats.ProjectActivityHours[p.name]; ok {
				type activityHour struct {
					name  string
					hours float64
				}
				activityList := make([]activityHour, 0, len(activities))
				for name, hours := range activities {
					activityList = append(activityList, activityHour{name, hours})
				}
				sort.Slice(activityList, func(i, j int) bool {
					return activityList[i].hours > activityList[j].hours
				})

				for _, a := range activityList {
					isSelected := false
					if m.SelectedItems[p.name] != nil {
						isSelected = m.SelectedItems[p.name][a.name]
					}

					checkbox := "[ ]"
					if isSelected {
						checkbox = "[✓]"
					}

					isFocused := m.FocusedInput == 4 && m.FocusedItemIndex == currentIndex

					prefix := "    "
					if isFocused {
						prefix = "  ▶ "
					}

					activityDays := a.hours / 8.0

					line := fmt.Sprintf("%s%s %s", prefix, checkbox, a.name)

					if isFocused {
						sb.WriteString(focusedStyle.Render(line))
						sb.WriteString(focusedStyle.Render(fmt.Sprintf(" %.1fh", a.hours)))
						sb.WriteString(focusedDaysStyle.Render(fmt.Sprintf("(%.1fd)", activityDays)))
					} else {
						sb.WriteString(activityStyle.Render(line))
						sb.WriteString(hoursStyle.Render(fmt.Sprintf(" %.1fh", a.hours)))
						sb.WriteString(daysStyle.Render(fmt.Sprintf("(%.1fd)", activityDays)))
					}
					sb.WriteString("\n")

					currentIndex++
				}
			}
		}
	}

	return sb.String()
}

func (m ReportGeneratorModal) generateReport() tea.Cmd {
	return func() tea.Msg {
		switch m.SelectedReportType {
		case int(ReportTypeOdooCSV):
			filePath, err := generator.GenerateOdooCSVReport(m.ViewMonth, m.ViewYear)
			if err != nil {
				return ReportGenerationFailedMsg{Error: err}
			}
			return ReportGeneratedMsg{FilePath: filePath}
		case int(ReportTypeMailReport):
			fromCompany := strings.TrimSpace(m.FromCompanyInput.Value())
			toCompany := strings.TrimSpace(m.ToCompanyInput.Value())
			invoiceName := strings.TrimSpace(m.InvoiceNameInput.Value())
			filePath, err := generator.GenerateMailReport(m.ViewMonth, m.ViewYear, fromCompany, toCompany, invoiceName, m.SignatureImagePath, m.SelectedItems)
			if err != nil {
				return ReportGenerationFailedMsg{Error: err}
			}
			return ReportGeneratedMsg{FilePath: filePath}
		default:
			return ReportGeneratorModalClosedMsg{}
		}
	}
}


func dispatchReportGeneratorModalClosedMsg() tea.Cmd {
	return func() tea.Msg {
		return ReportGeneratorModalClosedMsg{}
	}
}
