#!/bin/bash
set -e

cd tools/change-detector
go build -o change-detector .
output=$(./change-detector makefile-analysis)
echo "$output"
GO_RELATED=$(echo "$output" | grep "go-related=" | cut -d'=' -f2)
echo "go-related=$GO_RELATED" >> $GITHUB_OUTPUT