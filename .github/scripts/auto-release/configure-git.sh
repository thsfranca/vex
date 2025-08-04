#!/bin/bash
set -e

# Configure Git for automated releases
# Usage: configure-git.sh

echo "🔧 Configuring Git for auto-release..."

git config --global user.email "action@github.com"
git config --global user.name "GitHub Action"
git config --global push.default current

echo "Git configured:"
git config --list | grep -E "(user|push)"

echo "✅ Git configuration complete"