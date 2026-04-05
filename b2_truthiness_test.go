package liquid_test

// B2 — Truthiness: nil, false, blank, empty
//
// Liquid spec: ONLY nil and false are falsy.
// Everything else — 0, "", [], {}, whitespace strings — is truthy.
//
// This file provides intensive E2E tests using Go variable bindings
// (not template literals) to verify the full pipeline from binding → value
// wrapping → truthiness check in every relevant context.

import (
	"testing"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/require"
)

// renderB2 is a small helper: parse, render, return the string or fail.
func renderB2(t *testing.T, tpl string, bindings map[string]any) string {
	t.Helper()
	eng := liquid.NewEngine()
	out, err := eng.ParseAndRenderString(tpl, bindings)
	require.NoError(t, err, "template: %s", tpl)
	return out
}

// ── 1. Core falsy values (only nil and false) ─────────────────────────────────

func TestB2_NilIsFalsy(t *testing.T) {
	// Go nil via binding
	require.Equal(t, "falsy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": nil}))
}

func TestB2_FalseIsFalsy(t *testing.T) {
	require.Equal(t, "falsy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": false}))
}

func TestB2_UnsetVariableIsFalsy(t *testing.T) {
	// Missing key in bindings → nil → falsy
	require.Equal(t, "falsy", renderB2(t,
		`{% if missing %}truthy{% else %}falsy{% endif %}`,
		nil))
}

// ── 2. Core truthy values ────────────────────────────────────────────────────

func TestB2_TrueIsTruthy(t *testing.T) {
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": true}))
}

func TestB2_ZeroIntIsTruthy(t *testing.T) {
	// 0 is truthy in Liquid — NOT falsy like in JavaScript/Python
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": 0}))
}

func TestB2_ZeroFloatIsTruthy(t *testing.T) {
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": 0.0}))
}

func TestB2_EmptyStringIsTruthy(t *testing.T) {
	// "" is truthy in Liquid
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": ""}))
}

func TestB2_WhitespaceStringIsTruthy(t *testing.T) {
	// "   " is truthy (whitespace is truthy, even though it's "blank")
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": "   "}))
}

func TestB2_NonEmptyStringIsTruthy(t *testing.T) {
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": "hello"}))
}

func TestB2_EmptyArrayIsTruthy(t *testing.T) {
	// [] is truthy in Liquid (unlike Ruby arrays in Ruby truthiness)
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": []any{}}))
}

func TestB2_NonEmptyArrayIsTruthy(t *testing.T) {
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": []any{1, 2, 3}}))
}

func TestB2_EmptyMapIsTruthy(t *testing.T) {
	// {} is truthy in Liquid
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": map[string]any{}}))
}

func TestB2_NonEmptyMapIsTruthy(t *testing.T) {
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": map[string]any{"k": "val"}}))
}

func TestB2_PositiveIntIsTruthy(t *testing.T) {
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": 42}))
}

func TestB2_NegativeIntIsTruthy(t *testing.T) {
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": -1}))
}

func TestB2_PositiveFloatIsTruthy(t *testing.T) {
	require.Equal(t, "truthy", renderB2(t,
		`{% if v %}truthy{% else %}falsy{% endif %}`,
		map[string]any{"v": 3.14}))
}

// ── 3. unless — negation of truthiness ───────────────────────────────────────

func TestB2_UnlessNilRenders(t *testing.T) {
	// unless nil → nil is falsy → negated = truthy → renders
	require.Equal(t, "yes", renderB2(t,
		`{% unless v %}yes{% endunless %}`,
		map[string]any{"v": nil}))
}

