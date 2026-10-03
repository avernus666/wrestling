#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT/server-go"
go test -coverprofile=coverage.out ./... >/dev/null
go tool cover -func=coverage.out
coverage="$(go tool cover -func=coverage.out | awk '/^total:/ {gsub(/%/, "", $3); print $3}')"
awk -v coverage="$coverage" 'BEGIN { if (coverage < 70) { printf "Coverage %.1f%% is below the 70%% quality gate.\n", coverage; exit 1 } printf "Coverage quality gate passed: %.1f%%\n", coverage }'
