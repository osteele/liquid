package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// Multiple Diagnostics — Accumulation (M01–M11)
// ============================================================================

// M01 — undefined-filter + empty-block: both distinct codes in Diagnostics.
func TestParseAudit_Multi_M01_undefinedFilterAndEmptyBlock(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}{{ x | badfilter_m01 }}`)
	if firstParseDiag(r, "empty-block") == nil {
		t.Error("M01: expected empty-block diagnostic")
	}
	if firstParseDiag(r, "undefined-filter") == nil {
		t.Error("M01: expected undefined-filter diagnostic")
	}
}

// M02 — Two undefined-filter + one empty-block: len(Diagnostics)=3.
func TestParseAudit_Multi_M02_twoFiltersAndOneBlock(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}{{ x | bad_m02a }}{{ y | bad_m02b }}`)
	assertParseDiagCount(t, r, 3, "M02")
}

// M03 — syntax-error + undefined-filter: both codes present.
func TestParseAudit_Multi_M03_syntaxErrorAndUndefinedFilter(t *testing.T) {
	r := parseAudit(`{{ | bad_syntax }} {{ x | bad_filter_m03 }}`)
	assertParseResultNonNil(t, r, "M03")
	if firstParseDiag(r, "syntax-error") == nil {
		t.Error("M03: expected syntax-error diagnostic")
	}
	if firstParseDiag(r, "undefined-filter") == nil {
		t.Error("M03: expected undefined-filter diagnostic")
	}
}

// M04 — syntax-error + empty-block: both present.
func TestParseAudit_Multi_M04_syntaxErrorAndEmptyBlock(t *testing.T) {
	r := parseAudit(`{{ | bad_m04 }}{% if x %}{% endif %}`)
	assertParseResultNonNil(t, r, "M04")
	if firstParseDiag(r, "syntax-error") == nil {
		t.Error("M04: expected syntax-error diagnostic")
	}
	if firstParseDiag(r, "empty-block") == nil {
		t.Error("M04: expected empty-block diagnostic")
	}
}

// M05 — syntax-error + undefined-filter + empty-block: all three present.
func TestParseAudit_Multi_M05_allThreeCodes(t *testing.T) {
	r := parseAudit(`{{ | bad_m05 }}{% if x %}{% endif %}{{ z | unknown_m05 }}`)
	assertParseResultNonNil(t, r, "M05")
	if firstParseDiag(r, "syntax-error") == nil {
		t.Error("M05: expected syntax-error diagnostic")
	}
	if firstParseDiag(r, "empty-block") == nil {
		t.Error("M05: expected empty-block diagnostic")
	}
	if firstParseDiag(r, "undefined-filter") == nil {
		t.Error("M05: expected undefined-filter diagnostic")
	}
}

// M06 — Three undefined-filter for three different bad filters on different lines.
// Each must have a distinct Range.
func TestParseAudit_Multi_M06_threeFiltersDistinctRanges(t *testing.T) {
	r := parseAudit("{{ a | bad_a }}\n{{ b | bad_b }}\n{{ c | bad_c }}")
	filters := allParseDiags(r, "undefined-filter")
	if len(filters) != 3 {
		t.Fatalf("M06: expected 3 undefined-filter diagnostics, got %d", len(filters))
	}
	// All three should have distinct start positions.
	for i := 0; i < len(filters); i++ {
		for j := i + 1; j < len(filters); j++ {
			li := filters[i].Range.Start.Line
			lj := filters[j].Range.Start.Line
			if li == lj {
				t.Errorf("M06: diagnostics[%d] and diagnostics[%d] share line %d", i, j, li)
			}
		}
	}
}

// M07 — Two empty-blocks on separate blocks → exactly two empty-block diagnostics.
func TestParseAudit_Multi_M07_twoEmptyBlocks(t *testing.T) {
	r := parseAudit(`{% if a %}{% endif %}{% for x in items %}{% endfor %}`)
	blocks := allParseDiags(r, "empty-block")
	if len(blocks) != 2 {
		t.Errorf("M07: expected 2 empty-block diagnostics, got %d", len(blocks))
	}
}

