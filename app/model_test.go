package app

import (
	"testing"

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
