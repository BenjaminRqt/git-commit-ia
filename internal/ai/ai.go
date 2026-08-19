// Package ai handles prompt building and commit message generation.
// The actual HTTP call is delegated to a Generator adapter.
package ai

import (
	"context"
	"fmt"
	"strings"

	"git-ai-commit/internal/ticket"
)

// Generator is the interface that every AI backend adapter must implement.
// It receives already-built prompts and returns raw text.
type Generator interface {
	Complete(ctx context.Context, system, user, model string, maxTokens int) (string, error)
}

// Request gathers the context needed for message generation.
type Request struct {
	Diff           string
	Branch         string
	Ticket         string
	TicketInfo     *ticket.Ticket
	MaxTicketChars int
	Hint           string
	Language       string
	Types          []string
	GenerateBody   bool
	MaxDiffChars   int
}

// GenerateCommit builds the prompts and delegates the API call to the generator.
func GenerateCommit(ctx context.Context, gen Generator, model string, r Request) (string, error) {
	system := buildSystem(r)
	user := buildUser(r)
	raw, err := gen.Complete(ctx, system, user, model, 1024)
	if err != nil {
		return "", err
	}
	return cleanup(raw), nil
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
		bodyRule = "Next, a blank line and at most 3 bullets (\"- ...\") focused on WHY/impact, not a line-by-line inventory of the diff.\n"
	}
	ticketGuidance := ""
	if r.TicketInfo != nil {
		ticketGuidance = `
When ticket context is provided:
- The "What" must describe the BUSINESS RESULT for the user (problem solved, value delivered), not the technical detail of the diff.
- Use the ticket type as a hint to choose the commit Type, but stay within the allowed list.
- The ticket summary and description express the INTENT; the diff ensures accuracy.
- NEVER copy the ticket description verbatim into the commit message.
`
	}
	return fmt.Sprintf(`You generate Git commit messages in %s.

The SUBJECT must EXACTLY follow this format:
Type - What - Ticket

Rules:
- Type: a single word from: %s
- What: concise description in imperative mood (approx 60 chars max), capitalized, no trailing period
- Ticket: the reference provided as is, or "N/A" if none is provided
%s%s
Respond ONLY with the raw commit message, no backticks or commentary.`,
		r.Language, strings.Join(r.Types, ", "), bodyRule, ticketGuidance)
}

func buildUser(r Request) string {
	diff := r.Diff
	if r.MaxDiffChars > 0 {
		if rs := []rune(diff); len(rs) > r.MaxDiffChars {
			diff = string(rs[:r.MaxDiffChars]) + "\n... [diff truncated]"
		}
	}
	ticketRef := r.Ticket
	if ticketRef == "" {
		ticketRef = "N/A"
	}
	hint := ""
	if r.Hint != "" {
		hint = "Extra instruction: " + r.Hint + "\n"
	}
	ticketContext := ""
	if r.TicketInfo != nil {
		desc := r.TicketInfo.Description
		if r.MaxTicketChars > 0 {
			if rs := []rune(desc); len(rs) > r.MaxTicketChars {
				desc = string(rs[:r.MaxTicketChars]) + "... [truncated]"
			}
		}
		ticketContext = fmt.Sprintf("\nTicket context:\n- Suggested type: %s\n- Summary: %s\n- Description: %s\n",
			r.TicketInfo.Type, r.TicketInfo.Summary, desc)
	}
	return fmt.Sprintf("Branch: %s\nTicket: %s\n%s%s\nStaged diff:\n```diff\n%s\n```",
		r.Branch, ticketRef, ticketContext, hint, diff)
}
