package models

import (
	"fmt"
	"time-logger-tui/src/render"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Project struct {
	ID     int
	Name   string
	OdooID int
}

type ProjectsModel struct {
	Width  int
	Height int

	ProjectEditModal   *ProjectEditModal
	ProjectCreateModal *ProjectCreateModal
	ProjectDeleteModal *ProjectDeleteModal

	ProjectsTable    table.Model
	ProjectsViewport viewport.Model
	Projects         []Project
	NextID           int // Track next available ID for new projects
}

func NewProjectsModel() ProjectsModel {
	m := ProjectsModel{}
	m.Projects = FetchAllProjects()

	// Calculate next available ID
	m.NextID = 1
	for _, p := range m.Projects {
		if p.ID >= m.NextID {
			m.NextID = p.ID + 1
		}
	}

	// Setup table columns
	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Project Name", Width: 30},
		{Title: "Odoo ID", Width: 10},
	}

	// Convert projects to table rows
	rows := []table.Row{}
	for _, p := range m.Projects {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.ID),
			p.Name,
			fmt.Sprintf("%d", p.OdooID),
		})
	}

	// Create table
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
		newProject := Project{
			ID:     m.NextID,
			Name:   msg.Name,
			OdooID: msg.OdooID,
		}
		m.Projects = append(m.Projects, newProject)
		m.NextID++

		// Update table
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
		for i := range m.Projects {
			if m.Projects[i].ID == msg.ProjectID {
				m.Projects[i].Name = msg.Name
				m.Projects[i].OdooID = msg.OdooID
				break
			}
		}
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
		// Find and delete the project
		for i := range m.Projects {
			if m.Projects[i].ID == msg.ProjectID {
				// Remove project from slice
				m.Projects = append(m.Projects[:i], m.Projects[i+1:]...)
				break
			}
		}

		// Update table
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
		verticalMargin := 12 // Increased to account for tab bar (4 lines) + padding
		tableHeight := msg.Height - verticalMargin
		m.ProjectsViewport.Width = msg.Width - 2
		m.ProjectsViewport.Height = tableHeight
		m.ProjectsTable.SetHeight(tableHeight)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.ProjectEditModal != nil {
				m.ProjectEditModal = nil
				return m, nil
			}
			if m.ProjectCreateModal != nil {
				m.ProjectCreateModal = nil
				return m, nil
			}
			if m.ProjectDeleteModal != nil {
				m.ProjectDeleteModal = nil
				return m, nil
			}
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

// getSelectedProject returns the currently selected project from the table
func (m ProjectsModel) getSelectedProject() *Project {
	cursor := m.ProjectsTable.Cursor()
	if cursor >= 0 && cursor < len(m.Projects) {
		return &m.Projects[cursor]
	}
	return nil
}
