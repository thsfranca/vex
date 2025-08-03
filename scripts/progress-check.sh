#!/bin/bash

# Fugo Progress Check Script
# Run this weekly to update progress and plan next steps

echo "[PROGRESS] Fugo Language Progress Check"
echo "================================"
echo

# Get current date
CURRENT_DATE=$(date "+%Y-%m-%d")
echo "📅 Date: $CURRENT_DATE"
echo

# Check git status
echo "[STATUS] Git Status:"
git log --oneline -5
echo

# Check if parsers are working
echo "🔧 Parser Status:"
if [ -f "tools/gen/go/fugo_parser.go" ]; then
    echo "[SUCCESS] Go parser: Available"
else
    echo "[ERROR] Go parser: Missing"
fi
echo

# Show current phase based on what exists
echo "🎯 Current Phase:"
if [ -f "internal/evaluator/evaluator.go" ]; then
    echo "🚧 Phase 2: Tree-Walking Interpreter"
elif [ -f "internal/types/types.go" ]; then
    echo "🚧 Phase 3: Type System"
elif [ -f "internal/transpiler/transpiler.go" ]; then
    echo "🚧 Phase 4: Go Transpilation"
else
    echo "[SUCCESS] Phase 1: Parser Foundation (Complete)"
    echo "🎯 Ready for Phase 2: Tree-Walking Interpreter"
fi
echo

# Suggest next steps
echo "📋 Suggested Next Steps:"
echo "1. Update PROGRESS.md with this week's accomplishments"
echo "2. Review current phase goals"
echo "3. Plan next week's tasks"
echo "4. Commit and push progress"
echo

# Count lines of code
echo "📈 Project Size:"
echo "Fugo files: $(find examples -name "*.fugo" | wc -l | tr -d ' ')"
echo "Go files: $(find . -name "*.go" -not -path "./tools/gen/*" | wc -l | tr -d ' ')"
echo "Documentation: $(find docs -name "*.md" | wc -l | tr -d ' ')"
echo

echo "[TIP] Tip: Run 'git commit -am \"Weekly progress update\"' to save your work!"