package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// --------------------------------------------------------------------------
// Engine helpers
// --------------------------------------------------------------------------

// newParseAuditEngine creates a default Engine for parse audit tests.
func newParseAuditEngine() *liquid.Engine {
	return liquid.NewEngine()
}

// --------------------------------------------------------------------------
// ParseResult helpers
// --------------------------------------------------------------------------

// parseAudit calls ParseStringAudit on a fresh engine and returns the result.
func parseAudit(src string) *liquid.ParseResult {
	return newParseAuditEngine().ParseStringAudit(src)
}

// parseAuditWith calls ParseStringAudit on the provided engine.
func parseAuditWith(eng *liquid.Engine, src string) *liquid.ParseResult {
	return eng.ParseStringAudit(src)
}

// parseAuditBytes calls ParseTemplateAudit ([]byte variant) on a fresh engine.
func parseAuditBytes(src string) *liquid.ParseResult {
	return newParseAuditEngine().ParseTemplateAudit([]byte(src))
}

// --------------------------------------------------------------------------
// Assertion helpers for ParseResult
// --------------------------------------------------------------------------

// assertParseResultNonNil asserts the ParseResult itself is not nil.
func assertParseResultNonNil(t *testing.T, r *liquid.ParseResult, label string) {
	t.Helper()
	if r == nil {
		t.Fatalf("%s: ParseResult is nil", label)
	}
}

// assertTemplateNonNil asserts result.Template is not nil (parse succeeded).
func assertTemplateNonNil(t *testing.T, r *liquid.ParseResult, label string) {
	t.Helper()
	assertParseResultNonNil(t, r, label)
	if r.Template == nil {
		t.Fatalf("%s: Template is nil, want non-nil (parse should have succeeded)", label)
	}
}

// assertTemplateNil asserts result.Template is nil (fatal parse error).
func assertTemplateNil(t *testing.T, r *liquid.ParseResult, label string) {
	t.Helper()
	assertParseResultNonNil(t, r, label)
	if r.Template != nil {
		t.Fatalf("%s: Template is non-nil, want nil (expected fatal parse error)", label)
	}
}

// assertDiagsNonNil asserts result.Diagnostics is not nil (always non-nil per contract).
func assertDiagsNonNil(t *testing.T, r *liquid.ParseResult, label string) {
	t.Helper()
	assertParseResultNonNil(t, r, label)
	if r.Diagnostics == nil {
		t.Fatalf("%s: Diagnostics is nil, want non-nil empty slice", label)
	}
}

// assertNoParseDiags asserts the result has no diagnostics.
func assertNoParseDiags(t *testing.T, r *liquid.ParseResult, label string) {
	t.Helper()
	assertDiagsNonNil(t, r, label)
	if len(r.Diagnostics) != 0 {
		t.Errorf("%s: expected 0 diagnostics, got %d: %v", label, len(r.Diagnostics), r.Diagnostics)
	}
}

// assertParseDiagCount asserts the exact number of diagnostics.
func assertParseDiagCount(t *testing.T, r *liquid.ParseResult, want int, label string) {
	t.Helper()
	assertDiagsNonNil(t, r, label)
	if len(r.Diagnostics) != want {
		t.Errorf("%s: len(Diagnostics)=%d, want %d (codes: %v)",
			label, len(r.Diagnostics), want, parseDiagCodes(r.Diagnostics))
	}
}

// parseDiagCodes extracts the Code field of each diagnostic.
func parseDiagCodes(diags []liquid.Diagnostic) []string {
	codes := make([]string, len(diags))
	for i, d := range diags {
		codes[i] = d.Code
	}
	return codes
}

// firstParseDiag returns the first Diagnostic with the given code, or nil.
func firstParseDiag(r *liquid.ParseResult, code string) *liquid.Diagnostic {
	for i := range r.Diagnostics {
		if r.Diagnostics[i].Code == code {
			return &r.Diagnostics[i]
		}
	}
	return nil
}

// allParseDiags returns all Diagnostics with the given code.
func allParseDiags(r *liquid.ParseResult, code string) []liquid.Diagnostic {
	var out []liquid.Diagnostic
	for _, d := range r.Diagnostics {
		if d.Code == code {
			out = append(out, d)
		}
	}
	return out
}

// requireParseDiag returns the first Diagnostic with the given code, failing the
// test if none is found.
func requireParseDiag(t *testing.T, r *liquid.ParseResult, code string) liquid.Diagnostic {
	t.Helper()
	d := firstParseDiag(r, code)
	if d == nil {
		t.Fatalf("expected diagnostic with code=%q, got codes: %v", code, parseDiagCodes(r.Diagnostics))
	}
	return *d
}

// assertDiagField checks a string field of a Diagnostic.
func assertDiagField(t *testing.T, got, want, field, code string) {
	t.Helper()
	if got != want {
		t.Errorf("%s diagnostic: %s=%q, want %q", code, field, got, want)
	}
}

// assertDiagContains checks that a string field of a Diagnostic contains a substring.
func assertDiagContains(t *testing.T, field, value, substr, code string) {
	t.Helper()
	if !containsSubstr(value, substr) {
		t.Errorf("%s diagnostic: %s=%q, want it to contain %q", code, field, value, substr)
	}
}

// containsSubstr is a local reimplementation to avoid importing strings in every test file.
func containsSubstr(s, substr string) bool {
	if substr == "" {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// assertRangeValidParse checks Range.Start.Line >= 1 and Column >= 1.
func assertRangeValidParse(t *testing.T, r liquid.Range, label string) {
	t.Helper()
	if r.Start.Line < 1 {
		t.Errorf("%s: Range.Start.Line=%d, want >= 1", label, r.Start.Line)
	}
	if r.Start.Column < 1 {
		t.Errorf("%s: Range.Start.Column=%d, want >= 1", label, r.Start.Column)
	}
}

// assertRangeEndAfterStart checks that End is at or after Start.
func assertRangeEndAfterStart(t *testing.T, r liquid.Range, label string) {
	t.Helper()
	assertRangeValidParse(t, r, label)
	endOK := r.End.Line > r.Start.Line ||
		(r.End.Line == r.Start.Line && r.End.Column >= r.Start.Column)
	if !endOK {
		t.Errorf("%s: Range.End (%+v) is before Range.Start (%+v)", label, r.End, r.Start)
	}
}
