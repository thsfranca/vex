# Fugo Language Development Makefile

.PHONY: help install-extension auto-install build-transpiler test clean dev watch

help: ## Show this help message
	@echo "Fugo Development Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

install-extension: ## Quick reinstall VSCode extension
	@echo "🚀 Reinstalling Fugo VSCode Extension..."
	@cd vscode-extension && ./quick-install.sh

auto-install: ## Watch for changes and auto-reinstall extension
	@echo "👀 Starting auto-install watcher..."
	@cd vscode-extension && ./auto-install.sh

watch: auto-install ## Alias for auto-install

build-transpiler: ## Build the Go transpiler
	@echo "🔨 Building Fugo transpiler..."
	@go build -o bin/fugo ./cmd/fugo

test: ## Run all tests
	@echo "🧪 Running tests..."
	@go test ./...

clean: ## Clean build artifacts
	@echo "🧹 Cleaning up..."
	@rm -rf bin/
	@rm -f vscode-extension/*.vsix
	@rm -f vscode-extension/fugo-minimal-latest.vsix

dev: install-extension ## Quick development setup (reinstall extension)
	@echo "✅ Development environment ready!"