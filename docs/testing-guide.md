# Vex Testing Guide

## Overview

The Vex testing framework provides a comprehensive solution for testing Vex programs with automatic test discovery, macro-based assertions, detailed coverage reporting, and CI/CD integration. It follows familiar testing patterns while leveraging Vex's functional programming features and Hindley-Milner type system.

## Quick Start

### 1. Create a Test File

Create a file ending in `_test.vx`:

```vex
;; math_test.vx

;; Test a simple function with explicit types (current requirement)
(defn add [x: number y: number] -> number (+ x y))

;; Only deftest declarations are allowed in test files
(deftest "addition-test"
  (assert-eq (add 2 3) 5 "2 + 3 should equal 5"))
```

### 2. Run Tests

```bash
./vex test                    # Run all tests
./vex test -verbose           # Detailed output
./vex test -coverage          # Include basic coverage report
./vex test -enhanced-coverage # Advanced function/line/branch/quality analysis
./vex test -dir ./src         # Test specific directory
```

## Test Structure

### File Naming Convention

- Test files must end with `_test.vx` or `_test.vex`
- Place test files in the same directory as the code they test
- Example: `calculator.vx` → `calculator_test.vx`

### Test Organization

```vex
;; user_service_test.vx

;; Import dependencies
(import "strings")

;; Define helper functions with explicit types
(defn create-user [name: string email: string] -> map
  {"name" name "email" email})

(defn validate-email [email: string] -> bool
  (strings/Contains email "@"))

;; Test basic functionality - only deftest blocks allowed
(deftest "user-creation"
  (do
    (def user (create-user "Alice" "alice@example.com"))
    (assert-eq (get user "name") "Alice" "user name should be set")
    (assert-eq (get user "email") "alice@example.com" "user email should be set")))

;; Test edge cases
(deftest "email-validation"
  (do
    (assert-eq (validate-email "valid@email.com") true "valid email")
    (assert-eq (validate-email "invalid-email") false "invalid email")))

;; Test error conditions
(deftest "empty-user-fields"
  (do
    (def user (create-user "" ""))
    (assert-eq (get user "name") "" "empty name should be preserved")
    (assert-eq (get user "email") "" "empty email should be preserved")))
```

## Test Macros

### `assert-eq`

Compares two values for equality:

```vex
(assert-eq actual expected "description")
```

**Examples:**
```vex
(assert-eq (+ 1 1) 2 "basic arithmetic")
(assert-eq (len [1 2 3]) 3 "array length")
(assert-eq (:key {:key "value"}) "value" "map access")
```

**Output:**
- Success: `PASS: basic arithmetic`
- Failure: Test execution stops with error and diagnostic information from HM type system

### `deftest`

Defines a named test with setup and teardown:

```vex
(deftest "test-name" body)
```

**Single Assertion:**
```vex
(deftest "simple-test"
  (assert-eq (+ 1 1) 2 "addition"))
```

**Multiple Assertions:**
```vex
(deftest "complex-test"
  (do
    (def x 10)
    (def y 20)
    (assert-eq (+ x y) 30 "addition")
    (assert-eq (- x y) -10 "subtraction")))
```

## Advanced Testing Patterns

### Testing Functions with Side Effects

```vex
;; file_operations_test.vx

(import ["os" "fmt"])

;; Helper functions with explicit types
(defn write-test-file [filename: string content: string] -> bool
  (os/WriteFile filename content 0644)
  true)

(defn read-test-file [filename: string] -> string
  (os/ReadFile filename))

(deftest "file-operations"
  (do
    ;; Setup
    (def test-file "test-output.txt")
    
    ;; Test file creation
    (def content "Hello, Vex!")
    (assert-eq (write-test-file test-file content) true "file should be written")
    
    ;; Test file reading
    (def read-content (read-test-file test-file))
    (assert-eq read-content content "file content should match")
    
    ;; Cleanup
    (os/Remove test-file)))
```

