# Sony Camera Bluetooth Remote

A Terminal User Interface (TUI) application for controlling Sony cameras via Bluetooth Low Energy (BLE) using Go.

## Features

- **Device Discovery**: Scan and connect to Sony cameras automatically
- **Camera Controls**:
  - Focus (manual and auto)
  - Shutter (half and full press)
  - Zoom (in/out)
  - Record toggle
  - Custom button (C1)
  - Quick photo capture
- **Real-time Feedback**: Visual button states and command logging
- **Cross-platform**: Supports Linux, macOS, and Windows
- **ARM Support**: Designed for compilation on ARM64 devices

## Requirements

- Go 1.21 or later
- Bluetooth Low Energy adapter
- Sony camera with Bluetooth support (tested with α7 series)

## Supported Sony Cameras

This application should work with Sony cameras that support Bluetooth remote control, including:
- Sony α7 series (α7, α7R, α7S, α7 II, α7R II, α7S II, α7 III, α7R III, α7S III, α7 IV, α7R IV, α7R V)
- Sony α6000 series
- Sony FX series
- Many Sony Cyber-shot models with Bluetooth

## Installation

### Building from source

```bash
git clone <repository-url>
cd videonode_ble
go mod tidy
go build -o sony-camera-remote
```

### Cross-compilation for ARM64

```bash
GOOS=linux GOARCH=arm64 go build -o sony-camera-remote-arm64
```

## Usage

1. **Pair your Sony camera** with your device:
   - Enable Bluetooth on your camera (Camera Settings → Network → Bluetooth)
   - Put camera in pairing mode (usually under Bluetooth settings)
   - The camera must be paired before it will accept control commands
2. **Run the application**:
   ```bash
   ./sony-camera-remote
   ```
3. **Scan for devices** by pressing `Tab`
4. **Select and connect** to your camera using arrow keys and Enter
5. **Control your camera** using the keyboard shortcuts

## Controls

### Device List Mode
- `↑/↓` or `k/j` - Navigate devices
- `Tab` - Scan for devices
- `Enter` - Connect to selected device
- `Esc` - Stop scanning
- `q` - Quit application

### Camera Control Mode
- `F/f` - Focus control
- `S/s` - Shutter control
- `Z/z` - Zoom out/in
- `A/a` - AutoFocus
- `R/r` - Record toggle
- `C/c` - Custom button (C1)
- `Space` - Quick photo (focus + shutter sequence)
- `Esc` - Return to device list
- `q` - Quit application

## Protocol Details

This application uses the Sony Camera Bluetooth Low Energy protocol:

- **Service UUID**: `8000ff00-ff00-ffff-ffff-ffffffffffff`
- **Characteristic UUID**: `0000ff01-0000-1000-8000-00805f9b34fb`

Commands are sent as byte arrays to control different camera functions.

## Troubleshooting

### Connection Issues
- Ensure your camera is paired via the camera's Bluetooth menu (not OS settings)
- Make sure the camera's Bluetooth is enabled and device is still in pairing list
- Sony cameras reject commands from unpaired devices and will disconnect immediately
- Try restarting both the camera and the application

### No Devices Found
- Check that your Bluetooth adapter is working
- Ensure the camera is nearby and in pairing mode
- Try scanning multiple times

### Commands Not Working
- Verify the camera is fully connected (green status indicator)
- Some cameras may require specific modes to be enabled
- Check the log output for error messages

## Development

### Project Structure
```
videonode_ble/
├── main.go                 # Application entry point
├── internal/
│   ├── bluetooth/          # BLE communication
│   │   ├── client.go       # Connection management
│   │   └── commands.go     # Sony camera commands
│   └── ui/                 # TUI components
│       ├── model.go        # Application state
│       ├── view.go         # UI rendering
│       └── styles.go       # Visual styling
└── README.md
```

### Dependencies
- [TinyGo Bluetooth](https://github.com/tinygo-org/bluetooth) - BLE communication
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## Acknowledgments

- Sony camera protocol reverse engineering by [Greg Leeds](https://gregleeds.com/reverse-engineering-sony-camera-bluetooth/)
- Additional protocol details from [freemote project](https://github.com/tao-j/freemote)