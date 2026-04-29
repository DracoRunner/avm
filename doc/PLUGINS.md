# avm Plugins

`avm` features an extensible, language-agnostic plugin system inspired by `asdf`. Plugins allow `avm` to dynamically discover and expose aliases based on your current project environment.

## 1. Using Plugins

### Installing a Plugin
You can install a plugin from a Git URL or a local path (useful for development):

```bash
avm plugin add https://github.com/DracoRunner/avm-plugin-node
# OR
avm plugin add ./my-local-plugin
```

### Listing Plugins
See all installed plugins and their metadata:

```bash
avm plugin list
```

### Updating Plugins
Keep your plugins up to date with the latest features:

```bash
avm plugin update --all
# OR
avm plugin update node
```

### Removing a Plugin
```bash
avm plugin remove node
```

---

## 2. Official Plugins

| Name | Description |
| :--- | :--- |
| **node** | Automatically exposes `package.json` scripts as aliases using npm, yarn, pnpm, or bun. |

---

## 3. Creating a Plugin (The Contract)

A plugin is simply a directory with a specific structure. You can write them in any language (Node, Bash, Go, Python).

### Directory Structure
```text
my-plugin/
├── plugin.json                # Metadata (Required)
└── bin/
    ├── export-aliases         # Outputs JSON aliases (Required)
    ├── health-check           # Returns exit 0 if plugin applies (Optional)
    └── describe               # Outputs manifest JSON (Optional)
```

### `plugin.json`
```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "api_version": 1,
  "description": "My custom plugin",
  "section_label": "My Custom Aliases"
}
```

### `bin/export-aliases --dir <cwd>`
This hook is executed by `avm` to fetch aliases. It must print a JSON object to `stdout`:

```json
{
  "api_version": 1,
  "aliases": {
    "start": "npm run start",
    "test": {
      "command": "npm test",
      "description": "Run project tests"
    }
  }
}
```

### `bin/health-check --dir <cwd>`
Use this to skip expensive processing. If this exits with a non-zero code, `avm` will not call `export-aliases`.
