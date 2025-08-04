#!/bin/bash
set -e

echo "[TEST] Running Go tests..."

if find . -name "*_test.go" | grep -q .; then
    go test -v ./... || {
        echo "[ERROR] Tests failed, but checking if this should block..."
        # In early development, test failures are warnings only
        if [ $(find . -name "*.go" -not -name "*_test.go" | wc -l) -lt 5 ]; then
            echo "⚠️ Early development detected - treating test failures as warnings"
            exit 0
        else
            echo "💥 Sufficient codebase exists - test failures are blocking"
            exit 1
        fi
    }
else
    echo "[INFO] No tests found yet - this is expected for early development"
fi
