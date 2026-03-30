# Changelog

## Unreleased

- OAuth and PAT environment variables now use the `ASANA_CLI_*` namespace only.
- Added `ASANA_CLI_CLIENT_ID` support for OAuth login.
- Documented the new environment variable names in the README.
- Split Quick Start into separate PAT and OAuth flows.
- Added a post-login hint that points users to `asana config --workspace`.
- `auth login` の既定スコープを `cli-default` preset に変更し、`task-full` は互換 alias として維持。
