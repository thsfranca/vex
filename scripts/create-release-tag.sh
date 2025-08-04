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

echo "Current git status:"
git status

echo "Checking if VERSION file has changes:"
git diff --name-only
git diff VERSION || echo "No changes to VERSION file"

# Commit version bump
git add VERSION

# Check if there are changes to commit
if git diff --cached --quiet; then
    echo "⚠️ No changes to commit - VERSION file may not have been modified"
    exit 1
fi

git commit -m "release: bump version to $NEW_VERSION

Auto-release triggered by PR #$PR_NUMBER
Release type: $RELEASE_TYPE
Previous version: $OLD_VERSION"

echo "Commit created successfully, creating tag..."

# Create and push tag
git tag "v$NEW_VERSION"

echo "Pushing commit and tag to remote..."
git push origin main
git push origin "v$NEW_VERSION"

echo "🎉 Created and pushed release v$NEW_VERSION"
echo "🔗 Tag: v$NEW_VERSION"
echo "📋 Triggered by PR #$PR_NUMBER ($RELEASE_TYPE release)"
