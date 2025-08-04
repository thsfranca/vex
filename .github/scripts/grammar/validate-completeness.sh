#!/bin/bash
set -e

echo "[CHECK] Checking grammar completeness..."

# Check for common language constructs that should be supported
GRAMMAR_FILE="tools/grammar/Vex.g4"

echo "Checking for required grammar rules..."

# Check if basic rules exist (account for multi-line definitions)
if ! grep -q "^program" "$GRAMMAR_FILE"; then
    echo "[ERROR] Missing root rule 'program'"
    exit 1
fi

if ! grep -q "^list" "$GRAMMAR_FILE"; then
    echo "[ERROR] Missing 'list' rule"
    exit 1
fi

if ! grep -q "SYMBOL" "$GRAMMAR_FILE"; then
    echo "[ERROR] Missing SYMBOL token"
    exit 1
fi

if ! grep -q "STRING" "$GRAMMAR_FILE"; then
    echo "[ERROR] Missing STRING token"
    exit 1
fi

echo "[SUCCESS] Grammar contains required basic rules"
echo "[SUCCESS] Grammar validation completed successfully"
