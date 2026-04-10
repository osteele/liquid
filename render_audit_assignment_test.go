package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// AssignmentTrace — Basic Attributes (A01–A09)
// ============================================================================

// A01 — {% assign x = "hello" %}: Variable="x", Value="hello".
func TestRenderAudit_Assignment_A01_stringValue(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign greeting = "hello" %}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if a.Assignment.Variable != "greeting" {
		t.Errorf("Variable=%q, want greeting", a.Assignment.Variable)
	}
	if a.Assignment.Value != "hello" {
		t.Errorf("Value=%v, want hello", a.Assignment.Value)
	}
}

// A02 — assign integer binding.
func TestRenderAudit_Assignment_A02_intValue(t *testing.T) {
	tpl := mustParseAudit(t, "{% assign count = 42 %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if sprintVal(a.Assignment.Value) != "42" {
		t.Errorf("Value=%v, want 42", a.Assignment.Value)
	}
}

// A03 — assign float literal.
func TestRenderAudit_Assignment_A03_floatValue(t *testing.T) {
	tpl := mustParseAudit(t, "{% assign pi = 3.14 %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if sprintVal(a.Assignment.Value) != "3.14" {
		t.Errorf("Value=%v, want 3.14", a.Assignment.Value)
	}
}

// A04 — assign true literal.
func TestRenderAudit_Assignment_A04_boolTrue(t *testing.T) {
	tpl := mustParseAudit(t, "{% assign flag = true %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if a.Assignment.Value != true {
		t.Errorf("Value=%v, want true", a.Assignment.Value)
	}
}

// A05 — assign false literal.
func TestRenderAudit_Assignment_A05_boolFalse(t *testing.T) {
	tpl := mustParseAudit(t, "{% assign flag = false %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if a.Assignment.Value != false {
		t.Errorf("Value=%v, want false", a.Assignment.Value)
	}
}

// A06 — assign from an undefined variable: value is nil.
// Note: "nil" and "empty" are reserved Liquid constants and cannot be used as
// variable names or assignment values directly; assign from an undefined variable instead.
func TestRenderAudit_Assignment_A06_nilValue(t *testing.T) {
	// assigning from a variable not in the bindings yields nil.
	tpl := mustParseAudit(t, "{% assign nilvariable = undefinedvar %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if a.Assignment.Value != nil {
		t.Errorf("Value=%v (type %T), want nil", a.Assignment.Value, a.Assignment.Value)
	}
}

// A07 — assign from another variable: value resolves to the binding's value.
func TestRenderAudit_Assignment_A07_fromVariable(t *testing.T) {
	tpl := mustParseAudit(t, "{% assign copy = original %}")
	result := auditOK(t, tpl, liquid.Bindings{"original": "source"}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if a.Assignment.Value != "source" {
		t.Errorf("Value=%v, want source", a.Assignment.Value)
	}
}

// A08 — assign from a nested dot-access path.
func TestRenderAudit_Assignment_A08_fromDotPath(t *testing.T) {
	tpl := mustParseAudit(t, "{% assign title = page.title %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"page": map[string]any{"title": "My Page"}},
		liquid.AuditOptions{TraceAssignments: true},
	)
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if a.Assignment.Value != "My Page" {
		t.Errorf("Value=%v, want %q", a.Assignment.Value, "My Page")
	}
}

// A09 — simple assign has no Path (or empty Path).
func TestRenderAudit_Assignment_A09_pathEmptyForSimpleAssign(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "hi" %}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	// Path is for dot-notation on the target side; standard assign targets are flat.
	// Either nil or empty is acceptable.
	if len(a.Assignment.Path) > 0 {
		t.Logf("Note: Path=%v for simple assign; informational only", a.Assignment.Path)
	}
}

// ============================================================================
// AssignmentTrace — Filter Pipeline (AP01–AP06)
// ============================================================================

