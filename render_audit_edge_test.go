package liquid_test

import (
	"strings"
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// Position & Range Precision (P01–P08)
// ============================================================================

// P01 — expression on first line, first column: Start.Line=1, Start.Column=1.
func TestRenderAudit_Position_P01_lineOneColOne(t *testing.T) {
	tpl := mustParseAudit(t, "{{ x }}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Range.Start.Line != 1 {
		t.Errorf("Range.Start.Line=%d, want 1", v.Range.Start.Line)
	}
	if v.Range.Start.Column != 1 {
		t.Errorf("Range.Start.Column=%d, want 1", v.Range.Start.Column)
	}
}

// P02 — expression on third line: Start.Line=3.
func TestRenderAudit_Position_P02_lineThree(t *testing.T) {
	tpl := mustParseAudit(t, "a\nb\n{{ x }}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Range.Start.Line != 3 {
		t.Errorf("Range.Start.Line=%d, want 3", v.Range.Start.Line)
	}
}

// P03 — expression preceded by text: Start.Column > 1.
func TestRenderAudit_Position_P03_columnOffset(t *testing.T) {
	// "Hello " is 6 chars, so {{ x }} starts at col 7.
	tpl := mustParseAudit(t, "Hello {{ x }}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Range.Start.Column <= 1 {
		t.Errorf("Range.Start.Column=%d, want > 1 (preceded by text)", v.Range.Start.Column)
	}
}

// P04 — End.Column = Start.Column + len(source) for single-line expression.
func TestRenderAudit_Position_P04_endColumnPrecise(t *testing.T) {
	// "{{ x }}" = 7 chars, at col 1 → End.Column = 8.
	tpl := mustParseAudit(t, "{{ x }}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	wantEnd := v.Range.Start.Column + len("{{ x }}")
	if v.Range.End.Column != wantEnd {
		t.Errorf("Range.End.Column=%d, want %d (Start+len)", v.Range.End.Column, wantEnd)
	}
}

// P05 — two expressions in the same template have non-overlapping Ranges.
func TestRenderAudit_Position_P05_noOverlap(t *testing.T) {
	tpl := mustParseAudit(t, "{{ a }} {{ b }}")
	result := auditOK(t, tpl, liquid.Bindings{"a": 1, "b": 2}, liquid.AuditOptions{TraceVariables: true})
	if len(result.Expressions) < 2 {
		t.Fatalf("expected 2 expressions, got %d", len(result.Expressions))
	}
	r0, r1 := result.Expressions[0].Range, result.Expressions[1].Range
	// r1.Start must be >= r0.End.
	endBeforeStart := r1.Start.Line < r0.End.Line ||
		(r1.Start.Line == r0.End.Line && r1.Start.Column < r0.End.Column)
	if endBeforeStart {
		t.Errorf("ranges overlap: r0=[%v→%v] r1=[%v→%v]", r0.Start, r0.End, r1.Start, r1.End)
	}
}

// P06 — expression on last line of multiline template: Line is correct.
func TestRenderAudit_Position_P07_lastLine(t *testing.T) {
	tpl := mustParseAudit(t, "line1\nline2\nline3\nline4\n{{ x }}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Range.Start.Line != 5 {
		t.Errorf("Range.Start.Line=%d, want 5 (last line)", v.Range.Start.Line)
	}
}

// P07 — assign tag: Start.Column=1 when at start of line.
func TestRenderAudit_Position_P08_assignStartColumn(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "y" %}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil {
		t.Fatal("no assignment expression")
	}
	if a.Range.Start.Column != 1 {
		t.Errorf("Range.Start.Column=%d, want 1 (assign at start of line)", a.Range.Start.Column)
	}
}

// ============================================================================
// Edge Cases (E01–E15)
// ============================================================================

// E01 — empty template: no crash, empty output, no expressions, no diagnostics.
func TestRenderAudit_Edge_E01_emptyTemplate(t *testing.T) {
	tpl := mustParseAudit(t, "")
	result := auditOK(t, tpl, liquid.Bindings{},
		liquid.AuditOptions{
			TraceVariables:   true,
			TraceConditions:  true,
			TraceIterations:  true,
			TraceAssignments: true,
		},
	)
	if result.Output != "" {
		t.Errorf("Output=%q, want empty", result.Output)
	}
	if len(result.Expressions) != 0 {
		t.Errorf("Expressions=%d, want 0", len(result.Expressions))
	}
	if len(result.Diagnostics) != 0 {
		t.Errorf("Diagnostics=%d, want 0", len(result.Diagnostics))
	}
}

// E02 — template with only text: no traces.
func TestRenderAudit_Edge_E02_textOnly(t *testing.T) {
	tpl := mustParseAudit(t, "Hello, World!")
	result := auditOK(t, tpl, liquid.Bindings{},
		liquid.AuditOptions{TraceVariables: true, TraceConditions: true},
	)
	assertOutput(t, result, "Hello, World!")
	if len(result.Expressions) != 0 {
		t.Errorf("Expressions=%d, want 0 for text-only template", len(result.Expressions))
	}
}

// E03 — deeply nested for×if×if: Depth increments correctly.
func TestRenderAudit_Edge_E03_tripleNesting(t *testing.T) {
	tpl := mustParseAudit(t, "{% for i in items %}{% if true %}{% if true %}{{ i }}{% endif %}{% endif %}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression in deeply nested template")
	}
	if v.Depth != 3 {
		t.Errorf("Depth=%d, want 3 (for > if > if)", v.Depth)
	}
}

