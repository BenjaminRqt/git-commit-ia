package jira_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git-ai-commit/internal/ticket/jira"
)

func TestMapIssueType(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Bug", "fix"},
		{"bug", "fix"},
		{"Story", "feat"},
		{"Epic", "feat"},
		{"Task", "chore"},
		{"Sub-task", "chore"},
		{"Improvement", "perf"},
		{"New Feature", "feat"},
		{"Documentation", "docs"},
		{"Test", "test"},
		{"Refactoring", "refactor"},
		{"Unknown", "chore"},
		{"", "chore"},
	}
	for _, c := range cases {
		got := jira.MapIssueType(c.input)
		if got != c.expected {
			t.Errorf("MapIssueType(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}

func TestFetch_ParsesResponse(t *testing.T) {
	payload := map[string]any{
		"fields": map[string]any{
			"summary":     "Fix login bug",
			"description": "Users cannot log in after session expiration.",
			"issuetype":   map[string]any{"name": "Bug"},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	adapter := jira.New(srv.URL, "user@example.com", "token123")
	tk, err := adapter.Fetch("BOZ-909")
	if err != nil {
		t.Fatalf("Fetch() unexpected error: %v", err)
	}
	if tk == nil {
		t.Fatal("Fetch() returned nil")
	}
	if tk.Type != "fix" {
		t.Errorf("Type = %q, want %q", tk.Type, "fix")
	}
	if tk.Summary != "Fix login bug" {
		t.Errorf("Summary = %q", tk.Summary)
	}
	if tk.Description == "" {
		t.Error("Description must not be empty")
	}
}

func TestFetch_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	adapter := jira.New(srv.URL, "user@example.com", "token123")
	tk, err := adapter.Fetch("BOZ-000")
	if err == nil {
		t.Error("Fetch() should return an error for a 404 status")
	}
	if tk != nil {
		t.Error("Fetch() should return nil on error")
	}
}
