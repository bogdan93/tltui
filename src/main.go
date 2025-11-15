package main

import (
	"fmt"
	"os"
	"time"
	"tltui/src/models"

	tea "github.com/charmbracelet/bubbletea"
)

func initModel() models.AppModel {
	return models.AppModel{
		Mode:            models.ModeViewCalendar,
		Calendar:        models.NewCalendarModel(),
		Projects:        models.NewProjectsModel(),
		WorkhourDetails: models.NewWorkhourDetailsModel(),
	}
}

func seedSampleData() error {
	// Create sample workhours with dates
	now := time.Now()
	sampleWorkhours := []models.Workhour{
		{},
		{Date: now.AddDate(0, 0, -2), DetailsID: 3, ProjectID: 2, Hours: 1.0},
		{Date: now.AddDate(0, 0, -1), DetailsID: 1, ProjectID: 1, Hours: 7.0},
		{Date: now.AddDate(0, 0, -1), DetailsID: 2, ProjectID: 1, Hours: 1.0},
		{Date: now, DetailsID: 1, ProjectID: 1, Hours: 5.0},
		{Date: now, DetailsID: 1, ProjectID: 3, Hours: 2.0},
		{Date: now.AddDate(0, 0, 1), DetailsID: 2, ProjectID: 1, Hours: 8.0},
		{Date: now.AddDate(0, 0, 2), DetailsID: 3, ProjectID: 2, Hours: 8.0},
	}

	// Only seed if there are no workhours in the database
	existing, err := models.GetAllWorkhours()
	if err != nil {
		return fmt.Errorf("failed to check existing workhours: %w", err)
	}

	if len(existing) == 0 {
		for _, wh := range sampleWorkhours {
			if _, err := models.CreateWorkhour(wh); err != nil {
				return fmt.Errorf("failed to seed workhour: %w", err)
			}
		}
	}

	return nil
}

func main() {
	// Initialize database
	if err := models.InitDB(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer models.CloseDB()

	// Seed initial data (projects and workhour details)
	if err := models.SeedProjects(); err != nil {
		fmt.Printf("Failed to seed projects: %v\n", err)
		os.Exit(1)
	}

	if err := models.SeedWorkhourDetails(); err != nil {
		fmt.Printf("Failed to seed workhour details: %v\n", err)
		os.Exit(1)
	}

	// Run the application
	p := tea.NewProgram(initModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
