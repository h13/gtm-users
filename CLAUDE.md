# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gtm-users is a declarative CLI tool and GitHub Action for managing Google Tag Manager user permissions. It follows a Terraform-like plan/apply workflow: define desired state in YAML, diff against live GTM state, and apply changes.

## Build & Development Commands

```bash
make build         # Compile binary to bin/gtm-users
make test          # Run tests with race detection
make test-cover    # Tests with coverage report (coverage.out)
make lint          # golangci-lint (v2.10.1)
make vet           # go vet
make clean         # Remove bin/ and coverage.out
```

Single test:

```bash
go test -run TestFunctionName ./internal/config/ -v -race
```

Coverage by function:

```bash
go tool cover -func=coverage.out
```

## Architecture

Pipeline flow: `Config (YAML) → State (normalize) → Diff (plan) → Output (text/json) → GTM API (apply)`

```
cmd/gtm-users/main.go     # Entry point, version injected by GoReleaser
internal/
  cli/                     # Cobra commands: validate, plan, apply, export
  config/                  # YAML parsing with strict unknown-field rejection
  state/                   # Config → normalized AccountState conversion
  diff/                    # Computes Plan with UserChange/ContainerChange
  output/                  # Text (Terraform-like +/~/- symbols) and JSON formatters
  gtm/                     # Google Tag Manager API client with path caching
```

Each package has a single responsibility with minimal cross-package coupling. The `gtm/` package is the only one that talks to external APIs.

## Dependency Injection

Functional options pattern is used for testability without mocks:

- `gtm.NewClient` accepts `gtm.Option` (e.g. `WithAPIOptions`) to inject httptest-backed services
- `cli.NewRootCmd` accepts `cli.CmdOption` (e.g. `WithGTMOptions`) for end-to-end testing
- `rootOptions` holds injectable `stdout`, `stdin`, and `newClient` factory
- Tests use httptest servers as fakes, not mock libraries

## CLI Commands

```bash
gtm-users validate [--config FILE]                           # Offline validation
gtm-users plan    [--config FILE] [--credentials FILE] [--format text|json]
gtm-users apply   [--config FILE] [--credentials FILE] [--auto-approve]
gtm-users export  --account-id ID [--credentials FILE] [--format text|json]
```

## Key Conventions

- **Strict YAML parsing**: `dec.KnownFields(true)` rejects unknown fields
- **Validation errors**: Return `[]ValidationError` with `Field` (e.g., `users[0].email`) and `Message`
- **Deterministic output**: All slices sorted with `slices.SortFunc` before comparison/output
- **Error wrapping**: Always use `fmt.Errorf("context: %w", err)` for chain
- **Email handling**: Case-insensitive comparison, normalized to lowercase
- **Container ID format**: `GTM-` prefix pattern validation
- **Modes**: `additive` (only add/update) vs `authoritative` (also deletes unmanaged users)

## Testing

- 100% coverage target (all 6 internal packages at 100%)
- `cmd/gtm-users/main.go` excluded via `codecov.yml`
- Race detection enabled on all test runs (`-race` flag)
- Fuzz testing for config parsing (`FuzzParse`)
- Test data in `testdata/` directory

## CI Pipeline

GitHub Actions runs: actionlint → lint → test (with coverage gate) → build + vet → govulncheck

## Release

- GoReleaser builds multi-platform binaries on `v*` tags
- Checksums signed with Sigstore cosign (keyless OIDC)
- `checksums.txt.bundle` uploaded to GitHub release

## Branch Protection & Merge Workflow

- main branch requires PR with 1 approving review + all 5 CI checks passing
- `enforce_admins: false` — admin can bypass review requirement via `gh pr merge --admin`
- Linear history required (no merge commits)
- Solo project workflow: create PR → CI passes → `gh pr merge --squash --admin`

## GitHub Action

The project ships as a GitHub Action (`action.yml`). Key behavior:

- `has-changes` output uses `jq` to check plan JSON for non-empty changes
- Auto-posts/updates PR comments with plan output
- Credentials passed as temp file, auto-cleaned
