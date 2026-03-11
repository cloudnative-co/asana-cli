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

### OAuth Permission Scopes の設定パターン

Asana Developer Console の **OAuth > Permission scopes** は、次のどちらかで運用する。

1. Full permissions を使う（簡単だが権限が広い）
2. Specific scopes を使う（推奨: 最小権限）

#### 1) Full permissions 方式

- アプリ設定で `Full permissions` を有効化
- `asana auth login` は `--scopes` を指定しない
- もし `forbidden_scopes: ... default identity ...` が出る場合は、`Full permissions` が無効か配布設定不足

```bash
asana auth login \
  --profile default \
  --client-id "$ASANA_MCP_CLIENT_ID" \
  --client-secret "$ASANA_MCP_CLIENT_SECRET" \
  --redirect-uri "urn:ietf:wg:oauth:2.0:oob"
```

#### 2) Specific scopes 方式（推奨）

- `Full permissions` を無効化
- アプリ設定で必要scopeのみ有効化
- CLIでも同じscopeを `--scopes` で明示

```bash
asana auth login \
  --profile default \
  --client-id "$ASANA_MCP_CLIENT_ID" \
  --client-secret "$ASANA_MCP_CLIENT_SECRET" \
  --redirect-uri "urn:ietf:wg:oauth:2.0:oob" \
  --scope-preset task-full
```

`--scope-preset task-full` は task周辺API向けの推奨scopeセットを展開する:

- `tasks:read`, `tasks:write`, `tasks:delete`
- `projects:read`, `projects:write`, `projects:delete`
- `users:read`
- `stories:read`, `stories:write`
- `attachments:read`, `attachments:write`, `attachments:delete`
- `custom_fields:read`, `custom_fields:write`
- `tags:read`, `tags:write`
- `time_tracking_entries:read`
- `task_custom_types:read`, `task_templates:read`
- `teams:read`
- `workspaces:read`

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
- `attachments:write`（添付アップロード）
- `attachments:delete`（添付削除）
- `custom_fields:read`（タスクの custom field 読み取り）
- `custom_fields:write`（custom field 作成・更新）
- `tags:read`
- `tags:write`
- `time_tracking_entries:read`
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

### Frequent filtering patterns

Use local name filters on list responses:

```bash
asana task list-project \
  --project-gid 1199687679891327 \
  --query opt_fields=gid,name,completed,due_on,assignee.name,permalink_url \
  --name-contains pocketalk \
  --output json \
  --non-interactive
```

Regex filter is also supported:

```bash
asana task list-project --project-gid <gid> --name-regex 'pocketalk-[0-9]+' --output json
```

Notes:

- `--name-contains` is case-insensitive.
- `--name-regex` uses Go regular expressions.
- For list/search endpoints with `--all`, CLI now sets `limit=100` automatically if not specified to avoid large-result errors.

### Assigned tasks and resolved projects

List tasks for a specific assignee:

```bash
asana task list \
  --assignee me \
  --workspace <workspace_gid> \
  --output json \
  --non-interactive
```

If the active profile already has a workspace set, `--workspace` can be omitted.

Resolve projects from ancestor tasks without overwriting the task's direct `projects` field:

```bash
asana task get \
  --task-gid <task_gid> \
  --resolve-projects ancestors \
  --output json \
  --non-interactive
```

This adds:

- `resolved_projects`
- `resolved_from_task_gid`
- `resolved_from_task_name`
- `resolved_from_depth`
- `resolved_projects_status`

Expand descendant subtasks from any task-shaped response:

```bash
asana task list-project \
  --project-gid <project_gid> \
  --query completed_since=now \
  --include-subtasks descendants \
  --output json \
  --non-interactive
```

For list responses, descendant subtasks are flattened into `data` and annotated with:

- `expanded_from_task_gid`
- `expanded_from_task_name`
- `subtask_depth`

For single-task responses, the root task keeps its normal shape and adds:

- `descendant_subtasks`
- `descendant_subtasks_count`

`completed_since` is propagated to descendant subtask fetches, so `completed_since=now` keeps the expansion open-task only.

## Official Endpoint Groups

- `asana task ...` (29 endpoint mappings)
- `asana project ...` (19 endpoint mappings)
- `asana user ...` (8 endpoint mappings)
- `asana custom-field ...` (8 endpoint mappings)
- `asana tag ...` (9 endpoint mappings)
- `asana section ...` (8 endpoint mappings)
- `asana story ...` (5 endpoint mappings)
- `asana attachment ...` (4 endpoint mappings)
- `asana user-task-list ...` (2 endpoint mappings)
- `asana time-entry ...` (6 endpoint mappings)

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
