package main

import (
	"fmt"
	"os"
	"time-logger-tui/src/models"

	tea "github.com/charmbracelet/bubbletea"
)

func initModel() models.AppModel {
	return models.AppModel{
		Mode:          models.ModeViewProjects,
		ProjectsTable: models.ProjectsModelInit(),
		Projects:      models.GetProjects(),

		WorkhoursDetails: []models.WorkhourDetails{
			{ID: 1, Name: "Development", ShortName: "Dev", IsWork: true},
			{ID: 2, Name: "Meeting", ShortName: "Meet", IsWork: true},
		},
		Workhours: []models.Workhour{
			{DetailsID: 1, ProjectID: 1, Hours: 5.0},
			{DetailsID: 2, ProjectID: 1, Hours: 2.0},
			{DetailsID: 1, ProjectID: 2, Hours: 3.5},
		},
	}
}

func main() {
	p := tea.NewProgram(initModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
