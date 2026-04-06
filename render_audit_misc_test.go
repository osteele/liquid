package liquid_test

import (
	"encoding/json"
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// AuditOptions — Flag Isolation (O01–O09)
// ============================================================================

// O01 — all flags false: Expressions is empty.
func TestRenderAudit_Options_O01_allFlagsOff(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "hi" %}{{ x }}{% if true %}yes{% endif %}{% for i in items %}{{ i }}{% endfor %}`)
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2}},
		liquid.AuditOptions{}, // all false
	)
	if len(result.Expressions) != 0 {
		t.Errorf("Expressions=%d, want 0 when all trace flags are false", len(result.Expressions))
	}
}

// O02 — only TraceVariables: only KindVariable expressions.
func TestRenderAudit_Options_O02_onlyVariables(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "hi" %}{{ x }}{% if true %}yes{% endif %}{% for i in items %}{{ i }}{% endfor %}`)
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1}},
		liquid.AuditOptions{TraceVariables: true},
	)
	for i, e := range result.Expressions {
		if e.Kind != liquid.KindVariable {
			t.Errorf("Expressions[%d].Kind=%q, want variable (only TraceVariables set)", i, e.Kind)
		}
	}
}

// O03 — only TraceConditions: only KindCondition expressions.
func TestRenderAudit_Options_O03_onlyConditions(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "hi" %}{{ x }}{% if true %}yes{% endif %}`)
	result := auditOK(t, tpl,
		liquid.Bindings{},
		liquid.AuditOptions{TraceConditions: true},
	)
	for i, e := range result.Expressions {
		if e.Kind != liquid.KindCondition {
			t.Errorf("Expressions[%d].Kind=%q, want condition (only TraceConditions set)", i, e.Kind)
		}
	}
}

// O04 — only TraceIterations: only KindIteration expressions.
func TestRenderAudit_Options_O04_onlyIterations(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "hi" %}{{ x }}{% for i in items %}{{ i }}{% endfor %}`)
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1}},
		liquid.AuditOptions{TraceIterations: true},
	)
	for i, e := range result.Expressions {
		if e.Kind != liquid.KindIteration {
			t.Errorf("Expressions[%d].Kind=%q, want iteration (only TraceIterations set)", i, e.Kind)
		}
	}
}

// O05 — only TraceAssignments: KindAssignment and KindCapture, no others.
func TestRenderAudit_Options_O05_onlyAssignments(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "hi" %}{{ x }}{% capture y %}hi{% endcapture %}{% if true %}yes{% endif %}`)
	result := auditOK(t, tpl,
		liquid.Bindings{},
		liquid.AuditOptions{TraceAssignments: true},
	)
	for i, e := range result.Expressions {
		if e.Kind != liquid.KindAssignment && e.Kind != liquid.KindCapture {
			t.Errorf("Expressions[%d].Kind=%q, want assignment or capture (only TraceAssignments set)", i, e.Kind)
		}
	}
}

// O06 — all flags true: all kinds of expressions appear in a rich template.
func TestRenderAudit_Options_O06_allFlagsOn(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "hi" %}{{ x }}{% if true %}yes{% endif %}{% for i in items %}{{ i }}{% endfor %}{% capture z %}cap{% endcapture %}`)
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1}},
		liquid.AuditOptions{
			TraceVariables:   true,
			TraceConditions:  true,
			TraceIterations:  true,
			TraceAssignments: true,
		},
	)
	kinds := make(map[liquid.ExpressionKind]bool)
	for _, e := range result.Expressions {
		kinds[e.Kind] = true
	}
	expectedKinds := []liquid.ExpressionKind{
		liquid.KindVariable,
		liquid.KindCondition,
		liquid.KindIteration,
		liquid.KindAssignment,
		liquid.KindCapture,
	}
	for _, k := range expectedKinds {
		if !kinds[k] {
			t.Errorf("missing expression kind %q in all-flags-on audit", k)
		}
	}
}

