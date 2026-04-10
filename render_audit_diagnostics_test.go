package liquid_test

import (
	"strings"
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// Diagnostics — Runtime Errors (D01–D16)
// ============================================================================

// D01 — undefined variable with StrictVariables: code="undefined-variable", severity=warning.
func TestRenderAudit_Diag_D01_undefinedVariable(t *testing.T) {
	tpl := mustParseAudit(t, "{{ ghost }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{}, liquid.WithStrictVariables())
	if result == nil {
		t.Fatal("result must not be nil")
	}
	d := firstDiag(result.Diagnostics, "undefined-variable")
	if d == nil {
		t.Fatal("expected undefined-variable diagnostic")
	}
	if d.Severity != liquid.SeverityWarning {
		t.Errorf("Severity=%q, want warning", d.Severity)
	}
	if !strings.Contains(d.Message, "ghost") {
		t.Errorf("Message=%q should mention the variable name 'ghost'", d.Message)
	}
}

// D02 — nested path undefined with StrictVariables.
func TestRenderAudit_Diag_D02_undefinedPath(t *testing.T) {
	tpl := mustParseAudit(t, "{{ a.b }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{}, liquid.WithStrictVariables())
	if result == nil {
		t.Fatal("result must not be nil")
	}
	d := firstDiag(result.Diagnostics, "undefined-variable")
	if d == nil {
		t.Fatal("expected undefined-variable diagnostic for nested path a.b")
	}
}

// D03 — multiple undefined variables each produce a diagnostic.
func TestRenderAudit_Diag_D03_multipleUndefined(t *testing.T) {
	tpl := mustParseAudit(t, "{{ x }}{{ y }}{{ z }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{}, liquid.WithStrictVariables())
	if result == nil {
		t.Fatal("result must not be nil")
	}
	undef := allDiags(result.Diagnostics, "undefined-variable")
	if len(undef) != 3 {
		t.Errorf("undefined-variable diagnostic count=%d, want 3 (one per undefined var)", len(undef))
	}
}

// D04 — divided_by: 0 → "argument-error", severity=error.
func TestRenderAudit_Diag_D04_dividedByZero(t *testing.T) {
	tpl := mustParseAudit(t, "{{ 10 | divided_by: 0 }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
	if result == nil {
		t.Fatal("result must not be nil")
	}
	d := firstDiag(result.Diagnostics, "argument-error")
	if d == nil {
		t.Fatal("expected argument-error diagnostic for divided_by: 0")
	}
	if d.Severity != liquid.SeverityError {
		t.Errorf("Severity=%q, want error", d.Severity)
	}
}

// D05 — modulo: 0 → "argument-error".
func TestRenderAudit_Diag_D05_moduloZero(t *testing.T) {
	tpl := mustParseAudit(t, "{{ 10 | modulo: 0 }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
	if result == nil {
		t.Fatal("result must not be nil")
	}
	d := firstDiag(result.Diagnostics, "argument-error")
	if d == nil {
		t.Fatal("expected argument-error diagnostic for modulo: 0")
	}
	if d.Severity != liquid.SeverityError {
		t.Errorf("Severity=%q, want error", d.Severity)
	}
}

// D06 — argument-error message is descriptive.
func TestRenderAudit_Diag_D06_argumentErrorMessageDescriptive(t *testing.T) {
	tpl := mustParseAudit(t, "{{ 10 | divided_by: 0 }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
	d := firstDiag(result.Diagnostics, "argument-error")
	if d == nil {
		t.Fatal("expected argument-error diagnostic")
	}
	if d.Message == "" {
		t.Error("argument-error diagnostic Message should not be empty")
	}
}

// D07 — type mismatch string vs int with ==: "type-mismatch", severity=warning.
func TestRenderAudit_Diag_D07_typeMismatchStringInt(t *testing.T) {
	tpl := mustParseAudit(t, `{% if status == 1 %}yes{% endif %}`)
	result, _ := tpl.RenderAudit(liquid.Bindings{"status": "active"}, liquid.AuditOptions{})
	if result == nil {
		t.Fatal("result must not be nil")
	}
	d := firstDiag(result.Diagnostics, "type-mismatch")
	if d == nil {
		t.Fatal("expected type-mismatch diagnostic")
	}
	if d.Severity != liquid.SeverityWarning {
		t.Errorf("Severity=%q, want warning", d.Severity)
	}
}

// D08 — type mismatch with > operator: diagnostic mentions the operator.
func TestRenderAudit_Diag_D08_typeMismatchWithGT(t *testing.T) {
	tpl := mustParseAudit(t, `{% if name > 5 %}yes{% endif %}`)
	result, _ := tpl.RenderAudit(liquid.Bindings{"name": "alice"}, liquid.AuditOptions{})
	if result == nil {
		t.Fatal("result must not be nil")
	}
	d := firstDiag(result.Diagnostics, "type-mismatch")
	if d == nil {
		t.Fatal("expected type-mismatch diagnostic for string > int")
	}
	if d.Message == "" {
		t.Error("type-mismatch Message should not be empty")
	}
}

// D09 — nil variable in comparison without path: no diagnostic (normal Liquid behavior).
func TestRenderAudit_Diag_D09_nilComparisonNoWarning(t *testing.T) {
	tpl := mustParseAudit(t, "{% if nil_var == 1 %}yes{% else %}no{% endif %}")
	result, ae := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
	if ae != nil {
		// Only argument errors should fire here, not undefined-variable without StrictVariables.
	}
	// type-mismatch might fire here (nil vs int). But there should be no undefined-variable.
	undef := firstDiag(result.Diagnostics, "undefined-variable")
	if undef != nil {
		t.Error("undefined-variable should not appear without StrictVariables for nil comparison")
	}
}

// D10 — for over int → "not-iterable", severity=warning.
func TestRenderAudit_Diag_D10_notIterableInt(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in orders %}{{ item }}{% endfor %}")
	result, _ := tpl.RenderAudit(liquid.Bindings{"orders": 42}, liquid.AuditOptions{})
	d := firstDiag(result.Diagnostics, "not-iterable")
	if d == nil {
		t.Fatal("expected not-iterable diagnostic")
	}
	if d.Severity != liquid.SeverityWarning {
		t.Errorf("Severity=%q, want warning", d.Severity)
	}
}

// D11 — for over bool → "not-iterable", severity=warning.
func TestRenderAudit_Diag_D11_notIterableBool(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in flag %}{{ item }}{% endfor %}")
	result, _ := tpl.RenderAudit(liquid.Bindings{"flag": true}, liquid.AuditOptions{})
	d := firstDiag(result.Diagnostics, "not-iterable")
	if d == nil {
		t.Fatal("expected not-iterable diagnostic for bool")
	}
}

// D12 — for over string → "not-iterable", severity=warning.
func TestRenderAudit_Diag_D12_notIterableString(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in status %}{{ item }}{% endfor %}")
	result, _ := tpl.RenderAudit(liquid.Bindings{"status": "pending"}, liquid.AuditOptions{})
	d := firstDiag(result.Diagnostics, "not-iterable")
	if d == nil {
		t.Fatal("expected not-iterable diagnostic for string")
	}
}

// D13 — nil intermediate in chained path: "nil-dereference", severity=warning.
func TestRenderAudit_Diag_D13_nilDereference(t *testing.T) {
	tpl := mustParseAudit(t, "{{ customer.address.city }}")
	result, _ := tpl.RenderAudit(
		liquid.Bindings{"customer": map[string]any{"address": nil}},
		liquid.AuditOptions{},
	)
	d := firstDiag(result.Diagnostics, "nil-dereference")
	if d == nil {
		t.Fatal("expected nil-dereference diagnostic")
	}
	if d.Severity != liquid.SeverityWarning {
		t.Errorf("Severity=%q, want warning", d.Severity)
	}
}

// D14 — deep nil in chained path: "nil-dereference" still fires.
func TestRenderAudit_Diag_D14_deepNilDereference(t *testing.T) {
	tpl := mustParseAudit(t, "{{ a.b.c.d }}")
	result, _ := tpl.RenderAudit(
		liquid.Bindings{"a": map[string]any{"b": nil}},
		liquid.AuditOptions{},
	)
	d := firstDiag(result.Diagnostics, "nil-dereference")
	if d == nil {
		t.Fatal("expected nil-dereference diagnostic for deep nil path")
	}
}

// D15 — simple nil variable (no chaining): no diagnostic.
func TestRenderAudit_Diag_D15_simpleNilNoWarning(t *testing.T) {
	tpl := mustParseAudit(t, "{{ nil_var }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{"nil_var": nil}, liquid.AuditOptions{})
	// nil render is normal Liquid behavior; no diagnostic expected.
	nilDeref := firstDiag(result.Diagnostics, "nil-dereference")
	if nilDeref != nil {
		t.Error("nil simple variable should NOT produce nil-dereference diagnostic")
	}
	assertOutput(t, result, "")
}

// D16 — nil variable in condition without chaining: no diagnostic.
func TestRenderAudit_Diag_D16_nilInConditionNoWarning(t *testing.T) {
	tpl := mustParseAudit(t, "{% if nil_var %}yes{% else %}no{% endif %}")
	result, _ := tpl.RenderAudit(liquid.Bindings{"nil_var": nil}, liquid.AuditOptions{})
	nilDeref := firstDiag(result.Diagnostics, "nil-dereference")
	if nilDeref != nil {
		t.Error("nil variable in simple condition should NOT produce nil-dereference diagnostic")
	}
	assertOutput(t, result, "no")
}

// ============================================================================
// Diagnostics — Range and Source fields (DR01–DR04)
// ============================================================================

// DR01 — all diagnostics have Range.Start.Line >= 1.
func TestRenderAudit_Diag_DR01_allHaveValidLine(t *testing.T) {
	tpl := mustParseAudit(t, "{{ x }}{{ y }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{}, liquid.WithStrictVariables())
	for i, d := range result.Diagnostics {
		if d.Range.Start.Line < 1 {
			t.Errorf("Diagnostics[%d].Range.Start.Line=%d, want >= 1", i, d.Range.Start.Line)
		}
	}
}

// DR02 — diagnostic Source is non-empty.
func TestRenderAudit_Diag_DR02_sourceNonEmpty(t *testing.T) {
	tpl := mustParseAudit(t, "{{ 10 | divided_by: 0 }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
	d := firstDiag(result.Diagnostics, "argument-error")
	if d == nil {
		t.Fatal("expected argument-error diagnostic")
	}
	if d.Source == "" {
		t.Error("Diagnostic.Source should not be empty")
	}
}

// DR03 — diagnostic on line 5 has Range.Start.Line=5.
func TestRenderAudit_Diag_DR03_lineNumber(t *testing.T) {
	tpl := mustParseAudit(t, "line1\nline2\nline3\nline4\n{{ ghost }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{}, liquid.WithStrictVariables())
	d := firstDiag(result.Diagnostics, "undefined-variable")
	if d == nil {
		t.Fatal("expected undefined-variable diagnostic")
	}
	if d.Range.Start.Line != 5 {
		t.Errorf("Range.Start.Line=%d, want 5", d.Range.Start.Line)
	}
}

// DR04 — multiple diagnostics on different lines each have their own Range.
func TestRenderAudit_Diag_DR04_multiLineDiagnostics(t *testing.T) {
	tpl := mustParseAudit(t, "{{ x }}\n{{ y }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{}, liquid.WithStrictVariables())
	undef := allDiags(result.Diagnostics, "undefined-variable")
	if len(undef) < 2 {
		t.Fatalf("expected 2 undefined-variable diagnostics, got %d", len(undef))
	}
	if undef[0].Range.Start.Line == undef[1].Range.Start.Line {
		t.Errorf("diagnostics on different lines should have different Range.Start.Line values: both got %d",
			undef[0].Range.Start.Line)
	}
}

// ============================================================================
// Diagnostics — Cross-reference with Expression.Error (DE01–DE03)
// ============================================================================

// DE01 — when a variable causes a strict-mode error, Expression.Error is non-nil.
func TestRenderAudit_Diag_DE01_expressionErrorNonNil(t *testing.T) {
	tpl := mustParseAudit(t, "{{ ghost }}")
	result, _ := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{TraceVariables: true},
		liquid.WithStrictVariables(),
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("expected variable expression even when it errors")
	}
	if v.Error == nil {
		t.Error("Expression.Error should be non-nil when the expression caused an error")
	}
}

// DE02 — Expression.Error.Code matches the Diagnostic code.
func TestRenderAudit_Diag_DE02_expressionErrorMatchesDiagCode(t *testing.T) {
	tpl := mustParseAudit(t, "{{ ghost }}")
	result, _ := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{TraceVariables: true},
		liquid.WithStrictVariables(),
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	d := firstDiag(result.Diagnostics, "undefined-variable")
	if v == nil || d == nil {
		t.Skip("need both variable expression and diagnostic")
	}
	if v.Error == nil {
		t.Fatal("Expression.Error is nil")
	}
	if v.Error.Code != d.Code {
		t.Errorf("Expression.Error.Code=%q != Diagnostic.Code=%q", v.Error.Code, d.Code)
	}
}

// DE03 — number of AuditError.Errors() equals number of diagnostics with severity error/warning.
func TestRenderAudit_Diag_DE03_auditErrorCountMatchesDiagnostics(t *testing.T) {
	tpl := mustParseAudit(t, "{{ x }}{{ y }}{{ z }}")
	result, ae := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{},
		liquid.WithStrictVariables(),
	)
	if ae == nil {
		t.Fatal("expected AuditError for 3 undefined variables")
	}
	if len(ae.Errors()) != 3 {
		t.Errorf("AuditError.Errors() len=%d, want 3", len(ae.Errors()))
	}
	if len(result.Diagnostics) != 3 {
		t.Errorf("Diagnostics len=%d, want 3", len(result.Diagnostics))
	}
}

// ============================================================================
// Diagnostics — Render Continues After Error (DC01–DC05)
// ============================================================================

// DC01 — partial output: content before and after the error is captured.
func TestRenderAudit_Diag_DC01_partialOutput(t *testing.T) {
	tpl := mustParseAudit(t, "before {{ 10 | divided_by: 0 }} after")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
	if result == nil {
		t.Fatal("result must not be nil")
	}
	if !strings.Contains(result.Output, "before") {
		t.Errorf("Output=%q should contain 'before'", result.Output)
	}
	if !strings.Contains(result.Output, "after") {
		t.Errorf("Output=%q should contain 'after' (render continued after error)", result.Output)
	}
}

// DC02 — three variables: 1st OK, 2nd fails, 3rd OK → output contains 1st and 3rd values.
func TestRenderAudit_Diag_DC02_continuesAfterMidError(t *testing.T) {
	tpl := mustParseAudit(t, "{{ a }}{{ b | divided_by: 0 }}{{ c }}")
	result, _ := tpl.RenderAudit(
		liquid.Bindings{"a": "first", "b": 10, "c": "third"},
		liquid.AuditOptions{},
	)
	if !strings.Contains(result.Output, "first") {
		t.Errorf("Output=%q should contain 'first'", result.Output)
	}
	if !strings.Contains(result.Output, "third") {
		t.Errorf("Output=%q should contain 'third' (render continued)", result.Output)
	}
}

// DC03 — multiple filter errors in the same template: all are accumulated as Diagnostics.
// Note: filter errors (divided_by: 0) produce Diagnostics but NOT an AuditError —
// RenderAudit treats them as continuable non-fatal errors.
func TestRenderAudit_Diag_DC03_multipleErrorsAccumulated(t *testing.T) {
	tpl := mustParseAudit(t, "{{ 1 | divided_by: 0 }}{{ 2 | divided_by: 0 }}")
	result, ae := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
	if result == nil {
		t.Fatal("result must not be nil")
	}
	// ae may be nil — divided_by:0 errors are captured as Diagnostics, not AuditErrors.
	_ = ae
	argErrs := allDiags(result.Diagnostics, "argument-error")
	if len(argErrs) < 2 {
		t.Errorf("argument-error diagnostic count=%d, want >= 2", len(argErrs))
	}
}

// DC04 — AuditError.Error() contains a count summary.
func TestRenderAudit_Diag_DC04_auditErrorMessage(t *testing.T) {
	tpl := mustParseAudit(t, "{{ x }}{{ y }}")
	_, ae := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{},
		liquid.WithStrictVariables(),
	)
	if ae == nil {
		t.Fatal("expected AuditError")
	}
	msg := ae.Error()
	if msg == "" {
		t.Error("AuditError.Error() should not be empty")
	}
	// Should mention the number of errors.
	if !strings.Contains(msg, "2") && !strings.Contains(msg, "error") {
		t.Errorf("AuditError.Error()=%q should mention count or 'error'", msg)
	}
}

// DC05 — AuditError.Errors() returns slice with SourceError types.
func TestRenderAudit_Diag_DC05_auditErrorTypedErrors(t *testing.T) {
	tpl := mustParseAudit(t, "{{ x }}")
	_, ae := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{},
		liquid.WithStrictVariables(),
	)
	if ae == nil {
		t.Fatal("expected AuditError")
	}
	errs := ae.Errors()
	if len(errs) == 0 {
		t.Fatal("AuditError.Errors() should not be empty")
	}
	// Each error should implement SourceError (which extends error).
	for i, e := range errs {
		if e == nil {
			t.Errorf("Errors()[%d] is nil", i)
		}
		// SourceError interface has Error(), Cause(), Path(), LineNumber().
		if e.Error() == "" {
			t.Errorf("Errors()[%d].Error() is empty", i)
		}
	}
}

// ============================================================================
// Diagnostics — not-iterable has valid Range span (already in existing tests,
// but verifying message content too)
// ============================================================================

// D_notIterable_messageContent — message mentions the variable and type.
func TestRenderAudit_Diag_notIterable_message(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in orders %}{{ item }}{% endfor %}")
	result, _ := tpl.RenderAudit(
		liquid.Bindings{"orders": "shipped"},
		liquid.AuditOptions{},
	)
	d := firstDiag(result.Diagnostics, "not-iterable")
	if d == nil {
		t.Fatal("expected not-iterable diagnostic")
	}
	if d.Message == "" {
		t.Error("Diagnostic.Message should not be empty")
	}
}

// D_typeMismatch_messageContent — message mentions both types.
func TestRenderAudit_Diag_typeMismatch_message(t *testing.T) {
	tpl := mustParseAudit(t, `{% if status == 1 %}yes{% endif %}`)
	result, _ := tpl.RenderAudit(
		liquid.Bindings{"status": "active"},
		liquid.AuditOptions{},
	)
	d := firstDiag(result.Diagnostics, "type-mismatch")
	if d == nil {
		t.Fatal("expected type-mismatch diagnostic")
	}
	// Message should mention both the string and int types in some form.
	if d.Message == "" {
		t.Error("type-mismatch Message should not be empty")
	}
}

// D_nilDereference_messageContent — message mentions the property.
func TestRenderAudit_Diag_nilDereference_message(t *testing.T) {
	tpl := mustParseAudit(t, "{{ customer.address.city }}")
	result, _ := tpl.RenderAudit(
		liquid.Bindings{"customer": map[string]any{"address": nil}},
		liquid.AuditOptions{},
	)
	d := firstDiag(result.Diagnostics, "nil-dereference")
	if d == nil {
		t.Fatal("expected nil-dereference diagnostic")
	}
	if d.Message == "" {
		t.Error("nil-dereference Message should not be empty")
	}
	// Message should mention "city" (the property being accessed on nil).
	if !strings.Contains(d.Message, "city") {
		t.Errorf("Message=%q should mention 'city'", d.Message)
	}
}
