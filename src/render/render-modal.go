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

	// Ensure modal doesn't exceed terminal size
	if modalWidth > termWidth-4 {
		modalWidth = termWidth - 4
	}
	if modalHeight > termHeight-4 {
		modalHeight = termHeight - 4
	}

	// Create the modal style with border
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")). // Purple border
		Padding(1, 2).
		Width(modalWidth).
		Height(modalHeight)

	// Render the content inside the modal
	modalContent := modalStyle.Render(content)

	// Calculate vertical centering
	contentHeight := lipgloss.Height(modalContent)
	verticalPadding := max((termHeight-contentHeight)/2, 0)

	// Calculate horizontal centering
	contentWidth := lipgloss.Width(modalContent)
	horizontalPadding := max((termWidth-contentWidth)/2, 0)

	// Create the centered layout
	var sb strings.Builder

	// Add vertical padding (top)
	for range verticalPadding {
		sb.WriteString("\n")
	}

	// Add horizontal padding and content
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
