# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository. `CLAUDE.md` symlinks here.

## Build & Test Commands

```bash
make build          # Build binary to ./bin/sl (with ldflags for version injection)
make test           # go test ./...
make lint           # golangci-lint run ./... (CI uses v1.64)
make vet            # go vet ./...
make fmt            # gofmt -w .
make coverage       # Generate coverage report
make man            # Generate man pages via cmd/gen-man
```

Run a single test: `go test -run TestHandleError_401 ./internal/api/`
Run tests with race detector (as CI does): `go test -race ./...`
Release: `git tag vX.Y.Z && git push origin vX.Y.Z` — triggers goreleaser (builds, checksums, homebrew tap update).

Branch protection is enabled on `main` — all changes go through PRs with CI passing. Repo admins can bypass.

## Upstream Tracking

Three workflows automatically track `simple-login/app` for API changes and implement fixes:

| Workflow | Trigger | Role |
|----------|---------|------|
| `upstream-watcher.yml` | Weekly cron (Mon 8am UTC) + manual | Fetches upstream diffs, Claude analyzes all changed files, creates issues |
| `on-issue-open.yml` | `issues: [labeled]` / `@claude` comment / `workflow_dispatch` | Security gate (verifies actor has write access) + orchestrator |
| `claude-code.yml` | `workflow_call` (reusable) | Implements the issue and opens a PR |

**Flow:** watcher creates issue with `upstream-change` label → if auto-fixable, adds `claude-fix` label separately → `on-issue-open` triggers (exactly once, on the `claude-fix` label event) → calls `claude-code.yml` → PR opened.

**Checkpoint:** `simplelogin_app_commit_checkpoint` on `main` stores the last analyzed upstream commit SHA. Updated via `GH_PAT` (bypasses branch protection). `ci.yml` has `paths-ignore` for this file.

**Security:** `on-issue-open.yml` verifies `github.actor` has write/admin/maintain access before running Claude. The `GH_PAT` is only used in the watcher workflow, not in the fix pipeline.

**Secrets required:** `ANTHROPIC_API_KEY`, `GH_PAT` (repo scope + branch protection bypass).

**Manual triggers:** `gh workflow run upstream-watcher.yml` to check upstream now. `gh workflow run on-issue-open.yml -f issue_number=N` to implement a specific issue. Or comment `@claude` on any issue.

## Architecture

This is a cobra-based CLI (`sl`) for the SimpleLogin API. Binary entrypoint is `cmd/sl/main.go`.

### Package structure

- `cmd/sl/main.go` — entrypoint, version injection via ldflags, exit code handling (1=runtime error, 2=usage error)
- `cmd/root.go` — root cobra command, `--verbose` persistent flag, `PersistentPreRun` sets `api.Verbose`
- `cmd/<resource>/<verb>.go` — one file per subcommand (e.g., `cmd/alias/create.go`)
- `cmd/<resource>/<resource>.go` — parent command that wires subcommands via `Cmd.AddCommand()`
- `internal/api/client.go` — API client, all HTTP methods, error handling
- `internal/auth/auth.go` — config file management, API key priority chain, base URL resolution
- `internal/output/output.go` — table rendering, JSON/jq output, terminal detection, confirmation prompts

### Command pattern

Every command follows this structure:

1. Package-level `var xxxCmd = &cobra.Command{...}` with `RunE` pointing to `runXxx`
2. Package-level flag vars (e.g., `var createJSON bool`)
3. `init()` registers flags and is called by Go automatically
4. `runXxx` function: get API key → construct client → validate input → call API → branch on `--json`/`--jq` for output

Client construction always looks like: `client := api.NewClient(key, auth.GetAPIBase())`

### API client design

- `NewClient(apiKey, baseURL string)` — baseURL falls back to `https://app.simplelogin.io` when empty
- API methods return `(typedResult, rawJSON []byte, error)` — rawJSON is the untouched API response, passed directly to `output.PrintJSON`/`PrintJQ`. Never re-marshal typed results for JSON output.
- The `Authentication` header (not `Authorization`) is correct per SimpleLogin's API spec
- `wrapNetworkError` translates DNS/timeout/TLS/connection-refused errors into user-friendly messages
- `ResolveAliasID(idOrEmail)` lets commands accept either numeric IDs or email addresses

### Auth priority chain (`GetAPIKey`)

1. `SIMPLELOGIN_API_KEY` env var
2. `SL_API_KEY` env var
3. 1Password CLI (`op read`) if `op_ref` is configured
4. `api_key` from config file

## Agentic-First Design Principles

This CLI is designed for agent and script consumption first, human use second.

**Stdout vs stderr separation is critical.** stdout = machine-readable data only (JSON, table, created alias email). stderr = status messages, prompts, verbose logging. This enables `NEW_ALIAS=$(sl alias create --random 2>/dev/null)`.

