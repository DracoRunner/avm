#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DOCKER_COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
SCENARIO_DIR="$ROOT_DIR/docker/tests/scenarios"

cd "$ROOT_DIR"

log() {
  echo "[avm-test] $*"
}

log "Building Docker image (if needed)"
docker compose -f "$DOCKER_COMPOSE_FILE" build avm-rust-dev

log "Running Rust workspace tests"
docker compose -f "$DOCKER_COMPOSE_FILE" run --rm avm-rust-dev bash -lc "export PATH=/usr/local/cargo/bin:\$PATH && cd /workspace && cargo test --workspace && cargo build --package avm-cli --bin avm-bin"

run_one() {
  local scenario_file="$1"
  local name
  name="$(basename "$scenario_file")"
  log "Running $name"
  docker compose -f "$DOCKER_COMPOSE_FILE" run --rm avm-rust-dev bash -lc "export PATH=/usr/local/cargo/bin:/workspace/target/debug:\$PATH && bash /workspace/docker/tests/scenarios/$name"
}

resolve_scenario() {
  local arg="$1"
  local matches
  if [[ "$arg" == *.sh && -f "$arg" ]]; then
    basename "$arg"
  elif [[ -f "$SCENARIO_DIR/$arg" ]]; then
    basename "$SCENARIO_DIR/$arg"
  elif [[ -f "$SCENARIO_DIR/$arg.sh" ]]; then
    basename "$SCENARIO_DIR/$arg.sh"
  else
    matches=("$SCENARIO_DIR/${arg}"*.sh)
    if [ ! -e "${matches[0]}" ]; then
      echo ""
      return
    fi
    if [ "${#matches[@]}" -eq 1 ]; then
      basename "${matches[0]}"
    else
      echo ""
      return
    fi
  fi
}

if [ $# -gt 0 ]; then
  for arg in "$@"; do
    scenario="$(resolve_scenario "$arg")"
    if [ -z "$scenario" ]; then
      echo "scenario not found: $arg"
      echo "Available:"
      ls -1 "$SCENARIO_DIR"
      exit 1
    fi
    run_one "$SCENARIO_DIR/$scenario"
  done
else
  for scenario in "$SCENARIO_DIR"/*.sh; do
    run_one "$scenario"
  done
fi

log "All requested scenarios passed"
