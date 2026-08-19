// Package provider contains the factory that instantiates the right AI adapter.
package provider

import (
	"fmt"

	"git-ai-commit/internal/ai"
	"git-ai-commit/internal/ai/anthropic"
	"git-ai-commit/internal/ai/openai"
)

// Config holds the parameters needed by the factory.
type Config struct {
	Provider     string // "anthropic" (default) or "openai"
	OpenAIBaseURL string
}

// New instantiates the right AI adapter from the configuration.
// Returns an error if the selected provider is missing its API key.
func New(cfg Config) (ai.Generator, error) {
	switch cfg.Provider {
	case "", "anthropic":
		return anthropic.New()
	case "openai":
		return openai.New(cfg.OpenAIBaseURL)
	default:
		return nil, fmt.Errorf("unknown ai_provider %q (supported: anthropic, openai)", cfg.Provider)
	}
}
