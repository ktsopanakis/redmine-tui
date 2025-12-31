package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Settings struct {
	Redmine struct {
		URL    string `yaml:"url"`
		APIKey string `yaml:"api_key"`
	} `yaml:"redmine"`
	Colors struct {
		ActivePaneBorder   string `yaml:"active_pane_border"`
		InactivePaneBorder string `yaml:"inactive_pane_border"`
		HeaderBackground   string `yaml:"header_background"`
		HeaderText         string `yaml:"header_text"`
		FooterBackground   string `yaml:"footer_background"`
		FooterText         string `yaml:"footer_text"`
	} `yaml:"colors"`
}

var Current Settings

// GetConfigPath returns the path to the config file in the user's home directory
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "redmine-tui", "config.yaml"), nil
}

// ensureConfigDir creates the config directory if it doesn't exist
func ensureConfigDir() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0755)
}

func Load() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("could not determine config path: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return special error to indicate first run
			return fmt.Errorf("config file does not exist: first run")
		}
		return fmt.Errorf("could not load config: %w", err)
	}

	if err := yaml.Unmarshal(data, &Current); err != nil {
		return fmt.Errorf("could not parse config: %w", err)
	}

	return nil
}

func saveSettings() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(&Current)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// PromptForRedmineSetup interactively asks for Redmine URL and API key
func PromptForRedmineSetup() error {
	// Set default colors first
	Current.Colors.ActivePaneBorder = "#FF00FF"
	Current.Colors.InactivePaneBorder = "#874BFD"
	Current.Colors.HeaderBackground = "#7D56F4"
	Current.Colors.HeaderText = "#FAFAFA"
	Current.Colors.FooterBackground = "#3C3C3C"
	Current.Colors.FooterText = "#FAFAFA"

	fmt.Println("\n=== Redmine TUI Setup ===")
	fmt.Println("Please enter your Redmine configuration:")

	fmt.Print("\nRedmine URL (e.g., https://redmine.example.com): ")
	fmt.Scanln(&Current.Redmine.URL)

	fmt.Print("API Key: ")
	fmt.Scanln(&Current.Redmine.APIKey)

	if Current.Redmine.URL == "" || Current.Redmine.APIKey == "" {
		return fmt.Errorf("URL and API Key are required")
	}

	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	if err := saveSettings(); err != nil {
		return fmt.Errorf("could not save config: %w", err)
	}

	configPath, _ := GetConfigPath()
	fmt.Printf("\nConfiguration saved to: %s\n", configPath)
	return nil
}
