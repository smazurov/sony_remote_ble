package sony_remote_ble

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

// ConnectionState represents the current state of the Bluetooth connection to a Sony camera.
type ConnectionState int

const (
	// Disconnected indicates no active connection to a camera
	Disconnected ConnectionState = iota
	// Scanning indicates the client is actively scanning for nearby Sony cameras
	Scanning
	// Connecting indicates a connection attempt is in progress
	Connecting
	// Connected indicates an active connection to a camera with command capability
	Connected
	// Error indicates an error state that requires attention
	Error
)

// String returns a human-readable representation of the connection state.
func (cs ConnectionState) String() string {
	switch cs {
	case Disconnected:
		return "Disconnected"
	case Scanning:
		return "Scanning"
	case Connecting:
		return "Connecting"
	case Connected:
		return "Connected"
	case Error:
		return "Error"
	default:
		return "Unknown"
	}
}

// Client provides a high-level interface for connecting to and controlling Sony cameras via Bluetooth Low Energy.
// The client handles device discovery, connection management, and command transmission.
//
// Example usage:
//
//	client, err := sony_remote_ble.NewClient()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Disconnect()
//
//	// Connect to a specific camera
//	err = client.Connect("AA:BB:CC:DD:EE:FF")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Take a photo
//	err = client.TakePhoto()
//	if err != nil {
//		log.Fatal(err)
//	}
type Client struct {
	adapter     *bluetooth.Adapter
	device      bluetooth.Device
	service     bluetooth.DeviceService
	char        bluetooth.DeviceCharacteristic
	state       ConnectionState
	deviceName  string
	lastError   error
	stopScan    chan bool
}

// DeviceInfo contains information about a discovered Sony camera device.
// This information is provided during device scanning and can be used to
// identify and connect to specific cameras.
type DeviceInfo struct {
	// Name is the advertised device name (e.g., "ILCE-7M4", "FX30")
	Name string
	// Address is the platform-specific bluetooth address object
	Address bluetooth.Address
	// AddressStr is the string representation of the address for display
	AddressStr string
	// RSSI is the received signal strength indicator in dBm (typically -30 to -100)
	RSSI int16
}

// NewClient creates a new Sony camera BLE client and initializes the Bluetooth adapter.
// The client is ready to scan for devices and establish connections after creation.
//
// Returns an error if the Bluetooth adapter cannot be enabled or is not available.
//
// Example:
//
//	client, err := sony_remote_ble.NewClient()
//	if err != nil {
//		log.Fatal("Failed to create client:", err)
//	}
//	defer client.Disconnect()
func NewClient() (*Client, error) {
	adapter := bluetooth.DefaultAdapter
	err := adapter.Enable()
	if err != nil {
		return nil, fmt.Errorf("failed to enable adapter: %w", err)
	}

	return &Client{
		adapter:  adapter,
		state:    Disconnected,
		stopScan: make(chan bool, 1),
	}, nil
}

// State returns the current connection state of the client.
func (c *Client) State() ConnectionState {
	return c.state
}

// DeviceName returns the name of the currently connected device.
// Returns an empty string if not connected to any device.
func (c *Client) DeviceName() string {
	return c.deviceName
}

// LastError returns the last error that occurred during client operations.
// Returns nil if no error has occurred or if the error has been cleared.
func (c *Client) LastError() error {
	return c.lastError
}

