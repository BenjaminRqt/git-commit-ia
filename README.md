# git-ai-commit

Generate Git commit messages in the `Type - What - Ticket` format, straight from
your staged changes.

## What it does & how it works

`git-ai-commit` reads your staged diff and asks the Anthropic API for a commit
message that follows a strict `Type - What - Ticket` format. It runs as a native
git subcommand — `git ai-commit` — because git executes any binary named
`git-ai-commit` found on your `PATH`.

Each run:

1. reads the staged diff (`git diff --cached`);
2. extracts the ticket from the branch name (`feature/VAL-482-...` → `VAL-482`);
3. optionally pulls the ticket's type, summary and description from your ticket
   provider (Jira, …) to steer the message toward **business impact**;
4. generates the message — subject plus an optional short body (max 3 bullets);
5. lets you validate, edit, regenerate or cancel before committing.

Typical usage:

```bash
git add -p
git ai-commit                 # generate, then choose validate / edit / regenerate / quit
```

Handy flags:

| Flag           | Effect                                          |
|----------------|-------------------------------------------------|
| `-a`           | run `git add -A` before generating              |
| `-y`           | commit immediately, no confirmation             |
| `-r`           | amend the last commit                           |
| `--print`      | print the message without committing            |
| `--no-ticket`  | skip the ticket provider for this run           |
| `--model <id>` | override the model (e.g. `claude-sonnet-5`)     |

When regenerating, you can add a freeform hint like *"focus on the Doctrine
migration"* or *"type = fix"*.

## Installation

Prerequisite: Go ≥ 1.22.

```bash
go build -o git-ai-commit .
mv git-ai-commit ~/.local/bin/     # any directory on your PATH
```

The binary **must** keep the exact name `git-ai-commit` for the git subcommand to
work. Then set your API key (add it to `~/.zshrc` or `~/.bashrc` to persist it):

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

Check it works with `git ai-commit -h`. If the command isn't found, confirm the
install directory is on your `PATH` and the file is executable (`chmod +x`).

## Configuration

All settings are optional. The tool merges, in order (later files fill only what
earlier ones leave unset):

1. `.git-ai-commit.json` — at the repo root, for per-project settings;
2. `~/.config/git-ai-commit/config.json` — for your global defaults.

| Field              | Default                                               | Role                                              |
|--------------------|-------------------------------------------------------|---------------------------------------------------|
| `model`            | `claude-haiku-4-5-20251001`                           | Anthropic model                                   |
| `language`         | `français`                                            | Message language                                  |
| `types`            | `feat, fix, refactor, docs, style, test, chore, perf` | Allowed vocabulary for `Type`                     |
| `ticket_pattern`   | `[A-Z]{2,}-\d+`                                        | Regex to extract the ticket from the branch       |
| `max_diff_chars`   | `12000`                                               | Diff truncation before sending to the model       |
| `generate_body`    | `true`                                                | Add a bulleted body (max 3 bullets)               |
| `ticket_provider`  | *(empty = off)*                                       | `"jira"` or empty                                 |
| `jira_base_url`    | *(empty)*                                             | Jira base URL, e.g. `https://acme.atlassian.net`  |
| `max_ticket_chars` | `500`                                                 | Max ticket-description chars sent to the model    |

**Secrets** (API key and provider credentials) live in environment variables only,
never in config files:

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
export JIRA_EMAIL="you@example.com"        # only if ticket_provider = "jira"
export JIRA_API_TOKEN="your-jira-token"
```

### Ticket provider (Jira)

Set `ticket_provider` to `"jira"` and provide `jira_base_url` plus the two
environment variables above. The tool then enriches the message with the ticket's
type, summary and description.

If the provider is off, credentials are missing, or the API call fails (network,
401, 404, timeout), the commit is **never blocked**: the tool falls back to
diff-only generation and prints a warning on stderr. Use `--no-ticket` to skip it
for a single run.

## Technical

### Build & test

```bash
go build -o git-ai-commit .    # binary (exact name required)
go test ./...                  # tests (offline; httptest for external APIs)
go vet ./...                   # static analysis
gofmt -w .                     # formatting (before every commit)
```

Zero external dependencies — standard library only, so the binary is fully
self-contained. Keep it that way.

### Architecture

```
main.go                 entry point, flags, interactive loop
internal/
├── ai/                 Anthropic client + prompt building
├── config/             config struct, JSON + env loading
├── gitutil/            git wrappers (diff, branch, ticket, commit)
└── ticket/             Provider interface + agnostic Ticket type + Noop adapter
    ├── jira/           Jira adapter (API v2, Basic Auth) + type mapping
    └── provider/       factory: picks the adapter from config
```
