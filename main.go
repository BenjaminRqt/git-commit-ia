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
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "✖ "+err.Error())
		os.Exit(1)
	}
}

func run() error {
	autoYes := flag.Bool("y", false, "committer sans confirmation")
	stageAll := flag.Bool("a", false, "git add -A avant de générer")
	amend := flag.Bool("r", false, "amender le dernier commit")
	printOnly := flag.Bool("print", false, "afficher le message sans committer")
	modelOverride := flag.String("model", "", "surcharger le modèle")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if *modelOverride != "" {
		cfg.Model = *modelOverride
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("variable d'environnement ANTHROPIC_API_KEY manquante")
	}
	if !gitutil.InRepo() {
		return fmt.Errorf("ce dossier n'est pas un dépôt git")
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
				return fmt.Errorf("impossible d'amender : aucun commit dans ce dépôt")
			}
			return fmt.Errorf("aucun changement à amender (HEAD^ inaccessible ou aucun diff)")
		}
		return fmt.Errorf("aucun changement indexé — fais `git add ...` ou utilise -a")
	}

	branch, _ := gitutil.CurrentBranch()
	ticket := gitutil.ExtractTicket(branch, cfg.TicketPattern)

	var stat string
	if *amend {
		stat, _ = gitutil.AmendStat()
	} else {
		stat, _ = gitutil.StagedStat()
	}
	if stat != "" {
		fmt.Print(stat)
	}
	if ticket != "" {
		fmt.Printf("Ticket détecté : %s  (branche %s)\n", ticket, branch)
	} else {
		fmt.Printf("Aucun ticket détecté dans la branche %s\n", branch)
	}

	client := ai.New(cfg.APIKey, cfg.Model)
	reader := bufio.NewReader(os.Stdin)
	hint := ""

	for {
		msg, err := generate(client, cfg, diff, branch, ticket, hint)
		if err != nil {
			return err
		}

		fmt.Println("\n─────────── Message proposé ───────────")
		fmt.Println(msg)
		fmt.Println("───────────────────────────────────────")

		if *printOnly {
			return nil
		}
		if *autoYes {
			return gitutil.CommitFromMessage(msg, false, *amend)
		}

		fmt.Print("[v]alider  [e]diter  [r]égénérer  [q]uitter › ")
		choice, _ := reader.ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "v", "":
			return gitutil.CommitFromMessage(msg, false, *amend)
		case "e":
			return gitutil.CommitFromMessage(msg, true, *amend)
		case "r":
			fmt.Print("Instruction supplémentaire (Entrée pour ignorer) › ")
			h, _ := reader.ReadString('\n')
			hint = strings.TrimSpace(h)
		case "q":
			fmt.Println("Abandon.")
			return nil
		default:
			fmt.Println("Choix inconnu.")
		}
	}
}

func generate(client *ai.Client, cfg config.Config, diff, branch, ticket, hint string) (string, error) {
	done := make(chan struct{})
	go spinner("Génération du message", done)
	msg, err := client.GenerateCommit(context.Background(), ai.Request{
		Diff:         diff,
		Branch:       branch,
		Ticket:       ticket,
		Hint:         hint,
		Language:     cfg.Language,
		Types:        cfg.Types,
		GenerateBody: cfg.GenerateBody,
		MaxDiffChars: cfg.MaxDiffChars,
	})
	close(done)
	fmt.Print("\r\033[K") // efface la ligne du spinner
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
