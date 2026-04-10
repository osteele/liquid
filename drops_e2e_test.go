package liquid_test

// drops_e2e_test.go — Intensive E2E tests for section 6: Drops.
//
// Covers every item in the implementation checklist §6:
//   6.1  ForloopDrop  (index, index0, rindex, rindex0, first, last, length,
//                      name, parentloop)
//   6.2  TablerowloopDrop  (row, col, col0, col_first, col_last + standard props)
//   6.3  EmptyDrop / BlankDrop  (literal comparisons, Go-typed bindings,
//                                assign/capture interaction, symmetric equality)
//   6.4  Drop base class  (ToLiquid, DropMethodMissing, ContextDrop)
//
// These tests are designed to prevent silent regressions — covering edge cases
// and cross-feature interactions so that any behaviour change shows up here.

import (
	"regexp"
	"testing"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/require"
)

// -- test helpers ------------------------------------------------------------

func e2eRender(t *testing.T, tpl string, bindings map[string]any) string {
	t.Helper()
	eng := liquid.NewEngine()
	out, err := eng.ParseAndRenderString(tpl, bindings)
	require.NoError(t, err, "template: %s", tpl)
	return out
}

var stripHTMLRe = regexp.MustCompile(`<[^>]+>`)

func stripHTML(s string) string { return stripHTMLRe.ReplaceAllString(s, "") }

// ============================================================================
// 6.1  ForloopDrop — standard properties
// ============================================================================

func TestE2E_ForloopDrop_Index(t *testing.T) {
	b := map[string]any{"arr": []string{"a", "b", "c", "d"}}
	require.Equal(t, "1 2 3 4 ", e2eRender(t,
		`{% for x in arr %}{{ forloop.index }} {% endfor %}`, b))
}

func TestE2E_ForloopDrop_Index0(t *testing.T) {
	b := map[string]any{"arr": []string{"a", "b", "c", "d"}}
	require.Equal(t, "0 1 2 3 ", e2eRender(t,
		`{% for x in arr %}{{ forloop.index0 }} {% endfor %}`, b))
}

func TestE2E_ForloopDrop_Rindex(t *testing.T) {
	b := map[string]any{"arr": []string{"a", "b", "c"}}
	require.Equal(t, "3 2 1 ", e2eRender(t,
		`{% for x in arr %}{{ forloop.rindex }} {% endfor %}`, b))
}

func TestE2E_ForloopDrop_Rindex0(t *testing.T) {
	b := map[string]any{"arr": []string{"a", "b", "c"}}
	require.Equal(t, "2 1 0 ", e2eRender(t,
		`{% for x in arr %}{{ forloop.rindex0 }} {% endfor %}`, b))
}

func TestE2E_ForloopDrop_First(t *testing.T) {
	b := map[string]any{"arr": []string{"a", "b", "c"}}
	require.Equal(t, "true false false ", e2eRender(t,
		`{% for x in arr %}{{ forloop.first }} {% endfor %}`, b))
}

func TestE2E_ForloopDrop_Last(t *testing.T) {
	b := map[string]any{"arr": []string{"a", "b", "c"}}
	require.Equal(t, "false false true ", e2eRender(t,
		`{% for x in arr %}{{ forloop.last }} {% endfor %}`, b))
}

func TestE2E_ForloopDrop_Length(t *testing.T) {
	b := map[string]any{"arr": []string{"a", "b", "c", "d"}}
	require.Equal(t, "4 4 4 4 ", e2eRender(t,
		`{% for x in arr %}{{ forloop.length }} {% endfor %}`, b))
}

func TestE2E_ForloopDrop_SingleElement(t *testing.T) {
	// Single-element array: first=true AND last=true simultaneously.
	require.Equal(t, "true/true/1/1", e2eRender(t,
		`{% for x in arr %}{{forloop.first}}/{{forloop.last}}/{{forloop.index}}/{{forloop.length}}{% endfor %}`,
		map[string]any{"arr": []int{42}}))
}

func TestE2E_ForloopDrop_Length_RespectsLimit(t *testing.T) {
	// forloop.length reflects the capped iteration count, not the full array size.
	b := map[string]any{"arr": []int{1, 2, 3, 4, 5}}
	require.Equal(t, "2.2.", e2eRender(t,
		`{% for x in arr limit:2 %}{{forloop.length}}.{% endfor %}`, b))
}

func TestE2E_ForloopDrop_Length_RespectsOffsetAndLimit(t *testing.T) {
	b := map[string]any{"arr": []int{10, 20, 30, 40, 50}}
	require.Equal(t, "1/3 2/3 3/3 ", e2eRender(t,
		`{% for x in arr offset:1 limit:3 %}{{forloop.index}}/{{forloop.length}} {% endfor %}`, b))
}

func TestE2E_ForloopDrop_Reversed(t *testing.T) {
	// Reversed iteration: values in reverse order; index still ascends 1→N.
	b := map[string]any{"arr": []int{1, 2, 3}}
	require.Equal(t, "3-1-true-false 2-2-false-false 1-3-false-true ", e2eRender(t,
		`{% for x in arr reversed %}{{x}}-{{forloop.index}}-{{forloop.first}}-{{forloop.last}} {% endfor %}`,
		b))
}

func TestE2E_ForloopDrop_Range_AllProperties(t *testing.T) {
	// Range literal (1..3): length=3, indices from 1.
	require.Equal(t, "1/0/3 2/1/3 3/2/3 ", e2eRender(t,
		`{% for i in (1..3) %}{{forloop.index}}/{{forloop.index0}}/{{forloop.length}} {% endfor %}`,
		nil))
}

func TestE2E_ForloopDrop_BreakLeavesCorrectState(t *testing.T) {
	// After break: only iterations up to (not including) break point ran.
	b := map[string]any{"arr": []int{10, 20, 30, 40, 50}}
	require.Equal(t, "1.2.", e2eRender(t,
		`{% for x in arr %}{% if forloop.index == 3 %}{% break %}{% endif %}{{forloop.index}}.{% endfor %}`,
		b))
}

