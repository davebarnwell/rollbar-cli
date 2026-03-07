---
name: rollbar-cli
description: >-
  Investigates production errors and deploy regressions through the Rollbar CLI. 
  Use when debugging exceptions, incident spikes, fingerprint groups,
  affected-user impact, environment-specific failures, release correlation, or
  when the user mentions Rollbar, occurrences, items, traces, regressions, or
  error monitoring.
---

# Rollbar Recent Issues

Use this skill to quickly find and triage Rollbar issues with `rollbar-cli`.

## When To Use

- You need a fast view of current active issues.
- You want recent issues in stable JSON or NDJSON for automation or triage notes.
- You want to narrow by environment and severity level.
- You need to inspect raw occurrences for a specific item or fetch one occurrence directly.
- You need to look up Rollbar account users before assigning an item.

## Prerequisites

- `rollbar-cli` is installed or available from this repo.
- Auth token is set:
    - `export ROLLBAR_ACCESS_TOKEN=...`
    - or pass `--token ...`
- `users list` uses the account-level users endpoint, so the token must be able to read account users.
- Optional config profiles are supported via `--config`, `--profile`, `ROLLBAR_CLI_CONFIG`, or `~/.config/rollbar-cli/config.json`.

## Core Commands

### 1) Recent active issues (text/table)

```bash
rollbar-cli items list --status active --output text
```

### 2) Recent active issues (JSON)

```bash
rollbar-cli items list --status active --json
```

### 3) Raw API JSON or NDJSON for scripting

```bash
# raw Rollbar envelope
rollbar-cli items list --status active --raw-json

# normalized NDJSON
rollbar-cli items list --status active --ndjson --limit 20
```

### 4) Filter by environment, level, and time window

```bash
rollbar-cli items list \
  --status active \
  --environment production \
  --level error \
  --level critical \
  --last 24h \
  --sort counter_desc \
  --limit 25 \
  --json
```

### 5) Next pages of recent issues

```bash
rollbar-cli items list --status active --page 2 --pages 3 --json
```

### 6) Get one item by ID or UUID

```bash
# by item id
rollbar-cli items get 275123456 --json
# or
rollbar-cli items get --id 275123456 --json

# by occurrence UUID
rollbar-cli items get 01234567-89ab-cdef-0123-456789abcdef --json
# or
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef --json
```

### 7) Get item with instances and shaped payload

```bash
# stable JSON
rollbar-cli items get --id 275123456 --instances --json

# text output with request-only payload summary
rollbar-cli items get \
  --id 275123456 \
  --instances \
  --payload summary \
  --payload-section request

# fetch a specific instances page
rollbar-cli items get --id 275123456 --instances --instances-page 2 --json
```

### 8) Update item status/title/level

```bash
rollbar-cli items update --id 275123456 \
  --status resolved \
  --resolved-in-version aabbcc1 \
  --level error \
  --title "Checkout failure" \
  --json
```

### 9) Task-shaped item actions

```bash
rollbar-cli items resolve --id 275123456 --resolved-in-version aabbcc1 --json
rollbar-cli items mute --id 275123456 --json
rollbar-cli items assign --id 275123456 --assigned-user-id 321 --assigned-team-id 88 --json
rollbar-cli items snooze --id 275123456 --duration 1h --json
```

### 10) Update assignment/team/snooze via generic update

```bash
# clear assignment and snooze for 1 hour
rollbar-cli items update --id 275123456 \
  --clear-assigned-user \
  --clear-assigned-team \
  --snooze-enabled true \
  --snooze-expiration-seconds 3600 \
  --json
```

### 11) List occurrences for an item

```bash
# by item id
rollbar-cli occurrences list --item-id 275123456 --json
# or positional id-or-uuid
rollbar-cli occurrences list 275123456 --json

# by item uuid
rollbar-cli occurrences list --item-uuid 01234567-89ab-cdef-0123-456789abcdef --json

# NDJSON for downstream tooling
rollbar-cli occurrences list --item-id 275123456 --ndjson
```

### 12) Get one occurrence by ID or UUID

```bash
# by occurrence id
rollbar-cli occurrences get --id 501 --json
# or positional id-or-uuid
rollbar-cli occurrences get 501 --json

# by occurrence uuid
rollbar-cli occurrences get --uuid 89abcdef-0123-4567-89ab-cdef01234567 --json

# supported alias spelling
rollbar-cli occurences get --uuid 89abcdef-0123-4567-89ab-cdef01234567 --json
```

### 13) List account users

```bash
# default text output
rollbar-cli users list

# stable JSON
rollbar-cli users list --json

# raw Rollbar envelope
rollbar-cli users list --raw-json

# NDJSON for downstream tooling
rollbar-cli users list --ndjson

# narrow text columns
rollbar-cli users list --fields id,username,email --no-headers
```

### 14) Get one user by ID

```bash
# positional id
rollbar-cli users get 7 --json

# or explicit flag
rollbar-cli users get --id 7

# NDJSON for downstream tooling
rollbar-cli users get --id 7 --ndjson

# raw Rollbar envelope
rollbar-cli users get --id 7 --raw-json
```

## Optional: Watch Active Issues During Triage

```bash
rollbar-cli items watch \
  --status active \
  --environment production \
  --interval 30s \
  --count 10
```

## Optional: Show Top N Most Recent With `jq`

```bash
rollbar-cli items list --status active --json \
| jq '.items
      | sort_by(.last_occurrence_timestamp // 0)
      | reverse
      | .[:10]'
```

## Triage Workflow

1. Start with production + `error`/`critical`.
2. Narrow with `--last`, `--since`, `--sort`, and `--limit`.
3. Open top counters/IDs with `rollbar-cli items get --instances` for stack context.
4. Use `rollbar-cli occurrences list` when you want to inspect occurrence-level payloads for an item.
5. Use `rollbar-cli users list` to find candidate assignee IDs before assigning items.
6. Use `items resolve|mute|assign|snooze` for common triage actions.

## Example Follow-up Commands

```bash
rollbar-cli items list --status active --environment production --last 24h --sort counter_desc --limit 10 --json
rollbar-cli items get --id 275123456 --instances --payload summary --payload-section request
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef --instances --raw-json
rollbar-cli occurrences list --item-id 275123456 --ndjson
rollbar-cli occurrences get --uuid 89abcdef-0123-4567-89ab-cdef01234567 --json
rollbar-cli users list --json
rollbar-cli users get --id 7 --json
rollbar-cli items resolve --id 275123456 --resolved-in-version aabbcc1
rollbar-cli items assign --uuid 01234567-89ab-cdef-0123-456789abcdef --assigned-user-id 321
```
