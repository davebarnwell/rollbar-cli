---
name: rollbar-cli
description: >-
  Investigates production errors and deploy regressions through the Rollbar CLI
  server. Use when debugging exceptions, incident spikes, fingerprint groups,
  affected-user impact, environment-specific failures, release correlation, or
  when the user mentions Rollbar, occurrences, items, traces, regressions, or
  error monitoring.
---

# Rollbar Recent Issues

Use this skill to quickly find and triage Rollbar issues with `rollbar-cli`.

## When To Use

- You need a fast view of current active issues.
- You want recent issues in JSON for automation or triage notes.
- You want to narrow by environment and severity level.

## Prerequisites

- `rollbar-cli` is installed or available from this repo.
- Auth token is set:
  - `export ROLLBAR_ACCESS_TOKEN=...`
  - or pass `--token ...`

## Core Commands

### 1) Recent active issues (text/table)

```bash
rollbar-cli items list --status active --output text
```

### 2) Recent active issues (JSON)

```bash
rollbar-cli items list --status active --json
```

### 3) Filter by environment and level

```bash
rollbar-cli items list \
  --status active \
  --environment production \
  --level error \
  --level critical \
  --json
```

### 4) Next page of recent issues

```bash
rollbar-cli items list --status active --page 2 --json
```

### 5) Get one item by ID or UUID

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

### 6) Get item with instances (stack frames, payload)

```bash
rollbar-cli items get --id 275123456 --instances --json

# fetch a specific instances page
rollbar-cli items get --id 275123456 --instances --instances-page 2 --json
```

### 7) Update item status/title/level

```bash
rollbar-cli items update --id 275123456 \
  --status resolved \
  --resolved-in-version aabbcc1 \
  --level error \
  --title "Checkout failure" \
  --json
```

### 8) Update assignment/team/snooze

```bash
# assign to user + team
rollbar-cli items update --id 275123456 --assigned-user-id 321 --assigned-team-id 88 --json

# clear assignment and snooze for 1 hour
rollbar-cli items update --id 275123456 \
  --clear-assigned-user \
  --clear-assigned-team \
  --snooze-enabled true \
  --snooze-expiration-seconds 3600 \
  --json
```

## Optional: Show Top N Most Recent With `jq`

```bash
rollbar-cli items list --status active --json \
| jq '.result.items
      | sort_by(.last_occurrence_timestamp // 0)
      | reverse
      | .[:10]'
```

## Triage Workflow

1. Start with production + `error`/`critical`.
2. Open top counters/IDs with `rollbar-cli items get` and include `--instances` for stack context.
3. Update state/assignment using `rollbar-cli items update` when triaged.

## Example Follow-up Commands

```bash
rollbar-cli items get --id 275123456 --json
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef --instances --json
rollbar-cli items update --id 275123456 --status resolved --resolved-in-version aabbcc1
rollbar-cli items update --uuid 01234567-89ab-cdef-0123-456789abcdef --level error --title "Checkout failure"
```
