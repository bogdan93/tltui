package models

import (
	"fmt"
	"strings"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProjectDeleteModal struct {
	ProjectID   int
	ProjectName string
}

type ProjectDeletedMsg struct {
	ProjectID int
}

type ProjectDeleteCanceledMsg struct{}

func NewProjectDeleteModal(projectID int, projectName string) *ProjectDeleteModal {
	return &ProjectDeleteModal{
		ProjectID:   projectID,
		ProjectName: projectName,
	}
}

func (m *ProjectDeleteModal) Update(msg tea.Msg) (ProjectDeleteModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			// Confirm delete
			return *m, tea.Batch(
				dispatchProjectDeletedMsg(m.ProjectID),
			)

		case "n", "N", "esc":
			// Cancel delete
			return *m, tea.Batch(
				dispatchProjectDeleteCanceledMsg(),
			)
		}
	}

	return *m, nil
}

func (m *ProjectDeleteModal) View(Width, Height int) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")). // Red for delete
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("241"))

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")). // Orange warning
		Bold(true)

	sb.WriteString(titleStyle.Render("⚠ Delete Project"))
	sb.WriteString("\n\n")

	sb.WriteString(warningStyle.Render("Are you sure you want to delete this project?"))
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("ID: "))
	sb.WriteString(fmt.Sprintf("%d", m.ProjectID))
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("Name: "))
	sb.WriteString(m.ProjectName)
	sb.WriteString("\n\n")

	warningStyle2 := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
	sb.WriteString(warningStyle2.Render("This action cannot be undone!"))
	sb.WriteString("\n\n")

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
	sb.WriteString(helpStyle.Render("Y/Enter: confirm delete • N/ESC: cancel"))

	return render.RenderSimpleModal(Width, Height, sb.String())
}

func dispatchProjectDeletedMsg(projectID int) tea.Cmd {
	return func() tea.Msg {
		return ProjectDeletedMsg{
			ProjectID: projectID,
		}
	}
}

func dispatchProjectDeleteCanceledMsg() tea.Cmd {
	return func() tea.Msg {
		return ProjectDeleteCanceledMsg{}
	}
}
