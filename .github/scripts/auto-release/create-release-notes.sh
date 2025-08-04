#!/bin/bash
set -e

# Create release notes for GitHub Release
# Usage: create-release-notes.sh <new-version> <old-version> <pr-number> <pr-title> <pr-body> <pr-author> <release-type>

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

cd tools/release-manager

# Create PR data JSON for the tool
PR_DATA=$(cat << EOF
{
  "number": $PR_NUMBER,
  "title": "$PR_TITLE",
  "body": "$PR_BODY",
  "author": "$PR_AUTHOR",
  "release_type": "$RELEASE_TYPE"
}
EOF
)

# Create release notes via tool
./release-manager create-notes "$PR_DATA"

# Create release notes for GitHub Release using the dedicated script
cd ../..
.github/scripts/release/create-release-notes.sh \
  "$NEW_VERSION" \
  "$OLD_VERSION" \
  "$PR_NUMBER" \
  "$PR_TITLE" \
  "$RELEASE_TYPE"

echo "✅ Release notes created successfully"