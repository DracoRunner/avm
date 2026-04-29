I read through the whole codebase. Good news: you took most of my earlier suggestions and the structure is solid. But there are real bugs, races, and rough edges that will bite you. Below is an audit-only review — exact files and line numbers, ranked by severity. **No fixes applied.**

---

## High-Level Verdict

**What works well**
- Manifest-first plugin discovery (`plugin.json` + `bin/export-aliases|describe|health-check`) ✅
- Worker-pool concurrency with per-plugin + global timeouts ✅
- Local-path install via symlink ✅
- Hierarchy correctly implemented as Local > Global > Plugin ✅
- Output schema supports both shorthand string and `{command, description, source}` object ✅

**What's missing or broken**
- Real concurrency bugs in the cache and the `loaded` flag.
- Plugin stderr is silently swallowed → debugging is impossible.
- Non-deterministic plugin precedence and unsorted list output.
- Shell wrapper has `eval` quoting issues that will break aliases with arguments containing spaces or special chars.
- Zero plugin-side tests, despite the spec calling for them.
- No disk cache, so plugins re-spawn on every keystroke (kills tab-completion later).

---

## 🔴 Critical (fix before next release)

| # | File:Line | Issue |
|---|---|---|
| C1 | `internal/plugin/manager.go:32-37` | **Race window in cache.** Pattern reads `cache` under lock, releases, then proceeds. Two callers can both pass the nil check and both run all plugins. The declared `cacheOnce sync.Once` (line 16) is never used — replace the manual locking with `cacheOnce.Do(...)`. |
| C2 | `internal/plugin/manager.go:130-134` | **`cmd.Output()` discards stderr.** Plugin errors written to stderr (`export-aliases` already does this on line 51 of the node plugin) vanish. Capture both with `cmd.CombinedOutput()` or set `cmd.Stderr` to a buffer; print to user only when `AVM_DEBUG=1`. Without this, broken plugins are invisible. |
| C3 | `internal/plugin/manager.go:82-90` | **Non-deterministic plugin precedence.** "First wins" depends on goroutine scheduling. Two runs with the same setup can return different aliases. Either sort plugin names alphabetically before merging, or add a `priority` field in `plugin.json`. |
| C4 | `cmd/shell_init.go:49,80` | **Unsafe `eval`.** `eval "$resolved $@"` and `eval "$@"` re-tokenize and re-interpret quoting. Try `avm git:feature "my branch with spaces"` and watch it explode. Use `eval "$resolved \"\$@\""` or, better, drop `eval` and use `bash -c` with `--` and positional args, or move execution into Go (`exec.Command("sh", "-c", resolved, "--", args...)`). |
| C5 | `cmd/shell_init.go:30` | **Fragile parse of `which` output.** `grep "^Command:" \| sed 's/^Command: //'` breaks if the alias value contains `\n` or starts with whitespace. Add a dedicated `avm-bin resolve <key>` (or `avm-bin which --raw <key>`) that prints **only** the command on stdout, or empty + nonzero exit if not found. |
| C6 | `cmd/shell_init.go:38-46` | **Placeholder substitution in `sed` is unsafe.** If an arg contains `&`, `\1`, `/`, `|`, or newlines, the substitution corrupts. Move placeholder expansion into Go — `which` already knows the template, just have it return the rendered command. |

---

## 🟡 Important

