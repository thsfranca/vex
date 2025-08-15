# Vex Test Message Standards

## Overview

This document defines the comprehensive message standards for the Vex testing framework, covering both framework-generated messages and user-written test messages. These standards ensure consistency, clarity, and alignment with Vex's AI-first design principles and professional tooling requirements.

## Core Principles

**Professional Clarity**: All messages use clean text formatting without emojis, ensuring compatibility with CI/CD systems and professional development environments.

**AI-Friendly Patterns**: Consistent, predictable message structures that AI models can generate and parse reliably.

**Semantic Clarity**: Messages express intent and context clearly for both human developers and automated tools.

**Kebab-Case Consistency**: All user-defined identifiers in test messages follow Vex's kebab-case naming convention.

## Framework-Generated Messages

### Test Discovery & Execution Headers

**Standard Format**:
```
Running Vex tests in {directory} (found {count} test files)
Pattern filter: {pattern}
Fail-fast mode enabled
Timeout: {duration}

Executing: {test-file-path} ({current}/{total})
```

**Example**:
```
Running Vex tests in ./src (found 8 test files)
Pattern filter: user-service
Fail-fast mode enabled
Timeout: 30s

Executing: src/user_service_test.vx (1/8)
```

### Test Result Messages

**Status Message Format**:
```
{STATUS}: {file-path} ({duration})[optional-context]
```

**Status Types**:
- `PASS`: Test completed successfully with all assertions passing
- `FAIL`: Test completed with one or more assertion failures
- `BUILD_ERROR`: Go compilation failed during test execution
- `TRANSPILE_ERROR`: Vex to Go translation failed
- `TIMEOUT`: Test execution exceeded timeout duration
- `VALIDATION_ERROR`: Test file structure validation failed
- `SKIP`: Test excluded by pattern filter or other criteria

**Examples**:
```
PASS: src/user_service_test.vx (125ms)
FAIL: src/payment_test.vx (89ms) - assertion failures
BUILD_ERROR: src/api_test.vx (45ms) - Go compilation failed
TRANSPILE_ERROR: src/invalid_test.vx (12ms) - Vex to Go translation failed
TIMEOUT: src/slow_test.vx (30.1s) - exceeded 30s
VALIDATION_ERROR: src/malformed_test.vx (8ms) - test file structure invalid
SKIP: src/excluded_test.vx (0ms) - pattern filter excluded
```

### Summary Messages

**Standard Format**:
```
Test Summary
=======================
Total: {count} tests
Passed: {count}
Failed: {count}
Skipped: {count}
Duration: {duration}
Success rate: {percentage}%

Next steps: Use './vex test -verbose' to see detailed error information
```

**Example**:
```
Test Summary
=======================
Total: 12 tests
Passed: 10
Failed: 2
Skipped: 0
Duration: 2.3s
Success rate: 83.3%

Next steps: Use './vex test -verbose' to see detailed error information
```

### Coverage Report Messages

#### Basic Coverage Format

**Structure**:
```
Generating test coverage report...

Test Coverage Report
=======================
{package-path}: {percentage}% ({tested-files}/{total-files} files)
...
=======================
Overall: {percentage}% ({tested-packages}/{total-packages} packages have tests)
Consider adding tests to improve coverage
Target packages: {low-coverage-packages}
```

**Example**:
```
Generating test coverage report...

Test Coverage Report
=======================
./src/utils: 100.0% (3/3 files)
./src/api: 66.7% (2/3 files)
./src/models: 25.0% (1/4 files)
./src/legacy: 0.0% (0/2 files)
=======================
Overall: 54.5% (6/11 packages have tests)
Consider adding tests to improve coverage
Target packages: models, legacy
```

#### Enhanced Coverage Format

**Structure**:
```
Generating Enhanced Coverage Analysis...

Enhanced Coverage Report
===============================
Overall Coverage:
  Function-Level: {percentage}% ({tested-functions}/{total-functions} functions tested)
  File-Level: {percentage}% ({tested-files}/{total-files} files tested)

{package}: {function-percentage}% ({tested-functions}/{total-functions} functions), {file-percentage}% ({tested-files}/{total-files} files)
  Untested: {untested-function-list}
  Line: {percentage}% ({covered-lines}/{total-lines})
  Branch: {percentage}% ({covered-branches}/{total-branches})
  Quality: {score}/100 ({assertions-per-test} assertions/test)

Coverage Insights:
  Coverage precision: {old-percentage}% -> {new-percentage}% (function-level analysis)
  Priority functions to test: {priority-functions}
  Packages needing attention: {attention-packages}
  Well-tested packages: {well-tested-packages}
```

