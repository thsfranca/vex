#!/bin/bash
set -euo pipefail

# Wrapper that delegates to the Go release-manager
# Usage: create-github-release.sh <version> <release_type>

if [ $# -ne 2 ]; then
  echo "Usage: $0 <version> <release_type>"
  exit 1
fi

VERSION="$1"
RELEASE_TYPE="$2"

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$REPO_ROOT/tools/release-manager"

go build -o release-manager .
./release-manager publish-release "$VERSION" "$RELEASE_TYPE"
