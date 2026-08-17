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

// Client encapsule les appels à l'API Anthropic.
type Client struct {
	apiKey string
	model  string
	http   *http.Client
}

// New crée un client prêt à l'emploi.
func New(apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 60 * time.Second},
	}
}

// Request rassemble le contexte nécessaire à la génération d'un message.
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

// GenerateCommit demande au modèle un message de commit formaté.
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
		return "", fmt.Errorf("appel API : %w", err)
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
		return "", fmt.Errorf("réponse illisible : %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if out.Error != nil {
			return "", fmt.Errorf("API %d : %s", resp.StatusCode, out.Error.Message)
		}
		return "", fmt.Errorf("API a renvoyé le statut %d", resp.StatusCode)
	}

	var sb strings.Builder
	for _, blk := range out.Content {
		if blk.Type == "text" {
			sb.WriteString(blk.Text)
		}
	}
	return cleanup(sb.String()), nil
}

// cleanup retire d'éventuels backticks ou fences que le modèle ajouterait.
func cleanup(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func buildSystem(r Request) string {
	bodyRule := ""
	if r.GenerateBody {
		bodyRule = "Ensuite, une ligne vide puis 1 à 5 puces (\"- ...\") décrivant les changements clés.\n"
	}
	return fmt.Sprintf(`Tu génères des messages de commit Git en %s.

Le SUJET doit suivre EXACTEMENT ce format :
Type - Quoi - Ticket

Règles :
- Type : un seul mot parmi : %s
- Quoi : description concise à l'impératif (environ 60 caractères max), première lettre en majuscule, pas de point final
- Ticket : la référence fournie telle quelle, ou "N/A" si aucune n'est fournie
%s
Réponds UNIQUEMENT avec le message de commit brut, sans backticks ni commentaire.`,
		r.Language, strings.Join(r.Types, ", "), bodyRule)
}

func buildUser(r Request) string {
	diff := r.Diff
	if r.MaxDiffChars > 0 {
		if rs := []rune(diff); len(rs) > r.MaxDiffChars {
			diff = string(rs[:r.MaxDiffChars]) + "\n... [diff tronqué]"
		}
	}
	ticket := r.Ticket
	if ticket == "" {
		ticket = "N/A"
	}
	hint := ""
	if r.Hint != "" {
		hint = "Consigne supplémentaire : " + r.Hint + "\n"
	}
	return fmt.Sprintf("Branche : %s\nTicket : %s\n%s\nDiff indexé :\n```diff\n%s\n```",
		r.Branch, ticket, hint, diff)
}
