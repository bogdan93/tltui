package projects

import (
	"fmt"
	"tltui/src/common"
	"tltui/src/domain"
	"tltui/src/domain/repository"
	"tltui/src/render"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProjectsModel struct {
	Width  int
	Height int

	ProjectEditModal   *ProjectEditModal
	ProjectCreateModal *ProjectCreateModal
	ProjectDeleteModal *ProjectDeleteModal

	ProjectsTable    table.Model
	ProjectsViewport viewport.Model
	Projects         []domain.Project
	NextID           int
}

func NewProjectsModel() ProjectsModel {
	m := ProjectsModel{}

	projects, err := repository.GetAllProjectsFromDB()
	if err != nil {
		projects = []domain.Project{}
	}
	m.Projects = projects

	m.NextID = 1
	for _, p := range m.Projects {
		if p.ID >= m.NextID {
			m.NextID = p.ID + 1
		}
	}

	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Project Name", Width: 30},
		{Title: "Odoo ID", Width: 10},
	}

	rows := []table.Row{}
	for _, p := range m.Projects {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.ID),
			p.Name,
			fmt.Sprintf("%d", p.OdooID),
		})
	}

	m.ProjectsTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
		table.WithWidth(50),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("39"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)

	m.ProjectsTable.SetStyles(s)
	m.ProjectsViewport = viewport.New(100, 100)
	m.ProjectsTable.SetHeight(100)

	return m
}

func (m ProjectsModel) Init() tea.Cmd {
	return nil
}

func (m ProjectsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ProjectCreatedMsg:
		newProject := domain.Project{
			ID:     m.NextID,
			Name:   msg.Name,
			OdooID: msg.OdooID,
		}

		err := repository.CreateProject(newProject)
		if err != nil {
			m.ProjectCreateModal = nil
			return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to create project: %v", err))
		}

		projects, err := repository.GetAllProjectsFromDB()
		if err != nil {
			m.ProjectCreateModal = nil
			return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to reload projects: %v", err))
		}
		m.Projects = projects
		m.NextID++

		rows := []table.Row{}
		for _, p := range m.Projects {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", p.ID),
				p.Name,
				fmt.Sprintf("%d", p.OdooID),
			})
		}
		m.ProjectsTable.SetRows(rows)
		m.ProjectCreateModal = nil
		return m, nil

	case ProjectCreateCanceledMsg:
		m.ProjectCreateModal = nil
		return m, nil

	case ProjectEditedMsg:
		updatedProject := domain.Project{
			ID:     msg.ProjectID,
			Name:   msg.Name,
			OdooID: msg.OdooID,
		}
		err := repository.UpdateProject(updatedProject)
		if err != nil {
			m.ProjectEditModal = nil
			return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to update project: %v", err))
		}

		projects, err := repository.GetAllProjectsFromDB()
		if err != nil {
			m.ProjectEditModal = nil
			return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to reload projects: %v", err))
		}
		m.Projects = projects

		rows := []table.Row{}
		for _, p := range m.Projects {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", p.ID),
				p.Name,
				fmt.Sprintf("%d", p.OdooID),
			})
		}
		m.ProjectsTable.SetRows(rows)
		m.ProjectEditModal = nil
		return m, nil

	case ProjectEditCanceledMsg:
		m.ProjectEditModal = nil
		return m, nil

	case ProjectDeletedMsg:
		err := repository.DeleteProject(msg.ProjectID)
		if err != nil {
			m.ProjectDeleteModal = nil
			return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to delete project: %v", err))
		}

		projects, err := repository.GetAllProjectsFromDB()
		if err != nil {
			m.ProjectDeleteModal = nil
			return m, common.DispatchErrorNotification(fmt.Sprintf("Failed to reload projects: %v", err))
		}
		m.Projects = projects

		rows := []table.Row{}
		for _, p := range m.Projects {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", p.ID),
				p.Name,
				fmt.Sprintf("%d", p.OdooID),
			})
		}
		m.ProjectsTable.SetRows(rows)
		m.ProjectDeleteModal = nil
		return m, nil

	case ProjectDeleteCanceledMsg:
		m.ProjectDeleteModal = nil
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		verticalMargin := 12
		tableHeight := msg.Height - verticalMargin
		m.ProjectsViewport.Width = msg.Width - 2
		m.ProjectsViewport.Height = tableHeight
		m.ProjectsTable.SetHeight(tableHeight)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Only close delete modal on 'q' (edit/create modals have text inputs where user might type 'q')
			if m.ProjectDeleteModal != nil {
				m.ProjectDeleteModal = nil
				return m, nil
			}
			// If edit/create modal is open, don't intercept - let the modal/textinput handle it
			if m.ProjectEditModal != nil || m.ProjectCreateModal != nil {
				break // Don't quit, let it pass through to modal forwarding
			}
			// No modal open, quit
			return m, tea.Quit

		case "n":
			if m.ProjectEditModal == nil && m.ProjectCreateModal == nil && m.ProjectDeleteModal == nil {
				m.ProjectCreateModal = NewProjectCreateModal()
				return m, nil
			}

		case "d":
			if m.ProjectEditModal == nil && m.ProjectCreateModal == nil && m.ProjectDeleteModal == nil {
				selectedProject := m.getSelectedProject()
				if selectedProject != nil {
					m.ProjectDeleteModal = NewProjectDeleteModal(
						selectedProject.ID,
						selectedProject.Name,
					)
					return m, nil
				}
			}

		case "enter":
			if m.ProjectEditModal == nil && m.ProjectCreateModal == nil && m.ProjectDeleteModal == nil {
				selectedProject := m.getSelectedProject()
				if selectedProject != nil {
					m.ProjectEditModal = NewProjectEditModal(
						selectedProject.ID,
						selectedProject.Name,
						selectedProject.OdooID,
					)
					return m, nil
				}
			}
		}
	}

	if m.ProjectEditModal != nil {
		_, cmd := m.ProjectEditModal.Update(msg)
		return m, cmd
	}

	if m.ProjectCreateModal != nil {
		_, cmd := m.ProjectCreateModal.Update(msg)
		return m, cmd
	}

	if m.ProjectDeleteModal != nil {
		_, cmd := m.ProjectDeleteModal.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.ProjectsTable, cmd = m.ProjectsTable.Update(msg)
	cmds = append(cmds, cmd)

	m.ProjectsViewport, cmd = m.ProjectsViewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ProjectsModel) View() string {
	helpText := render.RenderHelpText("↑/↓: navigate", "enter: edit", "n: new", "d: delete", "q: quit")
	m.ProjectsViewport.SetContent(m.ProjectsTable.View())

	if m.ProjectEditModal != nil {
		return m.ProjectEditModal.View(m.Width, m.Height)
	}

	if m.ProjectCreateModal != nil {
		return m.ProjectCreateModal.View(m.Width, m.Height)
	}

	if m.ProjectDeleteModal != nil {
		return m.ProjectDeleteModal.View(m.Width, m.Height)
	}

	return m.ProjectsViewport.View() + "\n" + helpText
}

func (m ProjectsModel) getSelectedProject() *domain.Project {
	cursor := m.ProjectsTable.Cursor()
	if cursor >= 0 && cursor < len(m.Projects) {
		return &m.Projects[cursor]
	}
	return nil
}
