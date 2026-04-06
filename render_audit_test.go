package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// newEngine is a test helper that creates a default Engine.
func newAuditEngine() *liquid.Engine {
	return liquid.NewEngine()
}

// --------------------------------------------------------------------------
// RenderAudit — TraceVariables
// --------------------------------------------------------------------------

func TestRenderAudit_TraceVariables_simple(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString("Hello, {{ name }}!")
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"name": "Alice"},
		liquid.AuditOptions{TraceVariables: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if result.Output != "Hello, Alice!" {
		t.Errorf("output = %q, want %q", result.Output, "Hello, Alice!")
	}
	if len(result.Expressions) != 1 {
		t.Fatalf("len(Expressions) = %d, want 1", len(result.Expressions))
	}
	e := result.Expressions[0]
	if e.Kind != liquid.KindVariable {
		t.Errorf("Kind = %q, want %q", e.Kind, liquid.KindVariable)
	}
	if e.Variable == nil {
		t.Fatal("Variable is nil")
	}
	if e.Variable.Name != "name" {
		t.Errorf("Variable.Name = %q, want %q", e.Variable.Name, "name")
	}
	if e.Variable.Value != "Alice" {
		t.Errorf("Variable.Value = %v, want %q", e.Variable.Value, "Alice")
	}
}

func TestRenderAudit_TraceVariables_noTrace(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString("Hello, {{ name }}!")
	if err != nil {
		t.Fatal(err)
	}

	// TraceVariables not set → Expressions should be empty.
	result, ae := tpl.RenderAudit(
		liquid.Bindings{"name": "Bob"},
		liquid.AuditOptions{},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if result.Output != "Hello, Bob!" {
		t.Errorf("output = %q, want %q", result.Output, "Hello, Bob!")
	}
	if len(result.Expressions) != 0 {
		t.Errorf("len(Expressions) = %d, want 0", len(result.Expressions))
	}
}

func TestRenderAudit_TraceVariables_filterPipeline(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{{ name | upcase }}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"name": "alice"},
		liquid.AuditOptions{TraceVariables: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if result.Output != "ALICE" {
		t.Errorf("output = %q, want %q", result.Output, "ALICE")
	}
	if len(result.Expressions) == 0 {
		t.Fatal("Expressions is empty")
	}
	e := result.Expressions[0]
	if e.Variable == nil {
		t.Fatal("Variable is nil")
	}
	if len(e.Variable.Pipeline) == 0 {
		t.Fatal("Pipeline is empty, expected at least one filter step")
	}
	step := e.Variable.Pipeline[0]
	if step.Filter != "upcase" {
		t.Errorf("Pipeline[0].Filter = %q, want %q", step.Filter, "upcase")
	}
	if step.Input != "alice" {
		t.Errorf("Pipeline[0].Input = %v, want %q", step.Input, "alice")
	}
	if step.Output != "ALICE" {
		t.Errorf("Pipeline[0].Output = %v, want %q", step.Output, "ALICE")
	}
}

func TestRenderAudit_TraceVariables_depth(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% if true %}{{ x }}{% endif %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"x": 42},
		liquid.AuditOptions{TraceVariables: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if len(result.Expressions) == 0 {
		t.Fatal("Expressions is empty")
	}
	e := result.Expressions[0]
	if e.Depth != 1 {
		t.Errorf("Depth = %d, want 1 (inside if block)", e.Depth)
	}
}

// --------------------------------------------------------------------------
// RenderAudit — TraceConditions
// --------------------------------------------------------------------------

func TestRenderAudit_TraceConditions_if_taken(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% if x %}yes{% else %}no{% endif %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"x": true},
		liquid.AuditOptions{TraceConditions: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if result.Output != "yes" {
		t.Errorf("output = %q, want %q", result.Output, "yes")
	}

	var condExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindCondition {
			condExpr = &result.Expressions[i]
			break
		}
	}
	if condExpr == nil {
		t.Fatal("no condition expression found")
	}
	if condExpr.Condition == nil {
		t.Fatal("Condition is nil")
	}
	branches := condExpr.Condition.Branches
	if len(branches) != 2 {
		t.Fatalf("len(Branches) = %d, want 2 (if + else)", len(branches))
	}
	if !branches[0].Executed {
		t.Error("branches[0].Executed should be true (if branch taken)")
	}
	if branches[1].Executed {
		t.Error("branches[1].Executed should be false (else not taken)")
	}
}

