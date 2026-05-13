# avm — Features

**avm** is designed to simplify complex development environments by providing scoped, context-aware command aliases.

---

## 1. Directory-Aware Scoping
The core feature of `avm`. Aliases are loaded from the `.avm.json` file in your current directory.
When needed, command environment can also be declared in the same file under `env` and is applied automatically for commands run through `avm`.
*   **Use Case**: In a React Native project, `avm start` runs the bundler. In a Go project, `avm start` might run `go run main.go`.
*   **Benefit**: You never have to remember which project uses which start command.

## 2. Global & Local Hierarchy
`avm` maintains three layers of configuration:
*   **Local (`./.avm.json`)**: Specific to the project folder.
*   **Global (`~/.avm.json`)**: Available everywhere.
*   **Plugins**: Dynamically generated based on the environment.
*   **Shadowing**: **Local > Global > Plugins**. If multiple sources define the same key, the higher priority source wins. `avm list` will explicitly show overrides.

## 3. Extensible Plugin System
`avm` features a language-agnostic plugin system inspired by `asdf`.
*   **Dynamic Discovery**: Plugins can provide aliases based on the current directory (e.g., exposing `package.json` scripts).
*   **High Performance**: Plugins run in parallel with a bounded worker pool and are cached per-invocation to ensure zero perceptible lag.
*   **Management CLI**: Easily add, remove, and update plugins using `avm plugin add/remove/list/update`.

See [PLUGINS.md](PLUGINS.md) for more details.

## 4. Official Plugins

### `avm-plugin-node`
The first official plugin that automatically detects Node.js projects and exposes all `package.json` scripts as `avm` aliases.
*   **Smart Detection**: Detects your package manager automatically (npm, yarn, pnpm, bun) based on lockfiles.
*   **Zero Config**: Simply install the plugin and your Node scripts are available immediately.

## 5. Dynamic Placeholder Support
Aliases are not static strings; they can accept arguments.
*   **Placeholders**: Use `$1`, `$2`, etc., in your alias definition.
*   **Substitution**: `avm` replaces these markers with positional arguments passed at runtime.
*   **Example**:
    *   Alias: `"deploy": "ssh $1 'docker restart $2'"`
    *   Execution: `avm deploy prod-server web-app`
    *   Result: Executes `ssh prod-server 'docker restart web-app'`

## 6. Interactive Suggestions
Mistyped commands don't just fail. `avm` proactively helps you find what you meant.
*   **Fuzzy Search**: Uses edit-distance and phonetic-like normalization.
*   **UI**: Provides a clean, arrow-key navigable menu in your terminal.
*   **Instant Execution**: Selecting a suggestion runs it immediately, saving you from re-typing.

## 5. Shell Passthrough
`avm` is a non-destructive wrapper.
*   If a command matches an alias, it runs the alias.
*   If it doesn't match an alias but is a valid system command (e.g., `ls`, `git`), it passes it through to the shell.
*   **Benefit**: You can prefix everything with `avm` or use it as a replacement for your standard shell runner.

## 6. Zero-Config Initialization
Getting started is a single command:
*   `avm init`: Creates a clean `.avm.json` file.
*   `avm shell-init`: Provides the snippet to hook into your shell, with an `install.sh` script that automates this for most users.
