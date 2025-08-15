package diagnostics

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Severity distinguishes errors vs warnings, if needed later
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Diagnostic is a structured compiler diagnostic
type Diagnostic struct {
	Code       Code           `json:"code"`
	Severity   Severity       `json:"severity"`
	File       string         `json:"file,omitempty"`
	Line       int            `json:"line,omitempty"`
	Column     int            `json:"col,omitempty"`
	Message    string         `json:"message"`
	Params     map[string]any `json:"params,omitempty"`
	Suggestion string         `json:"suggestion,omitempty"`
}

// New constructs a Diagnostic with defaults
func New(code Code, severity Severity, file string, line, col int, params map[string]any) Diagnostic {
	if params == nil {
		params = make(map[string]any)
	}
	return Diagnostic{
		Code:     code,
		Severity: severity,
		File:     file,
		Line:     line,
		Column:   col,
		Params:   params,
	}
}

// WithSuggestion sets the suggestion text
func (d Diagnostic) WithSuggestion(s string) Diagnostic {
	d.Suggestion = s
	return d
}

// WithMessage overrides the computed message (rarely needed)
func (d Diagnostic) WithMessage(msg string) Diagnostic {
	d.Message = msg
	return d
}

// RenderText produces Go-style text with code and optional detail lines
// Format: path:line:col: error: [CODE]: short-message\n[Expected: …]\n[Got: …]\n[Offender: …]\n[Suggestion: …]
func (d Diagnostic) RenderText() string {
	header := renderHeader(d)
	// Prefer explicit Message, otherwise use catalog template
	message := d.Message
	if strings.TrimSpace(message) == "" {
		message = renderFromCatalog(d)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s: %s: [%s]: %s", header, strings.ToLower(string(d.Severity)), d.Code, message))

	if exp, ok := d.Params["Expected"]; ok {
		b.WriteString("\nExpected: ")
		b.WriteString(fmt.Sprint(exp))
	}
	if got, ok := d.Params["Got"]; ok {
		b.WriteString("\nGot: ")
		b.WriteString(fmt.Sprint(got))
	}
	if off, ok := d.Params["Offender"]; ok {
		b.WriteString("\nOffender: ")
		b.WriteString(fmt.Sprint(off))
	}
	if s := strings.TrimSpace(d.Suggestion); s != "" {
		b.WriteString("\nSuggestion: ")
		b.WriteString(s)
	}
	return b.String()
}

// RenderJSON renders a machine-friendly representation
func (d Diagnostic) RenderJSON() ([]byte, error) {
	// Ensure Message exists for JSON as well
	if strings.TrimSpace(d.Message) == "" {
		d.Message = renderFromCatalog(d)
	}
	return json.Marshal(d)
}

func renderHeader(d Diagnostic) string {
	if d.File != "" {
		return fmt.Sprintf("%s:%d:%d", d.File, d.Line, d.Column)
	}
	if d.Line > 0 || d.Column > 0 {
		return fmt.Sprintf("%d:%d", d.Line, d.Column)
	}
	return "error"
}

// RenderMessage returns only the canonical short message for this code and params.
// If Message is already set, it is returned as-is; otherwise the catalog template is used.
func (d Diagnostic) RenderMessage() string {
	if strings.TrimSpace(d.Message) != "" {
		return d.Message
	}
	return renderFromCatalog(d)
}

// RenderBody renders the diagnostic without the location and severity header.
// Format: [CODE]: short-message + optional detail lines (Expected/Got/Offender/Suggestion).
func (d Diagnostic) RenderBody() string {
	message := d.RenderMessage()
	var b strings.Builder
	b.WriteString("[")
	b.WriteString(string(d.Code))
	b.WriteString("]: ")
	b.WriteString(message)

	if exp, ok := d.Params["Expected"]; ok {
		b.WriteString("\nExpected: ")
		b.WriteString(fmt.Sprint(exp))
	}
	if got, ok := d.Params["Got"]; ok {
		b.WriteString("\nGot: ")
		b.WriteString(fmt.Sprint(got))
	}
	if off, ok := d.Params["Offender"]; ok {
		b.WriteString("\nOffender: ")
		b.WriteString(fmt.Sprint(off))
	}
	if s := strings.TrimSpace(d.Suggestion); s != "" {
		b.WriteString("\nSuggestion: ")
		b.WriteString(s)
	}
	return b.String()
}