// M08 — Template with clean and bad sections: only bad sections produce diagnostics.
func TestParseAudit_Multi_M08_onlyBadSectionsDiag(t *testing.T) {
	// "{{ name }}" is clean; "{{ x | nofilter_m08 }}" is the bad section.
	r := parseAudit(`Hello {{ name }}! {{ x | nofilter_m08 }}`)
	assertParseResultNonNil(t, r, "M08")
	// The clean {{ name }} should not produce any diagnostics.
	for _, d := range r.Diagnostics {
		if containsSubstr(d.Source, "name") && d.Code == "undefined-filter" {
			t.Errorf("M08: unexpected undefined-filter on clean '{{ name }}' expression")
		}
	}
	// The bad one should produce a diagnostic.
	if firstParseDiag(r, "undefined-filter") == nil {
		t.Error("M08: expected undefined-filter for 'nofilter_m08'")
	}
}

// M09 — Diagnostics are in source order (ascending by Range.Start.Line).
func TestParseAudit_Multi_M09_diagnosticsInSourceOrder(t *testing.T) {
	r := parseAudit("{{ a | bad_first }}\n{{ b | bad_second }}\n{{ c | bad_third }}")
	filters := allParseDiags(r, "undefined-filter")
	if len(filters) < 2 {
		t.Skip("M09: fewer than 2 undefined-filter diagnostics; skipping order check")
	}
	for i := 1; i < len(filters); i++ {
		prev := filters[i-1].Range.Start.Line
		curr := filters[i].Range.Start.Line
		if curr < prev {
			t.Errorf("M09: Diagnostics out of source order: diagnostics[%d].Line=%d > diagnostics[%d].Line=%d",
				i-1, prev, i, curr)
		}
	}
}

// M10 — Single fatal error template: exactly one diagnostic (not duplicated).
func TestParseAudit_Multi_M10_oneDiagnosticOnFatal(t *testing.T) {
	r := parseAudit(`{% if x %}`)
	assertParseResultNonNil(t, r, "M10")
	if len(r.Diagnostics) != 1 {
		t.Errorf("M10: expected exactly 1 diagnostic for simple unclosed-tag, got %d (codes: %v)",
			len(r.Diagnostics), parseDiagCodes(r.Diagnostics))
	}
}

// M11 — Fatal error: no spurious static analysis diagnostics (walk skipped when Template=nil).
func TestParseAudit_Multi_M11_noStaticDiagsOnFatal(t *testing.T) {
	// A template with unclosed-tag; if Template=nil we should not also get
	// empty-block or undefined-filter since the AST is not usable.
	r := parseAudit(`{% if x %}{{ y | nofilter_m11 }}`)
	assertTemplateNil(t, r, "M11")
	for _, d := range r.Diagnostics {
		if d.Code == "empty-block" || d.Code == "undefined-filter" {
			t.Errorf("M11: unexpected static analysis diagnostic code=%q on fatal-error template", d.Code)
		}
	}
}

// ============================================================================
// Diagnostic Field Completeness (DF01–DF15)
// ============================================================================

// DF01 — Every Diagnostic has a non-empty Code.
func TestParseAudit_Fields_DF01_allHaveCode(t *testing.T) {
	// Exercise multiple paths to generate diagnostics.
	templates := []string{
		`{% if x %}unclosed`,
		`{% endif %}`,
		`{{ | bad }}`,
		`{{ x | nofilter_df01 }}`,
		`{% if true %}{% endif %}`,
	}
	for _, src := range templates {
		r := parseAudit(src)
		for i, d := range r.Diagnostics {
			if d.Code == "" {
				t.Errorf("DF01 src=%q: Diagnostics[%d].Code is empty", src, i)
			}
		}
	}
}

