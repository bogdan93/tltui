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
	// UI
	Width          int
	Height         int
	Ready          bool // After initial window size received

	// Projects List
	ShowModal      bool
	ModalContent   string
	ProjectsTable    table.Model
	ProjectsViewport viewport.Model
	Projects         []Project

	// State
	Mode AppMode

	// Data
	WorkhoursDetails []WorkhourDetails
	Workhours        []Workhour
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Account for: padding top (1) + padding bottom (1) + title (2 lines with margin) + help text (2 lines with margin)
		verticalMargin := 8
		tableHeight := msg.Height - verticalMargin

		if !m.Ready {
			// Initialize viewport and table on first window size message
			m.ProjectsViewport = viewport.New(msg.Width-2, tableHeight)
			m.ProjectsTable.SetHeight(tableHeight)
			m.Ready = true
		} else {
			// Update viewport and table dimensions on resize
			m.ProjectsViewport.Width = msg.Width - 2
			m.ProjectsViewport.Height = tableHeight
			m.ProjectsTable.SetHeight(tableHeight)
		}

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
			if m.Mode == ModeViewProjects && !m.ShowModal {
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
	if m.Mode == ModeViewProjects && !m.ShowModal {
		m.ProjectsTable, cmd = m.ProjectsTable.Update(msg)
		cmds = append(cmds, cmd)

		if m.Ready {
			m.ProjectsViewport, cmd = m.ProjectsViewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}


// getSelectedProject returns the currently selected project from the table
func (m AppModel) getSelectedProject() *Project {
	cursor := m.ProjectsTable.Cursor()
	if cursor >= 0 && cursor < len(m.Projects) {
		return &m.Projects[cursor]
	}
	return nil
}

// formatProjectDetails formats a project's details for display in the modal
func (m AppModel) formatProjectDetails(project Project) string {
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

var globalStyle = lipgloss.NewStyle().Padding(1)

func (m AppModel) View() string {
	if !m.Ready {
		return "\n  Initializing..."
	}

	var title string
	var content string

	switch m.Mode {
	case ModeViewProjects:
		title = "Projects"
		helpText := render.RenderHelpText("↑/↓: navigate", "enter: select", "q: quit")

		// Set the viewport content to the table
		m.ProjectsViewport.SetContent(m.ProjectsTable.View())
		content = m.ProjectsViewport.View() + "\n" + helpText

	case ModeViewWorkhours:
		title = "Workhours"
		content = "Press 'q' to quit."

	case ModeViewWorkhourDetails:
		title = "Workhour Details"
		content = "Press 'q' to quit."

	case ModeViewCalendar:
		title = "Calendar"
		content = "Press 'q' to quit."
	}

	titleString := render.RenderPageTitle(title)
	fullContent := globalStyle.Render(titleString + "\n" + content)

	// If modal is active, overlay it on top of the content
	if m.ShowModal {
		return render.RenderModal(m.Width, m.Height, 0, 0, m.ModalContent)
	}

	return fullContent
}
