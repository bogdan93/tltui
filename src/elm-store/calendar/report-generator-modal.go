package calendar

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"tltui/src/domain"
	"tltui/src/domain/repository"
	"tltui/src/render"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
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
	PreviewStats       *WorkhourStats             // Cached stats for preview display
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
type SignatureImageSelectedMsg struct {
	ImagePath string
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

func (m ReportGeneratorModal) calculatePreviewStats() *WorkhourStats {
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

	stats := calculateWorkhourStats(workhours, detailsMap, projectsMap)
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
				return m, openImageFileDialog()
			}
		}

	case SignatureImageSelectedMsg:
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
			filePath, err := generateOdooCSVReport(m.ViewMonth, m.ViewYear)
			if err != nil {
				return ReportGenerationFailedMsg{Error: err}
			}
			return ReportGeneratedMsg{FilePath: filePath}
		case int(ReportTypeMailReport):
			fromCompany := strings.TrimSpace(m.FromCompanyInput.Value())
			toCompany := strings.TrimSpace(m.ToCompanyInput.Value())
			invoiceName := strings.TrimSpace(m.InvoiceNameInput.Value())
			filePath, err := generateMailReport(m.ViewMonth, m.ViewYear, fromCompany, toCompany, invoiceName, m.SignatureImagePath, m.SelectedItems)
			if err != nil {
				return ReportGenerationFailedMsg{Error: err}
			}
			return ReportGeneratedMsg{FilePath: filePath}
		default:
			return ReportGeneratorModalClosedMsg{}
		}
	}
}

