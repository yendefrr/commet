package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		expectedType string
		expectedScope string
		expectedBoard string
		expectedDesc string
		expectedForce bool
	}{
		{
			name:         "simple feature",
			message:      "Feature(log): added logger wrap",
			expectedType: "Feature",
			expectedScope: "log",
			expectedDesc: "added logger wrap",
			expectedForce: false,
		},
		{
			name:         "simple fix",
			message:      "Fix: some commit",
			expectedType: "Fix",
			expectedScope: "",
			expectedDesc: "some commit",
			expectedForce: false,
		},
		{
			name:         "board with wrapped type",
			message:      "B-378670(payment,spare): <Fix> removed spares update",
			expectedType: "Fix",
			expectedScope: "payment,spare",
			expectedBoard: "B-378670",
			expectedDesc: "removed spares update",
			expectedForce: false,
		},
		{
			name:         "board with unwrapped type",
			message:      "U-1234(user): Feature some feat",
			expectedType: "Feature",
			expectedScope: "user",
			expectedBoard: "U-1234",
			expectedDesc: "some feat",
			expectedForce: false,
		},
		{
			name:         "force major with exclamation",
			message:      "Fix!(auth): critical security patch",
			expectedType: "Fix",
			expectedScope: "auth",
			expectedDesc: "critical security patch",
			expectedForce: true,
		},
		{
			name:         "breaking keyword",
			message:      "Feature: Breaking change in API",
			expectedType: "Feature",
			expectedScope: "",
			expectedDesc: "Breaking change in API",
			expectedForce: true,
		},
		{
			name:         "refactor with scope",
			message:      "Refactor(db): optimize queries",
			expectedType: "Refactor",
			expectedScope: "db",
			expectedDesc: "optimize queries",
			expectedForce: false,
		},
		{
			name:         "docs without scope",
			message:      "Docs: update API documentation",
			expectedType: "Docs",
			expectedScope: "",
			expectedDesc: "update API documentation",
			expectedForce: false,
		},
		{
			name:         "board without scope wrapped type",
			message:      "B-12345: <Build> update dependencies",
			expectedType: "Build",
			expectedScope: "",
			expectedBoard: "B-12345",
			expectedDesc: "update dependencies",
			expectedForce: false,
		},
		{
			name:         "board without scope unwrapped type",
			message:      "U-9876: Tests added for parser",
			expectedType: "Tests",
			expectedScope: "",
			expectedBoard: "U-9876",
			expectedDesc: "added for parser",
			expectedForce: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.message)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			if result.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", result.Type, tt.expectedType)
			}

			if result.Scope != tt.expectedScope {
				t.Errorf("Scope = %v, want %v", result.Scope, tt.expectedScope)
			}

			if result.Board != tt.expectedBoard {
				t.Errorf("Board = %v, want %v", result.Board, tt.expectedBoard)
			}

			if result.Description != tt.expectedDesc {
				t.Errorf("Description = %v, want %v", result.Description, tt.expectedDesc)
			}

			if result.ForceMajor != tt.expectedForce {
				t.Errorf("ForceMajor = %v, want %v", result.ForceMajor, tt.expectedForce)
			}
		})
	}
}

func TestParseMultiple(t *testing.T) {
	messages := []string{
		"Feature(auth): add OAuth support",
		"Fix(api): handle null responses",
		"Invalid commit message without colon",
		"Refactor(db): optimize queries",
		"",
	}

	results := ParseMultiple(messages)

	if len(results) != 3 {
		t.Errorf("Expected 3 parsed commits, got %d", len(results))
	}

	if results[0].Type != "Feature" {
		t.Errorf("First commit type = %v, want Feature", results[0].Type)
	}
}

func TestIsValidCommit(t *testing.T) {
	tests := []struct {
		name    string
		commit  *Commit
		isValid bool
	}{
		{
			name:    "valid commit",
			commit:  &Commit{Type: "Feature", Description: "add feature"},
			isValid: true,
		},
		{
			name:    "invalid commit - no type",
			commit:  &Commit{Type: "", Description: "something"},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.commit.IsValidCommit(); got != tt.isValid {
				t.Errorf("IsValidCommit() = %v, want %v", got, tt.isValid)
			}
		})
	}
}

func TestCommitString(t *testing.T) {
	commit := &Commit{
		Type:        "Feature",
		Scope:       "auth",
		Board:       "B-123",
		Description: "add OAuth",
		ForceMajor:  true,
	}

	result := commit.String()

	if result == "" {
		t.Error("String() returned empty string")
	}

	expected := []string{"B-123", "Feature!", "(auth)", "add OAuth"}
	for _, exp := range expected {
		if !contains(result, exp) {
			t.Errorf("String() = %v, should contain %v", result, exp)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
