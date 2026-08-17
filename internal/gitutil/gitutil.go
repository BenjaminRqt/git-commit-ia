package gitutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// run executes git and returns stdout, or an error enriched with stderr.
func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return out.String(), nil
}

// InRepo indicates if the current folder is inside a git repository.
func InRepo() bool {
	_, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil
}

// StageAll performs a `git add -A`.
func StageAll() error {
	_, err := run("add", "-A")
	return err
}

// StagedDiff returns the staged diff (git diff --cached).
func StagedDiff() (string, error) { return run("diff", "--cached") }

// HasCommits returns true if the repository has at least one commit.
func HasCommits() bool {
	_, err := run("rev-parse", "--verify", "HEAD")
	return err == nil
}

// AmendDiff returns the diff including the last commit and staged changes.
func AmendDiff() (string, error) {
	if !HasCommits() {
		return "", fmt.Errorf("no commit to amend")
	}
	return run("diff", "--cached", "HEAD^")
}

// StagedStat returns the summary of staged files.
func StagedStat() (string, error) { return run("diff", "--cached", "--stat") }

// AmendStat returns the summary of changes in amend mode.
func AmendStat() (string, error) {
	if !HasCommits() {
		return "", nil
	}
	return run("diff", "--cached", "HEAD^", "--stat")
}

// CurrentBranch returns the name of the current branch.
func CurrentBranch() (string, error) {
	out, err := run("branch", "--show-current")
	return strings.TrimSpace(out), err
}

// ExtractTicket searches for a ticket reference in the branch name.
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

// CommitFromMessage creates the commit from the message.
// If edit is true, git opens the pre-filled editor before validating.
// If amend is true, the commit amends the previous one.
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
	// Connect streams to terminal for the editor to work.
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}
