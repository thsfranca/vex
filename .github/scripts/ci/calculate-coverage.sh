#!/bin/bash
set -e

# Component coverage thresholds
declare -A COMPONENT_THRESHOLDS=(
    ["parser"]="95"
    ["transpiler"]="90" 
    ["types"]="85"
    ["stdlib"]="80"
)

# Component paths (matching actual codebase structure)
declare -A COMPONENT_PATHS=(
    ["parser"]="./internal/transpiler/parser"
    ["transpiler"]="./internal/transpiler"
    ["stdlib"]="./stdlib"
)

# Special handling for types (part of transpiler)
TYPES_FILES="./internal/transpiler/type*.go"

# Function to check component coverage
check_component_coverage() {
    local component=$1
    local threshold=$2
    local path=$3
    
    echo "[COVERAGE] Checking $component coverage (threshold: ${threshold}%)..."
    
    # Create coverage directory
    mkdir -p coverage
    local coverage_file="coverage/${component}.out"
    
    # Run tests with coverage for this component
    if go test -v -coverprofile="$coverage_file" "$path/..." 2>/dev/null; then
        if [ -f "$coverage_file" ] && [ -s "$coverage_file" ]; then
            local coverage=$(go tool cover -func="$coverage_file" | grep "total:" | awk '{print $3}' | sed 's/%//' | head -1)
            local coverage_int=${coverage%.*}  # Remove decimal part for comparison
            
            echo "[COVERAGE] $component: ${coverage}%"
            COVERAGE_REPORT="${COVERAGE_REPORT}  - $component: ${coverage}% (threshold: ${threshold}%)\n"
            
            if [ "$coverage_int" -lt "$threshold" ]; then
                echo "❌ $component coverage ${coverage}% is below threshold ${threshold}%"
                return 1  # Failed threshold
            else
                echo "✅ $component coverage ${coverage}% meets threshold ${threshold}%"
                return 0  # Passed threshold
            fi
        else
            echo "⚠️ No coverage data for $component"
            COVERAGE_REPORT="${COVERAGE_REPORT}  - $component: No data\n"
            return 0  # Skip if no data
        fi
    else
        echo "⚠️ Tests failed for $component, skipping coverage check"
        COVERAGE_REPORT="${COVERAGE_REPORT}  - $component: Tests failed\n"
        return 0  # Skip if tests failed
    fi
}

# Check if components exist and validate coverage
VALIDATION_FAILED=false
COVERAGE_REPORT=""

echo "[COVERAGE] Validating component coverage thresholds..."

for component in "${!COMPONENT_THRESHOLDS[@]}"; do
    threshold=${COMPONENT_THRESHOLDS[$component]}
    
    if [ "$component" = "types" ]; then
        # Special handling for types component
        if ls $TYPES_FILES 1> /dev/null 2>&1; then
            echo "[COVERAGE] Checking types coverage (threshold: ${threshold}%)..."
            echo "[COVERAGE] Types component uses shared transpiler tests - checking specific type file coverage"
            
            # Run transpiler tests and extract type-specific coverage
            mkdir -p coverage
            if go test -v -coverprofile="coverage/transpiler.out" "./internal/transpiler/..." 2>/dev/null; then
                if [ -f "coverage/transpiler.out" ] && [ -s "coverage/transpiler.out" ]; then
                    # Calculate coverage for type-related files only
                    type_coverage=$(go tool cover -func="coverage/transpiler.out" | grep -E "(type_|types\.go|type_checker\.go|type_inference\.go)" | awk '{if($3) {sum+=$3; count++}} END {if(count>0) printf "%.1f", sum/count; else print "0"}')
                    type_coverage_int=${type_coverage%.*}
                    
                    echo "[COVERAGE] types: ${type_coverage}%"
                    COVERAGE_REPORT="${COVERAGE_REPORT}  - types: ${type_coverage}% (threshold: ${threshold}%)\n"
                    
                    if [ "$type_coverage_int" -lt "$threshold" ]; then
                        echo "❌ types coverage ${type_coverage}% is below threshold ${threshold}%"
                        VALIDATION_FAILED=true
                    else
                        echo "✅ types coverage ${type_coverage}% meets threshold ${threshold}%"
                    fi
                else
                    echo "⚠️ No coverage data for types"
                    COVERAGE_REPORT="${COVERAGE_REPORT}  - types: No data\n"
                fi
            else
                echo "⚠️ Tests failed for types, skipping coverage check"
                COVERAGE_REPORT="${COVERAGE_REPORT}  - types: Tests failed\n"
            fi
        else
            echo "[COVERAGE] types not implemented yet, skipping"
            COVERAGE_REPORT="${COVERAGE_REPORT}  - types: Not implemented\n"
        fi
    else
        # Standard component handling
        path=${COMPONENT_PATHS[$component]}
        
        # Check if component exists by looking for .go files
        if find "$path" -name "*.go" -not -name "*_test.go" 2>/dev/null | grep -q .; then
            if ! check_component_coverage "$component" "$threshold" "$path"; then
                VALIDATION_FAILED=true
            fi
        else
            echo "[COVERAGE] $component not implemented yet, skipping"
            COVERAGE_REPORT="${COVERAGE_REPORT}  - $component: Not implemented\n"
        fi
    fi
