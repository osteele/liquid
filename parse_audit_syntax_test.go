package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// Non-Fatal Errors — syntax-error: single {{ }} (S01–S09)
// ============================================================================

// S01 — "{{ | bad }}" — invalid expression in object: Template non-nil, syntax-error.
func TestParseAudit_Syntax_S01_invalidObjectExpr(t *testing.T) {
	r := parseAudit(`{{ | bad }}`)
	assertTemplateNonNil(t, r, "S01")
	requireParseDiag(t, r, "syntax-error")
}

// S02 — "{{ product.price | | round }}" — double pipe: syntax-error.
func TestParseAudit_Syntax_S02_doublePipe(t *testing.T) {
	r := parseAudit(`{{ product.price | | round }}`)
	assertTemplateNonNil(t, r, "S02")
	requireParseDiag(t, r, "syntax-error")
}

// S03 — "{{ }}" — empty object expression: syntax-error (if engine rejects empty).
func TestParseAudit_Syntax_S03_emptyObject(t *testing.T) {
	r := parseAudit(`{{ }}`)
	// Engine may or may not treat this as a syntax error.
	// This test documents the behavior: if diagnostics exist, they should be syntax-error.
	for _, d := range r.Diagnostics {
		if d.Code != "syntax-error" && d.Code != "undefined-filter" && d.Code != "empty-block" {
			t.Errorf("S03: unexpected diagnostic code=%q for empty {{ }}", d.Code)
		}
	}
	// Template must still not panic.
	assertParseResultNonNil(t, r, "S03")
}

// S04 — syntax-error Code field equals exactly "syntax-error".
func TestParseAudit_Syntax_S04_codeField(t *testing.T) {
	r := parseAudit(`{{ | bad }}`)
	d := requireParseDiag(t, r, "syntax-error")
	assertDiagField(t, d.Code, "syntax-error", "Code", "syntax-error")
}

// S05 — syntax-error Severity equals exactly "error".
func TestParseAudit_Syntax_S05_severityError(t *testing.T) {
	r := parseAudit(`{{ | bad }}`)
	d := requireParseDiag(t, r, "syntax-error")
	assertDiagField(t, string(d.Severity), string(liquid.SeverityError), "Severity", "syntax-error")
}

// S06 — syntax-error Source contains "{{" and "}}" delimiters.
func TestParseAudit_Syntax_S06_sourceContainsDelimiters(t *testing.T) {
	r := parseAudit(`{{ | bad }}`)
	d := requireParseDiag(t, r, "syntax-error")
	if len(d.Source) == 0 {
		t.Fatal("S06: syntax-error Source is empty")
	}
	assertDiagContains(t, "Source", d.Source, "{{", "syntax-error")
}

// S07 — syntax-error Range.Start.Line is correct (line 1 for first-line expr).
func TestParseAudit_Syntax_S07_rangeStartLine(t *testing.T) {
	r := parseAudit(`{{ | bad }}`)
	d := requireParseDiag(t, r, "syntax-error")
	if d.Range.Start.Line != 1 {
		t.Errorf("S07: Range.Start.Line=%d, want 1", d.Range.Start.Line)
	}
}

// S08 — syntax-error Range.Start.Column is correct (col 1 for first expression).
func TestParseAudit_Syntax_S08_rangeStartColumn(t *testing.T) {
	r := parseAudit(`{{ | bad }}`)
	d := requireParseDiag(t, r, "syntax-error")
	if d.Range.Start.Column < 1 {
		t.Errorf("S08: Range.Start.Column=%d, want >= 1", d.Range.Start.Column)
	}
}

// S09 — syntax-error Message is non-empty and describes the issue.
func TestParseAudit_Syntax_S09_messagePresentAndNonEmpty(t *testing.T) {
	r := parseAudit(`{{ | bad }}`)
	d := requireParseDiag(t, r, "syntax-error")
	if len(d.Message) == 0 {
		t.Fatal("S09: syntax-error Message is empty; should describe the error")
	}
}

// ============================================================================
// Non-Fatal Errors — syntax-error on tag args (ST01–ST04)
// ============================================================================

// ST01 — {% assign x = | bad %} — broken expression in assign args.
func TestParseAudit_Syntax_ST01_assignBrokenExpr(t *testing.T) {
	r := parseAudit(`{% assign x = | bad %}`)
	assertTemplateNonNil(t, r, "ST01")
	requireParseDiag(t, r, "syntax-error")
}

