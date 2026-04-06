package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// IterationTrace — Basic Attributes (I01–I07)
// ============================================================================

// I01 — basic for loop: Variable and Collection names.
func TestRenderAudit_Iteration_I01_basic(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"a", "b"}},
		liquid.AuditOptions{TraceIterations: true},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Variable != "item" {
		t.Errorf("Variable=%q, want item", it.Iteration.Variable)
	}
	if it.Iteration.Collection != "items" {
		t.Errorf("Collection=%q, want items", it.Iteration.Collection)
	}
}

// I02 — iteration over empty collection: Length=0, TracedCount=0.
func TestRenderAudit_Iteration_I02_emptyCollection(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl, liquid.Bindings{"items": []string{}}, liquid.AuditOptions{TraceIterations: true})
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Length != 0 {
		t.Errorf("Length=%d, want 0", it.Iteration.Length)
	}
	if it.Iteration.TracedCount != 0 {
		t.Errorf("TracedCount=%d, want 0", it.Iteration.TracedCount)
	}
	if it.Iteration.Truncated {
		t.Error("Truncated should be false for empty collection")
	}
}

// I03 — single-item collection: Length=1, TracedCount=1.
func TestRenderAudit_Iteration_I03_singleItem(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl, liquid.Bindings{"items": []string{"only"}}, liquid.AuditOptions{TraceIterations: true})
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Length != 1 {
		t.Errorf("Length=%d, want 1", it.Iteration.Length)
	}
	if it.Iteration.TracedCount != 1 {
		t.Errorf("TracedCount=%d, want 1", it.Iteration.TracedCount)
	}
}

// I04 — 100 items: Length=100.
func TestRenderAudit_Iteration_I04_manyItems(t *testing.T) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}
	tpl := mustParseAudit(t, "{% for item in items %}x{% endfor %}")
	result := auditOK(t, tpl, liquid.Bindings{"items": items}, liquid.AuditOptions{TraceIterations: true})
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Length != 100 {
		t.Errorf("Length=%d, want 100", it.Iteration.Length)
	}
}

// I05 — iteration over a map/hash: Length is the number of key-value pairs.
func TestRenderAudit_Iteration_I05_overMap(t *testing.T) {
	tpl := mustParseAudit(t, "{% for pair in hash %}{{ pair[0] }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"hash": map[string]any{"a": 1, "b": 2}},
		liquid.AuditOptions{TraceIterations: true},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Length != 2 {
		t.Errorf("Length=%d, want 2 (hash with 2 entries)", it.Iteration.Length)
	}
}

// I06 — range literal (1..5): Length=5.
func TestRenderAudit_Iteration_I06_rangeLiteral(t *testing.T) {
	tpl := mustParseAudit(t, "{% for i in (1..5) %}{{ i }}{% endfor %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceIterations: true})
	assertOutput(t, result, "12345")
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Variable != "i" {
		t.Errorf("Variable=%q, want i", it.Iteration.Variable)
	}
	if it.Iteration.Length != 5 {
		t.Errorf("Length=%d, want 5", it.Iteration.Length)
	}
}

// I07 — reversed range (5..1): this implementation yields a non-positive Length
// (computed as end-start+1 = 1-5+1 = -3) meaning no elements are iterated.
func TestRenderAudit_Iteration_I07_emptyRange(t *testing.T) {
	tpl := mustParseAudit(t, "{% for i in (5..1) %}{{ i }}{% endfor %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceIterations: true})
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	// For a reversed/empty range, Length is non-positive (no iterations executed).
	if it.Iteration.Length > 0 {
		t.Errorf("Length=%d, want <= 0 for reversed range (5..1)", it.Iteration.Length)
	}
	if it.Iteration.TracedCount != 0 {
		t.Errorf("TracedCount=%d, want 0 (no iterations for reversed range)", it.Iteration.TracedCount)
	}
}

// ============================================================================
// IterationTrace — Limit, Offset, Reversed (IL01–IL07)
// ============================================================================

// IL01 — limit:3 with 5 items: Limit=ptr(3), Length=3 (actual iterations run), TracedCount=3.
// Note: Length reflects the number of elements actually iterated (post-limit), not the
// original collection size.
func TestRenderAudit_Iteration_IL01_limit(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items limit:3 %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2, 3, 4, 5}},
		liquid.AuditOptions{TraceIterations: true},
	)
	assertOutput(t, result, "123")
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Limit == nil {
		t.Fatal("Limit should be non-nil when limit: is specified")
	}
	if *it.Iteration.Limit != 3 {
		t.Errorf("*Limit=%d, want 3", *it.Iteration.Limit)
	}
	if it.Iteration.Length != 3 {
		t.Errorf("Length=%d, want 3 (post-limit iteration count)", it.Iteration.Length)
	}
	if it.Iteration.TracedCount != 3 {
		t.Errorf("TracedCount=%d, want 3", it.Iteration.TracedCount)
	}
}

