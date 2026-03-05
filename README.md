# gtm-users

[![CI](https://github.com/h13/gtm-users/actions/workflows/ci.yaml/badge.svg)](https://github.com/h13/gtm-users/actions/workflows/ci.yaml)
[![Go](https://img.shields.io/github/go-mod/go-version/h13/gtm-users)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/h13/gtm-users)](https://goreportcard.com/report/github.com/h13/gtm-users)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Google Tag Manager のユーザー権限を YAML で宣言的に管理する CLI ツール。

## Features

- **Declarative** — YAML ファイルでユーザー権限の desired state を定義
- **Plan & Apply** — Terraform ライクなワークフローで変更を安全に適用
- **Export** — 既存の GTM 権限設定を YAML としてエクスポート
- **Validation** — 設定ファイルの構文・値チェック
- **Additive / Authoritative** — 未管理ユーザーの扱いを2モードで制御

## Install

```sh
go install github.com/h13/gtm-users/cmd/gtm-users@latest
```

## Quick Start

### 1. 設定ファイルを作成

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

### 2. バリデーション

```sh
gtm-users validate --config gtm-users.yaml
```

### 3. 差分を確認

```sh
gtm-users plan --config gtm-users.yaml --credentials sa.json
```

### 4. 適用

```sh
gtm-users apply --config gtm-users.yaml --credentials sa.json
```

## Usage

### validate

設定ファイルの構文とセマンティクスを検証する。

```sh
gtm-users validate --config gtm-users.yaml
```

### export

現在の GTM 権限設定を YAML 形式で出力する。

```sh
gtm-users export --account-id 123456789 --credentials sa.json
```

### plan

desired state と actual state の差分を表示する。変更は適用しない。

```sh
gtm-users plan --config gtm-users.yaml --credentials sa.json --format text
```

`--format json` で JSON 出力も可能。

### apply

plan の内容を GTM API に適用する。

```sh
gtm-users apply --config gtm-users.yaml --credentials sa.json
```

`--auto-approve` で確認プロンプトをスキップ。

## Configuration

| Field | Required | Description |
|-------|----------|-------------|
| `account_id` | Yes | GTM アカウント ID |
| `mode` | No | `additive`（デフォルト）または `authoritative` |
| `users` | Yes | ユーザー権限の配列 |

### Mode

- **additive** — 設定ファイルに記載されたユーザーのみ追加・更新する。未記載のユーザーには影響しない。
- **authoritative** — 設定ファイルに記載されていないユーザーを削除する。

### Account Access

| Value | Description |
|-------|-------------|
| `admin` | アカウント管理者 |
| `user` | 一般ユーザー |
| `noAccess` | アカウントレベルのアクセスなし |

### Container Permission

| Value | Description |
|-------|-------------|
| `read` | 読み取りのみ |
| `edit` | 編集可能 |
| `approve` | 承認可能 |
| `publish` | 公開可能 |

## Authentication

GCP サービスアカウントの認証情報 JSON ファイルを `--credentials` フラグで指定する。
サービスアカウントには `Tag Manager - Manage Users` のスコープが必要。

## License

[MIT](LICENSE)