### Error Output Messages

#### Assertion Failure Format

**Structure**:
```
FAIL: {test-name}
  Expected: {expected-value}
  Actual: {actual-value}
  Message: {assertion-message}
  Location: {file}:{line}:{column}
```

**Example**:
```
FAIL: user-authentication-validates-credentials
  Expected: true
  Actual: false
  Message: valid-email-format-passes-validation
  Location: user_service_test.vx:15:5
```

#### Build/Transpile Error Format

**Structure**:
```
{ERROR_TYPE}: {file-path}
  Error: {error-description}
  Location: {file}:{line}:{column}
  Suggestion: {helpful-suggestion}
```

**Example**:
```
TRANSPILE_ERROR: payment_processing_test.vx
  Error: undefined identifier 'process-payment'
  Location: payment_processing_test.vx:23:12
  Suggestion: define 'process-payment' or import the package that provides it
```

## User-Written Test Messages

### Test Definition Standards

#### Test Name Format

**Pattern**: `"subject-action-expectation"`

**Structure Components**:
- **Subject**: What system/feature is being tested (kebab-case)
- **Action**: What operation or behavior (kebab-case verb)
- **Expectation**: Expected outcome (kebab-case description)

**Examples**:
```vex
;; Business Logic Tests
(deftest "user-authentication-validates-credentials" ...)
(deftest "payment-processing-handles-invalid-cards" ...)
(deftest "order-calculation-applies-bulk-discounts" ...)

;; API/HTTP Tests
(deftest "api-endpoint-returns-proper-json-structure" ...)
(deftest "get-users-endpoint-returns-success-status" ...)
(deftest "post-user-endpoint-validates-required-fields" ...)

;; Data Processing Tests
(deftest "filter-operation-preserves-matching-records" ...)
(deftest "sort-operation-handles-null-values-gracefully" ...)
(deftest "aggregation-computes-correct-totals" ...)

;; Error Handling Tests
(deftest "validation-rejects-malformed-input" ...)
(deftest "error-handler-logs-appropriate-messages" ...)
(deftest "timeout-handler-cleans-up-resources" ...)
```

#### Test Organization Patterns

**Group by Domain**:
```vex
;; user-service_test.vx

;; Authentication tests
(deftest "user-authentication-succeeds-with-valid-credentials" ...)
(deftest "user-authentication-fails-with-invalid-password" ...)
(deftest "user-authentication-blocks-disabled-accounts" ...)

;; Profile management tests
(deftest "user-profile-updates-successfully-with-valid-data" ...)
(deftest "user-profile-validates-email-format-before-saving" ...)
(deftest "user-profile-preserves-existing-data-on-partial-update" ...)
```

**Edge Case Naming Convention**:
```vex
;; Standard cases
(deftest "calculation-handles-positive-numbers" ...)

;; Edge cases - use "edge-case-" prefix
(deftest "edge-case-calculation-handles-zero-input" ...)
(deftest "edge-case-calculation-handles-negative-numbers" ...)
(deftest "edge-case-calculation-handles-maximum-integer" ...)

;; Error cases - use "error-case-" prefix
(deftest "error-case-calculation-rejects-non-numeric-input" ...)
(deftest "error-case-calculation-handles-division-by-zero" ...)
```

### Assertion Message Standards

#### Message Format

**Pattern**: `"action-context-expectation"`

**Guidelines**:
- Use kebab-case for consistency with Vex naming
- Focus on business logic, not implementation details
- Explain why the assertion matters, not what it does
- Be specific about the expected behavior

#### Business Logic Assertions

```vex
(assert-eq (validate-email "user@test.com") true "valid-email-format-passes-validation")
(assert-eq (calculate-tax 100.00) 8.50 "standard-tax-rate-applies-correctly")
(assert-eq (user-permissions admin-user) ["read" "write" "admin"] "admin-role-grants-full-permissions")
(assert-eq (apply-discount order-total 0.10) 90.00 "ten-percent-discount-reduces-total-correctly")
(assert-eq (is-eligible-for-promotion user) true "active-users-qualify-for-current-promotion")
```