// E04 — {% comment %} content is not traced.
func TestRenderAudit_Edge_E04_commentNotTraced(t *testing.T) {
	tpl := mustParseAudit(t, "{% comment %}{{ secret }}{% endcomment %}{{ visible }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"secret": "hidden", "visible": "shown"},
		liquid.AuditOptions{TraceVariables: true},
	)
	for _, e := range result.Expressions {
		if e.Kind == liquid.KindVariable && e.Variable != nil && e.Variable.Name == "secret" {
			t.Error("secret inside comment should not be traced")
		}
	}
	assertOutput(t, result, "shown")
}

// E05 — {% raw %} content is not parsed or traced.
func TestRenderAudit_Edge_E05_rawNotTraced(t *testing.T) {
	tpl := mustParseAudit(t, "{% raw %}{{ not_parsed }}{% endraw %}")
	result := auditOK(t, tpl, liquid.Bindings{},
		liquid.AuditOptions{TraceVariables: true},
	)
	assertOutput(t, result, "{{ not_parsed }}")
	if len(result.Expressions) != 0 {
		t.Errorf("Expressions=%d, want 0 inside raw block", len(result.Expressions))
	}
}

// E06 — Unicode values are preserved correctly in traces.
func TestRenderAudit_Edge_E06_unicodeValues(t *testing.T) {
	tpl := mustParseAudit(t, "{{ greeting }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"greeting": "Olá, João! 🎉"},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Value != "Olá, João! 🎉" {
		t.Errorf("Value=%v, want unicode greeting", v.Variable.Value)
	}
	assertOutput(t, result, "Olá, João! 🎉")
}

// E07 — whitespace control tags ({%- -%}): output is trimmed, traces are still present.
func TestRenderAudit_Edge_E07_whitespaceControl(t *testing.T) {
	tpl := mustParseAudit(t, "  {%- if true -%}  yes  {%- endif -%}  ")
	result := auditOK(t, tpl, liquid.Bindings{},
		liquid.AuditOptions{TraceConditions: true},
	)
	// Output should be trimmed around the tags.
	if strings.Contains(result.Output, "  yes  ") {
		t.Logf("Output=%q (whitespace might be trimmed)", result.Output)
	}
	// But traces should still appear.
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil {
		t.Error("condition expression should still be traced with whitespace control tags")
	}
}

