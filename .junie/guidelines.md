# Project guidelines — git-ai-commit

This file describes the project to the Junie agent. Follow these conventions for
any code generation or modification.

## Overview

`git-ai-commit` is a command-line tool written in **Go** that generates a commit
message in the `Type - Quoi - Ticket` format from the staged diff, via the
**Anthropic API**. It is used as a git subcommand: `git ai-commit`.

Key mechanism: git automatically runs any binary named `git-ai-commit` present in
the `PATH` when you type `git ai-commit`. The binary name is therefore
significant — do not rename it without a reason.

## Stack & constraints

- **Go ≥ 1.22**, module named `git-ai-commit` (see `go.mod`).
- **No external dependencies**: standard library only. This is a deliberate choice
  for trivial distribution (a single self-contained binary). Do not introduce a
  third-party dependency without a strong need; prefer `net/http`, `encoding/json`,
  `os/exec`, etc.
- **Module mode must be enabled** (`GO111MODULE=on`) so that
  `git-ai-commit/internal/...` imports resolve.
- All **user-facing text is in English** (messages, errors, interactive prompts).
  Keep this convention.

## Layout

```
git-ai-commit/
├── go.mod                          # module git-ai-commit, go 1.22
├── main.go                         # entry point, flags, interactive loop, spinner
├── README.md
├── .git-ai-commit.example.json     # example of overridable config
└── internal/
    ├── ai/
    │   └── ai.go                   # Anthropic API client + prompt building
    ├── config/
    │   └── config.go               # Config struct, loading from JSON file + env
    ├── gitutil/
    │   ├── gitutil.go              # git wrappers (diff, branch, ticket, commit)
    │   └── gitutil_test.go         # ExtractTicket test
    └── ticket/
        ├── ticket.go               # Provider interface, Ticket type, Noop adapter
        ├── jira/
        │   ├── jira.go             # Jira adapter (API v2, Basic Auth)
        │   └── jira_test.go        # MapIssueType + response parsing tests (httptest)
        └── provider/
            ├── factory.go          # factory: instantiates the right adapter from config
            └── factory_test.go     # factory tests (noop for empty/unknown/incomplete config)
```

One package = one directory. Never place multiple packages in the same directory
(Go forbids it).

## Package responsibilities

- **`main`** (`main.go`): flag parsing, orchestration, interactive loop
  (`[v]alider / [e]diter / [r]égénérer / [q]uitter`), spinner. All error logic
  bubbles up through `run() error` and is then printed with the `✖ ` prefix.
- **`internal/config`**: `Config` + `loadConfig()`. Starts from the default values
  (`defaultConfig`/`Default`), applies the first file found (missing fields keep
  their default via `json.Unmarshal` onto the pre-filled struct), then reads the
  API key from the environment.
- **`internal/gitutil`**: all git interactions go through the `run` helper (via
  `os/exec`), which enriches the error with `stderr`. The commit is done via a
  temporary file (`git commit -F <tmp>`), with `-e` to open the editor.
- **`internal/ai`**: Anthropic HTTP client and building of the system / user
  prompts. Never log the API key.
- **`internal/ticket`**: `Provider` interface, provider-agnostic `Ticket` type,
  `Noop` adapter. `main` and `ai` depend ONLY on this package — never on a concrete
  adapter.
- **`internal/ticket/jira`**: Jira adapter (API v2, Basic Auth). Contains the
  Jira type → commit type mapping (Bug→fix, Story→feat, etc.).
- **`internal/ticket/provider`**: factory that instantiates the right adapter from
  config. Placed in a separate sub-package to avoid the import cycle with
  `ticket/jira`.

## Format contract (to preserve)

The commit **subject** strictly follows: `Type - Quoi - Ticket`.

- `Type`: one word from the configurable list (`feat, fix, refactor, docs, style,
  test, chore, perf` by default).
- `Quoi`: imperative description, ~60 characters, initial capital, no trailing period.
- `Ticket`: reference extracted from the branch name, or `N/A`.

