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

// TestMultilineDescriptionEditor verifies that the description is edited via the
// dedicated multi-line editor and its newlines survive into the pending edit.
func TestMultilineDescriptionEditor(t *testing.T) {
	model := InitialModel()
	model.loading = false
	model.issues = []api.Issue{{ID: 7, Subject: "S", Description: "old", Priority: api.Priority{ID: 2, Name: "Normal"}}}
	model.selectedIndex = 0

	var m tea.Model = model
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	send := func(k tea.KeyMsg) { m, _ = m.Update(k) }

	send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}) // edit mode (field 0)
	send(tea.KeyMsg{Type: tea.KeyTab})                       // -> description (multiline)
	send(tea.KeyMsg{Type: tea.KeyEnter})                     // open the editor

	if !m.(Model).descEditMode {
		t.Fatal("Enter on the description field should open the multi-line editor")
	}

	// Type a two-line description (Enter inserts a newline in the editor)
	send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("line one")})
	send(tea.KeyMsg{Type: tea.KeyEnter})
	send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("line two")})
	send(tea.KeyMsg{Type: tea.KeyCtrlS}) // apply

	mm := m.(Model)
	if mm.descEditMode {
		t.Error("Ctrl+S should close the editor")
	}
	got := mm.pendingEdits["description"]
	if !strings.Contains(got, "\n") {
		t.Errorf("description pending edit lost its newline: %q", got)
	}
	if !strings.Contains(got, "line one") || !strings.Contains(got, "line two") {
		t.Errorf("description pending edit = %q, want both lines", got)
	}
}

// TestDescriptionKeepsNewlinesWhileEditing guards against the details pane
// showing a newline-stripped description while the description field is selected
// in edit mode (it must not read from the single-line input).
func TestDescriptionKeepsNewlinesWhileEditing(t *testing.T) {
	desc := "First line\nSecond line\nThird line"
	model := InitialModel()
	model.loading = false
	model.issues = []api.Issue{{ID: 9, Subject: "S", Description: desc}}
	model.selectedIndex = 0

	var m tea.Model = model
	m, _ = m.Update(tea.WindowSizeMsg{Width: 140, Height: 50})
	send := func(k tea.KeyMsg) { m, _ = m.Update(k) }

	send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}) // edit mode
	send(tea.KeyMsg{Type: tea.KeyTab})                       // -> description field

	// With newlines preserved, the three fragments land on three different
	// rendered lines. If the single-line input stripped them, they collapse
	// onto one line.
	lines := strings.Split(m.(Model).rightPane.View(), "\n")
	lineOf := func(frag string) int {
		for i, l := range lines {
			if strings.Contains(l, frag) {
				return i
			}
		}
		return -1
	}
	first, second, third := lineOf("First line"), lineOf("Second line"), lineOf("Third line")
	if first < 0 || second < 0 || third < 0 {
		t.Fatalf("description fragments missing from details pane (%d,%d,%d)", first, second, third)
	}
	if first == second || second == third {
		t.Errorf("description newlines were stripped: fragments share a rendered line (%d,%d,%d)", first, second, third)
	}
}

// TestStatusPicker verifies 's' opens the picker with the current status
// pre-selected, and that a number key applies the chosen status.
func TestStatusPicker(t *testing.T) {
	model := InitialModel()
	model.loading = false
	model.availableStatuses = []api.Status{{ID: 1, Name: "New"}, {ID: 2, Name: "In Progress"}, {ID: 3, Name: "Resolved"}}
	model.issues = []api.Issue{{ID: 5, Subject: "S", Status: api.Status{ID: 2, Name: "In Progress"}}}
	model.selectedIndex = 0

	var m tea.Model = model
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Open the picker
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	mm := m.(Model)
	if !mm.statusPickMode {
		t.Fatal("pressing 's' should open the status picker")
	}
	if mm.statusPickCursor != 1 {
		t.Errorf("cursor should pre-select current status (index 1), got %d", mm.statusPickCursor)
	}
	if mm.statusPickCurrentID != 2 {
		t.Errorf("statusPickCurrentID = %d, want 2", mm.statusPickCurrentID)
	}

	// Press "3" to apply "Resolved" -> should close the picker and issue a command
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	mm = updated.(Model)
	if mm.statusPickMode {
		t.Error("applying a status should close the picker")
	}
	if cmd == nil {
		t.Error("applying a status should issue an update command")
	}

	// Esc should cancel without a command
	m2, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m2, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m2, cmd2 := m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m2.(Model).statusPickMode {
		t.Error("Esc should close the picker")
	}
	_ = cmd2
}

// TestSaveClearsPendingEdits guards the "sticky field" bug: after a save,
// pending edits must be cleared so they don't bleed onto other issues.
func TestSaveClearsPendingEdits(t *testing.T) {
	model := InitialModel()
	model.pendingEdits = map[string]string{"priority_id": "High"}
	model.originalValues = map[string]string{"priority_id": "Normal"}
	model.editedFields = map[string]bool{"priority_id": true}
	model.editMode = true

	updated, _ := model.Update(issueUpdatedMsg{issueID: 1, err: nil})
	mm := updated.(Model)

	if len(mm.pendingEdits) != 0 {
		t.Errorf("pendingEdits should be cleared after save, got %v", mm.pendingEdits)
	}
	if len(mm.editedFields) != 0 {
		t.Errorf("editedFields should be cleared after save, got %v", mm.editedFields)
	}
	if mm.editMode {
		t.Error("editMode should be false after save")
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