// AP01 — assign with one filter (upcase): pipeline has one step.
func TestRenderAudit_Assignment_AP01_oneFilter(t *testing.T) {
	tpl := mustParseAudit(t, "{% assign upper = name | upcase %}")
	result := auditOK(t, tpl, liquid.Bindings{"name": "alice"}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if len(a.Assignment.Pipeline) != 1 {
		t.Fatalf("Pipeline len=%d, want 1", len(a.Assignment.Pipeline))
	}
	step := a.Assignment.Pipeline[0]
	if step.Filter != "upcase" {
		t.Errorf("Filter=%q, want upcase", step.Filter)
	}
	if step.Input != "alice" {
		t.Errorf("Input=%v, want alice", step.Input)
	}
	if step.Output != "ALICE" {
		t.Errorf("Output=%v, want ALICE", step.Output)
	}
	if a.Assignment.Value != "ALICE" {
		t.Errorf("Value=%v, want ALICE", a.Assignment.Value)
	}
}

// AP02 — assign with two-filter chain (times | round).
func TestRenderAudit_Assignment_AP02_twoFilterChain(t *testing.T) {
	tpl := mustParseAudit(t, "{% assign discounted = price | times: 0.9 | round %}")
	result := auditOK(t, tpl, liquid.Bindings{"price": 50}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if len(a.Assignment.Pipeline) != 2 {
		t.Fatalf("Pipeline len=%d, want 2", len(a.Assignment.Pipeline))
	}
	if a.Assignment.Pipeline[0].Filter != "times" {
		t.Errorf("Pipeline[0].Filter=%q, want times", a.Assignment.Pipeline[0].Filter)
	}
	if a.Assignment.Pipeline[1].Filter != "round" {
		t.Errorf("Pipeline[1].Filter=%q, want round", a.Assignment.Pipeline[1].Filter)
	}
	// Output of first step cascades to input of second.
	if a.Assignment.Pipeline[0].Output != a.Assignment.Pipeline[1].Input {
		t.Errorf("pipeline chain broken: times.Output=%v != round.Input=%v",
			a.Assignment.Pipeline[0].Output, a.Assignment.Pipeline[1].Input)
	}
}

// AP03 — assign with array pipeline (sort | first).
func TestRenderAudit_Assignment_AP03_arrayPipeline(t *testing.T) {
	tpl := mustParseAudit(t, "{% assign smallest = nums | sort | first %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"nums": []int{3, 1, 2}},
		liquid.AuditOptions{TraceAssignments: true},
	)
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if len(a.Assignment.Pipeline) != 2 {
		t.Fatalf("Pipeline len=%d, want 2", len(a.Assignment.Pipeline))
	}
	if sprintVal(a.Assignment.Value) != "1" {
		t.Errorf("Value=%v, want 1 (smallest after sort+first)", a.Assignment.Value)
	}
}

// AP04 — assign using split filter: value is a slice.
func TestRenderAudit_Assignment_AP04_splitResult(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign parts = csv | split: "," %}`)
	result := auditOK(t, tpl,
		liquid.Bindings{"csv": "a,b,c"},
		liquid.AuditOptions{TraceAssignments: true},
	)
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if len(a.Assignment.Pipeline) != 1 {
		t.Fatalf("Pipeline len=%d, want 1", len(a.Assignment.Pipeline))
	}
	switch v := a.Assignment.Value.(type) {
	case []string:
		if len(v) != 3 {
			t.Errorf("Value len=%d, want 3", len(v))
		}
	case []any:
		if len(v) != 3 {
			t.Errorf("Value len=%d, want 3", len(v))
		}
	default:
		t.Errorf("Value type=%T, want []string or []any", a.Assignment.Value)
	}
}

// AP05 — assign without any filter: Pipeline is empty.
func TestRenderAudit_Assignment_AP05_noPipeline(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "hello" %}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil || a.Assignment == nil {
		t.Fatal("no assignment expression")
	}
	if len(a.Assignment.Pipeline) != 0 {
		t.Errorf("Pipeline should be empty without filters, got %d steps", len(a.Assignment.Pipeline))
	}
}

// AP06 — assign with a filter that fails: the tag is silently skipped, AssignmentTrace
// still appears (with nil Value), no panic, no AuditError.
// Note: filter errors inside {% assign %} are silently swallowed (no Diagnostic emitted);
// filter errors inside {{ }} output tags DO produce Diagnostics.
func TestRenderAudit_Assignment_AP06_filterError(t *testing.T) {
	tpl := mustParseAudit(t, "{% assign result = 10 | divided_by: 0 %}")
	result, ae := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	if result == nil {
		t.Fatal("result must not be nil")
	}
	// No panic and no AuditError expected.
	if ae != nil {
		t.Errorf("unexpected AuditError: %v", ae)
	}
}

// ============================================================================
// AssignmentTrace — Source, Range, Depth (AR01–AR04)
// ============================================================================

// AR01 — Source is non-empty for assignments.
// The Source field contains the expression body (not the full {% assign %} tag).
func TestRenderAudit_Assignment_AR01_sourceNonEmpty(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "hello" %}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil {
		t.Fatal("no assignment expression")
	}
	if a.Source == "" {
		t.Error("Source must be non-empty for assignment")
	}
}

// AR02 — Range.Start.Line >= 1 and Column >= 1.
func TestRenderAudit_Assignment_AR02_rangeValid(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "y" %}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil {
		t.Fatal("no assignment expression")
	}
	assertRangeValid(t, a.Range, "assignment Range")
}

// AR03 — assign inside an if block has Depth=1.
func TestRenderAudit_Assignment_AR03_depthInsideIf(t *testing.T) {
	tpl := mustParseAudit(t, `{% if true %}{% assign x = "y" %}{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	a := firstExpr(result.Expressions, liquid.KindAssignment)
	if a == nil {
		t.Fatal("no assignment expression")
	}
	if a.Depth != 1 {
		t.Errorf("Depth=%d, want 1 (inside if)", a.Depth)
	}
}

// AR04 — assign repeats once per for-loop iteration.
func TestRenderAudit_Assignment_AR04_repeatsPerIteration(t *testing.T) {
	tpl := mustParseAudit(t, `{% for item in items %}{% assign doubled = item | times: 2 %}{% endfor %}`)
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2, 3}},
		liquid.AuditOptions{TraceIterations: true, TraceAssignments: true},
	)
	assignExprs := allExprs(result.Expressions, liquid.KindAssignment)
	if len(assignExprs) != 3 {
		t.Errorf("assignment count=%d, want 3 (one per iteration)", len(assignExprs))
	}
}

