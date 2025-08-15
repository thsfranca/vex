package analysis

import "testing"

// Test isCollectionHelper function directly
func TestIsCollectionHelper_Coverage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Collection operations that should return true
		{"first operation", "first", true},
		{"rest operation", "rest", true},
		{"cons operation", "cons", true},
		{"count operation", "count", true},
		{"empty check", "empty?", true},
		{"len operation", "len", true},
		{"get operation", "get", true},
		{"slice operation", "slice", true},
		{"append operation", "append", true},

		// Non-collection operations that should return false
		{"arithmetic operation", "+", false},
		{"comparison operation", "=", false},
		{"user function", "my-function", false},
		{"package function", "fmt/Println", false},
		{"unknown function", "unknown", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCollectionHelper(tt.input)

			if result != tt.expected {
				t.Errorf("isCollectionHelper(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
