package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base colors
	primaryColor   = lipgloss.Color("#7C3AED")
	secondaryColor = lipgloss.Color("#10B981")
	dangerColor    = lipgloss.Color("#EF4444")
	warningColor   = lipgloss.Color("#F59E0B")
	textColor      = lipgloss.Color("#F3F4F6")
	mutedColor     = lipgloss.Color("#9CA3AF")
	bgColor        = lipgloss.Color("#1F2937")
	borderColor    = lipgloss.Color("#374151")

	// Title style
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1)

	// Status styles
	connectedStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	disconnectedStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	errorStyle = lipgloss.NewStyle().
			Foreground(dangerColor).
			Bold(true)

	// Button styles
	buttonStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true).
			Align(lipgloss.Center)

	activeButtonStyle = lipgloss.NewStyle().
				Foreground(bgColor).
				Background(secondaryColor).
				Padding(0, 2).
				Margin(0, 1).
				Bold(true).
				Align(lipgloss.Center)

	disabledButtonStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Background(borderColor).
				Padding(0, 2).
				Margin(0, 1).
				Align(lipgloss.Center)

	// Container styles
	containerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2)

	logStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1).
			Height(4)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Margin(1, 0)

	// Status bar style
	statusBarStyle = lipgloss.NewStyle().
			Background(borderColor).
			Foreground(textColor).
			Padding(0, 1)

	// Device list styles
	selectedDeviceStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	deviceStyle = lipgloss.NewStyle().
			Foreground(textColor)
)

// GetButtonStyle returns the appropriate button style based on state
func GetButtonStyle(active, disabled bool) lipgloss.Style {
	if disabled {
		return disabledButtonStyle
	}
	if active {
		return activeButtonStyle
	}
	return buttonStyle
}

// GetStatusStyle returns the appropriate status style
func GetStatusStyle(connected bool) lipgloss.Style {
	if connected {
		return connectedStyle
	}
	return disconnectedStyle
}