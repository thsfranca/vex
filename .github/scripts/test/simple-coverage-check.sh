#!/bin/bash
set -e

# Configuration
COVERAGE_THRESHOLD=80
COVERAGE_FILE="coverage.out"

echo "[COVERAGE] Running simplified coverage analysis..."

# Run tests with coverage
go test -v -race -coverprofile="$COVERAGE_FILE" -covermode=atomic ./... || {
    echo "❌ Tests failed during coverage analysis"
    exit 1
}

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ No coverage data generated"
    exit 1
fi

# Calculate overall coverage
TOTAL_COV=$(go tool cover -func="$COVERAGE_FILE" | grep "total:" | awk '{print $3}' | sed 's/%//' || echo "0.0")
COVERAGE_INT=${TOTAL_COV%.*}

echo "[COVERAGE] Total coverage: $TOTAL_COV%"

# Generate HTML report for artifacts
go tool cover -html="$COVERAGE_FILE" -o coverage.html

# Create informative PR report
STATUS_ICON="✅"
STATUS_TEXT="Coverage threshold met!"
DETAILS_SECTION=""

if [ "$COVERAGE_INT" -lt "$COVERAGE_THRESHOLD" ]; then
    STATUS_ICON="❌"
    STATUS_TEXT="Coverage below threshold"
    DETAILS_SECTION="
### 🔧 How to Improve Coverage:
1. **Run coverage locally:** \`go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out\`
2. **Add tests for uncovered code paths**
3. **Focus on critical functionality first**
4. **Consider edge cases and error conditions**

This helps maintain code quality and catch potential bugs early!"
fi

cat > coverage-report.md << EOF
## ${STATUS_ICON} Test Coverage Report

**Current Coverage:** \`${TOTAL_COV}%\`  
**Required Threshold:** \`${COVERAGE_THRESHOLD}%\`  
**Status:** ${STATUS_TEXT}

### 📊 Coverage Summary:
| Metric | Value | Status |
|--------|-------|--------|
| Overall Coverage | ${TOTAL_COV}% | $([ "$COVERAGE_INT" -ge "$COVERAGE_THRESHOLD" ] && echo "✅ Pass" || echo "❌ Fail") |
| Required Threshold | ${COVERAGE_THRESHOLD}% | - |
| Difference | $([ "$COVERAGE_INT" -ge "$COVERAGE_THRESHOLD" ] && echo "+$((COVERAGE_INT - COVERAGE_THRESHOLD))%" || echo "$((COVERAGE_INT - COVERAGE_THRESHOLD))%") | $([ "$COVERAGE_INT" -ge "$COVERAGE_THRESHOLD" ] && echo "Above target" || echo "Below target") |

### 🎯 What This Means:
$([ "$COVERAGE_INT" -ge "$COVERAGE_THRESHOLD" ] && echo "✅ **Great work!** Your changes maintain good test coverage." || echo "⚠️ **Action needed:** Please add tests to reach the ${COVERAGE_THRESHOLD}% coverage threshold.")

${DETAILS_SECTION}

### 🔍 Detailed Coverage Report:
Download the \`coverage.html\` artifact from this CI run to see line-by-line coverage details.

---
*Automated coverage validation • Target: ${COVERAGE_THRESHOLD}%+ for quality assurance*
EOF

# Set output for GitHub Actions
if [ -n "$GITHUB_OUTPUT" ]; then
    echo "coverage=$TOTAL_COV" >> $GITHUB_OUTPUT
    echo "coverage-status=$([ "$COVERAGE_INT" -ge "$COVERAGE_THRESHOLD" ] && echo "passed" || echo "failed")" >> $GITHUB_OUTPUT
fi

# Check threshold - but don't fail CI job here, let commenting happen first
if [ "$COVERAGE_INT" -lt "$COVERAGE_THRESHOLD" ]; then
    echo ""
    echo "❌ Coverage $TOTAL_COV% is below threshold $COVERAGE_THRESHOLD%"
    echo "💡 Add tests to improve coverage before merging"
    echo "⚠️ Coverage check will be handled by workflow after commenting"
else
    echo ""
    echo "✅ Coverage $TOTAL_COV% meets threshold $COVERAGE_THRESHOLD%"
fi

# Exit successfully to allow comment posting - CI failure handled by workflow