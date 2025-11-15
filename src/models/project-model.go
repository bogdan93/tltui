package models

import (
	"fmt"
	"strings"
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

	ShowModal        bool
	ModalContent     string
	ProjectsTable    table.Model
	ProjectsViewport viewport.Model
	Projects         []Project
}

// NewProjectsModel creates and initializes a new ProjectsModel
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
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Account for: padding top (1) + padding bottom (1) + title (2 lines with margin) + help text (2 lines with margin)
		verticalMargin := 8
		tableHeight := msg.Height - verticalMargin

		// Update viewport and table dimensions on resize
		m.ProjectsViewport.Width = msg.Width - 2
		m.ProjectsViewport.Height = tableHeight
		m.ProjectsTable.SetHeight(tableHeight)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Close modal if open
			if m.ShowModal {
				m.ShowModal = false
				m.ModalContent = ""
				return m, nil
			}

		case "enter":
			// Show project details modal when in projects view
			if !m.ShowModal {
				selectedProject := m.getSelectedProject()
				if selectedProject != nil {
					m.ModalContent = m.formatProjectDetails(*selectedProject)
					m.ShowModal = true
					return m, nil
				}
			}
		}
	}

	// Update the table and viewport based on current mode (only if modal is not shown)
	if !m.ShowModal {
		m.ProjectsTable, cmd = m.ProjectsTable.Update(msg)
		cmds = append(cmds, cmd)

		m.ProjectsViewport, cmd = m.ProjectsViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)

}

func (m ProjectsModel) View() string {
	helpText := render.RenderHelpText("↑/↓: navigate", "enter: select", "q: quit")
	m.ProjectsViewport.SetContent(m.ProjectsTable.View())

	if m.ShowModal {
		return render.RenderModal(m.Width, m.Height, 0, 0, m.ModalContent)
	}

	return render.RenderPageLayout(
		"Projects",
		m.ProjectsViewport.View() + "\n" + helpText,
	);
}

// getSelectedProject returns the currently selected project from the table
func (m ProjectsModel) getSelectedProject() *Project {
	cursor := m.ProjectsTable.Cursor()
	if cursor >= 0 && cursor < len(m.Projects) {
		return &m.Projects[cursor]
	}
	return nil
}

// formatProjectDetails formats a project's details for display in the modal
func (m ProjectsModel) formatProjectDetails(project Project) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	sb.WriteString(titleStyle.Render("Project Details"))
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("ID: "))
	sb.WriteString(valueStyle.Render(fmt.Sprintf("%d", project.ID)))
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("Name: "))
	sb.WriteString(valueStyle.Render(project.Name))
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("Odoo ID: "))
	sb.WriteString(valueStyle.Render(fmt.Sprintf("%d", project.OdooID)))
	sb.WriteString("\n\n")

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
	sb.WriteString(helpStyle.Render("Press ESC to close"))

	return sb.String()
}
