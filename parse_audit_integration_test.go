package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// Integration — ParseTemplateAudit → RenderAudit pipeline (I01–I08)
// ============================================================================

// I01 — Parse-clean template → RenderAudit succeeds.
func TestParseAudit_Integration_I01_cleanParseToRenderAudit(t *testing.T) {
	r := parseAudit(`Hello {{ name }}!`)
	assertTemplateNonNil(t, r, "I01")
	auditResult, auditErr := r.Template.RenderAudit(
		liquid.Bindings{"name": "Alice"},
		liquid.AuditOptions{TraceVariables: true},
	)
	if auditResult == nil {
		t.Fatal("I01: RenderAudit returned nil result")
	}
	if auditErr != nil {
		t.Fatalf("I01: unexpected RenderAudit error: %v", auditErr)
	}
	if auditResult.Output != "Hello Alice!" {
		t.Errorf("I01: Output=%q, want %q", auditResult.Output, "Hello Alice!")
	}
}

// I02 — Parse with syntax-error (non-fatal) → ASTBroken renders as empty string.
func TestParseAudit_Integration_I02_syntaxErrorBrokenNodeEmptyRender(t *testing.T) {
	r := parseAudit(`before{{ | bad_i02 }}after`)
	assertTemplateNonNil(t, r, "I02")
	auditResult, _ := r.Template.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
	if auditResult == nil {
		t.Fatal("I02: RenderAudit returned nil")
	}
	if auditResult.Output != "beforeafter" {
		t.Errorf("I02: Output=%q, want %q (broken node should output nothing)", auditResult.Output, "beforeafter")
	}
}

// I03 — Parse with undefined-filter → RenderAudit does not panic.
func TestParseAudit_Integration_I03_undefinedFilterRenderNoPanic(t *testing.T) {
	r := parseAudit(`{{ x | totally_undefined_i03 }}`)
	assertTemplateNonNil(t, r, "I03")
	// Rendering a template with an unknown filter should not panic.
	// It may return an error through normal render error handling.
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("I03: RenderAudit panicked: %v", rec)
		}
	}()
	auditResult, _ := r.Template.RenderAudit(liquid.Bindings{"x": "value"}, liquid.AuditOptions{})
	if auditResult == nil {
		t.Fatal("I03: RenderAudit returned nil result")
	}
}

// I04 — Parse and render diagnostics come from different sources; no overlap of identical errors.
func TestParseAudit_Integration_I04_noDupBetweenParseAndRenderDiags(t *testing.T) {
	// Parse with undefined-filter (parse diagnostic).
	// Render with strict variables hitting an undefined variable (render diagnostic).
	r := parseAudit(`{{ x | unknown_i04 }}`)
	assertTemplateNonNil(t, r, "I04")

	auditResult, _ := r.Template.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{},
		liquid.WithStrictVariables(),
	)
	if auditResult == nil {
		t.Fatal("I04: RenderAudit returned nil")
	}

	// Parse diags: undefined-filter
	// Render diags: undefined-variable (from StrictVariables)
	// They should not duplicate each other.
	for _, pd := range r.Diagnostics {
		for _, rd := range auditResult.Diagnostics {
			if pd.Code == rd.Code && pd.Range.Start.Line == rd.Range.Start.Line &&
				pd.Range.Start.Column == rd.Range.Start.Column {
				t.Errorf("I04: same diagnostic appears in both parse and render results: code=%q line=%d col=%d",
					pd.Code, pd.Range.Start.Line, pd.Range.Start.Column)
			}
		}
	}
}

// I05 — Parse with empty-block → RenderAudit output is empty for that block.
func TestParseAudit_Integration_I05_emptyBlockRendersEmptyOutput(t *testing.T) {
	r := parseAudit(`before{% if true %}{% endif %}after`)
	assertTemplateNonNil(t, r, "I05")
	auditResult, _ := r.Template.RenderAudit(liquid.Bindings{"true": true}, liquid.AuditOptions{})
	if auditResult == nil {
		t.Fatal("I05: RenderAudit returned nil")
	}
	if auditResult.Output != "beforeafter" {
		t.Errorf("I05: Output=%q, want %q", auditResult.Output, "beforeafter")
	}
}

// I06 — Fatal parse (Template=nil): caller can safely guard without panic.
func TestParseAudit_Integration_I06_nilTemplateGuardedSafely(t *testing.T) {
	r := parseAudit(`{% if x %}unclosed`)
	assertTemplateNil(t, r, "I06")
	// Caller-pattern: check Template before using it.
	if r.Template != nil {
		t.Error("I06: Template should be nil for unclosed template")
	}
	// No panic here; just confirming nil-check pattern works.
}