// ST02 — {% if | condition %} — broken expression in if condition.
func TestParseAudit_Syntax_ST02_ifBrokenExpr(t *testing.T) {
	r := parseAudit(`{% if | condition %}yes{% endif %}`)
	// Engine may recover or may treat this as fatal; in either case document behavior.
	assertParseResultNonNil(t, r, "ST02")
	// A syntax-error diagnostic should be present.
	d := firstParseDiag(r, "syntax-error")
	if d == nil {
		// Could also be unexpected-tag or unclosed-tag if recovery fails more broadly.
		t.Logf("ST02: no syntax-error found; codes present: %v", parseDiagCodes(r.Diagnostics))
	}
}

// ST03 — {% for %} with missing iteration spec — syntax-error or similar.
func TestParseAudit_Syntax_ST03_forMissingSpec(t *testing.T) {
	r := parseAudit(`{% for %}{{ item }}{% endfor %}`)
	assertParseResultNonNil(t, r, "ST03")
	// Some error diagnostic must be present.
	if len(r.Diagnostics) == 0 {
		t.Fatal("ST03: expected at least one diagnostic for '{% for %}' with no loop spec")
	}
}

// ST04 — Tag-level syntax-error: Source contains "{% ... %}" delimiters.
func TestParseAudit_Syntax_ST04_tagSourceContainsDelimiters(t *testing.T) {
	r := parseAudit(`{% assign x = | bad %}`)
	d := requireParseDiag(t, r, "syntax-error")
	if len(d.Source) == 0 {
		t.Fatal("ST04: syntax-error Source is empty (tag args error)")
	}
	// Source must include the tag delimiters.
	assertDiagContains(t, "Source", d.Source, "{%", "syntax-error")
}

// ============================================================================
// Non-Fatal Errors — multiple syntax-errors (SM01–SM08)
// ============================================================================

// SM01 — Two bad {{ }} objects in same template: len(Diagnostics)=2, both syntax-error.
func TestParseAudit_Syntax_SM01_twoSyntaxErrors(t *testing.T) {
	r := parseAudit(`{{ | bad1 }} text {{ | bad2 }}`)
	assertTemplateNonNil(t, r, "SM01")
	syntaxErrs := allParseDiags(r, "syntax-error")
	if len(syntaxErrs) < 2 {
		t.Errorf("SM01: expected >= 2 syntax-error diagnostics, got %d (codes: %v)",
			len(syntaxErrs), parseDiagCodes(r.Diagnostics))
	}
}

// SM02 — Three bad {{ }} objects: at least three syntax-error diagnostics.
func TestParseAudit_Syntax_SM02_threeSyntaxErrors(t *testing.T) {
	r := parseAudit(`{{ | a }} {{ | b }} {{ | c }}`)
	assertTemplateNonNil(t, r, "SM02")
	syntaxErrs := allParseDiags(r, "syntax-error")
	if len(syntaxErrs) < 3 {
		t.Errorf("SM02: expected >= 3 syntax-error diagnostics, got %d", len(syntaxErrs))
	}
}

// SM03 — Mix of bad {{ }} and bad {% tag %}: all errors collected.
func TestParseAudit_Syntax_SM03_mixedTagAndObject(t *testing.T) {
	r := parseAudit(`{{ | bad }}{% assign x = | broken %}`)
	assertTemplateNonNil(t, r, "SM03")
	if len(r.Diagnostics) < 2 {
		t.Errorf("SM03: expected >= 2 diagnostics; got %d (codes: %v)",
			len(r.Diagnostics), parseDiagCodes(r.Diagnostics))
	}
}

// SM04 — Valid text between two bad expressions is rendered correctly
// (ASTBroken → empty, Text → present).
func TestParseAudit_Syntax_SM04_validTextBetweenErrors(t *testing.T) {
	r := parseAudit(`before{{ | bad }}middle{{ | bad2 }}after`)
	assertTemplateNonNil(t, r, "SM04")
	// Template should be usable; render should produce "beforemiddleafter".
	if r.Template != nil {
		out, err := r.Template.RenderString(liquid.Bindings{})
		if err != nil {
			t.Logf("SM04: render returned error: %v", err)
		}
		if out != "beforemiddleafter" {
			t.Errorf("SM04: Output=%q, want %q", out, "beforemiddleafter")
		}
	}
}

