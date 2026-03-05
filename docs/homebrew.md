# Homebrew Distribution

This repository ships a Homebrew formula at `Formula/asana.rb`.

## Install from this repository tap

```bash
brew tap cloudnative-co/asana-cli https://github.com/cloudnative-co/asana-cli
brew install --HEAD cloudnative-co/asana-cli/asana
```

Notes:

- `brew tap cloudnative-co/asana-cli` without URL looks for `cloudnative-co/homebrew-asana-cli`.
- URL form pins the tap to this repository.

## Stable release update flow

1. Create and push a tag (example: `v0.1.0`).
2. Run:

```bash
./scripts/update-homebrew-formula.sh v0.1.0
```

3. Review `Formula/asana.rb`.
4. Validate formula:

```bash
brew install --HEAD cloudnative-co/asana-cli/asana --dry-run
```

5. Commit and push formula updates.

## Formula maintenance policy

- `head` always points to `main`.
- `url` + `sha256` are added/updated only for tagged releases.
- Keep `test do` lightweight (`asana --help`) to avoid external Asana dependencies.
