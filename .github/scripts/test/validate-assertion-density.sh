#!/bin/bash

set -e

echo "🎯 Validating Test Assertion Density"
echo "===================================="

# Configuration
ASSERTION_DENSITY_THRESHOLD=${ASSERTION_DENSITY_THRESHOLD:-1.5}
COVERAGE_FILE=${COVERAGE_FILE:-"stdlib-coverage.json"}
TEST_QUALITY_THRESHOLD=${TEST_QUALITY_THRESHOLD:-60}

# Check if coverage file exists
if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Error: Coverage file not found: $COVERAGE_FILE"
    echo "Make sure to run tests with enhanced coverage first to generate quality data."
    exit 1
fi

echo "📄 Reading test quality data from: $COVERAGE_FILE"

# Extract test quality data
if command -v jq >/dev/null 2>&1; then
    # Check if this is enhanced coverage format with quality metrics
    if jq -e '.quality_metrics' "$COVERAGE_FILE" >/dev/null 2>&1; then
        ENHANCED_QUALITY=true
        
        # Extract overall quality metrics
        OVERALL_ASSERTION_DENSITY=$(jq -r '.quality_metrics.overall.assertion_density // 0' "$COVERAGE_FILE")
        OVERALL_QUALITY_SCORE=$(jq -r '.quality_metrics.overall.overall_quality_score // 0' "$COVERAGE_FILE")
        PACKAGE_COUNT=$(jq -r '.quality_metrics | keys | length' "$COVERAGE_FILE")
        
        echo "📊 Overall Test Quality Analysis:"
        echo "   Assertion Density: ${OVERALL_ASSERTION_DENSITY}/test"
        echo "   Quality Score: ${OVERALL_QUALITY_SCORE}/100"
        echo "   Packages Analyzed: $PACKAGE_COUNT"
        echo "   Density Threshold: ${ASSERTION_DENSITY_THRESHOLD}/test"
        echo "   Quality Threshold: ${TEST_QUALITY_THRESHOLD}/100"
    else
        echo "⚠️  Enhanced coverage format with quality metrics not found"
        echo "   Run tests with: vex test -enhanced-coverage -coverage-out $COVERAGE_FILE"
        ENHANCED_QUALITY=false
        OVERALL_ASSERTION_DENSITY=0
        OVERALL_QUALITY_SCORE=0
        PACKAGE_COUNT=0
    fi
else
    echo "❌ Error: jq not available for JSON parsing"
    exit 1
fi

# Validate assertion density threshold
DENSITY_STATUS="passed"
QUALITY_STATUS="passed"

# Check assertion density
if [ "$ENHANCED_QUALITY" = true ]; then
    # Use awk for floating point comparison
    DENSITY_CHECK=$(awk -v density="$OVERALL_ASSERTION_DENSITY" -v threshold="$ASSERTION_DENSITY_THRESHOLD" 'BEGIN {print (density >= threshold) ? "pass" : "fail"}')
    if [ "$DENSITY_CHECK" = "fail" ]; then
        DENSITY_STATUS="failed"
    fi
    
    # Check overall quality score
    QUALITY_CHECK=$(awk -v score="$OVERALL_QUALITY_SCORE" -v threshold="$TEST_QUALITY_THRESHOLD" 'BEGIN {print (score >= threshold) ? "pass" : "fail"}')
    if [ "$QUALITY_CHECK" = "fail" ]; then
        QUALITY_STATUS="failed"
    fi
fi

# Generate detailed analysis per package
echo ""
echo "📋 Package-Level Assertion Density Analysis:"

FAILED_PACKAGES=()
LOW_DENSITY_PACKAGES=()

