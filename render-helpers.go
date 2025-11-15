package main

import (
	"github.com/charmbracelet/lipgloss"
)


func RenderPageTitle(title string) string {
	titleStyle := lipgloss.NewStyle().Bold(true).
		Underline(true).
		MarginBottom(1).
		Foreground(lipgloss.Color("#FFA500")) // Orange color

	return titleStyle.Render(title)
}
