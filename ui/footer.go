package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// FooterItem represents a single footer menu item
type FooterItem struct {
	Text     string
	Required bool // if true, item won't be dropped when space is limited
}

// RenderFooter renders a footer with the given content and width
func RenderFooter(content string, width int) string {
	return FooterStyle.Width(width).Render(content)
}

// RenderPromptFooter renders a footer with a prompt and input view
func RenderPromptFooter(prompt, inputView string, width int, promptColor string) string {
	styledPrompt := lipgloss.NewStyle().
		Foreground(lipgloss.Color(promptColor)).
		Bold(true).
		Render(prompt)
	return FooterStyle.Width(width).Render(styledPrompt + inputView)
}

// BuildAdaptiveMenu builds a menu string that adapts to available width
// It will drop optional items if they don't fit, but always keeps required items
func BuildAdaptiveMenu(items []FooterItem, maxWidth int, separator string) string {
	text := ""
	for _, item := range items {
		testText := text
		if testText != "" {
			testText += separator
		}
		testText += item.Text

		// Check if adding this item would exceed width
		if len(testText) > maxWidth {
			if !item.Required {
				continue
			}
		}

		if text != "" {
			text += separator
		}
		text += item.Text
	}

	return text
}