// IL02 — offset:2 with 5 items: Offset=ptr(2), Length=3 (items remaining after skip), TracedCount=3.
// Note: Length reflects elements actually iterated (collection size minus offset), not total.
func TestRenderAudit_Iteration_IL02_offset(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items offset:2 %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{10, 20, 30, 40, 50}},
		liquid.AuditOptions{TraceIterations: true},
	)
	assertOutput(t, result, "304050")
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Offset == nil {
		t.Fatal("Offset should be non-nil when offset: is specified")
	}
	if *it.Iteration.Offset != 2 {
		t.Errorf("*Offset=%d, want 2", *it.Iteration.Offset)
	}
	if it.Iteration.Length != 3 {
		t.Errorf("Length=%d, want 3 (5 items minus 2 offset)", it.Iteration.Length)
	}
}

// IL03 — limit:2 offset:1 combined.
func TestRenderAudit_Iteration_IL03_limitAndOffset(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items limit:2 offset:1 %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{10, 20, 30, 40}},
		liquid.AuditOptions{TraceIterations: true},
	)
	assertOutput(t, result, "2030")
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Limit == nil {
		t.Error("Limit should be non-nil")
	} else if *it.Iteration.Limit != 2 {
		t.Errorf("*Limit=%d, want 2", *it.Iteration.Limit)
	}
	if it.Iteration.Offset == nil {
		t.Error("Offset should be non-nil")
	} else if *it.Iteration.Offset != 1 {
		t.Errorf("*Offset=%d, want 1", *it.Iteration.Offset)
	}
}

// IL04 — reversed: Reversed=true.
func TestRenderAudit_Iteration_IL04_reversed(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items reversed %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2, 3}},
		liquid.AuditOptions{TraceIterations: true},
	)
	assertOutput(t, result, "321")
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if !it.Iteration.Reversed {
		t.Error("Reversed should be true when `reversed` modifier is used")
	}
}

// IL05 — no modifiers: Limit=nil, Offset=nil, Reversed=false.
func TestRenderAudit_Iteration_IL05_noModifiers(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2}},
		liquid.AuditOptions{TraceIterations: true},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Limit != nil {
		t.Errorf("Limit should be nil (not specified), got %d", *it.Iteration.Limit)
	}
	if it.Iteration.Offset != nil {
		t.Errorf("Offset should be nil (not specified), got %d", *it.Iteration.Offset)
	}
	if it.Iteration.Reversed {
		t.Error("Reversed should be false when not specified")
	}
}

// IL06 — limit:0: iterates zero times despite non-empty collection.
func TestRenderAudit_Iteration_IL06_limitZero(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items limit:0 %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2, 3}},
		liquid.AuditOptions{TraceIterations: true},
	)
	assertOutput(t, result, "")
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Limit == nil {
		t.Error("Limit should be non-nil for limit:0")
	} else if *it.Iteration.Limit != 0 {
		t.Errorf("*Limit=%d, want 0", *it.Iteration.Limit)
	}
	if it.Iteration.TracedCount != 0 {
		t.Errorf("TracedCount=%d, want 0 (no iterations)", it.Iteration.TracedCount)
	}
}