**Every command that returns data must support `--json` and `--jq`** (per-command flags, not global):
```go
if createJSON || createJQ != "" {
    if createJQ != "" {
        return output.PrintJQ(rawJSON, createJQ)
    }
    return output.PrintJSON(rawJSON)
}
```

**Errors must be actionable.** Bad: `"invalid ID"`. Good: `"invalid contact ID %q: expected a numeric ID (use 'sl contact list <alias>' to find IDs)"`.

**Non-interactive by default.** `IsInteractive()` gates all prompts. `ConfirmAction()` returns `false` in non-TTY mode. Every interactive prompt must have a flag alternative (`--key`, `--suffix N`, `--yes`). Cancelled operations return an error (non-zero exit), not nil.

**Prefer idempotent operations over toggles.** `enable`/`disable` and `block`/`unblock` check current state first. Agents should use these instead of `toggle`.

**Exit codes:** `0` success, `1` runtime error, `2` usage error (POSIX EX_USAGE — lets agents distinguish bad invocation from API failure).

**Self-documenting via `--help`.** Every command must have `Short`, `Long`, and `Example` fields. Delete commands must support `--yes` to skip confirmation and `--dry-run` to preview.

## AXI Compliance

This CLI follows the [AXI (Agent eXperience Interface)](https://axi.md/) standard for agent-friendly CLIs. Two skills from the [`cli-dev` plugin](https://github.com/mexcool/claude-code-toolkit) are available for auditing and improving CLI design:

- **`/cli-dev:axi`** — Deep AXI spec audit: token efficiency, `--fields`, truncation hints, content-first, contextual disclosure, structured errors, empty states
- **`/cli-dev:cursor-cli-dev`** — Practical checklist: non-interactive flags, `--dry-run`/`--yes`, idempotency, actionable errors, predictable structure, stdin/pipeline support

Use `/cli-dev:axi` for spec-level audits and `/cli-dev:cursor-cli-dev` when designing new commands or reviewing CLI ergonomics. Key AXI patterns already implemented:

- **Content first:** `sl alias`, `sl mailbox`, `sl domain` show live data (not help) when invoked with no subcommand
- **`--fields` flag:** All list commands accept `--fields id,email,...` for table column projection (uses `output.SelectColumns`/`FilterRow`)
- **Truncation with size hints:** `output.Truncate()` appends `... [N]` showing total original length
- **Contextual disclosure:** Mutation commands print a `Hint:` with a logical next-step command via `output.PrintHint()`; empty lists suggest creation commands; paginated lists hint at `--page N` or `--all`
- **Definitive empty states:** Every list command prints an explicit warning when results are empty
- **Structured JSON errors:** When `--json` is active and a command fails, a `{"error":"...","code":N}` envelope is emitted to stdout (handled in `main.go` via `cmd.ExecutedCmd()`)

When adding or modifying commands, maintain these patterns. Run `/cli-dev:axi` periodically to check for regressions.

## Adding a New Command

1. Create `cmd/<resource>/<verb>.go` — define `var xxxCmd` with `Use`, `Short`, `Long`, `Example`, `Args`, `RunE`
2. Define flag vars at package level, register them in `init()`
3. In `RunE`: call `auth.GetAPIKey()`, then `api.NewClient(key, auth.GetAPIBase())`
4. Add `--json` (bool) and `--jq` (string) flags. Branch output on them.
5. For list commands: add `--fields` flag, use `output.SelectColumns`/`FilterRow` for table rendering, handle empty state with `output.PrintWarning`
6. For mutation commands: add `output.PrintHint(...)` after success with a contextual next-step command
7. If the API method doesn't exist yet, add it to `internal/api/client.go` returning `(typed, rawJSON, error)`
8. Register the command in `cmd/<resource>/<resource>.go` via `Cmd.AddCommand(xxxCmd)`
9. Add an `Aliases` field if a short alias makes sense (`ls`, `rm`, `info`)
10. Run `go build ./...`, `go test ./...`, `go vet ./...`

## Non-Obvious Things

- `--json`/`--jq` flags are local per-command, not persistent on the root. Some commands (like `auth login`) intentionally don't have them.
- `api.Verbose` is a package-level var set in `PersistentPreRun`, snapshotted into the client struct at construction. Not a global that changes mid-execution.
- The `SuffixOption` struct has `IsCustom`/`IsPremium` fields. `GetAliasOptions(hostname)` accepts a hostname to tailor suggestions.
- `CreateRandomAlias(note, hostname, mode)` takes 3 string params. `mode` is `"uuid"`, `"word"`, or empty.
- Tests use `httptest.Server` — `newTestClient(t, srv)` in `client_test.go` points the client at a test server.
- GoReleaser builds for linux/darwin/windows on amd64/arm64 with CGO_ENABLED=0. Homebrew tap at `mexcool/homebrew-tap`.
