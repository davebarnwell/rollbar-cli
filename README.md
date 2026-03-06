# rollbar-cli

Enables an AI agent to query Rollbar and create bug fix PRs. You can also query manually too.

The CLI is for querying [Rollbar](https://rollbar.com) to triage error items and occurrences, with an associated
Agentic Skill to expose the CLI tool usage to the coding agent, built with:

- [Cobra command framework](https://cobra.dev)
- [Rollbar API](https://docs.rollbar.com/reference/getting-started-1) integration (`GET /api/1/items`)
- stable JSON and NDJSON output modes
- [Charm-powered TUI](https://charm.land) table output mode

## Why?

Using a scheduled/automation task of codex (or claude code), the prompt below and the provided agent skill in this repo
enables an AI agent to daily query Rollbar and create bug fix PRs automatically.

> Find all unresolved rollbar errors from the last 24 hours and create fixes with associated PRs

You'll also need an appropriate GitHub Skill installed to interact with PRs, such
as [Yeet](https://github.com/openai/skills/tree/main/skills/.curated/yeet).

It's magical to have this scheduled daily for 6am and start the work day reviewing bug fix PRs.

NOTE: The official [Rollbar CLI](https://github.com/rollbar/rollbar-cli) only supports source map uploads and
deployments, hence this CLI is a minimal alternative.

## Build It

```bash
go mod tidy
go build -o rollbar-cli .
```

## Makefile It (build, install, etc.)

```bash
# show available targets
make help

# build binary into bin/
make build

# install skill into common AI tool skill directories
make install-skill

# install skill + install CLI with go install
make install

# run unit tests
make test

# run unit tests with coverage
make test-cover

# remove build artifacts
make clean
```

## Testing It

```bash
# vet code
go vet ./...

# run all unit tests
go test ./...

# run tests with package coverage
go test ./... -cover

# generate an HTML coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Authentication

Provide a Rollbar project token with `read` scope (or `read` and `write` if you want to update items):

- flag: `--token`
- or env var: `ROLLBAR_ACCESS_TOKEN`

### Config profiles (optional)

Optional config profiles are supported via `--config` / `--profile`, `ROLLBAR_CLI_CONFIG`, or
`~/.config/rollbar-cli/config.json`:

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

## Install shell tab completion

```bash
# install bash completion
# bash one off
source <(rollbar-cli completion bash)

# OR bash global
rollbar-cli completion bash > rollbar-cli.bash
sudo cp rollbar-cli.bash /etc/bash_completion.d/

# install zsh completion
# zsh one off
source <(rollbar-cli completion zsh)

# OR zsh global
rollbar-cli completion zsh > _rollbar-cli
sudo cp _rollbar-cli /usr/local/share/zsh/site-functions/

# OR zsh user only
mkdir -p ~/.zsh/completions
rollbar-cli completion zsh > ~/.zsh/completions/_rollbar-cli
```

## Usage

If you've built the CLI with `go install` or `make install` you can run it directly with `rollbar-cli ...` which is
recommended ([see the go documentation](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies))
if you installed locally then replace `rollbar-cli` with `./rollbar-cli` prefix.

```bash
# text/TUI output
rollbar-cli items list --status active --environment production

# stable JSON output
rollbar-cli items list --json

# raw Rollbar API JSON
rollbar-cli items list --raw-json

# NDJSON for scripting
rollbar-cli items list --ndjson --limit 20

# get a single item by numeric item ID
rollbar-cli items get 275123456
# or
rollbar-cli items get --id 275123456

# get a single item by occurrence UUID
rollbar-cli items get 01234567-89ab-cdef-0123-456789abcdef
# or
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef

# get item JSON (stable schema)
rollbar-cli items get --id 275123456 --json

# get item + instance details with payload shaping
rollbar-cli items get --id 275123456 --instances --payload summary --payload-section request

# get item + instances JSON payloads
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef --instances --json

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

# page, time, and sort filtering (repeat --level)
rollbar-cli items list --page 2 --pages 3 --level error --level critical --last 24h --sort counter_desc --limit 25

# watch the list during incident triage
rollbar-cli items watch --status active --environment production --interval 30s --count 10

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

## Notes

- Uses `X-Rollbar-Access-Token` header for auth.
- Base API URL defaults to `https://api.rollbar.com` and can be overridden with `--base-url`.
- `--json` emits normalized, stable CLI JSON; `--raw-json` preserves Rollbar API envelopes.
- `items list` supports client-side `--since`, `--until`, `--last`, `--sort`, `--limit`, and `--pages`.
- The item TUI now shows item IDs and supports `enter` to load occurrences, `o` to toggle details, `y` to copy the item
  ID, and `r` / `m` to resolve or mute the selected row.