| # | File:Line | Issue |
|---|---|---|
| I1 | `internal/config/actions.go:90,104,118-131` | **Map iteration is randomized in Go.** `avm list` shows aliases in different order every run. Sort keys before printing — for local, global, sections map, and aliases inside each section. |
| I2 | `internal/config/resolver.go:41` | **`loaded = true` is a data race.** Read on line 16, write on line 41, no lock. Wrap state init in a `sync.Once` or guard with mutex. |
| I3 | `internal/plugin/manager.go:206-232` | **`InstallPlugin` doesn't validate.** After clone/symlink there is no check that `plugin.json` parses, `api_version` is supported, or `bin/export-aliases` exists and is executable. Validate immediately and rollback (delete clone or unlink symlink) on failure. |
| I4 | `internal/plugin/manager.go:217-220` | **`git clone` is silent and can clash.** No stdout/stderr piped → no progress shown. If target dir already exists you get a cryptic git error. Pipe `cmd.Stdout/Stderr = os.Stdout/Stderr`, and check `os.Stat(target)` first with a clear message ("plugin already installed; use `avm plugin update`"). |
| I5 | `internal/plugin/manager.go:234-236` | **`isGitURL` is loose.** `strings.HasPrefix(s, "http")` matches `httptest`. Use proper prefixes: `https://`, `http://`, `git@`, `git://`, `ssh://`. Also consider `user/repo` shorthand → expand to `https://github.com/user/repo`. |
| I6 | `internal/plugin/manager.go:127-134` | **No distinction between timeout, exec error, and bad JSON.** All three return `nil` silently. At minimum, check `errors.Is(ctx.Err(), context.DeadlineExceeded)` and surface a `[avm] plugin "<name>" timed out` warning under `AVM_DEBUG`. |
| I7 | `cmd/plugin.go:36-54` | **`os.ReadDir` error treated as "no plugins"** — masks permission errors. Distinguish `os.IsNotExist(err)` from real failures. Also: when dir exists but empty, the header `Installed Plugins:` prints with nothing below. |
| I8 | `internal/plugin/types.go:18-21` | **`map[string]interface{}` for aliases.** Works but loses type safety and forces the type-switch on `manager.go:153-163`. Consider a `json.RawMessage` + custom `UnmarshalJSON` on a typed struct that handles both `string` and full object form. Cleaner and unit-testable. |
| I9 | `internal/plugin/manager.go` (no file) | **No plugin-side tests.** The original spec said "install a mock plugin that returns static JSON … verify `avm list` and `avm <plugin-alias>` work." Add `internal/plugin/manager_test.go` with a `testdata/` dir containing fake plugins (one healthy, one timing out via `sleep 5`, one returning malformed JSON, one with no `package.json` to test the no-op path). |
| I10 | `cmd/plugin.go:21-30` | **No `--ref` flag for `plugin add`** — can't pin to a tag/branch. Add `--ref <tag-or-branch>` and pass `-b <ref>` to `git clone`. |
| I11 | `cmd/root.go:60,78,82` | **`os.Exit` inside Cobra `Run`.** Skips deferred cleanup (currently none, but brittle). Prefer returning from `RunE` with sentinel errors and a wrapper in `Execute` that maps to exit codes. |
| I12 | `cmd/root.go:37` | **No source label on suggestions.** `aliasSuggestions` and `subCommandSuggestions` are merged flat — user can't tell if "build" is a Cobra subcommand or an alias. Prefix with `[alias]` / `[cmd]` in the menu. |

---

## 🟢 Nice to have

