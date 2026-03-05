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
  --client-secret <client_secret>
```

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
