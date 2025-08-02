.PHONY: generate clean help java go python cpp javascript

# Default target
help:
	@echo "Fugo Grammar - ANTLR Parser Generator"
	@echo ""
	@echo "Available targets:"
	@echo "  generate     - Generate parsers for all languages"
	@echo "  java         - Generate Java parser"
	@echo "  go           - Generate Go parser"
	@echo "  python       - Generate Python parser"
	@echo "  cpp          - Generate C++ parser"
	@echo "  javascript   - Generate JavaScript parser"
	@echo "  clean        - Clean generated files"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Requirements:"
	@echo "  - ANTLR4 installed and available in PATH"
	@echo ""

# Generate parsers for all supported languages
generate: java go python cpp javascript

# Generate Java parser
java:
	@echo "Generating Java parser..."
	@mkdir -p tools/gen/java
	antlr -Dlanguage=Java -o tools/gen/java tools/grammar/Fugo.g4

# Generate Go parser
go:
	@echo "Generating Go parser..."
	@mkdir -p tools/gen/go
	antlr -Dlanguage=Go -o tools/gen/go tools/grammar/Fugo.g4

# Generate Python parser
python:
	@echo "Generating Python parser..."
	@mkdir -p tools/gen/python
	antlr -Dlanguage=Python3 -o tools/gen/python tools/grammar/Fugo.g4

# Generate C++ parser
cpp:
	@echo "Generating C++ parser..."
	@mkdir -p tools/gen/cpp
	antlr -Dlanguage=Cpp -o tools/gen/cpp tools/grammar/Fugo.g4

# Generate JavaScript parser
javascript:
	@echo "Generating JavaScript parser..."
	@mkdir -p tools/gen/javascript
	antlr -Dlanguage=JavaScript -o tools/gen/javascript tools/grammar/Fugo.g4

# Clean all generated files
clean:
	@echo "Cleaning generated files..."
	rm -rf tools/gen/

# Check if ANTLR4 is available
check-antlr:
	@which antlr > /dev/null || (echo "Error: antlr not found in PATH. Please install ANTLR4." && exit 1)