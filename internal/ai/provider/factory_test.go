package provider_test

import (
	"testing"

	aiprovider "git-ai-commit/internal/ai/provider"
)

func TestNew_DefaultIsAnthropic(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "test-key")
	gen, err := aiprovider.New(aiprovider.Config{Provider: ""})
	if err != nil {
		t.Fatalf("New() unexpected error for default provider: %v", err)
	}
	if gen == nil {
		t.Error("New() returned nil generator")
	}
}

func TestNew_AnthropicExplicit(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "test-key")
	gen, err := aiprovider.New(aiprovider.Config{Provider: "anthropic"})
	if err != nil {
		t.Fatalf("New() unexpected error for anthropic provider: %v", err)
	}
	if gen == nil {
		t.Error("New() returned nil generator")
	}
}

func TestNew_OpenAI(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")
	gen, err := aiprovider.New(aiprovider.Config{Provider: "openai"})
	if err != nil {
		t.Fatalf("New() unexpected error for openai provider: %v", err)
	}
	if gen == nil {
		t.Error("New() returned nil generator")
	}
}

func TestNew_UnknownProvider(t *testing.T) {
	_, err := aiprovider.New(aiprovider.Config{Provider: "unknown"})
	if err == nil {
		t.Error("New() should return an error for an unknown provider")
	}
}

func TestNew_MissingKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	_, err := aiprovider.New(aiprovider.Config{Provider: "anthropic"})
	if err == nil {
		t.Error("New() should return an error when the API key is missing")
	}
}