If `generate_body` is enabled: a blank line then **at most 3 bullets** `- ...`
oriented toward "why / business impact", not a line-by-line inventory of the diff.
This structure is defined in `buildSystem` / `buildUser` (package `ai`) — any change
to the format happens there.

When ticket context is available, the "Quoi" must describe the **business outcome**
(problem solved, value delivered), not the technical detail of the diff.

The ticket is extracted from the branch name via regex (`ticket_pattern`, default
`[A-Z]{2,}-\d+`) in `ExtractTicket`. Example: `feature/VAL-482-refonte` → `VAL-482`.

## Configuration

Files read, in order (missing fields = default values):

1. `.git-ai-commit.json` at the repository root (per-project settings);
2. `~/.config/git-ai-commit/config.json` (global settings).

Fields: `model`, `language`, `types`, `ticket_pattern`, `max_diff_chars`,
`generate_body`, `ticket_provider`, `jira_base_url`, `max_ticket_chars`.
The API key comes **exclusively** from the `ANTHROPIC_API_KEY` environment variable
(never from a file, never hardcoded).

Ticket provider secrets — environment variables only:
- `JIRA_EMAIL`: email address of the Jira account.
- `JIRA_API_TOKEN`: Jira API token.

## Anthropic API integration

- Endpoint: `POST https://api.anthropic.com/v1/messages`.
- Headers: `content-type: application/json`, `x-api-key: <key>`,
  `anthropic-version: 2023-06-01`.
- Default model: `claude-haiku-4-5-20251001` (fast and cheap, suited to a diff) —
  overridable via config or `--model`.
- The diff sent is truncated to `max_diff_chars` (default 12000) to keep cost under
  control.
- The response is parsed by extracting the `content[].text` blocks of type `text`.

## CLI flags

- `-y`: commit without confirmation.
- `-a`: `git add -A` before generating.
- `--print`: display the message without committing.
- `--model <id>`: override the model.
- `--no-ticket`: disable the ticket provider for this run.

## Adapter pattern — ticket providers

`main` and `ai` know ONLY the `ticket.Provider` interface and the `ticket.Ticket`
type. To add a new provider:

1. Create `internal/ticket/<name>/<name>.go` that implements `ticket.Provider`.
2. Put the provider-type → commit-type mapping **inside the adapter** (vocabulary
   specific to each tool).
3. Register the new case in `internal/ticket/provider/factory.go`.
4. Add the necessary config fields in `internal/config/config.go`.
5. No change to `main.go` or `internal/ai/ai.go`.

### Graceful degradation (mandatory)

If the provider is not configured, if credentials are missing, or if the call fails
(network, 401, 404, timeout): **never block the commit**. Fall back silently to
generation from the diff alone, with a warning on stderr. The ticket description
must never be copied verbatim into the message.

## Dev commands

```bash
go build -o git-ai-commit .     # build the binary (exact name required)
go test ./...                   # run the tests
go vet ./...                    # static analysis
gofmt -w .                      # formatting (mandatory before commit)
```

To test under real conditions as a git subcommand: place the `git-ai-commit` binary
in a directory on the `PATH`, then run `git ai-commit` in a repository with staged
changes (requires `ANTHROPIC_API_KEY`).

## Code conventions

- Systematic `gofmt` formatting; idiomatic Go naming (exported = PascalCase).
- Errors are returned, not handled with `panic`; messages in English, contextualized
  with `fmt.Errorf("...: %w", err)`.
- No network side effects in tests; use `net/http/httptest` to simulate external
  APIs (see `internal/ticket/jira/jira_test.go`).
- Add a test when introducing a non-trivial pure function (extraction, parsing,
  formatting, type mapping).

## Do not break

- The binary name `git-ai-commit` (git subcommand).
- The absence of external dependencies.
- The `Type - Quoi - Ticket` format contract.
- The origin of the API key (environment variable only).
- User-facing text in English.
- Operation without a ticket provider (identical to before).
- `main` and `ai` depend ONLY on `ticket.Provider` / `ticket.Ticket`, never on a
  concrete adapter.