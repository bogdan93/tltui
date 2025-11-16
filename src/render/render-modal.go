package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderModal(termWidth, termHeight int, modalWidth, modalHeight int, content string) string {
	if modalWidth == 0 {
		modalWidth = termWidth * 2 / 3
	}
	if modalHeight == 0 {
		modalHeight = termHeight * 2 / 3
	}

	if modalWidth > termWidth-4 {
		modalWidth = termWidth - 4
	}
	if modalHeight > termHeight-4 {
		modalHeight = termHeight - 4
	}

	modalStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Width(modalWidth).
		Height(modalHeight)

	modalContent := modalStyle.Render(content)

	contentHeight := lipgloss.Height(modalContent)
	verticalPadding := max((termHeight-contentHeight)/2, 0)

	contentWidth := lipgloss.Width(modalContent)
	horizontalPadding := max((termWidth-contentWidth)/2, 0)

	var sb strings.Builder

	for range verticalPadding {
		sb.WriteString("\n")
	}

	lines := strings.SplitSeq(modalContent, "\n")
	for line := range lines {
		sb.WriteString(strings.Repeat(" ", horizontalPadding))
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

/** RenderSimpleModal is a convenience function that auto-calculates modal size */
func RenderSimpleModal(termWidth, termHeight int, content string) string {
	return RenderModal(termWidth, termHeight, 0, 0, content)
}
