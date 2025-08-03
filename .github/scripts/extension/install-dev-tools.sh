#!/bin/bash
set -e

npm install -g eslint prettier
# Install specific vsce version that's compatible with Node 20
npm install -g @vscode/vsce@latest