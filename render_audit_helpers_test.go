package liquid_test

import (
	"fmt"
	"testing"

	"github.com/osteele/liquid"
)

// --------------------------------------------------------------------------
// Parse / render helpers
// --------------------------------------------------------------------------

// mustParseAudit parses a template string, failing the test on any error.
func mustParseAudit(t *testing.T, src string) *liquid.Template {
	t.Helper()
	tpl, err := newAuditEngine().ParseString(src)
	if err != nil {
		t.Fatalf("ParseString(%q): %v", src, err)
	}
	return tpl
}

// mustParseAuditWith is like mustParseAudit but uses a caller-provided engine.
func mustParseAuditWith(t *testing.T, eng *liquid.Engine, src string) *liquid.Template {
	t.Helper()
	tpl, err := eng.ParseString(src)
	if err != nil {
		t.Fatalf("ParseString(%q): %v", src, err)
	}
	return tpl
}

// auditOK renders with audit and asserts no AuditError is returned.
func auditOK(t *testing.T, tpl *liquid.Template, vars liquid.Bindings, opts liquid.AuditOptions, renderOpts ...liquid.RenderOption) *liquid.AuditResult {
	t.Helper()
	result, ae := tpl.RenderAudit(vars, opts, renderOpts...)
	if result == nil {
		t.Fatal("RenderAudit returned nil result")
	}
	if ae != nil {
		t.Fatalf("RenderAudit returned unexpected AuditError: %v", ae)
	}
	return result
}

// auditErr renders with audit and asserts that an AuditError is returned.
func auditErr(t *testing.T, tpl *liquid.Template, vars liquid.Bindings, opts liquid.AuditOptions, renderOpts ...liquid.RenderOption) (*liquid.AuditResult, *liquid.AuditError) {
	t.Helper()
	result, ae := tpl.RenderAudit(vars, opts, renderOpts...)
	if result == nil {
		t.Fatal("RenderAudit returned nil result (must be non-nil even on error)")
	}
	if ae == nil {
		t.Fatal("RenderAudit: expected AuditError but got nil")
	}
	return result, ae
}

// --------------------------------------------------------------------------
// Expression finders
// --------------------------------------------------------------------------

// firstExpr returns the first Expression with matching Kind, or nil.
func firstExpr(exprs []liquid.Expression, kind liquid.ExpressionKind) *liquid.Expression {
	for i := range exprs {
		if exprs[i].Kind == kind {
			return &exprs[i]
		}
	}
	return nil
}

// allExprs returns all Expressions with matching Kind.
func allExprs(exprs []liquid.Expression, kind liquid.ExpressionKind) []liquid.Expression {
	var out []liquid.Expression
	for _, e := range exprs {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// nthExpr returns the n-th (0-based) Expression with matching Kind, or nil.
func nthExpr(exprs []liquid.Expression, kind liquid.ExpressionKind, n int) *liquid.Expression {
	idx := 0
	for i := range exprs {
		if exprs[i].Kind == kind {
			if idx == n {
				return &exprs[i]
			}
			idx++
		}
	}
	return nil
}

// --------------------------------------------------------------------------
// Diagnostic finder
// --------------------------------------------------------------------------

// firstDiag returns the first Diagnostic with matching Code, or nil.
func firstDiag(diags []liquid.Diagnostic, code string) *liquid.Diagnostic {
	for i := range diags {
		if diags[i].Code == code {
			return &diags[i]
		}
	}
	return nil
}

// allDiags returns all Diagnostics with matching Code.
func allDiags(diags []liquid.Diagnostic, code string) []liquid.Diagnostic {
	var out []liquid.Diagnostic
	for _, d := range diags {
		if d.Code == code {
			out = append(out, d)
		}
	}
	return out
}

// --------------------------------------------------------------------------
// Assertion helpers
// --------------------------------------------------------------------------

// assertOutput checks result.Output equals want.
func assertOutput(t *testing.T, result *liquid.AuditResult, want string) {
	t.Helper()
	if result.Output != want {
		t.Errorf("Output=%q, want %q", result.Output, want)
	}
}

// assertExprCount checks the total number of Expressions.
func assertExprCount(t *testing.T, result *liquid.AuditResult, want int) {
	t.Helper()
	if len(result.Expressions) != want {
		t.Errorf("len(Expressions)=%d, want %d", len(result.Expressions), want)
	}
}

// assertNoDiags asserts the result has no Diagnostics.
func assertNoDiags(t *testing.T, result *liquid.AuditResult) {
	t.Helper()
	if len(result.Diagnostics) > 0 {
		t.Errorf("expected no diagnostics, got %d: %v", len(result.Diagnostics), result.Diagnostics)
	}
}

// assertRangeValid checks that Range.Start.Line >= 1.
func assertRangeValid(t *testing.T, r liquid.Range, label string) {
	t.Helper()
	if r.Start.Line < 1 {
		t.Errorf("%s: Range.Start.Line=%d, want >= 1", label, r.Start.Line)
	}
	if r.Start.Column < 1 {
		t.Errorf("%s: Range.Start.Column=%d, want >= 1", label, r.Start.Column)
	}
}

// assertRangeSpan checks that Range.End > Range.Start (valid non-zero span).
func assertRangeSpan(t *testing.T, r liquid.Range, label string) {
	t.Helper()
	assertRangeValid(t, r, label)
	endIsAfter := r.End.Line > r.Start.Line ||
		(r.End.Line == r.Start.Line && r.End.Column > r.Start.Column)
	if !endIsAfter {
		t.Errorf("%s: Range.End (%v) is not after Range.Start (%v)", label, r.End, r.Start)
	}
}

// --------------------------------------------------------------------------
// Numeric formatting helpers
// --------------------------------------------------------------------------

// sprintVal converts any value to its default string representation.
// Useful for comparing numeric values without caring about int vs float64.
func sprintVal(v any) string {
	return fmt.Sprintf("%v", v)
}

// iptr returns a pointer to the given int value.
func iptr(i int) *int { return &i }
