#!/bin/bash
set -e

# Script to decide if testing is needed based on changes and event type
# Usage: decide-testing-needed.sh <grammar_changed> <event_name>

GRAMMAR_CHANGED="$1"
EVENT_NAME="$2"

if [ "$GRAMMAR_CHANGED" == "true" ] || [ "$EVENT_NAME" == "schedule" ]; then
  echo "needs-testing=true" >> $GITHUB_OUTPUT
  echo "🧪 Tests will run: Grammar changes or scheduled run"
else
  echo "needs-testing=false" >> $GITHUB_OUTPUT
  echo "⏭️ Tests will be skipped: No relevant changes"
fi