// O07 — Diagnostics are always collected regardless of trace flags.
func TestRenderAudit_Options_O07_diagnosticsAlwaysCollected(t *testing.T) {
	tpl := mustParseAudit(t, "{{ 10 | divided_by: 0 }}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{ /* all flags false */ })
	if len(result.Diagnostics) == 0 {
		t.Error("Diagnostics should be collected even when all trace flags are false")
	}
}

// O08 — MaxIterationTraceItems=0 with all flags → no truncation.
func TestRenderAudit_Options_O08_maxIterUnlimited(t *testing.T) {
	tpl := mustParseAudit(t, "{% for i in items %}{{ i }}{% endfor %}")
	items := make([]int, 50)
	result := auditOK(t, tpl,
		liquid.Bindings{"items": items},
		liquid.AuditOptions{
			TraceIterations:        true,
			MaxIterationTraceItems: 0,
		},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Truncated {
		t.Error("Truncated should be false when MaxIterationTraceItems=0 (unlimited)")
	}
}

// O09 — trace flags do not affect Output correctness.
func TestRenderAudit_Options_O09_flagsDontAffectOutput(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "hi" %}{{ x }}{% if true %}yes{% endif %}{% for i in items %}{{ i }}{% endfor %}`)
	bindings := liquid.Bindings{"items": []int{1, 2}}

	// Get expected output without audit.
	eng := newAuditEngine()
	expected, se := eng.ParseAndRenderString(
		`{% assign x = "hi" %}{{ x }}{% if true %}yes{% endif %}{% for i in items %}{{ i }}{% endfor %}`,
		bindings,
	)
	if se != nil {
		t.Fatalf("baseline render error: %v", se)
	}

	// Render with all flags on should produce the same output.
	result := auditOK(t, tpl, bindings,
		liquid.AuditOptions{
			TraceVariables:   true,
			TraceConditions:  true,
			TraceIterations:  true,
			TraceAssignments: true,
		},
	)
	if result.Output != expected {
		t.Errorf("Output with audit=%q, want %q (identical to Render)", result.Output, expected)
	}
}

// ============================================================================
// AuditResult — Output (R01–R04)
// ============================================================================

// R01 — Output matches Render for a simple template.
func TestRenderAudit_Result_R01_outputMatchesRender(t *testing.T) {
	src := "Hello, {{ name }}!"
	bindings := liquid.Bindings{"name": "World"}
	eng := newAuditEngine()
	expected, _ := eng.ParseAndRenderString(src, bindings)

	tpl := mustParseAuditWith(t, eng, src)
	result := auditOK(t, tpl, bindings, liquid.AuditOptions{})
	if result.Output != expected {
		t.Errorf("Output=%q, want %q", result.Output, expected)
	}
}

// R02 — Output matches Render for a complex template with assign, for, if.
func TestRenderAudit_Result_R02_complexOutputMatchesRender(t *testing.T) {
	src := `{% assign greeting = "Hello" %}{% for name in names %}{% if name == "Alice" %}{{ greeting }}, {{ name }}!{% else %}Hi, {{ name }}.{% endif %}{% endfor %}`
	bindings := liquid.Bindings{"names": []string{"Alice", "Bob"}}

	eng := newAuditEngine()
	expected, _ := eng.ParseAndRenderString(src, bindings)

	tpl := mustParseAuditWith(t, eng, src)
	result := auditOK(t, tpl, bindings,
		liquid.AuditOptions{
			TraceVariables:   true,
			TraceConditions:  true,
			TraceIterations:  true,
			TraceAssignments: true,
		},
	)
	if result.Output != expected {
		t.Errorf("Output=%q, want %q", result.Output, expected)
	}
}

// R03 — AuditResult is always non-nil even on error; Output may be partial.
// Note: divided_by:0 filter errors are captured as Diagnostics; they do NOT produce
// an AuditError — the render continues and emits partial output.
func TestRenderAudit_Result_R03_nonNilOnError(t *testing.T) {
	tpl := mustParseAudit(t, "before{{ 10 | divided_by: 0 }}after")
	result, ae := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
	if result == nil {
		t.Fatal("AuditResult must never be nil")
	}
	// ae is nil for filter errors (captured as Diagnostics only).
	_ = ae
	// Output should contain at least the non-error text.
	if result.Output != "beforeafter" {
		t.Errorf("Output=%q, want \"beforeafter\" (error part skipped)", result.Output)
	}
	// Diagnostic should be present.
	if len(result.Diagnostics) == 0 {
		t.Error("expected at least one Diagnostic for divided_by:0")
	}
}

// R04 — empty template produces empty Output and zero Expressions.
func TestRenderAudit_Result_R04_emptyTemplate(t *testing.T) {
	tpl := mustParseAudit(t, "")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceVariables: true})
	if result.Output != "" {
		t.Errorf("Output=%q, want empty for empty template", result.Output)
	}
	if len(result.Expressions) != 0 {
		t.Errorf("Expressions=%d, want 0 for empty template", len(result.Expressions))
	}
}

// ============================================================================
// AuditResult — Expressions Ordering (RO01–RO05)
// ============================================================================

// RO01 — assign appears before variable in execution order.
func TestRenderAudit_Result_RO01_assignBeforeVariable(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign msg = "hello" %}{{ msg }}`)
	result := auditOK(t, tpl, liquid.Bindings{},
		liquid.AuditOptions{TraceAssignments: true, TraceVariables: true},
	)
	if len(result.Expressions) < 2 {
		t.Fatalf("expected >= 2 expressions")
	}
	if result.Expressions[0].Kind != liquid.KindAssignment {
		t.Errorf("Expressions[0].Kind=%q, want assignment", result.Expressions[0].Kind)
	}
	if result.Expressions[1].Kind != liquid.KindVariable {
		t.Errorf("Expressions[1].Kind=%q, want variable", result.Expressions[1].Kind)
	}
}

