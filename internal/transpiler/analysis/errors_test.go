package analysis

import (
	"strings"
	"testing"
)

// Moved from errors_extra_test.go
func TestErrorReporter_StringersAndCounts(t *testing.T) {
    er := NewErrorReporter()
    if er.HasErrors() { t.Fatalf("expected no errors initially") }
    er.ReportWarning(2, 3, "be careful")
    er.ReportTypedError(1, 2, "boom", TypeError)
    if !er.HasErrors() || er.GetErrorCount() != 1 || er.GetWarningCount() != 1 {
        t.Fatalf("counts mismatch: errors=%d warnings=%d", er.GetErrorCount(), er.GetWarningCount())
    }
    if er.FormatErrors() == "" || er.FormatWarnings() == "" {
        t.Fatalf("expected formatted outputs")
    }
    er.Clear()
    if er.HasErrors() || er.GetErrorCount() != 0 || er.GetWarningCount() != 0 {
        t.Fatalf("clear failed")
    }
}


func TestErrorType_String(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		want      string
	}{
		{SyntaxError, "Syntax Error"},
		{SemanticError, "Semantic Error"},
		{MacroError, "Macro Error"},
		{TypeError, "Type Error"},
		{ErrorType(999), "Unknown Error"}, // Test unknown error type
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.errorType.String()
			if got != tt.want {
				t.Errorf("ErrorType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompilerError_String(t *testing.T) {
	err := CompilerError{
		Line:    10,
		Column:  5,
		Message: "test error",
		Type:    SyntaxError,
	}
	
    expected := "10:5: error: test error"
	got := err.String()
	
	if got != expected {
		t.Errorf("CompilerError.String() = %v, want %v", got, expected)
	}
	
	// Test with file
	errWithFile := CompilerError{
		File:    "test.vx",
		Line:    5,
		Column:  2,
		Message: "file error",
		Type:    TypeError,
	}
	
	expectedWithFile := "test.vx:5:2: error: file error"
	gotWithFile := errWithFile.String()
	
	if gotWithFile != expectedWithFile {
		t.Errorf("CompilerError.String() with file = %v, want %v", gotWithFile, expectedWithFile)
	}
}

func TestNewErrorReporter(t *testing.T) {
	reporter := NewErrorReporter()
	
	if reporter == nil {
		t.Fatal("NewErrorReporter should return non-nil reporter")
	}
	
	if reporter.errors == nil {
		t.Error("ErrorReporter should initialize errors slice")
	}
	
	if reporter.warnings == nil {
		t.Error("ErrorReporter should initialize warnings slice")
	}
	
	if reporter.HasErrors() {
		t.Error("New ErrorReporter should not have errors")
	}
}

func TestErrorReporter_ReportError(t *testing.T) {
	reporter := NewErrorReporter()
	
	reporter.ReportError(10, 5, "test error")
	
	if !reporter.HasErrors() {
		t.Error("Reporter should have errors after ReportError")
	}
	
	errors := reporter.GetErrors()
	if len(errors) != 1 {
		t.Errorf("Should have 1 error, got %v", len(errors))
	}
	
	err := errors[0]
	if err.Line != 10 {
		t.Errorf("Error line = %v, want %v", err.Line, 10)
	}
	
	if err.Column != 5 {
		t.Errorf("Error column = %v, want %v", err.Column, 5)
	}
	
	if err.Message != "test error" {
		t.Errorf("Error message = %v, want %v", err.Message, "test error")
	}
	
	if err.Type != SemanticError {
		t.Errorf("Error type = %v, want %v", err.Type, SemanticError)
	}
}

func TestErrorReporter_ReportWarning(t *testing.T) {
	reporter := NewErrorReporter()
	
	reporter.ReportWarning(15, 3, "test warning")
	
	warnings := reporter.GetWarnings()
	if len(warnings) != 1 {
		t.Errorf("Should have 1 warning, got %v", len(warnings))
	}
	
	warning := warnings[0]
	if warning.Line != 15 {
		t.Errorf("Warning line = %v, want %v", warning.Line, 15)
	}
	
	if warning.Column != 3 {
		t.Errorf("Warning column = %v, want %v", warning.Column, 3)
	}
	
	if warning.Message != "test warning" {
		t.Errorf("Warning message = %v, want %v", warning.Message, "test warning")
	}
}

func TestErrorReporter_ReportTypedError(t *testing.T) {
	reporter := NewErrorReporter()
	
	reporter.ReportTypedError(20, 8, "macro error", MacroError)
	
	errors := reporter.GetErrors()
	if len(errors) != 1 {
		t.Errorf("Should have 1 error, got %v", len(errors))
	}
	
	err := errors[0]
	if err.Type != MacroError {
		t.Errorf("Error type = %v, want %v", err.Type, MacroError)
	}
	
	if err.Message != "macro error" {
		t.Errorf("Error message = %v, want %v", err.Message, "macro error")
	}
}

func TestErrorReporter_HasErrors(t *testing.T) {
	reporter := NewErrorReporter()
	
	if reporter.HasErrors() {
		t.Error("New reporter should not have errors")
	}
	
	reporter.ReportError(1, 1, "error")
	
	if !reporter.HasErrors() {
		t.Error("Reporter should have errors after reporting")
	}
	
	// Warnings should not affect HasErrors
	reporter2 := NewErrorReporter()
	reporter2.ReportWarning(1, 1, "warning")
	
	if reporter2.HasErrors() {
		t.Error("Reporter with only warnings should not have errors")
	}
}

func TestErrorReporter_GetErrors_Sorting(t *testing.T) {
	reporter := NewErrorReporter()
	
	// Add errors out of order
	reporter.ReportError(20, 5, "error 2")
	reporter.ReportError(10, 3, "error 1")
	reporter.ReportError(10, 8, "error 1.2")
	reporter.ReportError(30, 1, "error 3")
	
	errors := reporter.GetErrors()
	
	// Should be sorted by line, then by column
	expected := []struct {
		line   int
		column int
	}{
		{10, 3},
		{10, 8},
		{20, 5},
		{30, 1},
	}
	
	for i, err := range errors {
		if err.Line != expected[i].line {
			t.Errorf("Error %d line = %v, want %v", i, err.Line, expected[i].line)
		}
		if err.Column != expected[i].column {
			t.Errorf("Error %d column = %v, want %v", i, err.Column, expected[i].column)
		}
	}
}

func TestErrorReporter_GetWarnings_Sorting(t *testing.T) {
	reporter := NewErrorReporter()
	
	// Add warnings out of order
	reporter.ReportWarning(25, 2, "warning 2")
	reporter.ReportWarning(5, 1, "warning 1")
	reporter.ReportWarning(5, 10, "warning 1.2")
	
	warnings := reporter.GetWarnings()
	
	// Should be sorted by line, then by column
	expected := []struct {
		line   int
		column int
	}{
		{5, 1},
		{5, 10},
		{25, 2},
	}
	
	for i, warning := range warnings {
		if warning.Line != expected[i].line {
			t.Errorf("Warning %d line = %v, want %v", i, warning.Line, expected[i].line)
		}
		if warning.Column != expected[i].column {
			t.Errorf("Warning %d column = %v, want %v", i, warning.Column, expected[i].column)
		}
	}
}

func TestErrorReporter_GetErrorCount(t *testing.T) {
	reporter := NewErrorReporter()
	
	if reporter.GetErrorCount() != 0 {
		t.Errorf("Initial error count should be 0, got %v", reporter.GetErrorCount())
	}
	
	reporter.ReportError(1, 1, "error 1")
	reporter.ReportError(2, 1, "error 2")
	
	if reporter.GetErrorCount() != 2 {
		t.Errorf("Error count should be 2, got %v", reporter.GetErrorCount())
	}
	
	// Warnings should not affect error count
	reporter.ReportWarning(3, 1, "warning")
	
	if reporter.GetErrorCount() != 2 {
		t.Errorf("Error count should still be 2 after adding warning, got %v", reporter.GetErrorCount())
	}
}

func TestErrorReporter_GetWarningCount(t *testing.T) {
	reporter := NewErrorReporter()
	
	if reporter.GetWarningCount() != 0 {
		t.Errorf("Initial warning count should be 0, got %v", reporter.GetWarningCount())
	}
	
	reporter.ReportWarning(1, 1, "warning 1")
	reporter.ReportWarning(2, 1, "warning 2")
	
	if reporter.GetWarningCount() != 2 {
		t.Errorf("Warning count should be 2, got %v", reporter.GetWarningCount())
	}
	
	// Errors should not affect warning count
	reporter.ReportError(3, 1, "error")
	
	if reporter.GetWarningCount() != 2 {
		t.Errorf("Warning count should still be 2 after adding error, got %v", reporter.GetWarningCount())
	}
}

func TestErrorReporter_Clear(t *testing.T) {
	reporter := NewErrorReporter()
	
	// Add some errors and warnings
	reporter.ReportError(1, 1, "error")
	reporter.ReportWarning(2, 1, "warning")
	
	if !reporter.HasErrors() {
		t.Error("Should have errors before clear")
	}
	
	if reporter.GetWarningCount() == 0 {
		t.Error("Should have warnings before clear")
	}
	
	// Clear
	reporter.Clear()
	
	if reporter.HasErrors() {
		t.Error("Should not have errors after clear")
	}
	
	if reporter.GetErrorCount() != 0 {
		t.Errorf("Error count should be 0 after clear, got %v", reporter.GetErrorCount())
	}
	
	if reporter.GetWarningCount() != 0 {
		t.Errorf("Warning count should be 0 after clear, got %v", reporter.GetWarningCount())
	}
}

func TestErrorReporter_FormatErrors(t *testing.T) {
	reporter := NewErrorReporter()
	
	// Test empty case
	formatted := reporter.FormatErrors()
	if formatted != "" {
		t.Errorf("FormatErrors should return empty string for no errors, got: %v", formatted)
	}
	
	// Add errors
	reporter.ReportError(10, 5, "first error")
	reporter.ReportTypedError(20, 3, "second error", TypeError)
	
	formatted = reporter.FormatErrors()
	
	// Should contain both errors
	if !strings.Contains(formatted, "first error") {
		t.Error("Formatted errors should contain first error")
	}
	
	if !strings.Contains(formatted, "second error") {
		t.Error("Formatted errors should contain second error")
	}
	
    if !strings.Contains(formatted, "10:5: error:") {
		t.Error("Formatted errors should contain line info for first error")
	}
	
    if !strings.Contains(formatted, "20:3: error:") {
		t.Error("Formatted errors should contain line info for second error")
	}
}

func TestErrorReporter_FormatWarnings(t *testing.T) {
	reporter := NewErrorReporter()
	
	// Test empty case
	formatted := reporter.FormatWarnings()
	if formatted != "" {
		t.Errorf("FormatWarnings should return empty string for no warnings, got: %v", formatted)
	}
	
	// Add warnings
	reporter.ReportWarning(15, 8, "first warning")
	reporter.ReportWarning(25, 2, "second warning")
	
	formatted = reporter.FormatWarnings()
	
	// Should contain both warnings
	if !strings.Contains(formatted, "first warning") {
		t.Error("Formatted warnings should contain first warning")
	}
	
	if !strings.Contains(formatted, "second warning") {
		t.Error("Formatted warnings should contain second warning")
	}
	
    if !strings.Contains(strings.ToLower(formatted), "warning:") {
        t.Error("Formatted warnings should contain 'warning:' prefix")
    }
	
    if !strings.Contains(formatted, "15:8:") {
		t.Error("Formatted warnings should contain line info")
	}
}

func TestErrorReporter_MultipleOperations(t *testing.T) {
	reporter := NewErrorReporter()
	
	// Mix of operations
	reporter.ReportError(5, 1, "error 1")
	reporter.ReportWarning(6, 1, "warning 1")
	reporter.ReportTypedError(7, 1, "syntax error", SyntaxError)
	reporter.ReportError(8, 1, "error 2")
	
	if reporter.GetErrorCount() != 3 {
		t.Errorf("Should have 3 errors, got %v", reporter.GetErrorCount())
	}
	
	if reporter.GetWarningCount() != 1 {
		t.Errorf("Should have 1 warning, got %v", reporter.GetWarningCount())
	}
	
	// Check that error types are preserved
	errors := reporter.GetErrors()
	syntaxErrorFound := false
	for _, err := range errors {
		if err.Type == SyntaxError && err.Message == "syntax error" {
			syntaxErrorFound = true
			break
		}
	}
	
	if !syntaxErrorFound {
		t.Error("Should find the syntax error with correct type")
	}
}