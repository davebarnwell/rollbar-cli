# Rollbar Recent Issues

Use this skill to quickly find the most recent Rollbar issues with `rollbar-cli`.

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
2. Open top counters/IDs with `rollbar-cli items get --instances` to include stack frames, exact file/line, and payload details.
3. Update state using `rollbar-cli items update` when triaged.

## Example Follow-up Commands

```bash
rollbar-cli items get --id 275123456 --json
rollbar-cli items get --id 275123456 --instances
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef --instances --instances-page 1 --json
rollbar-cli items update --id 275123456 --status resolved --resolved-in-version aabbcc1
```

## Item Detail With Instances

Use `--instances` on `items get` when you need full occurrence context.

```bash
# text output: item details + instance details
rollbar-cli items get --id 275123456 --instances

# JSON output: returns both item and instances response payloads
rollbar-cli items get --id 275123456 --instances --json
```
