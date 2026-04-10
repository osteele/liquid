package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// VariableTrace — Name, Parts, Value (V01–V16)
// ============================================================================

// V01 — simple single-segment variable.
func TestRenderAudit_Variable_V01_simple(t *testing.T) {
	tpl := mustParseAudit(t, "{{ x }}")
	result := auditOK(t, tpl, liquid.Bindings{"x": "hello"}, liquid.AuditOptions{TraceVariables: true})
	assertExprCount(t, result, 1)
	e := result.Expressions[0]
	if e.Kind != liquid.KindVariable {
		t.Fatalf("Kind=%q, want %q", e.Kind, liquid.KindVariable)
	}
	if e.Variable == nil {
		t.Fatal("Variable is nil")
	}
	if e.Variable.Name != "x" {
		t.Errorf("Name=%q, want %q", e.Variable.Name, "x")
	}
	if len(e.Variable.Parts) != 1 || e.Variable.Parts[0] != "x" {
		t.Errorf("Parts=%v, want [\"x\"]", e.Variable.Parts)
	}
	if e.Variable.Value != "hello" {
		t.Errorf("Value=%v, want %q", e.Variable.Value, "hello")
	}
}

// V02 — dot-access two-level path.
func TestRenderAudit_Variable_V02_dotAccess(t *testing.T) {
	tpl := mustParseAudit(t, "{{ customer.name }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"customer": map[string]any{"name": "Alice"}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Name != "customer.name" {
		t.Errorf("Name=%q, want %q", v.Variable.Name, "customer.name")
	}
	if len(v.Variable.Parts) != 2 {
		t.Fatalf("len(Parts)=%d, want 2", len(v.Variable.Parts))
	}
	if v.Variable.Parts[0] != "customer" || v.Variable.Parts[1] != "name" {
		t.Errorf("Parts=%v, want [customer name]", v.Variable.Parts)
	}
	if v.Variable.Value != "Alice" {
		t.Errorf("Value=%v, want Alice", v.Variable.Value)
	}
}

// V03 — deep dot-access four-level path.
func TestRenderAudit_Variable_V03_deepDotAccess(t *testing.T) {
	tpl := mustParseAudit(t, "{{ a.b.c.d }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"a": map[string]any{"b": map[string]any{"c": map[string]any{"d": "deep"}}}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Name != "a.b.c.d" {
		t.Errorf("Name=%q, want %q", v.Variable.Name, "a.b.c.d")
	}
	if len(v.Variable.Parts) != 4 {
		t.Fatalf("len(Parts)=%d, want 4", len(v.Variable.Parts))
	}
	if v.Variable.Value != "deep" {
		t.Errorf("Value=%v, want %q", v.Variable.Value, "deep")
	}
}

// V04 — array index access via bracket notation.
func TestRenderAudit_Variable_V04_arrayIndex(t *testing.T) {
	tpl := mustParseAudit(t, "{{ items[0] }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"alpha", "beta"}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Value != "alpha" {
		t.Errorf("Value=%v, want %q", v.Variable.Value, "alpha")
	}
	// Name/Parts may vary by implementation; just verify they are non-empty.
	if v.Variable.Name == "" {
		t.Error("Name should be non-empty for bracket access")
	}
}