func TestRenderAudit_TraceConditions_else_taken(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% if x %}yes{% else %}no{% endif %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"x": false},
		liquid.AuditOptions{TraceConditions: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if result.Output != "no" {
		t.Errorf("output = %q, want %q", result.Output, "no")
	}

	var condExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindCondition {
			condExpr = &result.Expressions[i]
			break
		}
	}
	if condExpr == nil {
		t.Fatal("no condition expression found")
	}
	branches := condExpr.Condition.Branches
	if branches[0].Executed {
		t.Error("if branch should not be executed")
	}
	if !branches[1].Executed {
		t.Error("else branch should be executed")
	}
}

func TestRenderAudit_TraceConditions_unless(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% unless disabled %}active{% endunless %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"disabled": false},
		liquid.AuditOptions{TraceConditions: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if result.Output != "active" {
		t.Errorf("output = %q, want %q", result.Output, "active")
	}

	var condExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindCondition {
			condExpr = &result.Expressions[i]
			break
		}
	}
	if condExpr == nil {
		t.Fatal("no condition expression found")
	}
	branches := condExpr.Condition.Branches
	if len(branches) == 0 {
		t.Fatal("no branches")
	}
	if branches[0].Kind != "unless" {
		t.Errorf("branches[0].Kind = %q, want %q", branches[0].Kind, "unless")
	}
}

// --------------------------------------------------------------------------
// RenderAudit — TraceIterations
// --------------------------------------------------------------------------

func TestRenderAudit_TraceIterations_basic(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% for item in items %}{{ item }}{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"items": []string{"a", "b", "c"}},
		liquid.AuditOptions{TraceIterations: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if result.Output != "abc" {
		t.Errorf("output = %q, want %q", result.Output, "abc")
	}

	var iterExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindIteration {
			iterExpr = &result.Expressions[i]
			break
		}
	}
	if iterExpr == nil {
		t.Fatal("no iteration expression found")
	}
	it := iterExpr.Iteration
	if it == nil {
		t.Fatal("Iteration is nil")
	}
	if it.Variable != "item" {
		t.Errorf("Variable = %q, want %q", it.Variable, "item")
	}
	if it.Collection != "items" {
		t.Errorf("Collection = %q, want %q", it.Collection, "items")
	}
	if it.Length != 3 {
		t.Errorf("Length = %d, want 3", it.Length)
	}
}

// --------------------------------------------------------------------------
// RenderAudit — TraceAssignments
// --------------------------------------------------------------------------

func TestRenderAudit_TraceAssignments_assign(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% assign greeting = "Hello" %}{{ greeting }}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{TraceAssignments: true, TraceVariables: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if result.Output != "Hello" {
		t.Errorf("output = %q, want %q", result.Output, "Hello")
	}

	var assignExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindAssignment {
			assignExpr = &result.Expressions[i]
			break
		}
	}
	if assignExpr == nil {
		t.Fatal("no assignment expression found")
	}
	a := assignExpr.Assignment
	if a == nil {
		t.Fatal("Assignment is nil")
	}
	if a.Variable != "greeting" {
		t.Errorf("Variable = %q, want %q", a.Variable, "greeting")
	}
	if a.Value != "Hello" {
		t.Errorf("Value = %v, want %q", a.Value, "Hello")
	}
}

func TestRenderAudit_TraceAssignments_capture(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% capture msg %}Hi there!{% endcapture %}{{ msg }}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{TraceAssignments: true, TraceVariables: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if result.Output != "Hi there!" {
		t.Errorf("output = %q, want %q", result.Output, "Hi there!")
	}

	var capExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindCapture {
			capExpr = &result.Expressions[i]
			break
		}
	}
	if capExpr == nil {
		t.Fatal("no capture expression found")
	}
	c := capExpr.Capture
	if c == nil {
		t.Fatal("Capture is nil")
	}
	if c.Variable != "msg" {
		t.Errorf("Variable = %q, want %q", c.Variable, "msg")
	}
	if c.Value != "Hi there!" {
		t.Errorf("Value = %q, want %q", c.Value, "Hi there!")
	}
}

