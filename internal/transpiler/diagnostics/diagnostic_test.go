package diagnostics

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDiagnostic_RenderText_JSON_Body(t *testing.T) {
    d := New(CodeTypIfMismatch, SeverityError, "", 12, 34, map[string]any{
        "Expected": "type(then) == type(else)",
        "Got":      "number vs bool",
        "Offender": "if",
    }).WithSuggestion("align branch types")

    txt := d.RenderText()
    if !strings.Contains(txt, "[VEX-TYP-IF-MISMATCH]") {
        t.Fatalf("RenderText missing code: %s", txt)
    }
    if !strings.Contains(strings.ToLower(txt), "error:") {
        t.Fatalf("RenderText missing severity: %s", txt)
    }
    if !strings.Contains(txt, "Expected:") || !strings.Contains(txt, "Got:") || !strings.Contains(txt, "Offender:") {
        t.Fatalf("RenderText missing details: %s", txt)
    }

    body := d.RenderBody()
    if !strings.HasPrefix(body, "[VEX-TYP-IF-MISMATCH]:") {
        t.Fatalf("RenderBody prefix unexpected: %s", body)
    }

    // JSON must marshal and contain a non-empty message
    b, err := d.RenderJSON()
    if err != nil {
        t.Fatalf("RenderJSON error: %v", err)
    }
    var m map[string]any
    if err := json.Unmarshal(b, &m); err != nil {
        t.Fatalf("json.Unmarshal: %v", err)
    }
    if _, ok := m["message"]; !ok {
        t.Fatalf("JSON missing message field: %s", string(b))
    }
}

func TestDiagnostic_RenderHeaderVariants(t *testing.T) {
    // With file and location
    d := New(CodeDefArgs, SeverityWarning, "a.vx", 3, 7, nil)
    txt := d.RenderText()
    if !strings.HasPrefix(txt, "a.vx:3:7:") {
        t.Fatalf("header with file unexpected: %s", txt)
    }

    // Only line/col
    d2 := New(CodeFnArgs, SeverityError, "", 9, 2, nil)
    txt2 := d2.RenderText()
    if !strings.HasPrefix(txt2, "9:2:") {
        t.Fatalf("header without file unexpected: %s", txt2)
    }

    // Neither
    d3 := New(CodeSynEmpty, SeverityError, "", 0, 0, nil)
    txt3 := d3.RenderText()
    if !strings.HasPrefix(txt3, "error:") {
        t.Fatalf("default header unexpected: %s", txt3)
    }
}

func TestDiagnostic_RenderMessage_ExplicitAndCatalog(t *testing.T) {
    // Explicit message should be returned as-is
    d := New(CodeDefArgs, SeverityError, "", 0, 0, nil).WithMessage("explicit message")
    if d.RenderMessage() != "explicit message" {
        t.Fatalf("RenderMessage should prefer explicit message")
    }
    // Without explicit message, should use catalog
    d2 := New(CodeDefArgs, SeverityError, "", 0, 0, nil)
    if d2.RenderMessage() == "" {
        t.Fatalf("RenderMessage should use catalog when Message is empty")
    }
}