| # | File:Line | Issue |
|---|---|---|
| N1 | `internal/plugin/manager.go` | **No disk cache.** Spec mentioned this for v1.1. Key by `(plugin_name, cwd, mtime of trigger files)` and store in `~/.avm/cache/plugins/...`. Massive win for shell completion (which you'll want eventually). |
| N2 | `plugins/node/bin/export-aliases:34` | **`yarn` without `run` is risky.** If a `package.json` script is named `add`, `install`, `remove`, etc., `yarn add` does the wrong thing. Always emit `yarn run <name>` for safety. |
| N3 | `plugins/node/bin/export-aliases` | **Detection priority** — bun before pnpm before yarn before npm is good, but consider also reading `packageManager` field in `package.json` (Corepack) which is more authoritative than lockfiles. |
| N4 | `internal/config/actions.go:134` | **`Plugin: %s:`** has odd colon styling vs `Local aliases (.avm.json):`. Pick one format. |
| N5 | `cmd/plugin.go:57-69` | **`plugin remove` doesn't confirm or check existence.** Silent no-op if the plugin doesn't exist (`os.RemoveAll` returns nil for missing path). Print "plugin <name> not found" if dir didn't exist. |
| N6 | `internal/plugin/manager.go:238-245` | **`UpdatePlugin` is silent for non-git plugins** (e.g., symlinked local). Print "skipping <name>: not a git plugin (symlink)". |
| N7 | `cmd/version.go:9` | `version` is fine for ldflags injection (`-X avm/cmd.version=…`), but I don't see ldflags in `.goreleaser.yml`. Verify `goreleaser` actually wires the version through. |
| N8 | `cmd/root.go` | **No tab-completion for alias keys.** Cobra supports custom completion functions; add one that calls `config.GetAliases()` and emits keys. With disk cache (N1), this becomes fast. |
| N9 | `internal/config/resolver.go:50-66` | When the same key exists in multiple scopes, `Which` only reports the winning one. Consider `avm which --all <key>` that lists every source so users can debug shadowing. |
| N10 | `doc/ARCHITECTURE.md:58` | Two sections numbered "4." (Plugin Contract and Fuzzy Matching). Renumber. |

---

## Concrete next-step fixes (in priority order)

If you only have a couple hours, do these five — they fix the worst of the issues with minimal surface area:

```text
1. internal/plugin/manager.go  → use cacheOnce.Do, capture stderr to buffer,
                                  sort plugin names before merging.
2. internal/config/actions.go  → sort.Strings on map keys before all prints.
3. cmd/shell_init.go           → add `avm-bin resolve <key>` that prints raw
                                  command; replace grep|sed parse; quote $@
                                  in eval (or move execution into Go).
4. internal/plugin/manager.go  → InstallPlugin: validate manifest + bin
                                  exists; rollback on failure; pipe git output.
5. internal/plugin/manager_test.go (new)
                               → testdata/ with mock plugins (healthy,
                                  slow, malformed JSON, no-trigger-file).
```

---

## Bigger architectural call-out

Right now the shell wrapper calls `avm-bin which` then potentially `avm-bin <args>` — **two separate processes**, each independently spawning all plugins (because the Go in-process cache doesn't persist between invocations). That's the real reason a disk cache (N1) matters: without it, every `avm <something>` is paying the plugin-spawn cost twice.

A cleaner model:

```
shell function avm() {
    avm-bin run -- "$@"     # one process; resolves, expands placeholders,
                            # and prints final command on stdout (or signals
                            # interactive flow via exit code)
    eval "$(...captured output...)"
}
```

That collapses the two-call dance, lets all caching stay in-process, and removes the `grep|sed` parsing entirely. Worth considering before you ship the plugin system widely.

---

Want me to:
- **A)** turn the top 5 fixes into a structured agent prompt you can paste into Claude Code, or
- **B)** write the missing `manager_test.go` test plan with the testdata layout, or
- **C)** sketch the unified `avm-bin run` redesign in more detail?



# Unified `avm-bin run` Redesign

The goal: collapse the current two-process dance (`which` → `eval` or fallback to suggestion mode) into **one process per user command**. This keeps all plugin caching in-process, eliminates the fragile `grep|sed` parse, and moves placeholder expansion out of bash sed-hell.

---

## Current flow (problem statement)

```
┌─────────────────────────────────────────────────────────────────┐
│  user types: avm android:start emulator-5554                     │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
                  shell function avm()
                            │
            ┌───────────────┴────────────────┐
            ▼                                ▼
   PROCESS 1: avm-bin which        (alias resolved)
   - load .avm.json                          │
   - load ~/.avm.json                        ▼
   - spawn ALL plugins                bash sed for $1,$2…
   - print "Command: …"                      │
   - exit                                    ▼
            │                          eval "$cmd $@"
            │ stdout parsed via grep|sed
            ▼
   alias not found?
            │
            ▼
   PROCESS 2: avm-bin <key>
   - load .avm.json AGAIN
   - load ~/.avm.json AGAIN
   - spawn ALL plugins AGAIN  ← duplicate work
   - run promptui suggestion menu
   - write pick to AVM_RESULT_FILE, exit 10
            │
            ▼
   shell function recurses: avm <picked-key>
   - … and we go through the whole thing once more
```

**Real costs**:
- Plugins spawn 2× per command (in-process cache only helps within one process).
- `grep|sed` parse breaks on multiline alias values, embedded `Command:` strings, leading whitespace.
- Bash `sed` placeholder substitution corrupts on `&`, `|`, `/`, newlines in args.
- `eval "$resolved $@"` re-tokenizes user args — quoted args with spaces break.

---

## Proposed flow

