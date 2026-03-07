# rollbar-cli

[![Release](https://img.shields.io/github/v/release/davebarnwell/rollbar-cli?display_name=tag)](https://github.com/davebarnwell/rollbar-cli/releases)
[![CI](https://github.com/davebarnwell/rollbar-cli/actions/workflows/pr-tests.yml/badge.svg?branch=main)](https://github.com/davebarnwell/rollbar-cli/actions/workflows/pr-tests.yml)

Query, triage, and automate [Rollbar](https://rollbar.com) from the terminal.

`rollbar-cli` is a single-binary CLI for engineers who want fast access to Rollbar items, occurrences, deploys,
environments, and users without living in the browser. It supports terminal-friendly views for manual triage and stable
JSON or NDJSON output for scripts, CI jobs, and agent workflows.

## Why use it

- Find active production errors quickly with filtering by status, level, environment, time range, sort order, and
  paging.
- Inspect an item and its occurrences from one command.
- Resolve, mute, assign, snooze, or update items without opening the Rollbar UI.
- Export normalised JSON, raw API JSON, or NDJSON for automation.
- Track deploys and inspect account metadata from the same CLI.
- Install an optional AI skill for Codex, Claude Code, Cursor, Windsurf, and similar tools.

The official [Rollbar CLI](https://github.com/rollbar/rollbar-cli) is primarily aimed at source maps and deployment
workflows. `rollbar-cli` is aimed at querying and triaging Rollbar data during development and incident response.

## Install

### Prebuilt binaries

Download the latest archive for your platform
from [GitHub Releases](https://github.com/davebarnwell/rollbar-cli/releases).

### Go install

```bash
go install github.com/davebarnwell/rollbar-cli@latest
```

### Build from source

```bash
make build
./bin/rollbar-cli --help
```

### Install the optional AI skill

```bash
make install-skill
```

### Install both the CLI and the AI skill

```bash
make install-all
```

`make install` only installs the `rollbar-cli` binary with `go install`. `make install-skill` installs only the AI
skill. Use `make install-all` if you want both.

## Quick start

```bash
export ROLLBAR_ACCESS_TOKEN=rbac_...

# list active production issues
rollbar-cli items list --status active --environment production

# inspect one item with occurrence details
rollbar-cli items get --id 275123456 --instances --json

# take action on an item
rollbar-cli items resolve --id 275123456 --resolved-in-version aabbcc1

# inspect recent deploys
rollbar-cli deploys list --limit 10
```

If you built locally with `make build`, use `./bin/rollbar-cli ...` instead.

## Common workflows

### Triage active errors

```bash
rollbar-cli items list \
  --status active \
  --environment production \
  --level error \
  --last 24h \
  --sort last_occurrence_timestamp_desc
```

### Inspect one item with occurrences

```bash
rollbar-cli items get --id 275123456 --instances --payload summary --payload-section request
```

### Watch during an incident

```bash
rollbar-cli items watch --status active --environment production --interval 30s --count 10
```

### Script against JSON output

```bash
rollbar-cli items list --status active --json
rollbar-cli items list --status active --ndjson --limit 20
rollbar-cli items list --status active --raw-json
```

More examples: [EXAMPLES.md](./EXAMPLES.md)

## Authentication and config

Provide a Rollbar token with the scopes needed for the command you want to run.

- Queries generally need a project token with `read` scope.
- Item updates need a token with `read` and `write` scope.
- `users list` needs an account-scoped token that can read account users.

Configuration sources:

- `--token`
- `ROLLBAR_ACCESS_TOKEN`
- `--config` and `--profile`
- `ROLLBAR_CLI_CONFIG`
- default config file: `~/.config/rollbar-cli/config.json`

Additional environment overrides:

- `ROLLBAR_BASE_URL`
- `ROLLBAR_TIMEOUT`

Example config:

```json
{
  "default_profile": "prod",
  "profiles": {
    "prod": {
      "token": "rbac_...",
      "base_url": "https://api.rollbar.com",
      "timeout": "15s"
    }
  }
}
```

Flag values take precedence over config and environment defaults.

## Output modes

Use the output format that matches the job:

- default text/TUI output for interactive triage
- `--json` for normalized, stable CLI JSON
- `--raw-json` for Rollbar API envelopes
- `--ndjson` for line-oriented pipelines

The item TUI shows item IDs and supports `enter` to load occurrences, `o` to toggle details, `y` to copy the item ID,
and `r` or `m` to resolve or mute the selected row.

## Commands

Top-level command groups:

- `items`
- `occurrences`
- `deploys`
- `environments`
- `users`
- `completion`

For full examples and command patterns, see [EXAMPLES.md](./EXAMPLES.md).

## Shell completion

```bash
rollbar-cli completion bash
rollbar-cli completion zsh
rollbar-cli completion fish
rollbar-cli completion powershell
```

## AI skill

This repository includes an optional skill at `.ai/skills/rollbar-cli/SKILL.md` for agent-driven Rollbar investigation
workflows.

Install it with:

```bash
make install-skill
```

## Development

```bash
make test
make test-cover
make help
```

Contribution guidelines: [CONTRIBUTING.md](./CONTRIBUTING.md)

## License

MIT. See [LICENSE](./LICENSE).
