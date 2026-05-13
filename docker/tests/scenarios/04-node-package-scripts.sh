#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$SCRIPT_DIR/lib.sh"

WORKDIR="$(mk_workdir)"
trap 'rm -rf "$WORKDIR"' EXIT

log "Scenario 04: node package.json script aliases"
write_json_file "$WORKDIR/package.json" '{
  "name": "avm-smoke-node",
  "version": "1.0.0",
  "scripts": {
    "start": "echo from-start",
    "build": "echo from-build"
  }
}'
touch "$WORKDIR/pnpm-lock.yaml"

# Local project config with no explicit alias for start
write_json_file "$WORKDIR/.avm.json" '{
  "aliases": {},
  "env": {},
  "tools": {}
}'

out="$(run_avm "$WORKDIR" which start)"
assert_contains "$out" "plugin alias 'start' from node" "node script alias should come from provider"
assert_contains "$out" "pnpm run start" "pnpm lockfile should switch manager"

out="$(run_avm "$WORKDIR" which build)"
assert_contains "$out" "plugin alias 'build' from node" "second script alias should also resolve"

log "Scenario 04 passed"