// --------------------------------------------------------------------------
// RenderAudit — combined trace
// --------------------------------------------------------------------------

func TestRenderAudit_Combined(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% assign total = price | times: 2 %}{{ total }}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"price": 10},
		liquid.AuditOptions{TraceAssignments: true, TraceVariables: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if result.Output != "20" {
		t.Errorf("output = %q, want %q", result.Output, "20")
	}

	// Should have 2 expressions: an assignment and a variable trace.
	if len(result.Expressions) != 2 {
		t.Fatalf("len(Expressions) = %d, want 2", len(result.Expressions))
	}
	kinds := make(map[liquid.ExpressionKind]bool)
	for _, e := range result.Expressions {
		kinds[e.Kind] = true
	}
	if !kinds[liquid.KindAssignment] {
		t.Error("missing KindAssignment expression")
	}
	if !kinds[liquid.KindVariable] {
		t.Error("missing KindVariable expression")
	}
}

// --------------------------------------------------------------------------
// RenderAudit — AuditError
// --------------------------------------------------------------------------

func TestRenderAudit_Error_strictVariables(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{{ ghost }}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{},
		liquid.WithStrictVariables(),
	)
	// Result is always returned.
	if result == nil {
		t.Fatal("result is nil")
	}
	if ae == nil {
		t.Fatal("expected AuditError, got nil")
	}
	if len(ae.Errors()) == 0 {
		t.Error("AuditError.Errors() is empty")
	}
	if ae.Error() == "" {
		t.Error("AuditError.Error() is empty")
	}
}

func TestRenderAudit_ResultNonNilOnError(t *testing.T) {
	eng := newAuditEngine()
	// Template that will fail with strict variables.
	tpl, err := eng.ParseString(`before {{ missing }} after`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{},
		liquid.WithStrictVariables(),
	)
	if result == nil {
		t.Fatal("result must never be nil")
	}
	if ae == nil {
		t.Fatal("expected AuditError")
	}
	// Output may be partial but should not panic.
	_ = result.Output
}

// --------------------------------------------------------------------------
// Validate
// --------------------------------------------------------------------------

func TestValidate_emptyIF(t *testing.T) {
	eng := newAuditEngine()
	tpl, parseErr := eng.ParseString(`{% if true %}{% endif %}`)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	result, err := tpl.Validate()
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	// Should have at least one info-level empty-block diagnostic.
	found := false
	for _, d := range result.Diagnostics {
		if d.Code == "empty-block" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected empty-block diagnostic, got none")
	}
}

func TestValidate_nonEmpty(t *testing.T) {
	eng := newAuditEngine()
	tpl, parseErr := eng.ParseString(`{% if true %}hello{% endif %}`)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	result, err := tpl.Validate()
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}
	// No diagnostics expected for a non-empty block.
	for _, d := range result.Diagnostics {
		if d.Code == "empty-block" {
			t.Errorf("unexpected empty-block diagnostic: %s", d.Message)
		}
	}
}

// --------------------------------------------------------------------------
// Position / Range
// --------------------------------------------------------------------------

func TestRenderAudit_Position_lineNumber(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString("line1\nline2\n{{ x }}")
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"x": 1},
		liquid.AuditOptions{TraceVariables: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if len(result.Expressions) == 0 {
		t.Fatal("no expressions")
	}

	pos := result.Expressions[0].Range.Start
	if pos.Line != 3 {
		t.Errorf("Start.Line = %d, want 3", pos.Line)
	}
}

// --------------------------------------------------------------------------
// Gap-fix tests: assign source location and filter pipeline
// --------------------------------------------------------------------------

