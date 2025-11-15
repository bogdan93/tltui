package render

import "github.com/charmbracelet/lipgloss"

var globalStyle = lipgloss.NewStyle().Padding(1)

func RenderPageLayout(title, content string) string {
	titleString := RenderPageTitle(title)
	fullContent := globalStyle.Render(titleString + "\n" + content)
	return fullContent
}

func RenderPageLayoutWithTabs(activeTabIndex int, content string) string {
	tabs := []Tab{
		{Key: "1", Label: "Calendar"},
		{Key: "2", Label: "Projects"},
		{Key: "3", Label: "Workhour Details"},
	}

	tabBar := RenderTabBar(tabs, activeTabIndex)
	fullContent := globalStyle.Render(tabBar + "\n" + content)
	return fullContent
}
