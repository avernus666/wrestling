#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
command -v go >/dev/null || { echo 'Go не установлен: https://go.dev/dl/' >&2; exit 1; }
command -v npm >/dev/null || { echo 'Node.js/npm не установлен: https://nodejs.org/' >&2; exit 1; }
mkdir -p server-go/bin
(cd server-go && go test ./... && go build -o bin/server ./cmd/server)
if [[ ! -d node_modules ]]; then npm install; fi
npm run build
printf '\nСборка завершена. API: server-go/bin/server; frontend: build/\n'
