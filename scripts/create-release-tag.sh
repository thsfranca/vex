#!/bin/bash
set -euo pipefail

# Wrapper to delegate tag creation to the Go release-manager
# Usage: ./scripts/create-release-tag.sh <new-version> <old-version> <pr-number> <release-type>

if [ $# -ne 4 ]; then
  echo "Usage: $0 <new-version> <old-version> <pr-number> <release-type>"
  exit 1
fi

NEW_VERSION="$1"
OLD_VERSION="$2"
PR_NUMBER="$3"
RELEASE_TYPE="$4"

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$REPO_ROOT/tools/release-manager"

go build -o release-manager .
./release-manager create-tag "$NEW_VERSION" "$OLD_VERSION" "$PR_NUMBER" "$RELEASE_TYPE"