func TestB2_UnlessFalseRenders(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% unless v %}yes{% endunless %}`,
		map[string]any{"v": false}))
}

func TestB2_UnlessTrueDoesNotRender(t *testing.T) {
	require.Equal(t, "", renderB2(t,
		`{% unless v %}yes{% endunless %}`,
		map[string]any{"v": true}))
}

func TestB2_UnlessZeroDoesNotRender(t *testing.T) {
	// 0 is truthy → unless 0 = negated truthy = false → doesn't render
	require.Equal(t, "", renderB2(t,
		`{% unless v %}yes{% endunless %}`,
		map[string]any{"v": 0}))
}

func TestB2_UnlessEmptyStringDoesNotRender(t *testing.T) {
	// "" is truthy in Liquid
	require.Equal(t, "", renderB2(t,
		`{% unless v %}yes{% endunless %}`,
		map[string]any{"v": ""}))
}

func TestB2_UnlessEmptyArrayDoesNotRender(t *testing.T) {
	// [] is truthy in Liquid
	require.Equal(t, "", renderB2(t,
		`{% unless v %}yes{% endunless %}`,
		map[string]any{"v": []any{}}))
}

// ── 4. not operator ───────────────────────────────────────────────────────────

func TestB2_NotNilIsTruthy(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if not v %}yes{% endif %}`,
		map[string]any{"v": nil}))
}

func TestB2_NotFalseIsTruthy(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if not v %}yes{% endif %}`,
		map[string]any{"v": false}))
}

func TestB2_NotTrueIsFalsy(t *testing.T) {
	require.Equal(t, "", renderB2(t,
		`{% if not v %}yes{% endif %}`,
		map[string]any{"v": true}))
}

func TestB2_NotZeroIsFalsy(t *testing.T) {
	// 0 is truthy, so not 0 is falsy
	require.Equal(t, "", renderB2(t,
		`{% if not v %}yes{% endif %}`,
		map[string]any{"v": 0}))
}

func TestB2_NotEmptyStringIsFalsy(t *testing.T) {
	// "" is truthy, so not "" is falsy
	require.Equal(t, "", renderB2(t,
		`{% if not v %}yes{% endif %}`,
		map[string]any{"v": ""}))
}

func TestB2_NotEmptyArrayIsFalsy(t *testing.T) {
	// [] is truthy, so not [] is falsy
	require.Equal(t, "", renderB2(t,
		`{% if not v %}yes{% endif %}`,
		map[string]any{"v": []any{}}))
}

// ── 5. and / or with falsy/truthy values via bindings ─────────────────────────

func TestB2_NilAndTrueIsFalsy(t *testing.T) {
	// nil is falsy → nil and true = false
	require.Equal(t, "no", renderB2(t,
		`{% if v and true %}yes{% else %}no{% endif %}`,
		map[string]any{"v": nil}))
}

func TestB2_FalseAndTrueIsFalsy(t *testing.T) {
	require.Equal(t, "no", renderB2(t,
		`{% if v and true %}yes{% else %}no{% endif %}`,
		map[string]any{"v": false}))
}

func TestB2_ZeroAndTrueIsTruthy(t *testing.T) {
	// 0 is truthy → 0 and true = true
	require.Equal(t, "yes", renderB2(t,
		`{% if v and true %}yes{% else %}no{% endif %}`,
		map[string]any{"v": 0}))
}

func TestB2_EmptyStringAndTrueIsTruthy(t *testing.T) {
	// "" is truthy in Liquid
	require.Equal(t, "yes", renderB2(t,
		`{% if v and true %}yes{% else %}no{% endif %}`,
		map[string]any{"v": ""}))
}

func TestB2_NilOrTrueIsTruthy(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if v or true %}yes{% else %}no{% endif %}`,
		map[string]any{"v": nil}))
}

func TestB2_FalseOrFalseIsFalsy(t *testing.T) {
	require.Equal(t, "no", renderB2(t,
		`{% if v or false %}yes{% else %}no{% endif %}`,
		map[string]any{"v": false}))
}

func TestB2_NilOrNilIsFalsy(t *testing.T) {
	require.Equal(t, "no", renderB2(t,
		`{% if v or w %}yes{% else %}no{% endif %}`,
		map[string]any{"v": nil, "w": nil}))
}

func TestB2_ZeroOrFalseIsTruthy(t *testing.T) {
	// 0 is truthy
	require.Equal(t, "yes", renderB2(t,
		`{% if v or false %}yes{% else %}no{% endif %}`,
		map[string]any{"v": 0}))
}