func TestE2E_ForloopDrop_ContinueSkipsBody(t *testing.T) {
	// continue skips the remainder of the loop body for that iteration.
	b := map[string]any{"arr": []string{"a", "b", "c", "d", "e"}}
	require.Equal(t, "a.c.e.", e2eRender(t,
		`{% for x in arr %}{% if forloop.index0 == 1 or forloop.index0 == 3 %}{% continue %}{% endif %}{{x}}.{% endfor %}`,
		b))
}

func TestE2E_ForloopDrop_InsideCapture(t *testing.T) {
	b := map[string]any{"arr": []string{"a", "b", "c"}}
	require.Equal(t, "123", e2eRender(t,
		`{% capture s %}{% for x in arr %}{{forloop.index}}{% endfor %}{% endcapture %}{{s}}`, b))
}

func TestE2E_ForloopDrop_IndexAssignPersistedAfterLoop(t *testing.T) {
	// Variables assigned inside the loop survive after the loop ends.
	b := map[string]any{"arr": []string{"x", "y", "z", "w"}}
	require.Equal(t, "4", e2eRender(t,
		`{% for i in arr %}{% assign last_idx = forloop.index %}{% endfor %}{{last_idx}}`, b))
}

func TestE2E_ForloopDrop_CommaSeparatedListPattern(t *testing.T) {
	// Classic pattern: comma between items but not after the last one.
	b := map[string]any{"arr": []string{"alpha", "beta", "gamma"}}
	require.Equal(t, "alpha, beta, gamma", e2eRender(t,
		`{% for x in arr %}{{x}}{% unless forloop.last %}, {% endunless %}{% endfor %}`, b))
}

func TestE2E_ForloopDrop_FirstUsedForBrackets(t *testing.T) {
	b := map[string]any{"arr": []string{"a", "b", "c"}}
	require.Equal(t, "[a, b, c]", e2eRender(t,
		`{% for x in arr %}{% if forloop.first %}[{% endif %}{{x}}{% unless forloop.last %}, {% endunless %}{% if forloop.last %}]{% endif %}{% endfor %}`,
		b))
}

func TestE2E_ForloopDrop_RindexUsedForCountdown(t *testing.T) {
	b := map[string]any{"arr": []int{1, 2, 3}}
	require.Equal(t, "3 left 2 left 1 left ", e2eRender(t,
		`{% for x in arr %}{{forloop.rindex}} left {% endfor %}`, b))
}

func TestE2E_ForloopDrop_PropertiesWithNonStringArray(t *testing.T) {
	// int slice: same loop mechanics as string slice.
	b := map[string]any{"arr": []int{10, 20, 30}}
	require.Equal(t, "10:1:3 20:2:3 30:3:3 ", e2eRender(t,
		`{% for x in arr %}{{x}}:{{forloop.index}}:{{forloop.length}} {% endfor %}`, b))
}

// ============================================================================
// 6.1  ForloopDrop — forloop.name
// ============================================================================

func TestE2E_ForloopDrop_Name_SimpleArray(t *testing.T) {
	b := map[string]any{"products": []int{1}}
	require.Equal(t, "item-products", e2eRender(t,
		`{% for item in products %}{{forloop.name}}{% endfor %}`, b))
}

func TestE2E_ForloopDrop_Name_DifferentVariableName(t *testing.T) {
	b := map[string]any{"list": []int{1}}
	require.Equal(t, "p-list", e2eRender(t,
		`{% for p in list %}{{forloop.name}}{% endfor %}`, b))
}

func TestE2E_ForloopDrop_Name_Range(t *testing.T) {
	require.Equal(t, "i-(1..1)", e2eRender(t,
		`{% for i in (1..1) %}{{forloop.name}}{% endfor %}`, nil))
}

func TestE2E_ForloopDrop_Name_OuterVsInner(t *testing.T) {
	// Inner loop always sees its own forloop.name.
	b := map[string]any{"xs": []int{1}, "ys": []int{1}}
	require.Equal(t, "b-ys", e2eRender(t,
		`{% for a in xs %}{% for b in ys %}{{forloop.name}}{% endfor %}{% endfor %}`, b))
}

func TestE2E_ForloopDrop_Name_ConsistentAcrossIterations(t *testing.T) {
	// Name stays the same for every iteration of the same loop.
	b := map[string]any{"items": []string{"a", "b", "c"}}
	require.Equal(t, "item-items.item-items.item-items.", e2eRender(t,
		`{% for item in items %}{{forloop.name}}.{% endfor %}`, b))
}

// ============================================================================
// 6.1  ForloopDrop — forloop.parentloop
// ============================================================================

func TestE2E_ForloopDrop_ParentloopNilAtTopLevel(t *testing.T) {
	// Accessing parentloop properties at the top-level renders as empty string.
	b := map[string]any{"arr": []string{"a", "b", "c"}}
	require.Equal(t, "...", e2eRender(t,
		`{% for x in arr %}{{forloop.parentloop.index}}.{% endfor %}`, b))
}

func TestE2E_ForloopDrop_Parentloop_Index(t *testing.T) {
	b := map[string]any{"outer": []string{"u", "v", "w"}, "inner": []string{"x", "y"}}
	// outer has 3 elements: parent.index = 1,1, 2,2, 3,3
	require.Equal(t, "1,1,2,2,3,3,", e2eRender(t,
		`{% for o in outer %}{% for i in inner %}{{forloop.parentloop.index}},{% endfor %}{% endfor %}`, b))
}

func TestE2E_ForloopDrop_Parentloop_Index0(t *testing.T) {
	b := map[string]any{"outer": []string{"u", "v", "w"}, "inner": []string{"x", "y"}}
	require.Equal(t, "0,0,1,1,2,2,", e2eRender(t,
		`{% for o in outer %}{% for i in inner %}{{forloop.parentloop.index0}},{% endfor %}{% endfor %}`, b))
}

func TestE2E_ForloopDrop_Parentloop_Rindex(t *testing.T) {
	b := map[string]any{"outer": []string{"u", "v", "w"}, "inner": []string{"x", "y"}}
	require.Equal(t, "3,3,2,2,1,1,", e2eRender(t,
		`{% for o in outer %}{% for i in inner %}{{forloop.parentloop.rindex}},{% endfor %}{% endfor %}`, b))
}