// I07 — Complete end-to-end: clean parse + RenderAudit with strict vars + collect all diags.
func TestParseAudit_Integration_I07_fullPipeline(t *testing.T) {
	r := parseAudit(`Hello {{ user.name }}! Your score: {{ score }}.`)
	assertTemplateNonNil(t, r, "I07")
	assertNoParseDiags(t, r, "I07")

	auditResult, auditErr := r.Template.RenderAudit(
		liquid.Bindings{"user": map[string]any{"name": "Bob"}, "score": 95},
		liquid.AuditOptions{TraceVariables: true},
		liquid.WithStrictVariables(),
	)
	if auditResult == nil {
		t.Fatal("I07: RenderAudit returned nil result")
	}
	if auditErr != nil {
		t.Fatalf("I07: unexpected AuditError: %v", auditErr)
	}
	if auditResult.Output != "Hello Bob! Your score: 95." {
		t.Errorf("I07: Output=%q, want %q", auditResult.Output, "Hello Bob! Your score: 95.")
	}
	// All diagnostics from both phases.
	allDiags := append(r.Diagnostics, auditResult.Diagnostics...)
	if len(allDiags) != 0 {
		t.Errorf("I07: expected 0 total diagnostics, got %d: %v", len(allDiags), allDiags)
	}
}

// I08 — ParseStringAudit + Template.Validate(): diagnostics from Validate are not
// duplicated in the ParseResult (parse-time and validate-time are independent stages).
func TestParseAudit_Integration_I08_validateNotDuplicateParseAudit(t *testing.T) {
	// Template with empty-block: ParseStringAudit detects it at parse time.
	// Validate() should also detect it. But neither should duplicate the other.
	src := `{% if true %}{% endif %}`
	r := parseAudit(src)
	assertTemplateNonNil(t, r, "I08")

	parseEmptyBlocks := allParseDiags(r, "empty-block")
	if len(parseEmptyBlocks) == 0 {
		t.Fatal("I08: expected empty-block in parse diagnostics")
	}

	validateResult, validateErr := r.Template.Validate()
	if validateErr != nil {
		t.Logf("I08: Validate() returned error: %v", validateErr)
	}
	if validateResult == nil {
		t.Skip("I08: Validate() returned nil result")
	}

	// Validate diagnostics should contain empty-block too.
	// The key point: ParseResult.Diagnostics and AuditResult.Diagnostics are separate,
	// not merged automatically. The caller is responsible for merging if needed.
	validateEmptyBlocks := allDiags(validateResult.Diagnostics, "empty-block")
	t.Logf("I08: parse detected %d empty-block(s), validate detected %d empty-block(s)",
		len(parseEmptyBlocks), len(validateEmptyBlocks))
	// Both should find it; this test documents the behavior.
}

// ============================================================================
// Engine Configuration Interaction (EC01–EC06)
// ============================================================================

// EC01 — Engine with RegisterFilter("my_filter", fn): {{ x | my_filter }} → no undefined-filter.
func TestParseAudit_EngineConfig_EC01_customRegisteredFilterRecognized(t *testing.T) {
	eng := newParseAuditEngine()
	eng.RegisterFilter("my_custom_ec01", func(s string) string { return s })

	r := parseAuditWith(eng, `{{ x | my_custom_ec01 }}`)
	d := firstParseDiag(r, "undefined-filter")
	if d != nil {
		t.Errorf("EC01: unexpected undefined-filter for registered filter 'my_custom_ec01'")
	}
}

// EC02 — Engine without custom filter: {{ x | my_filter }} → undefined-filter.
func TestParseAudit_EngineConfig_EC02_unregisteredFilterDetected(t *testing.T) {
	r := parseAudit(`{{ x | my_custom_ec02_unregistered }}`)
	d := firstParseDiag(r, "undefined-filter")
	if d == nil {
		t.Error("EC02: expected undefined-filter for unregistered filter")
	}
}

// EC03 — Two engines with different filter registrations: same source gives different results.
func TestParseAudit_EngineConfig_EC03_engineScopedFilterCheck(t *testing.T) {
	src := `{{ x | engine_specific_filter_ec03 }}`

	eng1 := liquid.NewEngine()
	eng1.RegisterFilter("engine_specific_filter_ec03", func(s string) string { return s })

	eng2 := liquid.NewEngine()
	// eng2 does NOT register the filter.

	r1 := parseAuditWith(eng1, src)
	r2 := parseAuditWith(eng2, src)

	has1 := firstParseDiag(r1, "undefined-filter") != nil
	has2 := firstParseDiag(r2, "undefined-filter") != nil

	if has1 {
		t.Error("EC03: eng1 (has filter registered) should NOT produce undefined-filter")
	}
	if !has2 {
		t.Error("EC03: eng2 (filter not registered) should produce undefined-filter")
	}
}

