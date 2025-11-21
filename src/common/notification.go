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

func NotifyError(action string, err error) tea.Cmd {
	return DispatchErrorNotification(action + ": " + err.Error())
}

func NotifySuccess(message string) tea.Cmd {
	return DispatchSuccessNotification(message)
}

func NotifyInfo(message string) tea.Cmd {
	return DispatchInfoNotification(message)
}

func RenderNotification(notification Notification) string {
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

func StartNotificationTimeout(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return ClearNotificationMsg{}
	})
}

func RenderNotificationOverlay(notification Notification, mainView string, terminalWidth int) string {
	notificationBar := RenderNotification(notification)

	lines := strings.Split(mainView, "\n")
	if len(lines) == 0 {
		return notificationBar
	}

	notificationWidth := lipgloss.Width(notificationBar)

	lastLineIndex := len(lines) - 2

	if lastLineIndex < 0 {
		lastLineIndex = len(lines) - 1
	}

	padding := terminalWidth - notificationWidth;

	lines[lastLineIndex] = strings.Repeat(" ", padding) + notificationBar

	return strings.Join(lines, "\n")
}