func TestE2E_ForloopDrop_Parentloop_FirstLast(t *testing.T) {
	b := map[string]any{"outer": []string{"u", "v"}, "inner": []string{"x"}}
	require.Equal(t, "true:false,false:true,", e2eRender(t,
		`{% for o in outer %}{% for i in inner %}{{forloop.parentloop.first}}:{{forloop.parentloop.last}},{% endfor %}{% endfor %}`, b))
}

func TestE2E_ForloopDrop_Parentloop_Length(t *testing.T) {
	b := map[string]any{"outer": []string{"u", "v", "w"}, "inner": []string{"x"}}
	require.Equal(t, "3,3,3,", e2eRender(t,
		`{% for o in outer %}{% for i in inner %}{{forloop.parentloop.length}},{% endfor %}{% endfor %}`, b))
}

func TestE2E_ForloopDrop_Parentloop_Name(t *testing.T) {
	b := map[string]any{"outer": []int{1, 2}, "inner": []int{1}}
	require.Equal(t, "o-outero-outer", e2eRender(t,
		`{% for o in outer %}{% for i in inner %}{{forloop.parentloop.name}}{% endfor %}{% endfor %}`, b))
}

func TestE2E_ForloopDrop_ThreeLevelNesting(t *testing.T) {
	// 3-level nesting: access grandparent via parentloop.parentloop.
	// a=[1,2], b=[3,4], c=[5,6]: 2×2×2=8 iterations; grandparent index 1,1,1,1,2,2,2,2
	b := map[string]any{"a": []int{1, 2}, "b": []int{3, 4}, "c": []int{5, 6}}
	require.Equal(t, "11112222", e2eRender(t,
		`{% for x in a %}{% for y in b %}{% for z in c %}{{forloop.parentloop.parentloop.index}}{% endfor %}{% endfor %}{% endfor %}`,
		b))
}

func TestE2E_ForloopDrop_ParentloopUsedInCondition(t *testing.T) {
	// Only render inner body when parent is on its first iteration.
	b := map[string]any{"outer": []string{"A", "B"}, "inner": []string{"x", "y"}}
	require.Equal(t, "xy", e2eRender(t,
		`{% for o in outer %}{% for i in inner %}{% if forloop.parentloop.first %}{{i}}{% endif %}{% endfor %}{% endfor %}`,
		b))
}

func TestE2E_ForloopDrop_ParentloopIndexUsedAsOffset(t *testing.T) {
	// Use forloop.parentloop.index to build a unique key per outer iteration.
	b := map[string]any{"outer": []string{"A", "B"}, "inner": []string{"x"}}
	require.Equal(t, "A:1/x,B:2/x,", e2eRender(t,
		`{% for o in outer %}{% for i in inner %}{{o}}:{{forloop.parentloop.index}}/{{i}},{% endfor %}{% endfor %}`,
		b))
}

// ============================================================================
// 6.2  TablerowloopDrop — col/row properties
// ============================================================================

// NOTE: In this Go implementation, tablerow loop variables are exposed via
// `forloop` (not `tablerowloop`). The col/col0/col_first/col_last/row
// properties are added to the same forloop map for tablerow loops.

func TestE2E_TablerowDrop_Col_NoCols(t *testing.T) {
	// Without cols: all items are in row 1; col increments for each item.
	b := map[string]any{"items": []int{10, 20, 30}}
	require.Equal(t, "1 2 3 ", stripHTML(e2eRender(t,
		`{% tablerow i in items %}{{ forloop.col }} {% endtablerow %}`, b)))
}

func TestE2E_TablerowDrop_Col0_NoCols(t *testing.T) {
	b := map[string]any{"items": []int{10, 20, 30}}
	require.Equal(t, "0 1 2 ", stripHTML(e2eRender(t,
		`{% tablerow i in items %}{{ forloop.col0 }} {% endtablerow %}`, b)))
}

func TestE2E_TablerowDrop_ColFirst_NoCols(t *testing.T) {
	b := map[string]any{"items": []int{10, 20, 30}}
	require.Equal(t, "true false false ", stripHTML(e2eRender(t,
		`{% tablerow i in items %}{{ forloop.col_first }} {% endtablerow %}`, b)))
}

func TestE2E_TablerowDrop_ColLast_NoCols(t *testing.T) {
	b := map[string]any{"items": []int{10, 20, 30}}
	require.Equal(t, "false false true ", stripHTML(e2eRender(t,
		`{% tablerow i in items %}{{ forloop.col_last }} {% endtablerow %}`, b)))
}

func TestE2E_TablerowDrop_Row_NoCols(t *testing.T) {
	// No cols: all items in row 1.
	b := map[string]any{"items": []int{10, 20, 30}}
	require.Equal(t, "1 1 1 ", stripHTML(e2eRender(t,
		`{% tablerow i in items %}{{ forloop.row }} {% endtablerow %}`, b)))
}

func TestE2E_TablerowDrop_AllColProps_WithCols2(t *testing.T) {
	// 6 items, cols:2 → 3 rows; col resets after every 2 items.
	b := map[string]any{"items": []int{1, 2, 3, 4, 5, 6}}
	require.Equal(t, "1 2 1 2 1 2 ", stripHTML(e2eRender(t,
		`{% tablerow i in items cols:2 %}{{ forloop.col }} {% endtablerow %}`, b)))
	require.Equal(t, "0 1 0 1 0 1 ", stripHTML(e2eRender(t,
		`{% tablerow i in items cols:2 %}{{ forloop.col0 }} {% endtablerow %}`, b)))
	require.Equal(t, "true false true false true false ", stripHTML(e2eRender(t,
		`{% tablerow i in items cols:2 %}{{ forloop.col_first }} {% endtablerow %}`, b)))
	require.Equal(t, "false true false true false true ", stripHTML(e2eRender(t,
		`{% tablerow i in items cols:2 %}{{ forloop.col_last }} {% endtablerow %}`, b)))
	require.Equal(t, "1 1 2 2 3 3 ", stripHTML(e2eRender(t,
		`{% tablerow i in items cols:2 %}{{ forloop.row }} {% endtablerow %}`, b)))
}

