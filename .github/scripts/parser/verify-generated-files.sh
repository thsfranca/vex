#!/bin/bash
set -e

echo "[CHECK] Verifying generated Go parser files..."

if [ -f "tools/gen/go/vex_lexer.go" ] && [ -f "tools/gen/go/vex_parser.go" ]; then
    echo "[SUCCESS] Go parser files generated successfully"
else
    echo "[ERROR] Missing Go parser files"
    ls -la tools/gen/go/ || echo "Directory doesn't exist"
    exit 1
fi