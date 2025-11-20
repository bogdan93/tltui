package common

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TableView wraps a table with a viewport for common table view functionality
type TableView struct {
	Table    table.Model
	Viewport viewport.Model
}

// NewTableView creates a new TableView with standard styling
func NewTableView(columns []table.Column, rows []table.Row) TableView {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
		table.WithWidth(50),
	)

	// Apply standard styling
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

	t.SetStyles(s)

	return TableView{
		Table:    t,
		Viewport: viewport.New(100, 100),
	}
}

// Update handles window resize and forwards other messages to table and viewport
func (tv TableView) Update(msg tea.Msg) (TableView, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	tv.Table, cmd = tv.Table.Update(msg)
	cmds = append(cmds, cmd)

	tv.Viewport, cmd = tv.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return tv, tea.Batch(cmds...)
}

// SetSize updates the table and viewport dimensions
func (tv *TableView) SetSize(width, height int, verticalMargin int) {
	tableHeight := height - verticalMargin
	tv.Viewport.Width = width - 2
	tv.Viewport.Height = tableHeight
	tv.Table.SetHeight(tableHeight)
}

// View renders the table within the viewport
func (tv *TableView) View() string {
	tv.Viewport.SetContent(tv.Table.View())
	return tv.Viewport.View()
}

// Cursor returns the current cursor position in the table
func (tv *TableView) Cursor() int {
	return tv.Table.Cursor()
}

// SetRows updates the table rows
func (tv *TableView) SetRows(rows []table.Row) {
	tv.Table.SetRows(rows)
}
