# Grammar Validator

A Go tool for validating Vex grammar by parsing `.vx` example files using the generated ANTLR parser.

## Purpose

This tool is used in CI to ensure that:
1. The ANTLR grammar generates valid Go parser code
2. All example `.vx` files can be successfully parsed
3. The grammar covers the syntax used in examples

## Usage

```bash
# Build the tool
go build -o grammar-validator .

# Test a single file
./grammar-validator examples/main.vx

# Test multiple files
./grammar-validator examples/*.vx

# Test all .vx files
find examples -name "*.vx" -exec ./grammar-validator {} +
```

## Dependencies

- Requires generated ANTLR Go parser files (`vex_lexer.go`, `vex_parser.go`, etc.)
- Uses `github.com/antlr4-go/antlr/v4` runtime

## CI Integration

This tool is automatically used in the grammar validation workflow to test all example files whenever the grammar or examples change.