package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const endpoint = "https://api.anthropic.com/v1/messages"

// Client wraps calls to the Anthropic API.
type Client struct {
	apiKey string
	model  string
	http   *http.Client
}

// New creates a ready-to-use client.
func New(apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 60 * time.Second},
	}
}

// Request gathers the context needed for message generation.
type Request struct {
	Diff         string
	Branch       string
	Ticket       string
	Hint         string
	Language     string
	Types        []string
	GenerateBody bool
	MaxDiffChars int
}

// GenerateCommit asks the model for a formatted commit message.
func (c *Client) GenerateCommit(ctx context.Context, r Request) (string, error) {
	payload := map[string]any{
		"model":      c.model,
		"max_tokens": 1024,
		"system":     buildSystem(r),
		"messages": []map[string]string{
			{"role": "user", "content": buildUser(r)},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("API call: %w", err)
	}
	defer resp.Body.Close()
	
	var out struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("unreadable response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if out.Error != nil {
			return "", fmt.Errorf("API %d: %s", resp.StatusCode, out.Error.Message)
		}
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var sb strings.Builder
	for _, blk := range out.Content {
		if blk.Type == "text" {
			sb.WriteString(blk.Text)
		}
	}
	return cleanup(sb.String()), nil
}

// cleanup removes potential backticks or fences the model might add.
func cleanup(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func buildSystem(r Request) string {
	bodyRule := ""
	if r.GenerateBody {
		bodyRule = "Next, a blank line and 1 to 5 bullets (\"- ...\") describing key changes.\n"
	}
	return fmt.Sprintf(`You generate Git commit messages in %s.

The SUBJECT must EXACTLY follow this format:
Type - What - Ticket

Rules:
- Type: a single word from: %s
- What: concise description in imperative mood (approx 60 chars max), capitalized, no trailing period
- Ticket: the reference provided as is, or "N/A" if none is provided
%s
Respond ONLY with the raw commit message, no backticks or commentary.`,
		r.Language, strings.Join(r.Types, ", "), bodyRule)
}

func buildUser(r Request) string {
	diff := r.Diff
	if r.MaxDiffChars > 0 {
		if rs := []rune(diff); len(rs) > r.MaxDiffChars {
			diff = string(rs[:r.MaxDiffChars]) + "\n... [diff truncated]"
		}
	}
	ticket := r.Ticket
	if ticket == "" {
		ticket = "N/A"
	}
	hint := ""
	if r.Hint != "" {
		hint = "Extra instruction: " + r.Hint + "\n"
	}
	return fmt.Sprintf("Branch: %s\nTicket: %s\n%s\nStaged diff:\n```diff\n%s\n```",
		r.Branch, ticket, hint, diff)
}
