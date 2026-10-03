#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

command -v go >/dev/null || { echo 'Go не установлен: https://go.dev/dl/' >&2; exit 1; }
command -v npm >/dev/null || { echo 'Node.js/npm не установлен: https://nodejs.org/' >&2; exit 1; }

if [[ -z "${DATABASE_URL:-}" ]]; then
  if command -v docker >/dev/null && docker compose version >/dev/null 2>&1; then
    docker compose up -d postgres
    DATABASE_URL="postgres://wrestling:wrestling@localhost:5432/wrestling?sslmode=disable"
  else
    echo "DATABASE_URL не задан. Запусти PostgreSQL или docker compose up -d postgres." >&2
    exit 1
  fi
fi

mkdir -p server-go/bin
(cd server-go && go build -o bin/server ./cmd/server)

PORT="${PORT:-8080}" DATABASE_URL="$DATABASE_URL" PUBLIC_DIR="$ROOT/public" WEB_DIR="$ROOT/build" ./server-go/bin/server &
BACKEND_PID=$!
cleanup() { kill "$BACKEND_PID" 2>/dev/null || true; wait "$BACKEND_PID" 2>/dev/null || true; }
trap cleanup EXIT INT TERM

if [[ ! -d node_modules ]]; then npm install; fi
npm start
