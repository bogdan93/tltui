package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Tab struct {
	Key   string
	Label string
}

func RenderTabBar(tabs []Tab, activeIndex int) string {
	var sb strings.Builder

	// Style for active tab
	activeTabStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Padding(0, 1)

	// Style for inactive tab
	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 1)

	// Border style for the tab bar
	tabBarStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("240")).
		MarginBottom(1)

	for i, tab := range tabs {
		if i > 0 {
			sb.WriteString("  ")
		}

		tabText := "[" + tab.Key + "] " + tab.Label

		if i == activeIndex {
			sb.WriteString(activeTabStyle.Render(tabText))
		} else {
			sb.WriteString(inactiveTabStyle.Render(tabText))
		}
	}

	return tabBarStyle.Render(sb.String())
}