// ============================================================================
// IterationTrace — MaxIterationTraceItems / Truncation (IT01–IT07)
// ============================================================================

// IT01 — MaxIterItems=0 (unlimited) with 10 items: Truncated=false, TracedCount=10.
func TestRenderAudit_Iteration_IT01_noLimit(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% endfor %}")
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := auditOK(t, tpl,
		liquid.Bindings{"items": items},
		liquid.AuditOptions{TraceIterations: true, MaxIterationTraceItems: 0},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Truncated {
		t.Error("Truncated should be false when MaxIterationTraceItems=0 (unlimited)")
	}
	if it.Iteration.TracedCount != 10 {
		t.Errorf("TracedCount=%d, want 10", it.Iteration.TracedCount)
	}
}

// IT02 — MaxIterItems=5 with 10 items: Truncated=true, TracedCount=5.
func TestRenderAudit_Iteration_IT02_truncation(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}x{% endfor %}")
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := auditOK(t, tpl,
		liquid.Bindings{"items": items},
		liquid.AuditOptions{TraceIterations: true, MaxIterationTraceItems: 5},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if !it.Iteration.Truncated {
		t.Error("Truncated should be true when MaxIterationTraceItems=5 and 10 items")
	}
	if it.Iteration.TracedCount != 5 {
		t.Errorf("TracedCount=%d, want 5", it.Iteration.TracedCount)
	}
}

// IT03 — MaxIterItems=10 with only 5 items: Truncated=false, TracedCount=5.
func TestRenderAudit_Iteration_IT03_limitExceedsItems(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}x{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2, 3, 4, 5}},
		liquid.AuditOptions{TraceIterations: true, MaxIterationTraceItems: 10},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Truncated {
		t.Error("Truncated should be false (only 5 items, limit 10)")
	}
	if it.Iteration.TracedCount != 5 {
		t.Errorf("TracedCount=%d, want 5", it.Iteration.TracedCount)
	}
}

// IT04 — MaxIterItems=1 with 100 items: Truncated=true, TracedCount=1.
func TestRenderAudit_Iteration_IT04_limitOne(t *testing.T) {
	items := make([]int, 100)
	tpl := mustParseAudit(t, "{% for item in items %}x{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": items},
		liquid.AuditOptions{TraceIterations: true, MaxIterationTraceItems: 1},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if !it.Iteration.Truncated {
		t.Error("Truncated should be true")
	}
	if it.Iteration.TracedCount != 1 {
		t.Errorf("TracedCount=%d, want 1", it.Iteration.TracedCount)
	}
}

// IT05 — MaxIterItems limits inner expression tracing but NOT the render output.
func TestRenderAudit_Iteration_IT05_outputCompleteEvenWhenTruncated(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2, 3, 4, 5}},
		liquid.AuditOptions{TraceIterations: true, TraceVariables: true, MaxIterationTraceItems: 2},
	)
	// Output must be complete despite truncation.
	assertOutput(t, result, "12345")

	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if !it.Iteration.Truncated {
		t.Error("Truncated should be true (5 items, limit 2)")
	}

	// Variable expressions are only traced for the first 2 iterations.
	varExprs := allExprs(result.Expressions, liquid.KindVariable)
	if len(varExprs) != 2 {
		t.Errorf("variable expression count=%d, want 2 (only traced iterations)", len(varExprs))
	}
}

