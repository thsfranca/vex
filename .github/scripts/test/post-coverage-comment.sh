#!/bin/bash
set -e

echo "[COVERAGE COMMENT] Starting coverage comment process..."

# Check if this is a pull request
if [ "$GITHUB_EVENT_NAME" != "pull_request" ]; then
    echo "[COVERAGE COMMENT] Not a pull request, skipping coverage comment"
    exit 0
fi

# Get PR number from GitHub context
PR_NUMBER="$GITHUB_PR_NUMBER"
if [ -z "$PR_NUMBER" ]; then
    echo "[COVERAGE COMMENT] No PR number found, skipping coverage comment"
    exit 0
fi

# Verify required environment variables
if [ -z "$GITHUB_TOKEN" ]; then
    echo "[COVERAGE COMMENT] ❌ GITHUB_TOKEN not found"
    exit 1
fi

if [ -z "$GITHUB_REPOSITORY" ]; then
    echo "[COVERAGE COMMENT] ❌ GITHUB_REPOSITORY not found"
    exit 1
fi

echo "[COVERAGE COMMENT] Processing PR #$PR_NUMBER for $GITHUB_REPOSITORY"

# Check if coverage report exists
COVERAGE_REPORT="coverage-report.md"
if [ ! -f "$COVERAGE_REPORT" ]; then
    echo "[COVERAGE COMMENT] Coverage report file not found at: $COVERAGE_REPORT"
    echo "[COVERAGE COMMENT] Creating fallback coverage report"
    cat > "$COVERAGE_REPORT" << EOF
## ⚠️ Coverage Report Not Available

Coverage analysis could not be completed for this PR.

**Possible reasons:**
- Tests failed to run
- Coverage data was not generated
- Script configuration issue

Please check the CI logs for more details.
EOF
fi

# Read coverage report content
if [ ! -s "$COVERAGE_REPORT" ]; then
    echo "[COVERAGE COMMENT] Coverage report is empty, skipping comment"
    exit 0
fi

echo "[COVERAGE COMMENT] Coverage report found, length: $(wc -c < "$COVERAGE_REPORT") bytes"

# Comment marker for identifying our coverage comments
COMMENT_MARKER="📊 Test Coverage Report"

# Get existing comments using GitHub API
echo "[COVERAGE COMMENT] Checking for existing coverage comments..."
COMMENTS_RESPONSE=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    "https://api.github.com/repos/$GITHUB_REPOSITORY/issues/$PR_NUMBER/comments")

# Find existing coverage comment
EXISTING_COMMENT_ID=$(echo "$COMMENTS_RESPONSE" | grep -B5 -A5 "$COMMENT_MARKER" | grep '"id":' | head -n1 | sed 's/.*"id": *\([0-9]*\).*/\1/' || echo "")

# Prepare the comment body for JSON (preserve formatting)
COMMENT_BODY=$(jq -Rs . "$COVERAGE_REPORT")

if [ -n "$EXISTING_COMMENT_ID" ]; then
    echo "[COVERAGE COMMENT] Found existing coverage comment: $EXISTING_COMMENT_ID"
    echo "[COVERAGE COMMENT] Updating existing comment..."
    
    HTTP_CODE=$(curl -s -w "%{http_code}" -o /tmp/comment_response.json \
        -X PATCH \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        -H "Content-Type: application/json" \
        -d "{\"body\": $COMMENT_BODY}" \
        "https://api.github.com/repos/$GITHUB_REPOSITORY/issues/comments/$EXISTING_COMMENT_ID")
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo "[COVERAGE COMMENT] ✅ Successfully updated existing coverage comment"
    else
        echo "[COVERAGE COMMENT] ❌ Failed to update existing comment (HTTP $HTTP_CODE)"
        cat /tmp/comment_response.json 2>/dev/null || true
        exit 1
    fi
else
    echo "[COVERAGE COMMENT] No existing coverage comment found"
    echo "[COVERAGE COMMENT] Creating new coverage comment..."
    
    HTTP_CODE=$(curl -s -w "%{http_code}" -o /tmp/comment_response.json \
        -X POST \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        -H "Content-Type: application/json" \
        -d "{\"body\": $COMMENT_BODY}" \
        "https://api.github.com/repos/$GITHUB_REPOSITORY/issues/$PR_NUMBER/comments")
    
    if [ "$HTTP_CODE" = "201" ]; then
        echo "[COVERAGE COMMENT] ✅ Successfully created new coverage comment"
    else
        echo "[COVERAGE COMMENT] ❌ Failed to create coverage comment (HTTP $HTTP_CODE)"
        cat /tmp/comment_response.json 2>/dev/null || true
        exit 1
    fi
fi

# Cleanup
rm -f /tmp/comment_response.json

echo "[COVERAGE COMMENT] Coverage comment process completed successfully"