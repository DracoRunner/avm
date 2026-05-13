#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$SCRIPT_DIR/lib.sh"

WORKDIR="$(mk_workdir)"
trap 'rm -rf "$WORKDIR"' EXIT

log "Scenario 03: shim fallback to system binary"
write_json_file "$WORKDIR/.avm.json" '{
  "aliases": {
    "say": "echo alias-call"
  },
  "tools": {
    "node": "99.9.9-missing"
  }
}'

run_avm "$WORKDIR" shims install
mkdir -p "$WORKDIR/fake-bin"
cat > "$WORKDIR/fake-bin/node" <<'EOF'
#!/usr/bin/env bash
echo "system-node:$*"
EOF
chmod +x "$WORKDIR/fake-bin/node"

out="$(cd "$WORKDIR" && PATH="$WORKDIR/fake-bin:$HOME/.avm/shims:$PATH" "$HOME/.avm/shims/node" -v 2>&1)"
assert_contains "$out" "warning: managed node 99.9.9-missing is not installed; falling back to system node" "should warn when managed node is missing"
assert_contains "$out" "system-node:-v" "should execute system fallback node"

run_avm "$WORKDIR" shims remove node
log "Scenario 03 passed"
