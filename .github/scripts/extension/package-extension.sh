#!/bin/bash
set -e

cd vscode-extension
../tools/extension-tester/extension-tester package
