# avm — Alias Version Manager

A lightweight local/global command alias manager that works like `asdf` or `nvm`, but for command aliases. It reads from a local `.avm.json` in the current directory, falls back to a global `~/.avm.json`, and if no alias is found it passes the command through to the shell normally.

## How it works

- Local config: `.avm.json` in the current directory (per-project aliases)
- Global config: `~/.avm.json` in your home directory (shared aliases)
- Local overrides global
- If nothing matches, the command runs through as-is

## Installation

### Homebrew

```bash
brew tap DracoRunner/tap
brew install avm
```

Then add to your shell profile (see [Shell Setup](#shell-setup) below).

### npm

```bash
npm install -g @dracorunner/avm
```

### curl (recommended — auto-configures shell)

```bash
curl -fsSL https://raw.githubusercontent.com/DracoRunner/avm/main/install.sh | bash
```

This installer automatically:
- Downloads the right binary for your platform
- Creates `~/.avm.json` if missing
- Adds the shell function to `~/.zshrc` and/or `~/.bashrc`

## Shell Setup

The binary is installed as `avm-bin`. A shell function named `avm` wraps it so that typing `avm <key>` resolves aliases in the current directory context.

Add this one line to your `~/.zshrc` or `~/.bashrc`:

```bash
eval "$(avm-bin shell-init)"
```

Then reload:

```bash
source ~/.zshrc   # or source ~/.bashrc
```

That's it — avm will now resolve aliases based on your current directory every time.

## Commands

### `avm init`
Create a local `.avm.json` in the current directory.

```bash
avm init
```

### `avm add <key> <value>`
Add/update a local alias.

```bash
avm add start "npm run dev"
avm add deploy "sh ./scripts/deploy.sh"
```

### `avm add -g <key> <value>`
Add/update a global alias.

```bash
avm add -g clean "docker system prune -a"
```

### `avm list` (alias: `ls`)
List local + global aliases (local marked with `[override]` when it shadows a global).

```bash
avm list
```

### `avm remove <key>` (alias: `rm`)
Remove a local alias. Use `-g` for global.

```bash
avm remove start
avm remove -g clean
```

### `avm which <key>`
Show what a key resolves to and where it comes from.

```bash
avm which start
```

### `avm <key> [args...]`
Run the alias. Extra args are appended.

```bash
avm start
avm start --port 4000
```

### `avm version`
Print the version.

### `avm shell-init`
Print the shell function (used via `eval`).

## `.avm.json` format

Simple flat JSON:

```json
{
  "start": "npm run dev",
  "db": "docker-compose up postgres",
  "deploy": "sh ./scripts/deploy.sh"
}
```

Keys are the alias names. Values are full shell command strings (args/flags/pipes all OK).

## Local vs Global Resolution

1. **Local first** — checks `.avm.json` in the current directory
2. **Global fallback** — then `~/.avm.json`
3. **Override** — a local alias with the same name shadows the global one (shown as `[override]` in `avm list`)
4. **Passthrough** — if no alias matches, the command runs normally

## Example Workflow

```bash
# In a Node.js project
cd my-app
avm init
avm add start "npm run dev"
avm add test "npm test -- --watch"

# In a Python project
cd ../my-py-app
avm init
avm add start "python manage.py runserver"
avm add test "pytest -v"

# Both projects use 'avm start' — it runs the right thing
```

## Supported Shells

- bash
- zsh

## Contributing

1. Fork the repo
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Commit your changes (`git commit -m 'feat: add my feature'`)
4. Push (`git push origin feat/my-feature`)
5. Open a PR

## License

MIT — see [LICENSE](LICENSE)

## Acknowledgments

Inspired by [asdf](https://asdf-vm.com) and [nvm](https://github.com/nvm-sh/nvm).
