#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$ROOT_DIR/package.json" | head -n 1)"

if [ -z "$VERSION" ]; then
  echo "Unable to read version from package.json"
  exit 1
fi

if ! grep -q "^## \\[$VERSION\\]" "$ROOT_DIR/CHANGELOG.md"; then
  echo "CHANGELOG.md is missing an entry for package version $VERSION"
  echo "Add a section like: ## [$VERSION] - YYYY-MM-DD"
  exit 1
fi

echo "CHANGELOG.md contains version $VERSION"
