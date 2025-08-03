#!/bin/bash
set -e

# Arguments: EXTENSION_FILES VALIDATE_RESULT SKIP_RESULT
export EXTENSION_FILES="$1"
export VALIDATE_RESULT="$2" 
export SKIP_RESULT="$3"

cd tools/extension-tester
go build -o extension-tester .
./extension-tester summary >> $GITHUB_STEP_SUMMARY