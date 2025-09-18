package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/smazurov/sony_remote_ble/sony_remote_ble"
)

type AppMode int

const (
	ModeDeviceList AppMode = iota
	ModeControl
)

type Model struct {
	client     *sony_remote_ble.Client
	mode       AppMode
	devices    []sony_remote_ble.DeviceInfo
	selected   int
	scanning   bool
	logs       []string
	ctx        context.Context
	cancel     context.CancelFunc
	deviceChan chan sony_remote_ble.DeviceInfo
	width      int
	height     int
	version    string

	// Animation state
	spinnerIndex int

	// Button states for visual feedback
	buttonStates map[string]bool
}

type tickMsg time.Time
type scanStartMsg struct{}
type scanCompleteMsg struct{}
type deviceFoundMsg sony_remote_ble.DeviceInfo
type connectionMsg struct {
	connected bool
	err       error
}
type commandSentMsg struct {
	command string
	err     error
}

func NewModel(version string) (*Model, error) {
	client, err := sony_remote_ble.NewClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Model{
		client:       client,
		mode:         ModeDeviceList,
		devices:      make([]sony_remote_ble.DeviceInfo, 0),
		logs:         make([]string, 0),
		ctx:          ctx,
		cancel:       cancel,
		deviceChan:   make(chan sony_remote_ble.DeviceInfo, 10),
		width:        80, // Default width
		height:       24, // Default height
		version:      version,
		buttonStates: make(map[string]bool),
	}

	m.addLog("Sony Camera Remote started. Press Tab to scan for devices.")
	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		m.checkForDevicesCmd(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tickMsg:
		if m.scanning {
			m.spinnerIndex = (m.spinnerIndex + 1) % 4
			return m, tea.Batch(tickCmd(), m.checkForDevicesCmd())
		}
		return m, tickCmd()

	case scanStartMsg:
		m.scanning = true
		m.devices = make([]sony_remote_ble.DeviceInfo, 0)
		m.selected = 0
		m.addLog("Starting scan for Sony cameras...")
		return m, m.performScan()

	case deviceFoundMsg:
		m.addDevice(sony_remote_ble.DeviceInfo(msg))
		if m.scanning {
			return m, m.checkForDevicesCmd()
		}
		return m, nil

	case scanCompleteMsg:
		wasScanning := m.scanning
		m.scanning = false
		if wasScanning {
			if len(m.devices) > 0 {
				m.addLog(fmt.Sprintf("Scan stopped - found %d device(s)", len(m.devices)))
			} else {
				m.addLog("Scan stopped - no devices found")
			}
		}
		return m, nil

	case connectionMsg:
		if msg.err != nil {
			m.addLog(fmt.Sprintf("Connection failed: %v", msg.err))
		} else if msg.connected {
			m.addLog("Connected to " + m.client.DeviceName())
			m.mode = ModeControl
		} else {
			m.addLog("Disconnected")
			m.mode = ModeDeviceList
		}
		return m, nil

	case commandSentMsg:
		m.buttonStates[msg.command] = false // Reset button state
		if msg.err != nil {
			m.addLog(fmt.Sprintf("Command failed: %v", msg.err))
		} else {
			m.addLog(fmt.Sprintf("Sent: %s", msg.command))
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeDeviceList:
		return m.handleDeviceListKeys(msg)
	case ModeControl:
		return m.handleControlKeys(msg)
	}
	return m, nil
}

func (m *Model) handleDeviceListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	m.addLog(fmt.Sprintf("Key pressed: '%s' Type: %d (scanning: %t)", key, msg.Type, m.scanning))

	// Check for ESC specifically
	if msg.Type == tea.KeyEsc {
		if m.scanning {
			m.addLog("ESC key detected - stopping scan...")
			m.client.StopScan()
			return m, func() tea.Msg { return scanCompleteMsg{} }
		}
	}

	switch key {
	case "q", "ctrl+c":
		m.cancel()
		return m, tea.Quit

	case "tab":
		if !m.scanning {
			return m, func() tea.Msg { return scanStartMsg{} }
		}

	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}

	case "down", "j":
		if m.selected < len(m.devices)-1 {
			m.selected++
		}

	case "enter":
		if len(m.devices) > 0 && m.selected < len(m.devices) {
			device := m.devices[m.selected]
			m.addLog(fmt.Sprintf("Connecting to %s...", device.Name))
			if m.scanning {
				m.client.StopScan()
			}
			return m, m.connect(device)
		}

	case "esc", "escape":
		if m.scanning {
			m.addLog("ESC pressed - stopping scan...")
			m.client.StopScan()
			return m, func() tea.Msg { return scanCompleteMsg{} }
		}
	}
	return m, nil
}

