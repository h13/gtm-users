# gtm-users

Declarative Google Tag Manager permission management — plan, review, apply.

[![CI](https://github.com/h13/gtm-users/actions/workflows/ci.yaml/badge.svg)](https://github.com/h13/gtm-users/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/h13/gtm-users/graph/badge.svg)](https://codecov.io/gh/h13/gtm-users)
[![Go](https://img.shields.io/github/go-mod/go-version/h13/gtm-users)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/h13/gtm-users)](https://goreportcard.com/report/github.com/h13/gtm-users)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/h13/gtm-users/badge)](https://scorecard.dev/viewer/?uri=github.com/h13/gtm-users)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/12217/badge)](https://www.bestpractices.dev/projects/12217)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-gtm--users-blue?logo=github)](https://github.com/marketplace/actions/gtm-users)

## Features

- **Plan & Apply** — Terraform-like workflow for safe, reviewable changes
- **Roles & Policy** — Reusable permission templates with admin limits and approval rules
- **Drift Detection** — Detect unauthorized changes via CI cron
- **Color Diff** — Green/yellow/red output for add/update/delete
- **Config Includes** — Split configs across files for large organizations
- **Backup** — Save current GTM state as YAML snapshots
- **Permission Matrix** — User × container permission overview
- **GitHub Action** — Plan on PR, apply on merge with PR comments

## Install

### Homebrew

```sh
brew install h13/tap/gtm-users
```

### Go

```sh
go install github.com/h13/gtm-users/cmd/gtm-users@latest
```

### Shell completion

```sh
source <(gtm-users completion bash)   # bash
source <(gtm-users completion zsh)    # zsh
gtm-users completion fish | source    # fish
```

## Quick Start

### 1. Bootstrap from existing GTM state

```sh
gtm-users init --account-id 123456789 --credentials sa.json
```

### 2. Or create a config file manually

```yaml
account_id: "123456789"
mode: additive

roles:
  viewer:
    account_access: user
    container_access:
      - container_id: "GTM-XXXX"
        permission: read
  publisher:
    account_access: user
    container_access:
      - container_id: "GTM-XXXX"
        permission: publish

policy:
  max_admins: 2
  require_approve_for_publish:
    - "GTM-XXXX"

users:
  - email: alice@example.com
    role: publisher
  - email: bob@example.com
    role: viewer
  - email: carol@example.com
    account_access: admin
```

### 3. Validate

```sh
gtm-users validate --config gtm-users.yaml
```

### 4. Preview changes

```sh
gtm-users plan --config gtm-users.yaml --credentials sa.json
```

### 5. Apply

```sh
gtm-users apply --config gtm-users.yaml --credentials sa.json
```

## Commands

| Command      | Description                                       |
| ------------ | ------------------------------------------------- |
| `validate`   | Check config syntax and semantics offline         |
| `plan`       | Show diff between desired and actual state        |
| `apply`      | Apply planned changes to the GTM API              |
| `export`     | Export current GTM permissions as YAML            |
| `init`       | Bootstrap a config file from current GTM state    |
| `drift`      | Detect configuration drift (exit code 2 if found) |
| `matrix`     | Display user × container permission matrix        |
| `backup`     | Save current GTM state as a YAML snapshot         |
| `completion` | Generate shell completion scripts                 |

### Global flags

| Flag            | Default          | Description                                  |
| --------------- | ---------------- | -------------------------------------------- |
| `--config`      | `gtm-users.yaml` | Path to config file                          |
| `--credentials` | —                | Path to GCP service account credentials JSON |
| `--format`      | `text`           | Output format (`text` or `json`)             |
| `--no-color`    | `false`          | Disable colored output                       |

## Configuration

| Field        | Required | Description                                            |
| ------------ | -------- | ------------------------------------------------------ |
| `account_id` | Yes      | GTM account ID                                         |
| `mode`       | No       | `additive` (default) or `authoritative`                |
| `includes`   | No       | List of config files to include                        |
| `roles`      | No       | Reusable permission templates                          |
| `policy`     | No       | Validation rules (admin limits, approval requirements) |
| `users`      | Yes      | Array of user permission entries                       |

### Mode

- **additive** — Only add or update users listed in the config. Unmanaged users are left untouched.
- **authoritative** — Remove users not listed in the config.

### Roles

Define reusable permission templates and assign them to users:

```yaml
roles:
  editor:
    account_access: user
    container_access:
      - container_id: "GTM-XXXX"
        permission: edit

users:
  - email: alice@example.com
    role: editor
```

Users can override role fields by specifying them inline.

### Policy

Enforce organizational rules:

```yaml
policy:
  max_admins: 2 # limit admin count
  require_approve_for_publish:
    - "GTM-XXXX" # publish requires approve
```

### Config includes

Split large configs across files:

```yaml
includes:
  - roles/common.yaml
  - teams/engineering.yaml
```

Included roles merge with the main config (main overrides on conflict). Users are appended.

### Account access

| Value      | Description             |
| ---------- | ----------------------- |
| `admin`    | Account administrator   |
| `user`     | Standard user           |
| `noAccess` | No account-level access |

### Container permission

| Value     | Description    |
| --------- | -------------- |
| `read`    | Read only      |
| `edit`    | Edit access    |
| `approve` | Approve access |
| `publish` | Publish access |

## Drift detection

Run `drift` on a schedule to detect unauthorized changes:

```yaml
# .github/workflows/drift.yaml
on:
  schedule:
    - cron: "0 9 * * 1-5"

jobs:
  drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: h13/gtm-users@v1
        with:
          command: drift
          credentials: ${{ secrets.GCP_SA_KEY }}
```

Exit code 2 indicates drift was detected.

## GitHub Action

Available on the [GitHub Marketplace](https://github.com/marketplace/actions/gtm-users).

### Example: Plan on PR, apply on merge

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

## GitLab CI

```yaml
stages:
  - validate
  - deploy

variables:
  GTM_USERS_VERSION: "v1.2.0"

.gtm-users:
  image: golang:latest
  before_script:
    - go install github.com/h13/gtm-users/cmd/gtm-users@$GTM_USERS_VERSION

plan:
  extends: .gtm-users
  stage: validate
  script:
    - gtm-users plan --config gtm-users.yaml --credentials "$GCP_CREDENTIALS_FILE"
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"

apply:
  extends: .gtm-users
  stage: deploy
  script:
    - gtm-users apply --config gtm-users.yaml --credentials "$GCP_CREDENTIALS_FILE" --auto-approve
  rules:
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
      changes:
        - gtm-users.yaml
```

## Authentication

Provide a GCP service account credentials JSON file via the `--credentials` flag.
The service account requires the `Tag Manager - Manage Users` scope.

When using the GitHub Action, pass the JSON content directly via the `credentials` input
(typically stored as a [GitHub secret](https://docs.github.com/en/actions/security-for-github-actions/security-guides/using-secrets-in-github-actions)).

## License

[MIT](LICENSE)