func TestE2E_TablerowDrop_OddItems_ColLastOnLastItem(t *testing.T) {
	// 5 items, cols:2: last item in last row is col_last even though it's alone.
	b := map[string]any{"items": []int{1, 2, 3, 4, 5}}
	require.Equal(t, "1.1.false 1.2.true 2.1.false 2.2.true 3.1.true ", stripHTML(e2eRender(t,
		`{% tablerow i in items cols:2 %}{{forloop.row}}.{{forloop.col}}.{{forloop.col_last}} {% endtablerow %}`, b)))
}

func TestE2E_TablerowDrop_SingleItem_AllPropertiesTrueOrMinimal(t *testing.T) {
	// Single item: first=true, last=true, col_first=true, col_last=true
	require.Equal(t, "true/true/true/true", stripHTML(e2eRender(t,
		`{% tablerow i in items %}{{forloop.col_first}}/{{forloop.col_last}}/{{forloop.first}}/{{forloop.last}}{% endtablerow %}`,
		map[string]any{"items": []int{99}})))
}

func TestE2E_TablerowDrop_ColsLargerThanItems(t *testing.T) {
	// cols > len(items): all items in row 1.
	b := map[string]any{"items": []int{1, 2, 3}}
	require.Equal(t, "1.1 1.2 1.3 ", stripHTML(e2eRender(t,
		`{% tablerow i in items cols:10 %}{{forloop.row}}.{{forloop.col}} {% endtablerow %}`, b)))
}

func TestE2E_TablerowDrop_StandardProps_IndexLength(t *testing.T) {
	// Standard iteration properties still work in tablerow.
	b := map[string]any{"items": []int{10, 20, 30, 40}}
	require.Equal(t, "1 2 3 4 ", stripHTML(e2eRender(t,
		`{% tablerow i in items %}{{ forloop.index }} {% endtablerow %}`, b)))
	require.Equal(t, "4 4 4 4 ", stripHTML(e2eRender(t,
		`{% tablerow i in items %}{{ forloop.length }} {% endtablerow %}`, b)))
	require.Equal(t, "4 3 2 1 ", stripHTML(e2eRender(t,
		`{% tablerow i in items %}{{ forloop.rindex }} {% endtablerow %}`, b)))
}

func TestE2E_TablerowDrop_WithLimitAndOffset(t *testing.T) {
	// offset:2 limit:3 → items [3,4,5] from [1,2,3,4,5,6]; cols:2 → row1: 3,4; row2: 5
	b := map[string]any{"items": []int{1, 2, 3, 4, 5, 6}}
	require.Equal(t, "3-1.1 4-1.2 5-2.1 ", stripHTML(e2eRender(t,
		`{% tablerow i in items cols:2 offset:2 limit:3 %}{{i}}-{{forloop.row}}.{{forloop.col}} {% endtablerow %}`, b)))
}

func TestE2E_TablerowDrop_Range(t *testing.T) {
	// (1..4) cols:2 → 2 full rows.
	require.Equal(t, "1:false 2:true 3:false 4:true ", stripHTML(e2eRender(t,
		`{% tablerow i in (1..4) cols:2 %}{{i}}:{{forloop.col_last}} {% endtablerow %}`, nil)))
}

func TestE2E_TablerowDrop_ColLastUsedForLineBreakLogic(t *testing.T) {
	// col_last can drive row-boundary logic inside the template.
	b := map[string]any{"items": []int{1, 2, 3, 4}}
	raw := e2eRender(t,
		`{% tablerow i in items cols:2 %}{{i}}{% if forloop.col_last %}|{% endif %}{% endtablerow %}`, b)
	// After stripping HTML tags: "1 2|3 4|" with col_last after items 2 and 4.
	require.Equal(t, "12|34|", stripHTML(raw))
}

func TestE2E_TablerowDrop_WithReversed(t *testing.T) {
	// reversed applies to the values; row/col still advance forward.
	b := map[string]any{"items": []int{1, 2, 3, 4}}
	require.Equal(t, "4 3 2 1 ", stripHTML(e2eRender(t,
		`{% tablerow i in items cols:2 reversed %}{{i}} {% endtablerow %}`, b)))
}

// ============================================================================
// 6.3  EmptyDrop — literal comparisons with Go bindings
// ============================================================================

func TestE2E_EmptyDrop_RendersAsEmptyString(t *testing.T) {
	require.Equal(t, "", e2eRender(t, `{{empty}}`, nil))
}

func TestE2E_EmptyDrop_EmptyStringIsEmpty(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": ""}))
}

func TestE2E_EmptyDrop_EmptySliceIsEmpty(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": []int{}}))
}

func TestE2E_EmptyDrop_EmptyMapIsEmpty(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": map[string]any{}}))
}