```
┌─────────────────────────────────────────────────────────────────┐
│  user types: avm android:start emulator-5554                     │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
              shell function avm() (~10 lines)
                            │
                            ▼
   ┌────────────────────────────────────────────────┐
   │ PROCESS: avm-bin run -- android:start ...      │
   │                                                │
   │  load configs + plugins ONCE                   │
   │                                                │
   │  ┌─ resolved? ──► render placeholders in Go    │
   │  │                emit RUN protocol on stdout  │
   │  │                                             │
   │  ├─ subcommand? ► run normally, emit DONE      │
   │  │                                             │
   │  ├─ suggestion? ► promptui on /dev/tty         │
   │  │                user picks → re-resolve      │
   │  │                in-process (cache hot!)      │
   │  │                emit RUN protocol            │
   │  │                                             │
   │  └─ no match?  ──► emit PASSTHROUGH protocol   │
   └────────────────────────────────────────────────┘
                            │
                            ▼
              shell parses ONE protocol line
              and exec's the command
```

One process. One config load. One plugin spawn round. All resolution + suggestion logic stays in Go.

---

## The `run` protocol

`avm-bin run` writes a single structured line to **stdout** as its last action. Everything else (logs, prompts, errors) goes to stderr or `/dev/tty`. The shell wrapper reads that one line and acts on it.

I'd use a tiny line-prefixed protocol — easier than JSON to parse from bash and impossible to confuse with command output.

```
AVM_EXEC <argv-as-NUL-separated>      # run this exact argv vector
AVM_PASSTHROUGH                        # original args were not aliases — let shell run them
AVM_DONE <exit_code>                   # avm-bin already did the work (e.g. `avm list`); just exit
AVM_ABORT                              # user cancelled (ESC); exit 0, do nothing
AVM_ERROR <message>                    # something failed; exit nonzero
```

NUL separation is key — it survives any character in the argv (spaces, quotes, newlines, `&`, `|`). Bash reads it with `IFS=$'\0' read -rd ''`.

---

## Go side: `cmd/run.go` (sketch)

```go
package cmd

import (
    "avm/internal/config"
    "fmt"
    "os"
    "strings"

    "github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
    Use:                "run [-- args...]",
    Short:              "Resolve and emit execution plan for the shell wrapper",
    Hidden:             true, // not user-facing; the shell function calls it
    DisableFlagParsing: true, // we want raw argv after `--`
    Args:               cobra.ArbitraryArgs,
    RunE: func(_ *cobra.Command, args []string) error {
        // Strip leading "--" if present (cobra usually consumes it, belt+braces)
        if len(args) > 0 && args[0] == "--" {
            args = args[1:]
        }
        if len(args) == 0 {
            emit("AVM_DONE", "0")
            return nil
        }

        key, rest := args[0], args[1:]

        // 1. Is it an avm subcommand? (init, add, list, plugin, ...)
        //    These are handled by re-dispatching through Cobra normally.
        //    The shell wrapper already filters most of these out before
        //    calling `run`, but we double-check here.
        if isReservedSubcommand(key) {
            // Tell the shell to invoke us again without `run`.
            // This is rare — only if user bypassed shell wrapper.
            emitExec(append([]string{"avm-bin", key}, rest...))
            return nil
        }

        // 2. Try to resolve as alias (single load — cache hot for whole process)
        resolved, found, _, err := config.ResolveWithSource(key)
        if err != nil {
            emit("AVM_ERROR", err.Error())
            return nil
        }

        if found {
            // Render $1, $2, ... in Go — no sed, no quoting headaches
            cmdline := expandPlaceholders(resolved, rest)
            emitExec([]string{"sh", "-c", cmdline, "--"})
            return nil
        }

        // 3. Fuzzy suggestions — promptui runs on /dev/tty, not stdout
        suggestions := config.SuggestAliases(key)
        if len(suggestions) > 0 {
            picked, ok := runSuggestionUI(key, suggestions) // writes prompt to /dev/tty
            if !ok {
                // user hit ESC or chose "None"
                emit("AVM_PASSTHROUGH")
                return nil
            }
            // Re-resolve the picked key — cache is hot, ~zero cost
            resolved, found, _, _ := config.ResolveWithSource(picked)
            if !found {
                emit("AVM_PASSTHROUGH")
                return nil
            }
            cmdline := expandPlaceholders(resolved, rest)
            emitExec([]string{"sh", "-c", cmdline, "--"})
            return nil
        }

        // 4. No alias, no suggestion → tell shell to run user's args as-is
        emit("AVM_PASSTHROUGH")
        return nil
    },
}

// emit writes a single protocol line to stdout.
func emit(parts ...string) {
    fmt.Fprintln(os.Stdout, strings.Join(parts, " "))
}

// emitExec writes AVM_EXEC followed by NUL-separated argv on stdout.
func emitExec(argv []string) {
    fmt.Fprint(os.Stdout, "AVM_EXEC ")
    fmt.Fprint(os.Stdout, strings.Join(argv, "\x00"))
    fmt.Fprintln(os.Stdout)
}

// expandPlaceholders replaces $1..$N with args, leaves others empty.
// Uses string ops only — no exec, no shell.
func expandPlaceholders(template string, args []string) string {
    if !strings.ContainsRune(template, '$') {
        // No placeholders — append args at the end (current behavior)
        if len(args) == 0 {
            return template
        }
        return template + " " + shellQuoteAll(args)
    }
    // Replace $1..$N
    out := template
    for i, a := range args {
        out = strings.ReplaceAll(out, fmt.Sprintf("$%d", i+1), shellQuote(a))
    }
    // Strip leftover unsubstituted $N
    return placeholderRegex.ReplaceAllString(out, "")
}
```

