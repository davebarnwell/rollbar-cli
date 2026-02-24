# rollbar-cli

A Go CLI for querying [Rollbar](https://rollbar.com), initially just items, with:

- Cobra command framework
- Rollbar API integration (`GET /api/1/items`)
- JSON output mode
- Charm-powered TUI table output mode

## Why?

Primarily because I wanted to build an AI skill to allow AI interaction with Rollbar via prompts such as
"Look at unresolved rollbar errors from the last 24 hours and create fixes with associated PRs".
And the existing [Rollbar CLI](https://github.com/rollbar/rollbar-cli) did not cut it and appears abandoned.

## Build

```bash
go mod tidy
go build -o rollbar-cli .
```

## Testing

```bash
# run all unit tests
go test ./...

# run tests with package coverage
go test ./... -cover

# generate an HTML coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Authentication

Provide a Rollbar project token with `read` scope:

- flag: `--token`
- or env var: `ROLLBAR_ACCESS_TOKEN`

## Usage

```bash
# text/TUI output
./rollbar-cli items list --status active --environment production

# JSON output
./rollbar-cli items list --json
# or
./rollbar-cli items list --output json

# get a single item by numeric item ID
./rollbar-cli items get 275123456
# or
./rollbar-cli items get --id 275123456

# get a single item by occurrence UUID
./rollbar-cli items get 01234567-89ab-cdef-0123-456789abcdef
# or
./rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef

# get item JSON
./rollbar-cli items get --id 275123456 --json

# update status + resolved version
./rollbar-cli items update --id 275123456 --status resolved --resolved-in-version aabbcc1

# update by UUID and set level/title
./rollbar-cli items update 01234567-89ab-cdef-0123-456789abcdef --level error --title "Checkout failure"

# clear assignment and set snooze
./rollbar-cli items update --id 275123456 --clear-assigned-user --snooze-enabled true --snooze-expiration-seconds 3600

# update item JSON
./rollbar-cli items update --id 275123456 --status active --json

# page and level filtering (repeat --level)
./rollbar-cli items list --page 2 --level error --level critical
```

## Notes

- Uses `X-Rollbar-Access-Token` header for auth.
- Base API URL defaults to `https://api.rollbar.com` and can be overridden with `--base-url`.
