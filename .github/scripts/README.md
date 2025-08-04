# GitHub Scripts Organization

This directory contains all automation scripts used in GitHub Actions workflows, organized by functional purpose for better maintainability.

## Directory Structure

```
.github/scripts/
├── build/          # Build and compilation scripts
├── test/           # Testing and coverage scripts  
├── lint/           # Code quality and linting scripts
├── release/        # Release management scripts
├── extension/      # VSCode extension scripts
├── grammar/        # Grammar and parser scripts
└── utils/          # Shared utilities and helpers
```

## Directory Purposes

### 🔨 `build/`
Scripts for building and compiling the project:
- `build-project.sh` - Main project compilation
- `generate-parsers.sh` - ANTLR parser generation
- `install-antlr.sh` - ANTLR installation
- `setup-antlr-binary.sh` - ANTLR binary setup

### 🧪 `test/`
Scripts for testing and coverage analysis:
- `run-tests.sh` - Execute test suites
- `run-coverage.sh` - Basic coverage collection
- `run-coverage-full.sh` - Comprehensive coverage analysis
- `simple-coverage-check.sh` - Simplified coverage validation and PR reporting
- `post-coverage-comment.sh` - PR coverage comment automation
- `install-coverage-tools.sh` - Coverage tooling setup

### 🔍 `lint/`
Scripts for code quality and style checking:
- `run-linting.sh` - Execute linting checks
- `check-formatting.sh` - Code formatting validation
- `install-go-linting-tools.sh` - Linting tools installation

### 🚀 `release/`
Scripts for release management and automation:
- `create-github-release.sh` - GitHub release creation
- `create-release-notes.sh` - Manual release notes
- `auto-create-release-notes.sh` - Automated release notes
- `auto-bump-version.sh` - Version bumping
- `check-release-labels.sh` - PR label validation
- `configure-git.sh` - Git configuration for releases
- `success-notification.sh` - Release success notifications

### 🎨 `extension/`
Scripts specific to VSCode extension development:
- Extension validation and packaging scripts
- Syntax highlighting tests
- Theme validation
- Manifest checks

### 📝 `grammar/`
Scripts for grammar development and validation:
- Grammar syntax validation
- Parser generation tests
- Example validation

### 🛠️ `utils/`
Shared utilities and helper scripts:
- `install-go-deps.sh` - Go dependency management
- `generate-summary.sh` - Workflow summaries
- `makefile-analysis.sh` - Makefile validation
- `final-go-decision.sh` - Go workflow decisions
- `skip-build.sh` - Build skip logic
- `decide-validation-needed.sh` - Validation decisions
- `evaluate-extension-results.sh` - Extension result evaluation
- `evaluate-grammar-results.sh` - Grammar result evaluation

## Usage Guidelines

1. **Keep scripts focused** - Each script should have a single, clear purpose
2. **Use descriptive names** - Script names should indicate their function
3. **Document parameters** - Include usage comments in script headers
4. **Handle errors gracefully** - Use proper exit codes and error messages
5. **Follow project conventions** - Match existing style and patterns

## Migration Notes

This structure was reorganized from the previous approach to:
- Eliminate overlapping directories (`ci/`, `auto-release/`, `validation/`)
- Group scripts by functional purpose rather than arbitrary categories
- Make it easier to find and maintain related scripts
- Improve clarity for contributors and maintainers
