#!/bin/bash
set -e

cd vscode-extension
echo "[FORMAT] Checking code formatting..."

# Create .prettierrc if it doesn't exist
if [ ! -f .prettierrc ]; then
    cat > .prettierrc << 'EOF'
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 100,
  "tabWidth": 2
}
EOF
fi

# Check formatting
prettier --check *.js *.json || {
    echo "⚠️ Formatting issues found - showing diff:"
    prettier --list-different *.js *.json || true
    echo "[INFO] Run 'prettier --write *.js *.json' to fix formatting"
    exit 0
}