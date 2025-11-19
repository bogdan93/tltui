package report_generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"tltui/src/domain"
	"tltui/src/domain/repository"

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

// GenerateMailReport generates a PDF activity report for the given month
func GenerateMailReport(viewMonth, viewYear int, fromCompany, toCompany, invoiceName, signatureImagePath string, selectedItems map[string]map[string]bool) (string, error) {
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

	stats := CalculateWorkhourStats(filteredWorkhours, detailsMap, projectsMap)

	tmpDir := os.TempDir()
	monthName := time.Month(viewMonth).String()
	fileName := fmt.Sprintf("raport_activitate_%s_%d.pdf", strings.ToLower(monthName), viewYear)
	filePath := filepath.Join(tmpDir, fileName)

	err = generatePDFReport(filePath, viewMonth, viewYear, fromCompany, toCompany, invoiceName, signatureImagePath, stats)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	savePath, err := OpenPDFSaveDialog(filePath)
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

// generatePDFReport creates the actual PDF file with formatted content
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
				fmt.Sprintf("%g", entry.Hours),
			})
			totalHours += entry.Hours
		}
	}

	tableRows = append(tableRows, []string{
		"",
		"",
		"",
		fmt.Sprintf("%g", totalHours),
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
