// Package jira implements the Jira adapter for the ticket.Provider interface.
package jira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git-ai-commit/internal/ticket"
)

// issueTypeMap maps Jira issue types to commit types.
var issueTypeMap = map[string]string{
	"bug":           "fix",
	"story":         "feat",
	"epic":          "feat",
	"task":          "chore",
	"subtask":       "chore",
	"sub-task":      "chore",
	"improvement":   "perf",
	"new feature":   "feat",
	"documentation": "docs",
	"test":          "test",
	"refactoring":   "refactor",
}

// Adapter is the Jira ticket adapter.
type Adapter struct {
	baseURL  string
	email    string
	apiToken string
	http     *http.Client
}

// New creates a Jira adapter with the provided parameters.
func New(baseURL, email, apiToken string) *Adapter {
	return &Adapter{
		baseURL:  strings.TrimRight(baseURL, "/"),
		email:    email,
		apiToken: apiToken,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

// jiraResponse is the partial structure of the Jira API v2 response.
type jiraResponse struct {
	Fields struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		IssueType   struct {
			Name string `json:"name"`
		} `json:"issuetype"`
	} `json:"fields"`
}

// Fetch retrieves metadata for a Jira ticket.
func (a *Adapter) Fetch(key string) (*ticket.Ticket, error) {
	url := fmt.Sprintf("%s/rest/api/2/issue/%s?fields=summary,issuetype,description", a.baseURL, key)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("jira: build request: %w", err)
	}
	req.SetBasicAuth(a.email, a.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jira: API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira: HTTP status %d for ticket %s", resp.StatusCode, key)
	}

	var jr jiraResponse
	if err := json.NewDecoder(resp.Body).Decode(&jr); err != nil {
		return nil, fmt.Errorf("jira: decode response: %w", err)
	}

	commitType := MapIssueType(jr.Fields.IssueType.Name)

	return &ticket.Ticket{
		Type:        commitType,
		Summary:     jr.Fields.Summary,
		Description: jr.Fields.Description,
	}, nil
}

// MapIssueType converts a Jira issue type to a commit type.
// Exported for testing.
func MapIssueType(jiraType string) string {
	if t, ok := issueTypeMap[strings.ToLower(jiraType)]; ok {
		return t
	}
	return "chore"
}
