package main

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle     lipgloss.Style
	footerStyle     lipgloss.Style
	paneStyle       lipgloss.Style
	activePaneStyle lipgloss.Style
)

func initStyles() {
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(settings.Colors.HeaderText)).
		Background(lipgloss.Color(settings.Colors.HeaderBackground)).
		PaddingLeft(1)

	footerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(settings.Colors.FooterText)).
		Background(lipgloss.Color(settings.Colors.FooterBackground)).
		PaddingLeft(1)

	paneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(settings.Colors.InactivePaneBorder)).
		Padding(0, 1)

	activePaneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(settings.Colors.ActivePaneBorder)).
		Padding(0, 1)
}
