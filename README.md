# Sony Camera Remote (BLE)

A terminal-based remote control for Sony cameras using Bluetooth Low Energy (BLE). Control your Sony camera wirelessly with focus, shutter, zoom, and recording commands through an intuitive text user interface.

## Features

- **Device Discovery**: Automatically scan for nearby Sony cameras
- **Real-time Control**: Focus, shutter, zoom, autofocus, and recording controls
- **Cross-platform**: Works on Linux and macOS
- **TUI Interface**: Clean terminal interface with visual feedback
- **Library Support**: Can be used as a Go library for custom applications

## Supported Cameras

This tool works with Sony cameras that support Bluetooth Low Energy remote control, including:
- Sony Alpha series (ILCE models)
- Sony FX series
- Sony Cyber-shot series (DSC models)

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/smazurov/sony_remote_ble/releases).

### From Source

```bash
git clone https://github.com/smazurov/sony_remote_ble.git
cd sony_remote_ble
go build -o sony-remote
```

## Setup and Pairing

**Important**: Your Sony camera must be paired with your computer through system Bluetooth settings before using this tool.

### Pairing Steps

1. **Enable Bluetooth on your camera**:
   - Go to camera's Network menu
   - Enable Bluetooth function
   - Set to "Remote Control" mode

2. **Pair via system settings**:
   - **Linux**: Use `bluetoothctl` or your desktop's Bluetooth manager
   - **macOS**: System Preferences → Bluetooth
   - Look for your camera model (e.g., "ILCE-7M4", "FX30")
   - Complete the pairing process

3. **Run the remote**:
   ```bash
   ./sony-remote
   ```

## Usage

### Terminal Interface

1. **Launch**: Run `./sony-remote`
2. **Scan**: Press `Tab` to scan for paired Sony cameras
3. **Select**: Use `↑/↓` or `k/j` to navigate devices
4. **Connect**: Press `Enter` to connect to selected camera
5. **Control**: Use keyboard shortcuts to control your camera

### Camera Controls

Once connected, use these keyboard shortcuts:

- **F/f** - Focus control
- **S/s** - Shutter control
- **Z/z** - Zoom out/in
- **A** - Autofocus
- **Space** - Quick shot (take photo)
- **R** - Toggle recording
- **C** - Custom button (C1)
- **Esc** - Back to device list
- **Q** - Quit application

### Troubleshooting

**Camera not appearing in scan?**
- Ensure camera is paired via system Bluetooth settings
- Check camera's Bluetooth is enabled and in "Remote Control" mode
- Try moving closer to the camera

**Connection fails?**
- Verify the camera is paired (not just discovered)
- Restart both camera and computer's Bluetooth
- Some cameras require re-pairing after being turned off

**Commands not working?**
- Camera must be in appropriate mode (not in menu, sleep, etc.)
- Some commands only work in specific camera modes
- Check camera's remote control settings

## Library Usage

This project can be used as a Go library for building custom Sony camera control applications:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/smazurov/sony_remote_ble/sony_remote_ble"
)

func main() {
    // Create client
    client, err := sony_remote_ble.NewClient()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect()

    // Scan for devices
    ctx := context.Background()
    deviceChan := make(chan sony_remote_ble.DeviceInfo, 10)

    err = client.ScanForDevices(ctx, deviceChan)
    if err != nil {
        log.Fatal(err)
    }

    // Wait for a device
    device := <-deviceChan
    fmt.Printf("Found: %s\n", device.Name)

    // Connect and take photo
    err = client.Connect(device.Address)
    if err != nil {
        log.Fatal(err)
    }

    err = client.TakePhoto()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Photo taken!")
}
```

### Available Commands

The library provides these camera commands:

- `focus_down` / `focus_up` - Focus control
- `shutter_half_down` / `shutter_half_up` - Half-press shutter
- `shutter_full_down` / `shutter_full_up` - Full shutter press
- `zoom_in_down` / `zoom_in_up` - Zoom controls
- `zoom_out_down` / `zoom_out_up` - Zoom controls
- `autofocus_down` / `autofocus_up` - Autofocus
- `record_toggle` - Start/stop recording
- `c1_down` / `c1_up` - Custom button

### High-level Methods

- `TakePhoto()` - Complete photo capture sequence
- `SendCommand(cmd SonyCommand)` - Send individual command
- `SendCommandSequence(cmds []SonyCommand, delay time.Duration)` - Send command sequence

## Platform Notes

### Linux
- Requires BlueZ bluetooth stack
- Uses MAC addresses for device identification
- May require sudo for bluetooth access depending on system configuration

### macOS
- Uses Core Bluetooth framework
- Uses UUID for device identification
- Requires location permissions for BLE scanning

## Technical Details

- **Protocol**: Bluetooth Low Energy (BLE/GATT)
- **Service UUID**: `8000ff00-ff00-ffff-ffff-ffffffffffff`
- **Characteristic UUID**: `8001ff00-ff00-ffff-ffff-ffffffffffff`
- **Go Version**: 1.21+
- **Dependencies**: `tinygo.org/x/bluetooth`, `github.com/charmbracelet/bubbletea`

## Building

### Local Build
```bash
go build -ldflags "-s -w -X main.version=v0.0.1" -o sony-remote
```

### Cross-platform Builds
The project uses GitHub Actions for automated cross-platform builds. See `.github/workflows/release.yml` for the complete build matrix.

## Limitations

- Cannot detect pairing status during scanning - cameras must be pre-paired
- Some advanced camera features may not be available via BLE
- Connection stability depends on Bluetooth hardware and distance
- Platform-specific Bluetooth address formats require different handling

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

- Reverse engineering research by [Greg Leed](https://gregleeds.com/reverse-engineering-sony-camera-bluetooth/)
- Built with [tinygo.org/x/bluetooth](https://tinygo.org/x/bluetooth) for cross-platform BLE support
- Terminal UI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)