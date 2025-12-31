package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	loadSettings()
	initStyles()

	// Parse command-line flags
	altScreen := flag.Bool("alt-screen", false, "Use alternate screen buffer (clears on exit)")
	flag.Parse()

	// Build program options
	opts := []tea.ProgramOption{tea.WithMouseCellMotion()}
	if *altScreen {
		opts = append(opts, tea.WithAltScreen())
	}

	p := tea.NewProgram(
		initialModel(),
		opts...,
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
