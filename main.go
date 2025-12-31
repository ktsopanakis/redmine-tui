package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ktsopanakis/redmine-tui/config"
)

func main() {
	// Parse command-line flags
	altScreen := flag.Bool("alt-screen", false, "Use alternate screen buffer (clears on exit)")
	setup := flag.Bool("setup", false, "Run interactive setup to configure Redmine URL and API key")
	showConfig := flag.Bool("show-config", false, "Show the location of the config file")
	flag.Parse()

	// Handle --show-config flag
	if *showConfig {
		configPath, err := config.GetConfigPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config file location: %s\n", configPath)
		os.Exit(0)
	}

	// Handle --setup flag
	if *setup {
		if err := config.PromptForRedmineSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\nSetup complete! You can now run redmine-tui to start the application.")
		os.Exit(0)
	}

	// Load settings
	err := config.Load()
	if err != nil && err.Error() == "config file does not exist: first run" {
		// First run - automatically run setup
		fmt.Println("Welcome to Redmine TUI!")
		fmt.Println("No configuration found. Let's set up your Redmine connection.")
		if err := config.PromptForRedmineSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\nSetup complete! Starting Redmine TUI...\n")
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		configPath, _ := config.GetConfigPath()
		fmt.Fprintf(os.Stderr, "Please check your config file at: %s\n", configPath)
		os.Exit(1)
	}

	// Check if Redmine is configured
	if config.Current.Redmine.URL == "" || config.Current.Redmine.APIKey == "" {
		configPath, _ := config.GetConfigPath()
		fmt.Fprintf(os.Stderr, "Redmine URL and API Key are not configured.\n")
		fmt.Fprintf(os.Stderr, "Run 'redmine-tui --setup' to configure, or edit: %s\n", configPath)
		os.Exit(1)
	}

	initStyles()

	// Build program options
	var opts []tea.ProgramOption
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
