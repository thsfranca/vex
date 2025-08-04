#!/bin/bash
set -e

# Script to decide if validation is needed based on grammar/example changes
# Usage: decide-validation-needed.sh <grammar_changed>

GRAMMAR_CHANGED="$1"

if [ "$GRAMMAR_CHANGED" == "true" ]; then
  echo "needs-validation=true" >> $GITHUB_OUTPUT
  echo "🧪 Validation will run: Grammar or examples changed"
else
  echo "needs-validation=false" >> $GITHUB_OUTPUT
  echo "⏭️ Validation will be skipped: No relevant changes"
fi