if [ "$ENHANCED_QUALITY" = true ] && [ "$PACKAGE_COUNT" -gt 0 ]; then
    jq -r '.quality_metrics | to_entries[] | "\(.key):\(.value.assertion_density // 0):\(.value.overall_quality_score // 0)"' "$COVERAGE_FILE" | while IFS=: read -r package density quality; do
        # Format density for display
        FORMATTED_DENSITY=$(awk -v d="$density" 'BEGIN {printf "%.1f", d}')
        FORMATTED_QUALITY=$(awk -v q="$quality" 'BEGIN {printf "%.1f", q}')
        
        # Check if package meets thresholds
        PACKAGE_DENSITY_CHECK=$(awk -v density="$density" -v threshold="$ASSERTION_DENSITY_THRESHOLD" 'BEGIN {print (density >= threshold) ? "✅" : "❌"}')
        PACKAGE_QUALITY_CHECK=$(awk -v score="$quality" -v threshold="$TEST_QUALITY_THRESHOLD" 'BEGIN {print (score >= threshold) ? "✅" : "⚠️ "}')
        
        echo "   $PACKAGE_DENSITY_CHECK $package: ${FORMATTED_DENSITY}/test (Quality: ${FORMATTED_QUALITY}/100 $PACKAGE_QUALITY_CHECK)"
        
        # Track failed packages for reporting
        if [ "$PACKAGE_DENSITY_CHECK" = "❌" ]; then
            echo "$package" >> /tmp/failed_density_packages.txt
        fi
        
        DENSITY_BELOW_1=$(awk -v density="$density" 'BEGIN {print (density < 1.0) ? "yes" : "no"}')
        if [ "$DENSITY_BELOW_1" = "yes" ]; then
            echo "$package" >> /tmp/low_density_packages.txt
        fi
    done
    
    # Read failed packages (if any)
    if [ -f /tmp/failed_density_packages.txt ]; then
        FAILED_PACKAGES=($(cat /tmp/failed_density_packages.txt))
        rm -f /tmp/failed_density_packages.txt
    fi
    
    if [ -f /tmp/low_density_packages.txt ]; then
        LOW_DENSITY_PACKAGES=($(cat /tmp/low_density_packages.txt))
        rm -f /tmp/low_density_packages.txt
    fi
else
    echo "   No quality metrics available - run enhanced coverage analysis"
fi

# Generate status icons and messages
DENSITY_ICON="✅"
DENSITY_TEXT="Assertion density meets threshold"
QUALITY_ICON="✅"
QUALITY_TEXT="Test quality meets threshold"

if [ "$DENSITY_STATUS" = "failed" ]; then
    DENSITY_ICON="❌"
    DENSITY_TEXT="Assertion density below threshold"
fi

if [ "$QUALITY_STATUS" = "failed" ]; then
    QUALITY_ICON="⚠️"
    QUALITY_TEXT="Test quality below threshold"
fi

# Generate markdown section for inclusion in main report
cat > assertion-density-section.md << EOF
## 🎯 Test Assertion Density Report

