#!/bin/bash
set -e

cd vscode-extension
# Build and use extension tester
cd ../tools/extension-tester
go build -o extension-tester .
cd ../../vscode-extension
../tools/extension-tester/extension-tester create-samples
