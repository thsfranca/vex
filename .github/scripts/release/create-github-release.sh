#!/bin/bash
set -e

# Script to create GitHub Release with artifacts
# Usage: create-github-release.sh <version> <release_type>

VERSION="$1"
RELEASE_TYPE="$2"

if [ -z "$VERSION" ] || [ -z "$RELEASE_TYPE" ]; then
    echo "Error: Missing required parameters"
    echo "Usage: $0 <version> <release_type>"
    exit 1
fi

echo "🚀 Creating GitHub Release..."

# Determine if this is a prerelease
PRERELEASE_FLAG=""
if [[ "$RELEASE_TYPE" =~ ^(alpha|beta|rc)$ ]]; then
    PRERELEASE_FLAG="--prerelease"
fi

# Create placeholder artifacts
mkdir -p dist
echo "# Vex Language v${VERSION}" > dist/README.md
echo "Basic transpiler with working features:" >> dist/README.md
echo "- Variable definitions: (def x 10) → x := 10" >> dist/README.md
echo "- Arithmetic expressions: (+ 1 2) → 1 + 2" >> dist/README.md
echo "- CLI tool: vex-transpiler" >> dist/README.md

# Package examples
tar -czf "dist/vex-examples-v${VERSION}.tar.gz" examples/

# Read release notes from temporary file
RELEASE_NOTES=$(cat /tmp/release-notes.md)

# Create the GitHub Release
gh release create "v${VERSION}" \
    --title "Vex v${VERSION}" \
    --notes "$RELEASE_NOTES" \
    $PRERELEASE_FLAG \
    dist/*

echo "✅ GitHub Release v${VERSION} created successfully"