package common

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NotificationType int

const (
	NotificationError NotificationType = iota
	NotificationSuccess
	NotificationInfo
)

type Notification struct {
	Message string
	Type    NotificationType
}

// Message types for notification system
type ShowNotificationMsg struct {
	Message string
	Type    NotificationType
}

type ClearNotificationMsg struct{}

// Helper functions to dispatch notifications
func DispatchErrorNotification(message string) tea.Cmd {
	return func() tea.Msg {
		return ShowNotificationMsg{
			Message: message,
			Type:    NotificationError,
		}
	}
}

func DispatchSuccessNotification(message string) tea.Cmd {
	return func() tea.Msg {
		return ShowNotificationMsg{
			Message: message,
			Type:    NotificationSuccess,
		}
	}
}

func DispatchInfoNotification(message string) tea.Cmd {
	return func() tea.Msg {
		return ShowNotificationMsg{
			Message: message,
			Type:    NotificationInfo,
		}
	}
}

// Simplified notification helpers with shorter names
// NotifyError creates an error notification from an error and action context
func NotifyError(action string, err error) tea.Cmd {
	return DispatchErrorNotification(action + ": " + err.Error())
}

// NotifySuccess creates a success notification
func NotifySuccess(message string) tea.Cmd {
	return DispatchSuccessNotification(message)
}

// NotifyInfo creates an info notification
func NotifyInfo(message string) tea.Cmd {
	return DispatchInfoNotification(message)
}

// RenderNotification renders a notification bar
func RenderNotification(notification *Notification) string {
	if notification == nil {
		return ""
	}

	var style lipgloss.Style
	var prefix string

	switch notification.Type {
	case NotificationError:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("196")).
			Padding(0, 2).
			Bold(true)
		prefix = "✗ "
	case NotificationSuccess:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("34")).
			Padding(0, 2).
			Bold(true)
		prefix = "✓ "
	case NotificationInfo:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("39")).
			Padding(0, 2).
			Bold(true)
		prefix = "ℹ "
	}

	return style.Render(prefix + notification.Message)
}

// StartNotificationTimeout returns a command that clears notification after delay
func StartNotificationTimeout(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return ClearNotificationMsg{}
	})
}

// RenderNotificationOverlay overlays the notification at the bottom right of the main view
func RenderNotificationOverlay(notification *Notification, mainView string) string {
	if notification == nil {
		return mainView
	}

	// Get the notification bar
	notificationBar := RenderNotification(notification)

	// Split the main view into lines
	lines := strings.Split(mainView, "\n")
	if len(lines) == 0 {
		return notificationBar
	}

	// Get the width of the terminal (assume from first line)
	terminalWidth := len(lines[0])
	if terminalWidth == 0 {
		terminalWidth = 80 // fallback
	}

	// Get notification width (strip ANSI codes for accurate length)
	notificationWidth := lipgloss.Width(notificationBar)

	// Find the last non-empty line to place notification
	lastLineIndex := len(lines) - 1
	for lastLineIndex >= 0 && strings.TrimSpace(lines[lastLineIndex]) == "" {
		lastLineIndex--
	}

	if lastLineIndex < 0 {
		lastLineIndex = len(lines) - 1
	}

	// Calculate padding to align to the right
	padding := terminalWidth - notificationWidth
	if padding < 0 {
		padding = 0
	}

	// Position notification at bottom right by replacing the last line
	// Pad with spaces to push it to the right
	lines[lastLineIndex] = strings.Repeat(" ", padding) + notificationBar

	return strings.Join(lines, "\n")
}