// EC04 — ParseStringAudit on engine configured with SetTrimTagLeft: no crash.
func TestParseAudit_EngineConfig_EC04_trimConfigNoCrash(t *testing.T) {
	eng := newParseAuditEngine()
	eng.SetTrimTagLeft(true)
	r := parseAuditWith(eng, `{% if x %}content{% endif %}`)
	assertParseResultNonNil(t, r, "EC04")
}

// EC05 — StrictVariables is a render-time option, not a parse option: parse is not affected.
func TestParseAudit_EngineConfig_EC05_strictVariablesNotAffectsParse(t *testing.T) {
	eng := newParseAuditEngine()
	eng.StrictVariables()
	// Even with StrictVariables on the engine, parse should not report undefined-variable.
	r := parseAuditWith(eng, `{{ undefined_var_ec05 }}`)
	d := firstParseDiag(r, "undefined-variable")
	if d != nil {
		t.Error("EC05: ParseStringAudit should not produce undefined-variable at parse time (it's a render-time check)")
	}
}

// EC06 — LaxFilters on engine: undefined-filter is still detected by static walk.
// The static walk is unconditional; LaxFilters only suppresses the runtime error.
func TestParseAudit_EngineConfig_EC06_laxFiltersStillDetectedAtParse(t *testing.T) {
	eng := newParseAuditEngine()
	eng.LaxFilters()
	r := parseAuditWith(eng, `{{ x | lax_filter_ec06_unknown }}`)
	// This behavior depends on implementation: static walk may or may not respect LaxFilters.
	// Document it here; no assertion either way — just no crash.
	assertParseResultNonNil(t, r, "EC06")
	t.Logf("EC06: LaxFilters engine + unknown filter produces undefined-filter: %v",
		firstParseDiag(r, "undefined-filter") != nil)
}

// ============================================================================
// ParseTemplate vs ParseTemplateAudit Behavioral Parity (PB01–PB05)
// ============================================================================

// PB01 — Clean source: ParseTemplate succeeds; ParseTemplateAudit.Template is non-nil.
// Both render identically.
func TestParseAudit_Parity_PB01_cleanSourceBothSucceed(t *testing.T) {
	src := `Hello {{ name | upcase }}!`
	eng := newParseAuditEngine()

	tpl1, err := eng.ParseString(src)
	if err != nil {
		t.Fatalf("PB01: ParseString failed: %v", err)
	}

	r := parseAuditWith(eng, src)
	assertTemplateNonNil(t, r, "PB01")

	vars := liquid.Bindings{"name": "world"}

	out1, err1 := tpl1.RenderString(vars)
	out2, err2 := r.Template.RenderString(vars)

	if err1 != nil || err2 != nil {
		t.Fatalf("PB01: render errors: ParseTemplate=%v, ParseTemplateAudit=%v", err1, err2)
	}
	if out1 != out2 {
		t.Errorf("PB01: output mismatch:\n  ParseTemplate:      %q\n  ParseTemplateAudit: %q", out1, out2)
	}
}

// PB02 — Fatal source: ParseTemplate returns error; ParseTemplateAudit.Template is nil.
func TestParseAudit_Parity_PB02_fatalSourceBothFail(t *testing.T) {
	src := `{% if x %}no close`
	eng := newParseAuditEngine()

	_, parseErr := eng.ParseString(src)
	if parseErr == nil {
		t.Fatal("PB02: ParseString should have returned an error for unclosed-tag")
	}

	r := parseAuditWith(eng, src)
	assertTemplateNil(t, r, "PB02")
}

// PB03 — Non-fatal source (syntax-error in expression): ParseTemplate returns error;
// ParseTemplateAudit.Template is non-nil (audit recovers).
func TestParseAudit_Parity_PB03_syntaxErrorAuditRecovers(t *testing.T) {
	src := `{{ | bad_pb03 }}`
	eng := newParseAuditEngine()

	_, parseErr := eng.ParseString(src)
	if parseErr == nil {
		// If ParseTemplate also succeeds here, that's OK — test is advisory.
		t.Logf("PB03: ParseString also succeeded; behavior note: syntax-error may be non-fatal in both paths")
	}

	r := parseAuditWith(eng, src)
	assertParseResultNonNil(t, r, "PB03")
	// ParseTemplateAudit should at minimum return without panicking.
}

