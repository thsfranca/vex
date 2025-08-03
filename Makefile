# Flux Language Development Makefile

.PHONY: help install-extension auto-install build-transpiler test clean dev watch go

help: ## Show this help message
	@echo "Flux Development Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

install-extension: ## Quick reinstall VSCode extension
	@echo "🚀 Reinstalling Flux VSCode Extension..."
	@cd vscode-extension && ./quick-install.sh

auto-install: ## Watch for changes and auto-reinstall extension
	@echo "👀 Starting auto-install watcher..."
	@cd vscode-extension && ./auto-install.sh

watch: auto-install ## Alias for auto-install

build-transpiler: ## Build the Go transpiler
	@echo "🔨 Building Flux transpiler..."
	@go build -o bin/flux ./cmd/flux

go: ## Generate Go parser from grammar
	@cd tools/grammar && antlr4 -Dlanguage=Go -listener -visitor Flux.g4 -o ../gen/go/

test: ## Run all tests
	@echo "🧪 Running tests..."
	@go test ./...

clean: ## Clean build artifacts
	@echo "🧹 Cleaning up..."
	@rm -rf bin/
	@rm -f vscode-extension/*.vsix
	@rm -f vscode-extension/flux-minimal-latest.vsix

dev: install-extension ## Quick development setup (reinstall extension)
	@echo "✅ Development environment ready!"