package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type project struct {
	ID     int
	Name   string
	OdooID int
}

type workhour_details struct {
	ID        int
	Name      string
	ShortName string
	IsWork    bool
}

type workhour struct {
	DetailsID int
	ProjectID int
	Hours     float64
}

type AppMode int

const (
	ModeViewCalendar AppMode = iota
	ModeViewProjects
	ModeViewWorkhours
	ModeViewWorkhourDetails
)

type model struct {
	// UI
	Width  int
	Height int

	// State
	Mode AppMode

	// Data
	Projects         []project
	WorkhoursDetails []workhour_details
	Workhours        []workhour
}

func initModel() model {
	return model{
		Mode: ModeViewProjects,

		Projects: []project{
			{ID: 1, Name: "Project A", OdooID: 101},
			{ID: 2, Name: "Project B", OdooID: 102},
		},
		WorkhoursDetails: []workhour_details{
			{ID: 1, Name: "Development", ShortName: "Dev", IsWork: true},
			{ID: 2, Name: "Meeting", ShortName: "Meet", IsWork: true},
		},
		Workhours: []workhour{
			{DetailsID: 1, ProjectID: 1, Hours: 5.0},
			{DetailsID: 2, ProjectID: 1, Hours: 2.0},
			{DetailsID: 1, ProjectID: 2, Hours: 3.5},
		},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

var globalStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)

func (m model) View() string {
	var title string
	switch m.Mode {
	case ModeViewProjects:
		title = "Projects"
	case ModeViewWorkhours:
		title = "Workhours"
	case ModeViewWorkhourDetails:
		title = "Workhour Details"
	case ModeViewCalendar:
		title = "Calendar"
	}

	titleString := RenderPageTitle(title)

	content := globalStyle.Render(titleString + "\nPress 'q' to quit.")

	return content
}

func main() {
	p := tea.NewProgram(initModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
