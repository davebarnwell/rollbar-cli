# Contributing to rollbar-cli

Thanks for contributing.

## How to contribute

1. Open a pull request for features, bug fixes, or focused improvements.
2. Keep changes scoped to one concern per PR.
3. Explain what changed and why.
4. Run the relevant tests before opening or updating the PR.

## Local development

### Build

```bash
# build the local binary into bin/
make build

# build common macOS, Linux, and Windows binaries into bin/
make build-cross

# install the CLI with go install and install the AI skill
make install

# install only the AI skill into common tool directories
make install-skill

# show all available make targets
make help
```

### Manual build

```bash
go mod tidy
go build -o rollbar-cli .
```

### Test

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

## Before you open a PR

- Test locally.
- Keep the PR focused.
- Describe the change and the reason for it.
- Call out any follow-up work or known limitations.

## AI-assisted PRs

Built with Codex, Claude, or other AI tools? Mark it clearly.

Please include:

- [ ] Mark the PR as AI-assisted in the title or description
- [ ] Note the degree of testing: untested, lightly tested, or fully tested
- [ ] Include prompts or session logs when practical
- [ ] Confirm that you understand the code being proposed

## Release automation

Merged pull requests trigger `.github/workflows/release-on-merge.yml`, which tags the merged commit and publishes
release archives.

Current behavior:

- Linux and Windows artefacts are always built.
- macOS artefacts are built, signed, and notarised when the required Apple credentials are present.
- macOS assets are published as `.zip` archives so they can be notarised before release.
- If the macOS signing or notarisation secrets are missing, the workflow skips macOS assets and still publishes Linux
  and Windows artefacts.

Required GitHub Actions secrets for macOS releases:

- `APPLE_SIGNING_CERTIFICATE_P12_BASE64`: base64-encoded Developer ID Application certificate export
- `APPLE_SIGNING_CERTIFICATE_PASSWORD`: password for the `.p12`
- `APPLE_SIGNING_IDENTITY`: full codesigning identity name, for example
  `Developer ID Application: Example, Inc. (TEAMID)`
- `APPLE_KEYCHAIN_PASSWORD`: temporary keychain password used during the workflow run
- `APPLE_NOTARY_KEY_ID`: App Store Connect API key ID for notarization
- `APPLE_NOTARY_ISSUER_ID`: App Store Connect issuer ID for notarization
- `APPLE_NOTARY_PRIVATE_KEY_BASE64`: base64-encoded contents of the App Store Connect `.p8` key

## Current focus

- Expose more of the official [Rollbar API](https://docs.rollbar.com/reference/getting-started-1) through the CLI, where
  it improves automation and agent workflows.
