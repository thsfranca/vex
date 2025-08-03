# Vex Language Development Makefile

.PHONY: help install-extension auto-install build-transpiler test clean dev watch go validate-grammar build-tools

help: ## Show this help message
	@echo "Vex Development Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

install-extension: ## Quick reinstall VSCode extension
	@echo "🚀 Reinstalling Vex VSCode Extension..."
	@cd vscode-extension && ./quick-install.sh

auto-install: ## Watch for changes and auto-reinstall extension
	@echo "👀 Starting auto-install watcher..."
	@cd vscode-extension && ./auto-install.sh

watch: auto-install ## Alias for auto-install

build-transpiler: ## Build the Go transpiler
	@echo "🔨 Building Vex transpiler..."
	@go build -o bin/vex ./cmd/vex

go: ## Generate Go parser from grammar
	@cd tools/grammar && antlr -Dlanguage=Go -listener -visitor Vex.g4 -o ../gen/go/

validate-grammar: ## Validate grammar by testing example files (requires ANTLR4)
	@echo "🔨 Generating Vex parser files..."
	@mkdir -p tools/grammar-validator/parser
	@cd tools/grammar && antlr -Dlanguage=Go -listener -visitor Vex.g4 -o ../grammar-validator/parser/
	@echo "🔨 Building grammar validator..."
	@cd tools/grammar-validator && go build -o grammar-validator .
	@echo "🧪 Testing .vx example files..."
	@cd tools/grammar-validator && ./grammar-validator ../../examples/*.vx

test: ## Run all tests
	@echo "🧪 Running tests..."
	@go test ./...

clean: ## Clean build artifacts
	@echo "🧹 Cleaning up..."
	@rm -rf bin/
	@rm -f vscode-extension/*.vsix
	@rm -f vscode-extension/vex-minimal-latest.vsix

build-tools: ## Build all CI/development tools
	@echo "🔨 Building working tools..."
	@cd tools/change-detector && go build -o change-detector . && echo "✅ change-detector built"
	@cd tools/coverage-updater && go build -o coverage-updater . && echo "✅ coverage-updater built"
	@cd tools/extension-tester && go build -o extension-tester . && echo "✅ extension-tester built"
	@cd tools/debug-helper && go build -o debug-helper . && echo "✅ debug-helper built"
	@echo "⚠️ grammar-validator skipped (requires ANTLR4 to generate parser files)"
	@echo "✅ Working tools built successfully!"

dev: install-extension ## Quick development setup (reinstall extension)
	@echo "✅ Development environment ready!"