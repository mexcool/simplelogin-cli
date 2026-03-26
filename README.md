# SimpleLogin CLI (`sl`)

A command-line interface for the [SimpleLogin](https://simplelogin.io) email alias service. Manage aliases, contacts, mailboxes, domains, and settings directly from your terminal.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap mexcool/tap
brew install sl
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
# Stored in ~/.config/simplelogin/config.yml
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

### Verify authentication

```bash
sl auth status
```

### Get your API key

1. Go to https://app.simplelogin.io/dashboard/setting
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

# Toggle an alias on/off
sl alias toggle 12345

# Check account status
sl account status
```

## Command Reference

### `sl auth` — Manage authentication

| Command | Description |
|---------|-------------|
| `sl auth login` | Authenticate with SimpleLogin |
| `sl auth login --key <key>` | Store API key directly |
| `sl auth login --1password --vault <v> --item <i>` | Use 1Password integration |
| `sl auth logout` | Remove stored credentials |
| `sl auth status` | Show current user and masked key |

### `sl alias` — Manage email aliases

| Command | Description |
|---------|-------------|
| `sl alias list` | List aliases (first page) |
| `sl alias list --all` | List all aliases |
| `sl alias list --enabled` | List only enabled aliases |
| `sl alias list --disabled` | List only disabled aliases |
| `sl alias list --pinned` | List only pinned aliases |
| `sl alias list --query <q>` | Search aliases |
| `sl alias list --page <N>` | List specific page |
| `sl alias create --random` | Create a random alias |
| `sl alias create --random --note <n>` | Create random alias with note |
| `sl alias create --prefix <p>` | Create custom alias (interactive suffix) |
| `sl alias create --prefix <p> --suffix <N>` | Create custom alias with specific suffix |
| `sl alias create --prefix <p> --mailbox <id>` | Create alias assigned to mailbox |
| `sl alias view <id-or-email>` | View alias details |
| `sl alias delete <id-or-email>` | Delete an alias |
| `sl alias delete <id-or-email> --yes` | Delete without confirmation |
| `sl alias toggle <id-or-email>` | Enable/disable an alias |
| `sl alias edit <id-or-email> --name <n>` | Set display name |
| `sl alias edit <id-or-email> --note <n>` | Set note |
| `sl alias edit <id-or-email> --pin` | Pin alias |
| `sl alias edit <id-or-email> --unpin` | Unpin alias |
| `sl alias edit <id-or-email> --mailbox <id>` | Set mailbox |
| `sl alias activity <id-or-email>` | View activity log |
| `sl alias activity <id-or-email> --all` | View all activity |

### `sl contact` — Manage alias contacts

| Command | Description |
|---------|-------------|
| `sl contact list <alias-id-or-email>` | List contacts for an alias |
| `sl contact list <alias-id-or-email> --all` | List all contacts |
| `sl contact add <alias-id-or-email> <email>` | Add contact (creates reverse alias) |
| `sl contact delete <contact-id>` | Delete a contact |
| `sl contact toggle <contact-id>` | Block/unblock a contact |

### `sl mailbox` — Manage mailboxes

| Command | Description |
|---------|-------------|
| `sl mailbox list` | List all mailboxes |
| `sl mailbox add <email>` | Add a new mailbox |
| `sl mailbox delete <id>` | Delete a mailbox |
| `sl mailbox delete <id> --transfer-to <id>` | Delete and transfer aliases |
| `sl mailbox edit <id> --default` | Set as default mailbox |
| `sl mailbox edit <id> --email <new-email>` | Change mailbox email |
| `sl mailbox edit <id> --cancel-change` | Cancel pending email change |

### `sl domain` — Manage custom domains

| Command | Description |
|---------|-------------|
| `sl domain list` | List custom domains |
| `sl domain edit <id> --catch-all` | Enable catch-all |
| `sl domain edit <id> --no-catch-all` | Disable catch-all |
| `sl domain edit <id> --random-prefix` | Enable random prefix |
| `sl domain edit <id> --no-random-prefix` | Disable random prefix |
| `sl domain edit <id> --name <n>` | Set display name |
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

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SIMPLELOGIN_API_KEY` | API key (highest priority) |
| `SL_API_KEY` | API key (alternative) |
| `NO_COLOR` | Set to any value to disable colored output |

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

This CLI is designed for both human and programmatic use. Every list/view command supports `--json` and `--jq` flags for machine-readable output.

**Progressive disclosure**: Every command has rich `--help` text with:
- `Short`: One-line description (shown in parent help)
- `Long`: Detailed explanation with context and behavior notes
- `Example`: Real usage examples

This makes the CLI self-documenting — agents and scripts can discover capabilities by reading help text.

**Piping-friendly**: Success/status messages go to stderr, data goes to stdout. This means you can safely pipe output:

```bash
# Create alias and capture just the email
NEW_ALIAS=$(sl alias create --random 2>/dev/null)

# Export and process
sl export data | jq '.aliases | length'
```

## Configuration

Config file location: `~/.config/simplelogin/config.yml`

```yaml
api_key: sl_xxxxxxxxxxxxx
# or for 1Password:
op_ref: op://Personal/SimpleLogin API Key/credential
```

## License

MIT
