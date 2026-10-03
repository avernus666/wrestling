#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

command -v go >/dev/null || { echo 'Go is not installed.' >&2; exit 1; }
command -v npm >/dev/null || { echo 'Node.js/npm is not installed.' >&2; exit 1; }

./scripts/check-go-coverage.sh
if command -v git >/dev/null && [[ -d "$ROOT/.git" ]]; then git -C "$ROOT" diff --check; fi

if [[ ! -d node_modules ]]; then npm install; fi
npm test -- --runInBand
npm run build