#### API/HTTP Assertions

```vex
(assert-eq (status-code response) 200 "get-users-endpoint-returns-success-status")
(assert-eq (content-type response) "application/json" "api-response-uses-json-content-type")
(assert-eq (len (get response "data")) 5 "user-list-contains-expected-record-count")
(assert-eq (get response "status") "success" "successful-operation-returns-success-indicator")
(assert-eq (has-field? response "pagination") true "paginated-endpoint-includes-pagination-metadata")
```

#### Data Processing Assertions

```vex
(assert-eq (count filtered-results) 3 "filter-preserves-only-matching-records")
(assert-eq (first sorted-list) smallest-value "sort-operation-places-minimum-first")
(assert-eq (empty? processed-queue) true "processing-completely-drains-work-queue")
(assert-eq (len grouped-data) expected-groups "grouping-creates-correct-number-of-buckets")
(assert-eq (sum aggregated-values) total-expected "aggregation-computes-accurate-sum")
```

#### Error Condition Assertions

```vex
(assert-eq (has-error? result) true "invalid-input-triggers-appropriate-error")
(assert-eq (error-code result) "VALIDATION_FAILED" "validation-error-sets-correct-error-code")
(assert-eq (error-message result) "Email format invalid" "validation-provides-human-readable-message")
(assert-eq (retry-count failed-operation) 3 "failed-operation-attempts-configured-retry-limit")
(assert-eq (cleanup-completed? error-state) true "error-handler-performs-necessary-cleanup")
```

### Multi-Assertion Test Standards

For tests with multiple assertions, use **progressive context** with clear setup and verification phases:

```vex
(deftest "user-registration-creates-complete-profile"
  (do
    ;; Setup phase
    (def user-data {"name" "Alice" "email" "alice@test.com"})
    (def result (register-user user-data))
    
    ;; Verify creation success
    (assert-eq (get result "success") true "registration-completes-successfully")
    
    ;; Verify data integrity
    (assert-eq (get result "user-id") expected-id "registration-assigns-unique-identifier")
    (assert-eq (get result "status") "active" "new-user-starts-with-active-status")
    
    ;; Verify side effects
    (assert-eq (user-exists? "alice@test.com") true "registration-persists-user-in-database")
    (assert-eq (email-sent? "alice@test.com") true "registration-triggers-welcome-email")))
```

## Framework Implementation Standards

### Test Framework Output (Clean)

**Required Updates to stdlib/vex/test/test.vx**:
```vex
;; Clean assertion macro without emojis
(macro assert-eq [actual expected msg]
  (do
    (if (= actual expected)
      (do
        (fmt/Println "PASS:" msg))
      (do
        (fmt/Println "FAIL:" msg)
        (fmt/Println "  Expected:" expected)
        (fmt/Println "  Actual:" actual)))))

;; Clean test definition macro
(macro deftest [name body]
  (do
     (fmt/Printf "\nRunning test: %s\n" name)
     body))
```

### Verbosity Levels

#### Normal Output (Clean and Minimal)

```
Running Vex tests in ./src (found 8 test files)

PASS: user_service_test.vx (125ms)
FAIL: payment_test.vx (89ms)
PASS: api_endpoints_test.vx (234ms)

Test Summary
=======================
Total: 8 tests
Passed: 6
Failed: 2
Duration: 1.2s
Success rate: 75.0%
```

#### Verbose Output (Detailed but Clean)

```
Running Vex tests in ./src (found 8 test files)
Pattern filter: none
Timeout: 30s

Executing: user_service_test.vx (1/8)
  Running test: user-authentication-validates-credentials
  PASS: valid-email-format-passes-validation
  PASS: invalid-email-format-fails-validation
PASS: user_service_test.vx (125ms)

Executing: payment_test.vx (2/8)
  Running test: payment-processing-handles-valid-cards
  FAIL: valid-payment-processes-successfully
    Expected: "success"
    Actual: "error"
FAIL: payment_test.vx (89ms) - assertion failures

Test Summary
=======================
Total: 8 tests
Passed: 6
Failed: 2
Duration: 1.2s
Success rate: 75.0%

Next steps: Use './vex test -verbose' to see detailed error information
```

## AI Code Generation Guidelines

### Template for AI Models

