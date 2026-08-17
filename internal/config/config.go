package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all tool settings. Everything is overrideable via a
// JSON file; only the API key comes from the environment.
type Config struct {
	Model         string   `json:"model"`
	Language      string   `json:"language"`
	Types         []string `json:"types"`
	TicketPattern string   `json:"ticket_pattern"`
	MaxDiffChars  int      `json:"max_diff_chars"`
	GenerateBody  bool     `json:"generate_body"`

	APIKey string `json:"-"` // injected from ANTHROPIC_API_KEY
}

// Default returns a ready-to-use configuration.
func Default() Config {
	return Config{
		Model:         "claude-haiku-4-5-20251001",
		Language:      "français",
		Types:         []string{"feat", "fix", "refactor", "docs", "style", "test", "chore", "perf"},
		TicketPattern: `[A-Z]{2,}-\d+`,
		MaxDiffChars:  12000,
		GenerateBody:  true,
	}
}

// Load starts with default values then applies the first config file
// found (missing fields keep their default value), and reads the API key.
func Load() (Config, error) {
	cfg := Default()
	for _, p := range candidatePaths() {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("invalid config %s: %w", p, err)
		}
		break
	}
	cfg.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	return cfg, nil
}

// candidatePaths searches for a repository-specific config first, then a global one.
func candidatePaths() []string {
	paths := []string{".git-ai-commit.json"}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "git-ai-commit", "config.json"))
	}
	return paths
}
