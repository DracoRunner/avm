# avm — Internal Architecture

This document details the internal workings of **avm (Alias Version Manager)**, its components, and how they interact.

---

## 1. High-Level Design
`avm` uses a **hybrid architecture** to achieve directory-aware alias resolution without the performance overhead of a pure shell script or the environment limitations of a pure binary.

### Key Components:
1.  **Shell Wrapper (`avm` function)**:
    *   Defined in `cmd/shell_init.go`.
    *   Intercepts commands to handle `eval` execution within the current shell context.
2.  **Core Binary (`avm-bin`)**:
    *   Written in Go.
    *   Handles all logic: JSON parsing, fuzzy matching, plugin management, and alias resolution.
3.  **Plugin System**:
    *   Directory-based executable plugins in `~/.avm/plugins/`.
    *   Language-agnostic (Bash, Node, Go, etc.).
    *   Plugins provide dynamic aliases based on the current environment.
4.  **Config Store**:
    *   Uses simple JSON maps for speed and portability.

---

## 2. Directory Structure
*   `cmd/`: Entry points for all CLI commands (using Cobra).
    *   `root.go`: Handles the fallback logic and interactive suggestions.
    *   `plugin.go`: Plugin management commands.
*   `internal/plugin/`: Plugin engine.
    *   `manager.go`: Discovery, concurrent execution (worker pool), and caching.
    *   `types.go`: Plugin manifest and API schemas.
*   `internal/config/`: The engine of the application.
    *   `config.go`: File I/O for `.avm.json`.
    *   `resolver.go`: Logic for searching local/global/plugin scopes and fuzzy matching.
    *   `actions.go`: Business logic for CRUD operations on aliases.
*   `internal/tooling/`: Tool runtime resolution and installation.
    *   `tooling.go`: Provider registry and runtime env synthesis.
    *   `node.go`: Node.js provider with install/update path wiring.

---

## 3. Configuration Resolution
The `resolver.go` implements a hierarchical lookup:
1.  **Local Context**: Searches for `.avm.json` in the current working directory.
2.  **Global Context**: Searches for `~/.avm.json`.
3.  **Plugin Context**: Executes all installed plugins to fetch dynamic aliases.
4.  **Tool Context**: Reads local/global `tools` selections from `.avm.json` and resolves active versions.
5.  **Precedence**: Local > Global > Plugins.

---

## 6. Tool Runtime Injection Path
`avm env` now includes resolved runtime environment from selected tools.

`cmd/shell_init.go` loads this using:

1. `command avm-bin env` to obtain resolved exports.
2. `eval` to apply these exports before alias execution.
3. PATH prefixes for installed runtime providers to ensure selected versions take precedence.

---

## 4. Plugin Contract
Plugins follow a strict directory structure and hook system:
*   `plugin.json`: Metadata manifest.
*   `bin/export-aliases --dir <cwd>`: (Required) Returns a JSON map of aliases.
*   `bin/health-check --dir <cwd>`: (Optional) Fast check to skip plugin if it doesn't apply.
*   `bin/describe`: (Optional) Returns metadata dynamically.

---

## 4. Fuzzy Matching Logic
When a command is not found, `avm` employs two strategies in `internal/config/resolver.go`:
*   **Levenshtein Distance**: Calculates the edit distance between the input and known keys (threshold $\le 2$).
*   **Normalization**:
    1.  Replaces separators (`-`, `:`, `_`, `.`) with spaces.
    2.  Splits the string into parts.
    3.  Sorts the parts alphabetically.
    4.  Joins them back with dashes.
    *   *Result*: `ios:run` and `run-ios` both normalize to `ios-run`, allowing matches regardless of naming convention.

---

## 5. Inter-Process Communication
Since the shell function needs to know which suggestion the user picked in the Go binary's interactive menu, they communicate via:
1.  **Environment Variables**: `AVM_RESULT_FILE` is passed to the binary.
2.  **Temporary Files**: The binary writes the selected alias key to a `mktemp` file, which the shell function then reads to execute the command.
3.  **Exit Codes**:
    *   `0`: Passthrough (no alias found, run original command).
    *   `10`: Suggestion selected (re-run logic).
    *   `1`: Error.
