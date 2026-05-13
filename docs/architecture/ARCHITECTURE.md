# avm Architecture

`avm` is a Rust-native CLI for project-local aliases, runtime selection, shims, and provider plugins.

## Workspace crates

| Crate | Responsibility |
| --- | --- |
| `crates/avm-cli` | Binary entrypoint, command routing, shell protocol, and user-facing behavior. |
| `crates/avm-core` | `.avm.json` parsing, config migration, local/global merge rules, alias/env/tool resolution. |
| `crates/avm-shims` | Shim directory management and executable shim generation. |
| `crates/avm-plugin-api` | Shared plugin manifest, alias response, and tool provider contracts. |
| `crates/avm-runtime` | Plugin discovery, manifest validation, timeout handling, and legacy executable adapter. |
| `crates/avm-plugin-node` | Built-in Node provider for package scripts and Node version lookup. |

## Config model

Primary config file:

```json
{
  "aliases": {
    "dev": "pnpm run dev"
  },
  "env": {
    "NODE_ENV": "development"
  },
  "tools": {
    "node": "20.11.1"
  }
}
```

Legacy flat-map files are still accepted:

```json
{
  "dev": "pnpm run dev"
}
```

Resolution order:

1. Local `.avm.json`
2. Global `~/.avm.json`
3. Plugin/provider aliases
4. System command fallback

## Runtime boundaries

- CLI commands stay in `avm-cli`.
- Config and resolver logic stays in `avm-core`.
- Provider contracts stay in `avm-plugin-api`.
- Built-in Node behavior stays in `avm-plugin-node`.
- External plugin execution stays in `avm-runtime`.
- Shim creation and path handling stays in `avm-shims`.

## Tool behavior

`avm` does not auto-install missing tools during execution. If a configured Node version is missing, shim execution warns and falls back to the next matching system binary outside the avm shim directory.

The `tool install` command is present as the provider installer surface, but the current Node baseline reports installation as unsupported until the Node downloader/verifier is implemented.

## Plugin behavior

The runtime supports executable-style plugins through the compatibility adapter. New provider work should use host-owned contracts first and keep plugin failures isolated from the main CLI command.