func TestE2E_EmptyDrop_NilIsNotEmpty(t *testing.T) {
	require.Equal(t, "no", e2eRender(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": nil}))
}

func TestE2E_EmptyDrop_FalseIsNotEmpty(t *testing.T) {
	require.Equal(t, "no", e2eRender(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": false}))
}

func TestE2E_EmptyDrop_ZeroIsNotEmpty(t *testing.T) {
	require.Equal(t, "no", e2eRender(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": 0}))
}

func TestE2E_EmptyDrop_WhitespaceStringIsNotEmpty(t *testing.T) {
	require.Equal(t, "no", e2eRender(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "  "}))
}

func TestE2E_EmptyDrop_NonEmptyStringIsNotEmpty(t *testing.T) {
	require.Equal(t, "no", e2eRender(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "hello"}))
}

func TestE2E_EmptyDrop_NonEmptySliceIsNotEmpty(t *testing.T) {
	require.Equal(t, "no", e2eRender(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": []int{1}}))
}

func TestE2E_EmptyDrop_Symmetric_ValueOnRight(t *testing.T) {
	// empty == v should equal v == empty.
	require.Equal(t, "yes", e2eRender(t,
		`{% if empty == v %}yes{% else %}no{% endif %}`,
		map[string]any{"v": ""}))
}

func TestE2E_EmptyDrop_NotEqualToSelf(t *testing.T) {
	// empty == empty → false (special Liquid semantic).
	require.Equal(t, "false", e2eRender(t,
		`{% if empty == empty %}true{% else %}false{% endif %}`, nil))
}

func TestE2E_EmptyDrop_NotEqualToNil(t *testing.T) {
	require.Equal(t, "false", e2eRender(t,
		`{% if empty == nil %}true{% else %}false{% endif %}`, nil))
}

func TestE2E_EmptyDrop_OrderingAlwaysFalse(t *testing.T) {
	// empty has no ordering: <, >, <=, >= always false.
	require.Equal(t, "no", e2eRender(t, `{% if 1 < empty %}yes{% else %}no{% endif %}`, nil))
	require.Equal(t, "no", e2eRender(t, `{% if 1 > empty %}yes{% else %}no{% endif %}`, nil))
	require.Equal(t, "no", e2eRender(t, `{% if 1 <= empty %}yes{% else %}no{% endif %}`, nil))
	require.Equal(t, "no", e2eRender(t, `{% if 1 >= empty %}yes{% else %}no{% endif %}`, nil))
}

func TestE2E_EmptyDrop_InUnless(t *testing.T) {
	require.Equal(t, "has content", e2eRender(t,
		`{% unless v == empty %}has content{% endunless %}`,
		map[string]any{"v": "hello"}))
	require.Equal(t, "", e2eRender(t,
		`{% unless v == empty %}has content{% endunless %}`,
		map[string]any{"v": ""}))
}

func TestE2E_EmptyDrop_AfterAssign_EmptyString(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% assign x = "" %}{% if x == empty %}yes{% else %}no{% endif %}`, nil))
}

func TestE2E_EmptyDrop_AfterAssign_NonEmptyString(t *testing.T) {
	require.Equal(t, "no", e2eRender(t,
		`{% assign x = "hi" %}{% if x == empty %}yes{% else %}no{% endif %}`, nil))
}

func TestE2E_EmptyDrop_AfterCapture_EmptyBody(t *testing.T) {
	// Captured empty body → empty string → equals empty.
	require.Equal(t, "yes", e2eRender(t,
		`{% capture x %}{% endcapture %}{% if x == empty %}yes{% else %}no{% endif %}`, nil))
}

func TestE2E_EmptyDrop_AfterCapture_NonEmpty(t *testing.T) {
	require.Equal(t, "no", e2eRender(t,
		`{% capture x %}hello{% endcapture %}{% if x == empty %}yes{% else %}no{% endif %}`, nil))
}

func TestE2E_EmptyDrop_InCaseWhen(t *testing.T) {
	require.Equal(t, "is empty", e2eRender(t,
		`{% case v %}{% when empty %}is empty{% else %}not empty{% endcase %}`,
		map[string]any{"v": ""}))
	require.Equal(t, "not empty", e2eRender(t,
		`{% case v %}{% when empty %}is empty{% else %}not empty{% endcase %}`,
		map[string]any{"v": "x"}))
}

// ============================================================================
// 6.3  BlankDrop — literal comparisons with Go bindings
// ============================================================================

func TestE2E_BlankDrop_RendersAsEmptyString(t *testing.T) {
	require.Equal(t, "", e2eRender(t, `{{blank}}`, nil))
}

func TestE2E_BlankDrop_NilIsBlank(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": nil}))
}

func TestE2E_BlankDrop_FalseIsBlank(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": false}))
}

func TestE2E_BlankDrop_EmptyStringIsBlank(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": ""}))
}

func TestE2E_BlankDrop_WhitespaceStringIsBlank(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "  "}))
}

func TestE2E_BlankDrop_TabOnlyStringIsBlank(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "\t"}))
}

func TestE2E_BlankDrop_NewlineOnlyStringIsBlank(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "\n"}))
}

func TestE2E_BlankDrop_EmptySliceIsBlank(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": []int{}}))
}

func TestE2E_BlankDrop_EmptyMapIsBlank(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": map[string]any{}}))
}

func TestE2E_BlankDrop_ZeroIsNotBlank(t *testing.T) {
	// 0 is NOT blank — only nil, false, empty/whitespace strings, empty collections.
	require.Equal(t, "no", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": 0}))
}

func TestE2E_BlankDrop_TrueIsNotBlank(t *testing.T) {
	require.Equal(t, "no", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": true}))
}

func TestE2E_BlankDrop_NonEmptyStringIsNotBlank(t *testing.T) {
	require.Equal(t, "no", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "hello"}))
}

func TestE2E_BlankDrop_NonEmptySliceIsNotBlank(t *testing.T) {
	require.Equal(t, "no", e2eRender(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": []int{1}}))
}

func TestE2E_BlankDrop_Symmetric_ValueOnRight(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if blank == v %}yes{% else %}no{% endif %}`,
		map[string]any{"v": ""}))
}

func TestE2E_BlankDrop_NilSymmetric(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if blank == v %}yes{% else %}no{% endif %}`,
		map[string]any{"v": nil}))
}

func TestE2E_BlankDrop_NotEqualToSelf(t *testing.T) {
	require.Equal(t, "false", e2eRender(t,
		`{% if blank == blank %}true{% else %}false{% endif %}`, nil))
}

func TestE2E_BlankDrop_InUnless(t *testing.T) {
	require.Equal(t, "has content", e2eRender(t,
		`{% unless v == blank %}has content{% endunless %}`,
		map[string]any{"v": "hello"}))
}

func TestE2E_BlankDrop_InCaseWhen(t *testing.T) {
	require.Equal(t, "is blank", e2eRender(t,
		`{% case v %}{% when blank %}is blank{% else %}not blank{% endcase %}`,
		map[string]any{"v": "  "}))
	require.Equal(t, "not blank", e2eRender(t,
		`{% case v %}{% when blank %}is blank{% else %}not blank{% endcase %}`,
		map[string]any{"v": "x"}))
}

func TestE2E_BlankDrop_AfterCapture_WhitespaceBody(t *testing.T) {
	// Captured whitespace-only body → blank.
	require.Equal(t, "yes", e2eRender(t,
		`{% capture x %}   {% endcapture %}{% if x == blank %}yes{% else %}no{% endif %}`, nil))
}

// ============================================================================
// 6.3  EmptyDrop vs BlankDrop — cross-comparisons
// ============================================================================

func TestE2E_EmptyBlank_NotEqualToEachOther(t *testing.T) {
	// empty != blank and blank != empty (they are distinct singletons).
	require.Equal(t, "not equal", e2eRender(t,
		`{% if empty == blank %}equal{% else %}not equal{% endif %}`, nil))
	require.Equal(t, "not equal", e2eRender(t,
		`{% if blank == empty %}equal{% else %}not equal{% endif %}`, nil))
}

func TestE2E_EmptyBlank_EmptyStringIsBothEmptyAndBlank(t *testing.T) {
	// "" satisfies both empty and blank.
	require.Equal(t, "EB", e2eRender(t,
		`{% if v == empty %}E{% endif %}{% if v == blank %}B{% endif %}`,
		map[string]any{"v": ""}))
}

func TestE2E_EmptyBlank_WhitespaceIsBlankNotEmpty(t *testing.T) {
	// "  " is blank but NOT empty (blank is strictly more permissive).
	require.Equal(t, "not-E/B", e2eRender(t,
		`{% if v == empty %}E{% else %}not-E{% endif %}/{% if v == blank %}B{% else %}not-B{% endif %}`,
		map[string]any{"v": "  "}))
}

func TestE2E_EmptyBlank_NilIsBlankNotEmpty(t *testing.T) {
	// nil is blank but NOT empty.
	require.Equal(t, "not-E/B", e2eRender(t,
		`{% if v == empty %}E{% else %}not-E{% endif %}/{% if v == blank %}B{% else %}not-B{% endif %}`,
		map[string]any{"v": nil}))
}

func TestE2E_EmptyBlank_FalseIsBlankNotEmpty(t *testing.T) {
	// false is blank but NOT empty.
	require.Equal(t, "not-E/B", e2eRender(t,
		`{% if v == empty %}E{% else %}not-E{% endif %}/{% if v == blank %}B{% else %}not-B{% endif %}`,
		map[string]any{"v": false}))
}

// ============================================================================
// 6.4  Drop base class — ToLiquid interface
// ============================================================================

// e2eStringDrop wraps a string with ToLiquid.
type e2eStringDrop struct{ s string }

func (d e2eStringDrop) ToLiquid() any { return d.s }

// e2eMapDrop wraps a map[string]any with ToLiquid.
type e2eMapDrop struct{ data map[string]any }

func (d e2eMapDrop) ToLiquid() any { return d.data }

// e2eSliceDrop wraps a []string with ToLiquid.
type e2eSliceDrop struct{ items []string }

func (d e2eSliceDrop) ToLiquid() any { return d.items }

// e2eNestedDrop returns a map with nested fields.
type e2eNestedDrop struct{}

func (d e2eNestedDrop) ToLiquid() any {
	return map[string]any{
		"title": "Nested",
		"inner": map[string]any{"value": 42},
		"list":  []string{"x", "y", "z"},
	}
}

func TestE2E_Drop_ToLiquidString_Output(t *testing.T) {
	require.Equal(t, "hello", e2eRender(t, `{{obj}}`, map[string]any{"obj": e2eStringDrop{"hello"}}))
}

func TestE2E_Drop_ToLiquidString_InFilter(t *testing.T) {
	require.Equal(t, "HELLO", e2eRender(t, `{{obj | upcase}}`, map[string]any{"obj": e2eStringDrop{"hello"}}))
}

func TestE2E_Drop_ToLiquidString_InCondition(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if obj == "hi" %}yes{% else %}no{% endif %}`,
		map[string]any{"obj": e2eStringDrop{"hi"}}))
}

