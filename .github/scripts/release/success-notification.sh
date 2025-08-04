#!/bin/bash
set -e

# Display success notification for auto-release
# Usage: success-notification.sh <old-version> <new-version> <pr-number> <release-type>

OLD_VERSION="$1"
NEW_VERSION="$2"
PR_NUMBER="$3"
RELEASE_TYPE="$4"

if [ -z "$OLD_VERSION" ] || [ -z "$NEW_VERSION" ] || [ -z "$PR_NUMBER" ] || [ -z "$RELEASE_TYPE" ]; then
    echo "Usage: $0 <old-version> <new-version> <pr-number> <release-type>"
    exit 1
fi

echo "🎉 Auto-release completed successfully!"
echo ""
echo "📋 Details:"
echo "  • Version: $OLD_VERSION → $NEW_VERSION"
echo "  • Tag: v$NEW_VERSION"
echo "  • GitHub Release: Created"
echo "  • Triggered by: PR #$PR_NUMBER"
echo "  • Release type: $RELEASE_TYPE"
echo ""
echo "🔗 Check the release: https://github.com/$GITHUB_REPOSITORY/releases/tag/v$NEW_VERSION"
