package models

import (
	"time-logger-tui/src/render"

	tea "github.com/charmbracelet/bubbletea"
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
	ModeViewWorkhourDetails
)

type AppModel struct {
	Mode AppMode

	// Views
	Calendar        CalendarModel
	Projects        ProjectsModel
	WorkhourDetails WorkhourDetailsModel

	// Data
	Workhours []Workhour
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		// Forward window size to all models
		var cmd1, cmd2, cmd3 tea.Cmd
		var updatedModel tea.Model

		updatedModel, cmd1 = m.Calendar.Update(msg)
		m.Calendar = updatedModel.(CalendarModel)

		updatedModel, cmd2 = m.Projects.Update(msg)
		m.Projects = updatedModel.(ProjectsModel)

		updatedModel, cmd3 = m.WorkhourDetails.Update(msg)
		m.WorkhourDetails = updatedModel.(WorkhourDetailsModel)

		return m, tea.Batch(cmd1, cmd2, cmd3)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "1", "2", "3":
			// Don't handle tab switching if any modal is open
			if m.Projects.ProjectEditModal != nil ||
				m.Projects.ProjectCreateModal != nil ||
				m.WorkhourDetails.WorkhourDetailsEditModal != nil ||
				m.WorkhourDetails.WorkhourDetailsCreateModal != nil ||
				m.WorkhourDetails.WorkhourDetailsDeleteModal != nil {
				// Modal is open, let the number key pass through to the modal
				break
			}

			// No modal open, handle tab switching
			switch msg.String() {
			case "1":
				m.Mode = ModeViewCalendar
				// Reset calendar to current month when switching to it
				m.Calendar.ResetToCurrentMonth()
				return m, nil
			case "2":
				m.Mode = ModeViewProjects
				return m, nil
			case "3":
				m.Mode = ModeViewWorkhourDetails
				return m, nil
			}
		}
	}

	if m.Mode == ModeViewCalendar {
		var cmd tea.Cmd
		var updatedModel tea.Model
		updatedModel, cmd = m.Calendar.Update(msg)
		m.Calendar = updatedModel.(CalendarModel)
		return m, cmd
	}

	if m.Mode == ModeViewProjects {
		var cmd tea.Cmd
		var updatedModel tea.Model
		updatedModel, cmd = m.Projects.Update(msg)
		m.Projects = updatedModel.(ProjectsModel)
		return m, cmd
	}

	if m.Mode == ModeViewWorkhourDetails {
		var cmd tea.Cmd
		var updatedModel tea.Model
		updatedModel, cmd = m.WorkhourDetails.Update(msg)
		m.WorkhourDetails = updatedModel.(WorkhourDetailsModel)
		return m, cmd
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	var content string
	var activeTabIndex int

	switch m.Mode {
	case ModeViewCalendar:
		activeTabIndex = 0
		content = m.Calendar.View()

	case ModeViewProjects:
		activeTabIndex = 1
		content = m.Projects.View()

	case ModeViewWorkhourDetails:
		activeTabIndex = 2
		content = m.WorkhourDetails.View()

	default:
		return ""
	}

	return render.RenderPageLayoutWithTabs(activeTabIndex, content)
}
