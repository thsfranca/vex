#!/bin/bash
set -e

if [ -f "coverage/total.out" ] && [ -s "coverage/total.out" ]; then
    TOTAL_COV=$(go tool cover -func=coverage/total.out | grep "total:" | awk '{print $3}' | sed 's/%//' || echo "0")
    echo "total-coverage=$TOTAL_COV" >> $GITHUB_OUTPUT
    echo "[COVERAGE] Coverage: $TOTAL_COV%"
    
    # Coverage requirements based on project maturity
    GO_FILE_COUNT=$(find . -name "*.go" -not -name "*_test.go" | wc -l)
    if [ $GO_FILE_COUNT -lt 5 ]; then
        echo "⚠️ Early development - no coverage requirements yet"
    elif [ $GO_FILE_COUNT -lt 20 ]; then
        echo "📈 Growing project - aim for 40%+ coverage"
    else
        echo "🎯 Mature project - aim for 70%+ coverage"
    fi
else
    echo "total-coverage=0" >> $GITHUB_OUTPUT
    echo "[COVERAGE] No coverage data available"
fi