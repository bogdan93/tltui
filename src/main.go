package main

import (
	"fmt"
	"os"
	"time-logger-tui/src/models"

	tea "github.com/charmbracelet/bubbletea"
)

func initModel() models.AppModel {
	return models.AppModel{
		Mode:            models.ModeViewWorkhourDetails,
		Projects:        models.NewProjectsModel(),
		WorkhourDetails: models.NewWorkhourDetailsModel(),

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
