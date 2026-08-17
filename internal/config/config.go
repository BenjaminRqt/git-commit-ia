package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config regroupe tous les réglages de l'outil. Tout est surchargeable via un
// fichier JSON ; seule la clé API vient de l'environnement.
type Config struct {
	Model         string   `json:"model"`
	Language      string   `json:"language"`
	Types         []string `json:"types"`
	TicketPattern string   `json:"ticket_pattern"`
	MaxDiffChars  int      `json:"max_diff_chars"`
	GenerateBody  bool     `json:"generate_body"`

	APIKey string `json:"-"` // injectée depuis ANTHROPIC_API_KEY
}

// Default renvoie une configuration prête à l'emploi.
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

// Load part des valeurs par défaut puis applique le premier fichier de config
// trouvé (les champs absents gardent leur valeur par défaut), et lit la clé API.
func Load() (Config, error) {
	cfg := Default()
	for _, p := range candidatePaths() {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("config %s invalide : %w", p, err)
		}
		break
	}
	cfg.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	return cfg, nil
}

// candidatePaths cherche d'abord une config propre au dépôt, puis une globale.
func candidatePaths() []string {
	paths := []string{".git-ai-commit.json"}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "git-ai-commit", "config.json"))
	}
	return paths
}
