#!/bin/bash
set -e

echo "[BUILD] Setting up grammar validator..."

# Create parser subdirectory and copy generated files there
mkdir -p tools/grammar-validator/parser
cp /tmp/vex-parser/*.go tools/grammar-validator/parser/

# Build the grammar validator tool
cd tools/grammar-validator
go mod download
go build -o grammar-validator .

# Test valid example files (must pass)
echo "[TEST] Testing valid example files..."
if [ -d "$GITHUB_WORKSPACE/examples/valid" ]; then
    VALID_FILES=$(find $GITHUB_WORKSPACE/examples/valid -name "*.vx")
    if [ -n "$VALID_FILES" ]; then
        ./grammar-validator --valid $VALID_FILES
    else
        echo "⚠️ No valid .vx files found"
    fi
fi

# Test invalid example files (must fail)
echo "[TEST] Testing invalid example files..."
if [ -d "$GITHUB_WORKSPACE/examples/invalid" ]; then
    INVALID_FILES=$(find $GITHUB_WORKSPACE/examples/invalid -name "*.vx")
    if [ -n "$INVALID_FILES" ]; then
        ./grammar-validator --invalid $INVALID_FILES
    else
        echo "⚠️ No invalid .vx files found"
    fi
fi
