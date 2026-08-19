// Package provider contient la fabrique qui instancie le bon adaptateur de ticket.
package provider

import (
	"os"

	"git-ai-commit/internal/ticket"
	"git-ai-commit/internal/ticket/jira"
)

// Config contient les paramètres nécessaires à la fabrique.
type Config struct {
	Provider    string // "jira", "" = désactivé
	JiraBaseURL string
}

// New instancie le bon adaptateur selon la configuration.
// Renvoie toujours un Provider valide (Noop si rien n'est configuré ou si les
// identifiants sont manquants).
func New(cfg Config) ticket.Provider {
	switch cfg.Provider {
	case "jira":
		email := os.Getenv("JIRA_EMAIL")
		token := os.Getenv("JIRA_API_TOKEN")
		if cfg.JiraBaseURL != "" && email != "" && token != "" {
			return jira.New(cfg.JiraBaseURL, email, token)
		}
		return ticket.Noop{}
	default:
		return ticket.Noop{}
	}
}
