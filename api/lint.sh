#!/usr/bin/env bash
# Lint each Go file in the api/ directory.
#
# Vercel-go compiles each handler in isolation, so files cannot be linted
# as a package. We use `go vet FILE.go` per handler (same as `go build FILE.go`).
# dev_server.go is compiled as a standalone program and is also vetted.
#
# Usage: bash api/lint.sh   (from project root)

set -euo pipefail

GOPATH_BIN="$(go env GOPATH)/bin"
FILES=(report.go resolve.go stats.go history.go interactions.go dev_server.go)

cd "$(dirname "$0")"

ok=true

echo "=== go vet ==="
for f in "${FILES[@]}"; do
  echo "── $f"
  if ! go vet "$f" 2>&1; then
    ok=false
  fi
done

if command -v "$GOPATH_BIN/staticcheck" &>/dev/null; then
  echo ""
  echo "=== staticcheck ==="
  for f in "${FILES[@]}"; do
    echo "── $f"
    if ! "$GOPATH_BIN/staticcheck" "$f" 2>&1; then
      ok=false
    fi
  done
fi

if command -v "$GOPATH_BIN/errcheck" &>/dev/null; then
  echo ""
  echo "=== errcheck ==="
  for f in "${FILES[@]}"; do
    echo "── $f"
    if ! "$GOPATH_BIN/errcheck" -exclude .errcheck_excludes "$f" 2>&1; then
      ok=false
    fi
  done
fi

echo ""
if $ok; then
  echo "✓ all checks passed"
else
  echo "✗ lint errors found"
  exit 1
fi
