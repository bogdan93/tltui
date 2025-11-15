package render

import "github.com/charmbracelet/lipgloss"

var globalStyle = lipgloss.NewStyle().Padding(1)

func RenderPageLayout(title, content string) string {
	titleString := RenderPageTitle(title)
	fullContent := globalStyle.Render(titleString + "\n" + content)
	return fullContent
}
