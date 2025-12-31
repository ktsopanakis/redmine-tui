package app

import (
	appui "github.com/ktsopanakis/redmine-tui/ui"
)

// getHelpContent returns the help modal content
func (m Model) getHelpContent() []string {
	return []string{
		"Navigation:",
		"  ↑/k, ↓/j       - Move up/down in lists",
		"  PgUp/PgDn      - Page up/down",
		"  Tab            - Switch between panes",
		"  Home/End       - Go to first/last item",
		"",
		"Filtering & Views:",
		"  f              - Toggle filter mode (filter issues by text)",
		"  m              - Toggle between My Issues/All Issues",
		"  u              - Select users to filter by",
		"  p              - Select projects to filter by",
		"",
		"Issue Management:",
		"  e              - Enter edit mode (modify issue fields)",
		"  Enter          - When editing: save changes",
		"  Space          - When in selection list: toggle item",
		"",
		"Edit Mode:",
		"  ↑/k, ↓/j       - Navigate between fields",
		"  Enter          - Edit selected field",
		"  Ctrl+S         - Save all changes",
		"  Tab            - Switch to next field",
		"",
		"General:",
		"  ?              - Show this help",
		"  Esc            - Cancel/close current action",
		"  q              - Quit application",
		"",
		"Tips:",
		"  - Selected users/projects appear at the top of lists",
		"  - Use filter in selection lists to quickly find items",
		"  - Unsaved changes show a red border on the details pane",
		"  - Press Esc to discard changes in edit mode",
	}
}

// renderHelpModal renders the help modal
func (m Model) renderHelpModal() string {
	cfg := appui.ModalConfig{
		Title:       "Redmine TUI - Help",
		Content:     m.getHelpContent(),
		Width:       m.width,
		Height:      m.height,
		BorderColor: "#61AFEF",
		TitleColor:  "#FFFFFF",
	}
	return appui.RenderModal(cfg)
}
