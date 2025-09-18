// Package sony_remote_ble provides a Bluetooth Low Energy (BLE) client for controlling Sony cameras.
// This package implements the reverse-engineered Sony camera remote protocol, allowing you to
// trigger camera functions like focus, shutter, zoom, and recording via Bluetooth.
//
// The package supports various Sony camera models that implement the Sony camera remote BLE service.
// It provides both low-level command sending capabilities and high-level convenience functions.
//
// Example usage:
//
//	client, err := sony_remote_ble.NewClient()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Disconnect()
//
//	// Scan for cameras
//	ctx := context.Background()
//	deviceChan := make(chan sony_remote_ble.DeviceInfo, 10)
//	err = client.ScanForDevices(ctx, deviceChan)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Connect to first found device
//	device := <-deviceChan
//	err = client.Connect(device.Address)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Take a photo
//	err = client.TakePhoto()
//	if err != nil {
//		log.Fatal(err)
//	}
package sony_remote_ble

import "tinygo.org/x/bluetooth"

const (
	// SonyServiceUUID is the Bluetooth service UUID for Sony camera remote control.
	// This UUID identifies the BLE service that Sony cameras expose for remote control functionality.
	SonyServiceUUID = "8000ff00-ff00-ffff-ffff-ffffffffffff"

	// CommandCharUUID is the characteristic UUID for sending commands to Sony cameras.
	// Commands are written to this characteristic to trigger camera functions.
	CommandCharUUID = "0000ff01-0000-1000-8000-00805f9b34fb"
)

// SonyCommand represents a camera command that can be sent to a Sony camera.
// Each command has a human-readable name and the corresponding byte sequence
// that triggers the action when sent to the camera's command characteristic.
type SonyCommand struct {
	// Name is a human-readable description of what this command does
	Name string
	// Code is the byte sequence that triggers this command when sent to the camera
	Code []byte
}

// Commands is a map of all available camera commands indexed by their string identifiers.
// These commands are based on reverse engineering of the Sony camera remote protocol.
// Each command consists of a descriptive name and the byte sequence that triggers the action.
//
// Usage example:
//
//	cmd := sony_remote_ble.Commands["shutter_full_down"]
//	err := client.SendCommand(cmd)
var Commands = map[string]SonyCommand{
	// Focus commands - Control manual focus ring
	"focus_down":     {"Focus Down", []byte{0x01, 0x07}},        // Move focus towards infinity
	"focus_up":       {"Focus Up", []byte{0x01, 0x06}},          // Move focus towards near
	"autofocus_down": {"AutoFocus Down", []byte{0x01, 0x15}},    // Trigger autofocus start
	"autofocus_up":   {"AutoFocus Up", []byte{0x01, 0x14}},      // Release autofocus

	// Shutter commands - Control camera shutter
	"shutter_half_down": {"Shutter Half Down", []byte{0x01, 0x07}}, // Half-press shutter (focus)
	"shutter_half_up":   {"Shutter Half Up", []byte{0x01, 0x06}},   // Release half-press
	"shutter_full_down": {"Shutter Full Down", []byte{0x01, 0x09}}, // Full shutter press (take photo)
	"shutter_full_up":   {"Shutter Full Up", []byte{0x01, 0x08}},   // Release full press

	// Record commands - Control video recording
	"record_toggle": {"Toggle Record", []byte{0x01, 0x0e}}, // Start/stop video recording
	"record_down":   {"Record Down", []byte{0x01, 0x0f}},   // Press record button

	// Zoom commands - Control lens zoom (if supported)
	"zoom_in_down":  {"Zoom In Down", []byte{0x02, 0x6d, 0x20}},  // Start zooming in
	"zoom_in_up":    {"Zoom In Up", []byte{0x02, 0x6c, 0x00}},    // Stop zooming in
	"zoom_out_down": {"Zoom Out Down", []byte{0x02, 0x6b, 0x20}}, // Start zooming out
	"zoom_out_up":   {"Zoom Out Up", []byte{0x02, 0x6a, 0x00}},   // Stop zooming out

	// Custom button commands - Trigger custom function buttons
	"c1_down": {"C1 Down", []byte{0x01, 0x21}}, // Press custom button C1
	"c1_up":   {"C1 Up", []byte{0x01, 0x20}},   // Release custom button C1
}

// TakePhotoSequence returns a sequence of commands that performs a complete photo capture.
// This sequence focuses the camera and then takes a photo, similar to a full shutter press
// on a physical camera.
//
// The sequence:
//  1. Start focus
//  2. Take photo (full shutter press)
//  3. Release shutter
//  4. Release focus
//
// Example usage:
//
//	sequence := sony_remote_ble.TakePhotoSequence()
//	for _, cmd := range sequence {
//		err := client.SendCommand(cmd)
//		if err != nil {
//			log.Printf("Command failed: %v", err)
//			break
//		}
//		time.Sleep(100 * time.Millisecond) // Small delay between commands
//	}
func TakePhotoSequence() []SonyCommand {
	return []SonyCommand{
		Commands["focus_down"],
		Commands["shutter_full_down"],
		Commands["shutter_full_up"],
		Commands["focus_up"],
	}
}

// ServiceUUID returns the parsed Bluetooth service UUID for Sony camera remote control.
// This UUID is used to identify and connect to the camera's remote control service.
func ServiceUUID() bluetooth.UUID {
	uuid, _ := bluetooth.ParseUUID(SonyServiceUUID)
	return uuid
}

// CharacteristicUUID returns the parsed Bluetooth characteristic UUID for sending commands.
// Commands are written to this characteristic to control the camera.
func CharacteristicUUID() bluetooth.UUID {
	uuid, _ := bluetooth.ParseUUID(CommandCharUUID)
	return uuid
}