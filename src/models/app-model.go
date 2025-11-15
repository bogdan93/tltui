package models

import (
	"time-logger-tui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkhourDetails struct {
	ID        int
	Name      string
	ShortName string
	IsWork    bool
}

type Workhour struct {
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

type AppModel struct {
	Mode AppMode

	// Projects List
	Projects ProjectsModel

	// Data
	WorkhoursDetails []WorkhourDetails
	Workhours        []Workhour
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	if m.Mode == ModeViewProjects {
		return m.Projects.Update(msg)
	}

	return m, tea.Batch(cmds...)
}

var globalStyle = lipgloss.NewStyle().Padding(1)

func (m AppModel) View() string {
	switch m.Mode {
	case ModeViewProjects:
		return m.Projects.View()

	case ModeViewWorkhours:
		return render.RenderPageLayout("Workhours", "Workhours view is under construction. Press 'q' to quit.")

	case ModeViewWorkhourDetails:
		return render.RenderPageLayout("Workhour Details", "Workhour Details view is under construction. Press 'q' to quit.")

	case ModeViewCalendar:
		return render.RenderPageLayout("Calendar", "Calendar view is under construction. Press 'q' to quit.")
	}

	return ""
}