// IT06 — nested for loops each have their own TracedCount.
// Note: inner for IterationTraces are emitted BEFORE the outer for's IterationTrace
// (because the outer body executes before the outer event finishes).
func TestRenderAudit_Iteration_IT06_nestedForSeparateTracedCount(t *testing.T) {
	tpl := mustParseAudit(t, "{% for outer in outers %}{% for inner in inners %}x{% endfor %}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{
			"outers": []int{1, 2},
			"inners": []int{1, 2, 3},
		},
		liquid.AuditOptions{TraceIterations: true},
	)
	iterExprs := allExprs(result.Expressions, liquid.KindIteration)
	// Expect 3 iteration expressions: inner for × 2 outer iterations + outer for × 1.
	if len(iterExprs) < 2 {
		t.Fatalf("expected >= 2 iteration expressions, got %d", len(iterExprs))
	}
	// Find the outer for by variable name. The outer for's trace is the LAST one emitted.
	var outerIter *liquid.Expression
	for i := range iterExprs {
		if iterExprs[i].Iteration != nil && iterExprs[i].Iteration.Variable == "outer" {
			e := iterExprs[i]
			outerIter = &e
			break
		}
	}
	if outerIter == nil {
		t.Fatal("no iteration trace with variable=\"outer\"")
	}
	if outerIter.Iteration.Length != 2 {
		t.Errorf("outer.Length=%d, want 2", outerIter.Iteration.Length)
	}
	// Verify at least one inner-for trace with variable="inner".
	var innerIter *liquid.Expression
	for i := range iterExprs {
		if iterExprs[i].Iteration != nil && iterExprs[i].Iteration.Variable == "inner" {
			e := iterExprs[i]
			innerIter = &e
			break
		}
	}
	if innerIter == nil {
		t.Fatal("no iteration trace with variable=\"inner\"")
	}
	if innerIter.Iteration.Length != 3 {
		t.Errorf("inner.Length=%d, want 3", innerIter.Iteration.Length)
	}
}

// IT07 — MaxIterItems with empty collection: Truncated=false, TracedCount=0.
func TestRenderAudit_Iteration_IT07_maxIterWithEmptyCollection(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{}},
		liquid.AuditOptions{TraceIterations: true, MaxIterationTraceItems: 3},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Truncated {
		t.Error("Truncated should be false for empty collection")
	}
	if it.Iteration.TracedCount != 0 {
		t.Errorf("TracedCount=%d, want 0", it.Iteration.TracedCount)
	}
}

// ============================================================================
// IterationTrace — Inner Expressions appear per iteration (IF01–IF06)
// ============================================================================

// IF01 — variable inside for appears once per iteration.
func TestRenderAudit_Iteration_IF01_variablePerIteration(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"a", "b", "c"}},
		liquid.AuditOptions{TraceIterations: true, TraceVariables: true},
	)
	varExprs := allExprs(result.Expressions, liquid.KindVariable)
	if len(varExprs) != 3 {
		t.Errorf("variable expressions=%d, want 3 (one per iteration)", len(varExprs))
	}
	for i, v := range varExprs {
		if v.Variable == nil {
			continue
		}
		expected := []string{"a", "b", "c"}[i]
		if v.Variable.Value != expected {
			t.Errorf("varExprs[%d].Value=%v, want %q", i, v.Variable.Value, expected)
		}
	}
}

// IF02 — condition inside for appears once per iteration.
func TestRenderAudit_Iteration_IF02_conditionPerIteration(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{% if item > 2 %}big{% endif %}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2, 3}},
		liquid.AuditOptions{TraceIterations: true, TraceConditions: true},
	)
	condExprs := allExprs(result.Expressions, liquid.KindCondition)
	if len(condExprs) != 3 {
		t.Errorf("condition expressions=%d, want 3 (one per iteration)", len(condExprs))
	}
}

// IF03 — assign inside for appears once per iteration.
func TestRenderAudit_Iteration_IF03_assignPerIteration(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{% assign doubled = item | times: 2 %}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2, 3}},
		liquid.AuditOptions{TraceIterations: true, TraceAssignments: true},
	)
	assignExprs := allExprs(result.Expressions, liquid.KindAssignment)
	if len(assignExprs) != 3 {
		t.Errorf("assignment expressions=%d, want 3 (one per iteration)", len(assignExprs))
	}
}