func TestRenderAudit_AssignSourceLoc(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% assign x = "hello" %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{TraceAssignments: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if len(result.Expressions) == 0 {
		t.Fatal("no expressions")
	}
	e := result.Expressions[0]
	if e.Kind != liquid.KindAssignment {
		t.Fatalf("Kind = %q, want assignment", e.Kind)
	}
	// The range should have a real line number (not 0).
	if e.Range.Start.Line == 0 {
		t.Errorf("Range.Start.Line = 0, want a real line number (≥1)")
	}
}

func TestRenderAudit_AssignFilterPipeline(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% assign upper = name | upcase %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"name": "alice"},
		liquid.AuditOptions{TraceAssignments: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	if len(result.Expressions) == 0 {
		t.Fatal("no expressions")
	}
	a := result.Expressions[0].Assignment
	if a == nil {
		t.Fatal("Assignment is nil")
	}
	if len(a.Pipeline) == 0 {
		t.Fatal("Pipeline is empty — filter steps not captured for assign")
	}
	if a.Pipeline[0].Filter != "upcase" {
		t.Errorf("Pipeline[0].Filter = %q, want %q", a.Pipeline[0].Filter, "upcase")
	}
	if a.Pipeline[0].Input != "alice" {
		t.Errorf("Pipeline[0].Input = %v, want %q", a.Pipeline[0].Input, "alice")
	}
	if a.Pipeline[0].Output != "ALICE" {
		t.Errorf("Pipeline[0].Output = %v, want %q", a.Pipeline[0].Output, "ALICE")
	}
	if a.Value != "ALICE" {
		t.Errorf("Value = %v, want %q", a.Value, "ALICE")
	}
}

// --------------------------------------------------------------------------
// Gap-fix tests: MaxIterationTraceItems and TracedCount
// --------------------------------------------------------------------------

func TestRenderAudit_MaxIterItems_TracedCount(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% for item in items %}{{ item }}{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}

	items := []int{1, 2, 3, 4, 5}
	result, ae := tpl.RenderAudit(
		liquid.Bindings{"items": items},
		liquid.AuditOptions{
			TraceIterations:        true,
			TraceVariables:         true,
			MaxIterationTraceItems: 2,
		},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	// Output should still be complete (MaxIterItems only limits tracing, not rendering).
	if result.Output != "12345" {
		t.Errorf("output = %q, want %q", result.Output, "12345")
	}

	// Find the iteration expression.
	var iterExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindIteration {
			iterExpr = &result.Expressions[i]
			break
		}
	}
	if iterExpr == nil {
		t.Fatal("no iteration expression")
	}
	it := iterExpr.Iteration
	if it == nil {
		t.Fatal("Iteration is nil")
	}
	if it.Length != 5 {
		t.Errorf("Length = %d, want 5", it.Length)
	}
	if it.TracedCount != 2 {
		t.Errorf("TracedCount = %d, want 2 (limited by MaxIterationTraceItems)", it.TracedCount)
	}
	if !it.Truncated {
		t.Error("Truncated should be true")
	}

	// Only 2 variable expressions should appear (one per traced iteration).
	varCount := 0
	for _, e := range result.Expressions {
		if e.Kind == liquid.KindVariable {
			varCount++
		}
	}
	if varCount != 2 {
		t.Errorf("variable expression count = %d, want 2 (only traced iterations)", varCount)
	}
}

func TestRenderAudit_NoMaxIterItems_AllTraced(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% for item in items %}{{ item }}{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"items": []int{1, 2, 3}},
		liquid.AuditOptions{TraceIterations: true, TraceVariables: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	var iterExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindIteration {
			iterExpr = &result.Expressions[i]
			break
		}
	}
	if iterExpr == nil {
		t.Fatal("no iteration expression")
	}
	if iterExpr.Iteration.TracedCount != 3 {
		t.Errorf("TracedCount = %d, want 3", iterExpr.Iteration.TracedCount)
	}
	if iterExpr.Iteration.Truncated {
		t.Error("Truncated should be false when no limit set")
	}
}

// --------------------------------------------------------------------------
// Gap-fix tests: ConditionBranch.Comparisons
// --------------------------------------------------------------------------