// RO02 — for loop: inner variable expressions are emitted BEFORE the iteration's final trace.
// Ordering: 3 × KindVariable (one per iteration), then 1 × KindIteration (summary at the end).
func TestRenderAudit_Result_RO02_forLoopLinearized(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"a", "b", "c"}},
		liquid.AuditOptions{TraceIterations: true, TraceVariables: true},
	)
	// Pattern: 3 × KindVariable (iterations emit body traces first), then KindIteration.
	if len(result.Expressions) < 4 {
		t.Fatalf("expected >= 4 expressions, got %d", len(result.Expressions))
	}
	// Variables come first (body traces from iterations).
	for i := range 3 {
		if result.Expressions[i].Kind != liquid.KindVariable {
			t.Errorf("Expressions[%d].Kind=%q, want variable", i, result.Expressions[i].Kind)
		}
	}
	// Iteration trace is the last expression.
	last := result.Expressions[len(result.Expressions)-1]
	if last.Kind != liquid.KindIteration {
		t.Errorf("last expression Kind=%q, want iteration", last.Kind)
	}
}

// RO03 — in if(true), inner expression traces exist; in if(false), inner traces DO NOT exist.
func TestRenderAudit_Result_RO03_onlyExecutedBranchTraced(t *testing.T) {
	tpl := mustParseAudit(t, "{% if flag %}{{ inside_true }}{% else %}{{ inside_false }}{% endif %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"flag": true, "inside_true": "yes", "inside_false": "no"},
		liquid.AuditOptions{TraceConditions: true, TraceVariables: true},
	)
	for _, e := range result.Expressions {
		if e.Kind == liquid.KindVariable && e.Variable != nil && e.Variable.Name == "inside_false" {
			t.Error("inside_false should not be traced (unexecuted else branch)")
		}
	}
	found := false
	for _, e := range result.Expressions {
		if e.Kind == liquid.KindVariable && e.Variable != nil && e.Variable.Name == "inside_true" {
			found = true
		}
	}
	if !found {
		t.Error("inside_true should be traced (executed if branch)")
	}
}

// ============================================================================
// AuditResult — JSON Serialization (RJ01–RJ04)
// ============================================================================

