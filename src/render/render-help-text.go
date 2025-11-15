package render

import "github.com/charmbracelet/lipgloss"

var helpText = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241")).
	MarginTop(1)

// Variadic arguments array of strings
func RenderHelpText(
	args ...string,
) string {
	help := ""
	for i, arg := range args {
		if i > 0 {
			help += " • "
		}
		help += arg
	}
	return helpText.Render(help)
}
