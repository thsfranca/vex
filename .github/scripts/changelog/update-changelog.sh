#!/bin/bash
set -euo pipefail

# Update CHANGELOG.md by prepending a new entry from a release
# Usage: update-changelog.sh <tag> <name> <date> <body-file>

TAG="$1"
NAME="$2"
DATE="$3"
BODY_FILE="$4"

if [ -z "${TAG:-}" ] || [ -z "${NAME:-}" ] || [ -z "${DATE:-}" ] || [ -z "${BODY_FILE:-}" ]; then
  echo "Usage: $0 <tag> <name> <date> <body-file>"
  exit 1
fi

if [ ! -f "$BODY_FILE" ]; then
  echo "❌ Body file not found: $BODY_FILE"
  exit 1
fi

# Prepare the new entry content
NEW_ENTRY=$(cat <<EOF

## [$TAG] - $DATE

$(cat "$BODY_FILE")

EOF
)

# Insert after the Unreleased section header if present; otherwise, append at top after the main title
if grep -q '^## \[Unreleased\]' CHANGELOG.md; then
  # macOS/BSD sed compatibility: use temp file
  TMP_FILE="$(mktemp)"
  awk -v entry="$NEW_ENTRY" '
    BEGIN {printed=0}
    {
      print $0
      if (!printed && $0 ~ /^## \[Unreleased\]/) {
        print entry
        printed=1
      }
    }
  ' CHANGELOG.md > "$TMP_FILE"
  mv "$TMP_FILE" CHANGELOG.md
else
  # Prepend after first line (title) if Unreleased not found
  TMP_FILE="$(mktemp)"
  {
    read -r first_line
    echo "$first_line"
    echo "$NEW_ENTRY"
    cat
  } < CHANGELOG.md > "$TMP_FILE"
  mv "$TMP_FILE" CHANGELOG.md
fi

echo "✅ CHANGELOG.md updated for $TAG"


