package main

import (
	"fmt"
	"os"
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
