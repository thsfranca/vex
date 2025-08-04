#!/bin/bash
set -e

# Script to evaluate VSCode extension validation results
# Usage: evaluate-extension-results.sh <extension_files_changed> <validation_result> <skip_result>

EXTENSION_FILES_CHANGED="$1"
VALIDATION_RESULT="$2"
SKIP_RESULT="$3"

echo "📊 VSCode Extension Validation Summary:"
echo "  - Extension files changed: $EXTENSION_FILES_CHANGED"
echo "  - Validation result: $VALIDATION_RESULT"
echo "  - Skip result: $SKIP_RESULT"

if [ "$EXTENSION_FILES_CHANGED" == "true" ]; then
  echo "🧪 Extension validation was required and ran"
  if [ "$VALIDATION_RESULT" == "success" ]; then
    echo "✅ VSCode extension validation passed!"
    exit 0
  else
    echo "❌ VSCode extension validation failed!"
    exit 1
  fi
else
  echo "✅ No extension changes detected - validation skipped (OK)"
  exit 0
fi
