#!/bin/bash

# Create and push a release tag (no commit of VERSION file)
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

echo "Creating release tag v$NEW_VERSION"

# Configure git (should be done by caller, but just in case)
git config --local user.email "action@github.com" 2>/dev/null || true
git config --local user.name "GitHub Action" 2>/dev/null || true

# Validate version format (semver or semver with prerelease)
if [[ ! "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+([-][A-Za-z]+\.[0-9]+)?$ ]]; then
  echo "❌ Invalid version format: $NEW_VERSION"
  exit 1
fi

# Check if tag already exists
if git rev-parse "v$NEW_VERSION" >/dev/null 2>&1; then
  echo "⚠️ Tag v$NEW_VERSION already exists, nothing to do"
  exit 0
fi

# Create and push tag
git tag -m "Release $NEW_VERSION (from PR #$PR_NUMBER, $RELEASE_TYPE)" "v$NEW_VERSION"
git push origin "v$NEW_VERSION"

echo "🎉 Created and pushed tag v$NEW_VERSION"
echo "🔗 Tag: v$NEW_VERSION"
echo "📋 Triggered by PR #$PR_NUMBER ($RELEASE_TYPE release)"
