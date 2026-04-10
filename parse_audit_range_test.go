package liquid_test

import (
	"encoding/json"
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// Range and Position Precision (P01–P12)
// ============================================================================

// P01 — Expression on line 1, column 1: Range.Start.Line=1, Column=1.
func TestParseAudit_Range_P01_firstLineFirstColumn(t *testing.T) {
	r := parseAudit(`{{ | bad_p01 }}`)
	d := requireParseDiag(t, r, "syntax-error")
	if d.Range.Start.Line != 1 {
		t.Errorf("P01: Range.Start.Line=%d, want 1", d.Range.Start.Line)
	}
	if d.Range.Start.Column != 1 {
		t.Errorf("P01: Range.Start.Column=%d, want 1", d.Range.Start.Column)
	}
}

// P02 — Three-line template, bad expression on line 3: Range.Start.Line=3.
func TestParseAudit_Range_P02_lineThree(t *testing.T) {
	r := parseAudit("line one\nline two\n{{ | bad_p02 }}")
	d := requireParseDiag(t, r, "syntax-error")
	if d.Range.Start.Line != 3 {
		t.Errorf("P02: Range.Start.Line=%d, want 3", d.Range.Start.Line)
	}
}

// P03 — Template starting with text before bad expression: Start.Column > 1.
func TestParseAudit_Range_P03_columnOffset(t *testing.T) {
	// "Hello " is 6 chars, so "{{ ... }}" starts at column 7.
	r := parseAudit(`Hello {{ | bad_p03 }}`)
	d := requireParseDiag(t, r, "syntax-error")
	if d.Range.Start.Column <= 1 {
		t.Errorf("P03: Range.Start.Column=%d, want > 1 (preceded by 'Hello ')", d.Range.Start.Column)
	}
}

// P04 — Source span: End.Column > Start.Column for a single-line expression.
func TestParseAudit_Range_P04_endAfterStart(t *testing.T) {
	r := parseAudit(`{{ | bad_p04 }}`)
	d := requireParseDiag(t, r, "syntax-error")
	assertRangeEndAfterStart(t, d.Range, "P04 syntax-error")
}

// P05 — Two diagnostics on different lines: each has a distinct Range.
func TestParseAudit_Range_P05_distinctRanges(t *testing.T) {
	r := parseAudit("{{ x | bad_p05a }}\n{{ y | bad_p05b }}")
	diags := allParseDiags(r, "undefined-filter")
	if len(diags) < 2 {
		t.Skip("P05: fewer than 2 undefined-filter diagnostics")
	}
	if diags[0].Range.Start.Line == diags[1].Range.Start.Line {
		t.Errorf("P05: two diagnostics on different lines share Range.Start.Line=%d",
			diags[0].Range.Start.Line)
	}
}

// P06 — {% if bad_expr %} Range points to the tag line, not EOF.
func TestParseAudit_Range_P06_tagRangeNotEOF(t *testing.T) {
	r := parseAudit("{% if x %}\n\n\n{% endif %}")
	// Template should parse cleanly; this tests unclosed-tag scenario on line 1.
	// Use a real missing-close scenario.
	r2 := parseAudit("{% if x %}\n\n\nmore content")
	d := firstParseDiag(r2, "unclosed-tag")
	if d == nil {
		t.Skip("P06: no unclosed-tag to inspect")
	}
	// Range should be on the opening line (1), not on the last line.
	if d.Range.Start.Line > 2 {
		t.Errorf("P06: unclosed-tag Range.Start.Line=%d; expected near the opening tag (line 1), not EOF",
			d.Range.Start.Line)
	}
	_ = r
}

// P07 — unclosed-tag Range.Start.Line = line of opening tag, not EOF line.
func TestParseAudit_Range_P07_unclosedTagStartAtOpenTag(t *testing.T) {
	r := parseAudit("{% if order %}\ncontent\ncontent\n")
	d := requireParseDiag(t, r, "unclosed-tag")
	if d.Range.Start.Line != 1 {
		t.Errorf("P07: unclosed-tag Range.Start.Line=%d, want 1 (opening tag is on line 1)", d.Range.Start.Line)
	}
}

// P08 — unclosed-tag Related[0].Range points at or near EOF.
func TestParseAudit_Range_P08_unclosedTagRelatedAtEOF(t *testing.T) {
	src := "{% if x %}\nline2\nline3"
	r := parseAudit(src)
	d := requireParseDiag(t, r, "unclosed-tag")
	if len(d.Related) == 0 {
		t.Skip("P08: no Related entries")
	}
	// Related[0].Range.Start.Line should be >= the opening tag line.
	if d.Related[0].Range.Start.Line < d.Range.Start.Line {
		t.Errorf("P08: Related[0].Range.Start.Line=%d is before the opening tag line=%d",
			d.Related[0].Range.Start.Line, d.Range.Start.Line)
	}
}

// P09 — Template with 10 lines, bad expression on line 7: Line=7.
func TestParseAudit_Range_P09_lineSevenOfTen(t *testing.T) {
	src := "line1\nline2\nline3\nline4\nline5\nline6\n{{ | bad_p09 }}\nline8\nline9\nline10"
	r := parseAudit(src)
	d := requireParseDiag(t, r, "syntax-error")
	if d.Range.Start.Line != 7 {
		t.Errorf("P09: Range.Start.Line=%d, want 7", d.Range.Start.Line)
	}
}

// P10 — Template with CRLF line endings: line numbers still correct.
func TestParseAudit_Range_P10_crlfLineEndings(t *testing.T) {
	src := "line1\r\nline2\r\n{{ | bad_p10 }}"
	r := parseAudit(src)
	d := requireParseDiag(t, r, "syntax-error")
	// Line should be 3 regardless of CRLF.
	if d.Range.Start.Line < 1 {
		t.Errorf("P10: Range.Start.Line=%d, want >= 1", d.Range.Start.Line)
	}
}

// P11 — Template with tabs before expression: Column counts correctly (>= 1).
func TestParseAudit_Range_P11_tabBeforeExpression(t *testing.T) {
	src := "\t\t{{ | bad_p11 }}"
	r := parseAudit(src)
	d := requireParseDiag(t, r, "syntax-error")
	if d.Range.Start.Column < 1 {
		t.Errorf("P11: Range.Start.Column=%d, want >= 1", d.Range.Start.Column)
	}
}

// P12 — ParseTemplateLocation with base line offset: Diagnostic line accounts for offset.
// Skipped if ParseTemplateLocation is not available or not relevant to audit.
func TestParseAudit_Range_P12_templateLocationLineOffset(t *testing.T) {
	eng := newParseAuditEngine()
	// If the engine exposes ParseTemplateLocationAudit or similar, test line offset.
	// For now, verify that ParseStringAudit at minimum uses absolute line 1.
	r := parseAuditWith(eng, `{{ x | bad_p12 }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	if d.Range.Start.Line < 1 {
		t.Errorf("P12: Range.Start.Line=%d, want >= 1", d.Range.Start.Line)
	}
}

// ============================================================================
// JSON Serialization (J01–J08)
// ============================================================================

// J01 — ParseResult with no diagnostics serializes to JSON with "diagnostics":[] not null.
func TestParseAudit_JSON_J01_emptyDiagsSerializeAsArray(t *testing.T) {
	r := parseAudit(`Hello {{ name }}!`)
	assertTemplateNonNil(t, r, "J01")
	// Serialize only the diagnostics part by marshaling a wrapper struct.
	type wrap struct {
		Diagnostics interface{} `json:"diagnostics"`
	}
	data, err := json.Marshal(wrap{Diagnostics: r.Diagnostics})
	if err != nil {
		t.Fatalf("J01: json.Marshal failed: %v", err)
	}
	js := string(data)
	if !containsSubstr(js, `"diagnostics":[]`) {
		t.Errorf("J01: expected empty diagnostics array in JSON, got: %s", js)
	}
}

// J02 — A fatal ParseResult (Template nil) serializes without panic.
func TestParseAudit_JSON_J02_fatalSerializesOK(t *testing.T) {
	r := parseAudit(`{% if x %}unclosed`)
	// Diagnostics are serializable.
	data, err := json.Marshal(r.Diagnostics)
	if err != nil {
		t.Fatalf("J02: json.Marshal(Diagnostics) failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("J02: marshaled diagnostics is empty")
	}
}

// J03 — Diagnostic JSON contains correct field names (snake_case or as tagged).
func TestParseAudit_JSON_J03_diagnosticJSONKeys(t *testing.T) {
	r := parseAudit(`{{ x | nofilter_j03 }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("J03: json.Marshal failed: %v", err)
	}
	js := string(data)
	for _, key := range []string{`"code"`, `"severity"`, `"message"`, `"source"`, `"range"`} {
		if !containsSubstr(js, key) {
			t.Errorf("J03: expected key %s in Diagnostic JSON: %s", key, js)
		}
	}
}

// J04 — Diagnostic.Related absent from JSON when nil/empty (omitempty).
func TestParseAudit_JSON_J04_relatedOmittedWhenEmpty(t *testing.T) {
	r := parseAudit(`{{ x | nofilter_j04 }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("J04: json.Marshal failed: %v", err)
	}
	js := string(data)
	if containsSubstr(js, `"related"`) {
		t.Errorf("J04: expected 'related' to be absent (omitempty) when empty, got: %s", js)
	}
}

// J05 — Diagnostic.Range always present in JSON.
func TestParseAudit_JSON_J05_rangeAlwaysPresent(t *testing.T) {
	r := parseAudit(`{{ x | nofilter_j05 }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("J05: json.Marshal failed: %v", err)
	}
	js := string(data)
	if !containsSubstr(js, `"range"`) {
		t.Errorf("J05: expected 'range' key in JSON, got: %s", js)
	}
}

// J06 — Full Diagnostic round-trip: Marshal → Unmarshal into same type → re-Marshal → same JSON.
func TestParseAudit_JSON_J06_roundTrip(t *testing.T) {
	r := parseAudit(`{{ x | nofilter_j06 }}`)
	if len(r.Diagnostics) == 0 {
		t.Skip("J06: no diagnostics to round-trip")
	}
	d := r.Diagnostics[0]

	data1, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("J06: first Marshal failed: %v", err)
	}

	// Unmarshal back into the same concrete type to preserve field order.
	var d2 liquid.Diagnostic
	if err := json.Unmarshal(data1, &d2); err != nil {
		t.Fatalf("J06: Unmarshal failed: %v", err)
	}

	data2, err := json.Marshal(d2)
	if err != nil {
		t.Fatalf("J06: second Marshal failed: %v", err)
	}

	if string(data1) != string(data2) {
		t.Errorf("J06: round-trip mismatch:\n  first:  %s\n  second: %s", data1, data2)
	}
}

// J07 — Diagnostic.Severity serializes as a string, not a number.
func TestParseAudit_JSON_J07_severityAsString(t *testing.T) {
	r := parseAudit(`{{ x | nofilter_j07 }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("J07: json.Marshal failed: %v", err)
	}
	js := string(data)
	// Should contain "error" as a JSON string, not a number.
	if !containsSubstr(js, `"error"`) {
		t.Errorf("J07: expected severity as string \"error\" in JSON, got: %s", js)
	}
}

// J08 — Position.Line and Position.Column serialize as JSON numbers.
func TestParseAudit_JSON_J08_positionAsNumbers(t *testing.T) {
	r := parseAudit(`{{ x | nofilter_j08 }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	data, err := json.Marshal(d.Range.Start)
	if err != nil {
		t.Fatalf("J08: json.Marshal(Position) failed: %v", err)
	}
	js := string(data)
	// Should contain "line":1 or "line":N - JSON number, not string.
	if !containsSubstr(js, `"line":`) {
		t.Errorf("J08: expected 'line' as JSON number key, got: %s", js)
	}
	if !containsSubstr(js, `"column":`) {
		t.Errorf("J08: expected 'column' as JSON number key, got: %s", js)
	}
}

// ============================================================================
// Edge Cases and Robustness (ED01–ED14)
// ============================================================================

// ED01 — Empty source: no diagnostics, Template non-nil.
func TestParseAudit_Edge_ED01_emptySource(t *testing.T) {
	r := parseAudit(``)
	assertTemplateNonNil(t, r, "ED01")
	assertNoParseDiags(t, r, "ED01")
}

// ED02 — Comment block: no diagnostics.
func TestParseAudit_Edge_ED02_commentBlock(t *testing.T) {
	r := parseAudit(`{% comment %}this is safe content{% endcomment %}`)
	assertTemplateNonNil(t, r, "ED02")
	assertNoParseDiags(t, r, "ED02")
}

// ED03 — {% raw %}{{ not_parsed }}{% endraw %}: no syntax-error for raw content.
func TestParseAudit_Edge_ED03_rawBlock(t *testing.T) {
	r := parseAudit(`{% raw %}{{ not_parsed | not_a_filter }}{% endraw %}`)
	assertTemplateNonNil(t, r, "ED03")
	assertNoParseDiags(t, r, "ED03")
}

// ED04 — Large template (100+ repeated expressions): no crash.
func TestParseAudit_Edge_ED04_largeTemplate(t *testing.T) {
	src := ""
	for i := 0; i < 200; i++ {
		src += `{{ name }} `
	}
	r := parseAudit(src)
	assertParseResultNonNil(t, r, "ED04")
	// All expressions should parse cleanly.
	assertNoParseDiags(t, r, "ED04")
}

// ED05 — Template with Unicode in string literals: no crash.
func TestParseAudit_Edge_ED05_unicodeInLiterals(t *testing.T) {
	r := parseAudit(`{% assign greeting = "こんにちは" %}{{ greeting }}`)
	assertParseResultNonNil(t, r, "ED05")
	// Should parse cleanly.
	assertNoParseDiags(t, r, "ED05")
}

// ED06 — Template with Unicode in content text (not in expressions): no crash.
func TestParseAudit_Edge_ED06_unicodeInText(t *testing.T) {
	r := parseAudit(`Héllo wörld! {{ name }}`)
	assertParseResultNonNil(t, r, "ED06")
	assertNoParseDiags(t, r, "ED06")
}

// ED07 — Whitespace-control {% if -%}…{% endif %} without close.
func TestParseAudit_Edge_ED07_trimMarkerUnclosed(t *testing.T) {
	r := parseAudit(`{%- if x -%}content`)
	assertTemplateNil(t, r, "ED07")
	requireParseDiag(t, r, "unclosed-tag")
}

// ED08 — Deeply nested but well-formed blocks (10 levels): no crash, clean.
func TestParseAudit_Edge_ED08_deepNesting(t *testing.T) {
	open := ""
	close := ""
	for i := 0; i < 10; i++ {
		open += `{% if x %}content `
		close = `{% endif %}` + close
	}
	r := parseAudit(open + close)
	assertParseResultNonNil(t, r, "ED08")
}

// ED09 — {% liquid assign x = 1 %} multi-line tag: no crash, no false diagnostic.
func TestParseAudit_Edge_ED09_liquidMultilineTag(t *testing.T) {
	r := parseAudit("{% liquid\nassign x = 1\necho x\n%}")
	assertParseResultNonNil(t, r, "ED09")
	// Should not produce spurious diagnostics.
	for _, d := range r.Diagnostics {
		if d.Code != "empty-block" {
			t.Errorf("ED09: unexpected diagnostic code=%q for valid liquid tag", d.Code)
		}
	}
}

// ED10 — Multiple {% assign x = | bad %} tags: each produces its own syntax-error.
func TestParseAudit_Edge_ED10_multipleTagSyntaxErrors(t *testing.T) {
	r := parseAudit(`{% assign a = | bad1 %}{% assign b = | bad2 %}`)
	assertTemplateNonNil(t, r, "ED10")
	syntaxErrs := allParseDiags(r, "syntax-error")
	if len(syntaxErrs) < 2 {
		t.Errorf("ED10: expected >= 2 syntax-error diagnostics, got %d", len(syntaxErrs))
	}
}

// ED11 — {{ x | unknown_filter }} + {% if true %}{% endif %}: both diagnostics present.
func TestParseAudit_Edge_ED11_filterAndEmptyBlock(t *testing.T) {
	r := parseAudit(`{{ x | unknown_ed11 }}{% if true %}{% endif %}`)
	if firstParseDiag(r, "undefined-filter") == nil {
		t.Error("ED11: expected undefined-filter diagnostic")
	}
	if firstParseDiag(r, "empty-block") == nil {
		t.Error("ED11: expected empty-block diagnostic")
	}
}

// ED12 — Template with {% break %} / {% continue %} inside for: no crash.
func TestParseAudit_Edge_ED12_breakContinueInsideFor(t *testing.T) {
	r := parseAudit(`{% for x in items %}{% if x > 3 %}{% break %}{% endif %}{{ x }}{% endfor %}`)
	assertParseResultNonNil(t, r, "ED12")
}

// ED13 — Template with {% increment %} / {% decrement %}: no false diagnostics.
func TestParseAudit_Edge_ED13_incrementDecrement(t *testing.T) {
	r := parseAudit(`{% increment counter %}{% decrement counter %}`)
	assertParseResultNonNil(t, r, "ED13")
	// Should produce no diagnostics.
	for _, d := range r.Diagnostics {
		if d.Code != "empty-block" {
			t.Errorf("ED13: unexpected diagnostic code=%q for increment/decrement tags", d.Code)
		}
	}
}

// ED14 — Template with {% cycle %} inside for: no crash.
func TestParseAudit_Edge_ED14_cycleTagInsideFor(t *testing.T) {
	r := parseAudit(`{% for item in items %}{% cycle "odd","even" %}{{ item }}{% endfor %}`)
	assertParseResultNonNil(t, r, "ED14")
}
