#!/bin/bash
set -e

# Arguments: BASIC_GO MAKEFILE_GO
BASIC_GO="$1"
MAKEFILE_GO="$2"

cd tools/change-detector
if [ ! -f change-detector ]; then
    go build -o change-detector .
fi

output=$(./change-detector final-decision "$BASIC_GO" "$MAKEFILE_GO")
echo "$output"
GO_FILES=$(echo "$output" | grep "go-files=" | cut -d'=' -f2)
echo "go-files=$GO_FILES" >> $GITHUB_OUTPUT
