# asana-cli

Go implementation of an Asana CLI with:

- Legacy command aliases for index-based workflows
- Full `Tasks / Projects / Users` endpoint command surface
- OAuth + PAT authentication
- AWS CLI style profile switching
- Automation-friendly JSON output

## Install (local)

```bash
go install ./cmd/asana
```

Or build directly:

```bash
go build -o asana ./cmd/asana
```

## Install (Homebrew)

Tap this repository and install the formula:

```bash
brew tap cloudnative-co/asana-cli https://github.com/cloudnative-co/asana-cli
brew install cloudnative-co/asana-cli/asana
```

If `asana` is already linked from another source, run:

```bash
brew unlink asana || true
brew link --overwrite asana
```

### Stable formula updates

This repository includes `Formula/asana.rb` and a helper script:

```bash
./scripts/update-homebrew-formula.sh v0.1.0
```

The script downloads the release tarball and updates `url` + `sha256` in the formula.

To install the latest development snapshot:

```bash
brew install --HEAD cloudnative-co/asana-cli/asana
```

See also: [docs/homebrew.md](docs/homebrew.md)

## Quick Start

1. Import PAT:

```bash
asana auth import-pat --profile default --pat "$ASANA_PAT"
```

2. Configure workspace:

```bash
asana config --profile default --workspace <workspace_gid>
```

3. List tasks:

```bash
asana tasks
```

## OAuth Login

```bash
asana auth login \
  --profile default \
  --client-id <client_id> \
  --client-secret <client_secret> \
  --redirect-uri <redirect_uri_registered_in_asana_app>
```

### redirect_uri の指定ルール

- `--redirect-uri` は **Asana Developer Console の OAuth Redirect URLs に登録済みの値と完全一致** が必要
- 1文字でも違うとブラウザで `invalid_request: The redirect_uri parameter does not match...` になる

CLI用途の例（アプリ側で同じ値を登録している場合）:

```bash
asana auth login \
  --profile default \
  --client-id "$ASANA_MCP_CLIENT_ID" \
  --client-secret "$ASANA_MCP_CLIENT_SECRET" \
  --redirect-uri "urn:ietf:wg:oauth:2.0:oob"
```

OOBを使わず Web callback を登録しているアプリなら、その登録済みURLをそのまま指定:

```bash
asana auth login \
  --profile default \
  --client-id "$ASANA_MCP_CLIENT_ID" \
  --client-secret "$ASANA_MCP_CLIENT_SECRET" \
  --redirect-uri "https://<your-registered-callback>"
```

### `forbidden_scopes` が出る場合

- アプリ側で許可されていないスコープを要求すると `forbidden_scopes` になる
- このCLIは `--scopes` 未指定時、アプリの既定スコープを使う
- 必要な場合のみ、アプリで許可済みのスコープだけを明示指定する

```bash
asana auth login \
  --profile default \
  --client-id "$ASANA_MCP_CLIENT_ID" \
  --client-secret "$ASANA_MCP_CLIENT_SECRET" \
  --redirect-uri "urn:ietf:wg:oauth:2.0:oob" \
  --scopes "tasks:read,users:read,workspaces:read"
```

### Asanaアプリ側の OAuth scope 設定チェック

Asana Developer Console の **OAuth > Permission scopes** で、少なくとも以下を有効化:

- `tasks:read`
- `tasks:write`
- `tasks:delete`（`task delete` を使う場合）
- `projects:read`
- `projects:write`
- `projects:delete`（`project delete` を使う場合）
- `users:read`
- `stories:read`（タスク履歴表示）
- `stories:write`（コメント投稿）
- `attachments:read`（添付参照/ダウンロード）
- `workspaces:read`（workspace取得）

補足:

- `--scopes` で指定した値は、アプリ側で許可済みでないと `forbidden_scopes` になる
- `--scopes` を省略すると、アプリの既定スコープで認可を試行する
- `user update` / `user update-for-workspace` は OAuth scopes 一覧に `users:write` が存在しないため、アプリ設定（Full permissions など）と実際のAPI応答で要確認

公式ドキュメント:

- OAuth scopes: https://developers.asana.com/docs/oauth-scopes
- Authentication: https://developers.asana.com/docs/authentication

## Output Modes

```bash
asana task list --output json --non-interactive
asana user list --workspace <gid> --domain example.com --output csv --out users.csv
```

## Official Endpoint Groups

- `asana task ...` (27 endpoint mappings)
- `asana project ...` (19 endpoint mappings)
- `asana user ...` (8 endpoint mappings)

Use `--help` under each group for concrete operations.

## Legacy Alias Commands

- `config`
- `workspaces` (`w`)
- `tasks` (`ts`)
- `task <index|gid>`
- `comment` (`cm`)
- `done`
- `due`
- `browse` (`b`)
- `download` (`dl`)

## Profile Files

- `~/.config/asana-cli/config.toml`
- `~/.config/asana-cli/credentials.toml`

Secrets are stored in keyring where available. Fallback environment variables:

- `ASANA_PAT`
- `ASANA_ACCESS_TOKEN`
- `ASANA_REFRESH_TOKEN`
- `ASANA_CLIENT_SECRET`

## Notes

- Local slug memo for API lookup: `.git/info/asana-api-slugs.local.md` (not tracked)
- API scope snapshot docs: `docs/spec/`
