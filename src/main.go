package main

import (
	"fmt"
	"os"
	"time"
	"time-logger-tui/src/models"

	tea "github.com/charmbracelet/bubbletea"
)

func initModel() models.AppModel {
	// Create sample workhours with dates
	now := time.Now()
	workhours := []models.Workhour{
		{Date: now.AddDate(0, 0, -2), DetailsID: 1, ProjectID: 1, Hours: 8.0},
		{Date: now.AddDate(0, 0, -2), DetailsID: 3, ProjectID: 2, Hours: 1.0},
		{Date: now.AddDate(0, 0, -1), DetailsID: 1, ProjectID: 1, Hours: 7.0},
		{Date: now.AddDate(0, 0, -1), DetailsID: 2, ProjectID: 0, Hours: 1.0},
		{Date: now, DetailsID: 1, ProjectID: 1, Hours: 5.0},                 
		{Date: now, DetailsID: 1, ProjectID: 3, Hours: 2.0},                
		{Date: now.AddDate(0, 0, 1), DetailsID: 2, ProjectID: 1, Hours: 8.0},
		{Date: now.AddDate(0, 0, 2), DetailsID: 3, ProjectID: 0, Hours: 8.0},
	}

	workhourDetails := models.FetchAllWorkhourDetails()

	calendar := models.NewCalendarModel()
	calendar.Workhours = workhours
	calendar.WorkhourDetails = workhourDetails

	return models.AppModel{
		Mode:            models.ModeViewCalendar,
		Calendar:        calendar,
		Projects:        models.NewProjectsModel(),
		WorkhourDetails: models.NewWorkhourDetailsModel(),

		Workhours: workhours,
	}
}

func main() {
	p := tea.NewProgram(initModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
