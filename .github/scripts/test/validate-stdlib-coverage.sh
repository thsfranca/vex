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

# Extract coverage data (handle both legacy and enhanced formats)
if command -v jq >/dev/null 2>&1; then
    # Check if this is enhanced coverage format
    if jq -e '.overall_coverage.function_coverage' $COVERAGE_FILE >/dev/null 2>&1; then
        # Enhanced format - prioritize function coverage
        FUNCTION_COVERAGE=$(jq -r '.overall_coverage.function_coverage // 0' $COVERAGE_FILE)
        FILE_COVERAGE=$(jq -r '.overall_coverage.file_coverage // 0' $COVERAGE_FILE)
        OVERALL_COVERAGE=$(echo "$FUNCTION_COVERAGE" | cut -d. -f1)  # Use function coverage as primary
        TOTAL_FILES=$(jq -r '.overall_coverage.total_files // 0' $COVERAGE_FILE)
        TESTED_FILES=$(jq -r '.overall_coverage.tested_files // 0' $COVERAGE_FILE)
        TOTAL_FUNCTIONS=$(jq -r '.overall_coverage.total_functions // 0' $COVERAGE_FILE)
        TESTED_FUNCTIONS=$(jq -r '.overall_coverage.tested_functions // 0' $COVERAGE_FILE)
        ENHANCED_FORMAT=true
    else
        # Legacy format
        OVERALL_COVERAGE=$(jq -r '.overall_coverage // 0' $COVERAGE_FILE)
        TOTAL_FILES=$(jq -r '.total_files // 0' $COVERAGE_FILE)
        TESTED_FILES=$(jq -r '.tested_files // 0' $COVERAGE_FILE)
        FUNCTION_COVERAGE=0
        FILE_COVERAGE=$OVERALL_COVERAGE
        TOTAL_FUNCTIONS=0
        TESTED_FUNCTIONS=0
        ENHANCED_FORMAT=false
    fi
    PACKAGE_COUNT=$(jq -r '.packages | length' $COVERAGE_FILE)
else
    echo "⚠️  jq not available, using basic validation"
    OVERALL_COVERAGE=0
    TOTAL_FILES=0
    TESTED_FILES=0
    FUNCTION_COVERAGE=0
    FILE_COVERAGE=0
    TOTAL_FUNCTIONS=0
    TESTED_FUNCTIONS=0
    PACKAGE_COUNT=0
    ENHANCED_FORMAT=false
fi

echo ""
echo "📈 Coverage Analysis:"
if [ "$ENHANCED_FORMAT" = true ]; then
    echo "   Function Coverage: ${FUNCTION_COVERAGE}% (Primary Metric)"
    echo "   File Coverage: ${FILE_COVERAGE}%"
    echo "   Total Functions: $TOTAL_FUNCTIONS"
    echo "   Tested Functions: $TESTED_FUNCTIONS"
else
    echo "   Overall Coverage: ${OVERALL_COVERAGE}% (File-based)"
fi
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
    if [ "$ENHANCED_FORMAT" = true ]; then
        # Show enhanced per-package coverage with function metrics
        jq -r '.packages[] | 
        if .metrics then
            "   \(.package): \(.metrics.function_coverage // 0)% functions (\(.metrics.tested_functions // 0)/\(.metrics.total_functions // 0)), \(.metrics.file_coverage // 0)% files (\(.metrics.tested_files // 0)/\(.metrics.total_files // 0))"
        else
            "   \(.package): \(.coverage // 0)% (\(.test_count // 0)/\(.file_count // 0) files)"
        end' $COVERAGE_FILE
    else
        # Show legacy per-package coverage
        jq -r '.packages[] | "   \(.package): \(.coverage)% (\(.test_count)/\(.file_count) files)"' $COVERAGE_FILE
    fi
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

EOF

    if [ "$ENHANCED_FORMAT" = true ]; then
        cat >> stdlib-coverage-report.md << EOF
**Function Coverage:** \`${FUNCTION_COVERAGE}%\` (Primary Metric)  
**File Coverage:** \`${FILE_COVERAGE}%\`  
**Target Coverage:** \`${STDLIB_COVERAGE_THRESHOLD}%\`  
**Status:** ${COVERAGE_ICON} ${COVERAGE_TEXT}

EOF
    else
        cat >> stdlib-coverage-report.md << EOF
**Current Coverage:** \`${OVERALL_COVERAGE}%\` (File-based)  
**Target Coverage:** \`${STDLIB_COVERAGE_THRESHOLD}%\`  
**Status:** ${COVERAGE_ICON} ${COVERAGE_TEXT}

EOF
    fi

    if command -v jq >/dev/null 2>&1 && [ "$PACKAGE_COUNT" -gt 0 ]; then
        echo "#### Package Coverage" >> stdlib-coverage-report.md
        echo "" >> stdlib-coverage-report.md
        if [ "$ENHANCED_FORMAT" = true ]; then
            jq -r '.packages[] | 
            if .metrics then
                "- **\(.package)**: \(.metrics.function_coverage // 0)% functions (\(.metrics.tested_functions // 0)/\(.metrics.total_functions // 0)), \(.metrics.file_coverage // 0)% files (\(.metrics.tested_files // 0)/\(.metrics.total_files // 0))"
            else
                "- **\(.package)**: \(.coverage // 0)% (\(.test_count // 0)/\(.file_count // 0) files)"
            end' $COVERAGE_FILE >> stdlib-coverage-report.md
        else
            jq -r '.packages[] | "- **\(.package)**: \(.coverage)% (\(.test_count)/\(.file_count) files)"' $COVERAGE_FILE >> stdlib-coverage-report.md
        fi
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
    if [ "$ENHANCED_FORMAT" = true ]; then
        echo "   Function coverage ${FUNCTION_COVERAGE}% meets threshold ${STDLIB_COVERAGE_THRESHOLD}%"
        echo "   Additional metrics: File coverage ${FILE_COVERAGE}%"
    else
        echo "   Coverage ${OVERALL_COVERAGE}% meets threshold ${STDLIB_COVERAGE_THRESHOLD}%"
    fi
else
    echo "❌ Stdlib coverage validation failed!"
    if [ "$ENHANCED_FORMAT" = true ]; then
        echo "   Function coverage ${FUNCTION_COVERAGE}% is below threshold ${STDLIB_COVERAGE_THRESHOLD}%"
        echo "   File coverage: ${FILE_COVERAGE}%"
        echo "💡 Add more tests to improve function coverage"
    else
        echo "   Coverage ${OVERALL_COVERAGE}% is below threshold ${STDLIB_COVERAGE_THRESHOLD}%"
        echo "💡 Add more tests to improve stdlib coverage"
    fi
fi

echo "📄 Updated coverage report: stdlib-coverage-report.md"

# Don't exit with error here - let the workflow handle the failure
# This allows the comment to be posted before failing the build
exit 0
