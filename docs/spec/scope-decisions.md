# Scope Decisions

## Locked Decisions

1. Authentication methods: OAuth + PAT
2. Profile model: `~/.config/asana-cli/config.toml` + `~/.config/asana-cli/credentials.toml`
3. Secret storage: OS keyring first; environment variables as fallback
4. Domain filtering: exact match only, multi-domain allowed
5. User inclusion defaults: active members only
6. Output defaults: table for human, json for automation (`--output json --non-interactive`)
7. MCP: explicitly out of scope
