# rollbar-cli

`rollbar-cli` is a Go CLI for querying and triaging [Rollbar](https://rollbar.com) items and occurrences.
It supports terminal-friendly views, stable JSON and NDJSON output for automation, and common item actions such as
resolve, mute, assign, snooze, and update.

This repo also includes an AI skill at `.ai/skills/rollbar-cli` so coding agents can use the CLI from natural-language
workflows.

## What it does

- List Rollbar items with environment, level, status, time, sort, and paging filters
- Fetch a single item by ID or UUID, optionally with associated occurrences
- List or fetch occurrences for an item
- Update item status, title, level, assignment, and snooze state
- Render stable JSON, raw API JSON, NDJSON, or text/TUI output
- Install an agent skill for Codex, Claude Code, Cursor, Windsurf, and similar tools

## Why

The main use case is incident triage, both manually and through agent workflows.

For example, a scheduled coding agent can run a prompt like:

> Find all unresolved Rollbar errors from the last 24 hours and create fixes with associated PRs.

If you want that full loop, you will also need a GitHub-oriented skill that can open pull requests, such as
[Yeet](https://github.com/openai/skills/tree/main/skills/.curated/yeet).

The official [Rollbar CLI](https://github.com/rollbar/rollbar-cli) focuses on source map uploads and deployments.
This project is a lightweight alternative for querying and triaging Rollbar data.

## Quick Start

```bash
# build locally
make build

# authenticate for this shell
export ROLLBAR_ACCESS_TOKEN=rbac_...

# list recent active production items
./bin/rollbar-cli items list --status active --environment production

# inspect one item with occurrences
./bin/rollbar-cli items get --id 275123456 --instances --json
```

If you install with `go install` or `make install`, you can run `rollbar-cli ...` directly instead of `./bin/rollbar-cli`.

## Build and Install

```bash
# build binary into bin/
make build

# install skill into common AI tool skill directories
make install-skill

# install skill + CLI with go install
make install
```

Manual build:

```bash
go mod tidy
go build -o rollbar-cli .
```

Show all make targets:

```bash
make help
```

## Authentication and Config

Provide a Rollbar project token with `read` scope for queries, or `read` and `write` if you want to update items.

Supported auth/config inputs:

- `--token`
- `ROLLBAR_ACCESS_TOKEN`
- `--config` and `--profile`
- `ROLLBAR_CLI_CONFIG`
- default config file: `~/.config/rollbar-cli/config.json`

Additional environment overrides:

- `ROLLBAR_BASE_URL`
- `ROLLBAR_TIMEOUT`

Example config file:

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

## AI Skill

This repo ships with a Rollbar skill for agent workflows:

- skill path: `.ai/skills/rollbar-cli/SKILL.md`
- install command: `make install-skill`

The skill documents common triage flows and gives agents a stable set of CLI commands to use when investigating
Rollbar issues.

## Usage

### Items

```bash
# text/TUI output
rollbar-cli items list --status active --environment production

# stable JSON output
rollbar-cli items list --json

# raw Rollbar API JSON
rollbar-cli items list --raw-json

# NDJSON for scripting
rollbar-cli items list --ndjson --limit 20

# page, time, and sort filtering (repeat --level)
rollbar-cli items list --page 2 --pages 3 --level error --level critical --last 24h --sort counter_desc --limit 25

# watch the list during incident triage
rollbar-cli items watch --status active --environment production --interval 30s --count 10
```

```bash
# get a single item by numeric item ID
rollbar-cli items get 275123456
# or
rollbar-cli items get --id 275123456

# get a single item by UUID
rollbar-cli items get 01234567-89ab-cdef-0123-456789abcdef
# or
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef

# get item JSON (stable schema)
rollbar-cli items get --id 275123456 --json

# get item + instance details with payload shaping
rollbar-cli items get --id 275123456 --instances --payload summary --payload-section request

# get item + instances JSON payloads
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef --instances --json
```

```bash
# common task verbs
rollbar-cli items resolve --id 275123456 --resolved-in-version aabbcc1
rollbar-cli items mute --id 275123456
rollbar-cli items assign --id 275123456 --assigned-user-id 321
rollbar-cli items snooze --id 275123456 --duration 1h

# update status + resolved version
rollbar-cli items update --id 275123456 --status resolved --resolved-in-version aabbcc1

# update by UUID and set level/title
rollbar-cli items update 01234567-89ab-cdef-0123-456789abcdef --level error --title "Checkout failure"

# clear assignment and set snooze
rollbar-cli items update --id 275123456 --clear-assigned-user --snooze-enabled true --snooze-expiration-seconds 3600

# update item JSON
rollbar-cli items update --id 275123456 --status active --json
```

### Occurrences

```bash
# list occurrences for an item
rollbar-cli occurrences list --item-id 275123456
# or
rollbar-cli occurrences list 275123456

# list occurrences JSON payload
rollbar-cli occurrences list --item-uuid 01234567-89ab-cdef-0123-456789abcdef --json

# list occurrences in NDJSON
rollbar-cli occurrences list --item-id 275123456 --ndjson

# get one occurrence by numeric occurrence ID
rollbar-cli occurrences get --id 501
# or
rollbar-cli occurrences get 501

# get one occurrence by UUID
rollbar-cli occurrences get --uuid 89abcdef-0123-4567-89ab-cdef01234567

# alias spelling also works
rollbar-cli occurences get --uuid 89abcdef-0123-4567-89ab-cdef01234567

# get one occurrence JSON
rollbar-cli occurrences get --uuid 89abcdef-0123-4567-89ab-cdef01234567 --json
```

## Shell Completion

```bash
# bash, fish, zsh, and powershell are supported
rollbar-cli completion bash
rollbar-cli completion zsh
rollbar-cli completion fish
rollbar-cli completion powershell
```

Examples:

```bash
# bash one-off
source <(rollbar-cli completion bash)

# bash global
rollbar-cli completion bash > rollbar-cli.bash
sudo cp rollbar-cli.bash /etc/bash_completion.d/

# zsh one-off
source <(rollbar-cli completion zsh)

# zsh global
rollbar-cli completion zsh > _rollbar-cli
sudo cp _rollbar-cli /usr/local/share/zsh/site-functions/

# zsh user only
mkdir -p ~/.zsh/completions
rollbar-cli completion zsh > ~/.zsh/completions/_rollbar-cli
```

## Development

```bash
# run unit tests
make test

# run unit tests with coverage
make test-cover

# run vet manually
go vet ./...

# generate an HTML coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Release Automation

Merged pull requests trigger `.github/workflows/release-on-merge.yml`, which now signs and notarizes the macOS release archives on `macos-latest`.

Required GitHub Actions secrets for macOS releases:

- `APPLE_SIGNING_CERTIFICATE_P12_BASE64`: base64-encoded Developer ID Application certificate export
- `APPLE_SIGNING_CERTIFICATE_PASSWORD`: password for the `.p12`
- `APPLE_SIGNING_IDENTITY`: full codesigning identity name, for example `Developer ID Application: Example, Inc. (TEAMID)`
- `APPLE_KEYCHAIN_PASSWORD`: temporary keychain password used during the workflow run
- `APPLE_NOTARY_KEY_ID`: App Store Connect API key ID for notarization
- `APPLE_NOTARY_ISSUER_ID`: App Store Connect issuer ID for notarization
- `APPLE_NOTARY_PRIVATE_KEY_BASE64`: base64-encoded contents of the App Store Connect `.p8` key

macOS assets are published as `.zip` archives instead of `.tar.gz` so they can be submitted to Apple's notarization service before release.

## Notes

- Auth uses the `X-Rollbar-Access-Token` header.
- The default API base URL is `https://api.rollbar.com`.
- `--json` emits normalized CLI JSON, while `--raw-json` preserves Rollbar API envelopes.
- `items list` supports client-side `--since`, `--until`, `--last`, `--sort`, `--limit`, and `--pages` filtering.
- The item TUI shows item IDs and supports `enter` to load occurrences, `o` to toggle details, `y` to copy the item
  ID, and `r` or `m` to resolve or mute the selected row.
