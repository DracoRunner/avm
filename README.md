# avm — Alias Version Manager

A lightweight local/global command alias manager that works like `asdf` or `nvm`, but for command aliases. It reads from a local `.avm.json` in the current directory, falls back to a global `~/.avm.json`, and if no alias is found, it offers interactive suggestions or passes the command through to the shell normally.

## Why avm?

Development setups often start clean: a few git aliases, a tidy `.zshrc`, and some short commands. But as soon as you dive into complex ecosystems like **React Native**, everything changes. Suddenly, you're juggling:

- Complex ADB commands
- Cryptic Xcode flags
- Long Emulator IDs
- App IDs and deployment scripts

The problem isn't the platform—it's **context switching**. Every project and every platform has different commands and syntax. You waste time googling the same commands or digging through Slack history.

**avm was built to fix that.**

Instead of memorizing commands, you define them once per project in a `.avm.json`. avm keeps your shorthand scoped to the project and directory.

- **One command runner** for everything.
- **Project-scoped** aliases that don't pollute your global shell config.
- **Global fallback** for your most common tools.
- **Clean environment**: No more `.zshrc` bloat or git alias mess.

## Features

- **Directory-Aware**: Aliases change automatically based on your current folder.
- **Global & Local**: Use global aliases for general tools and local ones for project-specific tasks.
- **Interactive Suggestions**: (New!) If you mistype a command, avm suggests the closest match and lets you run it immediately.
- **Placeholder Support**: Pass arguments into your aliases using `$1`, `$2`, etc.
- **Passthrough**: If no alias is found, it runs the command through your shell as-is.
- **Extensible Plugins**: (New!) Dynamically extend avm with community or custom plugins.

## Plugins

`avm` supports a language-agnostic plugin system to dynamically discover project-specific aliases.

### Official Plugins
| Plugin | Description |
| :--- | :--- |
| [node](https://github.com/DracoRunner/avm-plugin-node) | Automatically exposes `package.json` scripts as aliases (npm, yarn, pnpm, bun). |

See the [Plugins Documentation](doc/PLUGINS.md) for installation and creation guides.

For technical details on the architecture, see [ARCHITECTURE.md](doc/ARCHITECTURE.md). For contribution and release instructions, see the [Development Guide](doc/DEVELOPMENT.md).

## Installation

### Homebrew

```bash
brew tap DracoRunner/tap
brew install avm
```

### npm

```bash
npm install -g @dracorunner/avm
```

### curl (recommended — auto-configures shell)

```bash
curl -fsSL https://raw.githubusercontent.com/DracoRunner/avm/main/install.sh | bash
```

## Shell Setup

Add this line to your `~/.zshrc` or `~/.bashrc`:

```bash
eval "$(avm-bin shell-init)"
```

## Usage Examples

Define your aliases in `.avm.json`:

```json
{
  "aliases": {
    "android:start": "npx react-native run-android",
    "ios:build": "npx react-native run-ios --configuration Release",
    "git:feature": "git checkout -b feature/$1",
    "docker:up": "docker-compose up -d"
  },
  "env": {
    "NODE_ENV": "development",
    "API_URL": "https://api.local"
  }
}
```

`avm` also still supports legacy `.avm.json` files that contain only aliases as a flat map.
Existing legacy `.avm.json` files are migrated to the structured form automatically on read, so users upgrading to this version do not need to manually edit existing files.

Now you just type:
- `avm android:start`
- `avm ios:build`
- `avm git:feature my-new-feature`
- `avm docker:up`

Run `avm list` to inspect active local/global aliases and `env` values side by side.

### Interactive Suggestions

Mistyped a command? `avm` has your back:

```bash
$ avm tv-run
avm: unknown command or alias "tv-run"
? Did you mean one of these?:
  ▸ tv:run
    None (run as-is)
```

## Commands

- `avm init`: Initialize local config.
- `avm add <key> <value>`: Add local alias.
- `avm add -g <key> <value>`: Add global alias.
- `avm list` (or `ls`): List all aliases.
- `avm remove <key>` (or `rm`): Remove an alias.
- `avm which <key>`: See where an alias points.

## Read More
Check out the story behind avm on [LinkedIn](https://lnkd.in/gEMzdm8P).

## License
MIT
