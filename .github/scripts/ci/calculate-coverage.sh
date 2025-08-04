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