func TestE2E_Drop_ToLiquidMap_PropertyAccess(t *testing.T) {
	require.Equal(t, "widget 7", e2eRender(t,
		`{{obj.name}} {{obj.count}}`,
		map[string]any{"obj": e2eMapDrop{map[string]any{"name": "widget", "count": 7}}}))
}

func TestE2E_Drop_ToLiquidMap_NestedAccess(t *testing.T) {
	require.Equal(t, "Nested 42 xyz", e2eRender(t,
		`{{obj.title}} {{obj.inner.value}} {% for s in obj.list %}{{s}}{% endfor %}`,
		map[string]any{"obj": e2eNestedDrop{}}))
}

func TestE2E_Drop_ToLiquidMap_InCondition(t *testing.T) {
	require.Equal(t, "yes", e2eRender(t,
		`{% if obj.active %}yes{% else %}no{% endif %}`,
		map[string]any{"obj": e2eMapDrop{map[string]any{"active": true}}}))
}

func TestE2E_Drop_ToLiquidSlice_ForLoop(t *testing.T) {
	require.Equal(t, "a.b.c.", e2eRender(t,
		`{% for x in collection %}{{x}}.{% endfor %}`,
		map[string]any{"collection": e2eSliceDrop{[]string{"a", "b", "c"}}}))
}

func TestE2E_Drop_ToLiquidSlice_FirstLast(t *testing.T) {
	require.Equal(t, "x/z", e2eRender(t,
		`{{obj.first}}/{{obj.last}}`,
		map[string]any{"obj": e2eSliceDrop{[]string{"x", "y", "z"}}}))
}

func TestE2E_Drop_ToLiquidSlice_JoinFilter(t *testing.T) {
	require.Equal(t, "a-b-c", e2eRender(t,
		`{{obj | join: "-"}}`,
		map[string]any{"obj": e2eSliceDrop{[]string{"a", "b", "c"}}}))
}

func TestE2E_Drop_ToLiquidSlice_Size(t *testing.T) {
	require.Equal(t, "3", e2eRender(t,
		`{{obj.size}}`,
		map[string]any{"obj": e2eSliceDrop{[]string{"a", "b", "c"}}}))
}

func TestE2E_Drop_ToLiquidInsideAssign(t *testing.T) {
	// ToLiquid result can be assigned and used later.
	require.Equal(t, "hello world", e2eRender(t,
		`{% assign words = obj %}{{words | join: " "}}`,
		map[string]any{"obj": e2eSliceDrop{[]string{"hello", "world"}}}))
}

