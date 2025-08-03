#!/bin/bash

# Create release commit and tag
# Usage: ./scripts/create-release-tag.sh <new-version> <old-version> <pr-number> <release-type>

set -e

NEW_VERSION="$1"
OLD_VERSION="$2"
PR_NUMBER="$3"
RELEASE_TYPE="$4"

if [ $# -ne 4 ]; then
    echo "Usage: $0 <new-version> <old-version> <pr-number> <release-type>"
    exit 1
fi

echo "Creating release commit and tag for v$NEW_VERSION"

# Configure git (should be done by caller, but just in case)
git config --local user.email "action@github.com" 2>/dev/null || true
git config --local user.name "GitHub Action" 2>/dev/null || true

# Commit version bump
git add VERSION
git commit -m "release: bump version to $NEW_VERSION

Auto-release triggered by PR #$PR_NUMBER
Release type: $RELEASE_TYPE
Previous version: $OLD_VERSION"

# Create and push tag
git tag "v$NEW_VERSION"
git push origin main
git push origin "v$NEW_VERSION"

echo "🎉 Created and pushed release v$NEW_VERSION"
echo "🔗 Tag: v$NEW_VERSION"
echo "📋 Triggered by PR #$PR_NUMBER ($RELEASE_TYPE release)"