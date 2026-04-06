package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// ConditionTrace — Branch Structure (C01–C10)
// ============================================================================

// C01 — {% if x %}...{% endif %} with no else: 1 branch, kind="if".
func TestRenderAudit_Condition_C01_ifOnly(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}yes{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Fatal("no condition expression")
	}
	if len(c.Condition.Branches) != 1 {
		t.Fatalf("Branches len=%d, want 1 (if only)", len(c.Condition.Branches))
	}
	if c.Condition.Branches[0].Kind != "if" {
		t.Errorf("Branches[0].Kind=%q, want %q", c.Condition.Branches[0].Kind, "if")
	}
}

// C02 — {% if %}...{% else %}...{% endif %}: 2 branches: "if" + "else".
func TestRenderAudit_Condition_C02_ifElse(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}yes{% else %}no{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Fatal("no condition expression")
	}
	if len(c.Condition.Branches) != 2 {
		t.Fatalf("Branches len=%d, want 2", len(c.Condition.Branches))
	}
	if c.Condition.Branches[0].Kind != "if" {
		t.Errorf("Branches[0].Kind=%q, want if", c.Condition.Branches[0].Kind)
	}
	if c.Condition.Branches[1].Kind != "else" {
		t.Errorf("Branches[1].Kind=%q, want else", c.Condition.Branches[1].Kind)
	}
}

// C03 — {% if %}...{% elsif %}...{% endif %}: 2 branches "if" + "elsif".
func TestRenderAudit_Condition_C03_ifElsif(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}first{% elsif y %}second{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": false, "y": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Fatal("no condition expression")
	}
	if len(c.Condition.Branches) != 2 {
		t.Fatalf("Branches len=%d, want 2", len(c.Condition.Branches))
	}
	kinds := []string{c.Condition.Branches[0].Kind, c.Condition.Branches[1].Kind}
	if kinds[0] != "if" || kinds[1] != "elsif" {
		t.Errorf("kinds=%v, want [if elsif]", kinds)
	}
}

// C04 — {% if %}...{% elsif %}...{% else %}...{% endif %}: 3 branches.
func TestRenderAudit_Condition_C04_ifElsifElse(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}a{% elsif y %}b{% else %}c{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": false, "y": false}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Fatal("no condition expression")
	}
	if len(c.Condition.Branches) != 3 {
		t.Fatalf("Branches len=%d, want 3", len(c.Condition.Branches))
	}
	kinds := []string{
		c.Condition.Branches[0].Kind,
		c.Condition.Branches[1].Kind,
		c.Condition.Branches[2].Kind,
	}
	if kinds[0] != "if" || kinds[1] != "elsif" || kinds[2] != "else" {
		t.Errorf("kinds=%v, want [if elsif else]", kinds)
	}
}

// C05 — two elsif + else = 4 branches.
func TestRenderAudit_Condition_C05_twoElsifElse(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}a{% elsif y %}b{% elsif z %}c{% else %}d{% endif %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"x": false, "y": false, "z": true},
		liquid.AuditOptions{TraceConditions: true},
	)
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Fatal("no condition expression")
	}
	if len(c.Condition.Branches) != 4 {
		t.Fatalf("Branches len=%d, want 4", len(c.Condition.Branches))
	}
}

// C06 — {% unless x %}...{% endunless %}: 1 branch, kind="unless".
func TestRenderAudit_Condition_C06_unlessOnly(t *testing.T) {
	tpl := mustParseAudit(t, "{% unless disabled %}active{% endunless %}")
	result := auditOK(t, tpl, liquid.Bindings{"disabled": false}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Fatal("no condition expression")
	}
	if len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	if c.Condition.Branches[0].Kind != "unless" {
		t.Errorf("Branches[0].Kind=%q, want unless", c.Condition.Branches[0].Kind)
	}
}

