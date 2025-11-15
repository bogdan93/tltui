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

	ProjectEditModal *ProjectEditModal

	ProjectsTable    table.Model
	ProjectsViewport viewport.Model
	Projects         []Project
}

func NewProjectsModel() ProjectsModel {
	m := ProjectsModel{}
	m.Projects = FetchAllProjects()

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
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		verticalMargin := 8
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
			return m, tea.Quit
		case "enter":
			if m.ProjectEditModal == nil {
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

	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.ProjectsTable, cmd = m.ProjectsTable.Update(msg)
	cmds = append(cmds, cmd)

	m.ProjectsViewport, cmd = m.ProjectsViewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ProjectsModel) View() string {
	helpText := render.RenderHelpText("↑/↓: navigate", "enter: select", "q: quit")
	m.ProjectsViewport.SetContent(m.ProjectsTable.View())

	if m.ProjectEditModal != nil {
		return m.ProjectEditModal.View(m.Width, m.Height)
	}

	return render.RenderPageLayout(
		"Projects",
		m.ProjectsViewport.View()+"\n"+helpText,
	)
}

// getSelectedProject returns the currently selected project from the table
func (m ProjectsModel) getSelectedProject() *Project {
	cursor := m.ProjectsTable.Cursor()
	if cursor >= 0 && cursor < len(m.Projects) {
		return &m.Projects[cursor]
	}
	return nil
}
