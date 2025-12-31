package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/ktsopanakis/redmine-tui/config"
)

// HeaderSection represents a section of the header with color and content
type HeaderSection struct {
	Text  string
	Color string
	Bold  bool
}

// RenderHeader builds and renders a header with left and right sections
// leftSections: array of sections to display on the left
// rightSections: array of sections to display on the right
// width: total width of the header
func RenderHeader(leftSections, rightSections []HeaderSection, width int) string {
	bg := lipgloss.Color(config.Current.Colors.HeaderBackground)

	// Build left side
	leftContent := ""
	leftLen := 0
	for i, section := range leftSections {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(section.Color)).
			Background(bg)
		if section.Bold {
			style = style.Bold(true)
		}

		leftContent += style.Render(section.Text)
		leftLen += len(section.Text)

		// Add space between sections (except after last)
		if i < len(leftSections)-1 {
			leftContent += lipgloss.NewStyle().Background(bg).Render(" ")
			leftLen++
		}
	}

	// Build right side
	rightContent := ""
	rightLen := 0
	for i, section := range rightSections {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(section.Color)).
			Background(bg)
		if section.Bold {
			style = style.Bold(true)
		}

		rightContent += style.Render(section.Text)
		rightLen += len(section.Text)

		// Add space between sections (except after last)
		if i < len(rightSections)-1 {
			rightContent += lipgloss.NewStyle().Background(bg).Render(" ")
			rightLen++
		}
	}

	// Calculate spacer
	spacer := ""
	if width > leftLen+rightLen+1 {
		spacer = lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", width-leftLen-rightLen-1))
	}

	// Combine and render
	return lipgloss.NewStyle().
		Background(bg).
		PaddingLeft(1).
		Width(width).
		Render(leftContent + spacer + rightContent)
}

// FormatDateTime returns formatted day and datetime strings
func FormatDateTime() (dayOfWeek, dateTime string) {
	now := time.Now()
	dayOfWeek = now.Format("Monday")
	dateTime = now.Format("2006-01-02 15:04:05")
	return
}
