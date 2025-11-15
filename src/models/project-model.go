package models

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type Project struct {
	ID     int
	Name   string
	OdooID int
}

func GetProjects() []Project {
	return []Project{
		{ID: 1, Name: "Project A", OdooID: 101},
		{ID: 2, Name: "Project B", OdooID: 102},
		{ID: 3, Name: "Internal Tools", OdooID: 103},
		{ID: 4, Name: "Client Website", OdooID: 104},
		{ID: 5, Name: "Mobile App", OdooID: 105},
		{ID: 6, Name: "E-Commerce Platform", OdooID: 106},
		{ID: 7, Name: "CRM System", OdooID: 107},
		{ID: 8, Name: "Analytics Dashboard", OdooID: 108},
		{ID: 9, Name: "Payment Gateway", OdooID: 109},
		{ID: 10, Name: "Inventory Management", OdooID: 110},
		{ID: 11, Name: "HR Portal", OdooID: 111},
		{ID: 12, Name: "Customer Support Tool", OdooID: 112},
		{ID: 13, Name: "Marketing Automation", OdooID: 113},
		{ID: 14, Name: "API Gateway", OdooID: 114},
		{ID: 15, Name: "DevOps Pipeline", OdooID: 115},
		{ID: 16, Name: "Security Audit", OdooID: 116},
		{ID: 17, Name: "Database Migration", OdooID: 117},
		{ID: 18, Name: "Cloud Infrastructure", OdooID: 118},
		{ID: 19, Name: "Machine Learning Model", OdooID: 119},
		{ID: 20, Name: "IoT Platform", OdooID: 120},
	}
}

func ProjectsModelInit() table.Model {
	projects := GetProjects()

	// Setup table columns
	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Project Name", Width: 30},
		{Title: "Odoo ID", Width: 10},
	}

	// Convert projects to table rows
	rows := []table.Row{}
	for _, p := range projects {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.ID),
			p.Name,
			fmt.Sprintf("%d", p.OdooID),
		})
	}

	// Create table
	projectsTable := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Style the table
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

	projectsTable.SetStyles(s)

	return projectsTable
}