### Testing Error Conditions

```vex
;; error_handling_test.vx

;; Helper functions with explicit types
(defn safe-divide [x: number y: number] -> map
  (if (= y 0)
    {"error" "division by zero"}
    {"result" (/ x y)}))

(deftest "division-error-handling"
  (do
    (def result (safe-divide 10 2))
    (assert-eq (get result "result") 5 "normal division")
    
    (def error-result (safe-divide 10 0))
    (assert-eq (get error-result "error") "division by zero" "division by zero error")))
```

### Testing with Mock Data

```vex
;; api_test.vx

;; Mock HTTP response with explicit typing
(def mock-user-response
  {"id" 1 "name" "John Doe" "email" "john@example.com"})

(defn parse-user [response: map] -> map
  {"id" (get response "id")
   "name" (get response "name")
   "email" (get response "email")})

(deftest "user-parsing"
  (do
    (def parsed (parse-user mock-user-response))
    (assert-eq (get parsed "id") 1 "user ID")
    (assert-eq (get parsed "name") "John Doe" "user name")
    (assert-eq (get parsed "email") "john@example.com" "user email")))
```

## Test Coverage

Vex provides **multi-dimensional coverage analysis** from basic file-level reporting to advanced function-level precision with quality scoring.

### Basic Coverage Analysis

File-level coverage analysis per package (directory):

```bash
./vex test -coverage
```

**Sample Output:**
```
📊 Generating test coverage report...

📋 Test Coverage Report
─────────────────────────
✅ ./src/utils: 100.0% (3/3 files)
📈 ./src/api: 66.7% (2/3 files)
⚠️ ./src/models: 25.0% (1/4 files)
❌ ./src/legacy: 0.0% (0/2 files)
─────────────────────────
📊 Overall: 54.5% (6/11 packages have tests)
💡 Consider adding more tests to improve coverage
```

### Enhanced Coverage Analysis

**Ultra-precise function-level analysis** with line coverage, branch coverage, and test quality scoring:

```bash
./vex test -enhanced-coverage
```

**Sample Output:**
```
🚀 Generating Enhanced Coverage Analysis...

📊 Enhanced Coverage Report
═══════════════════════════
📈 Overall Coverage:
   Function-Level: 65.4% (17/26 functions tested)
   File-Level: 83.3% (10/12 files tested)

✅ utils: 100.0% (8/8 functions), 100.0% (3/3 files)
     📝 Line: 92.3% (24/26)
     🌿 Branch: 87.5% (7/8)
     🎯 Quality: 92.0/100 (3.2 assertions/test)

📈 api: 75.0% (6/8 functions), 66.7% (2/3 files)
   Untested: validateInput, logError
     📝 Line: 68.4% (13/19)
     🌿 Branch: 50.0% (2/4)
     🎯 Quality: 75.0/100 (1.8 assertions/test)

❌ models: 20.0% (1/5 functions), 25.0% (1/4 files)
   Untested: User, validateUser, saveUser, deleteUser
     📝 Line: 15.8% (3/19)
     🎯 Quality: 45.0/100 (0.5 assertions/test)

❌ legacy: 0.0% (0/5 functions), 0.0% (0/2 files)
   Untested: OldAPI, processLegacy, migrate, cleanup, archive
     📝 Line: 0.0% (0/23)
     🌿 Branch: 0.0% (0/6)
     🎯 Quality: 0.0/100 (0.0 assertions/test)

💡 Coverage Insights:
   Coverage precision: 40% → 89% (function-level analysis)
   Priority functions to test: validateInput (api), User (models), OldAPI (legacy)
   📋 Packages needing attention: models, legacy
   🏆 Well-tested packages: utils
```

### Coverage Dimensions Explained

#### 1. **Function Coverage**
- Tracks specific functions (`defn`, `defmacro`, `def+fn`, `macro`) that are tested
- **More accurate** than file coverage - reveals untested functions in "covered" files
- Shows exactly which functions need test attention