// DF02 — Every Diagnostic has Severity in {"error","warning","info"}.
func TestParseAudit_Fields_DF02_allHaveValidSeverity(t *testing.T) {
	validSeverities := map[liquid.DiagnosticSeverity]bool{
		liquid.SeverityError:   true,
		liquid.SeverityWarning: true,
		liquid.SeverityInfo:    true,
	}
	templates := []string{
		`{% if x %}unclosed`,
		`{% endif %}`,
		`{{ | bad }}`,
		`{{ x | nofilt_df02 }}`,
		`{% if true %}{% endif %}`,
	}
	for _, src := range templates {
		r := parseAudit(src)
		for i, d := range r.Diagnostics {
			if !validSeverities[d.Severity] {
				t.Errorf("DF02 src=%q: Diagnostics[%d].Severity=%q is not a valid severity", src, i, d.Severity)
			}
		}
	}
}

// DF03 — Every Diagnostic has a non-empty Message.
func TestParseAudit_Fields_DF03_allHaveMessage(t *testing.T) {
	templates := []string{
		`{% if x %}unclosed`,
		`{% endif %}`,
		`{{ | bad }}`,
		`{{ x | nofilt_df03 }}`,
		`{% if true %}{% endif %}`,
	}
	for _, src := range templates {
		r := parseAudit(src)
		for i, d := range r.Diagnostics {
			if d.Message == "" {
				t.Errorf("DF03 src=%q: Diagnostics[%d].Message is empty (Code=%q)", src, i, d.Code)
			}
		}
	}
}

// DF04 — Every Diagnostic has a non-empty Source.
func TestParseAudit_Fields_DF04_allHaveSource(t *testing.T) {
	templates := []string{
		`{% if x %}unclosed`,
		`{% endif %}`,
		`{{ | bad }}`,
		`{{ x | nofilt_df04 }}`,
		`{% if true %}{% endif %}`,
	}
	for _, src := range templates {
		r := parseAudit(src)
		for i, d := range r.Diagnostics {
			if d.Source == "" {
				t.Errorf("DF04 src=%q: Diagnostics[%d].Source is empty (Code=%q)", src, i, d.Code)
			}
		}
	}
}

// DF05 — Every Diagnostic has Range.Start.Line >= 1.
func TestParseAudit_Fields_DF05_allHaveValidStartLine(t *testing.T) {
	templates := []string{
		`{% if x %}unclosed`,
		`{% endif %}`,
		`{{ | bad }}`,
		`{{ x | nofilt_df05 }}`,
		`{% if true %}{% endif %}`,
	}
	for _, src := range templates {
		r := parseAudit(src)
		for i, d := range r.Diagnostics {
			if d.Range.Start.Line < 1 {
				t.Errorf("DF05 src=%q: Diagnostics[%d].Range.Start.Line=%d (Code=%q), want >= 1",
					src, i, d.Range.Start.Line, d.Code)
			}
		}
	}
}

// DF06 — Every Diagnostic has Range.Start.Column >= 1.
func TestParseAudit_Fields_DF06_allHaveValidStartColumn(t *testing.T) {
	templates := []string{
		`{% if x %}unclosed`,
		`{{ | bad }}`,
		`{{ x | nofilt_df06 }}`,
		`{% if true %}{% endif %}`,
	}
	for _, src := range templates {
		r := parseAudit(src)
		for i, d := range r.Diagnostics {
			if d.Range.Start.Column < 1 {
				t.Errorf("DF06 src=%q: Diagnostics[%d].Range.Start.Column=%d (Code=%q), want >= 1",
					src, i, d.Range.Start.Column, d.Code)
			}
		}
	}
}

// DF07 — Every Diagnostic has Range.End.Line >= Range.Start.Line.
func TestParseAudit_Fields_DF07_endNotBeforeStart(t *testing.T) {
	templates := []string{
		`{% if x %}unclosed`,
		`{{ | bad }}`,
		`{{ x | nofilt_df07 }}`,
		`{% if true %}{% endif %}`,
	}
	for _, src := range templates {
		r := parseAudit(src)
		for i, d := range r.Diagnostics {
			if d.Range.End.Line < d.Range.Start.Line {
				t.Errorf("DF07 src=%q: Diagnostics[%d] Range.End.Line=%d < Range.Start.Line=%d (Code=%q)",
					src, i, d.Range.End.Line, d.Range.Start.Line, d.Code)
			}
		}
	}
}