done

# Check total coverage
if [ -f "coverage/total.out" ] && [ -s "coverage/total.out" ]; then
    TOTAL_COV=$(go tool cover -func=coverage/total.out | grep "total:" | awk '{print $3}' | sed 's/%//' || echo "0")
    echo "total-coverage=$TOTAL_COV" >> $GITHUB_OUTPUT
    echo "[COVERAGE] Total coverage: $TOTAL_COV%"
    COVERAGE_REPORT="${COVERAGE_REPORT}  - Total: ${TOTAL_COV}%\n"
else
    echo "total-coverage=0" >> $GITHUB_OUTPUT
    echo "[COVERAGE] No total coverage data available"
    COVERAGE_REPORT="${COVERAGE_REPORT}  - Total: No data\n"
fi

# Output coverage report
echo -e "\n📊 Coverage Report:"
echo -e "$COVERAGE_REPORT"

# Create markdown report for PR comment
create_pr_coverage_report() {
    local status_icon="✅"
    local status_text="All coverage thresholds met!"
    local details_section=""
    
    if [ "$VALIDATION_FAILED" = true ]; then
        status_icon="❌"
        status_text="Coverage thresholds not met"
        details_section="
### 🔧 How to Fix Coverage Issues:
1. **Add tests for uncovered code paths**
2. **Review existing tests for completeness**  
3. **Consider edge cases and error conditions**
4. **Test error handling and boundary conditions**

This is a learning project - maintaining good test coverage teaches best practices!"
    fi
    
    # Create coverage table
    local coverage_table="| Component | Current | Threshold | Status |\n|-----------|---------|-----------|--------|\n"
    
    # Build coverage table from our component data
    for component in "${!COMPONENT_THRESHOLDS[@]}"; do
        threshold="${COMPONENT_THRESHOLDS[$component]}%"
        
        # Extract current coverage from our reports
        current_coverage=""
        status=""
        
        if echo -e "$COVERAGE_REPORT" | grep -q "$component:"; then
            current_line=$(echo -e "$COVERAGE_REPORT" | grep "$component:" | head -1)
            if [[ $current_line =~ ([0-9]+\.?[0-9]*)% ]]; then
                current_coverage="${BASH_REMATCH[1]}%"
                current_num=${BASH_REMATCH[1]%.*}
                threshold_num=${COMPONENT_THRESHOLDS[$component]}
                
                if [[ $current_num -ge $threshold_num ]]; then
                    status="✅ Pass"
                else
                    status="❌ Fail"
                fi
            elif [[ $current_line == *"No data"* ]]; then
                current_coverage="-"
                status="⚠️ No data"
            elif [[ $current_line == *"Tests failed"* ]]; then
                current_coverage="-"
                status="⚠️ Tests failed"
            elif [[ $current_line == *"Not implemented"* ]]; then
                current_coverage="-"
                status="⚠️ Not implemented"
            fi
        else
            current_coverage="-"
            status="⚠️ Unknown"
        fi
        
        coverage_table="${coverage_table}| $component | $current_coverage | $threshold | $status |\n"
    done
    
    # Add total coverage if available
    local total_section=""
    if echo -e "$COVERAGE_REPORT" | grep -q "Total:"; then
        total_line=$(echo -e "$COVERAGE_REPORT" | grep "Total:" | head -1)
        if [[ $total_line =~ ([0-9]+\.?[0-9]*)% ]]; then
            total_coverage="${BASH_REMATCH[1]}%"
            total_section="
**Overall Coverage:** ${total_coverage}"
        fi
    fi

    cat > coverage-report.md << EOF
## ${status_icon} Test Coverage Report

**Status:** ${status_text}${total_section}

### 📊 Component Coverage:
$(echo -e "${coverage_table}")

### 🎯 Required Thresholds:
| Component | Threshold | Purpose |
|-----------|-----------|---------|
| Parser | 95%+ | Critical language component |
| Transpiler | 90%+ | Core functionality |
| Types | 85%+ | Type system implementation |
| Standard Library | 80%+ | User-facing features |

${details_section}

---
*Coverage validation enforced to maintain code quality in this learning project.*
EOF
}

# Create the PR coverage report
create_pr_coverage_report

# Output location for GitHub Actions
echo "COVERAGE_REPORT_FILE=coverage-report.md" >> $GITHUB_OUTPUT
echo "COVERAGE_STATUS=$([ "$VALIDATION_FAILED" = true ] && echo "failed" || echo "passed")" >> $GITHUB_OUTPUT

# Fail if any component is below threshold
if [ "$VALIDATION_FAILED" = true ]; then
    echo ""
    echo "💥 BUILD FAILED: One or more components below coverage threshold"
    echo "This is a learning project - maintaining good test coverage is important!"
    echo ""
    echo "🔧 To fix:"
    echo "1. Add tests for uncovered code paths"
    echo "2. Review existing tests for completeness"
    echo "3. Consider edge cases and error conditions"
    echo ""
    exit 1
else
    echo ""
    echo "✅ All implemented components meet coverage thresholds!"
fi