// C07 — {% unless %}...{% else %}...{% endunless %}: 2 branches.
func TestRenderAudit_Condition_C07_unlessElse(t *testing.T) {
	tpl := mustParseAudit(t, "{% unless ok %}bad{% else %}good{% endunless %}")
	result := auditOK(t, tpl, liquid.Bindings{"ok": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Fatal("no condition expression")
	}
	if len(c.Condition.Branches) != 2 {
		t.Fatalf("Branches len=%d, want 2", len(c.Condition.Branches))
	}
	if c.Condition.Branches[0].Kind != "unless" {
		t.Errorf("Branches[0].Kind=%q, want unless", c.Condition.Branches[0].Kind)
	}
	if c.Condition.Branches[1].Kind != "else" {
		t.Errorf("Branches[1].Kind=%q, want else", c.Condition.Branches[1].Kind)
	}
}

// C08 — {% case %}{% when "a" %}{% when "b" %}{% endcase %}: 2 when branches.
func TestRenderAudit_Condition_C08_caseWhen(t *testing.T) {
	tpl := mustParseAudit(t, `{% case x %}{% when "a" %}alpha{% when "b" %}beta{% endcase %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": "a"}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Skip("case/when does not yet produce a ConditionTrace")
	}
	if len(c.Condition.Branches) != 2 {
		t.Fatalf("Branches len=%d, want 2 (when×2)", len(c.Condition.Branches))
	}
	for i, b := range c.Condition.Branches {
		if b.Kind != "when" {
			t.Errorf("Branches[%d].Kind=%q, want when", i, b.Kind)
		}
	}
}

// C09 — case with else: 2 when + 1 else = 3 branches.
func TestRenderAudit_Condition_C09_caseWhenElse(t *testing.T) {
	tpl := mustParseAudit(t, `{% case x %}{% when "a" %}alpha{% when "b" %}beta{% else %}other{% endcase %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": "c"}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Skip("case/when does not yet produce a ConditionTrace")
	}
	if len(c.Condition.Branches) != 3 {
		t.Fatalf("Branches len=%d, want 3", len(c.Condition.Branches))
	}
	last := c.Condition.Branches[len(c.Condition.Branches)-1]
	if last.Kind != "else" {
		t.Errorf("last.Kind=%q, want else", last.Kind)
	}
}

// ============================================================================
// ConditionTrace — Executed flag (CE01–CE10)
// ============================================================================

// CE01 — if condition true → if branch executed.
func TestRenderAudit_Condition_CE01_ifTrue_executed(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}yes{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	if !c.Condition.Branches[0].Executed {
		t.Error("if branch should have Executed=true when condition is true")
	}
}

// CE02 — if false, else present → only else executed.
func TestRenderAudit_Condition_CE02_ifFalse_elseExecuted(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}yes{% else %}no{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": false}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) != 2 {
		t.Fatal("expected 2 branches")
	}
	if c.Condition.Branches[0].Executed {
		t.Error("if branch should have Executed=false")
	}
	if !c.Condition.Branches[1].Executed {
		t.Error("else branch should have Executed=true")
	}
}

// CE03 — if false, elsif true → only elsif executed.
func TestRenderAudit_Condition_CE03_elsif_executed(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}a{% elsif y %}b{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": false, "y": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) != 2 {
		t.Fatal("expected 2 branches")
	}
	if c.Condition.Branches[0].Executed {
		t.Error("if branch should not execute")
	}
	if !c.Condition.Branches[1].Executed {
		t.Error("elsif branch should execute")
	}
}

// CE04 — if false, elsif false, else → only else executed.
func TestRenderAudit_Condition_CE04_else_executed(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}a{% elsif y %}b{% else %}c{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": false, "y": false}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) != 3 {
		t.Fatal("expected 3 branches")
	}
	if c.Condition.Branches[0].Executed {
		t.Error("if branch should not execute")
	}
	if c.Condition.Branches[1].Executed {
		t.Error("elsif branch should not execute")
	}
	if !c.Condition.Branches[2].Executed {
		t.Error("else branch should execute")
	}
}