func TestE2E_Drop_ToLiquidInsideCapture(t *testing.T) {
	require.Equal(t, "result: hello", e2eRender(t,
		`{% capture r %}result: {{obj}}{% endcapture %}{{r}}`,
		map[string]any{"obj": e2eStringDrop{"hello"}}))
}

func TestE2E_Drop_ToLiquidMapSliceCombo(t *testing.T) {
	// Drop returns a map that contains a list; iterate the list.
	type item struct {
		Tags []string
	}
	require.Equal(t, "go.liquid.", e2eRender(t,
		`{% for t in obj.tags %}{{t}}.{% endfor %}`,
		map[string]any{"obj": e2eMapDrop{map[string]any{"tags": []string{"go", "liquid"}}}}))
}

// ============================================================================
// 6.4  DropMethodMissing
// ============================================================================

// e2eDynDrop exposes known fields normally; unknown keys go to MissingMethod.
type e2eDynDrop struct {
	Title   string
	dynamic map[string]any
}

func (d e2eDynDrop) MissingMethod(key string) any { return d.dynamic[key] }

func TestE2E_DropMethodMissing_KnownFieldNotIntercepted(t *testing.T) {
	// Struct field takes priority over MissingMethod.
	out := e2eRender(t, `{{obj.Title}}`, map[string]any{
		"obj": e2eDynDrop{Title: "Real", dynamic: map[string]any{"Title": "Shadow"}},
	})
	require.Equal(t, "Real", out)
}

func TestE2E_DropMethodMissing_UnknownFieldDispatched(t *testing.T) {
	out := e2eRender(t, `{{obj.color}} {{obj.count}}`, map[string]any{
		"obj": e2eDynDrop{dynamic: map[string]any{"color": "red", "count": 5}},
	})
	require.Equal(t, "red 5", out)
}

func TestE2E_DropMethodMissing_NilReturnIsEmpty(t *testing.T) {
	out := e2eRender(t, `{{obj.nope}}`, map[string]any{
		"obj": e2eDynDrop{dynamic: map[string]any{}},
	})
	require.Equal(t, "", out)
}

func TestE2E_DropMethodMissing_BoolValueUsableInCondition(t *testing.T) {
	out := e2eRender(t, `{% if obj.active %}yes{% else %}no{% endif %}`, map[string]any{
		"obj": e2eDynDrop{dynamic: map[string]any{"active": true}},
	})
	require.Equal(t, "yes", out)
}

func TestE2E_DropMethodMissing_FalseValueIsNotTruthy(t *testing.T) {
	out := e2eRender(t, `{% if obj.active %}yes{% else %}no{% endif %}`, map[string]any{
		"obj": e2eDynDrop{dynamic: map[string]any{"active": false}},
	})
	require.Equal(t, "no", out)
}

func TestE2E_DropMethodMissing_StringValueInFilter(t *testing.T) {
	out := e2eRender(t, `{{obj.name | upcase}}`, map[string]any{
		"obj": e2eDynDrop{dynamic: map[string]any{"name": "widget"}},
	})
	require.Equal(t, "WIDGET", out)
}

func TestE2E_DropMethodMissing_IntValue(t *testing.T) {
	out := e2eRender(t, `{{obj.score | plus: 10}}`, map[string]any{
		"obj": e2eDynDrop{dynamic: map[string]any{"score": 90}},
	})
	require.Equal(t, "100", out)
}

func TestE2E_DropMethodMissing_ArrayValueIterable(t *testing.T) {
	out := e2eRender(t, `{% for x in obj.tags %}{{x}}.{% endfor %}`, map[string]any{
		"obj": e2eDynDrop{dynamic: map[string]any{"tags": []string{"go", "liquid"}}},
	})
	require.Equal(t, "go.liquid.", out)
}

func TestE2E_DropMethodMissing_MapValuePropertyAccess(t *testing.T) {
	out := e2eRender(t, `{{obj.meta.version}}`, map[string]any{
		"obj": e2eDynDrop{dynamic: map[string]any{
			"meta": map[string]any{"version": "1.2.3"},
		}},
	})
	require.Equal(t, "1.2.3", out)
}

func TestE2E_DropMethodMissing_NestedMissingMethod(t *testing.T) {
	// MissingMethod returns another drop that also implements MissingMethod.
	out := e2eRender(t, `{{obj.inner.value}}`, map[string]any{
		"obj": e2eDynDrop{dynamic: map[string]any{
			"inner": e2eDynDrop{dynamic: map[string]any{"value": "deep"}},
		}},
	})
	require.Equal(t, "deep", out)
}

func TestE2E_DropMethodMissing_MultipleAccessesInTemplate(t *testing.T) {
	obj := e2eDynDrop{
		Title:   "Product",
		dynamic: map[string]any{"price": 9.99, "sku": "P-001", "available": true},
	}
	out := e2eRender(t,
		`{{obj.Title}}: {{obj.sku}} ${{obj.price}}{% if obj.available %} (in stock){% endif %}`,
		map[string]any{"obj": obj})
	require.Equal(t, "Product: P-001 $9.99 (in stock)", out)
}

func TestE2E_DropMethodMissing_ReturnedDropInFilterChain(t *testing.T) {
	// MissingMethod returns a slice; apply a filter to it.
	out := e2eRender(t, `{{obj.tags | join: ","}}`, map[string]any{
		"obj": e2eDynDrop{dynamic: map[string]any{"tags": []string{"a", "b", "c"}}},
	})
	require.Equal(t, "a,b,c", out)
}

func TestE2E_DropMethodMissing_AccessInsideForLoop(t *testing.T) {
	// Drop used as loop collection: each element is a map.
	rows := []map[string]any{
		{"id": 1, "name": "alpha"},
		{"id": 2, "name": "beta"},
	}
	out := e2eRender(t, `{% for r in obj.rows %}{{r.id}}:{{r.name}} {% endfor %}`, map[string]any{
		"obj": e2eDynDrop{dynamic: map[string]any{"rows": rows}},
	})
	require.Equal(t, "1:alpha 2:beta ", out)
}