Two important details:

1. **Placeholder expansion happens in Go** with proper shell-quoting (`shellQuote` wraps each arg in `'...'` with `'\''` escaping). This kills the entire class of sed-injection bugs from the current bash impl.

2. **Suggestion UI uses `/dev/tty` directly**, not stdin/stderr. That way it can stay interactive while stdout is being captured by the shell wrapper.

```go
func runSuggestionUI(query string, suggestions []string) (string, bool) {
    tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
    if err != nil {
        return "", false
    }
    defer tty.Close()

    options := append(suggestions, "None (run as-is)")
    prompt := promptui.Select{
        Label:        fmt.Sprintf("avm: unknown \"%s\" — did you mean?", query),
        Items:        options,
        Stdin:        tty,
        Stdout:       tty,
        HideSelected: true,
    }
    _, result, err := prompt.Run()
    if err != nil || result == "None (run as-is)" {
        return "", false
    }
    return result, true
}
```

---

## Shell side: `cmd/shell_init.go` (new wrapper)

The wrapper shrinks dramatically. No `grep|sed`, no `mktemp`, no `eval`:

```bash
avm() {
  # Fast path: pure avm subcommands bypass `run` entirely
  case "$1" in
    init|add|list|ls|remove|rm|which|version|help|shell-init|completion|plugin|--help|-h|--version|-v|"")
      command avm-bin "$@"
      return $?
      ;;
  esac

  # Capture only stdout; stderr and /dev/tty pass through for prompts/logs
  local _avm_out
  _avm_out="$(command avm-bin run -- "$@")"
  local _avm_rc=$?

  if [ $_avm_rc -ne 0 ] && [ -z "$_avm_out" ]; then
    return $_avm_rc
  fi

  # Read exactly the first line as the verb
  local _verb="${_avm_out%% *}"
  local _payload="${_avm_out#"$_verb"}"
  _payload="${_payload# }"

  case "$_verb" in
    AVM_EXEC)
      # _payload is argv joined by NUL — split safely
      local -a _argv=()
      while IFS= read -r -d '' _part; do
        _argv+=("$_part")
      done < <(printf '%s\0' "$_payload")
      "${_argv[@]}"
      return $?
      ;;
    AVM_PASSTHROUGH)
      "$@"
      return $?
      ;;
    AVM_DONE)
      return "${_payload:-0}"
      ;;
    AVM_ABORT)
      return 0
      ;;
    AVM_ERROR)
      printf 'avm: %s\n' "$_payload" >&2
      return 1
      ;;
    *)
      # Unknown protocol — print stdout verbatim and bail safely
      printf '%s\n' "$_avm_out"
      return 0
      ;;
  esac
}
```

Roughly **35 lines vs the current ~80**, and no `eval` anywhere.

---

## What this fixes (mapping back to the audit)

