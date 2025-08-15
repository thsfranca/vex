// Package transpiler provides a minimal Vex to Go transpiler
package transpiler

// Transpiler defines high-level entrypoints for Vex->Go transpilation.
type Transpiler interface {
	TranspileFromInput(input string) (string, error)
	GetDetectedModules() map[string]string
}




