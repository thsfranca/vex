#!/bin/bash

set -e

echo "📊 Validating Stdlib Coverage"
echo "============================="

# Configuration
STDLIB_COVERAGE_THRESHOLD=${STDLIB_COVERAGE_THRESHOLD:-98}
COVERAGE_FILE="stdlib-coverage.json"

# Check if coverage file exists
if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Error: Stdlib coverage file not found: $COVERAGE_FILE"
    echo "Make sure to run stdlib tests first to generate coverage data."
    exit 1
fi

echo "📄 Reading coverage data from: $COVERAGE_FILE"

# Extract coverage data
if command -v jq >/dev/null 2>&1; then
    OVERALL_COVERAGE=$(jq -r '.overall_coverage // 0' $COVERAGE_FILE)
    TOTAL_FILES=$(jq -r '.total_files // 0' $COVERAGE_FILE)
    TESTED_FILES=$(jq -r '.tested_files // 0' $COVERAGE_FILE)
    PACKAGE_COUNT=$(jq -r '.packages | length' $COVERAGE_FILE)
else
    echo "⚠️  jq not available, using basic validation"
    OVERALL_COVERAGE=0
    TOTAL_FILES=0
    TESTED_FILES=0
    PACKAGE_COUNT=0
fi

echo ""
echo "📈 Coverage Analysis:"
echo "   Overall Coverage: ${OVERALL_COVERAGE}%"
echo "   Total Files: $TOTAL_FILES"
echo "   Tested Files: $TESTED_FILES"
echo "   Packages: $PACKAGE_COUNT"
echo "   Coverage Threshold: ${STDLIB_COVERAGE_THRESHOLD}%"

# Validate coverage threshold
COVERAGE_STATUS="passed"
if [ "$OVERALL_COVERAGE" -lt "$STDLIB_COVERAGE_THRESHOLD" ]; then
    COVERAGE_STATUS="failed"
fi

# Generate detailed coverage report
echo ""
echo "📋 Package-Level Coverage:"

if command -v jq >/dev/null 2>&1 && [ "$PACKAGE_COUNT" -gt 0 ]; then
    # Show per-package coverage
    jq -r '.packages[] | "   \(.package): \(.coverage)% (\(.test_count)/\(.file_count) files)"' $COVERAGE_FILE
else
    echo "   No detailed package data available"
fi

# Update the markdown report with coverage details
COVERAGE_ICON="✅"
COVERAGE_TEXT="Coverage threshold met!"

if [ "$COVERAGE_STATUS" = "failed" ]; then
    COVERAGE_ICON="❌"
    COVERAGE_TEXT="Coverage below threshold"
fi

# Append coverage details to existing report
if [ -f "stdlib-coverage-report.md" ]; then
    cat >> stdlib-coverage-report.md << EOF

### 📊 Coverage Details

**Current Coverage:** \`${OVERALL_COVERAGE}%\`  
**Target Coverage:** \`${STDLIB_COVERAGE_THRESHOLD}%\`  
**Status:** ${COVERAGE_ICON} ${COVERAGE_TEXT}

EOF

    if command -v jq >/dev/null 2>&1 && [ "$PACKAGE_COUNT" -gt 0 ]; then
        echo "#### Package Coverage" >> stdlib-coverage-report.md
        echo "" >> stdlib-coverage-report.md
        jq -r '.packages[] | "- **\(.package)**: \(.coverage)% (\(.test_count)/\(.file_count) files)"' $COVERAGE_FILE >> stdlib-coverage-report.md
        echo "" >> stdlib-coverage-report.md
    fi

    if [ "$COVERAGE_STATUS" = "failed" ]; then
        cat >> stdlib-coverage-report.md << EOF
⚠️ **Coverage below threshold!** Please add more tests to improve stdlib coverage.

EOF
    fi
fi

# Set output for GitHub Actions
if [ -n "$GITHUB_OUTPUT" ]; then
    echo "coverage=$OVERALL_COVERAGE" >> $GITHUB_OUTPUT
    echo "coverage-status=$COVERAGE_STATUS" >> $GITHUB_OUTPUT
    echo "total-files=$TOTAL_FILES" >> $GITHUB_OUTPUT
    echo "tested-files=$TESTED_FILES" >> $GITHUB_OUTPUT
fi

echo ""
if [ "$COVERAGE_STATUS" = "passed" ]; then
    echo "✅ Stdlib coverage validation passed!"
    echo "   Coverage ${OVERALL_COVERAGE}% meets threshold ${STDLIB_COVERAGE_THRESHOLD}%"
else
    echo "❌ Stdlib coverage validation failed!"
    echo "   Coverage ${OVERALL_COVERAGE}% is below threshold ${STDLIB_COVERAGE_THRESHOLD}%"
    echo "💡 Add more tests to improve stdlib coverage"
fi

echo "📄 Updated coverage report: stdlib-coverage-report.md"

# Don't exit with error here - let the workflow handle the failure
# This allows the comment to be posted before failing the build
exit 0
