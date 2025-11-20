package models

import (
	"time"
	"tltui/src/common"
	"tltui/src/elm-store/calendar"
	"tltui/src/elm-store/projects"
	"tltui/src/elm-store/workhour_details"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
)

type AppMode int

const (
	ModeViewCalendar AppMode = iota
	ModeViewProjects
	ModeViewWorkhourDetails
)

type AppModel struct {
	Mode AppMode

	Calendar        calendar.CalendarModel
	Projects        projects.ProjectsModel
	WorkhourDetails workhour_details.WorkhourDetailsModel

	Notification *common.Notification
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case common.ShowNotificationMsg:
		m.Notification = &common.Notification{
			Message: msg.Message,
			Type:    msg.Type,
		}
		return m, common.StartNotificationTimeout(3 * time.Second)

	case common.ClearNotificationMsg:
		m.Notification = nil
		return m, nil

	case tea.WindowSizeMsg:
		var cmd1, cmd2, cmd3 tea.Cmd
		var updatedModel tea.Model

		updatedModel, cmd1 = m.Calendar.Update(msg)
		m.Calendar = updatedModel.(calendar.CalendarModel)

		updatedModel, cmd2 = m.Projects.Update(msg)
		m.Projects = updatedModel.(projects.ProjectsModel)

		updatedModel, cmd3 = m.WorkhourDetails.Update(msg)
		m.WorkhourDetails = updatedModel.(workhour_details.WorkhourDetailsModel)

		return m, tea.Batch(cmd1, cmd2, cmd3)

	case tea.KeyMsg:
		isModalOpen := m.Calendar.ActiveModal != nil ||
			m.Calendar.ShowHelp ||
			m.Projects.ActiveModal != nil ||
			m.WorkhourDetails.ActiveModal != nil

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			if !isModalOpen {
				return m, tea.Quit
			}
		case "1", "2", "3":
			if isModalOpen {
				break
			}

			switch msg.String() {
			case "1":
				m.Mode = ModeViewCalendar
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
		m.Calendar = updatedModel.(calendar.CalendarModel)
		return m, cmd
	}

	if m.Mode == ModeViewProjects {
		var cmd tea.Cmd
		var updatedModel tea.Model
		updatedModel, cmd = m.Projects.Update(msg)
		m.Projects = updatedModel.(projects.ProjectsModel)
		return m, cmd
	}

	if m.Mode == ModeViewWorkhourDetails {
		var cmd tea.Cmd
		var updatedModel tea.Model
		updatedModel, cmd = m.WorkhourDetails.Update(msg)
		m.WorkhourDetails = updatedModel.(workhour_details.WorkhourDetailsModel)
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

	isModalOpened := m.Calendar.ActiveModal != nil || m.Projects.ActiveModal != nil || m.WorkhourDetails.ActiveModal != nil

	mainView := ""
	if !isModalOpened {
		mainView = render.RenderPageLayoutWithTabs(activeTabIndex, content)
	} else {
		mainView = content;
	}

	if m.Notification != nil {
		return common.RenderNotificationOverlay(m.Notification, mainView)
	}

	return mainView
}
