package models

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"tltui/src/domain/repository"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ReportType int

const (
	ReportTypeOdooCSV ReportType = iota
)

type ReportGeneratorModal struct {
	SelectedReportType int
	ReportTypes        []string
	Generating         bool
	ErrorMessage       string
	ViewMonth          int // Month to generate report for
	ViewYear           int // Year to generate report for
}

type ReportGeneratorModalClosedMsg struct{}
type ReportGeneratedMsg struct {
	FilePath string
}
type ReportGenerationFailedMsg struct {
	Error error
}

func NewReportGeneratorModal(viewMonth, viewYear int) *ReportGeneratorModal {
	return &ReportGeneratorModal{
		SelectedReportType: 0,
		ReportTypes:        []string{"Odoo CSV"},
		Generating:         false,
		ViewMonth:          viewMonth,
		ViewYear:           viewYear,
	}
}

func (m ReportGeneratorModal) Init() tea.Cmd {
	return nil
}

func (m ReportGeneratorModal) Update(msg tea.Msg) (ReportGeneratorModal, tea.Cmd) {
	if m.Generating {
		// Don't process input while generating
		return m, nil
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
			// Generate report based on selected type
			m.Generating = true
			return m, m.generateReport()
		}
	}

	return m, nil
}

func (m ReportGeneratorModal) View(width, height int) string {
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

		for i, reportType := range m.ReportTypes {
			prefix := "  "
			style := lipgloss.NewStyle()

			if i == m.SelectedReportType {
				prefix = "▶ "
				style = style.Bold(true).Foreground(lipgloss.Color("39"))
			}

			sb.WriteString(prefix + style.Render(reportType) + "\n")
		}

		sb.WriteString("\n")
		helpItems := []string{"↑/↓: select", "enter: generate", "esc/q: cancel"}
		sb.WriteString(render.RenderHelpText(helpItems...))
	}

	return render.RenderSimpleModal(width, height, sb.String())
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
		default:
			return ReportGeneratorModalClosedMsg{}
		}
	}
}

func generateOdooCSVReport(viewMonth, viewYear int) (string, error) {
	// Calculate date range for the specified month
	startDate := time.Date(viewYear, time.Month(viewMonth), 1, 0, 0, 0, 0, time.Local)
	// Last day of month: first day of next month minus one day
	endDate := time.Date(viewYear, time.Month(viewMonth+1), 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)

	// Get workhours for the specified month
	workhours, err := repository.GetWorkhoursByDateRange(startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get workhours: %w", err)
	}

	// Get workhour details and projects for lookup
	workhourDetails, err := repository.GetAllWorkhourDetailsFromDB()
	if err != nil {
		return "", fmt.Errorf("failed to get workhour details: %w", err)
	}

	projects, err := repository.GetAllProjectsFromDB()
	if err != nil {
		return "", fmt.Errorf("failed to get projects: %w", err)
	}

	// Create maps for quick lookup
	detailsMap := make(map[int]WorkhourDetails)
	for _, wd := range workhourDetails {
		detailsMap[wd.ID] = wd
	}

	projectsMap := make(map[int]Project)
	for _, p := range projects {
		projectsMap[p.ID] = p
	}

	// Create temporary file
	tmpDir := os.TempDir()
	monthName := time.Month(viewMonth).String()
	fileName := fmt.Sprintf("odoo_timesheet_%s_%d.csv", monthName, viewYear)
	filePath := filepath.Join(tmpDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	writer := csv.NewWriter(file)

	// Write CSV header
	header := []string{"date", "account_id/id", "journal_id/id", "name", "unit_amount"}
	if err := writer.Write(header); err != nil {
		file.Close()
		return "", fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows
	for _, wh := range workhours {
		details, ok := detailsMap[wh.DetailsID]
		if !ok {
			continue // Skip if details not found
		}

		project, ok := projectsMap[wh.ProjectID]
		if !ok {
			continue // Skip if project not found
		}

		// Format: date,account_id/id,journal_id/id,name,unit_amount
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

	// Flush and close file BEFORE opening save dialog
	writer.Flush()
	if err := writer.Error(); err != nil {
		file.Close()
		return "", fmt.Errorf("failed to flush csv: %w", err)
	}
	file.Close()

	// Open file save dialog using OS-specific command
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

	// Platform-specific file save dialog
	switch {
	case commandExists("zenity"):
		// Linux - GTK dialog
		cmd = exec.Command("zenity", "--file-selection", "--save", "--confirm-overwrite",
			"--filename="+defaultPath,
			"--title=Save Odoo CSV Report")
	case commandExists("kdialog"):
		// Linux - KDE dialog
		cmd = exec.Command("kdialog", "--getsavefilename", defaultPath, "*.csv")
	case commandExists("osascript"):
		// macOS
		script := fmt.Sprintf(`
			set defaultPath to POSIX file "%s"
			set saveFile to choose file name with prompt "Save Odoo CSV Report" default name "%s" default location (path to home folder)
			return POSIX path of saveFile
		`, defaultPath, defaultFileName)
		cmd = exec.Command("osascript", "-e", script)
	default:
		// Fallback: just use the temp file
		return sourceFile, nil
	}

	output, err := cmd.Output()
	if err != nil {
		// User cancelled or error - keep temp file
		return sourceFile, nil
	}

	targetPath := strings.TrimSpace(string(output))
	if targetPath == "" {
		return sourceFile, nil
	}

	// Copy file to selected location
	if err := copyFile(sourceFile, targetPath); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// Remove temp file
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

func dispatchReportGeneratorModalClosedMsg() tea.Cmd {
	return func() tea.Msg {
		return ReportGeneratorModalClosedMsg{}
	}
}
