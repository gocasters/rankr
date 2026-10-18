#!/usr/bin/env bash
# The change gate: gofmt, vet, race tests and build for the Go packages this
# branch touches (vs $BASE, default origin/main). The whole repo still has to
# build. Use `make check-all` for every package (red until the baseline issue
# is fixed).
set -euo pipefail

BASE="${BASE:-origin/main}"
base_commit="$(git merge-base "$BASE" HEAD 2>/dev/null || echo "$BASE")"

files="$( {
  git diff --name-only --diff-filter=ACMR "$base_commit" -- '*.go'
  git ls-files --others --exclude-standard -- '*.go'
} | grep -v '^protobuf/golang/' | sort -u || true)"

echo "check: go build ./..."
go build ./...

if [ -z "$files" ]; then
  echo "check: no changed Go files vs $BASE; build OK"
  exit 0
fi

echo "check: gofmt"
unformatted="$(gofmt -l $files)"
if [ -n "$unformatted" ]; then
  echo "gofmt needed (run: gofmt -w <file>):"
  echo "$unformatted"
  exit 1
fi

pkgs="$(for f in $files; do echo "./$(dirname "$f")"; done | sort -u)"
echo "check: go vet + go test -race -short on:"
echo "$pkgs"
go vet $pkgs
go test -race -short -count=1 $pkgs
echo "check: OK"
