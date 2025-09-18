package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/smazurov/sony_remote_ble/sony_remote_ble"
)

func (m *Model) deviceListView() string {
	var sections []string

	// Title
	title := titleStyle.Render(fmt.Sprintf("Sony Camera Remote %s", m.version))
	sections = append(sections, title)

	// Status with spinner
	spinner := ""
	if m.scanning {
		spinners := []string{"|", "/", "-", "\\"}
		spinner = spinners[m.spinnerIndex%len(spinners)] + " "
	}

	statusText := "Status: Ready to scan"
	if m.scanning {
		if len(m.devices) == 0 {
			statusText = fmt.Sprintf("Status: %sScanning for Sony cameras... (none found yet)", spinner)
		} else {
			statusText = fmt.Sprintf("Status: %sScanning... found %d device(s) so far", spinner, len(m.devices))
		}
	} else if len(m.devices) == 0 {
		statusText = "Status: No devices found. Press Tab to scan."
	} else {
		statusText = fmt.Sprintf("Status: Scan complete - found %d device(s)", len(m.devices))
	}
	sections = append(sections, statusText)

	// Device list - show devices found during scanning
	if len(m.devices) > 0 || m.scanning {
		if len(m.devices) > 0 {
			sections = append(sections, "\nDevices:")
			for i, device := range m.devices {
				prefix := "  "
				style := deviceStyle
				if i == m.selected {
					prefix = "▶ "
					style = selectedDeviceStyle
				}

				// Add scanning indicator
				scanIndicator := ""
				if m.scanning {
					scanIndicator = " [scanning...]"
				}

				deviceLine := fmt.Sprintf("%s%s (%s) RSSI: %d%s",
					prefix, device.Name, device.Address, device.RSSI, scanIndicator)
				sections = append(sections, style.Render(deviceLine))
			}
		} else if m.scanning {
			sections = append(sections, "\nDevices:")
			sections = append(sections, "  Searching for Sony cameras... [scanning]")
		}
	}

	// Controls help
	help := []string{
		"Controls:",
		"↑/↓ or k/j - Navigate devices",
		"Tab - Scan for devices",
		"Enter - Connect to selected device",
		"Esc - Stop scanning",
		"q - Quit",
	}
	sections = append(sections, "\n"+helpStyle.Render(strings.Join(help, "\n")))

	// Logs
	if len(m.logs) > 0 {
		logLines := m.getLastLogs(3)
		logContent := strings.Join(logLines, "\n")

		// Calculate log width - needs to fit inside container
		logWidth := 60 // Default width
		if m.width > 20 {
			logWidth = m.width - 12 // Account for container + log borders and padding
		}
		if logWidth < 40 {
			logWidth = 40
		}

		logStyleWithWidth := logStyle.Width(logWidth)
		sections = append(sections, "\n"+logStyleWithWidth.Render(logContent))
	}

	// Use nearly full terminal width
	containerWidth := max(m.width-2, 60)

	return containerStyle.Width(containerWidth).Render(strings.Join(sections, "\n"))
}

func (m *Model) controlView() string {
	var sections []string

	// Title with connection status
	connected := m.client.State() == sony_remote_ble.Connected
	connectionStatus := "Disconnected"
	statusStyle := disconnectedStyle
	if connected {
		connectionStatus = "Connected to " + m.client.DeviceName()
		statusStyle = connectedStyle
	}

	title := titleStyle.Render("Sony Camera Remote")
	status := statusStyle.Render(connectionStatus)
	sections = append(sections, title+" | "+status)

	// Main control interface using the compact design
	controlInterface := m.renderControlInterface()
	sections = append(sections, controlInterface)

	// Quick actions
	quickActions := m.renderQuickActions()
	sections = append(sections, quickActions)

	// Controls help
	help := []string{
		"Controls:",
		"F/f - Focus | S/s - Shutter | Z/z - Zoom | A - AutoFocus",
		"Space - Quick Shot | R - Record | C - Custom | Esc - Back",
		"Q - Quit",
	}
	sections = append(sections, helpStyle.Render(strings.Join(help, "\n")))

	// Logs
	if len(m.logs) > 0 {
		logLines := m.getLastLogs(3)
		logContent := strings.Join(logLines, "\n")

		// Calculate log width - needs to fit inside container
		logWidth := 60 // Default width
		if m.width > 20 {
			logWidth = m.width - 12 // Account for container + log borders and padding
		}
		if logWidth < 40 {
			logWidth = 40
		}

		logStyleWithWidth := logStyle.Width(logWidth)
		sections = append(sections, logStyleWithWidth.Render(logContent))
	}

	// Use nearly full terminal width
	containerWidth := m.width - 2 // Just a small margin
	if containerWidth < 60 {
		containerWidth = 60
	}

	return containerStyle.Width(containerWidth).Render(strings.Join(sections, "\n"))
}

func (m *Model) renderControlInterface() string {
	disabled := m.client.State() != sony_remote_ble.Connected

	// Zoom controls row
	zoomOut := GetButtonStyle(m.buttonStates["zoom_out"], disabled).Render("Z-")
	zoomIn := GetButtonStyle(m.buttonStates["zoom_in"], disabled).Render("Z+")
	zoomRow := fmt.Sprintf("    %s  ◀──── ZOOM ────▶  %s", zoomOut, zoomIn)

	// Main control buttons in 2x2 grid
	autoFocus := GetButtonStyle(m.buttonStates["autofocus"], disabled).
		Width(6).Render("AF")
	focus := GetButtonStyle(m.buttonStates["focus"], disabled).
		Width(6).Render("FOCUS")
	shutter := GetButtonStyle(m.buttonStates["shutter"], disabled).
		Width(6).Render("SHUTR")
	record := GetButtonStyle(m.buttonStates["record"], disabled).
		Width(6).Render("REC")

	controlGrid := lipgloss.JoinVertical(lipgloss.Center,
		"",
		zoomRow,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top,
			"         ",
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.JoinHorizontal(lipgloss.Top, autoFocus, " ", focus),
				lipgloss.JoinHorizontal(lipgloss.Top, shutter, " ", record),
			),
		),
		"",
	)

	return controlGrid
}

func (m *Model) renderQuickActions() string {
	disabled := m.client.State() != sony_remote_ble.Connected

	custom := GetButtonStyle(m.buttonStates["custom"], disabled).Render("C1")
	quickShot := GetButtonStyle(m.buttonStates["shutter"], disabled).Render("Quick Shot")
	recordBtn := GetButtonStyle(m.buttonStates["record"], disabled).Render("Record")

	actions := lipgloss.JoinHorizontal(lipgloss.Top,
		"Quick Actions: ",
		custom, " Custom  ",
		quickShot, " Take Photo  ",
		recordBtn, " Record",
	)

	return actions
}

// Helper to render button with consistent styling
func (m *Model) renderButton(text string, key string, disabled bool) string {
	active := m.buttonStates[key]
	return GetButtonStyle(active, disabled).Render(text)
}