// ── 6. case/when matching nil and false ───────────────────────────────────────

func TestB2_CaseWhenNilMatchesNil(t *testing.T) {
	require.Equal(t, "was nil", renderB2(t,
		`{% case v %}{% when nil %}was nil{% when false %}was false{% else %}other{% endcase %}`,
		map[string]any{"v": nil}))
}

func TestB2_CaseWhenFalseMatchesFalse(t *testing.T) {
	require.Equal(t, "was false", renderB2(t,
		`{% case v %}{% when nil %}was nil{% when false %}was false{% else %}other{% endcase %}`,
		map[string]any{"v": false}))
}

func TestB2_CaseWhenZeroDoesNotMatchNilOrFalse(t *testing.T) {
	// 0 should not match nil or false — they are not equal
	require.Equal(t, "other", renderB2(t,
		`{% case v %}{% when nil %}was nil{% when false %}was false{% else %}other{% endcase %}`,
		map[string]any{"v": 0}))
}

func TestB2_CaseWhenEmptyStringDoesNotMatchNilOrFalse(t *testing.T) {
	require.Equal(t, "other", renderB2(t,
		`{% case v %}{% when nil %}was nil{% when false %}was false{% else %}other{% endcase %}`,
		map[string]any{"v": ""}))
}

// ── 7. default filter ─────────────────────────────────────────────────────────

func TestB2_DefaultFilter_NilActivatesDefault(t *testing.T) {
	require.Equal(t, "fallback", renderB2(t,
		`{{ v | default: "fallback" }}`,
		map[string]any{"v": nil}))
}

func TestB2_DefaultFilter_FalseActivatesDefault(t *testing.T) {
	require.Equal(t, "fallback", renderB2(t,
		`{{ v | default: "fallback" }}`,
		map[string]any{"v": false}))
}

func TestB2_DefaultFilter_EmptyStringActivatesDefault(t *testing.T) {
	// "" triggers default (empty string is blank for the default filter)
	require.Equal(t, "fallback", renderB2(t,
		`{{ v | default: "fallback" }}`,
		map[string]any{"v": ""}))
}

func TestB2_DefaultFilter_EmptyArrayActivatesDefault(t *testing.T) {
	require.Equal(t, "fallback", renderB2(t,
		`{{ v | default: "fallback" }}`,
		map[string]any{"v": []any{}}))
}

func TestB2_DefaultFilter_EmptyMapActivatesDefault(t *testing.T) {
	require.Equal(t, "fallback", renderB2(t,
		`{{ v | default: "fallback" }}`,
		map[string]any{"v": map[string]any{}}))
}

func TestB2_DefaultFilter_ZeroDoesNotActivateDefault(t *testing.T) {
	// 0 is NOT blank for default filter — it should pass through
	require.Equal(t, "0", renderB2(t,
		`{{ v | default: "fallback" }}`,
		map[string]any{"v": 0}))
}

func TestB2_DefaultFilter_TrueDoesNotActivateDefault(t *testing.T) {
	require.Equal(t, "true", renderB2(t,
		`{{ v | default: "fallback" }}`,
		map[string]any{"v": true}))
}

func TestB2_DefaultFilter_NonEmptyStringDoesNotActivateDefault(t *testing.T) {
	require.Equal(t, "hello", renderB2(t,
		`{{ v | default: "fallback" }}`,
		map[string]any{"v": "hello"}))
}

func TestB2_DefaultFilter_NonEmptyArrayDoesNotActivateDefault(t *testing.T) {
	require.Equal(t, "1 2 3", renderB2(t,
		`{{ v | default: "fallback" | join: " " }}`,
		map[string]any{"v": []any{1, 2, 3}}))
}

func TestB2_DefaultFilter_AllowFalsePreservesFalse(t *testing.T) {
	// allow_false: true → false is preserved, not replaced by default
	require.Equal(t, "false", renderB2(t,
		`{{ v | default: "fallback", allow_false: true }}`,
		map[string]any{"v": false}))
}