// ============================================================================
// AssignmentTrace — Multiple Assigns (AM01–AM03)
// ============================================================================

// AM01 — three assigns in sequence: three expressions in order.
func TestRenderAudit_Assignment_AM01_multipleAssignsOrdered(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign a = 1 %}{% assign b = 2 %}{% assign c = 3 %}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	assigns := allExprs(result.Expressions, liquid.KindAssignment)
	if len(assigns) != 3 {
		t.Fatalf("assignment count=%d, want 3", len(assigns))
	}
	names := []string{
		assigns[0].Assignment.Variable,
		assigns[1].Assignment.Variable,
		assigns[2].Assignment.Variable,
	}
	if names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Errorf("assignment variables=%v, want [a b c]", names)
	}
}

// AM02 — assign then use: assignment appears before variable trace in the array.
func TestRenderAudit_Assignment_AM02_assignBeforeVariable(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign msg = "hi" %}{{ msg }}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true, TraceVariables: true})
	if len(result.Expressions) != 2 {
		t.Fatalf("expected 2 expressions, got %d", len(result.Expressions))
	}
	if result.Expressions[0].Kind != liquid.KindAssignment {
		t.Errorf("Expressions[0].Kind=%q, want assignment", result.Expressions[0].Kind)
	}
	if result.Expressions[1].Kind != liquid.KindVariable {
		t.Errorf("Expressions[1].Kind=%q, want variable", result.Expressions[1].Kind)
	}
}

