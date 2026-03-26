# SimpleLogin CLI (`sl`)

[![CI](https://github.com/mexcool/simplelogin-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/mexcool/simplelogin-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/mexcool/simplelogin-cli)](https://github.com/mexcool/simplelogin-cli/releases/latest)
[![Go version](https://img.shields.io/badge/go-1.24-blue?logo=go)](https://golang.org/doc/go1.24)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A command-line interface for the [SimpleLogin](https://simplelogin.io) email alias service. Manage aliases, contacts, mailboxes, domains, and settings directly from your terminal. For you and your agents.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install mexcool/tap/simplelogin-cli
```

### Go install

```bash
go install github.com/mexcool/simplelogin-cli/cmd/sl@latest
```

### GitHub Releases

Download pre-built binaries for your platform from the [Releases](https://github.com/mexcool/simplelogin-cli/releases) page.

### From source

```bash
git clone https://github.com/mexcool/simplelogin-cli.git
cd simplelogin-cli
make build
# Binary is at ./bin/sl
```

## Authentication

The CLI looks for your API key in this order:

### 1. Environment variable (highest priority)

```bash
export SIMPLELOGIN_API_KEY=your_api_key_here
# or
export SL_API_KEY=your_api_key_here
```

### 2. Config file

```bash
sl auth login --key sl_xxxxxxxxxxxxx
# Stored in $XDG_CONFIG_HOME/simplelogin/config.yml (defaults to ~/.config/simplelogin/config.yml)
```

Or interactively:

```bash
sl auth login
# Prompts for your API key
```

### 3. 1Password integration (most secure)

If you use 1Password, the CLI can retrieve your API key on each request without ever storing it on disk:

```bash
sl auth login --1password --vault Personal --item "SimpleLogin API Key"
```

This stores an `op://` reference in the config file. The actual key is fetched via the `op` CLI each time it's needed. You must have the [1Password CLI](https://developer.1password.com/docs/cli/) installed and signed in.

### Self-hosted instances

If you run a self-hosted SimpleLogin instance, pass your URL at login:

```bash
sl auth login --key sl_xxxxxxxxxxxxx --url https://sl.example.com
```

The URL is stored in the config file and used for all subsequent API calls.

### Verify authentication

```bash
sl auth status
```

### Get your API key

1. Go to https://app.simplelogin.io/dashboard/setting (or your self-hosted URL)
2. Scroll to the "API Key" section
3. Generate or copy your API key

## Quick Start

```bash
# Authenticate
sl auth login --key sl_xxxxxxxxxxxxx

# Create a random alias
sl alias create --random

# List your aliases
sl alias list --all

# View alias details
sl alias view my-alias@simplelogin.co

# Enable/disable an alias (idempotent, safe for scripts)
sl alias enable 12345
sl alias disable 12345

# Check account status
sl account status
```

## Version

```bash
sl --version
# sl version 0.2.0 (abc1234, 2026-03-26T20:00:00Z)
```

Binaries from [Releases](https://github.com/mexcool/simplelogin-cli/releases) and Homebrew include the version tag, commit hash, and build date. When installed via `go install`, the Go module version is shown if available; otherwise it falls back to `dev`.

## Command Reference

Most commands support `--json` and `--jq` flags for machine-readable output. Many commands have short aliases (`ls`, `rm`, `info`) for convenience.

### `sl auth` — Manage authentication

| Command | Description |
|---------|-------------|
| `sl auth login` | Authenticate interactively |
| `sl auth login --key <key>` | Store API key directly |
| `sl auth login --key <key> --url <url>` | Authenticate against a self-hosted instance |
| `sl auth login --1password --vault <v> --item <i>` | Use 1Password integration |
| `sl auth logout` | Remove stored credentials |
| `sl auth status` | Show current user, key source, and API URL |

### `sl alias` — Manage email aliases

| Command | Description |
|---------|-------------|
| `sl alias list` (alias: `ls`) | List aliases (first page) |
| `sl alias list --all` | List all aliases |
| `sl alias list --enabled` / `--disabled` / `--pinned` | Filter aliases |
| `sl alias list --query <q>` | Search aliases |
| `sl alias view <id-or-email>` (alias: `info`) | View alias details |
| `sl alias create --random` | Create a random alias |
| `sl alias create --random --mode word` | Create word-based random alias |
| `sl alias create --random --mode uuid` | Create UUID-based random alias |
| `sl alias create --random --hostname github.com` | Associate alias with a website |
| `sl alias create --prefix <p> --suffix <N>` | Create custom alias with specific suffix |
| `sl alias create --prefix <p> --mailbox <id>` | Create alias assigned to mailbox(es) |
| `sl alias options` | Show available suffixes and creation options |
| `sl alias options --hostname github.com` | Show options tailored for a hostname |
| `sl alias enable <id-or-email>` | Ensure alias is enabled (idempotent) |
| `sl alias disable <id-or-email>` | Ensure alias is disabled (idempotent) |
| `sl alias toggle <id-or-email>` | Flip alias enabled/disabled state |
| `sl alias edit <id-or-email> --name <n>` | Set display name |
| `sl alias edit <id-or-email> --note <n>` | Set note |
| `sl alias edit <id-or-email> --pin` / `--unpin` | Pin/unpin alias |
| `sl alias edit <id-or-email> --mailbox <id>` | Set mailbox(es) |
| `sl alias delete <id-or-email>` (alias: `rm`) | Delete an alias (with confirmation) |
| `sl alias delete <id-or-email> --yes` | Delete without confirmation |
| `sl alias delete <id-or-email> --dry-run` | Preview what would be deleted |
| `sl alias activity <id-or-email>` | View activity log |
| `sl alias activity <id-or-email> --all` | View all activity |

### `sl contact` — Manage alias contacts

| Command | Description |
|---------|-------------|
| `sl contact list <alias-id-or-email>` (alias: `ls`) | List contacts for an alias |
| `sl contact list <alias-id-or-email> --all` | List all contacts |
| `sl contact add <alias-id-or-email> <email>` | Add contact (creates reverse alias) |
| `sl contact delete <contact-id>` (alias: `rm`) | Delete a contact |
| `sl contact delete <contact-id> --dry-run` | Preview deletion |
| `sl contact block <contact-id>` | Ensure contact is blocked (idempotent) |
| `sl contact unblock <contact-id>` | Ensure contact is unblocked (idempotent) |
| `sl contact toggle <contact-id>` | Flip block/unblock state |

### `sl mailbox` — Manage mailboxes

| Command | Description |
|---------|-------------|
| `sl mailbox list` (alias: `ls`) | List all mailboxes |
| `sl mailbox add <email>` | Add a new mailbox |
| `sl mailbox delete <id>` (alias: `rm`) | Delete a mailbox |
| `sl mailbox delete <id> --transfer-to <id>` | Delete and transfer aliases |
| `sl mailbox delete <id> --dry-run` | Preview deletion |
| `sl mailbox edit <id> --default` | Set as default mailbox |
| `sl mailbox edit <id> --email <new-email>` | Change mailbox email |
| `sl mailbox edit <id> --cancel-change` | Cancel pending email change |

### `sl domain` — Manage custom domains

| Command | Description |
|---------|-------------|
| `sl domain list` (alias: `ls`) | List custom domains |
| `sl domain view <id>` | View domain details (verification, mailboxes, alias count) |
| `sl domain edit <id> --catch-all` / `--no-catch-all` | Toggle catch-all |
| `sl domain edit <id> --random-prefix` / `--no-random-prefix` | Toggle random prefix |
| `sl domain edit <id> --name <n>` | Set display name |
| `sl domain edit <id> --mailbox <id>` | Assign mailbox(es) to domain |
| `sl domain trash <id>` | View deleted aliases |

### `sl setting` — Manage account settings

| Command | Description |
|---------|-------------|
| `sl setting view` | View current settings |
| `sl setting edit --generator word\|uuid` | Set alias generator mode |
| `sl setting edit --sender-format <fmt>` | Set sender format (AT, A, NAME_ONLY, AT_ONLY, NO_NAME) |
| `sl setting edit --notifications on\|off` | Toggle notifications |
| `sl setting edit --default-domain <d>` | Set default domain for random aliases |
| `sl setting edit --suffix-type word\|random_string` | Set suffix type |
| `sl setting domains` | List available domains for alias creation |

### `sl account` — Account information

| Command | Description |
|---------|-------------|
| `sl account status` | View account info and stats |

### `sl export` — Export data

| Command | Description |
|---------|-------------|
| `sl export data` | Export all data as JSON |
| `sl export data --output <file>` | Export to file |
| `sl export aliases` | Export aliases as CSV |
| `sl export aliases --output <file>` | Export to file |

### `sl completion` — Shell completions

| Command | Description |
|---------|-------------|
| `sl completion bash` | Generate bash completions |
| `sl completion zsh` | Generate zsh completions |
| `sl completion fish` | Generate fish completions |
| `sl completion powershell` | Generate PowerShell completions |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SIMPLELOGIN_API_KEY` | API key (highest priority) |
| `SL_API_KEY` | API key (alternative) |
| `SL_VERBOSE` or `SL_DEBUG` | Set to `1` to log HTTP requests to stderr |
| `NO_COLOR` | Set to any value to disable colored output |

## Global Flags

| Flag | Description |
|------|-------------|
| `--verbose` | Log HTTP requests to stderr (method, URL, status, latency) |
| `--json` | Output as JSON (available on most commands) |
| `--jq <expr>` | Apply jq expression to JSON output |

## Output Formats

### Default (table)

```
$ sl alias list
ID     Email                          Status   Fwd  Block  Reply  Pinned  Note
12345  my-alias@simplelogin.co        enabled  42   3      5      yes     Shopping sites
12346  newsletter@simplelogin.co      enabled  128  0      0      no      Newsletters
```

### JSON (`--json`)

```bash
sl alias list --json
```

Returns the raw API JSON response, pretty-printed.

### jq filtering (`--jq`)

```bash
# Get just alias emails
sl alias list --all --json --jq '.aliases[].email'

# Find high-traffic aliases
sl alias list --all --json --jq '.aliases[] | select(.nb_forward > 100) | {email, nb_forward}'

# Get default mailbox email
sl mailbox list --json --jq '.mailboxes[] | select(.default) | .email'
```

## Agent Usage

This CLI is designed for both human and programmatic use.

**Idempotent operations**: Use `sl alias enable`/`disable` and `sl contact block`/`unblock` instead of `toggle` — they check the current state first and are safe to call repeatedly.

**Exit codes**: `0` for success, `1` for runtime errors, `2` for usage errors (wrong arguments/flags). Scripts can distinguish "bad invocation" from "API failure."

**Self-documenting errors**: Error messages include actionable guidance, e.g., "invalid contact ID: use 'sl contact list <alias>' to find IDs."

**Progressive disclosure**: Every command has rich `--help` text with examples. Agents can discover capabilities by reading help text.

**Piping-friendly**: Status messages go to stderr, data goes to stdout:

```bash
# Create alias and capture just the email
NEW_ALIAS=$(sl alias create --random 2>/dev/null)

# Export and process
sl export data | jq '.aliases | length'
```

**Debugging**: Use `--verbose` or `SL_VERBOSE=1` to see HTTP requests without exposing API keys.

## Configuration

Config file location: `$XDG_CONFIG_HOME/simplelogin/config.yml` (defaults to `~/.config/simplelogin/config.yml`)

```yaml
api_key: sl_xxxxxxxxxxxxx
api_base: https://sl.example.com  # only for self-hosted instances
# or for 1Password:
op_ref: op://Personal/SimpleLogin API Key/credential
```

## Related Projects

Other community-built SimpleLogin CLIs:

- [KennethWussmann/simplelogin-cli](https://github.com/KennethWussmann/simplelogin-cli) — TypeScript/oclif, includes a separate SDK library.
- [joedemcher/simplelogin-cli](https://github.com/joedemcher/simplelogin-cli) — Python/docopt, available on PyPI. Rich interactive prompts with questionary.

## Disclaimer

This is an unofficial, community-developed CLI tool and is not affiliated with, endorsed by, or supported by SimpleLogin or Proton AG. SimpleLogin is a trademark of Proton AG. Please direct any issues with this tool to [this repository's issue tracker](https://github.com/mexcool/simplelogin-cli/issues), not to SimpleLogin support.

## License

[MIT](LICENSE)
