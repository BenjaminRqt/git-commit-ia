package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"git-ai-commit/internal/ai"
	"git-ai-commit/internal/config"
	"git-ai-commit/internal/gitutil"
	"git-ai-commit/internal/ticket"
	"git-ai-commit/internal/ticket/provider"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "✖ "+err.Error())
		os.Exit(1)
	}
}

func run() error {
	autoYes := flag.Bool("y", false, "commit without confirmation")
	stageAll := flag.Bool("a", false, "git add -A before generating")
	amend := flag.Bool("r", false, "amend the last commit")
	printOnly := flag.Bool("print", false, "print the message without committing")
	modelOverride := flag.String("model", "", "override the model")
	noTicket := flag.Bool("no-ticket", false, "disable ticket provider for this run")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if *modelOverride != "" {
		cfg.Model = *modelOverride
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("missing ANTHROPIC_API_KEY environment variable")
	}
	if !gitutil.InRepo() {
		return fmt.Errorf("this folder is not a git repository")
	}

	if *stageAll {
		if err := gitutil.StageAll(); err != nil {
			return err
		}
	}

	var diff string
	if *amend {
		diff, err = gitutil.AmendDiff()
	} else {
		diff, err = gitutil.StagedDiff()
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(diff) == "" {
		if *amend {
			if !gitutil.HasCommits() {
				return fmt.Errorf("cannot amend: no commits in this repository")
			}
			return fmt.Errorf("no changes to amend (HEAD^ inaccessible or no diff)")
		}
		return fmt.Errorf("no staged changes — use `git add ...` or use -a")
	}

	branch, _ := gitutil.CurrentBranch()
	ticketKey := gitutil.ExtractTicket(branch, cfg.TicketPattern)

	var stat string
	if *amend {
		stat, _ = gitutil.AmendStat()
	} else {
		stat, _ = gitutil.StagedStat()
	}
	if stat != "" {
		fmt.Print(stat)
	}
	if ticketKey != "" {
		fmt.Printf("Detected ticket: %s  (branch %s)\n", ticketKey, branch)
	} else {
		fmt.Printf("No ticket detected in branch %s\n", branch)
	}

	// Fetch ticket context via the factory
	var ticketInfo *ticket.Ticket
	if !*noTicket && ticketKey != "" {
		p := provider.New(provider.Config{
			Provider:    cfg.TicketProvider,
			JiraBaseURL: cfg.JiraBaseURL,
		})
		info, err := p.Fetch(ticketKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ Could not fetch ticket %s: %v\n", ticketKey, err)
		} else {
			ticketInfo = info
		}
	}

	client := ai.New(cfg.APIKey, cfg.Model)
	reader := bufio.NewReader(os.Stdin)
	hint := ""

	for {
		msg, err := generate(client, cfg, diff, branch, ticketKey, ticketInfo, hint)
		if err != nil {
			return err
		}

		if ticketInfo != nil {
			fmt.Println("\n─────────── Initial request ───────────")
			fmt.Printf("Ticket: %s — %s\n", ticketKey, ticketInfo.Summary)
			if ticketInfo.Description != "" {
				desc := ticketInfo.Description
				if rs := []rune(desc); len(rs) > 200 {
					desc = string(rs[:200]) + "…"
				}
				fmt.Printf("Request: %s\n", desc)
			}
		}
		fmt.Println("\n─────────── Proposed Message ───────────")
		fmt.Println(msg)
		fmt.Println("────────────────────────────────────────")

		if *printOnly {
			return nil
		}
		if *autoYes {
			return gitutil.CommitFromMessage(msg, false, *amend)
		}

		fmt.Print("[v]alidate  [e]dit  [r]egenerate  [q]uit › ")
		choice, _ := reader.ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "v", "":
			return gitutil.CommitFromMessage(msg, false, *amend)
		case "e":
			return gitutil.CommitFromMessage(msg, true, *amend)
		case "r":
			fmt.Print("Extra instruction (Enter to skip) › ")
			h, _ := reader.ReadString('\n')
			hint = strings.TrimSpace(h)
		case "q":
			fmt.Println("Aborted.")
			return nil
		default:
			fmt.Println("Unknown choice.")
		}
	}
}

func generate(client *ai.Client, cfg config.Config, diff, branch, ticketKey string, ticketInfo *ticket.Ticket, hint string) (string, error) {
	done := make(chan struct{})
	go spinner("Generating message", done)
	msg, err := client.GenerateCommit(context.Background(), ai.Request{
		Diff:           diff,
		Branch:         branch,
		Ticket:         ticketKey,
		TicketInfo:     ticketInfo,
		MaxTicketChars: cfg.MaxTicketChars,
		Hint:           hint,
		Language:       cfg.Language,
		Types:          cfg.Types,
		GenerateBody:   cfg.GenerateBody,
		MaxDiffChars:   cfg.MaxDiffChars,
	})
	close(done)
	fmt.Print("\r\033[K") // clear spinner line
	return msg, err
}

func spinner(label string, done chan struct{}) {
	frames := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	for i := 0; ; i++ {
		select {
		case <-done:
			return
		default:
			fmt.Printf("\r%c %s...", frames[i%len(frames)], label)
			time.Sleep(80 * time.Millisecond)
		}
	}
}