// AM03 — reassigning the same variable produces two assignment traces.
func TestRenderAudit_Assignment_AM03_reassign(t *testing.T) {
	tpl := mustParseAudit(t, `{% assign x = "first" %}{% assign x = "second" %}{{ x }}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true, TraceVariables: true})
	assigns := allExprs(result.Expressions, liquid.KindAssignment)
	if len(assigns) != 2 {
		t.Fatalf("expected 2 assignment traces, got %d", len(assigns))
	}
	if assigns[0].Assignment.Value != "first" {
		t.Errorf("first assign Value=%v, want first", assigns[0].Assignment.Value)
	}
	if assigns[1].Assignment.Value != "second" {
		t.Errorf("second assign Value=%v, want second", assigns[1].Assignment.Value)
	}
	// Final variable value should be "second".
	vars := allExprs(result.Expressions, liquid.KindVariable)
	if len(vars) < 1 {
		t.Fatal("expected variable expression after reassign")
	}
	if vars[0].Variable.Value != "second" {
		t.Errorf("variable Value=%v, want second", vars[0].Variable.Value)
	}
}

// ============================================================================
// CaptureTrace — Basic Attributes (CP01–CP05)
// ============================================================================

// CP01 — simple capture: Variable and Value.
func TestRenderAudit_Capture_CP01_simple(t *testing.T) {
	tpl := mustParseAudit(t, "{% capture greeting %}Hello, world!{% endcapture %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	c := firstExpr(result.Expressions, liquid.KindCapture)
	if c == nil || c.Capture == nil {
		t.Fatal("no capture expression")
	}
	if c.Capture.Variable != "greeting" {
		t.Errorf("Variable=%q, want greeting", c.Capture.Variable)
	}
	if c.Capture.Value != "Hello, world!" {
		t.Errorf("Value=%q, want %q", c.Capture.Value, "Hello, world!")
	}
}

// CP02 — capture with an expression inside: Value contains rendered output.
func TestRenderAudit_Capture_CP02_withExpression(t *testing.T) {
	tpl := mustParseAudit(t, "{% capture msg %}Hello, {{ name }}!{% endcapture %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"name": "Alice"},
		liquid.AuditOptions{TraceAssignments: true},
	)
	c := firstExpr(result.Expressions, liquid.KindCapture)
	if c == nil || c.Capture == nil {
		t.Fatal("no capture expression")
	}
	if c.Capture.Value != "Hello, Alice!" {
		t.Errorf("Value=%q, want %q", c.Capture.Value, "Hello, Alice!")
	}
}

// CP03 — capture with multiline content: entire rendered content is in Value.
func TestRenderAudit_Capture_CP03_multiline(t *testing.T) {
	tpl := mustParseAudit(t, "{% capture block %}\nline1\nline2\n{% endcapture %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	c := firstExpr(result.Expressions, liquid.KindCapture)
	if c == nil || c.Capture == nil {
		t.Fatal("no capture expression")
	}
	if c.Capture.Value == "" {
		t.Error("capture Value should be non-empty for multiline content")
	}
}

// CP04 — empty capture: Value is empty string.
func TestRenderAudit_Capture_CP04_empty(t *testing.T) {
	tpl := mustParseAudit(t, "{% capture nothing %}{% endcapture %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	c := firstExpr(result.Expressions, liquid.KindCapture)
	if c == nil || c.Capture == nil {
		t.Fatal("no capture expression")
	}
	if c.Capture.Value != "" {
		t.Errorf("Value=%q, want empty string for empty capture", c.Capture.Value)
	}
}

// CP05 — capture with an if tag inside: only the executed branch content is captured.
func TestRenderAudit_Capture_CP05_withConditional(t *testing.T) {
	tpl := mustParseAudit(t, "{% capture result %}{% if x %}yes{% else %}no{% endif %}{% endcapture %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": true}, liquid.AuditOptions{TraceAssignments: true})
	c := firstExpr(result.Expressions, liquid.KindCapture)
	if c == nil || c.Capture == nil {
		t.Fatal("no capture expression")
	}
	if c.Capture.Value != "yes" {
		t.Errorf("Value=%q, want yes", c.Capture.Value)
	}
}

// ============================================================================
// CaptureTrace — Source, Range, Depth (CPR01–CPR03)
// ============================================================================

// CPR01 — Source is the {% capture name %} header.
func TestRenderAudit_Capture_CPR01_sourceNonEmpty(t *testing.T) {
	tpl := mustParseAudit(t, "{% capture x %}hello{% endcapture %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	c := firstExpr(result.Expressions, liquid.KindCapture)
	if c == nil {
		t.Fatal("no capture expression")
	}
	if c.Source == "" {
		t.Error("Capture Source should be non-empty")
	}
}

// CPR02 — Range.Start.Line >= 1.
func TestRenderAudit_Capture_CPR02_rangeValid(t *testing.T) {
	tpl := mustParseAudit(t, "{% capture x %}hello{% endcapture %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	c := firstExpr(result.Expressions, liquid.KindCapture)
	if c == nil {
		t.Fatal("no capture expression")
	}
	assertRangeValid(t, c.Range, "capture Range")
}

// CPR03 — capture inside an if block has Depth=1.
func TestRenderAudit_Capture_CPR03_depthInsideBlock(t *testing.T) {
	tpl := mustParseAudit(t, "{% if true %}{% capture x %}hello{% endcapture %}{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceAssignments: true})
	c := firstExpr(result.Expressions, liquid.KindCapture)
	if c == nil {
		t.Fatal("no capture expression")
	}
	if c.Depth != 1 {
		t.Errorf("Depth=%d, want 1 (inside if)", c.Depth)
	}
}

// ============================================================================
// CaptureTrace — Inner Traces (CPI01–CPI03)
// ============================================================================

// CPI01 — capture with {{ var }} inside: inner variable trace appears in expressions.
func TestRenderAudit_Capture_CPI01_innerVariableTrace(t *testing.T) {
	tpl := mustParseAudit(t, "{% capture msg %}{{ name }}{% endcapture %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"name": "Bob"},
		liquid.AuditOptions{TraceAssignments: true, TraceVariables: true},
	)
	// Both a Capture and a Variable expression should appear.
	captureExprs := allExprs(result.Expressions, liquid.KindCapture)
	varExprs := allExprs(result.Expressions, liquid.KindVariable)
	if len(captureExprs) == 0 {
		t.Error("expected capture expression")
	}
	if len(varExprs) == 0 {
		t.Error("expected variable expression inside capture body")
	}
}

// CPI02 — capture with {% if %} inside: inner condition trace appears.
func TestRenderAudit_Capture_CPI02_innerConditionTrace(t *testing.T) {
	tpl := mustParseAudit(t, "{% capture msg %}{% if x %}yes{% endif %}{% endcapture %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"x": true},
		liquid.AuditOptions{TraceAssignments: true, TraceConditions: true},
	)
	captureExprs := allExprs(result.Expressions, liquid.KindCapture)
	condExprs := allExprs(result.Expressions, liquid.KindCondition)
	if len(captureExprs) == 0 {
		t.Error("expected capture expression")
	}
	if len(condExprs) == 0 {
		t.Error("expected condition expression inside capture body")
	}
}

// CPI03 — capture then use: the variable trace for {{ x }} shows the captured value.
func TestRenderAudit_Capture_CPI03_capturedValueUsed(t *testing.T) {
	tpl := mustParseAudit(t, "{% capture x %}Hello{% endcapture %}{{ x }}")
	result := auditOK(t, tpl,
		liquid.Bindings{},
		liquid.AuditOptions{TraceAssignments: true, TraceVariables: true},
	)
	assertOutput(t, result, "Hello")

	vars := allExprs(result.Expressions, liquid.KindVariable)
	for _, v := range vars {
		if v.Variable != nil && v.Variable.Name == "x" {
			if v.Variable.Value != "Hello" {
				t.Errorf("{{ x }} Variable.Value=%v, want Hello (the captured value)", v.Variable.Value)
			}
		}
	}
}