func TestB2_DefaultFilter_AllowFalseNilStillActivatesDefault(t *testing.T) {
	// allow_false: true → nil still activates default
	require.Equal(t, "fallback", renderB2(t,
		`{{ v | default: "fallback", allow_false: true }}`,
		map[string]any{"v": nil}))
}

func TestB2_DefaultFilter_AllowFalseEmptyStringStillActivatesDefault(t *testing.T) {
	// allow_false: true → "" still activates default (empty string is blank)
	require.Equal(t, "fallback", renderB2(t,
		`{{ v | default: "fallback", allow_false: true }}`,
		map[string]any{"v": ""}))
}

// ── 8. where filter (no value → truthy semantics) ────────────────────────────

func TestB2_WhereFilter_NoValueUsesLiquidTruthiness(t *testing.T) {
	// Only truthy-property items pass; nil and false are falsy
	items := []any{
		map[string]any{"x": nil},   // falsy
		map[string]any{"x": false}, // falsy
		map[string]any{"x": true},  // truthy
		map[string]any{"x": 0},     // truthy (0 is truthy in Liquid!)
		map[string]any{"x": ""},    // truthy ("" is truthy in Liquid!)
		map[string]any{"x": 1},     // truthy
		map[string]any{"x": "hi"},  // truthy
	}
	out := renderB2(t,
		`{% assign filtered = arr | where: "x" %}{{ filtered | size }}`,
		map[string]any{"arr": items})
	// nil(0), false(0), true(1), 0(1), ""(1), 1(1), "hi"(1) → 5 truthy items
	require.Equal(t, "5", out)
}

func TestB2_WhereFilter_OnlyNilAndFalseAreFiltered(t *testing.T) {
	// Confirm nil and false are the only items removed
	items := []any{
		map[string]any{"active": nil},
		map[string]any{"active": false},
		map[string]any{"active": 0},
		map[string]any{"active": ""},
		map[string]any{"active": "yes"},
	}
	out := renderB2(t,
		`{% assign filtered = arr | where: "active" %}{{ filtered | size }}`,
		map[string]any{"arr": items})
	// 0, "", "yes" are truthy → 3 pass
	require.Equal(t, "3", out)
}

// ── 9. blank keyword comparisons via Go bindings ─────────────────────────────

func TestB2_BlankKeyword_NilIsBlank(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": nil}))
}

func TestB2_BlankKeyword_FalseIsBlank(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": false}))
}

func TestB2_BlankKeyword_EmptyStringIsBlank(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": ""}))
}

func TestB2_BlankKeyword_WhitespaceStringIsBlank(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "   "}))
}

func TestB2_BlankKeyword_EmptyArrayIsBlank(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": []any{}}))
}

func TestB2_BlankKeyword_EmptyMapIsBlank(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": map[string]any{}}))
}

