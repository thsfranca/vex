# Vex Testing Guide

## Overview

The Vex testing framework provides a comprehensive solution for testing Vex programs with automatic test discovery, macro-based assertions, real execution-based coverage reporting, and CI/CD integration. It follows familiar testing patterns while leveraging Vex's functional programming features and Hindley-Milner type system.

## Quick Start

### 1. Create a Test File

Create a file ending in `_test.vx`:

```vex
;; math_test.vx

;; Import required modules for testing
(import ["fmt" "test"])

;; Only deftest declarations are allowed in test files
(deftest "addition-test"
  (do
    (fmt/Println "Testing basic addition")
    (def result (+ 2 3))
    (assert-eq result 5 "2 + 3 should equal 5")))
```

### 2. Run Tests

```bash
./vex test                    # Run all tests
./vex test -verbose           # Detailed output
./vex test -coverage          # Include basic coverage report
./vex test -enhanced-coverage # Real execution-based coverage analysis
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

;; Import required dependencies
(import ["fmt" "test" "strings"])

;; Test basic functionality - only deftest blocks allowed
(deftest "string-operations"
  (do
    (fmt/Println "Testing string operations")
    (def email "alice@example.com")
    (assert-true (strings/Contains email "@") "email contains @ symbol")
    (assert-eq (strings/ToUpper "hello") "HELLO" "uppercase conversion")))

;; Test edge cases  
(deftest "empty-string-handling"
  (do
    (fmt/Println "Testing empty strings")
    (def empty "")
    (assert-eq (len empty) 0 "empty string length")
    (assert-false (strings/Contains empty "@") "empty string contains check")))

;; Test multiple assertions
(deftest "comprehensive-string-test"
  (do
    (fmt/Println "Testing comprehensive string operations")
    (def name "Alice")
    (def domain "example.com")
    (def email (strings/Join [name "@" domain] ""))
    (assert-eq email "Alice@example.com" "string concatenation")
    (assert-true (> (len email) 0) "result is not empty")))
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
(assert-eq "hello" "hello" "string comparison")
```

**Output:**
- Success: `✅ PASS: basic arithmetic`
- Failure: `❌ FAIL: Expected 2, got 3: basic arithmetic`

### `assert-true` and `assert-false`

Test boolean conditions:

```vex
(assert-true condition "description")
(assert-false condition "description")
```

**Examples:**
```vex
(assert-true (> 5 3) "5 is greater than 3")
(assert-false (< 10 5) "10 is not less than 5")
(assert-true (strings/Contains "hello" "ell") "substring check")
```

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

Vex provides **real execution-based coverage analysis** from basic file-level reporting to advanced runtime instrumentation with 100% accurate data.

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

**Real execution-based analysis** using Go runtime instrumentation for 100% accurate coverage data:

```bash
./vex test -enhanced-coverage
```

**Sample Output (Successful Tests):**
```
🚀 Generating Enhanced Coverage Analysis...

📊 Enhanced Coverage Report
═══════════════════════════
📄 Coverage detected: /path/to/generated/test_file.go
📈 Overall Coverage:
   Execution-Based: 100.0% (1/1 files executed)
   Profile Sources: 1 coverage profile(s)
   Data Quality: REAL execution data ✅

💡 Coverage Insights:
   Coverage precision: REAL execution data (100% accurate)
   Data source: Go runtime instrumentation
```

**Sample Output (Failed Tests):**
```
🚀 Generating Enhanced Coverage Analysis...

📊 Enhanced Coverage Report
═══════════════════════════
No coverage data available (tests did not execute successfully)
```

### Coverage Analysis Explained

#### **Real Execution-Based Coverage**
- Uses Go's runtime instrumentation (`-cover` flag) during test execution
- **100% accurate data** - only counts code that actually runs during tests
- **Execution profiles** parsed from Go coverage output format
- **No heuristics** - every coverage metric reflects real code execution

#### **Coverage Data Sources**
- **Go Runtime Instrumentation**: Uses `go run -cover` with `GOCOVERDIR` environment
- **Coverage Profile Parsing**: Reads actual Go coverage format (`filename:line.col,line.col numstmt count`)
- **Test Result Integration**: Coverage only collected when tests pass successfully
- **Profile Cleanup**: Temporary coverage files automatically cleaned up after analysis

#### **Accuracy Guarantees**
- **Real Execution Only**: No coverage reported unless code actually executed
- **Test Failure Handling**: "No coverage data available" when tests fail to run
- **Profile Validation**: Checks for actual coverage file generation before analysis
- **Quality Indicators**: Always shows "REAL execution data ✅" for valid coverage

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
