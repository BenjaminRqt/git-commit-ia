// Package openai implements the ai.Generator interface for the OpenAI API.
package openai

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
	defaultBaseURL = "https://api.openai.com/v1"
	DefaultModel   = "gpt-4o-mini"
)

// Adapter calls the OpenAI Chat Completions API.
type Adapter struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// New creates an OpenAI adapter. It reads OPENAI_API_KEY from the environment.
// baseURL overrides the default endpoint (useful for Azure or compatible APIs); empty = default.
func New(baseURL string) (*Adapter, error) {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return newAdapter(baseURL)
}

// newAdapter is the internal constructor shared by New and test helpers.
func newAdapter(baseURL string) (*Adapter, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("missing OPENAI_API_KEY environment variable")
	}
	return &Adapter{
		apiKey:  key,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Complete sends the prompts to OpenAI and returns the generated text.
func (a *Adapter) Complete(ctx context.Context, system, user, model string, maxTokens int) (string, error) {
	if model == "" {
		model = DefaultModel
	}
	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"max_completion_tokens": maxTokens,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := a.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai: API call: %w", err)
	}
	defer resp.Body.Close()

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("openai: unreadable response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if out.Error != nil {
			return "", fmt.Errorf("openai: API %d: %s", resp.StatusCode, out.Error.Message)
		}
		return "", fmt.Errorf("openai: API returned status %d", resp.StatusCode)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("openai: empty response (no choices)")
	}
	return out.Choices[0].Message.Content, nil
}