func TestRenderAudit_ConditionComparisons_simple(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% if x >= 10 %}big{% else %}small{% endif %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"x": 15},
		liquid.AuditOptions{TraceConditions: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	var condExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindCondition {
			condExpr = &result.Expressions[i]
			break
		}
	}
	if condExpr == nil {
		t.Fatal("no condition expression")
	}
	branches := condExpr.Condition.Branches
	if len(branches) == 0 {
		t.Fatal("no branches")
	}

	// The if branch should have items with a leaf comparison.
	ifBranch := branches[0]
	if len(ifBranch.Items) == 0 {
		t.Fatal("if branch has no items — comparison tracing not working")
	}
	cmpItem := ifBranch.Items[0].Comparison
	if cmpItem == nil {
		t.Fatal("first item is not a comparison")
	}
	if cmpItem.Operator != ">=" {
		t.Errorf("Operator = %q, want %q", cmpItem.Operator, ">=")
	}
	if cmpItem.Left != 15 {
		t.Errorf("Left = %v, want 15", cmpItem.Left)
	}
	if cmpItem.Right != 10 {
		t.Errorf("Right = %v, want 10", cmpItem.Right)
	}
	if !cmpItem.Result {
		t.Error("Result should be true (15 >= 10)")
	}
}

func TestRenderAudit_ConditionComparisons_else_noComparisons(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% if x > 100 %}big{% else %}small{% endif %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"x": 5},
		liquid.AuditOptions{TraceConditions: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	var condExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindCondition {
			condExpr = &result.Expressions[i]
			break
		}
	}
	if condExpr == nil {
		t.Fatal("no condition expression")
	}
	branches := condExpr.Condition.Branches

	// else branch should have no items.
	for _, b := range branches {
		if b.Kind == "else" && len(b.Items) > 0 {
			t.Errorf("else branch should have 0 items, got %d", len(b.Items))
		}
	}
}

func TestRenderAudit_ConditionComparisons_equality(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% if status == "active" %}yes{% endif %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"status": "active"},
		liquid.AuditOptions{TraceConditions: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	var condExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindCondition {
			condExpr = &result.Expressions[i]
			break
		}
	}
	if condExpr == nil {
		t.Fatal("no condition expression")
	}
	if len(condExpr.Condition.Branches) == 0 {
		t.Fatal("no branches")
	}
	items := condExpr.Condition.Branches[0].Items
	if len(items) == 0 {
		t.Fatal("no items for == expression")
	}
	if items[0].Comparison == nil {
		t.Fatal("first item is not a comparison")
	}
	if items[0].Comparison.Operator != "==" {
		t.Errorf("Operator = %q, want %q", items[0].Comparison.Operator, "==")
	}
}

func TestRenderAudit_ConditionComparisons_groupTrace_and(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% if x >= 10 and y < 5 %}yes{% else %}no{% endif %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"x": 15, "y": 3},
		liquid.AuditOptions{TraceConditions: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	var condExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindCondition {
			condExpr = &result.Expressions[i]
			break
		}
	}
	if condExpr == nil {
		t.Fatal("no condition expression")
	}
	branches := condExpr.Condition.Branches
	if len(branches) == 0 {
		t.Fatal("no branches")
	}
	// The if branch (index 0) should have one item: a GroupTrace for and.
	ifItems := branches[0].Items
	if len(ifItems) == 0 {
		t.Fatal("if branch has no items")
	}
	group := ifItems[0].Group
	if group == nil {
		t.Fatalf("expected a GroupTrace at items[0], got comparison %+v", ifItems[0].Comparison)
	}
	if group.Operator != "and" {
		t.Errorf("group.Operator = %q, want \"and\"", group.Operator)
	}
	if !group.Result {
		t.Error("group.Result should be true (15 >= 10 and 3 < 5)")
	}
	// The group should contain exactly two child items (the >= and < comparisons).
	if len(group.Items) != 2 {
		t.Fatalf("group.Items len = %d, want 2", len(group.Items))
	}
	// First child: the >= comparison.
	geCmp := group.Items[0].Comparison
	if geCmp == nil {
		t.Fatal("group.Items[0] should be a Comparison, got Group")
	}
	if geCmp.Operator != ">=" {
		t.Errorf("group.Items[0].Comparison.Operator = %q, want \">=\"", geCmp.Operator)
	}
	// Second child: the < comparison.
	ltCmp := group.Items[1].Comparison
	if ltCmp == nil {
		t.Fatal("group.Items[1] should be a Comparison, got Group")
	}
	if ltCmp.Operator != "<" {
		t.Errorf("group.Items[1].Comparison.Operator = %q, want \"<\"", ltCmp.Operator)
	}
}

