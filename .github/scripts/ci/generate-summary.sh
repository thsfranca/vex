#!/bin/bash
set -e

# Arguments: GO_FILES_CHANGED SKIP_BUILD_RESULT BUILD_TEST_RESULT COVERAGE_RESULT LINT_RESULT

GO_FILES_CHANGED="$1"
SKIP_BUILD_RESULT="$2"
BUILD_TEST_RESULT="$3"
COVERAGE_RESULT="$4"
LINT_RESULT="$5"

echo "## [SUMMARY] CI Summary" >> $GITHUB_STEP_SUMMARY

if [[ "$GO_FILES_CHANGED" == "false" ]]; then
    echo "[SUCCESS] **Fast skip**: No Go-related files changed" >> $GITHUB_STEP_SUMMARY
    echo "[INFO] **Time saved**: ~5-8 minutes" >> $GITHUB_STEP_SUMMARY
    echo "- Skip Build: $SKIP_BUILD_RESULT" >> $GITHUB_STEP_SUMMARY
else
    echo "🔄 **Full CI**: Go-related files changed" >> $GITHUB_STEP_SUMMARY
    echo "- Build & Test: $BUILD_TEST_RESULT" >> $GITHUB_STEP_SUMMARY
    echo "- Coverage: $COVERAGE_RESULT" >> $GITHUB_STEP_SUMMARY  
    echo "- Linting: $LINT_RESULT" >> $GITHUB_STEP_SUMMARY
    echo "" >> $GITHUB_STEP_SUMMARY
    echo "## 🧠 Smart Failure Handling" >> $GITHUB_STEP_SUMMARY
    echo "- [SUCCESS] **Changed files**: Strict error checking" >> $GITHUB_STEP_SUMMARY
    echo "- ⚠️  **Unchanged files**: Warning-only for legacy issues" >> $GITHUB_STEP_SUMMARY
    echo "- [INFO] **Early development**: Graceful failure handling" >> $GITHUB_STEP_SUMMARY
    echo "- [INFO] **Coverage**: Adaptive requirements based on project size" >> $GITHUB_STEP_SUMMARY
    
    # Intelligent failure handling - only fail if jobs actually failed
    if [[ "$BUILD_TEST_RESULT" == "failure" || 
          "$COVERAGE_RESULT" == "failure" || 
          "$LINT_RESULT" == "failure" ]]; then
        echo "" >> $GITHUB_STEP_SUMMARY
        echo "[ERROR] **CI blocked** - Issues found in changed files" >> $GITHUB_STEP_SUMMARY
        echo "Check individual jobs for specific issues that need fixing." >> $GITHUB_STEP_SUMMARY
        exit 1
    else
        echo "" >> $GITHUB_STEP_SUMMARY
        echo "[SUCCESS] **All checks passed** - Ready to merge!" >> $GITHUB_STEP_SUMMARY
    fi
fi