package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/smazurov/sony_remote_ble/internal/ui"
)

// version is set via ldflags during build
var version = "dev"

func main() {
	// Initialize the model
	model, err := ui.NewModel(version)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Create the Bubble Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