// IF04 — nested for: inner expressions have Depth=2.
func TestRenderAudit_Iteration_IF04_nestedForDepth(t *testing.T) {
	tpl := mustParseAudit(t, "{% for outer in outers %}{% for inner in inners %}{{ inner }}{% endfor %}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"outers": []int{1}, "inners": []int{1}},
		liquid.AuditOptions{TraceIterations: true, TraceVariables: true},
	)
	varExprs := allExprs(result.Expressions, liquid.KindVariable)
	for _, v := range varExprs {
		if v.Variable != nil && v.Variable.Name == "inner" {
			if v.Depth != 2 {
				t.Errorf("inner variable Depth=%d, want 2 (nested for×for)", v.Depth)
			}
		}
	}
}

// IF05 — MaxIterItems truncates inner expressions.
func TestRenderAudit_Iteration_IF05_innerExpressionsAreTruncated(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2, 3, 4, 5}},
		liquid.AuditOptions{TraceIterations: true, TraceVariables: true, MaxIterationTraceItems: 2},
	)
	varExprs := allExprs(result.Expressions, liquid.KindVariable)
	if len(varExprs) != 2 {
		t.Errorf("variable expressions=%d, want 2 (only first 2 traced)", len(varExprs))
	}
}

// IF06 — forloop special variables (forloop.index) can be traced as variables.
func TestRenderAudit_Iteration_IF06_forloopVariables(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ forloop.index }}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"a", "b"}},
		liquid.AuditOptions{TraceIterations: true, TraceVariables: true},
	)
	// forloop.index should be accessible and traced.
	varExprs := allExprs(result.Expressions, liquid.KindVariable)
	if len(varExprs) == 0 {
		t.Fatal("expected variable expressions for forloop.index")
	}
	// On first iteration, forloop.index should be 1.
	first := varExprs[0]
	if first.Variable != nil && sprintVal(first.Variable.Value) != "1" {
		t.Errorf("forloop.index[0]=%v, want 1", first.Variable.Value)
	}
}

// ============================================================================
// IterationTrace — Tablerow (TR01–TR03)
// ============================================================================

// TR01 — tablerow produces an IterationTrace.
func TestRenderAudit_Iteration_TR01_tablerow(t *testing.T) {
	tpl := mustParseAudit(t, "{% tablerow item in items %}{{ item }}{% endtablerow %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"a", "b", "c"}},
		liquid.AuditOptions{TraceIterations: true},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("tablerow should produce an IterationTrace")
	}
	if it.Iteration.Variable != "item" {
		t.Errorf("Variable=%q, want item", it.Iteration.Variable)
	}
	if it.Iteration.Length != 3 {
		t.Errorf("Length=%d, want 3", it.Iteration.Length)
	}
}

// TR02 — tablerow with cols: Length is correct and output contains table structure.
func TestRenderAudit_Iteration_TR02_tablerowCols(t *testing.T) {
	tpl := mustParseAudit(t, "{% tablerow item in items cols:2 %}{{ item }}{% endtablerow %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"a", "b", "c", "d"}},
		liquid.AuditOptions{TraceIterations: true},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Length != 4 {
		t.Errorf("Length=%d, want 4", it.Iteration.Length)
	}
}

// TR03 — tablerow with limit: Limit field populated.
func TestRenderAudit_Iteration_TR03_tablerowLimit(t *testing.T) {
	tpl := mustParseAudit(t, "{% tablerow item in items limit:2 %}{{ item }}{% endtablerow %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"a", "b", "c", "d"}},
		liquid.AuditOptions{TraceIterations: true},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil || it.Iteration == nil {
		t.Fatal("no iteration expression")
	}
	if it.Iteration.Limit == nil {
		t.Error("Limit should be non-nil for tablerow with limit:")
	} else if *it.Iteration.Limit != 2 {
		t.Errorf("*Limit=%d, want 2", *it.Iteration.Limit)
	}
}

// ============================================================================
// IterationTrace — Source, Range, Depth (IR01–IR03)
// ============================================================================

// IR01 — Source contains the {% for ... %} header.
func TestRenderAudit_Iteration_IR01_sourceNonEmpty(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}x{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1}},
		liquid.AuditOptions{TraceIterations: true},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil {
		t.Fatal("no iteration expression")
	}
	if it.Source == "" {
		t.Error("iteration Source should be non-empty")
	}
}

