#!/bin/bash
set -euo pipefail

# Compute next version using release manager (from latest git tag)
# Usage: auto-bump-version.sh <release-type>

if [ $# -ne 1 ]; then
  echo "Usage: $0 <release-type>"
  exit 1
fi

RELEASE_TYPE="$1"

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$REPO_ROOT/tools/release-manager"

go build -o release-manager .

OUTPUT=$(./release-manager bump-version "$RELEASE_TYPE")
echo "$OUTPUT"

OLD_VERSION=$(echo "$OUTPUT" | grep "^old-version=" | cut -d'=' -f2)
NEW_VERSION=$(echo "$OUTPUT" | grep "^new-version=" | cut -d'=' -f2)

echo "old-version=$OLD_VERSION" >> "$GITHUB_OUTPUT"
echo "new-version=$NEW_VERSION" >> "$GITHUB_OUTPUT"

echo "✅ Computed next version: $OLD_VERSION -> $NEW_VERSION"
