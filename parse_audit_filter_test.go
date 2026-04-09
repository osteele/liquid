package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// undefined-filter — Static Analysis (F01–F16)
// ============================================================================

// F01 — {{ x | no_such_filter }}: Diagnostics contains code="undefined-filter".
func TestParseAudit_Filter_F01_unknownFilter(t *testing.T) {
	r := parseAudit(`{{ x | no_such_filter }}`)
	requireParseDiag(t, r, "undefined-filter")
}

// F02 — {{ x | upcase }}: valid filter → no undefined-filter diagnostic.
func TestParseAudit_Filter_F02_validFilter(t *testing.T) {
	r := parseAudit(`{{ x | upcase }}`)
	assertNoParseDiags(t, r, "F02")
}

// F03 — {{ x | no_such | upcase }}: one bad filter in chain → one undefined-filter.
func TestParseAudit_Filter_F03_oneBadOneGood(t *testing.T) {
	r := parseAudit(`{{ x | no_such | upcase }}`)
	filters := allParseDiags(r, "undefined-filter")
	if len(filters) != 1 {
		t.Errorf("F03: expected 1 undefined-filter diagnostic (only 'no_such' is bad), got %d", len(filters))
	}
}

// F04 — {{ x | one_bad | two_bad }}: two unknown filters → two undefined-filter diagnostics.
func TestParseAudit_Filter_F04_twoBadFilters(t *testing.T) {
	r := parseAudit(`{{ x | one_bad | two_bad }}`)
	filters := allParseDiags(r, "undefined-filter")
	if len(filters) != 2 {
		t.Errorf("F04: expected 2 undefined-filter diagnostics, got %d", len(filters))
	}
}

// F05 — {{ x | bad }} and {{ y | also_bad }}: two bad object nodes → two undefined-filter.
func TestParseAudit_Filter_F05_twoBadObjects(t *testing.T) {
	r := parseAudit(`{{ x | bad_filter_one }} text {{ y | bad_filter_two }}`)
	filters := allParseDiags(r, "undefined-filter")
	if len(filters) != 2 {
		t.Errorf("F05: expected 2 undefined-filter diagnostics, got %d", len(filters))
	}
}

// F06 — {% assign x = val | bad_filter %}: unknown filter in assign → undefined-filter.
func TestParseAudit_Filter_F06_assignUnknownFilter(t *testing.T) {
	r := parseAudit(`{% assign x = val | bad_filter %}`)
	requireParseDiag(t, r, "undefined-filter")
}

// F07 — {% capture %}{{ val | bad_filter }}{% endcapture %}: unknown filter in capture body.
func TestParseAudit_Filter_F07_captureUnknownFilter(t *testing.T) {
	r := parseAudit(`{% capture x %}{{ val | bad_filter }}{% endcapture %}`)
	requireParseDiag(t, r, "undefined-filter")
}

// F08 — Code field equals exactly "undefined-filter".
func TestParseAudit_Filter_F08_codeField(t *testing.T) {
	r := parseAudit(`{{ x | no_such_filter }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	assertDiagField(t, d.Code, "undefined-filter", "Code", "undefined-filter")
}

// F09 — Severity equals exactly "error".
func TestParseAudit_Filter_F09_severityError(t *testing.T) {
	r := parseAudit(`{{ x | no_such_filter }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	assertDiagField(t, string(d.Severity), string(liquid.SeverityError), "Severity", "undefined-filter")
}

// F10 — Source contains the full expression including both filter and variable.
func TestParseAudit_Filter_F10_sourceContainsExpression(t *testing.T) {
	r := parseAudit(`{{ order.total | my_custom | round }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	if len(d.Source) == 0 {
		t.Fatal("F10: undefined-filter Source is empty")
	}
	// Source should contain at least the object delimiters.
	assertDiagContains(t, "Source", d.Source, "{{", "undefined-filter")
}

// F11 — Range points to the expression line.
func TestParseAudit_Filter_F11_rangeLineCorrect(t *testing.T) {
	r := parseAudit(`{{ x | bad_filter }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	if d.Range.Start.Line < 1 {
		t.Errorf("F11: Range.Start.Line=%d, want >= 1", d.Range.Start.Line)
	}
}

// F12 — Message mentions the unknown filter name.
func TestParseAudit_Filter_F12_messageContainsFilterName(t *testing.T) {
	r := parseAudit(`{{ x | my_unusual_filter_xyz }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	assertDiagContains(t, "Message", d.Message, "my_unusual_filter_xyz", "undefined-filter")
}

// F13 — undefined-filter co-exists with syntax-error: both codes in Diagnostics.
func TestParseAudit_Filter_F13_coexistsWithSyntaxError(t *testing.T) {
	r := parseAudit(`{{ | bad_syntax }} {{ x | unknown_filter }}`)
	assertParseResultNonNil(t, r, "F13")
	hasSyntax := firstParseDiag(r, "syntax-error") != nil
	hasFilter := firstParseDiag(r, "undefined-filter") != nil
	if !hasSyntax {
		t.Error("F13: expected a syntax-error diagnostic")
	}
	if !hasFilter {
		t.Error("F13: expected an undefined-filter diagnostic")
	}
}

// F14 — Engine with custom registered filter: that filter does not produce undefined-filter.
func TestParseAudit_Filter_F14_customRegisteredFilterNoFalsePositive(t *testing.T) {
	eng := newParseAuditEngine()
	eng.RegisterFilter("my_custom_filter", func(s string) string { return s })
	r := parseAuditWith(eng, `{{ x | my_custom_filter }}`)
	d := firstParseDiag(r, "undefined-filter")
	if d != nil {
		t.Errorf("F14: unexpected undefined-filter diagnostic for registered filter 'my_custom_filter'")
	}
}

// F15 — Parse is independent of render-time lax-filters flag; undefined-filter is still
// detected at parse time when the filter is not registered.
func TestParseAudit_Filter_F15_parseIndependentOfLaxFilters(t *testing.T) {
	// ParseStringAudit uses the engine's filter registry, not a render option.
	// This test confirms the static walk is not suppressed by any "lax" setting
	// that might be on a future parse option (there is none in the current design).
	r := parseAudit(`{{ x | definitely_unknown_filter_zzzz }}`)
	d := firstParseDiag(r, "undefined-filter")
	if d == nil {
		t.Error("F15: expected undefined-filter diagnostic; static walk should always report unknown filters")
	}
}

// F16 — Template is still non-nil for undefined-filter (non-fatal).
func TestParseAudit_Filter_F16_templateNonNilForUndefinedFilter(t *testing.T) {
	r := parseAudit(`{{ x | no_such_filter }}`)
	assertTemplateNonNil(t, r, "F16")
}
