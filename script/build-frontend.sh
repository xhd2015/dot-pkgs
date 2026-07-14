#!/usr/bin/env bash
# Build dot-pkgs-react and copy dist into go-pkgs/wrkcli/web/dist for //go:embed.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REACT="$ROOT/react"
DST="$ROOT/go-pkgs/wrkcli/web/dist"

if ! command -v bun >/dev/null 2>&1; then
  echo "bun is not installed; see https://bun.sh/docs/installation" >&2
  exit 1
fi

if [[ ! -d "$REACT/node_modules" ]]; then
  (cd "$REACT" && bun install)
fi

(cd "$REACT" && bun run build)

rm -rf "$DST"
mkdir -p "$DST"
cp -R "$REACT/dist/." "$DST/"
echo "built frontend → $DST"
