#!/bin/bash
set -euo pipefail

# Create release notes for GitHub Release
# Usage: auto-create-release-notes.sh <new-version> <old-version> <pr-number> <pr-title> <pr-body> <pr-author> <release-type>

NEW_VERSION="$1"
OLD_VERSION="$2"
PR_NUMBER="$3"
PR_TITLE="$4"
PR_BODY="$5"
PR_AUTHOR="$6"
RELEASE_TYPE="$7"

if [ -z "$NEW_VERSION" ] || [ -z "$OLD_VERSION" ] || [ -z "$PR_NUMBER" ] || [ -z "$PR_TITLE" ] || [ -z "$RELEASE_TYPE" ]; then
    echo "Usage: $0 <new-version> <old-version> <pr-number> <pr-title> <pr-body> <pr-author> <release-type>"
    exit 1
fi

echo "📝 Creating release notes..."

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$REPO_ROOT/tools/release-manager"

# Build the release-manager tool
go build -o release-manager .

# Create PR data JSON for the tool using jq to properly escape values
PR_DATA=$(jq -n \
  --arg number "$PR_NUMBER" \
  --arg title "$PR_TITLE" \
  --arg body "$PR_BODY" \
  --arg author "$PR_AUTHOR" \
  --arg release_type "$RELEASE_TYPE" \
  '{
    number: ($number | tonumber),
    title: $title,
    body: $body,
    author: $author,
    release_type: $release_type
  }')

# Create release notes via tool
./release-manager create-notes "$PR_DATA"

echo "Using release notes at: $(pwd)/release-notes.md"
cp release-notes.md /tmp/release-notes.md

echo "✅ Release notes created successfully"
