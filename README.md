# gtm-users

[![CI](https://github.com/h13/gtm-users/actions/workflows/ci.yaml/badge.svg)](https://github.com/h13/gtm-users/actions/workflows/ci.yaml)
[![Go](https://img.shields.io/github/go-mod/go-version/h13/gtm-users)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/h13/gtm-users)](https://goreportcard.com/report/github.com/h13/gtm-users)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-gtm--users-blue?logo=github)](https://github.com/marketplace/actions/gtm-users)

A CLI tool for declaratively managing Google Tag Manager user permissions via YAML.

## Features

- **Declarative** — Define desired user permissions in a YAML file
- **Plan & Apply** — Terraform-like workflow for safe, reviewable changes
- **Export** — Export existing GTM permissions as YAML
- **Validation** — Syntax and semantic checks on config files
- **Additive / Authoritative** — Two modes to control how unmanaged users are handled

## Install

```sh
go install github.com/h13/gtm-users/cmd/gtm-users@latest
```

## Quick Start

### 1. Create a config file

```yaml
account_id: "123456789"
mode: additive

users:
  - email: alice@example.com
    account_access: admin
    container_access:
      - container_id: "GTM-XXXX"
        permission: publish
  - email: bob@example.com
    account_access: user
    container_access:
      - container_id: "GTM-XXXX"
        permission: read
```

### 2. Validate

```sh
gtm-users validate --config gtm-users.yaml
```

### 3. Preview changes

```sh
gtm-users plan --config gtm-users.yaml --credentials sa.json
```

### 4. Apply

```sh
gtm-users apply --config gtm-users.yaml --credentials sa.json
```

## Usage

### validate

Validate config file syntax and semantics.

```sh
gtm-users validate --config gtm-users.yaml
```

### export

Export current GTM permissions as YAML.

```sh
gtm-users export --account-id 123456789 --credentials sa.json
```

### plan

Show the diff between desired and actual state. No changes are applied.

```sh
gtm-users plan --config gtm-users.yaml --credentials sa.json --format text
```

Use `--format json` for JSON output.

### apply

Apply planned changes to the GTM API.

```sh
gtm-users apply --config gtm-users.yaml --credentials sa.json
```

Use `--auto-approve` to skip the confirmation prompt.

## Configuration

| Field | Required | Description |
|-------|----------|-------------|
| `account_id` | Yes | GTM account ID |
| `mode` | No | `additive` (default) or `authoritative` |
| `users` | Yes | Array of user permission entries |

### Mode

- **additive** — Only add or update users listed in the config. Unmanaged users are left untouched.
- **authoritative** — Remove users not listed in the config.

### Account Access

| Value | Description |
|-------|-------------|
| `admin` | Account administrator |
| `user` | Standard user |
| `noAccess` | No account-level access |

### Container Permission

| Value | Description |
|-------|-------------|
| `read` | Read only |
| `edit` | Edit access |
| `approve` | Approve access |
| `publish` | Publish access |

## GitHub Action

Available on the [GitHub Marketplace](https://github.com/marketplace/actions/gtm-users).

### Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `command` | Yes | — | Command to run (`validate`, `plan`, `apply`, `export`) |
| `config` | No | `gtm-users.yaml` | Path to config file |
| `credentials` | No | — | GCP service account credentials JSON content |
| `format` | No | `text` | Output format (`text`, `json`) |
| `auto-approve` | No | `false` | Skip confirmation prompt for `apply` |
| `account-id` | No | — | GTM account ID (required for `export`) |
| `comment-on-pr` | No | `true` | Post plan output as a PR comment |

### Outputs

| Output | Description |
|--------|-------------|
| `result` | Command output |
| `has-changes` | Whether the plan contains changes (`true`/`false`) |

### Example: Plan on PR, Apply on merge

```yaml
name: GTM Users

on:
  pull_request:
    paths: ["gtm-users.yaml"]
  push:
    branches: [main]
    paths: ["gtm-users.yaml"]

permissions:
  contents: read
  pull-requests: write

jobs:
  plan:
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: h13/gtm-users@v1
        with:
          command: plan
          credentials: ${{ secrets.GCP_SA_KEY }}

  apply:
    if: github.event_name == 'push'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: h13/gtm-users@v1
        with:
          command: apply
          credentials: ${{ secrets.GCP_SA_KEY }}
          auto-approve: "true"
```

## Authentication

Provide a GCP service account credentials JSON file via the `--credentials` flag.
The service account requires the `Tag Manager - Manage Users` scope.

When using the GitHub Action, pass the JSON content directly via the `credentials` input
(typically stored as a [GitHub secret](https://docs.github.com/en/actions/security-for-github-actions/security-guides/using-secrets-in-github-actions)).

## License

[MIT](LICENSE)
