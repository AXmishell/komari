#!/usr/bin/env bash
# Rebuild the embedded default frontend (web/public/defaultTheme) from komari-web.
#
# Mirrors .github/actions/build-frontend so local and CI produce the same archive,
# then runs the self-consistency check via cmd/pack-frontend.
#
# Requirements: git, node, npm, go.
# Environment overrides:
#   KOMARI_WEB_REPO  git URL of the frontend (default: AXmishell/komari-web)
#   KOMARI_WEB_REF   branch/tag/sha to build (default: repository default branch)
#   KOMARI_WEB_DIR   work directory (default: a fresh temp directory)

set -euo pipefail

REPO_URL="${KOMARI_WEB_REPO:-https://github.com/AXmishell/komari-web}"
REF="${KOMARI_WEB_REF:-}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -n "${KOMARI_WEB_DIR:-}" ]; then
  WORK_DIR="$KOMARI_WEB_DIR"
  mkdir -p "$WORK_DIR"
else
  WORK_DIR="$(mktemp -d)"
fi

SRC="$WORK_DIR/komari-web"
rm -rf "$SRC"

if [ -n "$REF" ]; then
  git clone --depth=1 --branch "$REF" "$REPO_URL" "$SRC"
else
  git clone --depth=1 "$REPO_URL" "$SRC"
fi

(
  cd "$SRC"
  npm install
  npm run build
)

cd "$ROOT"
go run ./cmd/pack-frontend \
  -dist "$SRC/dist" \
  -out web/public/defaultTheme/dist.tar.zst \
  -theme "$SRC/komari-theme.json"

echo "Done. Build the server with: go build -o komari ."
