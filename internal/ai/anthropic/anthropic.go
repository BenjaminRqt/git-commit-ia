// Package anthropic implements the ai.Generator interface for the Anthropic API.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	endpoint     = "https://api.anthropic.com/v1/messages"
	DefaultModel = "claude-haiku-4-5-20251001"
)

// Adapter calls the Anthropic Messages API.
type Adapter struct {
	apiKey   string
	endpoint string
	http     *http.Client
}

// New creates an Anthropic adapter. It reads ANTHROPIC_API_KEY from the environment.
func New() (*Adapter, error) {
	return NewWithEndpoint(endpoint)
}

// NewWithEndpoint creates an Anthropic adapter with a custom endpoint URL (useful for tests).
func NewWithEndpoint(endpointURL string) (*Adapter, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("missing ANTHROPIC_API_KEY environment variable")
	}
	return &Adapter{
		apiKey:   key,
		endpoint: endpointURL,
		http:     &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Complete sends the prompts to Anthropic and returns the generated text.
func (a *Adapter) Complete(ctx context.Context, system, user, model string, maxTokens int) (string, error) {
	if model == "" {
		model = DefaultModel
	}
	payload := map[string]any{
		"model":      model,
		"max_tokens": maxTokens,
		"system":     system,
		"messages": []map[string]string{
			{"role": "user", "content": user},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic: API call: %w", err)
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
		return "", fmt.Errorf("anthropic: unreadable response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if out.Error != nil {
			return "", fmt.Errorf("anthropic: API %d: %s", resp.StatusCode, out.Error.Message)
		}
		return "", fmt.Errorf("anthropic: API returned status %d", resp.StatusCode)
	}

	var sb strings.Builder
	for _, blk := range out.Content {
		if blk.Type == "text" {
			sb.WriteString(blk.Text)
		}
	}
	return sb.String(), nil
}
