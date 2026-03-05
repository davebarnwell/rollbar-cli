# rollbar-cli

A Go CLI for querying [Rollbar](https://rollbar.com), including item and occurrence workflows, with:

- Cobra command framework
- Rollbar API integration (`GET /api/1/items`)
- JSON output mode
- Charm-powered TUI table output mode

## Why?

Primarily because I wanted to build an AI Agent Skill to allow AI interaction with Rollbar via prompts such as

> Find all unresolved rollbar errors from the last 24 hours and create fixes with associated PRs

When combined with an appropriate GitHub Skill to create PRs etc..
 
NOTE: The official [Rollbar CLI](https://github.com/rollbar/rollbar-cli) only supports source map uploads and deployments.

The Official [Rollbar API docs](https://docs.rollbar.com/reference/getting-started-1).

## Build

```bash
go mod tidy
go build -o rollbar-cli .
```

## Makefile Commands

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

## Testing

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

## Usage

If you've built the CLI with `go install` or `make install` you can run it directly with `rollbar-cli ...` which is
recommended ([see the go documentation](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies))
if you installed locally then replace `rollbar-cli` with `./rollbar-cli` prefix.

```bash
# text/TUI output
rollbar-cli items list --status active --environment production

# JSON output
rollbar-cli items list --json
# or
rollbar-cli items list --output json

# get a single item by numeric item ID
rollbar-cli items get 275123456
# or
rollbar-cli items get --id 275123456

# get a single item by occurrence UUID
rollbar-cli items get 01234567-89ab-cdef-0123-456789abcdef
# or
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef

# get item JSON
rollbar-cli items get --id 275123456 --json

# get item + instance details (stack frames, file/line, payload)
rollbar-cli items get --id 275123456 --instances

# get item + instances JSON payloads
rollbar-cli items get --uuid 01234567-89ab-cdef-0123-456789abcdef --instances --json

# update status + resolved version
rollbar-cli items update --id 275123456 --status resolved --resolved-in-version aabbcc1

# update by UUID and set level/title
rollbar-cli items update 01234567-89ab-cdef-0123-456789abcdef --level error --title "Checkout failure"

# clear assignment and set snooze
rollbar-cli items update --id 275123456 --clear-assigned-user --snooze-enabled true --snooze-expiration-seconds 3600

# update item JSON
rollbar-cli items update --id 275123456 --status active --json

# page and level filtering (repeat --level)
rollbar-cli items list --page 2 --level error --level critical

# list occurrences for an item
rollbar-cli occurrences list --item-id 275123456
# or
rollbar-cli occurrences list 275123456

# list occurrences JSON payload
rollbar-cli occurrences list --item-uuid 01234567-89ab-cdef-0123-456789abcdef --json

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
