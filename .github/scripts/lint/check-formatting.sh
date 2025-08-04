#!/bin/bash
set -e

if ! find . -name "*.go" | grep -q .; then
    echo "[INFO] No Go files found - skipping formatting check"
    exit 0
fi

echo "[CHECK] Checking Go code formatting..."

# Get list of changed Go files
if [ -n "$GITHUB_BASE_SHA" ]; then
    CHANGED_GO_FILES=$(git diff --name-only $GITHUB_BASE_SHA..HEAD | grep '\.go$' || true)
    echo "Checking changed files since $GITHUB_BASE_SHA"
else
    CHANGED_GO_FILES=$(find . -name "*.go")
    echo "No base SHA available - checking all Go files"
fi

if [ -n "$CHANGED_GO_FILES" ]; then
    echo "Changed Go files: $CHANGED_GO_FILES"
    echo "$CHANGED_GO_FILES" | xargs goimports -l > /tmp/goimports.out || true
    
    if [ -s /tmp/goimports.out ]; then
        echo "[ERROR] Formatting issues found in changed files:"
        cat /tmp/goimports.out
        echo ""
        echo "🔧 Run: goimports -w $(echo $CHANGED_GO_FILES | tr '\n' ' ')"
        exit 1
    else
        echo "[SUCCESS] All changed Go files are properly formatted"
    fi
    
    # Check all files but only warn for unchanged ones
    goimports -l . > /tmp/all_goimports.out || true
    if [ -s /tmp/all_goimports.out ]; then
        echo ""
        echo "⚠️ Additional formatting issues found in unchanged files (not blocking):"
        cat /tmp/all_goimports.out
    fi
else
    echo "[INFO] No Go files changed in this PR"
fi