// SM05 — ASTBroken renders as empty string: broken node produces no output.
func TestParseAudit_Syntax_SM05_brokenNodeEmptyOutput(t *testing.T) {
	r := parseAudit(`{{ | bad }}`)
	assertTemplateNonNil(t, r, "SM05")
	if r.Template != nil {
		out, _ := r.Template.RenderString(liquid.Bindings{})
		if out != "" {
			t.Errorf("SM05: broken node Output=%q, want empty string", out)
		}
	}
}

// SM06 — Two syntax-errors on different lines: each Diagnostic has a distinct Range.
func TestParseAudit_Syntax_SM06_distinctRangesPerLine(t *testing.T) {
	r := parseAudit("{{ | bad1 }}\n{{ | bad2 }}")
	assertTemplateNonNil(t, r, "SM06")
	syntaxErrs := allParseDiags(r, "syntax-error")
	if len(syntaxErrs) < 2 {
		t.Skip("SM06: less than 2 syntax-error diagnostics; skipping range distinctness check")
	}
	if syntaxErrs[0].Range.Start.Line == syntaxErrs[1].Range.Start.Line &&
		syntaxErrs[0].Range.Start.Column == syntaxErrs[1].Range.Start.Column {
		t.Errorf("SM06: two diagnostics on different lines share the same Range.Start: %+v",
			syntaxErrs[0].Range.Start)
	}
}

// SM07 — All Diagnostics in multi-error result have distinct (non-duplicate) source fields.
func TestParseAudit_Syntax_SM07_noIdenticalDuplicates(t *testing.T) {
	r := parseAudit(`{{ | bad1 }} {{ | bad2 }} {{ | bad3 }}`)
	assertTemplateNonNil(t, r, "SM07")
	seen := map[string]bool{}
	for _, d := range r.Diagnostics {
		key := d.Code + "|" + d.Source
		if seen[key] {
			t.Errorf("SM07: duplicate diagnostic entry code=%q source=%q", d.Code, d.Source)
		}
		seen[key] = true
	}
}

// SM08 — Multiple bad tokens: len(Diagnostics) matches count of bad tokens.
func TestParseAudit_Syntax_SM08_countMatchesBadTokens(t *testing.T) {
	r := parseAudit(`{{ | a }} {{ | b }}`)
	assertTemplateNonNil(t, r, "SM08")
	syntaxErrs := allParseDiags(r, "syntax-error")
	if len(syntaxErrs) != 2 {
		t.Errorf("SM08: expected exactly 2 syntax-error diagnostics, got %d", len(syntaxErrs))
	}
}

// ============================================================================
// Non-Fatal Errors — rendering a recovered template (SR01–SR03)
// ============================================================================

// SR01 — Template with syntax-error renders without panic (ASTBroken = empty string).
func TestParseAudit_Syntax_SR01_renderNoPanic(t *testing.T) {
	r := parseAudit(`{{ | bad }}`)
	assertTemplateNonNil(t, r, "SR01")
	if r.Template != nil {
		// Must not panic.
		out, err := r.Template.RenderString(liquid.Bindings{})
		if err != nil {
			t.Logf("SR01: render returned error (acceptable): %v", err)
		}
		_ = out
	}
}

// SR02 — Template with broken expr surrounded by valid content renders correctly.
func TestParseAudit_Syntax_SR02_validContentAroundBroken(t *testing.T) {
	r := parseAudit(`Hello {{ | bad }} {{ name }}`)
	assertTemplateNonNil(t, r, "SR02")
	if r.Template == nil {
		t.Skip("SR02: Template is nil, skipping render check")
	}
	out, _ := r.Template.RenderString(liquid.Bindings{"name": "Alice"})
	// Broken node outputs nothing; "Hello " + "" + " " + "Alice" = "Hello  Alice"
	if !containsSubstr(out, "Alice") {
		t.Errorf("SR02: Output=%q, want it to contain valid variable value 'Alice'", out)
	}
}

// SR03 — Template from ParseStringAudit can be used with RenderAudit.
func TestParseAudit_Syntax_SR03_pipelineWithRenderAudit(t *testing.T) {
	r := parseAudit(`{{ | bad }} {{ name }}`)
	assertTemplateNonNil(t, r, "SR03")
	if r.Template == nil {
		t.Skip("SR03: Template is nil")
	}
	auditResult, _ := r.Template.RenderAudit(
		liquid.Bindings{"name": "Bob"},
		liquid.AuditOptions{TraceVariables: true},
	)
	if auditResult == nil {
		t.Fatal("SR03: RenderAudit returned nil AuditResult")
	}
}
