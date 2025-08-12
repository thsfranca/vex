package analysis

import (
	"fmt"
	"sort"
)

// Adapter to accept structured diagnostics without breaking existing API
// We avoid importing diagnostics in the type definitions to keep analysis decoupled.
// A thin shim function is provided to map a pre-rendered body into the reporter.

// ErrorType represents the type of error
type ErrorType int

const (
	SyntaxError ErrorType = iota
	SemanticError
	MacroError
	TypeError
)

func (et ErrorType) String() string {
	switch et {
	case SyntaxError:
		return "Syntax Error"
	case SemanticError:
		return "Semantic Error"
	case MacroError:
		return "Macro Error"
	case TypeError:
		return "Type Error"
	default:
		return "Unknown Error"
	}
}

// CompilerError represents a compilation error
type CompilerError struct {
	File    string
	Line    int
	Column  int
	Message string
	Type    ErrorType
}

func (e *CompilerError) String() string {
	// Go-style: file:line:col: error: message (file optional)
	if e.File != "" {
		return fmt.Sprintf("%s:%d:%d: error: %s", e.File, e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("%d:%d: error: %s", e.Line, e.Column, e.Message)
}

// ErrorReporterImpl implements the ErrorReporter interface
type ErrorReporterImpl struct {
	errors   []CompilerError
	warnings []CompilerError
}

// NewErrorReporter creates a new error reporter
func NewErrorReporter() *ErrorReporterImpl {
	return &ErrorReporterImpl{
		errors:   make([]CompilerError, 0),
		warnings: make([]CompilerError, 0),
	}
}

// ReportError reports a compilation error
func (er *ErrorReporterImpl) ReportError(line, column int, message string) {
	er.ReportTypedError(line, column, message, SemanticError)
}

// ReportWarning reports a compilation warning
func (er *ErrorReporterImpl) ReportWarning(line, column int, message string) {
	warning := CompilerError{
		File:    "",
		Line:    line,
		Column:  column,
		Message: message,
		Type:    SemanticError, // Warnings are semantic by default
	}
	er.warnings = append(er.warnings, warning)
}

// ReportTypedError reports an error with a specific type
func (er *ErrorReporterImpl) ReportTypedError(line, column int, message string, errorType ErrorType) {
	error := CompilerError{
		File:    "",
		Line:    line,
		Column:  column,
		Message: message,
		Type:    errorType,
	}
	er.errors = append(er.errors, error)
}

// ReportDiagnosticBody reports a diagnostic by taking the already-rendered body
// e.g., "[VEX-TYP-IF-MISMATCH]: branch types differ\nExpected: ...\nGot: ..."
// Callers should pass appropriate error type and position.
func (er *ErrorReporterImpl) ReportDiagnosticBody(line, column int, body string, errorType ErrorType) {
	er.ReportTypedError(line, column, body, errorType)
}

// HasErrors returns true if any errors have been reported
func (er *ErrorReporterImpl) HasErrors() bool {
	return len(er.errors) > 0
}

// GetErrors returns all reported errors
func (er *ErrorReporterImpl) GetErrors() []CompilerError {
	// Sort errors by line number for consistent output
	sorted := make([]CompilerError, len(er.errors))
	copy(sorted, er.errors)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Line == sorted[j].Line {
			return sorted[i].Column < sorted[j].Column
		}
		return sorted[i].Line < sorted[j].Line
	})

	return sorted
}

// GetWarnings returns all reported warnings
func (er *ErrorReporterImpl) GetWarnings() []CompilerError {
	// Sort warnings by line number for consistent output
	sorted := make([]CompilerError, len(er.warnings))
	copy(sorted, er.warnings)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Line == sorted[j].Line {
			return sorted[i].Column < sorted[j].Column
		}
		return sorted[i].Line < sorted[j].Line
	})

	return sorted
}

// GetErrorCount returns the number of errors
func (er *ErrorReporterImpl) GetErrorCount() int {
	return len(er.errors)
}

// GetWarningCount returns the number of warnings
func (er *ErrorReporterImpl) GetWarningCount() int {
	return len(er.warnings)
}

// Clear removes all errors and warnings
func (er *ErrorReporterImpl) Clear() {
	er.errors = er.errors[:0]
	er.warnings = er.warnings[:0]
}

// FormatErrors returns a formatted string of all errors
func (er *ErrorReporterImpl) FormatErrors() string {
	if !er.HasErrors() {
		return ""
	}

	result := ""
	for _, err := range er.GetErrors() {
		result += err.String() + "\n"
	}

	return result
}

// FormatWarnings returns a formatted string of all warnings
func (er *ErrorReporterImpl) FormatWarnings() string {
	if len(er.warnings) == 0 {
		return ""
	}

	result := ""
	for _, warning := range er.GetWarnings() {
		// lowercase 'warning:' after position for consistency with Go style
		if warning.File != "" {
			result += fmt.Sprintf("%s:%d:%d: warning: %s\n", warning.File, warning.Line, warning.Column, warning.Message)
		} else {
			result += fmt.Sprintf("%d:%d: warning: %s\n", warning.Line, warning.Column, warning.Message)
		}
	}

	return result
}
