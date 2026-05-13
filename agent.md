# Using avm with an Agent

This file is for users who want an LLM or coding agent to understand how their project uses `avm`.

## Give your agent this context

Point the agent to these files first:

- `README.md`
- `llm.txt`
- `agent.skill.md`
- `.avm.json` in the current project

## What the agent should know

- `avm` reads local `.avm.json` first.
- If no local value exists, `avm` falls back to global `~/.avm.json`.
- Aliases live under `aliases`.
- Environment values live under `env`.
- Tool versions live under `tools`.
- Plain commands such as `node` resolve through shims when shell integration is active.
- Missing managed tool versions do not auto-install; avm warns and falls back to the system binary.

## Useful prompts

```text
Read .avm.json and explain which aliases, env values, and tools this project uses.
```

```text
Create a clean .avm.json for this project with aliases for dev, test, lint, build, and release.
```

```text
Check why node is not resolving through avm and suggest the commands to fix it.
```
