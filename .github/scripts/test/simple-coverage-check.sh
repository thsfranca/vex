#!/bin/bash
set -e

# Configuration
COVERAGE_THRESHOLD=85
COVERAGE_FILE="coverage.out"

echo "[COVERAGE] Running simplified coverage analysis..."

# Run tests with coverage, excluding generated parser files
echo "[DEBUG] Running coverage test command..."
go test -v -coverprofile="$COVERAGE_FILE" -covermode=atomic -coverpkg=./internal/transpiler ./... > coverage_test_output.log 2>&1
TEST_EXIT_CODE=$?
echo "[DEBUG] Test command completed with exit code: $TEST_EXIT_CODE"

# Check if tests actually passed by looking for "PASS" in output and coverage file exists
if grep -q "PASS" coverage_test_output.log && [ -f "$COVERAGE_FILE" ]; then
    echo "[COVERAGE] Tests passed successfully"
    cat coverage_test_output.log
else
    echo "❌ Tests failed during coverage analysis (exit code: $TEST_EXIT_CODE)"
    cat coverage_test_output.log
    exit 1
fi

# Clean up temporary log file
rm -f coverage_test_output.log

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

if [ "$COVERAGE_INT" -lt "$COVERAGE_THRESHOLD" ]; then
    STATUS_ICON="❌"
    STATUS_TEXT="Coverage below threshold"
fi

cat > coverage-report.md << EOF
## 📊 Transpiler Core Coverage Report

**Package:** \`internal/transpiler\` (Vex language implementation)  
**Current Coverage:** \`${TOTAL_COV}%\`  
**Target Coverage:** \`${COVERAGE_THRESHOLD}%\`  
**Status:** ${STATUS_ICON} ${STATUS_TEXT}

*Coverage focuses on transpiler core business logic, excluding CLI scaffolding and generated parser code.*
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