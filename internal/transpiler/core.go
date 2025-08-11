// Package transpiler provides a minimal Vex to Go transpiler
package transpiler

import (
	"fmt"
)

// Transpiler defines high-level entrypoints for Vex->Go transpilation.
type Transpiler interface {
	TranspileFromInput(input string) (string, error)
	TranspileFromFile(filename string) (string, error)
	GetDetectedModules() map[string]string
}

// New creates a new transpiler instance using default configuration.
func New() Transpiler {
	transpiler, err := NewVexTranspiler()
	if err != nil {
		// In practice, this shouldn't happen with default config
		panic(fmt.Sprintf("Failed to create transpiler: %v", err))
	}
	return transpiler
}

// NewWithDebug creates a new transpiler instance with debug options (placeholder).
func NewWithDebug() Transpiler {
	// For now, return the same as New() - debug mode will be configurable later
	return New()
}