**Test Structure Template**:
```vex
(deftest "{domain}-{action}-{expected-outcome}"
  (do
    ;; Setup with meaningful variable names (kebab-case)
    (def test-data {meaningful-data})
    
    ;; Action under test
    (def result (function-under-test test-data))
    
    ;; Assertions with clear context
    (assert-eq result expected-value "{action}-{context}-{expectation}")))
```

### AI Generation Patterns

1. **Start with the domain**: Identify what system/feature you're testing
2. **Use action verbs**: validates, calculates, processes, handles, returns
3. **Be specific about expectations**: Instead of "works correctly", say "returns-valid-json"
4. **Follow kebab-case convention**: All test and assertion messages use dashes
5. **Include edge cases**: Always consider boundary conditions and error states

### Quality Metrics Alignment

These standards support Vex's enhanced coverage analysis quality scoring:

- **Assertion Density**: 1-10 meaningful assertions per test
- **Edge Case Coverage**: Explicit edge-case and error-case test naming
- **Method Diversity**: Clear patterns for different assertion types
- **Naming Quality**: Descriptive test names that explain business value

### Example High-Quality Test

```vex
(deftest "payment-processing-handles-multiple-scenarios"
  (do
    ;; Valid payment scenario
    (assert-eq (process-payment valid-card 100.00) "success" "valid-payment-processes-successfully")
    
    ;; Edge case: minimum amount
    (assert-eq (process-payment valid-card 0.01) "success" "minimum-payment-amount-processes-correctly")
    
    ;; Error case: insufficient funds
    (assert-eq (process-payment empty-card 100.00) "insufficient-funds" "empty-card-returns-appropriate-error")
    
    ;; Edge case: maximum amount
    (assert-eq (process-payment premium-card 99999.99) "success" "maximum-payment-amount-processes-correctly")))
```

## Implementation Checklist

To implement these standards across the Vex testing framework:

### Framework Updates Required

- [ ] **cmd/vex-transpiler/main.go**: Remove all emojis from framework messages
- [ ] **stdlib/vex/test/test.vx**: Update assertion macros to use clean text output
- [ ] **Test discovery messages**: Implement clean header format
- [ ] **Test result messages**: Implement clean status reporting
- [ ] **Summary messages**: Implement clean summary format
- [ ] **Coverage reports**: Implement clean coverage formatting

### Documentation Updates Required

- [ ] **docs/testing-guide.md**: Update examples to show clean text output
- [ ] **docs/ai-quick-reference.md**: Add test message pattern examples
- [ ] **docs/getting-started.md**: Update test examples with clean messages
- [ ] **All example files**: Ensure test messages follow standards

### Quality Assurance

- [ ] **CI/CD compatibility**: Verify all messages work in automated environments
- [ ] **AI model testing**: Validate that AI can generate compliant test messages
- [ ] **Professional tooling**: Ensure compatibility with IDEs and development tools
- [ ] **Internationalization**: Verify clean text works across different locales

## Compliance Examples

### Compliant Test File

```vex
;; user-authentication_test.vx

(import ["fmt" "user-service"])

(deftest "user-authentication-validates-correct-credentials"
  (do
    (def user-credentials {"username" "alice" "password" "secure123"})
    (def auth-result (authenticate-user user-credentials))
    (assert-eq (get auth-result "success") true "valid-credentials-authenticate-successfully")
    (assert-eq (get auth-result "user-id") expected-user-id "authentication-returns-correct-user-identifier")))

(deftest "edge-case-authentication-handles-empty-credentials"
  (do
    (def empty-credentials {"username" "" "password" ""})
    (def auth-result (authenticate-user empty-credentials))
    (assert-eq (get auth-result "success") false "empty-credentials-fail-authentication")
    (assert-eq (get auth-result "error") "INVALID_CREDENTIALS" "empty-credentials-return-appropriate-error-code")))

(deftest "error-case-authentication-handles-database-failure"
  (do
    (def valid-credentials {"username" "alice" "password" "secure123"})
    (def mock-db-failure (simulate-database-error))
    (def auth-result (authenticate-user valid-credentials))
    (assert-eq (get auth-result "success") false "database-failure-prevents-authentication")
    (assert-eq (get auth-result "error") "SERVICE_UNAVAILABLE" "database-error-returns-service-unavailable")))
```

These standards ensure that Vex testing maintains professional quality while supporting AI code generation and providing clear, actionable feedback to developers.
