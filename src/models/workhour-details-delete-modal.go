package models

import (
	"fmt"
	"strings"
	"tltui/src/render"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkhourDetailsDeleteModal struct {
	WorkhourDetailID   int
	WorkhourDetailName string
}

type WorkhourDetailsDeletedMsg struct {
	WorkhourDetailID int
}

type WorkhourDetailsDeleteCanceledMsg struct{}

func NewWorkhourDetailsDeleteModal(workhourDetailID int, workhourDetailName string) *WorkhourDetailsDeleteModal {
	return &WorkhourDetailsDeleteModal{
		WorkhourDetailID:   workhourDetailID,
		WorkhourDetailName: workhourDetailName,
	}
}

func (m *WorkhourDetailsDeleteModal) Update(msg tea.Msg) (WorkhourDetailsDeleteModal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			// Confirm delete
			return *m, tea.Batch(
				dispatchWorkhourDetailsDeletedMsg(m.WorkhourDetailID),
			)

		case "n", "N", "esc":
			// Cancel delete
			return *m, tea.Batch(
				dispatchWorkhourDetailsDeleteCanceledMsg(),
			)
		}
	}

	return *m, nil
}

func (m *WorkhourDetailsDeleteModal) View(Width, Height int) string {
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

	sb.WriteString(titleStyle.Render("⚠ Delete Workhour Detail"))
	sb.WriteString("\n\n")

	sb.WriteString(warningStyle.Render("Are you sure you want to delete this workhour detail?"))
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("ID: "))
	sb.WriteString(fmt.Sprintf("%d", m.WorkhourDetailID))
	sb.WriteString("\n\n")

	sb.WriteString(labelStyle.Render("Name: "))
	sb.WriteString(m.WorkhourDetailName)
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

func dispatchWorkhourDetailsDeletedMsg(workhourDetailID int) tea.Cmd {
	return func() tea.Msg {
		return WorkhourDetailsDeletedMsg{
			WorkhourDetailID: workhourDetailID,
		}
	}
}

func dispatchWorkhourDetailsDeleteCanceledMsg() tea.Cmd {
	return func() tea.Msg {
		return WorkhourDetailsDeleteCanceledMsg{}
	}
}
