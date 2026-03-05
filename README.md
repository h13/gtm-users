# gtm-users

[![CI](https://github.com/h13/gtm-users/actions/workflows/ci.yaml/badge.svg)](https://github.com/h13/gtm-users/actions/workflows/ci.yaml)
[![Go](https://img.shields.io/github/go-mod/go-version/h13/gtm-users)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/h13/gtm-users)](https://goreportcard.com/report/github.com/h13/gtm-users)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

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

## Authentication

Provide a GCP service account credentials JSON file via the `--credentials` flag.
The service account requires the `Tag Manager - Manage Users` scope.

## License

[MIT](LICENSE)
