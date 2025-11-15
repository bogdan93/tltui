package models

import (
	"time"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
)

type WorkhourDetails struct {
	ID        int
	Name      string
	ShortName string
	IsWork    bool
}

type Workhour struct {
	ID        int
	Date      time.Time
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

	// Notification system
	Notification *Notification
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case ShowNotificationMsg:
		// Show notification and set timeout to clear it
		m.Notification = &Notification{
			Message: msg.Message,
			Type:    msg.Type,
		}
		return m, StartNotificationTimeout(3 * time.Second)

	case ClearNotificationMsg:
		// Clear notification
		m.Notification = nil
		return m, nil

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
		// Check if any modal is open
		isModalOpen := m.Calendar.WorkhoursViewModal != nil ||
			m.Calendar.ReportGeneratorModal != nil ||
			m.Calendar.ShowHelp ||
			m.Projects.ProjectEditModal != nil ||
			m.Projects.ProjectCreateModal != nil ||
			m.Projects.ProjectDeleteModal != nil ||
			m.WorkhourDetails.WorkhourDetailsEditModal != nil ||
			m.WorkhourDetails.WorkhourDetailsCreateModal != nil ||
			m.WorkhourDetails.WorkhourDetailsDeleteModal != nil

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			// If a modal is open, let it handle the quit key to close itself
			if !isModalOpen {
				return m, tea.Quit
			}
		case "1", "2", "3":
			// Don't handle tab switching if any modal is open
			if isModalOpen {
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

	mainView := render.RenderPageLayoutWithTabs(activeTabIndex, content)

	// Overlay notification bar at the top if present
	if m.Notification != nil {
		return RenderNotificationOverlay(m.Notification, mainView)
	}

	return mainView
}
