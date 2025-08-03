#!/bin/bash
set -e

if ! find . -name "*.go" | grep -q .; then
    echo "[INFO] No Go files found - skipping linting"
    exit 0
fi

echo "[LINT] Running Go linting..."

# Get list of changed Go files
if [ "$GITHUB_EVENT_NAME" = "pull_request" ]; then
    CHANGED_GO_FILES=$(git diff --name-only $GITHUB_BASE_SHA..HEAD | grep '\.go$' || true)
else
    CHANGED_GO_FILES=$(find . -name "*.go")
fi

# Track if we found any blocking issues in changed files
BLOCKING_ISSUES=0

if [ -n "$CHANGED_GO_FILES" ]; then
    echo "Linting changed files: $CHANGED_GO_FILES"
    
    # Run golint on changed files only - these are blocking
    for file in $CHANGED_GO_FILES; do
        if [ -f "$file" ]; then
            golint_output=$(golint "$file" || true)
            if [ -n "$golint_output" ]; then
                echo "[ERROR] Linting issues in changed file: $file"
                echo "$golint_output"
                BLOCKING_ISSUES=1
            fi
        fi
    done
    
    # Run golint on all files for informational purposes
    echo ""
    echo "[LINT] Checking all files (warnings for unchanged files)..."
    golint ./... > /tmp/all_golint.out 2>/dev/null || true
    
    if [ -s /tmp/all_golint.out ]; then
        echo "⚠️ Additional linting suggestions (not blocking):"
        cat /tmp/all_golint.out
    fi
    
    if [ $BLOCKING_ISSUES -eq 1 ]; then
        echo ""
        echo "[ERROR] Linting issues found in changed files - these must be fixed"
        exit 1
    else
        echo "[SUCCESS] No blocking linting issues in changed files"
    fi
else
    echo "[INFO] No Go files changed in this PR"
fi