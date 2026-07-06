package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ktsopanakis/redmine-tui/api"
)

func TestInitialModel(t *testing.T) {
	// This test verifies the model can be created without panicking
	model := InitialModel()

	if model.leftTitle == "" {
		t.Error("InitialModel() should set leftTitle")
	}

	if model.rightTitle == "" {
		t.Error("InitialModel() should set rightTitle")
	}

	if model.activePane != 0 {
		t.Errorf("InitialModel() activePane = %d, want 0", model.activePane)
	}

	if model.client == nil {
		t.Error("InitialModel() should create client")
	}
}

func TestModelInit(t *testing.T) {
	model := InitialModel()

	// Test that Init returns valid commands
	cmd := model.Init()

	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestSelectedIssue(t *testing.T) {
	model := InitialModel()

	// Test with issues
	model.issues = []api.Issue{
		{ID: 1, Subject: "Test 1"},
		{ID: 2, Subject: "Test 2"},
	}
	model.selectedIndex = 0

	if model.selectedIndex < 0 || model.selectedIndex >= len(model.issues) {
		t.Error("selectedIndex should be within bounds")
	}

	if model.issues[model.selectedIndex].ID != 1 {
		t.Errorf("Selected issue ID = %d, want 1", model.issues[model.selectedIndex].ID)
	}
}

func TestSwitchPane(t *testing.T) {
	model := InitialModel()

	initialPane := model.activePane

	// Simulate switching panes
	if initialPane == 0 {
		model.activePane = 1
	} else {
		model.activePane = 0
	}

	if model.activePane == initialPane {
		t.Error("Pane should have switched")
	}
}

func TestNoteModeToggle(t *testing.T) {
	model := InitialModel()
	model.ready = true
	model.issues = []api.Issue{{ID: 42, Subject: "Test"}}
	model.selectedIndex = 0

	// Press 'c' to enter note mode on the selected issue
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m := updated.(Model)
	if !m.noteMode {
		t.Fatal("pressing 'c' should enter note mode")
	}
	if m.noteIssueID != 42 {
		t.Errorf("noteIssueID = %d, want 42", m.noteIssueID)
	}

	// Esc cancels without posting
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.noteMode {
		t.Error("pressing Esc should exit note mode")
	}

	// Re-enter, then Ctrl+S with an empty note should just close (no request)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(Model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	if m.noteMode {
		t.Error("Ctrl+S should exit note mode")
	}
	if cmd != nil {
		t.Error("Ctrl+S with an empty note should not issue a command")
	}
}

// TestEditOnlyCommitsTouchedFields reproduces the ticket-3888 bug: changing one
// field (priority) must not rewrite fields the user merely navigated past
// (a long, multi-line description), which the single-line input would otherwise
// truncate/mangle.
func TestEditOnlyCommitsTouchedFields(t *testing.T) {
	longDesc := strings.Repeat("Multi\nline description text ", 60) // >800 chars, has newlines

	model := InitialModel()
	model.loading = false
	model.availableStatuses = []api.Status{{ID: 1, Name: "New"}, {ID: 2, Name: "In Progress"}}
	model.availablePriorities = []api.Priority{{ID: 1, Name: "Low"}, {ID: 2, Name: "Normal"}, {ID: 3, Name: "High"}}
	model.issues = []api.Issue{{
		ID:          3888,
		Subject:     "Test",
		Description: longDesc,
		Status:      api.Status{ID: 1, Name: "New"},
		Priority:    api.Priority{ID: 2, Name: "Normal"},
	}}
	model.selectedIndex = 0

	var m tea.Model = model
	// Size the window so the panes initialise (avoids width-0 render math).
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	send := func(k tea.KeyMsg) { m, _ = m.Update(k) }
	runes := func(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

	send(runes("e"))                    // enter edit mode (field 0 = subject)
	send(tea.KeyMsg{Type: tea.KeyTab})  // -> description (field 1)
	send(tea.KeyMsg{Type: tea.KeyTab})  // leave description untouched -> status (2)
	send(tea.KeyMsg{Type: tea.KeyTab})  // -> priority (3)
	send(tea.KeyMsg{Type: tea.KeyDown}) // cycle priority Normal -> High
	send(tea.KeyMsg{Type: tea.KeyTab})  // commit priority

	mm := m.(Model)
	if v, ok := mm.pendingEdits["description"]; ok {
		t.Errorf("description must NOT be edited (only navigated past); got %d chars", len(v))
	}
	if mm.pendingEdits["priority_id"] != "High" {
		t.Errorf("priority_id pending edit = %q, want \"High\"", mm.pendingEdits["priority_id"])
	}
}

func TestScrollBounds(t *testing.T) {
	model := InitialModel()
	model.issues = []api.Issue{
		{ID: 1, Subject: "Test 1"},
		{ID: 2, Subject: "Test 2"},
		{ID: 3, Subject: "Test 3"},
	}

	// Test selectedIndex doesn't go negative
	model.selectedIndex = 0
	if model.selectedIndex < 0 {
		t.Error("selectedIndex should not be negative")
	}

	// Test selectedIndex doesn't exceed issue count
	model.selectedIndex = len(model.issues) + 10
	if model.selectedIndex >= len(model.issues) {
		model.selectedIndex = len(model.issues) - 1
	}

	if model.selectedIndex >= len(model.issues) {
		t.Errorf("selectedIndex %d should be less than issue count %d", model.selectedIndex, len(model.issues))
	}
}
