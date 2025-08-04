#!/bin/bash
set -e

# Script to create release notes for GitHub Release
# Usage: create-release-notes.sh <new_version> <old_version> <pr_number> <pr_title> <release_type>

NEW_VERSION="$1"
OLD_VERSION="$2"
PR_NUMBER="$3"
PR_TITLE="$4"
RELEASE_TYPE="$5"

if [ -z "$NEW_VERSION" ] || [ -z "$OLD_VERSION" ] || [ -z "$PR_NUMBER" ] || [ -z "$PR_TITLE" ] || [ -z "$RELEASE_TYPE" ]; then
    echo "Error: Missing required parameters"
    echo "Usage: $0 <new_version> <old_version> <pr_number> <pr_title> <release_type>"
    exit 1
fi

echo "📝 Creating release notes..."

# Create release notes and output to file
cat > /tmp/release-notes.md << EOF
## Release v${NEW_VERSION}

**Release Type**: ${RELEASE_TYPE}

**Changes:**
- ${PR_TITLE}

**Full Changelog**: https://github.com/${GITHUB_REPOSITORY}/compare/v${OLD_VERSION}...v${NEW_VERSION}
EOF

echo "Release notes created successfully"
cat /tmp/release-notes.md