// ============================================================================
// 6.4  ContextDrop
// ============================================================================

// e2eScopeDrop reads another binding from the rendering scope via ContextDrop.
type e2eScopeDrop struct {
	key string
	ctx liquid.DropRenderContext
}

func (d *e2eScopeDrop) SetContext(ctx liquid.DropRenderContext) { d.ctx = ctx }
func (d *e2eScopeDrop) Peek() any {
	if d.ctx == nil {
		return nil
	}
	return d.ctx.Get(d.key)
}

func TestE2E_ContextDrop_ReadsOtherBinding(t *testing.T) {
	out := e2eRender(t, `{{ probe.Peek }}`, map[string]any{
		"probe": &e2eScopeDrop{key: "msg"},
		"msg":   "hello",
	})
	require.Equal(t, "hello", out)
}

func TestE2E_ContextDrop_ReadsIntBinding(t *testing.T) {
	out := e2eRender(t, `{{ probe.Peek }}`, map[string]any{
		"probe": &e2eScopeDrop{key: "val"},
		"val":   42,
	})
	require.Equal(t, "42", out)
}

func TestE2E_ContextDrop_SeesAssignedVariable(t *testing.T) {
	// assign runs before the probe is evaluated, so probe sees "dynamic".
	out := e2eRender(t,
		`{% assign msg = "dynamic" %}{{ probe.Peek }}`,
		map[string]any{"probe": &e2eScopeDrop{key: "msg"}, "msg": "original"})
	require.Equal(t, "dynamic", out)
}

func TestE2E_ContextDrop_MissingKeyReturnsEmpty(t *testing.T) {
	out := e2eRender(t, `{{ probe.Peek }}`,
		map[string]any{"probe": &e2eScopeDrop{key: "nonexistent"}})
	require.Equal(t, "", out)
}

func TestE2E_ContextDrop_InsideForLoop_ReadsAssign(t *testing.T) {
	// Assign forloop.index to n inside the loop; probe reads n each iteration.
	out := e2eRender(t,
		`{% for x in arr %}{% assign n = forloop.index %}{{ probe.Peek }}.{% endfor %}`,
		map[string]any{
			"probe": &e2eScopeDrop{key: "n"},
			"arr":   []int{1, 2, 3},
		})
	require.Equal(t, "1.2.3.", out)
}

func TestE2E_ContextDrop_InsideNestedForLoop(t *testing.T) {
	// Drop reads 'idx' which is updated to forloop.index of the INNER loop.
	out := e2eRender(t,
		`{% for x in outer %}{% for y in inner %}{% assign idx = forloop.index %}{{ probe.Peek }},{% endfor %}{% endfor %}`,
		map[string]any{
			"probe": &e2eScopeDrop{key: "idx"},
			"outer": []int{1, 2},
			"inner": []int{1, 2, 3},
		})
	require.Equal(t, "1,2,3,1,2,3,", out)
}

func TestE2E_ContextDrop_MultipleDropsSameScopeDifferentKeys(t *testing.T) {
	out := e2eRender(t, `{{ a.Peek }}-{{ b.Peek }}`, map[string]any{
		"a": &e2eScopeDrop{key: "x"},
		"b": &e2eScopeDrop{key: "y"},
		"x": "hello",
		"y": "world",
	})
	require.Equal(t, "hello-world", out)
}

func TestE2E_ContextDrop_SameDropAccessedTwice(t *testing.T) {
	// Accessing the same drop twice in the same template.
	out := e2eRender(t, `{{ probe.Peek }}-{{ probe.Peek }}`, map[string]any{
		"probe": &e2eScopeDrop{key: "v"},
		"v":     "X",
	})
	require.Equal(t, "X-X", out)
}

func TestE2E_ContextDrop_ValueChangedBetweenAccesses(t *testing.T) {
	// Probe reads the binding value at the moment of access; assign changes it midway.
	out := e2eRender(t,
		`{{ probe.Peek }}-{% assign v = "Y" %}-{{ probe.Peek }}`,
		map[string]any{
			"probe": &e2eScopeDrop{key: "v"},
			"v":     "X",
		})
	require.Equal(t, "X--Y", out)
}

func TestE2E_ContextDrop_UsableInCondition(t *testing.T) {
	out := e2eRender(t,
		`{% if probe.Peek == "secret" %}found{% else %}not found{% endif %}`,
		map[string]any{"probe": &e2eScopeDrop{key: "pw"}, "pw": "secret"})
	require.Equal(t, "found", out)
}

func TestE2E_ContextDrop_UsableInFilterChain(t *testing.T) {
	out := e2eRender(t,
		`{{ probe.Peek | upcase }}`,
		map[string]any{"probe": &e2eScopeDrop{key: "word"}, "word": "hello"})
	require.Equal(t, "HELLO", out)
}

// e2eCombinedDrop implements both DropMethodMissing and ContextDrop.
type e2eCombinedDrop struct {
	ctx     liquid.DropRenderContext
	dynamic map[string]any
}

func (d *e2eCombinedDrop) SetContext(ctx liquid.DropRenderContext) { d.ctx = ctx }
func (d *e2eCombinedDrop) MissingMethod(key string) any {
	if key == "ctx_var" && d.ctx != nil {
		return d.ctx.Get("_secret")
	}
	return d.dynamic[key]
}

func TestE2E_ContextDrop_CombinedWithMissingMethod(t *testing.T) {
	// Drop delegates to ctx for one key and to dynamic map for others.
	obj := &e2eCombinedDrop{
		dynamic: map[string]any{"name": "combo"},
	}
	out := e2eRender(t, `{{obj.name}} {{obj.ctx_var}}`, map[string]any{
		"obj":     obj,
		"_secret": "revealed",
	})
	require.Equal(t, "combo revealed", out)
}

func TestE2E_ContextDrop_InjectionBeforeAnyPropertyAccess(t *testing.T) {
	// Context is always injected before any property is accessed on the drop,
	// so the first property access already has a valid ctx.
	out := e2eRender(t, `{{ probe.Peek }}`,
		map[string]any{"probe": &e2eScopeDrop{key: "val"}, "val": 99})
	require.Equal(t, "99", out)
}
