# Rollbar CLI Command Reference

Read this reference when exact command flags are needed. Prefer `--json` or `--ndjson` for analysis and `--raw-json`
only when the normalised output omits required evidence.

## Items

```bash
# Active production errors
rollbar-cli items list \
  --status active \
  --environment production \
  --level error \
  --level critical \
  --last 24h \
  --sort counter_desc \
  --limit 25 \
  --json

# All statuses in the same window
rollbar-cli items list \
  --status all \
  --environment production \
  --level error \
  --level critical \
  --last 24h \
  --json

# Pagination and streaming formats
rollbar-cli items list --status active --page 2 --pages 3 --json
rollbar-cli items list --status active --limit 20 --ndjson
rollbar-cli items list --status active --raw-json

# One item by ID or occurrence UUID
rollbar-cli items get 275123456 --json
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef --json

# Instances and a narrowed payload summary
rollbar-cli items get --id 275123456 --instances --json
rollbar-cli items get \
  --id 275123456 \
  --instances \
  --payload summary \
  --payload-section request
rollbar-cli items get --id 275123456 --instances --instances-page 2 --json
```

## Occurrences

```bash
rollbar-cli occurrences list --item-id 275123456 --json
rollbar-cli occurrences list --item-uuid 01234567-89ab-cdef-0123-456789abcdef --ndjson
rollbar-cli occurrences get --id 501 --json
rollbar-cli occurrences get --uuid 89abcdef-0123-4567-89ab-cdef01234567 --json
```

The historical alias `occurences` is accepted, but use the correctly spelt `occurrences` in new commands.

## Environments And Deploys

```bash
rollbar-cli environments list --json
rollbar-cli environments list --fields environment,project_id --no-headers

rollbar-cli deploys list --page 1 --limit 20 --json
rollbar-cli deploys get --id 12345 --json
```

Create or update deploy records only when the user explicitly asks:

```bash
rollbar-cli deploys create \
  --environment production \
  --revision aabbcc1 \
  --status started \
  --comment "Deploy started from CI" \
  --local-username ci-bot \
  --json

rollbar-cli deploys update --id 12345 --status succeeded --json
```

## Users And Assignment

`users list` requires a token that can read account users.

```bash
rollbar-cli users list --json
rollbar-cli users get --id 7 --json
rollbar-cli users list --fields id,username,email --no-headers
```

Mutate an item only with explicit user authorisation:

```bash
rollbar-cli items resolve --id 275123456 --resolved-in-version aabbcc1 --json
rollbar-cli items mute --id 275123456 --json
rollbar-cli items assign --id 275123456 --assigned-user-id 321 --assigned-team-id 88 --json
rollbar-cli items snooze --id 275123456 --duration 1h --json

rollbar-cli items update \
  --id 275123456 \
  --status resolved \
  --resolved-in-version aabbcc1 \
  --level error \
  --title "Checkout failure" \
  --json
```

## Watching And Narrowing

```bash
rollbar-cli items watch \
  --status active \
  --environment production \
  --interval 30s \
  --count 10

rollbar-cli items list --status active --json \
  | jq '.items
      | sort_by(.last_occurrence_timestamp // 0)
      | reverse
      | .[:10]'
```