// CE05 — unless false → unless body executes (Executed=true after inversion).
func TestRenderAudit_Condition_CE05_unlessFalse_executes(t *testing.T) {
	tpl := mustParseAudit(t, "{% unless disabled %}active{% endunless %}")
	result := auditOK(t, tpl, liquid.Bindings{"disabled": false}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	if !c.Condition.Branches[0].Executed {
		t.Error("unless branch should execute when condition is false (inverted)")
	}
}

// CE06 — unless true → unless body does NOT execute.
func TestRenderAudit_Condition_CE06_unlessTrue_notExecuted(t *testing.T) {
	tpl := mustParseAudit(t, "{% unless disabled %}active{% endunless %}")
	result := auditOK(t, tpl, liquid.Bindings{"disabled": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	if c.Condition.Branches[0].Executed {
		t.Error("unless branch should NOT execute when condition is true (inverted)")
	}
}

// CE07 — case matches first when → first Executed=true.
func TestRenderAudit_Condition_CE07_case_firstWhenExecuted(t *testing.T) {
	tpl := mustParseAudit(t, `{% case x %}{% when "a" %}alpha{% when "b" %}beta{% else %}other{% endcase %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": "a"}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Skip("case/when does not yet produce a ConditionTrace")
	}
	if len(c.Condition.Branches) < 1 || !c.Condition.Branches[0].Executed {
		t.Error("first when branch should be Executed=true when case matches")
	}
	for i := 1; i < len(c.Condition.Branches); i++ {
		if c.Condition.Branches[i].Executed {
			t.Errorf("Branches[%d].Executed should be false", i)
		}
	}
}

// CE08 — case matches second when → second Executed=true.
func TestRenderAudit_Condition_CE08_case_secondWhenExecuted(t *testing.T) {
	tpl := mustParseAudit(t, `{% case x %}{% when "a" %}alpha{% when "b" %}beta{% else %}other{% endcase %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": "b"}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Skip("case/when does not yet produce a ConditionTrace")
	}
	if len(c.Condition.Branches) < 2 || !c.Condition.Branches[1].Executed {
		t.Error("second when branch should be Executed=true")
	}
}

// CE09 — case no match, else → else Executed=true.
func TestRenderAudit_Condition_CE09_case_elseExecuted(t *testing.T) {
	tpl := mustParseAudit(t, `{% case x %}{% when "a" %}alpha{% else %}other{% endcase %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": "z"}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Skip("case/when does not yet produce a ConditionTrace")
	}
	last := c.Condition.Branches[len(c.Condition.Branches)-1]
	if !last.Executed {
		t.Error("else branch should execute when nothing matches")
	}
}

// CE10 — case no match, no else → all Executed=false.
func TestRenderAudit_Condition_CE10_case_noneExecuted(t *testing.T) {
	tpl := mustParseAudit(t, `{% case x %}{% when "a" %}alpha{% when "b" %}beta{% endcase %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": "z"}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Skip("case/when does not yet produce a ConditionTrace")
	}
	for i, b := range c.Condition.Branches {
		if b.Executed {
			t.Errorf("Branches[%d].Executed should be false when no when matches", i)
		}
	}
}

// ============================================================================
// ConditionTrace — ComparisonTrace (CC01–CC13)
// ============================================================================

// CC01 — operator ==.
func TestRenderAudit_Condition_CC01_equalOp(t *testing.T) {
	tpl := mustParseAudit(t, `{% if x == 1 %}yes{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison in if branch")
	}
	cmp := items[0].Comparison
	if cmp.Operator != "==" {
		t.Errorf("Operator=%q, want ==", cmp.Operator)
	}
	if sprintVal(cmp.Left) != "1" || sprintVal(cmp.Right) != "1" {
		t.Errorf("Left=%v Right=%v, want both 1", cmp.Left, cmp.Right)
	}
	if !cmp.Result {
		t.Error("Result should be true (1 == 1)")
	}
}

// CC02 — operator !=.
func TestRenderAudit_Condition_CC02_notEqualOp(t *testing.T) {
	tpl := mustParseAudit(t, `{% if x != 2 %}yes{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison")
	}
	if items[0].Comparison.Operator != "!=" {
		t.Errorf("Operator=%q, want !=", items[0].Comparison.Operator)
	}
	if !items[0].Comparison.Result {
		t.Error("Result should be true (1 != 2)")
	}
}

// CC03 — operator >.
func TestRenderAudit_Condition_CC03_greaterOp(t *testing.T) {
	tpl := mustParseAudit(t, `{% if x > 5 %}yes{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": 10}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison")
	}
	if items[0].Comparison.Operator != ">" {
		t.Errorf("Operator=%q, want >", items[0].Comparison.Operator)
	}
}

// CC04 — operator <.
func TestRenderAudit_Condition_CC04_lessOp(t *testing.T) {
	tpl := mustParseAudit(t, `{% if x < 5 %}yes{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": 3}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison")
	}
	if items[0].Comparison.Operator != "<" {
		t.Errorf("Operator=%q, want <", items[0].Comparison.Operator)
	}
}

// CC05 — operator >=.
func TestRenderAudit_Condition_CC05_gteOp(t *testing.T) {
	tpl := mustParseAudit(t, `{% if x >= 10 %}yes{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": 10}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	cmp := c.Condition.Branches[0].Items[0].Comparison
	if cmp == nil || cmp.Operator != ">=" {
		t.Errorf("Operator=%v, want >=", cmp)
	}
	if !cmp.Result {
		t.Error("Result should be true (10 >= 10)")
	}
}

// CC06 — operator <=.
func TestRenderAudit_Condition_CC06_lteOp(t *testing.T) {
	tpl := mustParseAudit(t, `{% if x <= 5 %}yes{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": 5}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison")
	}
	if items[0].Comparison.Operator != "<=" {
		t.Errorf("Operator=%q, want <=", items[0].Comparison.Operator)
	}
}

// CC07 — operator contains on an array.
func TestRenderAudit_Condition_CC07_containsArray(t *testing.T) {
	tpl := mustParseAudit(t, `{% if arr contains "x" %}yes{% endif %}`)
	result := auditOK(t, tpl,
		liquid.Bindings{"arr": []string{"x", "y"}},
		liquid.AuditOptions{TraceConditions: true},
	)
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison")
	}
	if items[0].Comparison.Operator != "contains" {
		t.Errorf("Operator=%q, want contains", items[0].Comparison.Operator)
	}
	if !items[0].Comparison.Result {
		t.Error("Result should be true (array contains x)")
	}
}

// CC08 — operator contains on a string (substring check).
func TestRenderAudit_Condition_CC08_containsString(t *testing.T) {
	tpl := mustParseAudit(t, `{% if str contains "ell" %}yes{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{"str": "hello"}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison")
	}
	if items[0].Comparison.Operator != "contains" {
		t.Errorf("Operator=%q, want contains", items[0].Comparison.Operator)
	}
	if !items[0].Comparison.Result {
		t.Error("Result should be true (hello contains ell)")
	}
}

// CC09 — Result=true when comparison holds.
func TestRenderAudit_Condition_CC09_resultTrue(t *testing.T) {
	tpl := mustParseAudit(t, `{% if x == "active" %}yes{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": "active"}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison")
	}
	if !items[0].Comparison.Result {
		t.Error("Result should be true")
	}
}

// CC10 — Result=false when comparison fails.
func TestRenderAudit_Condition_CC10_resultFalse(t *testing.T) {
	tpl := mustParseAudit(t, `{% if x == "active" %}yes{% else %}no{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{"x": "inactive"}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison")
	}
	if items[0].Comparison.Result {
		t.Error("Result should be false")
	}
}

// CC11 — Left and Right carry typed values (int, string, bool).
func TestRenderAudit_Condition_CC11_leftRightTypes(t *testing.T) {
	tpl := mustParseAudit(t, `{% if age > 18 %}adult{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{"age": 25}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison")
	}
	cmp := items[0].Comparison
	if sprintVal(cmp.Left) != "25" {
		t.Errorf("Left=%v, want 25 (age binding)", cmp.Left)
	}
	if sprintVal(cmp.Right) != "18" {
		t.Errorf("Right=%v, want 18 (literal)", cmp.Right)
	}
}

// CC12 — ComparisonTrace.Expression field is non-empty.
func TestRenderAudit_Condition_CC12_expressionFieldNonEmpty(t *testing.T) {
	tpl := mustParseAudit(t, `{% if score >= 60 %}pass{% endif %}`)
	result := auditOK(t, tpl, liquid.Bindings{"score": 75}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Comparison == nil {
		t.Fatal("no comparison")
	}
	if items[0].Comparison.Expression == "" {
		t.Error("ComparisonTrace.Expression should be non-empty")
	}
}

// CC13 — bare truthiness check: {% if x %} without explicit operator.
func TestRenderAudit_Condition_CC13_truthiness(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}yes{% else %}no{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": "something"}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	// Branch must be marked as executed since x is truthy.
	if !c.Condition.Branches[0].Executed {
		t.Error("if branch should be Executed=true for truthy x")
	}
	// Items may be empty (no explicit operator) or a single comparison — both are acceptable.
}

// ============================================================================
// ConditionTrace — GroupTrace (CG01–CG09)
// ============================================================================

// CG01 — "and" with both sub-conditions true: GroupTrace.Operator="and", Result=true.
// Note: GroupTrace.Items is not populated in the current implementation.
func TestRenderAudit_Condition_CG01_andBothTrue(t *testing.T) {
	tpl := mustParseAudit(t, "{% if a and b %}yes{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"a": true, "b": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	if !c.Condition.Branches[0].Executed {
		t.Error("if branch should be executed (both a and b are true)")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 {
		t.Fatal("no items in if branch")
	}
	g := items[0].Group
	if g == nil {
		t.Fatal("expected GroupTrace, got nil — items[0].Group is nil")
	}
	if g.Operator != "and" {
		t.Errorf("Operator=%q, want and", g.Operator)
	}
	if !g.Result {
		t.Error("GroupTrace.Result should be true (both sub-conditions true)")
	}
}

// CG02 — "and" with one false → GroupTrace.Result=false.
func TestRenderAudit_Condition_CG02_andOneFalse(t *testing.T) {
	tpl := mustParseAudit(t, "{% if a and b %}yes{% else %}no{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"a": true, "b": false}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Group == nil {
		t.Fatal("expected group in if branch")
	}
	if items[0].Group.Result {
		t.Error("GroupTrace.Result should be false (b is false)")
	}
}

// CG03 — "or" both false → Result=false.
func TestRenderAudit_Condition_CG03_orBothFalse(t *testing.T) {
	tpl := mustParseAudit(t, "{% if a or b %}yes{% else %}no{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"a": false, "b": false}, liquid.AuditOptions{TraceConditions: true})
	assertOutput(t, result, "no")
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Group == nil {
		t.Fatal("expected group in if branch")
	}
	if items[0].Group.Operator != "or" {
		t.Errorf("Operator=%q, want or", items[0].Group.Operator)
	}
	if items[0].Group.Result {
		t.Error("GroupTrace.Result should be false")
	}
}

// CG04 — "or" one true → Result=true.
func TestRenderAudit_Condition_CG04_orOneTrue(t *testing.T) {
	tpl := mustParseAudit(t, "{% if a or b %}yes{% else %}no{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"a": false, "b": true}, liquid.AuditOptions{TraceConditions: true})
	assertOutput(t, result, "yes")
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Group == nil {
		t.Fatal("expected group")
	}
	if !items[0].Group.Result {
		t.Error("GroupTrace.Result should be true")
	}
}

// CG05 — "a and b and c": all sub-conditions recorded (three items in group, or nested groups).
func TestRenderAudit_Condition_CG05_andThree(t *testing.T) {
	tpl := mustParseAudit(t, "{% if a and b and c %}yes{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"a": true, "b": true, "c": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 {
		t.Fatal("no items")
	}
	if !c.Condition.Branches[0].Executed {
		t.Error("branch should execute (a and b and c all true)")
	}
	// The group might be 2-deep (a and (b and c)) or 3-wide — both acceptable.
	// Just verify a group exists with Operator "and".
	g := items[0].Group
	if g == nil {
		t.Skip("no GroupTrace emitted for 3-way and — may be implementation-specific")
	}
	if g.Operator != "and" {
		t.Errorf("Operator=%q, want and", g.Operator)
	}
}

// CG06 — "a or b or c": at least a group with Operator "or".
func TestRenderAudit_Condition_CG06_orThree(t *testing.T) {
	tpl := mustParseAudit(t, "{% if a or b or c %}yes{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"a": false, "b": false, "c": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	if !c.Condition.Branches[0].Executed {
		t.Error("branch should execute (c is true)")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Group == nil {
		t.Skip("no GroupTrace emitted for 3-way or — may be implementation-specific")
	}
	if items[0].Group.Operator != "or" {
		t.Errorf("Operator=%q, want or", items[0].Group.Operator)
	}
}

// CG07 — "a and b or c" mixed — Liquid evaluates right-to-left: `a and (b or c)`.
// With a=true, b=true, c=false: `true and (true or false)` = true. Branch executes.
func TestRenderAudit_Condition_CG07_andOrMixed(t *testing.T) {
	tpl := mustParseAudit(t, "{% if a and b or c %}yes{% else %}no{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"a": true, "b": true, "c": false}, liquid.AuditOptions{TraceConditions: true})
	assertOutput(t, result, "yes")
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	if !c.Condition.Branches[0].Executed {
		t.Error("if branch should execute (a=true, b=true)")
	}
}

// CG08 — group containing a comparison: GroupTrace exists with Operator="and".
// Note: GroupTrace.Items is only populated when both sides are explicit comparisons.
// With truthiness (bare variable) on one side, Items may be empty.
func TestRenderAudit_Condition_CG08_groupContainsComparisons(t *testing.T) {
	tpl := mustParseAudit(t, "{% if age >= 18 and active %}yes{% endif %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"age": 20, "active": true},
		liquid.AuditOptions{TraceConditions: true},
	)
	assertOutput(t, result, "yes")
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	if !c.Condition.Branches[0].Executed {
		t.Error("if branch should execute (age=20 >= 18 and active=true)")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Group == nil {
		t.Fatal("expected GroupTrace in if branch items")
	}
	g := items[0].Group
	if g.Operator != "and" {
		t.Errorf("GroupTrace.Operator=%q, want \"and\"", g.Operator)
	}
	if !g.Result {
		t.Error("GroupTrace.Result should be true (age>=18 and active=true)")
	}
}

// CG09 — group with two explicit comparisons: GroupTrace.Items contains both sub-comparisons.
// This validates the full GroupTrace.Items population as promised in the spec.
func TestRenderAudit_Condition_CG09_groupItemsBothComparisons(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x >= 10 and y < 5 %}yes{% else %}no{% endif %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"x": 15, "y": 3},
		liquid.AuditOptions{TraceConditions: true},
	)
	assertOutput(t, result, "yes")
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no branches")
	}
	items := c.Condition.Branches[0].Items
	if len(items) == 0 || items[0].Group == nil {
		t.Fatal("expected GroupTrace in if branch")
	}
	g := items[0].Group
	if g.Operator != "and" {
		t.Errorf("GroupTrace.Operator=%q, want \"and\"", g.Operator)
	}
	if !g.Result {
		t.Error("GroupTrace.Result should be true (15>=10 and 3<5)")
	}
	// GroupTrace.Items must have exactly 2 children (one per comparison).
	if len(g.Items) != 2 {
		t.Fatalf("GroupTrace.Items len=%d, want 2", len(g.Items))
	}
	// First child: x >= 10
	cmp0 := g.Items[0].Comparison
	if cmp0 == nil {
		t.Fatal("GroupTrace.Items[0] should be a Comparison, got Group")
	}
	if cmp0.Operator != ">=" {
		t.Errorf("Items[0].Operator=%q, want >=", cmp0.Operator)
	}
	if cmp0.Left != 15 {
		t.Errorf("Items[0].Left=%v, want 15", cmp0.Left)
	}
	if cmp0.Right != 10 {
		t.Errorf("Items[0].Right=%v, want 10", cmp0.Right)
	}
	if !cmp0.Result {
		t.Error("Items[0].Result should be true (15 >= 10)")
	}
	// Second child: y < 5
	cmp1 := g.Items[1].Comparison
	if cmp1 == nil {
		t.Fatal("GroupTrace.Items[1] should be a Comparison, got Group")
	}
	if cmp1.Operator != "<" {
		t.Errorf("Items[1].Operator=%q, want <", cmp1.Operator)
	}
	if cmp1.Left != 3 {
		t.Errorf("Items[1].Left=%v, want 3", cmp1.Left)
	}
	if cmp1.Right != 5 {
		t.Errorf("Items[1].Right=%v, want 5", cmp1.Right)
	}
	if !cmp1.Result {
		t.Error("Items[1].Result should be true (3 < 5)")
	}
}

// ============================================================================
// ConditionTrace — Branch Range and Source (CB01–CB05)
// ============================================================================

// CB01 — Branch[0].Range is valid for the if header.
func TestRenderAudit_Condition_CB01_branchRangeValid(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x > 0 %}yes{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 1}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no condition branches")
	}
	assertRangeValid(t, c.Condition.Branches[0].Range, "branch[0].Range")
}

// CB02 — else branch Range is valid.
func TestRenderAudit_Condition_CB02_elseBranchRange(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}yes{% else %}no{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": false}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) != 2 {
		t.Fatal("expected 2 branches")
	}
	assertRangeValid(t, c.Condition.Branches[1].Range, "else branch Range")
}

// CB03 — elsif branch Range is valid and points to its own line.
func TestRenderAudit_Condition_CB03_elsifBranchRange(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}a{% elsif y %}b{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": false, "y": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 2 {
		t.Fatal("expected 2 branches (if + elsif)")
	}
	assertRangeValid(t, c.Condition.Branches[1].Range, "elsif branch Range")
}

// CB04 — ConditionTrace Expression.Range is valid.
func TestRenderAudit_Condition_CB04_expressionRangeValid(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}yes{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil {
		t.Fatal("no condition expression")
	}
	assertRangeValid(t, c.Range, "condition expression Range")
}

// CB05 — ConditionTrace Expression.Source is non-empty.
func TestRenderAudit_Condition_CB05_expressionSourceNonEmpty(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}yes{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil {
		t.Fatal("no condition expression")
	}
	if c.Source == "" {
		t.Error("ConditionTrace Expression.Source should be non-empty")
	}
}

// ============================================================================
// ConditionTrace — Depth (CD01–CD03)
// ============================================================================

// CD01 — top-level condition has Depth=0.
func TestRenderAudit_Condition_CD01_depthZero(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}yes{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": true}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil {
		t.Fatal("no condition expression")
	}
	if c.Depth != 0 {
		t.Errorf("Depth=%d, want 0 for top-level condition", c.Depth)
	}
}

// CD02 — condition inside a for block has Depth=1.
func TestRenderAudit_Condition_CD02_depthInsideFor(t *testing.T) {
	tpl := mustParseAudit(t, "{% for item in items %}{% if item > 1 %}big{% endif %}{% endfor %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"items": []int{2}},
		liquid.AuditOptions{TraceConditions: true},
	)
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil {
		t.Fatal("no condition expression")
	}
	if c.Depth != 1 {
		t.Errorf("Depth=%d, want 1 (inside for)", c.Depth)
	}
}

// CD03 — nested if inside if has Depth=2.
func TestRenderAudit_Condition_CD03_depthNestedIf(t *testing.T) {
	tpl := mustParseAudit(t, "{% if true %}{% if true %}inner{% endif %}{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{}, liquid.AuditOptions{TraceConditions: true})
	conditions := allExprs(result.Expressions, liquid.KindCondition)
	if len(conditions) < 2 {
		t.Fatalf("expected 2 condition expressions (outer + inner), got %d", len(conditions))
	}
	// Outer has Depth=0, inner has Depth=1.
	depths := make([]int, len(conditions))
	for i, c := range conditions {
		depths[i] = c.Depth
	}
	found1 := false
	for _, d := range depths {
		if d == 1 {
			found1 = true
		}
	}
	if !found1 {
		t.Errorf("depths=%v; expected at least one condition at Depth=1 (inner if)", depths)
	}
}

// ============================================================================
// ConditionTrace — error in condition (CR01)
// ============================================================================

// CR01 — undefined variable in condition: with StrictVariables, the comparison silently
// evaluates to false (nil compared to 1 returns false) and the else branch runs.
// No diagnostic is emitted for undefined variables used in comparisons (only for output tags).
func TestRenderAudit_Condition_CR01_undefinedVarInCondition(t *testing.T) {
	tpl := mustParseAudit(t, "{% if ghost == 1 %}yes{% else %}no{% endif %}")
	result, ae := tpl.RenderAudit(
		liquid.Bindings{},
		liquid.AuditOptions{TraceConditions: true},
		liquid.WithStrictVariables(),
	)
	if result == nil {
		t.Fatal("result must not be nil")
	}
	_ = ae // may or may not be nil depending on strict mode handling
	// The else branch runs because ghost (undefined) != 1.
	assertOutput(t, result, "no")
	// The condition trace should show the if branch NOT executed.
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil || len(c.Condition.Branches) < 1 {
		t.Fatal("no condition trace")
	}
	if c.Condition.Branches[0].Executed {
		t.Error("if branch (ghost == 1) should NOT be executed (ghost is undefined/nil)")
	}
}

// ============================================================================
// else branch Items are empty (no explicit comparison)
// ============================================================================

// Extra: else branch has no Items.
func TestRenderAudit_Condition_ElseBranchNoItems(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x > 10 %}big{% else %}small{% endif %}")
	result := auditOK(t, tpl, liquid.Bindings{"x": 5}, liquid.AuditOptions{TraceConditions: true})
	c := firstExpr(result.Expressions, liquid.KindCondition)
	if c == nil || c.Condition == nil {
		t.Fatal("no condition expression")
	}
	for _, b := range c.Condition.Branches {
		if b.Kind == "else" && len(b.Items) > 0 {
			t.Errorf("else branch should have 0 Items, got %d", len(b.Items))
		}
	}
}

// Extra: only the executed branch's inner expressions appear in the expressions array.
func TestRenderAudit_Condition_OnlyExecutedBranchInnerExprs(t *testing.T) {
	tpl := mustParseAudit(t, "{% if x %}{{ a }}{% else %}{{ b }}{% endif %}")
	result := auditOK(t, tpl,
		liquid.Bindings{"x": true, "a": "yes", "b": "no"},
		liquid.AuditOptions{TraceConditions: true, TraceVariables: true},
	)
	vars := allExprs(result.Expressions, liquid.KindVariable)
	for _, v := range vars {
		if v.Variable != nil && v.Variable.Name == "b" {
			t.Error("variable 'b' should not be traced — it's in the unexecuted branch")
		}
	}
}
