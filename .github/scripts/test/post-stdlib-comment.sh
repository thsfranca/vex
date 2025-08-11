#!/bin/bash

set -e

echo "💬 Posting stdlib test results to PR"

# Only run for pull requests
if [ "$GITHUB_EVENT_NAME" != "pull_request" ]; then
    echo "⏭️ Skipping PR comment - not a pull request"
    exit 0
fi

# Check if we have the required environment variables
if [ -z "$GITHUB_TOKEN" ] || [ -z "$GITHUB_PR_NUMBER" ] || [ -z "$GITHUB_REPOSITORY" ]; then
    echo "❌ Missing required environment variables for PR commenting"
    echo "   GITHUB_TOKEN: $([ -n "$GITHUB_TOKEN" ] && echo "✓" || echo "✗")"
    echo "   GITHUB_PR_NUMBER: $([ -n "$GITHUB_PR_NUMBER" ] && echo "✓" || echo "✗")"
    echo "   GITHUB_REPOSITORY: $([ -n "$GITHUB_REPOSITORY" ] && echo "✓" || echo "✗")"
    exit 1
fi

# Check if we have the coverage report
if [ ! -f "stdlib-coverage-report.md" ]; then
    echo "⚠️ No stdlib coverage report found - creating minimal comment"
    
    cat > stdlib-coverage-report.md << EOF
## 📚 Stdlib Test Report

**Status:** ⚠️ No Report Generated  

The stdlib test report could not be generated. This may indicate:
- No stdlib files were changed
- Tests failed to run
- Report generation encountered an error

EOF
fi

echo "📄 Reading coverage report..."
COMMENT_BODY=$(cat stdlib-coverage-report.md)

# Create the comment using GitHub API
COMMENT_JSON=$(jq -n --arg body "$COMMENT_BODY" '{body: $body}')

echo "📤 Posting comment to PR #$GITHUB_PR_NUMBER..."

# Use curl to post the comment
HTTP_STATUS=$(curl -s -w "%{http_code}" -o response.json \
    -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: application/json" \
    -d "$COMMENT_JSON" \
    "https://api.github.com/repos/$GITHUB_REPOSITORY/issues/$GITHUB_PR_NUMBER/comments")

echo "📬 GitHub API response status: $HTTP_STATUS"

if [ "$HTTP_STATUS" -eq 201 ]; then
    echo "✅ Successfully posted stdlib test results to PR #$GITHUB_PR_NUMBER"
    COMMENT_ID=$(jq -r '.id' response.json)
    echo "🔗 Comment ID: $COMMENT_ID"
elif [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ Successfully updated stdlib test results comment"
else
    echo "❌ Failed to post comment. HTTP Status: $HTTP_STATUS"
    echo "📄 Response:"
    cat response.json
    exit 1
fi

# Clean up
rm -f response.json

echo "🏁 PR comment completed!"
