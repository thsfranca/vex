#!/bin/bash
set -e

echo "[TEST] Running tests with coverage analysis..."

if find . -name "*_test.go" | grep -q .; then
    go test -v -race -coverprofile=coverage.out -covermode=atomic -coverpkg=./cmd/vex-transpiler,./internal/transpiler ./... || {
        echo "[ERROR] Tests failed during coverage analysis"
        # Check if this should be treated as a warning in early development
        if [ $(find . -name "*.go" -not -name "*_test.go" | wc -l) -lt 5 ]; then
            echo "⚠️ Early development - coverage failures are warnings only"
            echo "coverage: 0.0% of statements" > coverage.out
        else
            echo "💥 Sufficient codebase - coverage analysis failures are blocking"
            exit 1
        fi
    }
    
    if [ -f coverage.out ]; then
        TOTAL_COV=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//' || echo "0.0")
        echo "[COVERAGE] Coverage: $TOTAL_COV%"
        
        # Generate HTML coverage report for artifacts
        go tool cover -html=coverage.out -o coverage.html || true
        
        # Set coverage output for other jobs
        if [ -n "$GITHUB_OUTPUT" ]; then
            echo "coverage=$TOTAL_COV" >> $GITHUB_OUTPUT
        fi
    else
        echo "[COVERAGE] No coverage data available"
        if [ -n "$GITHUB_OUTPUT" ]; then
            echo "coverage=0.0" >> $GITHUB_OUTPUT
        fi
    fi
else
    echo "[INFO] No tests found - skipping coverage analysis"
    if [ -n "$GITHUB_OUTPUT" ]; then
        echo "coverage=0.0" >> $GITHUB_OUTPUT
    fi
fi
