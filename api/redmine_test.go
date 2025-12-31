package api

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://example.com", "test-key")

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.BaseURL != "https://example.com" {
		t.Errorf("BaseURL = %s, want https://example.com", client.BaseURL)
	}

	if client.APIKey != "test-key" {
		t.Errorf("APIKey = %s, want test-key", client.APIKey)
	}
}

func TestIssueStructure(t *testing.T) {
	issue := Issue{
		ID:      1,
		Subject: "Test Issue",
		Status: Status{
			ID:   1,
			Name: "New",
		},
		Priority: Priority{
			ID:   1,
			Name: "Normal",
		},
		Project: Project{
			ID:   1,
			Name: "Test Project",
		},
		Author: User{
			ID:   1,
			Name: "Test User",
		},
	}

	if issue.ID != 1 {
		t.Errorf("Issue.ID = %d, want 1", issue.ID)
	}

	if issue.Subject != "Test Issue" {
		t.Errorf("Issue.Subject = %s, want Test Issue", issue.Subject)
	}

	if issue.Status.Name != "New" {
		t.Errorf("Issue.Status.Name = %s, want New", issue.Status.Name)
	}
}
