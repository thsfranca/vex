#!/bin/bash
set -e

# Script to evaluate grammar validation results
# Usage: evaluate-grammar-results.sh <grammar_changed> <needs_validation> <validation_result>

GRAMMAR_CHANGED="$1"
NEEDS_VALIDATION="$2"
VALIDATION_RESULT="$3"

echo "📊 Grammar Validation Summary:"
echo "  - Grammar changed: $GRAMMAR_CHANGED"
echo "  - Validation needed: $NEEDS_VALIDATION"
echo "  - Validation result: $VALIDATION_RESULT"

if [ "$NEEDS_VALIDATION" == "true" ]; then
  echo "🧪 Grammar validation was required and ran"
  if [ "$VALIDATION_RESULT" == "success" ]; then
    echo "✅ Grammar validation passed!"
    exit 0
  else
    echo "❌ Grammar validation failed!"
    exit 1
  fi
else
  echo "✅ No grammar/example changes detected - validation skipped (OK)"
  exit 0
fi