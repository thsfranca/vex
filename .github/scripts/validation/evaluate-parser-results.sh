#!/bin/bash
set -e

# Script to evaluate parser generation test results
# Usage: evaluate-parser-results.sh <grammar_changed> <needs_testing> <test_result>

GRAMMAR_CHANGED="$1"
NEEDS_TESTING="$2"
TEST_RESULT="$3"

echo "📊 Evaluation Summary:"
echo "  - Grammar changed: $GRAMMAR_CHANGED"
echo "  - Tests needed: $NEEDS_TESTING"
echo "  - Tests result: $TEST_RESULT"

if [ "$NEEDS_TESTING" == "true" ]; then
  echo "🧪 Grammar tests were required and ran"
  if [ "$TEST_RESULT" == "success" ]; then
    echo "✅ All parser generation tests passed!"
    exit 0
  else
    echo "❌ Parser generation tests failed!"
    exit 1
  fi
else
  echo "✅ No grammar changes detected - parser generation tests skipped (OK)"
  exit 0
fi