func generateOdooCSVReport(viewMonth, viewYear int) (string, error) {
	startDate := time.Date(viewYear, time.Month(viewMonth), 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(viewYear, time.Month(viewMonth+1), 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)

	workhours, err := repository.GetWorkhoursByDateRange(startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get workhours: %w", err)
	}

	workhourDetails, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		return "", fmt.Errorf("failed to get workhour details: %w", err)
	}

	projects, err := repository.GetAllProjectsFromDB()
	if err != nil {
		return "", fmt.Errorf("failed to get projects: %w", err)
	}

	detailsMap := make(map[int]domain.WorkhourDetails)
	for _, wd := range workhourDetails {
		detailsMap[wd.ID] = wd
	}

	projectsMap := make(map[int]domain.Project)
	for _, p := range projects {
		projectsMap[p.ID] = p
	}

	tmpDir := os.TempDir()
	monthName := time.Month(viewMonth).String()
	fileName := fmt.Sprintf("odoo_timesheet_%s_%d.csv", monthName, viewYear)
	filePath := filepath.Join(tmpDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	writer := csv.NewWriter(file)

	header := []string{"date", "account_id/id", "journal_id/id", "name", "unit_amount"}
	if err := writer.Write(header); err != nil {
		file.Close()
		return "", fmt.Errorf("failed to write header: %w", err)
	}

	for _, wh := range workhours {
		details, ok := detailsMap[wh.DetailsID]
		if !ok {
			continue
		}

		project, ok := projectsMap[wh.ProjectID]
		if !ok {
			continue
		}

		row := []string{
			repository.DateToString(wh.Date),
			fmt.Sprintf("__export__.account_analytic_account_%d", project.OdooID),
			"hr_timesheet.analytic_journal",
			details.Name,
			fmt.Sprintf("%.1f", wh.Hours),
		}

		if err := writer.Write(row); err != nil {
			file.Close()
			return "", fmt.Errorf("failed to write row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		file.Close()
		return "", fmt.Errorf("failed to flush csv: %w", err)
	}
	file.Close()

	return openFileSaveDialog(filePath)
}

func openFileSaveDialog(sourceFile string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	defaultFileName := filepath.Base(sourceFile)
	defaultPath := filepath.Join(homeDir, defaultFileName)

	var cmd *exec.Cmd

	switch {
	case commandExists("zenity"):
		cmd = exec.Command("zenity", "--file-selection", "--save", "--confirm-overwrite",
			"--filename="+defaultPath,
			"--title=Save Odoo CSV Report")
	case commandExists("kdialog"):
		cmd = exec.Command("kdialog", "--getsavefilename", defaultPath, "*.csv")
	case commandExists("osascript"):
		script := fmt.Sprintf(`
			set defaultPath to POSIX file "%s"
			set saveFile to choose file name with prompt "Save Odoo CSV Report" default name "%s" default location (path to home folder)
			return POSIX path of saveFile
		`, defaultPath, defaultFileName)
		cmd = exec.Command("osascript", "-e", script)
	default:
		return sourceFile, nil
	}

	output, err := cmd.Output()
	if err != nil {
		return sourceFile, nil
	}

	targetPath := strings.TrimSpace(string(output))
	if targetPath == "" {
		return sourceFile, nil
	}

	if err := copyFile(sourceFile, targetPath); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	os.Remove(sourceFile)

	return targetPath, nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func openImageFileDialog() tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "/"
		}

		var cmd *exec.Cmd

		switch {
		case commandExists("zenity"):
			cmd = exec.Command("zenity", "--file-selection",
				"--title=Select Signature Image",
				"--file-filter=Images | *.png *.jpg *.jpeg",
				"--filename="+homeDir+"/")
		case commandExists("kdialog"):
			cmd = exec.Command("kdialog", "--getopenfilename", homeDir, "*.png *.jpg *.jpeg")
		case commandExists("osascript"):
			script := `
				set imageFile to choose file with prompt "Select Signature Image" of type {"public.image"}
				return POSIX path of imageFile
			`
			cmd = exec.Command("osascript", "-e", script)
		default:
			return SignatureImageSelectedMsg{ImagePath: ""}
		}

		output, err := cmd.Output()
		if err != nil {
			return SignatureImageSelectedMsg{ImagePath: ""}
		}

		imagePath := strings.TrimSpace(string(output))
		return SignatureImageSelectedMsg{ImagePath: imagePath}
	}
}

func dispatchReportGeneratorModalClosedMsg() tea.Cmd {
	return func() tea.Msg {
		return ReportGeneratorModalClosedMsg{}
	}
}

type WorkhourStats struct {
	TotalHours           float64
	TotalDays            int
	AveragePerDay        float64
	ProjectHours         map[string]float64            // project name -> hours
	ActivityHours        map[string]float64            // activity name -> hours
	ProjectActivityHours map[string]map[string]float64 // project name -> activity name -> hours
	DailyBreakdown       map[string][]WorkhourEntry    // date -> list of entries
}

type WorkhourEntry struct {
	ProjectName  string
	ActivityName string
	Hours        float64
}

func calculateWorkhourStats(
	workhours []domain.Workhour,
	detailsMap map[int]domain.WorkhourDetails,
	projectsMap map[int]domain.Project,
) WorkhourStats {
	stats := WorkhourStats{
		ProjectHours:         make(map[string]float64),
		ActivityHours:        make(map[string]float64),
		ProjectActivityHours: make(map[string]map[string]float64),
		DailyBreakdown:       make(map[string][]WorkhourEntry),
	}

	daysWorked := make(map[string]bool)

	for _, wh := range workhours {
		stats.TotalHours += wh.Hours

		dateStr := repository.DateToString(wh.Date)
		daysWorked[dateStr] = true

		var projectName, activityName string

		if project, ok := projectsMap[wh.ProjectID]; ok {
			projectName = project.Name
			stats.ProjectHours[projectName] += wh.Hours
		}

		if details, ok := detailsMap[wh.DetailsID]; ok {
			activityName = details.Name
			stats.ActivityHours[activityName] += wh.Hours
		}

		if projectName != "" && activityName != "" {
			if stats.ProjectActivityHours[projectName] == nil {
				stats.ProjectActivityHours[projectName] = make(map[string]float64)
			}
			stats.ProjectActivityHours[projectName][activityName] += wh.Hours
		}

		entry := WorkhourEntry{
			Hours:        wh.Hours,
			ProjectName:  projectName,
			ActivityName: activityName,
		}
		stats.DailyBreakdown[dateStr] = append(stats.DailyBreakdown[dateStr], entry)
	}

	stats.TotalDays = len(daysWorked)
	if stats.TotalDays > 0 {
		stats.AveragePerDay = stats.TotalHours / float64(stats.TotalDays)
	}

	return stats
}

func formatMailReport(
	viewMonth, viewYear int,
	fromCompany, toCompany string,
	stats WorkhourStats,
) string {
	monthName := time.Month(viewMonth).String()
	currentTime := time.Now().Format("January 2, 2006 at 3:04 PM")

	var sb strings.Builder

	sb.WriteString(strings.Repeat("=", 70))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("WORKHOUR REPORT - %s %d\n", monthName, viewYear))
	sb.WriteString(strings.Repeat("=", 70))
	sb.WriteString("\n\n")

	if fromCompany != "" {
		sb.WriteString(fmt.Sprintf("From: %s\n", fromCompany))
	}
	if toCompany != "" {
		sb.WriteString(fmt.Sprintf("To: %s\n", toCompany))
	}
	sb.WriteString(fmt.Sprintf("Date: %s\n", currentTime))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("-", 70))
	sb.WriteString("\n\n")

	sb.WriteString("SUMMARY\n")
	sb.WriteString(strings.Repeat("-", 70))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Total Work Hours:     %.1f hours\n", stats.TotalHours))
	sb.WriteString(fmt.Sprintf("Total Days Worked:    %d days\n", stats.TotalDays))
	sb.WriteString(fmt.Sprintf("Average Hours/Day:    %.1f hours\n", stats.AveragePerDay))
	sb.WriteString("\n\n")

	if len(stats.ProjectHours) > 0 {
		sb.WriteString("BREAKDOWN BY PROJECT\n")
		sb.WriteString(strings.Repeat("-", 70))
		sb.WriteString("\n")

		projectNames := make([]string, 0, len(stats.ProjectHours))
		for name := range stats.ProjectHours {
			projectNames = append(projectNames, name)
		}
		sort.Strings(projectNames)

		for _, projectName := range projectNames {
			hours := stats.ProjectHours[projectName]
			percentage := (hours / stats.TotalHours) * 100
			sb.WriteString(fmt.Sprintf("%-40s %8.1f hours (%5.1f%%)\n", projectName, hours, percentage))
		}
		sb.WriteString("\n\n")
	}

	if len(stats.ActivityHours) > 0 {
		sb.WriteString("BREAKDOWN BY ACTIVITY TYPE\n")
		sb.WriteString(strings.Repeat("-", 70))
		sb.WriteString("\n")

		activityNames := make([]string, 0, len(stats.ActivityHours))
		for name := range stats.ActivityHours {
			activityNames = append(activityNames, name)
		}
		sort.Strings(activityNames)

		for _, activityName := range activityNames {
			hours := stats.ActivityHours[activityName]
			percentage := (hours / stats.TotalHours) * 100
			sb.WriteString(fmt.Sprintf("%-40s %8.1f hours (%5.1f%%)\n", activityName, hours, percentage))
		}
		sb.WriteString("\n\n")
	}

	sb.WriteString("DAILY BREAKDOWN\n")
	sb.WriteString(strings.Repeat("-", 70))
	sb.WriteString("\n")

	dates := make([]string, 0, len(stats.DailyBreakdown))
	for date := range stats.DailyBreakdown {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	for _, dateStr := range dates {
		entries := stats.DailyBreakdown[dateStr]

		parsedDate, _ := time.Parse("2006-01-02", dateStr)
		dayOfWeek := parsedDate.Format("Monday")

		dailyTotal := 0.0
		for _, entry := range entries {
			dailyTotal += entry.Hours
		}

		sb.WriteString(fmt.Sprintf("%s - %s: %.1f hours\n", dateStr, dayOfWeek, dailyTotal))

		for _, entry := range entries {
			sb.WriteString(fmt.Sprintf("  - %s / %s: %.1f hours\n",
				entry.ProjectName, entry.ActivityName, entry.Hours))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("=", 70))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Generated on %s\n", currentTime))
	sb.WriteString(strings.Repeat("=", 70))
	sb.WriteString("\n")

	return sb.String()
}

func generatePDFReport(filePath string, viewMonth, viewYear int, fromCompany, toCompany, invoiceName, signatureImagePath string, stats WorkhourStats) error {
	cfg := config.NewBuilder().
		WithPageNumber().
		Build()

	m := maroto.New(cfg)
	monthName := time.Month(viewMonth).String()

	// Add title and company info as regular rows (only on first page)
	m.AddRow(5)
	m.AddRow(10,
		text.NewCol(12, "Raport de activitate", props.Text{
			Top:   3,
			Size:  16,
			Style: fontstyle.Bold,
			Align: align.Center,
		}),
	)
	m.AddRow(15)
	m.AddRow(5,
		text.NewCol(4, "Firma prestatoare:", props.Text{
			Size:  10,
			Top:   1,
			Style: fontstyle.Bold,
		}),
		text.NewCol(8, fromCompany, props.Text{
			Size: 10,
			Top:  1,
		}),
	)
	m.AddRow(5,
		text.NewCol(4, "Catre:", props.Text{
			Size:  10,
			Top:   1,
			Style: fontstyle.Bold,
		}),
		text.NewCol(8, toCompany, props.Text{
			Size: 10,
			Top:  1,
		}),
	)
	m.AddRow(5,
		text.NewCol(4, "Referitor la factura numarul:", props.Text{
			Size:  10,
			Top:   1,
			Style: fontstyle.Bold,
		}),
		text.NewCol(8, fmt.Sprintf("%s - %s %d", invoiceName, monthName, viewYear), props.Text{
			Size: 10,
			Top:  1,
		}),
	)

	m.AddRow(30)

	m.AddRow(8,
		text.NewCol(12, "Raport de ore lucrate", props.Text{
			Top:   2,
			Size:  11,
			Style: fontstyle.Bold,
			Align: align.Center,
		}),
	)

	tableHeaders := []string{"Data", "Proiect", "Descriere", "Ore lucrate"}
	var tableRows [][]string

	dates := make([]string, 0, len(stats.DailyBreakdown))
	for date := range stats.DailyBreakdown {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	totalHours := 0.0
	for _, dateStr := range dates {
		entries := stats.DailyBreakdown[dateStr]

		for _, entry := range entries {
			tableRows = append(tableRows, []string{
				dateStr,
				entry.ProjectName,
				entry.ActivityName,
				fmt.Sprintf("%.0f", entry.Hours),
			})
			totalHours += entry.Hours
		}
	}

	tableRows = append(tableRows, []string{
		"",
		"",
		"",
		fmt.Sprintf("%.0f", totalHours),
	})

	darkBlue := &props.Color{Red: 54, Green: 69, Blue: 92}
	lightBlue := &props.Color{Red: 207, Green: 226, Blue: 243}
	white := &props.Color{Red: 255, Green: 255, Blue: 255}
	black := &props.Color{Red: 0, Green: 0, Blue: 0}

	headerCellStyle := &props.Cell{
		BackgroundColor: darkBlue,
		BorderType:      border.Full,
		BorderColor:     black,
		BorderThickness: 0.5,
	}

	// Register table header to repeat on every page
	err := m.RegisterHeader(
		row.New(7).Add(
			col.New(3).Add(text.New(tableHeaders[0], props.Text{
				Top:   1.5,
				Size:  9,
				Style: fontstyle.Bold,
				Align: align.Center,
				Color: white,
			})).WithStyle(headerCellStyle),
			col.New(3).Add(text.New(tableHeaders[1], props.Text{
				Top:   1.5,
				Size:  9,
				Style: fontstyle.Bold,
				Align: align.Center,
				Color: white,
			})).WithStyle(headerCellStyle),
			col.New(3).Add(text.New(tableHeaders[2], props.Text{
				Top:   1.5,
				Size:  9,
				Style: fontstyle.Bold,
				Align: align.Center,
				Color: white,
			})).WithStyle(headerCellStyle),
			col.New(3).Add(text.New(tableHeaders[3], props.Text{
				Top:   1.5,
				Size:  9,
				Style: fontstyle.Bold,
				Align: align.Center,
				Color: white,
			})).WithStyle(headerCellStyle),
		),
	)

	if err != nil {
		return fmt.Errorf("failed to register table header: %w", err)
	}

	for i, rowData := range tableRows {
		backgroundColor := white
		isLastRow := i == len(tableRows)-1

		if isLastRow {
			backgroundColor = lightBlue
		} else if i%2 == 1 {
			backgroundColor = lightBlue
		}

		cellStyle := &props.Cell{
			BackgroundColor: backgroundColor,
			BorderType:      border.Full,
			BorderColor:     black,
			BorderThickness: 0.5,
		}

		m.AddRow(6,
			col.New(3).Add(text.New(rowData[0], props.Text{
				Top:   1,
				Size:  8,
				Align: align.Center,
			})).WithStyle(cellStyle),
			col.New(3).Add(text.New(rowData[1], props.Text{
				Top:   1,
				Size:  8,
				Align: align.Center,
			})).WithStyle(cellStyle),
			col.New(3).Add(text.New(rowData[2], props.Text{
				Top:   1,
				Size:  8,
				Align: align.Center,
			})).WithStyle(cellStyle),
			col.New(3).Add(text.New(rowData[3], props.Text{
				Top:   1,
				Size:  8,
				Align: align.Center,
			})).WithStyle(cellStyle),
		)
	}

	m.AddRow(10,
		text.NewCol(6, "Semnatura Prestator,", props.Text{
			Size:  10,
			Top:   4,
			Align: align.Left,
			Style: fontstyle.Bold,
		}),
		text.NewCol(6, "Semnatura Beneficiar,", props.Text{
			Size:  10,
			Top:   4,
			Align: align.Right,
			Style: fontstyle.Bold,
		}),
	)

	if signatureImagePath != "" {
		if _, err := os.Stat(signatureImagePath); err == nil {
			m.AddRow(30,
				image.NewFromFileCol(6, signatureImagePath, props.Rect{
					Center:  false,
					Left:    0,
					Top:     0,
					Percent: 50,
				}),
			)
		}
	}

	document, err := m.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate PDF document: %w", err)
	}

	err = document.Save(filePath)
	if err != nil {
		return fmt.Errorf("failed to save PDF file: %w", err)
	}

	return nil
}

func generateMailReport(viewMonth, viewYear int, fromCompany, toCompany, invoiceName, signatureImagePath string, selectedItems map[string]map[string]bool) (string, error) {
	startDate := time.Date(viewYear, time.Month(viewMonth), 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(viewYear, time.Month(viewMonth+1), 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)

	workhours, err := repository.GetWorkhoursByDateRange(startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to fetch workhours: %w", err)
	}

	workhourDetails, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		return "", fmt.Errorf("failed to fetch workhour details: %w", err)
	}

	projects, err := repository.GetAllProjectsFromDB()
	if err != nil {
		return "", fmt.Errorf("failed to fetch projects: %w", err)
	}

	detailsMap := make(map[int]domain.WorkhourDetails)
	for _, wd := range workhourDetails {
		detailsMap[wd.ID] = wd
	}

	projectsMap := make(map[int]domain.Project)
	for _, p := range projects {
		projectsMap[p.ID] = p
	}

	filteredWorkhours := make([]domain.Workhour, 0)
	for _, wh := range workhours {
		project, projectOk := projectsMap[wh.ProjectID]
		details, detailsOk := detailsMap[wh.DetailsID]

		if projectOk && detailsOk {
			if selectedItems[project.Name] != nil && selectedItems[project.Name][details.Name] {
				filteredWorkhours = append(filteredWorkhours, wh)
			}
		}
	}

	stats := calculateWorkhourStats(filteredWorkhours, detailsMap, projectsMap)

	tmpDir := os.TempDir()
	monthName := time.Month(viewMonth).String()
	fileName := fmt.Sprintf("raport_activitate_%s_%d.pdf", strings.ToLower(monthName), viewYear)
	filePath := filepath.Join(tmpDir, fileName)

	err = generatePDFReport(filePath, viewMonth, viewYear, fromCompany, toCompany, invoiceName, signatureImagePath, stats)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	savePath, err := openMailReportSaveDialog(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open save dialog: %w", err)
	}

	if savePath != filePath {
		err = copyFile(filePath, savePath)
		if err != nil {
			return "", fmt.Errorf("failed to save report: %w", err)
		}
		os.Remove(filePath)
	}

	return savePath, nil
}

func openMailReportSaveDialog(sourceFile string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	defaultFileName := filepath.Base(sourceFile)
	defaultPath := filepath.Join(homeDir, defaultFileName)

	var cmd *exec.Cmd

	switch {
	case commandExists("zenity"):
		cmd = exec.Command("zenity", "--file-selection", "--save", "--confirm-overwrite",
			"--filename="+defaultPath,
			"--title=Save Mail Report")
	case commandExists("kdialog"):
		cmd = exec.Command("kdialog", "--getsavefilename", defaultPath, "*.pdf")
	case commandExists("osascript"):
		script := fmt.Sprintf(`
			set defaultPath to POSIX file "%s"
			set saveFile to choose file name with prompt "Save Mail Report" default name "%s" default location (path to home folder)
			return POSIX path of saveFile
		`, defaultPath, defaultFileName)
		cmd = exec.Command("osascript", "-e", script)
	default:
		return sourceFile, nil
	}

	output, err := cmd.Output()
	if err != nil {
		return sourceFile, nil
	}

	targetPath := strings.TrimSpace(string(output))
	if targetPath == "" {
		return sourceFile, nil
	}

	return targetPath, nil
}