// RJ01 — AuditResult serializes to JSON without error.
func TestRenderAudit_Result_RJ01_jsonMarshal(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "hi" %}{{ x }}`)
	result := auditOK(t, tpl, liquid.Bindings{},
		liquid.AuditOptions{TraceAssignments: true, TraceVariables: true},
	)
	b, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal(AuditResult): %v", err)
	}
	if len(b) == 0 {
		t.Error("marshaled JSON should not be empty")
	}
}

// RJ02 — JSON output contains snake_case keys matching the spec.
func TestRenderAudit_Result_RJ02_jsonKeys(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}x{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2, 3}},
		liquid.AuditOptions{TraceIterations: true},
	)
	b, _ := json.Marshal(result)
	s := string(b)
	expectedKeys := []string{"output", "expressions", "diagnostics", "traced_count"}
	for _, key := range expectedKeys {
		found := false
		for i := 0; i < len(s)-len(key); i++ {
			if s[i:i+len(key)] == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected key %q in JSON output: %s", key, s)
		}
	}
}

// RJ03 — omitempty works: nil optional fields are omitted from JSON.
func TestRenderAudit_Result_RJ03_omitempty(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}x{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1}},
		liquid.AuditOptions{TraceIterations: true},
	)
	b, _ := json.Marshal(result)
	s := string(b)
	// Limit and Offset should be absent when nil (omitempty).
	if contains(s, `"limit":null`) {
		t.Error(`"limit":null should be omitted (omitempty), not present as null`)
	}
	if contains(s, `"offset":null`) {
		t.Error(`"offset":null should be omitted (omitempty), not present as null`)
	}
}

// RJ04 — roundtrip: marshal → unmarshal → Output preserved.
func TestRenderAudit_Result_RJ04_roundtrip(t *testing.T) {
	tpl := mustParseAudit(t, "Hello, {{ name }}!")
	result := auditOK(t, tpl,
		liquid.Bindings{"name": "World"},
		liquid.AuditOptions{TraceVariables: true},
	)
	b, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var decoded liquid.AuditResult
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if decoded.Output != result.Output {
		t.Errorf("roundtrip Output=%q, want %q", decoded.Output, result.Output)
	}
}

// ============================================================================
// Validate — Static Analysis (VA01–VA12)
// ============================================================================

// VA01 — empty if block: info-level empty-block diagnostic.
func TestRenderAudit_Validate_VA01_emptyIf(t *testing.T) {
	tpl := mustParseAudit(t, "{% if true %}{% endif %}")
	result, err := tpl.Validate()
	if err != nil {
		t.Fatal(err)
	}
	d := firstDiag(result.Diagnostics, "empty-block")
	if d == nil {
		t.Fatal("expected empty-block diagnostic")
	}
	if d.Severity != liquid.SeverityInfo {
		t.Errorf("Severity=%q, want info for empty-block", d.Severity)
	}
}

// VA02 — empty unless block: info-level empty-block.
func TestRenderAudit_Validate_VA02_emptyUnless(t *testing.T) {
	tpl := mustParseAudit(t, "{% unless true %}{% endunless %}")
	result, err := tpl.Validate()
	if err != nil {
		t.Fatal(err)
	}
	d := firstDiag(result.Diagnostics, "empty-block")
	if d == nil {
		t.Fatal("expected empty-block for empty unless block")
	}
}

// VA03 — empty for block: info-level empty-block.
func TestRenderAudit_Validate_VA03_emptyFor(t *testing.T) {
	tpl := mustParseAudit(t, "{% for x in items %}{% endfor %}")
	result, err := tpl.Validate()
	if err != nil {
		t.Fatal(err)
	}
	d := firstDiag(result.Diagnostics, "empty-block")
	if d == nil {
		t.Fatal("expected empty-block for empty for block")
	}
}

// VA05 — normal template with content: no empty-block.
func TestRenderAudit_Validate_VA05_normalTemplate(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}hello{% endif %}")
	result, err := tpl.Validate()
	if err != nil {
		t.Fatal(err)
	}
	d := firstDiag(result.Diagnostics, "empty-block")
	if d != nil {
		t.Errorf("unexpected empty-block diagnostic for non-empty if block")
	}
}

// VA06 — undefined filter: error-level diagnostic.
func TestRenderAudit_Validate_VA06_undefinedFilter(t *testing.T) {
	eng := liquid.NewEngine()
	tpl := mustParseAuditWith(t, eng, "{{ name | no_such_filter }}")
	result, err := tpl.Validate()
	if err != nil {
		t.Fatal(err)
	}
	d := firstDiag(result.Diagnostics, "undefined-filter")
	if d == nil {
		t.Fatal("expected undefined-filter diagnostic")
	}
	if d.Severity != liquid.SeverityError {
		t.Errorf("Severity=%q, want error for undefined-filter", d.Severity)
	}
}

// VA07 — defined filter (upcase): no undefined-filter diagnostic.
func TestRenderAudit_Validate_VA07_definedFilter(t *testing.T) {
	tpl := mustParseAudit(t, "{{ name | upcase }}")
	result, err := tpl.Validate()
	if err != nil {
		t.Fatal(err)
	}
	d := firstDiag(result.Diagnostics, "undefined-filter")
	if d != nil {
		t.Error("unexpected undefined-filter diagnostic for standard upcase filter")
	}
}

// VA08 — Validate returns empty Output string (does not render).
func TestRenderAudit_Validate_VA08_noOutput(t *testing.T) {
	tpl := mustParseAudit(t, "{{ name }}")
	result, err := tpl.Validate()
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "" {
		t.Errorf("Validate Output=%q, want empty (no rendering)", result.Output)
	}
}

// VA09 — Validate returns empty Expressions (no execution).
func TestRenderAudit_Validate_VA09_noExpressions(t *testing.T) {
	tpl := mustParseAudit(t, "{{ name }}")
	result, err := tpl.Validate()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Expressions) != 0 {
		t.Errorf("Validate Expressions len=%d, want 0 (no execution)", len(result.Expressions))
	}
}

// VA10 — multiple empty blocks detected together.
func TestRenderAudit_Validate_VA10_multipleEmptyBlocks(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}{% endif %}{% for y in items %}{% endfor %}")
	result, err := tpl.Validate()
	if err != nil {
		t.Fatal(err)
	}
	emptyBlocks := allDiags(result.Diagnostics, "empty-block")
	if len(emptyBlocks) < 2 {
		t.Errorf("expected >= 2 empty-block diagnostics, got %d", len(emptyBlocks))
	}
}

// VA11 — block with only whitespace: may or may not be empty-block (implementation-defined).
// The test documents the behavior, not requires a specific outcome.
func TestRenderAudit_Validate_VA11_whitespaceOnlyBlock(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %} {% endif %}")
	result, err := tpl.Validate()
	if err != nil {
		t.Fatal(err)
	}
	// Just verify no panic and result is non-nil.
	if result == nil {
		t.Fatal("Validate result must not be nil")
	}
	t.Logf("whitespace-only if block: empty-block count=%d (implementation-defined)",
		len(allDiags(result.Diagnostics, "empty-block")))
}

// VA12 — nested empty block: inner empty block detected.
func TestRenderAudit_Validate_VA12_nestedEmptyBlock(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}{% if y %}{% endif %}{% endif %}")
	result, err := tpl.Validate()
	if err != nil {
		t.Fatal(err)
	}
	emptyBlocks := allDiags(result.Diagnostics, "empty-block")
	if len(emptyBlocks) < 1 {
		t.Error("expected at least 1 empty-block diagnostic for inner empty if")
	}
}

// ============================================================================
// RenderOptions Interaction (RO01–RO06)
// ============================================================================

// RO01 — WithStrictVariables: undefined-variable captured as warning.
func TestRenderAudit_RenderOpts_RO01_strictVariables(t *testing.T) {
	tpl := mustParseAudit(t, "{{ undefined }}")
	result, ae := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{},
		liquid.WithStrictVariables(),
	)
	if ae == nil {
		t.Fatal("expected AuditError with StrictVariables")
	}
	d := firstDiag(result.Diagnostics, "undefined-variable")
	if d == nil {
		t.Fatal("expected undefined-variable diagnostic")
	}
}

// RO02 — without StrictVariables: no diagnostic for undefined.
func TestRenderAudit_RenderOpts_RO02_noStrict(t *testing.T) {
	tpl := mustParseAudit(t, "{{ undefined }}")
	result, ae := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
	if ae != nil {
		t.Fatalf("unexpected AuditError without StrictVariables: %v", ae)
	}
	if len(result.Diagnostics) > 0 {
		t.Errorf("expected no diagnostics without StrictVariables, got %v", result.Diagnostics)
	}
}

// RO03 — WithLaxFilters: unknown filter does not produce an error.
func TestRenderAudit_RenderOpts_RO03_laxFilters(t *testing.T) {
	tpl := mustParseAudit(t, "{{ name | unknown_filter }}")
	result, ae := tpl.RenderAudit(
		liquid.Bindings{"name": "Alice"},
		liquid.AuditOptions{},
		liquid.WithLaxFilters(),
	)
	if ae != nil {
		t.Fatalf("unexpected AuditError with LaxFilters: %v", ae)
	}
	if result == nil {
		t.Fatal("result must not be nil")
	}
}

// RO04 — WithGlobals: global variables accessible.
func TestRenderAudit_RenderOpts_RO04_withGlobals(t *testing.T) {
	tpl := mustParseAudit(t, "{{ site_name }}")
	result := auditOK(t, tpl,
		liquid.Bindings{},
		liquid.AuditOptions{TraceVariables: true},
		liquid.WithGlobals(map[string]any{"site_name": "My Site"}),
	)
	assertOutput(t, result, "My Site")
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Value != "My Site" {
		t.Errorf("Value=%v, want My Site", v.Variable.Value)
	}
}

// RO05 — WithSizeLimit: output is limited but trace still collected.
func TestRenderAudit_RenderOpts_RO05_sizeLimit(t *testing.T) {
	tpl := mustParseAudit(t, "{{ a }}{{ b }}")
	result, _ := tpl.RenderAudit(
		liquid.Bindings{"a": "hello", "b": "world"},
		liquid.AuditOptions{TraceVariables: true},
		liquid.WithSizeLimit(5), // limit to 5 bytes
	)
	if result == nil {
		t.Fatal("result must not be nil")
	}
	// Output should be truncated.
	if len(result.Output) > 5 {
		t.Errorf("Output len=%d, want <= 5 (size limit)", len(result.Output))
	}
}

// ============================================================================
// Helper function used in JSON tests
// ============================================================================

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
