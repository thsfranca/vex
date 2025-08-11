#!/bin/bash

set -e

echo "🧪 Running Vex Stdlib Tests"
echo "=============================="

# Configuration
STDLIB_DIR="stdlib"
STDLIB_COVERAGE_THRESHOLD=${STDLIB_COVERAGE_THRESHOLD:-98}
COVERAGE_FILE="stdlib-coverage.json"
TEST_LOG="stdlib-test-results.log"

# Ensure we're in the project root
if [ ! -d "$STDLIB_DIR" ]; then
    echo "❌ Error: stdlib directory not found. Are you in the project root?"
    exit 1
fi

# Build the Vex transpiler if not already built
if [ ! -f "bin/vex" ]; then
    echo "📦 Building Vex transpiler..."
    go build -o bin/vex ./cmd/vex-transpiler
fi

echo "🔍 Discovering stdlib test files..."

# Find all test files in stdlib
TEST_FILES=$(find $STDLIB_DIR -name "*_test.vx" -o -name "*_test.vex" 2>/dev/null || true)

if [ -z "$TEST_FILES" ]; then
    echo "⚠️  No stdlib test files found in $STDLIB_DIR"
    echo "   Expected files: *_test.vx or *_test.vex"
    
    # Create a minimal report for artifacts
    cat > stdlib-coverage-report.md << EOF
## 📚 Stdlib Test Report

**Status:** ⚠️ No Tests Found  
**Test Files:** 0  
**Coverage:** N/A  