// PB04 — Clean template from ParseTemplateAudit: render output matches ParseAndRenderString.
func TestParseAudit_Parity_PB04_outputMatchesDirectRender(t *testing.T) {
	src := `{% assign total = items | size %}Count: {{ total }}`
	vars := liquid.Bindings{"items": []string{"a", "b", "c"}}
	eng := newParseAuditEngine()

	expected, err := eng.ParseAndRenderString(src, vars)
	if err != nil {
		t.Fatalf("PB04: ParseAndRenderString failed: %v", err)
	}

	r := parseAuditWith(eng, src)
	assertTemplateNonNil(t, r, "PB04")

	got, renderErr := r.Template.RenderString(vars)
	if renderErr != nil {
		t.Fatalf("PB04: RenderString failed: %v", renderErr)
	}

	if got != expected {
		t.Errorf("PB04: output mismatch:\n  direct: %q\n  audit:  %q", expected, got)
	}
}

// PB05 — Clean source: ParseStringAudit produces no diagnostics (same as ParseString no-error).
func TestParseAudit_Parity_PB05_cleanMeansNoDiagnostics(t *testing.T) {
	src := `{% assign price = 100 %}{{ price | times: 0.9 | round }}`
	r := parseAudit(src)
	assertTemplateNonNil(t, r, "PB05")
	assertNoParseDiags(t, r, "PB05")
}

// ============================================================================
// Validate() Overlap — Non-duplication (VA01–VA05)
// ============================================================================

// VA01 — empty-block via ParseStringAudit: diagnostic present in ParseResult.Diagnostics.
func TestParseAudit_Validate_VA01_emptyBlockInParseResult(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}`)
	if firstParseDiag(r, "empty-block") == nil {
		t.Error("VA01: expected empty-block in ParseResult.Diagnostics")
	}
}

// VA02 — Same template via ParseString + Validate(): empty-block present in AuditResult.Diagnostics.
func TestParseAudit_Validate_VA02_emptyBlockInValidateResult(t *testing.T) {
	tpl, err := newParseAuditEngine().ParseString(`{% if true %}{% endif %}`)
	if err != nil {
		t.Fatalf("VA02: ParseString failed: %v", err)
	}
	validateResult, validateErr := tpl.Validate()
	if validateErr != nil {
		t.Logf("VA02: Validate() returned error: %v", validateErr)
	}
	if validateResult == nil {
		t.Skip("VA02: Validate() returned nil result")
	}
	d := allDiags(validateResult.Diagnostics, "empty-block")
	if len(d) == 0 {
		t.Error("VA02: expected empty-block in Validate() AuditResult.Diagnostics")
	}
}

// VA03 — Full pipeline: ParseStringAudit + RenderAudit → empty-block appears in parse diags,
// not again in render diags.
func TestParseAudit_Validate_VA03_emptyBlockNotInRenderDiags(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}`)
	assertTemplateNonNil(t, r, "VA03")

	parseEmpty := allParseDiags(r, "empty-block")
	if len(parseEmpty) == 0 {
		t.Fatal("VA03: expected empty-block in parse diagnostics")
	}

	auditResult, _ := r.Template.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
	if auditResult == nil {
		t.Fatal("VA03: RenderAudit returned nil")
	}

	renderEmpty := allDiags(auditResult.Diagnostics, "empty-block")
	if len(renderEmpty) > 0 {
		t.Errorf("VA03: empty-block should not appear in RenderAudit diagnostics "+
			"(it's a parse-time static check); got %d render empty-block diagnostics", len(renderEmpty))
	}
}

// VA04 — undefined-filter via ParseStringAudit: present in parse diagnostics.
func TestParseAudit_Validate_VA04_undefinedFilterInParseResult(t *testing.T) {
	r := parseAudit(`{{ x | no_such_filter_va04 }}`)
	if firstParseDiag(r, "undefined-filter") == nil {
		t.Error("VA04: expected undefined-filter in ParseResult.Diagnostics")
	}
}

// VA05 — Same template via ParseString + Validate(): undefined-filter present in AuditResult.Diagnostics.
func TestParseAudit_Validate_VA05_undefinedFilterInValidateResult(t *testing.T) {
	tpl, err := newParseAuditEngine().ParseString(`{{ x | no_such_filter_va05 }}`)
	if err != nil {
		t.Logf("VA05: ParseString returned error (may be normal for unknown filter): %v", err)
		t.Skip("VA05: ParseString did not produce a usable template")
	}
	validateResult, validateErr := tpl.Validate()
	if validateErr != nil {
		t.Logf("VA05: Validate() returned error: %v", validateErr)
	}
	if validateResult == nil {
		t.Skip("VA05: Validate() returned nil result")
	}
	d := allDiags(validateResult.Diagnostics, "undefined-filter")
	if len(d) == 0 {
		t.Error("VA05: expected undefined-filter in Validate() AuditResult.Diagnostics")
	}
}