| Audit issue | How `run` fixes it |
|---|---|
| **C2** stderr swallowed | One process — plugin stderr can be surfaced under `AVM_DEBUG` without timing/race issues. |
| **C4** `eval` quoting | Removed entirely. Final exec is `"${_argv[@]}"` with proper array expansion. |
| **C5** fragile `grep\|sed` | Replaced with explicit protocol; first whitespace-separated token is the verb. |
| **C6** sed-based placeholders | Moved to Go with `shellQuote`; no shell metacharacter issues. |
| Plugins spawn 2× | One process for the whole flow — config + plugins load once, suggestion path reuses cache. |
| `os.Exit(10)` IPC dance | Gone. No `AVM_RESULT_FILE`, no temp files, no exit-code semaphores. |

---

## Edge cases worth thinking through before you build it

1. **Subshell prompts.** `_avm_out="$(...)"` runs `avm-bin` in a subshell, but the suggestion UI talks to `/dev/tty` directly — that still works because `/dev/tty` is the controlling terminal, not stdin. Verify on macOS Terminal, iTerm2, and a basic Linux tty (and inside `tmux`, which can be weird).

2. **Empty args / `avm` alone.** Bash case statement short-circuits to `avm-bin` directly when `$1` is empty — that calls root cmd → help. Same as today.

3. **`avm <existing-cobra-subcommand>`.** Caught by the `case` filter before `run`. If a user accidentally aliases a key like `list`, the alias gets shadowed by the subcommand — which is current behavior, just be aware.

4. **Exit code propagation.** `AVM_EXEC` runs the resolved command; its exit code becomes the function's exit code. `AVM_DONE` carries an explicit code. Make sure `avm-bin run` itself always exits 0 unless something is genuinely broken — otherwise the bash wrapper's `[ $_avm_rc -ne 0 ]` check kicks in too eagerly.

5. **Output buffering.** `fmt.Fprintln` is line-buffered to stdout, but if any code earlier in the process writes to stdout without a newline, the shell parse breaks. Either redirect all noise to stderr, or wrap `os.Stdout` so only `emit()` can write to it (a `protocolWriter` that panics on misuse is a nice debug aid).

6. **NUL handling on macOS bash 3.2.** `read -d ''` is bash-only. zsh handles it. If you support `bash 3.2` (macOS default), `process substitution + read -d` works there too — but worth testing. Fallback: base64-encode the argv on the Go side and decode in shell.

7. **Plugin completion later.** Once `run` is the single entrypoint, adding `avm-bin complete <partial>` for tab-completion is trivial — same in-process cache, just a different output path.

---

## Migration plan (low-risk)

1. **Land `avm-bin run` alongside the existing flow.** Don't remove `which` or the old shell wrapper yet.
2. **Add `shell-init --v2`** that emits the new wrapper. Old `shell-init` keeps the current behavior. Users opt-in.
3. **Dogfood for a release cycle.** Make the v2 wrapper default in install.sh, but leave v1 reachable.
4. **Deprecate `which` parsing.** `avm which` stays as a user-facing inspection command (it's useful), but the shell wrapper no longer parses it.
5. **Remove old code path** once telemetry/feedback says nobody's hitting bugs.

---

## What I'd build first if I were you

```
1. internal/protocol/protocol.go    — emit() helpers, NUL-argv encoder
2. internal/runner/runner.go        — Resolve + Render + Suggest, used by `run`
3. cmd/run.go                       — thin Cobra command wiring it together
4. cmd/shell_init.go (--v2 flag)    — new wrapper
5. internal/runner/runner_test.go   — golden-file tests of protocol output
                                       for: alias hit, alias miss, suggestion
                                       picked, suggestion cancelled, $-args,
                                       weird-char args
```

The golden-file tests are the big win — current code has zero coverage of the resolution/dispatch surface, and protocol output is trivially testable without spawning real plugins.

---

Want me to:
- **A)** write out the full `cmd/run.go` and `internal/protocol/protocol.go` as ready-to-paste files,
- **B)** turn this into a structured agent prompt for Claude Code (audit-first, exact file/line refs, no fixes without approval — the way you usually like them), or
- **C)** draft the golden-file test cases for `runner_test.go` first so you can TDD the redesign?