// DF08 — error-severity diagnostics have Severity="error".
func TestParseAudit_Fields_DF08_errorCodesHaveErrorSeverity(t *testing.T) {
	errorCodes := []struct {
		src  string
		code string
	}{
		{`{% if x %}unclosed`, "unclosed-tag"},
		{`{% endif %}`, "unexpected-tag"},
		{`{{ | bad_df08 }}`, "syntax-error"},
		{`{{ x | nofilt_df08 }}`, "undefined-filter"},
	}
	for _, tc := range errorCodes {
		r := parseAudit(tc.src)
		d := firstParseDiag(r, tc.code)
		if d == nil {
			t.Logf("DF08: code=%q not present for src=%q", tc.code, tc.src)
			continue
		}
		if d.Severity != liquid.SeverityError {
			t.Errorf("DF08 code=%q: Severity=%q, want %q", tc.code, d.Severity, liquid.SeverityError)
		}
	}
}

// DF09 — info-severity diagnostics have Severity="info".
func TestParseAudit_Fields_DF09_infoCodesHaveInfoSeverity(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}`)
	d := requireParseDiag(t, r, "empty-block")
	if d.Severity != liquid.SeverityInfo {
		t.Errorf("DF09 code=empty-block: Severity=%q, want %q", d.Severity, liquid.SeverityInfo)
	}
}

// DF10 — unclosed-tag: Related is non-nil and non-empty.
func TestParseAudit_Fields_DF10_unclosedTagRelatedNonEmpty(t *testing.T) {
	r := parseAudit(`{% if x %}`)
	d := requireParseDiag(t, r, "unclosed-tag")
	if d.Related == nil || len(d.Related) == 0 {
		t.Fatal("DF10: unclosed-tag Related is nil/empty; expected at least one related entry")
	}
}

// DF11 — unclosed-tag Related[0].Range.Start.Line >= 1.
func TestParseAudit_Fields_DF11_unclosedTagRelatedRangeValid(t *testing.T) {
	r := parseAudit(`{% if x %}`)
	d := requireParseDiag(t, r, "unclosed-tag")
	if len(d.Related) == 0 {
		t.Skip("DF11: no Related entries")
	}
	if d.Related[0].Range.Start.Line < 1 {
		t.Errorf("DF11: Related[0].Range.Start.Line=%d, want >= 1", d.Related[0].Range.Start.Line)
	}
}

// DF12 — unclosed-tag Related[0].Message is non-empty.
func TestParseAudit_Fields_DF12_unclosedTagRelatedMessageNonEmpty(t *testing.T) {
	r := parseAudit(`{% if x %}`)
	d := requireParseDiag(t, r, "unclosed-tag")
	if len(d.Related) == 0 {
		t.Skip("DF12: no Related entries")
	}
	if d.Related[0].Message == "" {
		t.Fatal("DF12: Related[0].Message is empty; should describe expected closing tag")
	}
}

// DF13 — syntax-error: Related field is nil or empty (not used for expression errors).
func TestParseAudit_Fields_DF13_syntaxErrorNoRelated(t *testing.T) {
	r := parseAudit(`{{ | bad_df13 }}`)
	d := requireParseDiag(t, r, "syntax-error")
	if len(d.Related) > 0 {
		t.Errorf("DF13: syntax-error has unexpected Related entries: %v", d.Related)
	}
}

// DF14 — undefined-filter: Related field is nil or empty.
func TestParseAudit_Fields_DF14_undefinedFilterNoRelated(t *testing.T) {
	r := parseAudit(`{{ x | nofilt_df14 }}`)
	d := requireParseDiag(t, r, "undefined-filter")
	if len(d.Related) > 0 {
		t.Errorf("DF14: undefined-filter has unexpected Related entries: %v", d.Related)
	}
}

// DF15 — empty-block: Related field is nil or empty.
func TestParseAudit_Fields_DF15_emptyBlockNoRelated(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}`)
	d := requireParseDiag(t, r, "empty-block")
	if len(d.Related) > 0 {
		t.Errorf("DF15: empty-block has unexpected Related entries: %v", d.Related)
	}
}