**Assertion Density:** \`${OVERALL_ASSERTION_DENSITY}/test\` ${DENSITY_ICON}  
**Quality Score:** \`${OVERALL_QUALITY_SCORE}/100\` ${QUALITY_ICON}  
**Density Threshold:** \`${ASSERTION_DENSITY_THRESHOLD}/test\`  
**Quality Threshold:** \`${TEST_QUALITY_THRESHOLD}/100\`  

### 📊 Status
- **Density Status:** ${DENSITY_ICON} ${DENSITY_TEXT}
- **Quality Status:** ${QUALITY_ICON} ${QUALITY_TEXT}

EOF

# Add package details if available
if [ "$ENHANCED_QUALITY" = true ] && [ "$PACKAGE_COUNT" -gt 0 ]; then
    echo "### 📋 Package Analysis" >> assertion-density-report.md
    echo "" >> assertion-density-report.md
    
    jq -r '.quality_metrics | to_entries[] | "\(.key):\(.value.assertion_density // 0):\(.value.overall_quality_score // 0)"' "$COVERAGE_FILE" | while IFS=: read -r package density quality; do
        FORMATTED_DENSITY=$(awk -v d="$density" 'BEGIN {printf "%.1f", d}')
        FORMATTED_QUALITY=$(awk -v q="$quality" 'BEGIN {printf "%.1f", q}')
        PACKAGE_DENSITY_CHECK=$(awk -v density="$density" -v threshold="$ASSERTION_DENSITY_THRESHOLD" 'BEGIN {print (density >= threshold) ? "✅" : "❌"}')
        
        echo "- **$package**: ${FORMATTED_DENSITY}/test, Quality: ${FORMATTED_QUALITY}/100 $PACKAGE_DENSITY_CHECK" >> assertion-density-section.md
    done
    echo "" >> assertion-density-section.md
fi

# Add recommendations if there are issues
if [ "$DENSITY_STATUS" = "failed" ] || [ "$QUALITY_STATUS" = "failed" ]; then
    cat >> assertion-density-section.md << EOF
### 💡 Recommendations

EOF
    
    if [ "$DENSITY_STATUS" = "failed" ]; then
        cat >> assertion-density-section.md << EOF
**Improve Assertion Density:**
- Add more \`assert-eq\` statements to verify function behavior
- Test edge cases with additional assertions
- Ensure each test verifies multiple aspects of functionality

EOF
    fi
    
    if [ "${#LOW_DENSITY_PACKAGES[@]}" -gt 0 ]; then
        cat >> assertion-density-section.md << EOF
**Critical Issues (< 1.0 assertions/test):**
EOF
        for package in "${LOW_DENSITY_PACKAGES[@]}"; do
            echo "- \`$package\`: Tests may not be verifying behavior properly" >> assertion-density-section.md
        done
        echo "" >> assertion-density-section.md
    fi
    
    cat >> assertion-density-section.md << EOF
**General Quality Improvements:**
- Use descriptive test names that explain the scenario
- Test boundary values and error conditions
- Use variety of assertion types (\`assert-eq\`, \`assert-true\`, \`assert-false\`)

EOF
fi

echo "" >> assertion-density-section.md

# Set output for GitHub Actions
if [ -n "$GITHUB_OUTPUT" ]; then
    echo "density-status=$DENSITY_STATUS" >> "$GITHUB_OUTPUT"
    echo "quality-status=$QUALITY_STATUS" >> "$GITHUB_OUTPUT"
    echo "assertion-density=$OVERALL_ASSERTION_DENSITY" >> "$GITHUB_OUTPUT"
    echo "quality-score=$OVERALL_QUALITY_SCORE" >> "$GITHUB_OUTPUT"
    echo "failed-packages=${#FAILED_PACKAGES[@]}" >> "$GITHUB_OUTPUT"
fi

# Final status report
echo ""
echo "🎯 Assertion Density Validation Results:"
echo "   Density Status: ${DENSITY_ICON} ${DENSITY_TEXT}"
echo "   Quality Status: ${QUALITY_ICON} ${QUALITY_TEXT}"

if [ "$ENHANCED_QUALITY" = true ]; then
    echo "   Overall Density: ${OVERALL_ASSERTION_DENSITY}/test (threshold: ${ASSERTION_DENSITY_THRESHOLD})"
    echo "   Overall Quality: ${OVERALL_QUALITY_SCORE}/100 (threshold: ${TEST_QUALITY_THRESHOLD})"
else
    echo "   ⚠️  Run enhanced coverage to get detailed metrics"
fi

if [ "${#FAILED_PACKAGES[@]}" -gt 0 ]; then
    echo ""
    echo "❌ Packages below assertion density threshold:"
    for package in "${FAILED_PACKAGES[@]}"; do
        echo "   $package"
    done
fi

if [ "${#LOW_DENSITY_PACKAGES[@]}" -gt 0 ]; then
    echo ""
    echo "🚨 Critical: Packages with very low assertion density (< 1.0):"
    for package in "${LOW_DENSITY_PACKAGES[@]}"; do
        echo "   $package"
    done
fi

echo ""
echo "📄 Assertion density section saved to: assertion-density-section.md"

# Don't exit with error here - let the workflow handle the failure
# This allows the comment to be posted before failing the build
exit 0