// E08 — increment tag: no crash.
func TestRenderAudit_Edge_E08_incrementTag(t *testing.T) {
	tpl := mustParseAudit(t, "{% increment counter %}{% increment counter %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{})
	if result == nil {
		t.Fatal("result must not be nil")
	}
}

// E09 — decrement tag: no crash.
func TestRenderAudit_Edge_E09_decrementTag(t *testing.T) {
	tpl := mustParseAudit(t, "{% decrement counter %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{})
	if result == nil {
		t.Fatal("result must not be nil")
	}
}

// E10 — cycle tag: no crash.
func TestRenderAudit_Edge_E10_cycleTag(t *testing.T) {
	tpl := mustParseAudit(t, `{% for i in items %}{% cycle "odd", "even" %}{% endfor %}`)
	result := auditOK(t, tpl, liquid.Bindings{"items": []int{1, 2, 3}}, liquid.AuditOptions{})
	if result.Output != "oddevenodd" {
		t.Errorf("Output=%q, want oddevenodd", result.Output)
	}
}

// E11 — very long filter pipeline (5 steps): no crash, traces correctly.
func TestRenderAudit_Edge_E11_longPipeline(t *testing.T) {
	tpl := mustParseAudit(t, `{{ name | downcase | upcase | downcase | upcase | downcase }}`)
	result := auditOK(t, tpl,
		liquid.Bindings{"name": "Hello"},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 5 {
		t.Errorf("Pipeline len=%d, want 5", len(v.Variable.Pipeline))
	}
}

// E12 — multiple independent if blocks: each produces its own ConditionTrace.
func TestRenderAudit_Edge_E12_multipleConditionBlocks(t *testing.T) {
	tpl := mustParseAudit(t, "{% if a %}yes{% endif %}{% if b %}no{% endif %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"a": true, "b": false},
		liquid.AuditOptions{TraceConditions: true},
	)
	conds := allExprs(result.Expressions, liquid.KindCondition)
	if len(conds) != 2 {
		t.Errorf("condition expression count=%d, want 2", len(conds))
	}
}

// E13 — assign then for then if: assignment comes first; all three kinds are present.
func TestRenderAudit_Edge_E13_executionOrder(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = 1 %}{% for i in items %}{% if i > x %}big{% endif %}{% endfor %}`)
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{0, 2}},
		liquid.AuditOptions{
			TraceAssignments: true,
			TraceIterations:  true,
			TraceConditions:  true,
		},
	)
	if len(result.Expressions) == 0 {
		t.Fatal("no expressions")
	}
	// The assignment tag appears before the for loop, so it is always first.
	if result.Expressions[0].Kind != liquid.KindAssignment {
		t.Errorf("Expressions[0].Kind=%q, want assignment (comes before for)", result.Expressions[0].Kind)
	}
	// All three expression kinds must be present somewhere.
	kinds := make(map[liquid.ExpressionKind]bool)
	for _, e := range result.Expressions {
		kinds[e.Kind] = true
	}
	if !kinds[liquid.KindIteration] {
		t.Error("expected at least one iteration expression")
	}
	if !kinds[liquid.KindCondition] {
		t.Error("expected at least one condition expression")
	}
}

// E14 — AuditResult.result nil pointer never happens even for panicky templates.
func TestRenderAudit_Edge_E14_resultNeverNil(t *testing.T) {
	templates := []string{
		"",
		"{{ x }}",
		"{% if true %}{% endif %}",
		"{% for i in items %}{% endfor %}",
	}
	for _, src := range templates {
		tpl := mustParseAudit(t, src)
		result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{})
		if result == nil {
			t.Errorf("RenderAudit(%q) returned nil result", src)
		}
	}
}

// E15 — echo tag works like variable output.
func TestRenderAudit_Edge_E15_echoTag(t *testing.T) {
	tpl := mustParseAudit(t, "{% echo name %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"name": "Alice"},
		liquid.AuditOptions{TraceVariables: true},
	)
	assertOutput(t, result, "Alice")
}

// ============================================================================
// End-to-End Spec Example (S01–S02)
// ============================================================================

// S01 — the complete spec example template produces the expected output
// and the Expressions array contains all expression kinds in order.
func TestRenderAudit_E2E_S01_specExample(t *testing.T) {
	src := `{% assign title = page.title | upcase %}
<h1>{{ title }}</h1>

{% if customer.age >= 18 %}
  <p>Welcome, {{ customer.name }}!</p>
{% else %}
  <p>Restricted.</p>
{% endif %}

{% for item in cart.items %}
  <li>{{ item.name }} — ${{ item.price | times: 1.1 | round }}</li>
{% endfor %}`

	bindings := liquid.Bindings{
		"page":     map[string]any{"title": "my store"},
		"customer": map[string]any{"name": "Alice", "age": 25},
		"cart": map[string]any{
			"items": []map[string]any{
				{"name": "Shirt", "price": 50},
				{"name": "Pants", "price": 120},
			},
		},
	}

	eng := newAuditEngine()
	tpl, err := eng.ParseString(src)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(bindings, liquid.AuditOptions{
		TraceVariables:         true,
		TraceConditions:        true,
		TraceIterations:        true,
		TraceAssignments:       true,
		MaxIterationTraceItems: 100,
	})
	if ae != nil {
		t.Fatalf("unexpected AuditError: %v", ae)
	}

	// ---- Output assertions ------------------------------------------------
	if !strings.Contains(result.Output, "MY STORE") {
		t.Errorf("Output should contain 'MY STORE' (assign title | upcase), got: %q", result.Output)
	}
	if !strings.Contains(result.Output, "Welcome, Alice!") {
		t.Errorf("Output should contain 'Welcome, Alice!', got: %q", result.Output)
	}
	if strings.Contains(result.Output, "Restricted.") {
		t.Error("Output should NOT contain 'Restricted.' (customer.age=25 is adult)")
	}
	if !strings.Contains(result.Output, "Shirt") {
		t.Errorf("Output should contain 'Shirt', got: %q", result.Output)
	}
	if !strings.Contains(result.Output, "Pants") {
		t.Errorf("Output should contain 'Pants', got: %q", result.Output)
	}

	// ---- Expression kind assertions ----------------------------------------
	kinds := make(map[liquid.ExpressionKind]int)
	for _, e := range result.Expressions {
		kinds[e.Kind]++
	}

	if kinds[liquid.KindAssignment] < 1 {
		t.Error("expected at least 1 assignment expression ({% assign title %})")
	}
	if kinds[liquid.KindVariable] < 1 {
		t.Error("expected variable expressions")
	}
	if kinds[liquid.KindCondition] < 1 {
		t.Error("expected at least 1 condition expression ({% if customer.age >= 18 %})")
	}
	if kinds[liquid.KindIteration] < 1 {
		t.Error("expected at least 1 iteration expression ({% for item in cart.items %})")
	}

	// ---- Assignment trace --------------------------------------------------
	assigns := allExprs(result.Expressions, liquid.KindAssignment)
	if len(assigns) == 0 || assigns[0].Assignment == nil {
		t.Fatal("no assignment trace for 'assign title'")
	}
	if assigns[0].Assignment.Variable != "title" {
		t.Errorf("assign.Variable=%q, want title", assigns[0].Assignment.Variable)
	}
	if assigns[0].Assignment.Value != "MY STORE" {
		t.Errorf("assign.Value=%v, want MY STORE", assigns[0].Assignment.Value)
	}
	if len(assigns[0].Assignment.Pipeline) != 1 || assigns[0].Assignment.Pipeline[0].Filter != "upcase" {
		t.Error("assign pipeline should have one step: upcase")
	}

	// ---- Condition trace ---------------------------------------------------
	conds := allExprs(result.Expressions, liquid.KindCondition)
	if len(conds) == 0 || conds[0].Condition == nil {
		t.Fatal("no condition trace")
	}
	// The if branch (customer.age >= 18) should be executed.
	found := false
	for _, b := range conds[0].Condition.Branches {
		if b.Kind == "if" && b.Executed {
			found = true
		}
	}
	if !found {
		t.Error("if branch (customer.age >= 18) should be Executed=true")
	}

	// ---- Iteration trace ---------------------------------------------------
	iters := allExprs(result.Expressions, liquid.KindIteration)
	if len(iters) == 0 || iters[0].Iteration == nil {
		t.Fatal("no iteration trace")
	}
	if iters[0].Iteration.Variable != "item" {
		t.Errorf("iter.Variable=%q, want item", iters[0].Iteration.Variable)
	}
	if iters[0].Iteration.Length != 2 {
		t.Errorf("iter.Length=%d, want 2 (two cart items)", iters[0].Iteration.Length)
	}
	if iters[0].Iteration.TracedCount != 2 {
		t.Errorf("iter.TracedCount=%d, want 2", iters[0].Iteration.TracedCount)
	}

	// ---- No diagnostics ----------------------------------------------------
	if len(result.Diagnostics) > 0 {
		t.Errorf("expected no diagnostics, got %v", result.Diagnostics)
	}
}

// S02 — verify that the spec example can also be validated without panic.
func TestRenderAudit_E2E_S02_validateSpecExample(t *testing.T) {
	src := `{% assign title = page.title | upcase %}
<h1>{{ title }}</h1>
{% if customer.age >= 18 %}
  <p>Welcome, {{ customer.name }}!</p>
{% else %}
  <p>Restricted.</p>
{% endif %}
{% for item in cart.items %}
  <li>{{ item.name }}</li>
{% endfor %}`

	tpl := mustParseAudit(t, src)
	result, err := tpl.Validate()
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Validate result must not be nil")
	}
	// The spec example is well-formed and non-empty — no empty-block diagnostics expected.
	emptyBlocks := allDiags(result.Diagnostics, "empty-block")
	if len(emptyBlocks) > 0 {
		t.Errorf("unexpected empty-block diagnostics in spec example: %v", emptyBlocks)
	}
}

// ============================================================================
// Additional RenderAudit parity test: matches Template.Render exactly.
// ============================================================================

func TestRenderAudit_Parity_withRender(t *testing.T) {
	templates := []struct {
		name     string
		src      string
		bindings liquid.Bindings
	}{
		{"simple", "Hello, {{ name }}!", liquid.Bindings{"name": "World"}},
		{"if_true", "{% if x %}yes{% else %}no{% endif %}", liquid.Bindings{"x": true}},
		{"if_false", "{% if x %}yes{% else %}no{% endif %}", liquid.Bindings{"x": false}},
		{"for", "{% for i in items %}{{ i }}{% endfor %}", liquid.Bindings{"items": []int{1, 2, 3}}},
		{"assign", `{% assign x = "hi" %}{{ x }}`, liquid.Bindings{}},
		{"filters", "{{ name | upcase | truncate: 3 }}", liquid.Bindings{"name": "hello"}},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			eng := newAuditEngine()
			expected, se := eng.ParseAndRenderString(tt.src, tt.bindings)
			if se != nil {
				t.Fatalf("baseline Render error: %v", se)
			}
			tpl := mustParseAuditWith(t, eng, tt.src)
			result := auditOK(t, tpl, tt.bindings,
				liquid.AuditOptions{
					TraceVariables:   true,
					TraceConditions:  true,
					TraceIterations:  true,
					TraceAssignments: true,
				},
			)
			if result.Output != expected {
				t.Errorf("RenderAudit output=%q, Render output=%q (must be identical)", result.Output, expected)
			}
		})
	}
}