#### 2. **Line Coverage** 
- Analyzes individual lines of code (excluding comments/imports)
- Identifies **precise lines** that lack test coverage
- Helps write targeted tests for specific code sections

#### 3. **Branch Coverage**
- Detects conditional branches (`if`, `when`, `unless`, `cond`)
- Tracks **true/false path coverage** for each condition
- Ensures comprehensive testing of all code paths

#### 4. **Test Quality Score (0-100)**
- **Assertion Density**: Optimal 1-10 assertions per test
- **Edge Case Coverage**: Tests boundary values, nil/empty inputs
- **Method Diversity**: Uses variety of assertion types (`assert-eq`, `assert-true`, etc.)
- **Naming Quality**: Descriptive test names that explain what's being tested

### Coverage Indicators

**File-Level:**
- **✅ Green (80%+)**: Excellent coverage
- **📈 Blue (50-79%)**: Good coverage  
- **⚠️ Yellow (<50%)**: Needs improvement
- **❌ Red (0%)**: No tests

**Function-Level:**
- **Function %**: Percentage of functions with any test coverage
- **Line %**: Percentage of executable lines covered by tests
- **Branch %**: Percentage of conditional branches tested
- **Quality Score**: Overall test quality rating (0-100)

### Systematic Coverage Improvement

#### 1. **Identify Priority Areas**
```bash
# Enhanced analysis shows exactly what needs attention
./vex test -enhanced-coverage -coverage-out analysis.json
```

#### 2. **Focus on Untested Functions**
```bash
# The report shows specific functions like "validateInput, logError"
# Create targeted tests for these functions
```

#### 3. **Improve Test Quality**
Based on quality scores and suggestions:
- **Low assertion density**: Add more `assert-eq` calls per test
- **Missing edge cases**: Test empty inputs, boundary values, error conditions
- **Poor naming**: Use descriptive test names that explain the scenario

#### 4. **Verify Improvements**  
```bash
# Compare before/after with enhanced coverage
./vex test -enhanced-coverage
```

### Example: Improving a Package

**Before** (Poor Quality):
```vex
(deftest "test1"
  (fmt/Println "testing something"))  ; No assertions!
```

**After** (High Quality):
```vex
(deftest "user-validation-accepts-valid-email"
  (assert-eq (validate-email "user@example.com") true "valid email should pass")
  (assert-eq (validate-email "") false "empty email should fail")
  (assert-eq (validate-email "invalid") false "malformed email should fail"))

(deftest "user-validation-handles-edge-cases"
  (assert-eq (validate-email nil) false "nil email should fail")
  (assert-eq (validate-email "a@b.co") true "minimal valid email should pass"))
```

**Result**: Quality score improves from 15/100 to 88/100

## Best Practices

> **📖 Reference**: For complete test message standards, see [test-message-standards.md](test-message-standards.md)

### Test Naming

Follow the **subject-action-expectation** pattern for test names:

- Use descriptive test names that explain what is being tested
- Group related tests logically
- Use kebab-case naming convention
- Follow the pattern: `"subject-action-expectation"`

```vex
;; Good - Clear, descriptive names following kebab-case
(deftest "user-authentication-validates-correct-credentials"
  (assert-eq (authenticate "user" "pass") true "valid-credentials-authenticate-successfully"))

(deftest "user-authentication-rejects-invalid-credentials"
  (assert-eq (authenticate "user" "wrong") false "invalid-credentials-fail-authentication"))

;; Avoid - Vague or unclear names
(deftest "test1"
  (assert-eq (authenticate "user" "pass") true "test"))

(deftest "user_test"  ; Wrong: uses underscores instead of kebab-case
  (assert-eq (authenticate "user" "pass") true "check user"))
```

