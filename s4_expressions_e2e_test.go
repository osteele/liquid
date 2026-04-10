package liquid_test

// s4_expressions_e2e_test.go — Intensive E2E tests for Section 4: Expressões / Literais
//
// Coverage matrix:
//   A. Literal output — all Go scalar types, nil, true, false, int, float, string, range
//   B. Comparison operators — ==, !=, <>, <, >, <=, >= across all type combinations
//   C. empty literal — emptiness semantics for every Go container and scalar
//   D. blank literal — blanking semantics for every Go container and scalar
//   E. Range literal — output, for-loop iteration, contains (boundary/mid/far/variable)
//   F. not operator — basic, compound, and precedence over and/or
//   G. nil/null with ordering — all four ordering operators on both sides
//   H. String escape sequences — all supported escapes in output and comparison
//   I. Logical operators — and/or right-associativity, short-circuit edge cases
//   J. Integration — templates combining multiple section 4 features
//   K. Edge cases — assigns, captures, nested loops, unless, case/when
//
// Every test function is self-contained: it creates its own engine, so test
// sharding or parallel agents cannot share state.

import (
	"testing"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func s4eng() *liquid.Engine { return liquid.NewEngine() }

func s4render(t *testing.T, tpl string, binds map[string]any) string {
	t.Helper()
	out, err := s4eng().ParseAndRenderString(tpl, binds)
	require.NoError(t, err, "template: %q", tpl)
	return out
}

func s4renderErr(t *testing.T, tpl string, binds map[string]any) (string, error) {
	t.Helper()
	return s4eng().ParseAndRenderString(tpl, binds)
}

func s4eq(t *testing.T, want, tpl string, binds map[string]any) {
	t.Helper()
	require.Equal(t, want, s4render(t, tpl, binds))
}

// ═════════════════════════════════════════════════════════════════════════════
// A. Literal Output
// ═════════════════════════════════════════════════════════════════════════════

// A1 — nil / null render as empty string
func TestS4_Literal_NilRendersEmpty(t *testing.T) {
	s4eq(t, "", `{{ nil }}`, nil)
}

func TestS4_Literal_NullRendersEmpty(t *testing.T) {
	s4eq(t, "", `{{ null }}`, nil)
}

func TestS4_Literal_GoNilBindingRendersEmpty(t *testing.T) {
	s4eq(t, "", `{{ v }}`, map[string]any{"v": nil})
}

func TestS4_Literal_UnsetVarRendersEmpty(t *testing.T) {
	s4eq(t, "", `{{ missing }}`, nil)
}

// A2 — boolean literals
func TestS4_Literal_TrueRendersTrue(t *testing.T) {
	s4eq(t, "true", `{{ true }}`, nil)
}

func TestS4_Literal_FalseRendersFalse(t *testing.T) {
	s4eq(t, "false", `{{ false }}`, nil)
}

func TestS4_Literal_GoBoolBindingTrue(t *testing.T) {
	s4eq(t, "true", `{{ v }}`, map[string]any{"v": true})
}

func TestS4_Literal_GoBoolBindingFalse(t *testing.T) {
	s4eq(t, "false", `{{ v }}`, map[string]any{"v": false})
}

// A3 — integer literals (template literals and Go bindings)
func TestS4_Literal_PositiveInt(t *testing.T) {
	s4eq(t, "42", `{{ 42 }}`, nil)
}

func TestS4_Literal_NegativeInt(t *testing.T) {
	s4eq(t, "-7", `{{ -7 }}`, nil)
}

func TestS4_Literal_Zero(t *testing.T) {
	s4eq(t, "0", `{{ 0 }}`, nil)
}

func TestS4_Literal_GoInt(t *testing.T) {
	s4eq(t, "100", `{{ v }}`, map[string]any{"v": 100})
}

func TestS4_Literal_GoInt64(t *testing.T) {
	s4eq(t, "9876543210", `{{ v }}`, map[string]any{"v": int64(9876543210)})
}

func TestS4_Literal_GoUint(t *testing.T) {
	s4eq(t, "255", `{{ v }}`, map[string]any{"v": uint(255)})
}

// A4 — float literals
func TestS4_Literal_PositiveFloat(t *testing.T) {
	s4eq(t, "2.5", `{{ 2.5 }}`, nil)
}

func TestS4_Literal_NegativeFloat(t *testing.T) {
	s4eq(t, "-17.42", `{{ -17.42 }}`, nil)
}

func TestS4_Literal_GoFloat64(t *testing.T) {
	s4eq(t, "3.14", `{{ v }}`, map[string]any{"v": 3.14})
}

// A5 — string literals
func TestS4_Literal_SingleQuotedString(t *testing.T) {
	s4eq(t, "hello", `{{ 'hello' }}`, nil)
}

func TestS4_Literal_DoubleQuotedString(t *testing.T) {
	s4eq(t, "world", `{{ "world" }}`, nil)
}

func TestS4_Literal_EmptyString(t *testing.T) {
	s4eq(t, "", `{{ "" }}`, nil)
}

func TestS4_Literal_StringWithSpaces(t *testing.T) {
	s4eq(t, "hello world", `{{ "hello world" }}`, nil)
}

func TestS4_Literal_StringWithEmoji(t *testing.T) {
	s4eq(t, "🔥", `{{ '🔥' }}`, nil)
}

func TestS4_Literal_GoStringBinding(t *testing.T) {
	s4eq(t, "bound", `{{ v }}`, map[string]any{"v": "bound"})
}

// A6 — range literals
func TestS4_Literal_RangeOutputFormat(t *testing.T) {
	// Range renders as "start..end"
	s4eq(t, "1..5", `{{ (1..5) }}`, nil)
}

func TestS4_Literal_RangeNegativeBound(t *testing.T) {
	s4eq(t, "-3..3", `{{ (-3..3) }}`, nil)
}

func TestS4_Literal_RangeSingleElement(t *testing.T) {
	s4eq(t, "4..4", `{{ (4..4) }}`, nil)
}

func TestS4_Literal_RangeWithVariableBound(t *testing.T) {
	// Range bound from variable
	s4eq(t, "1..5", `{{ (1..n) }}`, map[string]any{"n": 5})
}

func TestS4_Literal_RangeForLoopIterates(t *testing.T) {
	s4eq(t, "1-2-3-4-5", `{% for i in (1..5) %}{{ i }}{% unless forloop.last %}-{% endunless %}{% endfor %}`, nil)
}

func TestS4_Literal_RangeForLoopNegative(t *testing.T) {
	s4eq(t, "-2-1012", `{% for i in (-2..2) %}{{ i }}{% endfor %}`, nil)
}

func TestS4_Literal_RangeForLoopSingleItem(t *testing.T) {
	s4eq(t, "7", `{% for i in (7..7) %}{{ i }}{% endfor %}`, nil)
}

func TestS4_Literal_RangeForLoopWithVariableBound(t *testing.T) {
	s4eq(t, "123", `{% for i in (1..n) %}{{ i }}{% endfor %}`, map[string]any{"n": 3})
}

// ═════════════════════════════════════════════════════════════════════════════
// B. Comparison Operators ×  type combinations
// ═════════════════════════════════════════════════════════════════════════════

// B1 — == (equality)
func TestS4_Eq_IntInt(t *testing.T) {
	s4eq(t, "yes", `{% if 3 == 3 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Eq_IntFloat(t *testing.T) {
	// 3 == 3.0 should be true (numeric equality across types)
	s4eq(t, "yes", `{% if 3 == 3.0 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Eq_StringString(t *testing.T) {
	s4eq(t, "yes", `{% if "foo" == "foo" %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Eq_NilNil(t *testing.T) {
	s4eq(t, "yes", `{% if nil == nil %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Eq_NullNull(t *testing.T) {
	s4eq(t, "yes", `{% if null == null %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Eq_NilNull(t *testing.T) {
	s4eq(t, "yes", `{% if nil == null %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Eq_BindingNil(t *testing.T) {
	s4eq(t, "yes", `{% if v == nil %}yes{% else %}no{% endif %}`, map[string]any{"v": nil})
}

func TestS4_Eq_BindingString(t *testing.T) {
	s4eq(t, "yes", `{% if v == "hello" %}yes{% else %}no{% endif %}`, map[string]any{"v": "hello"})
}

func TestS4_Eq_BoolTrue(t *testing.T) {
	s4eq(t, "yes", `{% if true == true %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Eq_BoolFalseTrue(t *testing.T) {
	s4eq(t, "no", `{% if false == true %}yes{% else %}no{% endif %}`, nil)
}

// B2 — != (inequality)
func TestS4_Neq_IntDifferent(t *testing.T) {
	s4eq(t, "yes", `{% if 1 != 2 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Neq_IntSame(t *testing.T) {
	s4eq(t, "no", `{% if 1 != 1 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Neq_StringDifferent(t *testing.T) {
	s4eq(t, "yes", `{% if "a" != "b" %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Neq_NilNotEqualTrue(t *testing.T) {
	s4eq(t, "yes", `{% if nil != true %}yes{% else %}no{% endif %}`, nil)
}

// B3 — <> (alias for !=)
func TestS4_Diamond_IntDifferent(t *testing.T) {
	s4eq(t, "yes", `{% if 5 <> 3 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Diamond_IntSame(t *testing.T) {
	s4eq(t, "no", `{% if 5 <> 5 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Diamond_StringDifferent(t *testing.T) {
	s4eq(t, "yes", `{% if "x" <> "y" %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Diamond_StringSame(t *testing.T) {
	s4eq(t, "no", `{% if "x" <> "x" %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Diamond_FloatDifferent(t *testing.T) {
	s4eq(t, "yes", `{% if 1.5 <> 2.5 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Diamond_CrossTypeEqual(t *testing.T) {
	// 3 == 3.0 → so 3 <> 3.0 is false
	s4eq(t, "no", `{% if 3 <> 3.0 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Diamond_BindingBinding(t *testing.T) {
	s4eq(t, "yes", `{% if a <> b %}yes{% else %}no{% endif %}`,
		map[string]any{"a": "foo", "b": "bar"})
}

func TestS4_Diamond_IdenticalToNeq(t *testing.T) {
	// <> and != must produce exactly the same result
	out1 := s4render(t, `{% if v <> "x" %}1{% else %}0{% endif %}`, map[string]any{"v": "y"})
	out2 := s4render(t, `{% if v != "x" %}1{% else %}0{% endif %}`, map[string]any{"v": "y"})
	require.Equal(t, out1, out2)
}

// B4 — ordering operators
func TestS4_Lt_True(t *testing.T) {
	s4eq(t, "yes", `{% if 1 < 2 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Lt_False(t *testing.T) {
	s4eq(t, "no", `{% if 2 < 2 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Lte_True(t *testing.T) {
	s4eq(t, "yes", `{% if 2 <= 2 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Gt_True(t *testing.T) {
	s4eq(t, "yes", `{% if 3 > 2 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Gte_True(t *testing.T) {
	s4eq(t, "yes", `{% if 3 >= 3 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Compare_FloatInt(t *testing.T) {
	s4eq(t, "yes", `{% if 2.5 > 2 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Compare_StringOrder(t *testing.T) {
	s4eq(t, "yes", `{% if "b" > "a" %}yes{% else %}no{% endif %}`, nil)
}

// ═════════════════════════════════════════════════════════════════════════════
// C. empty literal
// ═════════════════════════════════════════════════════════════════════════════

// C1 — what IS empty
func TestS4_Empty_EmptyStringIsEmpty(t *testing.T) {
	s4eq(t, "yes", `{% if v == empty %}yes{% else %}no{% endif %}`, map[string]any{"v": ""})
}

func TestS4_Empty_EmptyArrayIsEmpty(t *testing.T) {
	s4eq(t, "yes", `{% if v == empty %}yes{% else %}no{% endif %}`, map[string]any{"v": []any{}})
}

func TestS4_Empty_EmptyMapIsEmpty(t *testing.T) {
	s4eq(t, "yes", `{% if v == empty %}yes{% else %}no{% endif %}`, map[string]any{"v": map[string]any{}})
}

// C2 — what is NOT empty
func TestS4_Empty_NilIsNotEmpty(t *testing.T) {
	// nil is NOT empty — empty = collection/string with zero length
	s4eq(t, "no", `{% if v == empty %}yes{% else %}no{% endif %}`, map[string]any{"v": nil})
}

func TestS4_Empty_FalseIsNotEmpty(t *testing.T) {
	s4eq(t, "no", `{% if v == empty %}yes{% else %}no{% endif %}`, map[string]any{"v": false})
}

func TestS4_Empty_ZeroIsNotEmpty(t *testing.T) {
	s4eq(t, "no", `{% if v == empty %}yes{% else %}no{% endif %}`, map[string]any{"v": 0})
}

func TestS4_Empty_WhitespaceStringIsNotEmpty(t *testing.T) {
	s4eq(t, "no", `{% if v == empty %}yes{% else %}no{% endif %}`, map[string]any{"v": "  "})
}

func TestS4_Empty_NonEmptyStringIsNotEmpty(t *testing.T) {
	s4eq(t, "no", `{% if v == empty %}yes{% else %}no{% endif %}`, map[string]any{"v": "a"})
}

func TestS4_Empty_NonEmptyArrayIsNotEmpty(t *testing.T) {
	s4eq(t, "no", `{% if v == empty %}yes{% else %}no{% endif %}`, map[string]any{"v": []any{1}})
}

func TestS4_Empty_NonEmptyMapIsNotEmpty(t *testing.T) {
	s4eq(t, "no", `{% if v == empty %}yes{% else %}no{% endif %}`, map[string]any{"v": map[string]any{"k": 1}})
}

// C3 — empty never equals itself
func TestS4_Empty_EmptyDoesNotEqualEmpty(t *testing.T) {
	// Liquid spec: empty == empty → false (it's a special asymmetric sentinel)
	s4eq(t, "no", `{% if empty == empty %}yes{% else %}no{% endif %}`, nil)
}

// C4 — empty renderers as ""
func TestS4_Empty_RendersAsEmptyString(t *testing.T) {
	s4eq(t, "", `{{ empty }}`, nil)
}

// C5 — ordering operators always return false with empty
func TestS4_Empty_OrderingAlwaysFalse(t *testing.T) {
	cases := []string{
		`{% if 1 < empty %}y{% else %}n{% endif %}`,
		`{% if 1 <= empty %}y{% else %}n{% endif %}`,
		`{% if 1 > empty %}y{% else %}n{% endif %}`,
		`{% if 1 >= empty %}y{% else %}n{% endif %}`,
		`{% if empty < 1 %}y{% else %}n{% endif %}`,
		`{% if empty <= 1 %}y{% else %}n{% endif %}`,
		`{% if empty > 1 %}y{% else %}n{% endif %}`,
		`{% if empty >= 1 %}y{% else %}n{% endif %}`,
	}
	for _, c := range cases {
		assert.Equal(t, "n", s4render(t, c, nil), "template: %s", c)
	}
}

// C6 — symmetric: v == empty and empty == v give same result
func TestS4_Empty_SymmetricComparison(t *testing.T) {
	binds := map[string]any{"v": ""}
	out1 := s4render(t, `{% if v == empty %}yes{% else %}no{% endif %}`, binds)
	out2 := s4render(t, `{% if empty == v %}yes{% else %}no{% endif %}`, binds)
	require.Equal(t, out1, out2)
}

// C7 — != empty
func TestS4_Empty_NotEqualNonEmpty(t *testing.T) {
	s4eq(t, "yes", `{% if v != empty %}yes{% else %}no{% endif %}`, map[string]any{"v": "hello"})
}

func TestS4_Empty_NotEqualOnEmptyString(t *testing.T) {
	s4eq(t, "no", `{% if v != empty %}yes{% else %}no{% endif %}`, map[string]any{"v": ""})
}

// C8 — assign + empty comparison
func TestS4_Empty_AssignedEmptyString(t *testing.T) {
	s4eq(t, "is empty", `{% assign v = "" %}{% if v == empty %}is empty{% else %}not empty{% endif %}`, nil)
}

func TestS4_Empty_AssignedNonEmpty(t *testing.T) {
	s4eq(t, "not empty", `{% assign v = "x" %}{% if v == empty %}is empty{% else %}not empty{% endif %}`, nil)
}

// ═════════════════════════════════════════════════════════════════════════════
// D. blank literal
// ═════════════════════════════════════════════════════════════════════════════

// D1 — what IS blank
func TestS4_Blank_NilIsBlank(t *testing.T) {
	s4eq(t, "yes", `{% if v == blank %}yes{% else %}no{% endif %}`, map[string]any{"v": nil})
}

func TestS4_Blank_FalseIsBlank(t *testing.T) {
	s4eq(t, "yes", `{% if v == blank %}yes{% else %}no{% endif %}`, map[string]any{"v": false})
}

func TestS4_Blank_EmptyStringIsBlank(t *testing.T) {
	s4eq(t, "yes", `{% if v == blank %}yes{% else %}no{% endif %}`, map[string]any{"v": ""})
}

func TestS4_Blank_WhitespaceStringIsBlank(t *testing.T) {
	s4eq(t, "yes", `{% if v == blank %}yes{% else %}no{% endif %}`, map[string]any{"v": "   \t\n"})
}

func TestS4_Blank_EmptyArrayIsBlank(t *testing.T) {
	s4eq(t, "yes", `{% if v == blank %}yes{% else %}no{% endif %}`, map[string]any{"v": []any{}})
}

func TestS4_Blank_EmptyMapIsBlank(t *testing.T) {
	s4eq(t, "yes", `{% if v == blank %}yes{% else %}no{% endif %}`, map[string]any{"v": map[string]any{}})
}

// D2 — what is NOT blank
func TestS4_Blank_TrueIsNotBlank(t *testing.T) {
	s4eq(t, "no", `{% if v == blank %}yes{% else %}no{% endif %}`, map[string]any{"v": true})
}

func TestS4_Blank_ZeroIsNotBlank(t *testing.T) {
	s4eq(t, "no", `{% if v == blank %}yes{% else %}no{% endif %}`, map[string]any{"v": 0})
}

func TestS4_Blank_OneIsNotBlank(t *testing.T) {
	s4eq(t, "no", `{% if v == blank %}yes{% else %}no{% endif %}`, map[string]any{"v": 1})
}

func TestS4_Blank_NonEmptyStringIsNotBlank(t *testing.T) {
	s4eq(t, "no", `{% if v == blank %}yes{% else %}no{% endif %}`, map[string]any{"v": "x"})
}

func TestS4_Blank_NonEmptyArrayIsNotBlank(t *testing.T) {
	s4eq(t, "no", `{% if v == blank %}yes{% else %}no{% endif %}`, map[string]any{"v": []any{0}})
}

// D3 — blank equals nil (the nil is blank special case)
func TestS4_Blank_BlankEqualsNilLiteral(t *testing.T) {
	s4eq(t, "yes", `{% if blank == nil %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Blank_NilEqualsBlank(t *testing.T) {
	s4eq(t, "yes", `{% if nil == blank %}yes{% else %}no{% endif %}`, nil)
}

// D4 — blank renders as ""
func TestS4_Blank_RendersAsEmptyString(t *testing.T) {
	s4eq(t, "", `{{ blank }}`, nil)
}

// D5 — blank vs empty: nil is blank but NOT empty
func TestS4_Blank_NilIsBlankButNotEmpty(t *testing.T) {
	s4eq(t, "blank", `{% if v == blank %}blank{% elsif v == empty %}empty{% else %}other{% endif %}`,
		map[string]any{"v": nil})
}

// D6 — assign + blank check
func TestS4_Blank_AssignedWhitespaceIsBlank(t *testing.T) {
	s4eq(t, "blank", `{% assign v = "  " %}{% if v == blank %}blank{% else %}not blank{% endif %}`, nil)
}

func TestS4_Blank_AssignedNonEmpty(t *testing.T) {
	s4eq(t, "not blank", `{% assign v = "hi" %}{% if v == blank %}blank{% else %}not blank{% endif %}`, nil)
}

// D7 — symmetric comparison
func TestS4_Blank_SymmetricComparison(t *testing.T) {
	binds := map[string]any{"v": ""}
	out1 := s4render(t, `{% if v == blank %}yes{% else %}no{% endif %}`, binds)
	out2 := s4render(t, `{% if blank == v %}yes{% else %}no{% endif %}`, binds)
	require.Equal(t, out1, out2)
}

// ═════════════════════════════════════════════════════════════════════════════
// E. Range literal
// ═════════════════════════════════════════════════════════════════════════════

// E1 — contains operator: membership inside range
func TestS4_Range_Contains_Inside(t *testing.T) {
	s4eq(t, "yes", `{% if (1..10) contains 5 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Range_Contains_AtLowerBound(t *testing.T) {
	s4eq(t, "yes", `{% if (1..10) contains 1 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Range_Contains_AtUpperBound(t *testing.T) {
	s4eq(t, "yes", `{% if (1..10) contains 10 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Range_Contains_BelowLower(t *testing.T) {
	s4eq(t, "no", `{% if (1..10) contains 0 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Range_Contains_AboveUpper(t *testing.T) {
	s4eq(t, "no", `{% if (1..10) contains 11 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Range_Contains_NegativeRange(t *testing.T) {
	s4eq(t, "yes", `{% if (-5..5) contains -3 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Range_Contains_NegativeOutside(t *testing.T) {
	s4eq(t, "no", `{% if (-5..5) contains -6 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Range_Contains_SingleElementRange(t *testing.T) {
	s4eq(t, "yes", `{% if (7..7) contains 7 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Range_Contains_SingleElementRangeMiss(t *testing.T) {
	s4eq(t, "no", `{% if (7..7) contains 8 %}yes{% else %}no{% endif %}`, nil)
}

// E2 — contains with variable bounds and check value
func TestS4_Range_Contains_VariableBound(t *testing.T) {
	s4eq(t, "yes", `{% if (1..n) contains 4 %}yes{% else %}no{% endif %}`, map[string]any{"n": 5})
}

func TestS4_Range_Contains_VariableValue(t *testing.T) {
	s4eq(t, "yes", `{% if (1..10) contains v %}yes{% else %}no{% endif %}`, map[string]any{"v": 7})
}

func TestS4_Range_Contains_BothVariable(t *testing.T) {
	s4eq(t, "yes", `{% if (a..b) contains v %}yes{% else %}no{% endif %}`,
		map[string]any{"a": 3, "b": 8, "v": 5})
}

// E3 — range in for loop — correct count and order
func TestS4_Range_ForLoop_Count(t *testing.T) {
	s4eq(t, "5", `{% assign c = 0 %}{% for i in (1..5) %}{% assign c = c | plus: 1 %}{% endfor %}{{ c }}`, nil)
}

func TestS4_Range_ForLoop_Ascending(t *testing.T) {
	s4eq(t, "12345", `{% for i in (1..5) %}{{ i }}{% endfor %}`, nil)
}

func TestS4_Range_ForLoop_Reversed(t *testing.T) {
	s4eq(t, "54321", `{% for i in (1..5) reversed %}{{ i }}{% endfor %}`, nil)
}

func TestS4_Range_ForLoop_FirstLast(t *testing.T) {
	s4eq(t, "F.L",
		`{% for i in (1..3) %}{% if forloop.first %}F{% elsif forloop.last %}L{% else %}.{% endif %}{% endfor %}`,
		nil)
}

func TestS4_Range_ForLoop_WithLimit(t *testing.T) {
	s4eq(t, "123", `{% for i in (1..10) limit:3 %}{{ i }}{% endfor %}`, nil)
}

func TestS4_Range_ForLoop_WithOffset(t *testing.T) {
	s4eq(t, "345", `{% for i in (1..5) offset:2 %}{{ i }}{% endfor %}`, nil)
}

// E4 — range in capture and assign
func TestS4_Range_CaptureAndCompare(t *testing.T) {
	// Count via iterating range
	s4eq(t, "3",
		`{% assign count = 0 %}{% for i in (1..3) %}{% assign count = count | plus: 1 %}{% endfor %}{{ count }}`,
		nil)
}

// ═════════════════════════════════════════════════════════════════════════════
// F. not operator
// ═════════════════════════════════════════════════════════════════════════════

// F1 — basic not
func TestS4_Not_TrueIsFalse(t *testing.T) {
	s4eq(t, "no", `{% if not true %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Not_FalseIsTrue(t *testing.T) {
	s4eq(t, "yes", `{% if not false %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Not_NilIsTrue(t *testing.T) {
	s4eq(t, "yes", `{% if not nil %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Not_ZeroIsFalse(t *testing.T) {
	// 0 is truthy in Liquid, so not 0 is false
	s4eq(t, "no", `{% if not 0 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Not_EmptyStringIsFalse(t *testing.T) {
	// "" is truthy in Liquid, so not "" is false
	s4eq(t, "no", `{% if not "" %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Not_CustomBinding(t *testing.T) {
	s4eq(t, "yes", `{% if not v %}yes{% else %}no{% endif %}`, map[string]any{"v": false})
	s4eq(t, "no", `{% if not v %}yes{% else %}no{% endif %}`, map[string]any{"v": true})
	s4eq(t, "yes", `{% if not v %}yes{% else %}no{% endif %}`, map[string]any{"v": nil})
}

// F2 — not applied to comparisons
func TestS4_Not_NotLessThan(t *testing.T) {
	s4eq(t, "no", `{% if not 1 < 5 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Not_NotGreaterThan(t *testing.T) {
	s4eq(t, "yes", `{% if not 5 < 3 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Not_NotEquals(t *testing.T) {
	s4eq(t, "yes", `{% if not "a" == "b" %}yes{% else %}no{% endif %}`, nil)
}

// F3 — not precedence over and/or
func TestS4_Not_PrecedenceOverOr(t *testing.T) {
	// not 1 < 2 or not 1 > 2
	// = (not true) or (not false)
	// = false or true = true
	s4eq(t, "yes", `{% if not 1 < 2 or not 1 > 2 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Not_PrecedenceOverAnd(t *testing.T) {
	// not 1 < 2 and not 1 > 2
	// = (not true) and (not false)
	// = false and true = false
	s4eq(t, "no", `{% if not 1 < 2 and not 1 > 2 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Not_DoubleNot(t *testing.T) {
	// not not true = not false = true
	s4eq(t, "yes", `{% if not not true %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Not_NotWithContains(t *testing.T) {
	s4eq(t, "yes", `{% if not arr contains "x" %}yes{% else %}no{% endif %}`,
		map[string]any{"arr": []any{"a", "b"}})
}

func TestS4_Not_NotWithContains_False(t *testing.T) {
	s4eq(t, "no", `{% if not arr contains "a" %}yes{% else %}no{% endif %}`,
		map[string]any{"arr": []any{"a", "b"}})
}

// F4 — not in unless (double negation)
func TestS4_Not_InUnless(t *testing.T) {
	// unless not x = unless (not truthy) = unless false = renders (truthy)
	s4eq(t, "yes", `{% unless not v %}yes{% endunless %}`, map[string]any{"v": true})
}

// ═════════════════════════════════════════════════════════════════════════════
// G. nil/null with ordering operators
// ═════════════════════════════════════════════════════════════════════════════

// G1 — null literal on the left
func TestS4_NilOrder_NullLtZero(t *testing.T) {
	s4eq(t, "false", `{% if null < 0 %}true{% else %}false{% endif %}`, nil)
}

func TestS4_NilOrder_NullLteZero(t *testing.T) {
	s4eq(t, "false", `{% if null <= 0 %}true{% else %}false{% endif %}`, nil)
}

func TestS4_NilOrder_NullGtZero(t *testing.T) {
	s4eq(t, "false", `{% if null > 0 %}true{% else %}false{% endif %}`, nil)
}

func TestS4_NilOrder_NullGteZero(t *testing.T) {
	s4eq(t, "false", `{% if null >= 0 %}true{% else %}false{% endif %}`, nil)
}

// G2  — null literal on the right
func TestS4_NilOrder_ZeroLtNull(t *testing.T) {
	s4eq(t, "false", `{% if 0 < null %}true{% else %}false{% endif %}`, nil)
}

func TestS4_NilOrder_ZeroLteNull(t *testing.T) {
	s4eq(t, "false", `{% if 0 <= null %}true{% else %}false{% endif %}`, nil)
}

func TestS4_NilOrder_ZeroGtNull(t *testing.T) {
	s4eq(t, "false", `{% if 0 > null %}true{% else %}false{% endif %}`, nil)
}

func TestS4_NilOrder_ZeroGteNull(t *testing.T) {
	s4eq(t, "false", `{% if 0 >= null %}true{% else %}false{% endif %}`, nil)
}

// G3 — nil keyword (same as null)
func TestS4_NilOrder_NilLteZero(t *testing.T) {
	s4eq(t, "false", `{% if nil <= 0 %}true{% else %}false{% endif %}`, nil)
}

func TestS4_NilOrder_ZeroLteNil(t *testing.T) {
	s4eq(t, "false", `{% if 0 <= nil %}true{% else %}false{% endif %}`, nil)
}

// G4 — Go nil binding in ordering
func TestS4_NilOrder_GoBindingLt(t *testing.T) {
	s4eq(t, "false", `{% if v < 1 %}true{% else %}false{% endif %}`, map[string]any{"v": nil})
}

func TestS4_NilOrder_GoBindingGte(t *testing.T) {
	s4eq(t, "false", `{% if v >= 0 %}true{% else %}false{% endif %}`, map[string]any{"v": nil})
}

// G5 — nil IS equal to nil (equality is fine, ordering is not)
func TestS4_NilOrder_NilEqualsNilIsTrue(t *testing.T) {
	s4eq(t, "true", `{% if nil == nil %}true{% else %}false{% endif %}`, nil)
}

func TestS4_NilOrder_NilNotEqualOneIsTrue(t *testing.T) {
	s4eq(t, "true", `{% if nil != 1 %}true{% else %}false{% endif %}`, nil)
}

// ═════════════════════════════════════════════════════════════════════════════
// H. String escape sequences
// ═════════════════════════════════════════════════════════════════════════════

func TestS4_Escape_Newline(t *testing.T) {
	s4eq(t, "a\nb", `{{ "a\nb" }}`, nil)
}

func TestS4_Escape_Tab(t *testing.T) {
	s4eq(t, "a\tb", `{{ "a\tb" }}`, nil)
}

func TestS4_Escape_CarriageReturn(t *testing.T) {
	s4eq(t, "a\rb", `{{ "a\rb" }}`, nil)
}

func TestS4_Escape_SingleQuoteInSingleQuoted(t *testing.T) {
	s4eq(t, "it's", `{{ 'it\'s' }}`, nil)
}

func TestS4_Escape_DoubleQuoteInDoubleQuoted(t *testing.T) {
	s4eq(t, `say "hi"`, `{{ "say \"hi\"" }}`, nil)
}

func TestS4_Escape_Backslash(t *testing.T) {
	s4eq(t, `a\b`, `{{ 'a\\b' }}`, nil)
}

func TestS4_Escape_InSingleQuotedNewline(t *testing.T) {
	s4eq(t, "x\ny", `{{ 'x\ny' }}`, nil)
}

// H2 — escape sequences in comparisons
func TestS4_Escape_CompareWithNewline(t *testing.T) {
	s4eq(t, "yes", `{% if v == "a\nb" %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "a\nb"})
}

func TestS4_Escape_CompareWithBackslash(t *testing.T) {
	s4eq(t, "yes", `{% if v == "a\\b" %}yes{% else %}no{% endif %}`,
		map[string]any{"v": `a\b`})
}

func TestS4_Escape_CompareWithTab(t *testing.T) {
	s4eq(t, "yes", `{% if v == "x\ty" %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "x\ty"})
}

// H3 — assign escape sequence then use
func TestS4_Escape_AssignAndOutput(t *testing.T) {
	s4eq(t, "line1\nline2", `{% assign v = "line1\nline2" %}{{ v }}`, nil)
}

// ═════════════════════════════════════════════════════════════════════════════
// I. Logical operators (and / or) — right-associativity and section-4 operands
// ═════════════════════════════════════════════════════════════════════════════

func TestS4_Logic_FalseOrTrue(t *testing.T) {
	s4eq(t, "yes", `{% if false or true %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Logic_TrueAndFalse(t *testing.T) {
	s4eq(t, "no", `{% if true and false %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Logic_RightAssoc_OrAndOr(t *testing.T) {
	// true or false and false
	// right-assoc: true or (false and false) = true or false = true
	s4eq(t, "yes", `{% if true or false and false %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Logic_RightAssoc_FourTerms(t *testing.T) {
	// true and false and false or true
	// right-assoc: true and (false and (false or true)) = true and (false and true) = true and false = false
	s4eq(t, "no", `{% if true and false and false or true %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Logic_RangeContainsInOr(t *testing.T) {
	// (1..5) contains 3 or false = true or false = true
	s4eq(t, "yes", `{% if (1..5) contains 3 or false %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Logic_RangeContainsInAnd(t *testing.T) {
	// (1..5) contains 3 and true = true and true = true
	s4eq(t, "yes", `{% if (1..5) contains 3 and true %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Logic_EmptyInOr(t *testing.T) {
	// "" == empty or false = true or false = true
	s4eq(t, "yes", `{% if "" == empty or false %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Logic_BlankInAnd(t *testing.T) {
	// nil == blank and true = true and true = true
	s4eq(t, "yes", `{% if nil == blank and true %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Logic_NilOrderingInOr(t *testing.T) {
	// null < 0 = false; or true = true
	s4eq(t, "yes", `{% if null < 0 or true %}yes{% else %}no{% endif %}`, nil)
}

// ═════════════════════════════════════════════════════════════════════════════
// J. Integration — multiple section 4 features in one template
// ═════════════════════════════════════════════════════════════════════════════

func TestS4_Integration_RangeContainsGate(t *testing.T) {
	// Use range contains to filter output
	out := s4render(t,
		`{% for i in (1..5) %}{% if (2..4) contains i %}{{ i }}{% endif %}{% endfor %}`,
		nil)
	require.Equal(t, "234", out)
}

func TestS4_Integration_NotEmptyAndRange(t *testing.T) {
	// Only output if items is not empty and count is in range
	tpl := `{% if items != empty and (1..10) contains items.size %}ok{% else %}bad{% endif %}`
	s4eq(t, "ok", tpl, map[string]any{"items": []any{1, 2, 3}})
	s4eq(t, "bad", tpl, map[string]any{"items": []any{}})
}

func TestS4_Integration_BlankFallbackWithDefault(t *testing.T) {
	// blank binding → default filter activates
	s4eq(t, "anonymous",
		`{{ name | default: "anonymous" }}`,
		map[string]any{"name": ""})
}

func TestS4_Integration_NilNullAlias(t *testing.T) {
	// nil and null are interchangeable in same template
	s4eq(t, "equal",
		`{% if null == nil %}equal{% else %}not equal{% endif %}`,
		nil)
}

func TestS4_Integration_EscapeInOutput(t *testing.T) {
	// String with escape sequence piped through filter
	s4eq(t, "LINE1 LINE2",
		`{{ "line1\nline2" | upcase | replace: "\n", " " }}`,
		nil)
}

func TestS4_Integration_CaseWhenRange(t *testing.T) {
	// case/when with literal values (not range contains — case doesn't use contains)
	out := s4render(t,
		`{% case v %}{% when 1 %}one{% when 2 %}two{% when 3 %}three{% else %}other{% endcase %}`,
		map[string]any{"v": 2})
	require.Equal(t, "two", out)
}

func TestS4_Integration_AssignEscapedAndCompare(t *testing.T) {
	// Assign escape sequence then compare
	s4eq(t, "yes",
		`{% assign newline = "\n" %}{% if newline == "\n" %}yes{% else %}no{% endif %}`,
		nil)
}

func TestS4_Integration_RangeForLoopWithNotEmpty(t *testing.T) {
	// Loop over range, only print items whose string is not empty
	out := s4render(t,
		`{% for i in (1..3) %}{% assign s = i | append: "" %}{% if s != empty %}[{{ s }}]{% endif %}{% endfor %}`,
		nil)
	require.Equal(t, "[1][2][3]", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// K. Edge cases — unless, case/when, captures, nested loops
// ═════════════════════════════════════════════════════════════════════════════

func TestS4_Edge_UnlessEmpty(t *testing.T) {
	// unless empty string == empty → unless true → don't render
	s4eq(t, "",
		`{% unless v == empty %}show{% endunless %}`,
		map[string]any{"v": ""})
}

func TestS4_Edge_UnlessNonEmpty(t *testing.T) {
	s4eq(t, "show",
		`{% unless v == empty %}show{% endunless %}`,
		map[string]any{"v": "hi"})
}

func TestS4_Edge_CaseWhenWithBlank(t *testing.T) {
	s4eq(t, "blank case",
		`{% case v %}{% when blank %}blank case{% when "" %}empty string{% else %}other{% endcase %}`,
		map[string]any{"v": nil}) // nil is blank
}

func TestS4_Edge_NestedRangeContains(t *testing.T) {
	// Inner loop using range contains as filter
	out := s4render(t, `{% for i in (1..5) %}{% if (2..4) contains i %}{{ i }}{% endif %}{% endfor %}`, nil)
	require.Equal(t, "234", out)
}

func TestS4_Edge_RangeInCapture(t *testing.T) {
	// Capture from a range-driven for loop
	out := s4render(t,
		`{% capture result %}{% for i in (1..3) %}{{ i }}{% unless forloop.last %},{% endunless %}{% endfor %}{% endcapture %}{{ result }}`,
		nil)
	require.Equal(t, "1,2,3", out)
}

func TestS4_Edge_DiamondInElsif(t *testing.T) {
	tpl := `{% if v == 1 %}one{% elsif v <> 2 %}not two{% else %}two{% endif %}`
	s4eq(t, "not two", tpl, map[string]any{"v": 3})
	s4eq(t, "two", tpl, map[string]any{"v": 2})
}

func TestS4_Edge_BlankEmpty_ChainedCheck(t *testing.T) {
	// Distinguish between blank and empty: whitespace is blank but not empty
	tpl := `{% if v == blank and v != empty %}only blank{% elsif v == empty %}empty{% else %}other{% endif %}`
	// "  " is blank but NOT empty (has length > 0)
	s4eq(t, "only blank", tpl, map[string]any{"v": "  "})
	// "" is both blank and empty
	s4eq(t, "empty", tpl, map[string]any{"v": ""})
	// "hi" is neither
	s4eq(t, "other", tpl, map[string]any{"v": "hi"})
}

func TestS4_Edge_NilOrderingShortCircuit(t *testing.T) {
	// Nil ordering returns false; should not cause render error
	out, err := s4renderErr(t, `{% if nil < nil %}y{% else %}n{% endif %}`, nil)
	require.NoError(t, err)
	require.Equal(t, "n", out)
}

func TestS4_Edge_RangeContainsZero(t *testing.T) {
	// Boundary: 0 in range that spans 0
	s4eq(t, "yes", `{% if (-1..1) contains 0 %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Edge_LargeRange(t *testing.T) {
	// Large range — contains should be O(1), not iterate
	out, err := s4renderErr(t, `{% if (1..1000) contains 999 %}yes{% else %}no{% endif %}`, nil)
	require.NoError(t, err)
	require.Equal(t, "yes", out)
}

func TestS4_Edge_NotInConditionChain(t *testing.T) {
	// Realistic: show element only if not in "skip" range
	tpl := `{% for i in (1..6) %}{% if not (3..4) contains i %}{{ i }}{% endif %}{% endfor %}`
	s4eq(t, "1256", tpl, nil)
}

func TestS4_Edge_EmptyAfterAssign_Nil(t *testing.T) {
	// Assign nil-valued expression then check empty
	// nil is not empty (it's blank but not empty)
	s4eq(t, "no", `{% assign v = nothing %}{% if v == empty %}yes{% else %}no{% endif %}`, nil)
}

func TestS4_Edge_BlankAfterCapture_Empty(t *testing.T) {
	// capture nothing → ""  → blank AND empty
	s4eq(t, "blank", `{% capture v %}{% endcapture %}{% if v == blank %}blank{% else %}not blank{% endif %}`, nil)
}
