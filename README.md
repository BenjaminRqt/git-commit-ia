# git-ai-commit

Generates a commit message in the `Type - What - Ticket` format from the staged diff,
via the Anthropic API. Used as a git subcommand: `git ai-commit`.

## How it works

Git automatically executes any binary named `git-ai-commit` present in the `PATH`
when you type `git ai-commit` — no alias to configure. The tool:

1. reads the staged diff (`git diff --cached`);
2. extracts the ticket number from the branch name (e.g., `feature/VAL-482-...` → `VAL-482`);
3. asks the model for a message in the `Type - What - Ticket` format, with an
   optional body (changelog-style bullets);
4. offers to validate, edit, regenerate, or cancel before committing.

## Installation

Prerequisite: Go ≥ 1.22.

```bash
# compile the binary
go build -o git-ai-commit .

# move it to a folder in your PATH (e.g., /usr/local/bin or ~/.local/bin)
# IMPORTANT: The binary MUST be named exactly 'git-ai-commit'
sudo mv git-ai-commit /usr/local/bin/
```

If after installation, `git ai-commit` is not recognized:
1. Verify that `/usr/local/bin` is in your `$PATH`.
2. Try restarting your terminal.
3. Check permissions: `chmod +x /usr/local/bin/git-ai-commit`.
4. Type `which git-ai-commit` to confirm the location.

Or via `go install`:
```bash
go install .
# Ensure $GOPATH/bin (often ~/go/bin) is in your PATH.
```

## API Key

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

## Usage

```bash
git add -p
git ai-commit            # generate, then propose [v]alidate / [e]dit / [r]egenerate / [q]uit

git ai-commit -a         # automatic git add -A before generating
git ai-commit -r         # amend the last commit (equivalent to git commit --amend)
git ai-commit -y         # validate directly without confirmation
git ai-commit --print    # print the message without committing
git ai-commit --model claude-sonnet-5
```

When regenerating (`r`), you can add a freeform instruction
("focus on Doctrine migration", "type = fix", etc.).

## Configuration

Optional. The tool reads, in order:

1. `.git-ai-commit.json` at the root of the repository (per-project settings);
2. `~/.config/git-ai-commit/config.json` (global settings).

Missing fields keep their default value. See
`.git-ai-commit.example.json`.

| Field            | Default                                                  | Role                                        |
|------------------|----------------------------------------------------------|---------------------------------------------|
| `model`          | `claude-haiku-4-5-20251001`                              | Anthropic model                             |
| `language`       | `français`                                               | Message language                            |
| `types`          | `feat, fix, refactor, docs, style, test, chore, perf`    | Allowed vocabulary for `Type`               |
| `ticket_pattern` | `[A-Z]{2,}-\d+`                                           | Regex for ticket extraction from branch     |
| `max_diff_chars` | `12000`                                                  | Truncation of diff sent to the model        |
| `generate_body`  | `true`                                                   | Adds a bulleted body under the subject      |

## Tests

```bash
go test ./...
```