No test files were found in the stdlib directory.
Expected test file patterns: \`*_test.vx\` or \`*_test.vex\`
EOF

    # Set output for GitHub Actions
    if [ -n "$GITHUB_OUTPUT" ]; then
        echo "status=no-tests" >> $GITHUB_OUTPUT
        echo "test-count=0" >> $GITHUB_OUTPUT
        echo "passed-count=0" >> $GITHUB_OUTPUT
        echo "failed-count=0" >> $GITHUB_OUTPUT
    fi
    
    exit 0
fi

echo "📋 Found test files:"
echo "$TEST_FILES" | while read -r file; do
    echo "   $file"
done

# Run tests and generate coverage
echo ""
echo "🚀 Running stdlib tests with coverage..."

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
FAILED_FILES=()

# Initialize test log
echo "Vex Stdlib Test Results" > $TEST_LOG
echo "========================" >> $TEST_LOG
echo "Timestamp: $(date -u +"%Y-%m-%d %H:%M:%S UTC")" >> $TEST_LOG
echo "" >> $TEST_LOG

# Run tests on each stdlib directory and generate coverage
while IFS= read -r test_file; do
    if [ -n "$test_file" ]; then
        TEST_DIR=$(dirname "$test_file")
        TEST_NAME=$(basename "$test_file")
        
        echo "🧪 Running: $test_file"
        echo "Running: $test_file" >> $TEST_LOG
        
        # Run the test with enhanced coverage output
        if timeout 30s ./bin/vex test -dir "$TEST_DIR" -enhanced-coverage -coverage-out "${TEST_DIR}-coverage.json" -verbose >> $TEST_LOG 2>&1; then
            echo "✅ PASSED: $test_file"
            echo "PASSED: $test_file" >> $TEST_LOG
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo "❌ FAILED: $test_file"
            echo "FAILED: $test_file" >> $TEST_LOG
            FAILED_TESTS=$((FAILED_TESTS + 1))
            FAILED_FILES+=("$test_file")
        fi
        
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        echo "" >> $TEST_LOG
    fi
done <<< "$TEST_FILES"

# Combine coverage reports
echo ""
echo "📊 Combining coverage reports..."

# Merge all individual coverage files into one
COMBINED_COVERAGE_DATA='{"timestamp":"'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'","overall_coverage":{"file_coverage":0,"function_coverage":0,"tested_files":0,"total_files":0,"tested_functions":0,"total_functions":0},"packages":[],"summary":{"precision_improvement":"Enhanced function-level coverage analysis","untested_functions":[],"top_coverage_packages":[],"needs_attention_packages":[]}}'

if command -v jq >/dev/null 2>&1; then
    echo "$COMBINED_COVERAGE_DATA" > $COVERAGE_FILE
    
    # Merge enhanced coverage data from all test runs
    TOTAL_FILES=0
    TESTED_FILES=0
    TOTAL_FUNCTIONS=0
    TESTED_FUNCTIONS=0
    
    for coverage_file in $(find $STDLIB_DIR -name "*-coverage.json" 2>/dev/null || true); do
        if [ -f "$coverage_file" ]; then
            # Check if this is enhanced coverage format
            if jq -e '.overall_coverage.function_coverage' "$coverage_file" >/dev/null 2>&1; then
                # Enhanced coverage format - merge package data and accumulate metrics
                jq --slurpfile new "$coverage_file" '.packages += $new[0].packages' $COVERAGE_FILE > temp_coverage.json && mv temp_coverage.json $COVERAGE_FILE
                
                # Accumulate totals
                TOTAL_FILES=$((TOTAL_FILES + $(jq -r '.overall_coverage.total_files // 0' "$coverage_file")))
                TESTED_FILES=$((TESTED_FILES + $(jq -r '.overall_coverage.tested_files // 0' "$coverage_file")))
                TOTAL_FUNCTIONS=$((TOTAL_FUNCTIONS + $(jq -r '.overall_coverage.total_functions // 0' "$coverage_file")))
                TESTED_FUNCTIONS=$((TESTED_FUNCTIONS + $(jq -r '.overall_coverage.tested_functions // 0' "$coverage_file")))
            else
                # Legacy format - convert to enhanced format
                jq --slurpfile new "$coverage_file" '.packages += $new[0].packages' $COVERAGE_FILE > temp_coverage.json && mv temp_coverage.json $COVERAGE_FILE
                TOTAL_FILES=$((TOTAL_FILES + $(jq -r '.total_files // 0' "$coverage_file")))
                TESTED_FILES=$((TESTED_FILES + $(jq -r '.tested_files // 0' "$coverage_file")))
            fi
        fi
    done
    
    # Calculate overall coverage metrics
    FILE_COVERAGE=0
    FUNCTION_COVERAGE=0
    if [ $TOTAL_FILES -gt 0 ]; then
        # Use awk for floating point arithmetic if bc is not available
        if command -v bc >/dev/null 2>&1; then
            FILE_COVERAGE=$(echo "scale=1; $TESTED_FILES * 100 / $TOTAL_FILES" | bc -l)
        else
            FILE_COVERAGE=$(awk "BEGIN {printf \"%.1f\", $TESTED_FILES * 100 / $TOTAL_FILES}")
        fi
    fi
    if [ $TOTAL_FUNCTIONS -gt 0 ]; then
        # Use awk for floating point arithmetic if bc is not available
        if command -v bc >/dev/null 2>&1; then
            FUNCTION_COVERAGE=$(echo "scale=1; $TESTED_FUNCTIONS * 100 / $TOTAL_FUNCTIONS" | bc -l)
        else
            FUNCTION_COVERAGE=$(awk "BEGIN {printf \"%.1f\", $TESTED_FUNCTIONS * 100 / $TOTAL_FUNCTIONS}")
        fi
    fi
    
    # Use function coverage as the primary metric, fall back to file coverage
    OVERALL_COVERAGE=${FUNCTION_COVERAGE:-$FILE_COVERAGE}
    OVERALL_COVERAGE=$(echo "$OVERALL_COVERAGE" | cut -d. -f1)  # Convert to integer
    
    # Update the combined coverage file with calculated metrics
    jq --argjson file_cov "$FILE_COVERAGE" \
       --argjson func_cov "$FUNCTION_COVERAGE" \
       --argjson total_files "$TOTAL_FILES" \
       --argjson tested_files "$TESTED_FILES" \
       --argjson total_funcs "$TOTAL_FUNCTIONS" \
       --argjson tested_funcs "$TESTED_FUNCTIONS" \
       '.overall_coverage = {
         "file_coverage": $file_cov,
         "function_coverage": $func_cov, 
         "tested_files": $tested_files,
         "total_files": $total_files,
         "tested_functions": $tested_funcs,
         "total_functions": $total_funcs
       }' $COVERAGE_FILE > temp_coverage.json && mv temp_coverage.json $COVERAGE_FILE
    
    echo "📈 Combined coverage - File: ${FILE_COVERAGE}%, Function: ${FUNCTION_COVERAGE}% (Primary: ${OVERALL_COVERAGE}%)"
else
    echo "⚠️  jq not available, using basic coverage calculation"
    OVERALL_COVERAGE=0
    echo "$COMBINED_COVERAGE_DATA" > $COVERAGE_FILE
fi

# Clean up individual coverage files
find $STDLIB_DIR -name "*-coverage.json" -delete 2>/dev/null || true

# Generate test results summary
echo ""
echo "📋 Test Summary:"
echo "   Total: $TOTAL_TESTS"
echo "   Passed: $PASSED_TESTS"
echo "   Failed: $FAILED_TESTS"

if [ $FAILED_TESTS -gt 0 ]; then
    echo ""
    echo "❌ Failed Tests:"
    for failed_file in "${FAILED_FILES[@]}"; do
        echo "   $failed_file"
    done
fi

# Create markdown report for PR comments
SUCCESS_RATE=0
if [ $TOTAL_TESTS -gt 0 ]; then
    SUCCESS_RATE=$(( (PASSED_TESTS * 100) / TOTAL_TESTS ))
fi

STATUS_ICON="✅"
STATUS_TEXT="All tests passed!"

if [ $FAILED_TESTS -gt 0 ]; then
    STATUS_ICON="❌"
    STATUS_TEXT="Some tests failed"
elif [ $TOTAL_TESTS -eq 0 ]; then
    STATUS_ICON="⚠️"
    STATUS_TEXT="No tests found"
fi

cat > stdlib-coverage-report.md << EOF
## 📚 Stdlib Test Report

**Status:** ${STATUS_ICON} ${STATUS_TEXT}  
**Test Files:** ${TOTAL_TESTS}  
**Passed:** ${PASSED_TESTS}  
**Failed:** ${FAILED_TESTS}  
**Success Rate:** ${SUCCESS_RATE}%  
**Coverage:** ${OVERALL_COVERAGE}%  
**Coverage Threshold:** ${STDLIB_COVERAGE_THRESHOLD}%  

EOF

if [ $FAILED_TESTS -gt 0 ]; then
    echo "### ❌ Failed Tests" >> stdlib-coverage-report.md
    echo "" >> stdlib-coverage-report.md
    for failed_file in "${FAILED_FILES[@]}"; do
        echo "- \`$failed_file\`" >> stdlib-coverage-report.md
    done
    echo "" >> stdlib-coverage-report.md
fi

echo "*Stdlib tests validate the standard library functionality and ensure API compatibility.*" >> stdlib-coverage-report.md

# Set output for GitHub Actions
if [ -n "$GITHUB_OUTPUT" ]; then
    echo "status=$([ $FAILED_TESTS -eq 0 ] && echo "passed" || echo "failed")" >> $GITHUB_OUTPUT
    echo "test-count=$TOTAL_TESTS" >> $GITHUB_OUTPUT
    echo "passed-count=$PASSED_TESTS" >> $GITHUB_OUTPUT
    echo "failed-count=$FAILED_TESTS" >> $GITHUB_OUTPUT
    echo "coverage=$OVERALL_COVERAGE" >> $GITHUB_OUTPUT
fi

echo ""
echo "🏁 Stdlib tests completed!"
echo "📄 Coverage report saved to: $COVERAGE_FILE"
echo "📋 Test log saved to: $TEST_LOG"

# Note: Don't exit with error here for now - allow CI to report results
# The workflow will check the status and decide whether to fail the build
# This allows comprehensive reporting even when tests fail

echo ""
if [ $FAILED_TESTS -gt 0 ]; then
    echo "⚠️  Some stdlib tests failed - results will be reported to PR"
else
    echo "✅ All stdlib tests passed!"
fi