// ScanForDevices starts scanning for nearby Sony cameras and sends discovered devices
// to the provided channel. The scan runs asynchronously until stopped with StopScan()
// or until the context is cancelled.
//
// The function filters devices to only include those that appear to be Sony cameras
// based on their advertised names. Found devices are sent to deviceChan as DeviceInfo structs.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - deviceChan: Channel to receive discovered devices (should be buffered to prevent blocking)
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	deviceChan := make(chan sony_remote_ble.DeviceInfo, 10)
//	err := client.ScanForDevices(ctx, deviceChan)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Wait for devices
//	select {
//	case device := <-deviceChan:
//		fmt.Printf("Found camera: %s (%s)\n", device.Name, device.Address)
//	case <-ctx.Done():
//		fmt.Println("Scan timeout")
//	}
func (c *Client) ScanForDevices(ctx context.Context, deviceChan chan<- DeviceInfo) error {
	c.state = Scanning
	c.lastError = nil

	go func() {
		for {
			select {
			case <-c.stopScan:
				return
			case <-ctx.Done():
				return
			default:
			}

			err := c.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
				select {
				case <-c.stopScan:
					return
				case <-ctx.Done():
					return
				default:
					// Look for Sony cameras (they typically advertise with specific names)
					name := result.LocalName()
					if name == "" {
						return
					}

					// Check if this might be a Sony camera
					if containsSonyIdentifier(name) {
						deviceChan <- DeviceInfo{
							Name:       name,
							Address:    result.Address,
							AddressStr: result.Address.String(),
							RSSI:       result.RSSI,
						}
					}
				}
			})

			if err != nil {
				c.lastError = err
				c.state = Error
				return
			}
			// If adapter.Scan() returns without error, restart it
		}
	}()

	return nil
}

// StopScan stops the active device scanning process.
// This method is safe to call multiple times and from different goroutines.
// After stopping the scan, the client state returns to Disconnected (unless already connected).
func (c *Client) StopScan() {
	select {
	case c.stopScan <- true:
	default:
	}
	c.adapter.StopScan()
	if c.state == Scanning {
		c.state = Disconnected
	}
}

// Connect establishes a connection to a Sony camera using the provided Bluetooth address.
// The function performs the complete connection sequence including service and characteristic discovery.
//
// The address should be obtained from a DeviceInfo struct during device scanning.
// After successful connection, the client will be ready to send commands to the camera.
//
// Parameters:
//   - address: Bluetooth address object from device scanning
//
// Returns an error if:
//   - The connection cannot be established
//   - The Sony camera service is not found on the device
//   - The command characteristic is not available
//
// Example:
//
//	// Get device from scanning
//	device := <-deviceChan
//	err := client.Connect(device.Address)
//	if err != nil {
//		log.Fatal("Connection failed:", err)
//	}
//	fmt.Println("Connected to camera successfully")
func (c *Client) Connect(address bluetooth.Address) error {
	c.state = Connecting
	c.lastError = nil

	// Connect to device
	device, err := c.adapter.Connect(address, bluetooth.ConnectionParams{})
	if err != nil {
		c.lastError = fmt.Errorf("failed to connect: %w", err)
		c.state = Error
		return c.lastError
	}

	c.device = device

	// Discover services
	services, err := device.DiscoverServices([]bluetooth.UUID{ServiceUUID()})
	if err != nil {
		c.lastError = fmt.Errorf("failed to discover services: %w", err)
		c.state = Error
		return c.lastError
	}

	if len(services) == 0 {
		c.lastError = errors.New("Sony camera service not found")
		c.state = Error
		return c.lastError
	}

	c.service = services[0]

	// Discover characteristics
	chars, err := c.service.DiscoverCharacteristics([]bluetooth.UUID{CharacteristicUUID()})
	if err != nil {
		c.lastError = fmt.Errorf("failed to discover characteristics: %w", err)
		c.state = Error
		return c.lastError
	}

	if len(chars) == 0 {
		c.lastError = errors.New("command characteristic not found")
		c.state = Error
		return c.lastError
	}

	c.char = chars[0]
	c.state = Connected
	c.deviceName = address.String() // Could be enhanced to get actual device name

	return nil
}

// Disconnect terminates the connection to the currently connected Sony camera.
// This method is safe to call even if not currently connected.
// After disconnection, the client can be used to connect to the same or different camera.
//
// Example:
//
//	defer client.Disconnect() // Ensure cleanup
//
//	err := client.Disconnect()
//	if err != nil {
//		log.Printf("Disconnect error: %v", err)
//	}
func (c *Client) Disconnect() error {
	if c.state == Connected {
		err := c.device.Disconnect()
		if err != nil {
			c.lastError = err
			c.state = Error
			return err
		}
	}
	c.state = Disconnected
	c.deviceName = ""
	return nil
}

