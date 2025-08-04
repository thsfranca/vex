#!/bin/bash
set -e

# Check for release labels in PR
# Usage: check-release-labels.sh <labels-json> <pr-number> <pr-title>

LABELS_JSON="$1"
PR_NUMBER="$2"
PR_TITLE="$3"

if [ -z "$LABELS_JSON" ] || [ -z "$PR_NUMBER" ] || [ -z "$PR_TITLE" ]; then
    echo "Usage: $0 <labels-json> <pr-number> <pr-title>"
    exit 1
fi

echo "🔍 Checking PR labels for release triggers..."
echo "PR Number: $PR_NUMBER"
echo "PR Title: $PR_TITLE"
echo "PR Labels: $LABELS_JSON"

cd tools/release-manager
go build -o release-manager .

echo "Parsed labels: $LABELS_JSON"
OUTPUT=$(./release-manager check-labels "$LABELS_JSON")
echo "$OUTPUT"

# Extract release-type from output and set GitHub output
RELEASE_TYPE=$(echo "$OUTPUT" | grep "^release-type=" | cut -d'=' -f2)
echo "Detected release type: $RELEASE_TYPE"
echo "release-type=$RELEASE_TYPE" >> $GITHUB_OUTPUT