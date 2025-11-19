package report_generator

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"tltui/src/domain"
	"tltui/src/domain/repository"
)

// GenerateOdooCSVReport generates an Odoo-compatible CSV timesheet export
func GenerateOdooCSVReport(viewMonth, viewYear int) (string, error) {
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
			fmt.Sprintf("%g", wh.Hours),
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

	return OpenCSVSaveDialog(filePath)
}