func TestRenderAudit_Diagnostic_undefinedVariable(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{{ ghost }}`)
	if err != nil {
		t.Fatal(err)
	}

	result, _ := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{},
		liquid.WithStrictVariables(),
	)
	if result == nil {
		t.Fatal("result is nil")
	}
	var found *liquid.Diagnostic
	for i := range result.Diagnostics {
		if result.Diagnostics[i].Code == "undefined-variable" {
			found = &result.Diagnostics[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected diagnostic code \"undefined-variable\", got: %v", result.Diagnostics)
	}
	if found.Severity != liquid.SeverityWarning {
		t.Errorf("severity = %q, want %q", found.Severity, liquid.SeverityWarning)
	}
}

func TestRenderAudit_Diagnostic_argumentError(t *testing.T) {
	eng := newAuditEngine()
	// divided_by: 0 produces a ZeroDivisionError which maps to "argument-error".
	tpl, err := eng.ParseString(`{{ 10 | divided_by: 0 }}`)
	if err != nil {
		t.Fatal(err)
	}

	result, _ := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{},
	)
	if result == nil {
		t.Fatal("result is nil")
	}
	var found *liquid.Diagnostic
	for i := range result.Diagnostics {
		if result.Diagnostics[i].Code == "argument-error" {
			found = &result.Diagnostics[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected diagnostic code \"argument-error\", got: %v", result.Diagnostics)
	}
	if found.Severity != liquid.SeverityError {
		t.Errorf("severity = %q, want %q", found.Severity, liquid.SeverityError)
	}
}

func TestRenderAudit_Diagnostic_typeMismatch(t *testing.T) {
	eng := newAuditEngine()
	// Comparing a string with an integer — type mismatch.
	tpl, err := eng.ParseString(`{% if status == 1 %}yes{% endif %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, _ := tpl.RenderAudit(
		liquid.Bindings{"status": "active"},
		liquid.AuditOptions{},
	)
	if result == nil {
		t.Fatal("result is nil")
	}
	var found *liquid.Diagnostic
	for i := range result.Diagnostics {
		if result.Diagnostics[i].Code == "type-mismatch" {
			found = &result.Diagnostics[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected diagnostic code \"type-mismatch\", got: %v", result.Diagnostics)
	}
	if found.Severity != liquid.SeverityWarning {
		t.Errorf("severity = %q, want %q", found.Severity, liquid.SeverityWarning)
	}
}

func TestRenderAudit_Diagnostic_notIterable(t *testing.T) {
	eng := newAuditEngine()
	// for over a string — not-iterable.
	tpl, err := eng.ParseString(`{% for item in status %}{{ item }}{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, _ := tpl.RenderAudit(
		liquid.Bindings{"status": "pending"},
		liquid.AuditOptions{},
	)
	if result == nil {
		t.Fatal("result is nil")
	}
	var found *liquid.Diagnostic
	for i := range result.Diagnostics {
		if result.Diagnostics[i].Code == "not-iterable" {
			found = &result.Diagnostics[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected diagnostic code \"not-iterable\", got: %v", result.Diagnostics)
	}
	if found.Severity != liquid.SeverityWarning {
		t.Errorf("severity = %q, want %q", found.Severity, liquid.SeverityWarning)
	}
}

func TestRenderAudit_Diagnostic_nilDereference(t *testing.T) {
	eng := newAuditEngine()
	// customer.address.city where address is nil — nil-dereference.
	tpl, err := eng.ParseString(`{{ customer.address.city }}`)
	if err != nil {
		t.Fatal(err)
	}

	result, _ := tpl.RenderAudit(
		liquid.Bindings{"customer": map[string]any{"address": nil}},
		liquid.AuditOptions{},
	)
	if result == nil {
		t.Fatal("result is nil")
	}
	var found *liquid.Diagnostic
	for i := range result.Diagnostics {
		if result.Diagnostics[i].Code == "nil-dereference" {
			found = &result.Diagnostics[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected diagnostic code \"nil-dereference\", got: %v", result.Diagnostics)
	}
	if found.Severity != liquid.SeverityWarning {
		t.Errorf("severity = %q, want %q", found.Severity, liquid.SeverityWarning)
	}
}

func TestRenderAudit_ConditionComparisons_expressionField(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% if x >= 10 %}big{% endif %}`)
	if err != nil {
		t.Fatal(err)
	}

	result, ae := tpl.RenderAudit(
		liquid.Bindings{"x": 15},
		liquid.AuditOptions{TraceConditions: true},
	)
	if ae != nil {
		t.Fatalf("unexpected error: %v", ae)
	}
	var condExpr *liquid.Expression
	for i := range result.Expressions {
		if result.Expressions[i].Kind == liquid.KindCondition {
			condExpr = &result.Expressions[i]
			break
		}
	}
	if condExpr == nil || len(condExpr.Condition.Branches) == 0 {
		t.Fatal("no condition expression or no branches")
	}
	items := condExpr.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison item in if branch")
	}
	expr := items[0].Comparison.Expression
	if expr == "" {
		t.Error("ComparisonTrace.Expression should be non-empty for a simple comparison branch")
	}
}

func TestRenderAudit_Diagnostic_typeMismatch_hasRange(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% if status == 1 %}yes{% endif %}`)
	if err != nil {
		t.Fatal(err)
	}
	result, _ := tpl.RenderAudit(liquid.Bindings{"status": "active"}, liquid.AuditOptions{})
	var found *liquid.Diagnostic
	for i := range result.Diagnostics {
		if result.Diagnostics[i].Code == "type-mismatch" {
			found = &result.Diagnostics[i]
			break
		}
	}
	if found == nil {
		t.Fatal("expected type-mismatch diagnostic")
	}
	if found.Range.Start.Line == 0 {
		t.Error("type-mismatch diagnostic Range.Start.Line should be non-zero")
	}
	if found.Source == "" {
		t.Error("type-mismatch diagnostic Source should be non-empty")
	}
}

func TestRenderAudit_Diagnostic_nilDereference_hasRange(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{{ customer.address.city }}`)
	if err != nil {
		t.Fatal(err)
	}
	result, _ := tpl.RenderAudit(liquid.Bindings{"customer": map[string]any{"address": nil}}, liquid.AuditOptions{})
	var found *liquid.Diagnostic
	for i := range result.Diagnostics {
		if result.Diagnostics[i].Code == "nil-dereference" {
			found = &result.Diagnostics[i]
			break
		}
	}
	if found == nil {
		t.Fatal("expected nil-dereference diagnostic")
	}
	if found.Range.Start.Line == 0 {
		t.Error("nil-dereference diagnostic Range.Start.Line should be non-zero")
	}
	if found.Source == "" {
		t.Error("nil-dereference diagnostic Source should be non-empty")
	}
}