// SendCommand sends a low-level command to the connected Sony camera.
// The client must be in Connected state for this method to succeed.
//
// Commands are typically obtained from the Commands map or created manually.
// This is the fundamental method for all camera control operations.
//
// Parameters:
//   - cmd: The SonyCommand to send, containing both name and byte code
//
// Returns an error if not connected or if the command transmission fails.
//
// Example:
//
//	// Send a focus command
//	cmd := sony_remote_ble.Commands["focus_down"]
//	err := client.SendCommand(cmd)
//	if err != nil {
//		log.Printf("Command failed: %v", err)
//	}
//
//	// Send custom command
//	customCmd := sony_remote_ble.SonyCommand{
//		Name: "Custom Action",
//		Code: []byte{0x01, 0x42},
//	}
//	err = client.SendCommand(customCmd)
func (c *Client) SendCommand(cmd SonyCommand) error {
	if c.state != Connected {
		return errors.New("not connected to device")
	}

	_, err := c.char.WriteWithoutResponse(cmd.Code)
	if err != nil {
		c.lastError = fmt.Errorf("failed to send command %s: %w", cmd.Name, err)
		return c.lastError
	}

	return nil
}

// SendCommandSequence sends a series of commands to the camera with optional delays between them.
// This is useful for complex operations that require multiple commands in sequence.
//
// Parameters:
//   - commands: Slice of SonyCommand structs to send in order
//   - delay: Duration to wait between each command (use 0 for no delay)
//
// The function stops and returns an error if any command in the sequence fails.
//
// Example:
//
//	// Create a custom sequence
//	sequence := []sony_remote_ble.SonyCommand{
//		sony_remote_ble.Commands["focus_down"],
//		sony_remote_ble.Commands["shutter_full_down"],
//		sony_remote_ble.Commands["shutter_full_up"],
//		sony_remote_ble.Commands["focus_up"],
//	}
//
//	err := client.SendCommandSequence(sequence, 100*time.Millisecond)
//	if err != nil {
//		log.Printf("Sequence failed: %v", err)
//	}
func (c *Client) SendCommandSequence(commands []SonyCommand, delay time.Duration) error {
	for _, cmd := range commands {
		if err := c.SendCommand(cmd); err != nil {
			return err
		}
		if delay > 0 {
			time.Sleep(delay)
		}
	}
	return nil
}

// TakePhoto is a high-level convenience method that captures a photo using the camera.
// This method sends the complete photo capture sequence: focus, capture, and release.
//
// It's equivalent to calling SendCommandSequence with TakePhotoSequence() and a 50ms delay.
// The delay between commands ensures proper camera response timing.
//
// Returns an error if not connected or if any part of the photo sequence fails.
//
// Example:
//
//	err := client.TakePhoto()
//	if err != nil {
//		log.Printf("Photo capture failed: %v", err)
//	} else {
//		fmt.Println("Photo captured successfully!")
//	}
func (c *Client) TakePhoto() error {
	return c.SendCommandSequence(TakePhotoSequence(), 50*time.Millisecond)
}

// Helper function to identify Sony cameras
func containsSonyIdentifier(name string) bool {
	sonyIdentifiers := []string{
		"Sony",
		"ILCE",  // Sony Alpha series
		"DSC",   // Sony Cyber-shot series
		"FX",    // Sony FX series
		"α",     // Alpha symbol
		"Alpha",
	}

	for _, identifier := range sonyIdentifiers {
		if len(name) >= len(identifier) {
			for i := 0; i <= len(name)-len(identifier); i++ {
				if name[i:i+len(identifier)] == identifier {
					return true
				}
			}
		}
	}
	return false
}