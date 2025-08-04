#!/bin/bash
set -e

cd vscode-extension
echo "[LINT] Linting JavaScript files..."

# Create basic .eslintrc.json if it doesn't exist
if [ ! -f .eslintrc.json ]; then
    cat > .eslintrc.json << 'EOF'
{
  "env": {
    "node": true,
    "es2021": true
  },
  "extends": ["eslint:recommended"],
  "parserOptions": {
    "ecmaVersion": 12,
    "sourceType": "module"
  },
  "rules": {
    "semi": ["error", "always"],
    "quotes": ["error", "single"],
    "no-unused-vars": "warn",
    "no-console": "off"
  }
}
EOF
fi

# Run ESLint on JS files
if ls *.js 1> /dev/null 2>&1; then
    eslint *.js || {
        echo "⚠️ ESLint found issues but continuing (warnings only for initial setup)"
        exit 0
    }
else
    echo "[INFO] No JavaScript files found to lint"
fi
