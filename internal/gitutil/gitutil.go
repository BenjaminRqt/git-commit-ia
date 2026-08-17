package gitutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// run exécute git et renvoie stdout, ou une erreur enrichie de stderr.
func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s : %s", strings.Join(args, " "), msg)
	}
	return out.String(), nil
}

// InRepo indique si le dossier courant est dans un dépôt git.
func InRepo() bool {
	_, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil
}

// StageAll fait un `git add -A`.
func StageAll() error {
	_, err := run("add", "-A")
	return err
}

// StagedDiff renvoie le diff indexé (git diff --cached).
func StagedDiff() (string, error) { return run("diff", "--cached") }

// HasCommits renvoie vrai si le dépôt a au moins un commit.
func HasCommits() bool {
	_, err := run("rev-parse", "--verify", "HEAD")
	return err == nil
}

// AmendDiff renvoie le diff incluant le dernier commit et les changements indexés.
func AmendDiff() (string, error) {
	if !HasCommits() {
		return "", fmt.Errorf("aucun commit à amender")
	}
	return run("diff", "--cached", "HEAD^")
}

// StagedStat renvoie le résumé des fichiers indexés.
func StagedStat() (string, error) { return run("diff", "--cached", "--stat") }

// AmendStat renvoie le résumé des changements en mode amend.
func AmendStat() (string, error) {
	if !HasCommits() {
		return "", nil
	}
	return run("diff", "--cached", "HEAD^", "--stat")
}

// CurrentBranch renvoie le nom de la branche courante.
func CurrentBranch() (string, error) {
	out, err := run("branch", "--show-current")
	return strings.TrimSpace(out), err
}

// ExtractTicket cherche la référence de ticket dans le nom de branche.
func ExtractTicket(branch, pattern string) string {
	if pattern == "" {
		return ""
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	return re.FindString(branch)
}

// CommitFromMessage crée le commit à partir du message.
// Si edit vaut true, git ouvre l'éditeur pré-rempli avant de valider.
// Si amend vaut true, le commit amende le précédent.
func CommitFromMessage(message string, edit, amend bool) error {
	f, err := os.CreateTemp("", "git-ai-commit-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(message + "\n"); err != nil {
		return err
	}
	f.Close()

	args := []string{"commit", "-F", f.Name()}
	if edit {
		args = append(args, "-e")
	}
	if amend {
		args = append(args, "--amend")
	}
	cmd := exec.Command("git", args...)
	// On relie les flux au terminal pour que l'éditeur fonctionne.
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}