// V05 — string literal in an object expression.
func TestRenderAudit_Variable_V05_stringLiteral(t *testing.T) {
	tpl := mustParseAudit(t, `{{ "hello" }}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Value != "hello" {
		t.Errorf("Value=%v, want %q", v.Variable.Value, "hello")
	}
}

// V06 — integer literal.
func TestRenderAudit_Variable_V06_intLiteral(t *testing.T) {
	tpl := mustParseAudit(t, "{{ 42 }}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if sprintVal(v.Variable.Value) != "42" {
		t.Errorf("Value=%v, want 42", v.Variable.Value)
	}
}

// V07 — float literal.
func TestRenderAudit_Variable_V07_floatLiteral(t *testing.T) {
	tpl := mustParseAudit(t, "{{ 3.14 }}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if sprintVal(v.Variable.Value) != "3.14" {
		t.Errorf("Value=%v, want 3.14", v.Variable.Value)
	}
}

// V08 — boolean true literal.
func TestRenderAudit_Variable_V08_boolTrue(t *testing.T) {
	tpl := mustParseAudit(t, "{{ true }}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Value != true {
		t.Errorf("Value=%v, want true", v.Variable.Value)
	}
}

// V09 — boolean false literal.
func TestRenderAudit_Variable_V09_boolFalse(t *testing.T) {
	tpl := mustParseAudit(t, "{{ false }}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Value != false {
		t.Errorf("Value=%v, want false", v.Variable.Value)
	}
}

// V10 — nil literal renders as empty string; value is nil.
func TestRenderAudit_Variable_V10_nilLiteral(t *testing.T) {
	tpl := mustParseAudit(t, "{{ nil }}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceVariables: true})
	assertOutput(t, result, "")
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Value != nil {
		t.Errorf("Value=%v, want nil", v.Variable.Value)
	}
}

// V13 — undefined variable without StrictVariables → Value nil, no error.
func TestRenderAudit_Variable_V13_undefinedNoStrict(t *testing.T) {
	tpl := mustParseAudit(t, "{{ ghost }}")
	result, ae := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{TraceVariables: true})
	if ae != nil {
		t.Fatalf("unexpected AuditError without StrictVariables: %v", ae)
	}
	assertNoDiags(t, result)
	assertOutput(t, result, "")
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression (undefined vars are still traced)")
	}
	if v.Variable.Value != nil {
		t.Errorf("Value=%v, want nil for undefined var", v.Variable.Value)
	}
}

// V14 — undefined variable WITH StrictVariables → Error on expression + Diagnostic.
func TestRenderAudit_Variable_V14_undefinedWithStrict(t *testing.T) {
	tpl := mustParseAudit(t, "{{ ghost }}")
	result, ae := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{TraceVariables: true}, liquid.WithStrictVariables())
	if result == nil {
		t.Fatal("result is nil")
	}
	if ae == nil {
		t.Fatal("expected AuditError for undefined variable with StrictVariables")
	}
	d := firstDiag(result.Diagnostics, "undefined-variable")
	if d == nil {
		t.Fatal("expected undefined-variable diagnostic")
	}
	// Expression should also carry the error reference.
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("variable expression should still appear even when it errored")
	}
	if v.Error == nil {
		t.Error("Expression.Error should be non-nil when variable caused an error")
	}
}

// V15 — multiple variables in sequence; all traced.
func TestRenderAudit_Variable_V15_multipleVars(t *testing.T) {
	tpl := mustParseAudit(t, "{{ a }}{{ b }}{{ c }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"a": 1, "b": 2, "c": 3},
		liquid.AuditOptions{TraceVariables: true},
	)
	assertExprCount(t, result, 3)
	for i, e := range result.Expressions {
		if e.Kind != liquid.KindVariable {
			t.Errorf("Expressions[%d].Kind=%q, want variable", i, e.Kind)
		}
	}
	names := []string{
		result.Expressions[0].Variable.Name,
		result.Expressions[1].Variable.Name,
		result.Expressions[2].Variable.Name,
	}
	if names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Errorf("Names=%v, want [a b c]", names)
	}
}

// V16 — bracket string-key access on a map.
func TestRenderAudit_Variable_V16_bracketStringKey(t *testing.T) {
	tpl := mustParseAudit(t, `{{ hash["key"] }}`)
	result := auditOK(t, tpl,
		liquid.Bindings{"hash": map[string]any{"key": "val"}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Value != "val" {
		t.Errorf("Value=%v, want %q", v.Variable.Value, "val")
	}
}

// ============================================================================
// VariableTrace — Filter Pipeline (VP01–VP24)
// ============================================================================

// VP01 — no filters → Pipeline is empty (not nil).
func TestRenderAudit_Variable_VP01_noPipeline(t *testing.T) {
	tpl := mustParseAudit(t, "{{ name }}")
	result := auditOK(t, tpl, liquid.Bindings{"name": "alice"}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 0 {
		t.Errorf("Pipeline should be empty when no filters, got %d steps", len(v.Variable.Pipeline))
	}
}

// VP02 — single filter, no args (upcase).
func TestRenderAudit_Variable_VP02_singleFilterNoArgs(t *testing.T) {
	tpl := mustParseAudit(t, "{{ name | upcase }}")
	result := auditOK(t, tpl, liquid.Bindings{"name": "alice"}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 {
		t.Fatalf("Pipeline len=%d, want 1", len(v.Variable.Pipeline))
	}
	step := v.Variable.Pipeline[0]
	if step.Filter != "upcase" {
		t.Errorf("Filter=%q, want %q", step.Filter, "upcase")
	}
	if len(step.Args) != 0 {
		t.Errorf("Args=%v, want []", step.Args)
	}
	if step.Input != "alice" {
		t.Errorf("Input=%v, want %q", step.Input, "alice")
	}
	if step.Output != "ALICE" {
		t.Errorf("Output=%v, want %q", step.Output, "ALICE")
	}
}

// VP03 — single filter with one integer arg (truncate: 5).
func TestRenderAudit_Variable_VP03_singleFilterOneArg(t *testing.T) {
	tpl := mustParseAudit(t, `{{ msg | truncate: 8 }}`)
	result := auditOK(t, tpl, liquid.Bindings{"msg": "hello world"}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 {
		t.Fatalf("Pipeline len=%d, want 1", len(v.Variable.Pipeline))
	}
	step := v.Variable.Pipeline[0]
	if step.Filter != "truncate" {
		t.Errorf("Filter=%q, want truncate", step.Filter)
	}
	if len(step.Args) == 0 {
		t.Error("Args should not be empty for truncate: 8")
	}
	if sprintVal(step.Args[0]) != "8" {
		t.Errorf("Args[0]=%v, want 8", step.Args[0])
	}
}

// VP04 — single filter with two args (truncate: 10, "...").
func TestRenderAudit_Variable_VP04_singleFilterTwoArgs(t *testing.T) {
	tpl := mustParseAudit(t, `{{ msg | truncate: 10, "~" }}`)
	result := auditOK(t, tpl, liquid.Bindings{"msg": "hello, world!"}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 {
		t.Fatalf("Pipeline len=%d, want 1", len(v.Variable.Pipeline))
	}
	step := v.Variable.Pipeline[0]
	if step.Filter != "truncate" {
		t.Errorf("Filter=%q, want truncate", step.Filter)
	}
	if len(step.Args) < 2 {
		t.Fatalf("Args len=%d, want >= 2", len(step.Args))
	}
	if step.Args[1] != "~" {
		t.Errorf("Args[1]=%v, want %q", step.Args[1], "~")
	}
}

// VP05 — chain of two filters: Output[0] == Input[1].
func TestRenderAudit_Variable_VP05_twoFilterChain(t *testing.T) {
	tpl := mustParseAudit(t, "{{ name | upcase | truncate: 3 }}")
	result := auditOK(t, tpl, liquid.Bindings{"name": "alice"}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 2 {
		t.Fatalf("Pipeline len=%d, want 2", len(v.Variable.Pipeline))
	}
	step0, step1 := v.Variable.Pipeline[0], v.Variable.Pipeline[1]
	if step0.Filter != "upcase" {
		t.Errorf("Pipeline[0].Filter=%q, want upcase", step0.Filter)
	}
	if step1.Filter != "truncate" {
		t.Errorf("Pipeline[1].Filter=%q, want truncate", step1.Filter)
	}
	// Output of step0 must equal Input of step1.
	if step0.Output != step1.Input {
		t.Errorf("step0.Output=%v != step1.Input=%v (chain broken)", step0.Output, step1.Input)
	}
	// Final value on the trace should be step1.Output.
	if v.Variable.Value != step1.Output {
		t.Errorf("Variable.Value=%v != Pipeline[-1].Output=%v", v.Variable.Value, step1.Output)
	}
}

// VP06 — chain of three filters: downcase | prepend | upcase.
func TestRenderAudit_Variable_VP06_threeFilterChain(t *testing.T) {
	tpl := mustParseAudit(t, `{{ name | downcase | prepend: "hi " | upcase }}`)
	result := auditOK(t, tpl, liquid.Bindings{"name": "ALICE"}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 3 {
		t.Fatalf("Pipeline len=%d, want 3", len(v.Variable.Pipeline))
	}
	// Each step's Output should equal the next step's Input.
	for i := range 2 {
		if v.Variable.Pipeline[i].Output != v.Variable.Pipeline[i+1].Input {
			t.Errorf("pipeline chain broken between step %d and %d", i, i+1)
		}
	}
	// Final value.
	expected := "HI ALICE"
	if v.Variable.Value != expected {
		t.Errorf("Value=%v, want %q", v.Variable.Value, expected)
	}
}

// VP07 — filter `default` with nil value.
func TestRenderAudit_Variable_VP07_defaultFilter(t *testing.T) {
	tpl := mustParseAudit(t, `{{ missing | default: "fallback" }}`)
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 {
		t.Fatalf("Pipeline len=%d, want 1", len(v.Variable.Pipeline))
	}
	step := v.Variable.Pipeline[0]
	if step.Filter != "default" {
		t.Errorf("Filter=%q, want default", step.Filter)
	}
	if step.Output != "fallback" {
		t.Errorf("Output=%v, want fallback", step.Output)
	}
	if v.Variable.Value != "fallback" {
		t.Errorf("Value=%v, want fallback", v.Variable.Value)
	}
}

// VP08 — filter `split` returns a slice.
func TestRenderAudit_Variable_VP08_splitFilter(t *testing.T) {
	tpl := mustParseAudit(t, `{{ csv | split: "," }}`)
	result := auditOK(t, tpl, liquid.Bindings{"csv": "a,b,c"}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 {
		t.Fatalf("Pipeline len=%d, want 1", len(v.Variable.Pipeline))
	}
	step := v.Variable.Pipeline[0]
	if step.Input != "a,b,c" {
		t.Errorf("Input=%v, want %q", step.Input, "a,b,c")
	}
	// Output should be a slice of strings.
	out, ok := step.Output.([]string)
	if !ok {
		t.Errorf("Output type=%T, want []string", step.Output)
	} else if len(out) != 3 {
		t.Errorf("output slice len=%d, want 3", len(out))
	}
}

// VP09 — filter `size` on a string returns its length.
func TestRenderAudit_Variable_VP09_sizeOnString(t *testing.T) {
	tpl := mustParseAudit(t, "{{ word | size }}")
	result := auditOK(t, tpl, liquid.Bindings{"word": "hello"}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if sprintVal(v.Variable.Value) != "5" {
		t.Errorf("Value=%v, want 5", v.Variable.Value)
	}
}

// VP10 — filter `size` on an array returns its length.
func TestRenderAudit_Variable_VP10_sizeOnArray(t *testing.T) {
	tpl := mustParseAudit(t, "{{ items | size }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{1, 2, 3, 4}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if sprintVal(v.Variable.Value) != "4" {
		t.Errorf("Value=%v, want 4", v.Variable.Value)
	}
}

// VP11 — filter `times` on a number.
func TestRenderAudit_Variable_VP11_timesFilter(t *testing.T) {
	tpl := mustParseAudit(t, "{{ price | times: 2 }}")
	result := auditOK(t, tpl, liquid.Bindings{"price": 10}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 {
		t.Fatalf("Pipeline len=%d, want 1", len(v.Variable.Pipeline))
	}
	step := v.Variable.Pipeline[0]
	if step.Filter != "times" {
		t.Errorf("Filter=%q, want times", step.Filter)
	}
	if sprintVal(step.Output) != "20" {
		t.Errorf("Output=%v, want 20", step.Output)
	}
}

// VP12 — filter `round` converts float to int.
func TestRenderAudit_Variable_VP12_roundFilter(t *testing.T) {
	tpl := mustParseAudit(t, "{{ price | round }}")
	result := auditOK(t, tpl, liquid.Bindings{"price": 3.7}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if sprintVal(v.Variable.Value) != "4" {
		t.Errorf("Value=%v, want 4", v.Variable.Value)
	}
}

// VP13 — filter `join` on an array produces a string.
func TestRenderAudit_Variable_VP13_joinFilter(t *testing.T) {
	tpl := mustParseAudit(t, `{{ tags | join: ", " }}`)
	result := auditOK(t, tpl,
		liquid.Bindings{"tags": []string{"go", "liquid", "test"}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Value != "go, liquid, test" {
		t.Errorf("Value=%v, want %q", v.Variable.Value, "go, liquid, test")
	}
}

// VP15 — filter `first` on an array.
func TestRenderAudit_Variable_VP15_firstFilter(t *testing.T) {
	tpl := mustParseAudit(t, "{{ items | first }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"a", "b", "c"}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Value != "a" {
		t.Errorf("Value=%v, want %q", v.Variable.Value, "a")
	}
}

// VP16 — filter `last` on an array.
func TestRenderAudit_Variable_VP16_lastFilter(t *testing.T) {
	tpl := mustParseAudit(t, "{{ items | last }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"a", "b", "c"}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if v.Variable.Value != "c" {
		t.Errorf("Value=%v, want %q", v.Variable.Value, "c")
	}
}

// VP17 — filter `map` returns a slice of extracted values.
func TestRenderAudit_Variable_VP17_mapFilter(t *testing.T) {
	tpl := mustParseAudit(t, `{{ products | map: "name" }}`)
	result := auditOK(t, tpl,
		liquid.Bindings{
			"products": []map[string]any{
				{"name": "Widget", "price": 10},
				{"name": "Gadget", "price": 20},
			},
		},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 {
		t.Fatalf("Pipeline len=%d, want 1", len(v.Variable.Pipeline))
	}
	step := v.Variable.Pipeline[0]
	if step.Filter != "map" {
		t.Errorf("Filter=%q, want map", step.Filter)
	}
	// Output should be a slice.
	switch step.Output.(type) {
	case []any, []string:
		// acceptable
	default:
		t.Errorf("Output type=%T, want slice", step.Output)
	}
}

// VP18 — filter `where` on an array.
func TestRenderAudit_Variable_VP18_whereFilter(t *testing.T) {
	tpl := mustParseAudit(t, `{{ products | where: "active", true }}`)
	result := auditOK(t, tpl,
		liquid.Bindings{
			"products": []map[string]any{
				{"name": "A", "active": true},
				{"name": "B", "active": false},
			},
		},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 {
		t.Fatalf("Pipeline len=%d, want 1", len(v.Variable.Pipeline))
	}
}

// VP19 — filter `sort` on a numeric array.
func TestRenderAudit_Variable_VP19_sortFilter(t *testing.T) {
	tpl := mustParseAudit(t, "{{ nums | sort }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"nums": []int{3, 1, 2}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 || v.Variable.Pipeline[0].Filter != "sort" {
		t.Error("expected sort filter step")
	}
}

// VP20 — filter `reverse` on an array.
func TestRenderAudit_Variable_VP20_reverseFilter(t *testing.T) {
	tpl := mustParseAudit(t, "{{ items | reverse }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"a", "b", "c"}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 || v.Variable.Pipeline[0].Filter != "reverse" {
		t.Error("expected reverse filter step")
	}
}

// VP21 — filter `compact` removes nil values.
func TestRenderAudit_Variable_VP21_compactFilter(t *testing.T) {
	tpl := mustParseAudit(t, "{{ items | compact }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []any{"a", nil, "b", nil, "c"}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 || v.Variable.Pipeline[0].Filter != "compact" {
		t.Error("expected compact filter step")
	}
}

// VP22 — filter `uniq` removes duplicates.
func TestRenderAudit_Variable_VP22_uniqFilter(t *testing.T) {
	tpl := mustParseAudit(t, "{{ items | uniq }}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []string{"a", "b", "a", "c"}},
		liquid.AuditOptions{TraceVariables: true},
	)
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil || v.Variable == nil {
		t.Fatal("no variable expression")
	}
	if len(v.Variable.Pipeline) != 1 || v.Variable.Pipeline[0].Filter != "uniq" {
		t.Error("expected uniq filter step")
	}
}

// VP23 — undefined filter with LaxFilters → no error, value passes through.
func TestRenderAudit_Variable_VP23_laxFilters(t *testing.T) {
	tpl := mustParseAudit(t, "{{ name | no_such_filter }}")
	result, ae := tpl.RenderAudit(
		liquid.Bindings{"name": "alice"},
		liquid.AuditOptions{TraceVariables: true},
		liquid.WithLaxFilters(),
	)
	if result == nil {
		t.Fatal("result is nil")
	}
	if ae != nil {
		t.Fatalf("unexpected AuditError with LaxFilters: %v", ae)
	}
}

// VP24 — filter that causes an error (divided_by: 0) → Error on expression + Diagnostic.
func TestRenderAudit_Variable_VP24_filterError(t *testing.T) {
	tpl := mustParseAudit(t, "{{ 10 | divided_by: 0 }}")
	result, _ := tpl.RenderAudit(liquid.Bindings{}, liquid.AuditOptions{TraceVariables: true})
	if result == nil {
		t.Fatal("result is nil")
	}
	d := firstDiag(result.Diagnostics, "argument-error")
	if d == nil {
		t.Fatal("expected argument-error diagnostic")
	}
}

// ============================================================================
// VariableTrace — Source and Range (VR01–VR07)
// ============================================================================

// VR01 — Source includes the {{ }} delimiters.
func TestRenderAudit_Variable_VR01_sourceIncludesDelimiters(t *testing.T) {
	tpl := mustParseAudit(t, "{{ name }}")
	result := auditOK(t, tpl, liquid.Bindings{"name": "x"}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Source != "{{ name }}" {
		t.Errorf("Source=%q, want %q", v.Source, "{{ name }}")
	}
}

// VR02 — Range.Start.Line = 1 when expression is on first line.
func TestRenderAudit_Variable_VR02_lineOne(t *testing.T) {
	tpl := mustParseAudit(t, "{{ x }}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Range.Start.Line != 1 {
		t.Errorf("Range.Start.Line=%d, want 1", v.Range.Start.Line)
	}
}

// VR03 — Range.Start.Line = 3 when expression is on third line.
func TestRenderAudit_Variable_VR03_lineThree(t *testing.T) {
	tpl := mustParseAudit(t, "line1\nline2\n{{ x }}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Range.Start.Line != 3 {
		t.Errorf("Range.Start.Line=%d, want 3", v.Range.Start.Line)
	}
}

// VR04 — Range.Start.Column >= 1 (never zero).
func TestRenderAudit_Variable_VR04_columnAtLeastOne(t *testing.T) {
	tpl := mustParseAudit(t, "{{ x }}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Range.Start.Column < 1 {
		t.Errorf("Range.Start.Column=%d, want >= 1", v.Range.Start.Column)
	}
}

// VR05 — Range.End is after Range.Start (non-zero span).
func TestRenderAudit_Variable_VR05_rangeIsSpan(t *testing.T) {
	tpl := mustParseAudit(t, "{{ name }}")
	result := auditOK(t, tpl, liquid.Bindings{"name": "x"}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	assertRangeSpan(t, v.Range, "variable {{ name }}")
}

// VR06 — Range.End.Column = Start.Column + len("{{ name }}") for single-line expression at col 1.
func TestRenderAudit_Variable_VR06_endColumnPrecise(t *testing.T) {
	// "{{ name }}" is 10 chars; at col 1, End.Column should be 11 (exclusive).
	tpl := mustParseAudit(t, "{{ name }}")
	result := auditOK(t, tpl, liquid.Bindings{"name": "x"}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	src := "{{ name }}"
	wantEndCol := v.Range.Start.Column + len(src)
	if v.Range.End.Column != wantEndCol {
		t.Errorf("Range.End.Column=%d, want %d (Start.Column + len(source))", v.Range.End.Column, wantEndCol)
	}
}

// VR07 — Multiple expressions in same template have non-overlapping Ranges.
func TestRenderAudit_Variable_VR07_noOverlappingRanges(t *testing.T) {
	tpl := mustParseAudit(t, "{{ a }} {{ b }}")
	result := auditOK(t, tpl, liquid.Bindings{"a": 1, "b": 2}, liquid.AuditOptions{TraceVariables: true})
	if len(result.Expressions) < 2 {
		t.Fatalf("expected 2 expressions, got %d", len(result.Expressions))
	}
	r0 := result.Expressions[0].Range
	r1 := result.Expressions[1].Range
	// r1.Start must be after r0.End
	if r1.Start.Line < r0.End.Line ||
		(r1.Start.Line == r0.End.Line && r1.Start.Column < r0.End.Column) {
		t.Errorf("ranges overlap: r0=[%v→%v] r1=[%v→%v]", r0.Start, r0.End, r1.Start, r1.End)
	}
}

// ============================================================================
// VariableTrace — Depth (VD01–VD06)
// ============================================================================

// VD01 — top-level variable has Depth = 0.
func TestRenderAudit_Variable_VD01_depthZeroTopLevel(t *testing.T) {
	tpl := mustParseAudit(t, "{{ x }}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Depth != 0 {
		t.Errorf("Depth=%d, want 0 for top-level variable", v.Depth)
	}
}

// VD02 — variable inside one {% if %} block has Depth = 1.
func TestRenderAudit_Variable_VD02_depthOneInsideIf(t *testing.T) {
	tpl := mustParseAudit(t, "{% if true %}{{ x }}{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Depth != 1 {
		t.Errorf("Depth=%d, want 1 (inside if)", v.Depth)
	}
}

// VD03 — variable inside one {% for %} block has Depth = 1.
func TestRenderAudit_Variable_VD03_depthOneInsideFor(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{{ item }}{% endfor %}")
	result := auditOK(t, tpl, liquid.Bindings{"items": []int{1}}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Depth != 1 {
		t.Errorf("Depth=%d, want 1 (inside for)", v.Depth)
	}
}

// VD04 — variable inside nested {% if %}{% if %} has Depth = 2.
func TestRenderAudit_Variable_VD04_depthTwoNestedIf(t *testing.T) {
	tpl := mustParseAudit(t, "{% if true %}{% if true %}{{ x }}{% endif %}{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Depth != 2 {
		t.Errorf("Depth=%d, want 2 (nested if×if)", v.Depth)
	}
}

// VD05 — variable inside {% for %}{% if %} has Depth = 2.
func TestRenderAudit_Variable_VD05_depthTwoForIf(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{% if true %}{{ item }}{% endif %}{% endfor %}")
	result := auditOK(t, tpl, liquid.Bindings{"items": []int{1}}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Depth != 2 {
		t.Errorf("Depth=%d, want 2 (for > if)", v.Depth)
	}
}

// VD06 — after exiting a block, subsequent top-level variable is Depth = 0.
func TestRenderAudit_Variable_VD06_depthResetsAfterBlock(t *testing.T) {
	tpl := mustParseAudit(t, "{% if true %}{% endif %}{{ x }}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceVariables: true})
	v := firstExpr(result.Expressions, liquid.KindVariable)
	if v == nil {
		t.Fatal("no variable expression")
	}
	if v.Depth != 0 {
		t.Errorf("Depth=%d, want 0 (after block exits)", v.Depth)
	}
}