func TestB2_BlankKeyword_ZeroIsNotBlank(t *testing.T) {
	require.Equal(t, "no", renderB2(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": 0}))
}

func TestB2_BlankKeyword_TrueIsNotBlank(t *testing.T) {
	require.Equal(t, "no", renderB2(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": true}))
}

func TestB2_BlankKeyword_NonEmptyStringIsNotBlank(t *testing.T) {
	require.Equal(t, "no", renderB2(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "hello"}))
}

func TestB2_BlankKeyword_NonEmptyArrayIsNotBlank(t *testing.T) {
	require.Equal(t, "no", renderB2(t,
		`{% if v == blank %}yes{% else %}no{% endif %}`,
		map[string]any{"v": []any{1}}))
}

// ── 10. empty keyword comparisons via Go bindings ─────────────────────────────

func TestB2_EmptyKeyword_EmptyStringIsEmpty(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": ""}))
}

func TestB2_EmptyKeyword_EmptyArrayIsEmpty(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": []any{}}))
}

func TestB2_EmptyKeyword_EmptyMapIsEmpty(t *testing.T) {
	require.Equal(t, "yes", renderB2(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": map[string]any{}}))
}

func TestB2_EmptyKeyword_NilIsNotEmpty(t *testing.T) {
	// nil is NOT empty (empty = string/array/map with 0 length)
	require.Equal(t, "no", renderB2(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": nil}))
}

func TestB2_EmptyKeyword_FalseIsNotEmpty(t *testing.T) {
	// false is NOT empty (only blank)
	require.Equal(t, "no", renderB2(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": false}))
}

func TestB2_EmptyKeyword_ZeroIsNotEmpty(t *testing.T) {
	require.Equal(t, "no", renderB2(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": 0}))
}

func TestB2_EmptyKeyword_WhitespaceStringIsNotEmpty(t *testing.T) {
	// "  " is NOT empty (len > 0) — it IS blank though
	require.Equal(t, "no", renderB2(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "   "}))
}

func TestB2_EmptyKeyword_NonEmptyStringIsNotEmpty(t *testing.T) {
	require.Equal(t, "no", renderB2(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": "hello"}))
}

func TestB2_EmptyKeyword_NonEmptyArrayIsNotEmpty(t *testing.T) {
	require.Equal(t, "no", renderB2(t,
		`{% if v == empty %}yes{% else %}no{% endif %}`,
		map[string]any{"v": []any{1, 2}}))
}

// ── 11. Combination: nil/false in conditional chains ─────────────────────────

func TestB2_ElsifWithNilFallsThrough(t *testing.T) {
	out := renderB2(t, `{% if a %}A{% elsif b %}B{% else %}neither{% endif %}`,
		map[string]any{"a": nil, "b": nil})
	require.Equal(t, "neither", out)
}

func TestB2_ElsifWithZeroTakesFirstBranch(t *testing.T) {
	// a=0 is truthy → takes first branch
	out := renderB2(t, `{% if a %}A{% elsif b %}B{% else %}neither{% endif %}`,
		map[string]any{"a": 0, "b": "hi"})
	require.Equal(t, "A", out)
}

func TestB2_CompoundAndWithNilShortCircuits(t *testing.T) {
	// nil and truthy_value = false
	out := renderB2(t, `{% if a and b %}yes{% else %}no{% endif %}`,
		map[string]any{"a": nil, "b": "anything"})
	require.Equal(t, "no", out)
}

func TestB2_CompoundOrWithNilFallbackToSecond(t *testing.T) {
	// nil or truthy = true
	out := renderB2(t, `{% if a or b %}yes{% else %}no{% endif %}`,
		map[string]any{"a": nil, "b": 1})
	require.Equal(t, "yes", out)
}

func TestB2_CompoundOrBothFalsy(t *testing.T) {
	out := renderB2(t, `{% if a or b %}yes{% else %}no{% endif %}`,
		map[string]any{"a": false, "b": nil})
	require.Equal(t, "no", out)
}

// ── 12. Capture / assign do not affect truthiness semantics ──────────────────

func TestB2_CapturedEmptyStringIsTruthy(t *testing.T) {
	// {% capture x %}{% endcapture %} captures "" which is truthy
	out := renderB2(t,
		`{% capture x %}{% endcapture %}{% if x %}truthy{% else %}falsy{% endif %}`,
		nil)
	require.Equal(t, "truthy", out)
}

func TestB2_AssignedZeroIsTruthy(t *testing.T) {
	out := renderB2(t,
		`{% assign x = 0 %}{% if x %}truthy{% else %}falsy{% endif %}`,
		nil)
	require.Equal(t, "truthy", out)
}

func TestB2_AssignedFalseIsFalsy(t *testing.T) {
	out := renderB2(t,
		`{% assign x = false %}{% if x %}truthy{% else %}falsy{% endif %}`,
		nil)
	require.Equal(t, "falsy", out)
}

func TestB2_AssignedNilIsFalsy(t *testing.T) {
	out := renderB2(t,
		`{% assign x = nil %}{% if x %}truthy{% else %}falsy{% endif %}`,
		nil)
	require.Equal(t, "falsy", out)
}