func TestRenderAudit_Diagnostic_notIterable_hasRange(t *testing.T) {
	eng := newAuditEngine()
	tpl, err := eng.ParseString(`{% for item in order %}{{ item }}{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}
	result, _ := tpl.RenderAudit(liquid.Bindings{"order": 42}, liquid.AuditOptions{})
	var found *liquid.Diagnostic
	for i := range result.Diagnostics {
		if result.Diagnostics[i].Code == "not-iterable" {
			found = &result.Diagnostics[i]
			break
		}
	}
	if found == nil {
		t.Fatal("expected not-iterable diagnostic")
	}
	if found.Range.Start == found.Range.End {
		t.Error("not-iterable diagnostic Range should be a span (Start != End)")
	}
}

func TestValidate_UndefinedFilter(t *testing.T) {
	eng := liquid.NewEngine()
	tpl, err := eng.ParseString(`{{ product.price | no_such_filter }}`)
	if err != nil {
		t.Fatal(err)
	}
	result, valErr := tpl.Validate()
	if valErr != nil {
		t.Fatal(valErr)
	}
	var found *liquid.Diagnostic
	for i := range result.Diagnostics {
		if result.Diagnostics[i].Code == "undefined-filter" {
			found = &result.Diagnostics[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected undefined-filter diagnostic, got: %v", result.Diagnostics)
	}
	if found.Severity != liquid.SeverityError {
		t.Errorf("severity = %q, want %q", found.Severity, liquid.SeverityError)
	}
}
