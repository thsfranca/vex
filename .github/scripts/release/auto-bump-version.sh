#!/bin/bash
set -e

# Bump version using release manager tool
# Usage: bump-version.sh <release-type>

RELEASE_TYPE="$1"

if [ -z "$RELEASE_TYPE" ]; then
    echo "Usage: $0 <release-type>"
    exit 1
fi

echo "⬆️ Bumping version..."
cd tools/release-manager

# Build the release-manager tool
go build -o release-manager .

# Bump version and capture output
OUTPUT=$(./release-manager bump-version "$RELEASE_TYPE")
echo "$OUTPUT"

# Extract version info from output and set GitHub outputs
OLD_VERSION=$(echo "$OUTPUT" | grep "^old-version=" | cut -d'=' -f2)
NEW_VERSION=$(echo "$OUTPUT" | grep "^new-version=" | cut -d'=' -f2)

echo "old-version=$OLD_VERSION" >> $GITHUB_OUTPUT
echo "new-version=$NEW_VERSION" >> $GITHUB_OUTPUT

echo "✅ Version bumped from $OLD_VERSION to $NEW_VERSION"