**Naming Pattern Examples**:
- Business Logic: `"payment-processing-handles-invalid-cards"`
- API Testing: `"get-users-endpoint-returns-success-status"`
- Edge Cases: `"edge-case-calculation-handles-zero-input"`
- Error Cases: `"error-case-validation-rejects-malformed-input"`

### Test Organization

- One test file per source file when possible
- Group related functionality in test suites
- Keep tests focused and atomic

```vex
;; calculator_test.vx

;; Basic operations
(deftest "addition" ...)
(deftest "subtraction" ...)
(deftest "multiplication" ...)
(deftest "division" ...)

;; Edge cases
(deftest "division-by-zero" ...)
(deftest "large-numbers" ...)
(deftest "negative-numbers" ...)
```

### Test Data

- Use meaningful test data that reflects real usage
- Create helper functions for complex test data
- Keep test data close to the tests that use it

```vex
;; Helper function for test data
(defn sample-user []
  {:id 123
   :name "Test User"
   :email "test@example.com"
   :created-at "2024-01-01"})

(deftest "user-serialization"
  (do
    (def user (sample-user))
    (def json (to-json user))
    (assert-eq (contains json "Test User") true "name in JSON")))
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Vex Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      
      - name: Build Vex CLI
        run: go build -o vex cmd/vex-transpiler/main.go
      
      - name: Run Tests with Coverage
        run: ./vex test -coverage -verbose
        env:
          VEX_STDLIB_PATH: ${{ github.workspace }}/stdlib
      
      - name: Check Coverage Threshold
        run: |
          # Custom script to check if coverage meets minimum threshold
          ./scripts/check-coverage.sh 70  # Require 70% coverage
```

### Coverage Thresholds

Create a script to enforce minimum coverage:

```bash
#!/bin/bash
# scripts/check-coverage.sh
THRESHOLD=${1:-80}
COVERAGE=$(./vex test -coverage 2>&1 | grep "Overall:" | grep -o '[0-9.]*%' | cut -d'%' -f1)

if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
    echo "❌ Coverage $COVERAGE% is below threshold $THRESHOLD%"
    exit 1
else
    echo "✅ Coverage $COVERAGE% meets threshold $THRESHOLD%"
fi
```

## Troubleshooting

### Common Issues

**Test files not discovered:**
- Ensure files end with `_test.vx` or `_test.vex`
- Check file permissions
- Verify directory structure
- Use `./vex test -verbose` to see discovery process

**Test validation errors:**
- Only `(deftest ...)` declarations are allowed in test files
- All other code (helper functions, imports) must be inside deftest blocks
- Check syntax: `(deftest "name" body)`
- Ensure `assert-eq` has exactly 3 arguments

**Type errors in tests:**
- Vex uses complete Hindley-Milner type inference with explicit type requirements
- All function parameters must have explicit type annotations: `[param: type]`
- Function return types must be specified: `-> returnType`
- Ensure test data types match function expectations
- Review HM diagnostic codes (VEX-TYP-*) for specific type errors

**Build failures:**
- Check that all imports are available and properly typed
- Verify Go compiler is installed and accessible
- Review package resolution with `-verbose` flag
- Check for circular dependencies in local packages

### Debug Mode

Run tests with maximum verbosity:

```bash
./vex test -verbose -dir . 2>&1 | tee test-debug.log
```

This captures all output including:
- Test discovery process
- Transpilation details
- Build commands
- Test execution results
- Coverage analysis

## Future Enhancements

The Vex testing framework is continuously evolving. Planned features include:

- **Property-based testing** with generators
- **Test fixtures** for setup/teardown
- **Parallel test execution** for faster runs
- **Benchmark testing** for performance analysis
- **Integration with Go testing.T** for advanced CI/CD workflows
- **Visual coverage reports** with HTML output
- **Test mutation** for testing test quality

## Examples Repository

For complete examples and advanced usage patterns, see:
- `examples/testing/` - Basic test examples
- `examples/coverage/` - Coverage analysis examples
- `examples/integration/` - CI/CD integration examples
