package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Settings struct {
	Colors struct {
		ActivePaneBorder   string `yaml:"active_pane_border"`
		InactivePaneBorder string `yaml:"inactive_pane_border"`
		HeaderBackground   string `yaml:"header_background"`
		HeaderText         string `yaml:"header_text"`
		FooterBackground   string `yaml:"footer_background"`
		FooterText         string `yaml:"footer_text"`
	} `yaml:"colors"`
}

var settings Settings

func loadSettings() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("Warning: Could not load config.yaml, using defaults: %v\n", err)
		settings.Colors.ActivePaneBorder = "#FF00FF"
		settings.Colors.InactivePaneBorder = "#874BFD"
		settings.Colors.HeaderBackground = "#7D56F4"
		settings.Colors.HeaderText = "#FAFAFA"
		settings.Colors.FooterBackground = "#3C3C3C"
		settings.Colors.FooterText = "#FAFAFA"
		return
	}

	if err := yaml.Unmarshal(data, &settings); err != nil {
		fmt.Printf("Warning: Could not parse config.yaml: %v\n", err)
	}
}
