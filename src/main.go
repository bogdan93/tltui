package main

import (
	"fmt"
	"os"
	"tltui/src/domain/repository"
	store "tltui/src/elm-store"
	"tltui/src/elm-store/calendar"
	"tltui/src/elm-store/projects"
	"tltui/src/elm-store/workhour_details"

	tea "github.com/charmbracelet/bubbletea"
)

func initModel() store.AppModel {
	return store.AppModel{
		Mode:            store.ModeViewCalendar,
		Calendar:        calendar.NewCalendarModel(),
		Projects:        projects.NewProjectsModel(),
		WorkhourDetails: workhour_details.NewWorkhourDetailsModel(),
	}
}

func main() {
	if err := repository.InitDB(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer repository.CloseDB()

	if err := repository.SeedProjects(); err != nil {
		fmt.Printf("Failed to seed projects: %v\n", err)
		os.Exit(1)
	}

	if err := repository.SeedWorkhourDetails(); err != nil {
		fmt.Printf("Failed to seed workhour details: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