// IR02 — Range.Start.Line is valid (>= 1).
func TestRenderAudit_Iteration_IR02_rangeValid(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}x{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1}},
		liquid.AuditOptions{TraceIterations: true},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil {
		t.Fatal("no iteration expression")
	}
	assertRangeValid(t, it.Range, "iteration Range")
}

// IR03 — top-level for has Depth=0; nested for inside if has Depth=1.
func TestRenderAudit_Iteration_IR03_depth(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}x{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1}},
		liquid.AuditOptions{TraceIterations: true},
	)
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it == nil {
		t.Fatal("no iteration expression")
	}
	if it.Depth != 0 {
		t.Errorf("Depth=%d, want 0 for top-level for", it.Depth)
	}
}

// ============================================================================
// IterationTrace — Error/Edge Cases (IE01–IE05)
// ============================================================================

// IE01 — for over an int → not-iterable warning, zero iterations.
func TestRenderAudit_Iteration_IE01_notIterableInt(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in orders %}{{ item }}{% endfor %}")
	result, _ := tpl.RenderAudit(
		liquid.Bindings{"orders": 42},
		liquid.AuditOptions{TraceIterations: true},
	)
	if result == nil {
		t.Fatal("result must not be nil")
	}
	d := firstDiag(result.Diagnostics, "not-iterable")
	if d == nil {
		t.Fatal("expected not-iterable diagnostic for for over int")
	}
	if d.Severity != liquid.SeverityWarning {
		t.Errorf("severity=%q, want warning", d.Severity)
	}
	// Output should be empty (zero iterations).
	assertOutput(t, result, "")
}

// IE02 — for over bool → not-iterable warning.
func TestRenderAudit_Iteration_IE02_notIterableBool(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in flag %}{{ item }}{% endfor %}")
	result, _ := tpl.RenderAudit(
		liquid.Bindings{"flag": true},
		liquid.AuditOptions{TraceIterations: true},
	)
	if result == nil {
		t.Fatal("result must not be nil")
	}
	d := firstDiag(result.Diagnostics, "not-iterable")
	if d == nil {
		t.Fatal("expected not-iterable diagnostic for for over bool")
	}
}

// IE03 — for over a string → not-iterable warning (string is not iterable in Liquid).
func TestRenderAudit_Iteration_IE03_notIterableString(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in status %}{{ item }}{% endfor %}")
	result, _ := tpl.RenderAudit(
		liquid.Bindings{"status": "pending"},
		liquid.AuditOptions{TraceIterations: true},
	)
	if result == nil {
		t.Fatal("result must not be nil")
	}
	d := firstDiag(result.Diagnostics, "not-iterable")
	if d == nil {
		t.Fatal("expected not-iterable diagnostic for for over string")
	}
}

// IE04 — for-else: when collection is empty the else block runs, and no IterationTrace is emitted.
// Note: the current implementation does NOT emit an IterationTrace for an empty collection
// (no iterations to trace).
func TestRenderAudit_Iteration_IE04_forElse_emptyCollection(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% else %}empty{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{}},
		liquid.AuditOptions{TraceIterations: true},
	)
	assertOutput(t, result, "empty")
	// No iteration trace is emitted when there are zero iterations.
	it := firstExpr(result.Expressions, liquid.KindIteration)
	if it != nil && it.Iteration != nil && it.Iteration.Length != 0 {
		t.Errorf("unexpected non-zero iteration length: %d", it.Iteration.Length)
	}
}

// IE05 — for-else: when collection is non-empty, the else block does not run.
func TestRenderAudit_Iteration_IE05_forElse_nonEmptyCollection(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% else %}empty{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"x"}},
		liquid.AuditOptions{TraceIterations: true},
	)
	assertOutput(t, result, "x")
}