func (m *Model) handleControlKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg.String() {
	case "q", "ctrl+c":
		m.cancel()
		m.client.Disconnect()
		return m, tea.Quit

	case "esc", "backspace":
		m.client.Disconnect()
		m.mode = ModeDeviceList
		m.devices = make([]sony_remote_ble.DeviceInfo, 0)
		m.addLog("Disconnected")
		return m, nil

	// Focus controls
	case "f", "F":
		m.buttonStates["focus"] = true
		cmds = append(cmds, m.sendCommand("focus_down"))

	// Shutter controls
	case "s", "S":
		m.buttonStates["shutter"] = true
		cmds = append(cmds, m.sendCommand("shutter_full_down"))

	// Zoom controls
	case "z":
		m.buttonStates["zoom_out"] = true
		cmds = append(cmds, m.sendCommand("zoom_out_down"))

	case "Z":
		m.buttonStates["zoom_in"] = true
		cmds = append(cmds, m.sendCommand("zoom_in_down"))

	// AutoFocus
	case "a", "A":
		m.buttonStates["autofocus"] = true
		cmds = append(cmds, m.sendCommand("autofocus_down"))

	// Record
	case "r", "R":
		m.buttonStates["record"] = true
		cmds = append(cmds, m.sendCommand("record_toggle"))

	// Custom button
	case "c", "C":
		m.buttonStates["custom"] = true
		cmds = append(cmds, m.sendCommand("c1_down"))

	// Quick photo
	case " ":
		m.buttonStates["shutter"] = true
		m.addLog("Taking photo...")
		cmds = append(cmds, m.takePhoto())
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	switch m.mode {
	case ModeDeviceList:
		return m.deviceListView()
	case ModeControl:
		return m.controlView()
	}
	return ""
}

func (m *Model) addLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	m.logs = append(m.logs, fmt.Sprintf("[%s] %s", timestamp, message))
	if len(m.logs) > 10 {
		m.logs = m.logs[1:]
	}
}

func (m *Model) addDevice(device sony_remote_ble.DeviceInfo) {
	// Check if device already exists
	for _, d := range m.devices {
		if d.AddressStr == device.AddressStr {
			return
		}
	}
	m.devices = append(m.devices, device)
	m.addLog(fmt.Sprintf("Found: %s (%s)", device.Name, device.AddressStr))
}

// Command functions
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) checkForDevicesCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		select {
		case device := <-m.deviceChan:
			return deviceFoundMsg(device)
		default:
			return nil // Don't return anything if no device found
		}
	})
}

func (m *Model) performScan() tea.Cmd {
	return func() tea.Msg {
		err := m.client.ScanForDevices(m.ctx, m.deviceChan)
		if err != nil {
			return scanCompleteMsg{} // End scan on error
		}
		return nil // Continue scanning
	}
}

func (m *Model) connect(device sony_remote_ble.DeviceInfo) tea.Cmd {
	return func() tea.Msg {
		err := m.client.Connect(device.Address)
		return connectionMsg{
			connected: err == nil,
			err:       err,
		}
	}
}

func (m *Model) sendCommand(command string) tea.Cmd {
	return func() tea.Msg {
		cmd, exists := sony_remote_ble.Commands[command]
		if !exists {
			return commandSentMsg{
				command: command,
				err:     fmt.Errorf("unknown command: %s", command),
			}
		}

		err := m.client.SendCommand(cmd)
		return commandSentMsg{
			command: cmd.Name,
			err:     err,
		}
	}
}

func (m *Model) takePhoto() tea.Cmd {
	return func() tea.Msg {
		err := m.client.TakePhoto()
		return commandSentMsg{
			command: "Take Photo",
			err:     err,
		}
	}
}

func (m *Model) getLastLogs(n int) []string {
	if len(m.logs) <= n {
		return m.logs
	}
	return m.logs[len(m.logs)-n:]
}
