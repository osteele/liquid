package liquid_test

// Ported expression/literal tests from:
//   - Ruby Liquid: test/integration/expression_test.rb
//   - Ruby Liquid: test/integration/tags/statements_test.rb
//   - Ruby Liquid: test/unit/condition_unit_test.rb
//   - LiquidJS:    src/render/expression.spec.ts
//   - LiquidJS:    src/render/boolean.spec.ts
//   - LiquidJS:    src/render/string.spec.ts
//   - LiquidJS:    test/integration/drop/empty-drop.spec.ts
//   - LiquidJS:    test/integration/drop/blank-drop.spec.ts
//
// Covers checklist section 4: Expressões / Literais.
// Added in second pass (2026-04):
//   - 4.10 Range contains operator (JS expression.spec.ts)
//   - 4.11 nil/null with ordering operators (Ruby statements_test.rb)

import (
	"testing"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/require"
)

// render is a helper that renders src with the given bindings and returns the
// output. It fails the test immediately on any error.
func renderExpr(t *testing.T, eng *liquid.Engine, src string, bindings map[string]any) string {
	t.Helper()
	out, err := eng.ParseAndRenderString(src, bindings)
	require.NoError(t, err, "template: %s", src)
	return out
}

// ── 4.1 Basic Literals ────────────────────────────────────────────────────────
// Sources: Ruby expression_test.rb (test_keyword_literals, test_string,
//          test_int, test_float, test_range)

func TestPortedLiterals_BasicOutput(t *testing.T) {
	eng := liquid.NewEngine()

	t.Run("bool true renders as true", func(t *testing.T) {
		// Ruby: assert_template_result("true", "{{ true }}")
		require.Equal(t, "true", renderExpr(t, eng, `{{ true }}`, nil))
	})

	t.Run("bool false renders as false", func(t *testing.T) {
		require.Equal(t, "false", renderExpr(t, eng, `{{ false }}`, nil))
	})

	t.Run("nil renders as empty string", func(t *testing.T) {
		require.Equal(t, "", renderExpr(t, eng, `{{ nil }}`, nil))
	})

	t.Run("null renders as empty string (alias for nil)", func(t *testing.T) {
		require.Equal(t, "", renderExpr(t, eng, `{{ null }}`, nil))
	})

	t.Run("integer literal", func(t *testing.T) {
		// Ruby: assert_template_result("456", "{{ 456 }}")
		require.Equal(t, "456", renderExpr(t, eng, `{{ 456 }}`, nil))
	})

	t.Run("negative integer literal", func(t *testing.T) {
		require.Equal(t, "-7", renderExpr(t, eng, `{{ -7 }}`, nil))
	})

	t.Run("float literal", func(t *testing.T) {
		// Ruby: assert_template_result("2.5", "{{ 2.5 }}")
		require.Equal(t, "2.5", renderExpr(t, eng, `{{ 2.5 }}`, nil))
	})

	t.Run("negative float literal", func(t *testing.T) {
		// Ruby: assert_template_result("-17.42", "{{ -17.42 }}")
		require.Equal(t, "-17.42", renderExpr(t, eng, `{{ -17.42 }}`, nil))
	})

	t.Run("single-quoted string literal", func(t *testing.T) {
		// Ruby: assert_template_result("single quoted", "{{'single quoted'}}")
		require.Equal(t, "single quoted", renderExpr(t, eng, `{{'single quoted'}}`, nil))
	})

	t.Run("double-quoted string literal", func(t *testing.T) {
		// Ruby: assert_template_result("double quoted", '{{"double quoted"}}')
		require.Equal(t, "double quoted", renderExpr(t, eng, `{{"double quoted"}}`, nil))
	})

	t.Run("string with emoji", func(t *testing.T) {
		// Ruby: assert_template_result("emoji🔥", "{{ 'emoji🔥' }}")
		require.Equal(t, "emoji🔥", renderExpr(t, eng, `{{ 'emoji🔥' }}`, nil))
	})

	t.Run("range output includes dots", func(t *testing.T) {
		// Ruby: assert_template_result("3..4", "{{ ( 3 .. 4 ) }}")
		require.Equal(t, "3..4", renderExpr(t, eng, `{{ ( 3 .. 4 ) }}`, nil))
	})
}

// ── 4.2 Truthiness ───────────────────────────────────────────────────────────
// Source: LiquidJS src/render/boolean.spec.ts
// Spec: https://shopify.github.io/liquid/basics/truthy-and-falsy/
// In Shopify Liquid only nil and false are falsy; everything else is truthy.

func TestPortedLiterals_Truthiness(t *testing.T) {
	eng := liquid.NewEngine()

	check := func(src string, bindings map[string]any, want string) {
		t.Helper()
		require.Equal(t, want, renderExpr(t, eng, src, bindings))
	}

	// true is truthy
	t.Run("true is truthy", func(t *testing.T) {
		check(`{% if true %}truthy{% else %}falsy{% endif %}`, nil, "truthy")
	})

	// false is falsy
	t.Run("false is falsy", func(t *testing.T) {
		check(`{% if false %}truthy{% else %}falsy{% endif %}`, nil, "falsy")
	})

	// nil is falsy
	t.Run("nil is falsy", func(t *testing.T) {
		check(`{% if nil %}truthy{% else %}falsy{% endif %}`, nil, "falsy")
	})

	// null is falsy (alias for nil)
	t.Run("null is falsy", func(t *testing.T) {
		check(`{% if null %}truthy{% else %}falsy{% endif %}`, nil, "falsy")
	})

	// "foo" is truthy
	t.Run("non-empty string is truthy", func(t *testing.T) {
		check(`{% if "foo" %}truthy{% else %}falsy{% endif %}`, nil, "truthy")
	})

	// "" is truthy (empty string — NOT falsy in Shopify Liquid)
	t.Run("empty string is truthy", func(t *testing.T) {
		// LiquidJS boolean.spec.ts: '"" is truthy'
		check(`{% if "" %}truthy{% else %}falsy{% endif %}`, nil, "truthy")
	})

	// 0 is truthy (zero — NOT falsy in Shopify Liquid)
	t.Run("zero integer is truthy", func(t *testing.T) {
		// LiquidJS boolean.spec.ts: '0 is truthy'
		check(`{% if 0 %}truthy{% else %}falsy{% endif %}`, nil, "truthy")
	})

	// 1 is truthy
	t.Run("positive integer is truthy", func(t *testing.T) {
		check(`{% if 1 %}truthy{% else %}falsy{% endif %}`, nil, "truthy")
	})

	// 1.1 is truthy
	t.Run("float is truthy", func(t *testing.T) {
		check(`{% if 1.1 %}truthy{% else %}falsy{% endif %}`, nil, "truthy")
	})

	// [] is truthy (empty array is NOT falsy in Shopify Liquid)
	t.Run("empty array is truthy", func(t *testing.T) {
		// LiquidJS boolean.spec.ts: '[] is truthy'
		check(`{% if arr %}truthy{% else %}falsy{% endif %}`, map[string]any{"arr": []any{}}, "truthy")
	})

	// [1] is truthy
	t.Run("non-empty array is truthy", func(t *testing.T) {
		check(`{% if arr %}truthy{% else %}falsy{% endif %}`, map[string]any{"arr": []any{1}}, "truthy")
	})

	// unset variable → nil → falsy
	t.Run("unset variable is falsy (nil)", func(t *testing.T) {
		check(`{% if missing_var %}truthy{% else %}falsy{% endif %}`, nil, "falsy")
	})
}

// ── 4.3 empty literal ────────────────────────────────────────────────────────
// Sources: LiquidJS test/integration/drop/empty-drop.spec.ts
//          Ruby   test/integration/tags/statements_test.rb

func TestPortedLiterals_Empty(t *testing.T) {
	eng := liquid.NewEngine()

	check := func(src string, bindings map[string]any, want string) {
		t.Helper()
		require.Equal(t, want, renderExpr(t, eng, src, bindings))
	}

	// LiquidJS: render empty drop as empty string
	t.Run("empty renders as empty string", func(t *testing.T) {
		check(`{{empty}}`, nil, "")
	})

	// "" is empty
	t.Run("empty string equals empty", func(t *testing.T) {
		// LiquidJS: '"" is empty'
		check(`{%if "" == empty %}"" == empty{%else%}"" != empty{% endif %}`, nil, `"" == empty`)
	})

	// "  " is NOT empty (only blank handles whitespace)
	t.Run("whitespace string is not empty", func(t *testing.T) {
		// LiquidJS: '"  " is not empty'
		check(`{%if "  " == empty %}"  " == empty{%else%}"  " != empty{% endif %}`, nil, `"  " != empty`)
	})

	// nil is NOT empty — Ruby: nil != empty, JS: nil != empty
	t.Run("nil is not empty", func(t *testing.T) {
		// LiquidJS: 'nil is not empty'
		check(`{%if nil == empty %}nil == empty{%else%}nil != empty{% endif %}`, nil, "nil != empty")
	})

	// false is NOT empty
	t.Run("false is not empty", func(t *testing.T) {
		// LiquidJS: 'false is not empty'
		check(`{%if false == empty %}false == empty{%else%}false != empty{% endif %}`, nil, "false != empty")
	})

	// [] is empty
	t.Run("empty array equals empty", func(t *testing.T) {
		// LiquidJS: '[] is empty' ; Ruby: test_is_collection_empty
		check(`{%if arr == empty %}[] == empty{%else%}[] != empty{% endif %}`, map[string]any{"arr": []any{}}, "[] == empty")
	})

	// [1] is NOT empty
	t.Run("non-empty array is not empty", func(t *testing.T) {
		// Ruby: test_is_not_collection_empty
		check(`{%if arr == empty %}[1] == empty{%else%}[1] != empty{% endif %}`, map[string]any{"arr": []any{1}}, "[1] != empty")
	})

	// {} is empty (empty map)
	t.Run("empty map equals empty", func(t *testing.T) {
		// LiquidJS: '{} is empty'
		check(`{%if obj == empty %}{} == empty{%else%}{} != empty{% endif %}`, map[string]any{"obj": map[string]any{}}, "{} == empty")
	})

	// {foo:1} is NOT empty
	t.Run("non-empty map is not empty", func(t *testing.T) {
		// LiquidJS: '{foo: 1} is not empty'
		check(`{%if obj == empty %}{foo: 1} == empty{%else%}{foo: 1} != empty{% endif %}`, map[string]any{"obj": map[string]any{"foo": 1}}, "{foo: 1} != empty")
	})

	// Numeric comparisons with empty always return false (empty has no ordering)
	t.Run("1 less than empty is false", func(t *testing.T) {
		// LiquidJS: '1 < empty should be false'
		check(`{%if 1 < empty %}true{%else%}false{% endif %}`, nil, "false")
	})
	t.Run("1 less-or-equal empty is false", func(t *testing.T) {
		check(`{%if 1 <= empty %}true{%else%}false{% endif %}`, nil, "false")
	})
	t.Run("1 greater than empty is false", func(t *testing.T) {
		check(`{%if 1 > empty %}true{%else%}false{% endif %}`, nil, "false")
	})
	t.Run("1 greater-or-equal empty is false", func(t *testing.T) {
		check(`{%if 1 >= empty %}true{%else%}false{% endif %}`, nil, "false")
	})
	t.Run("1 equals empty is false", func(t *testing.T) {
		check(`{%if 1 == empty %}true{%else%}false{% endif %}`, nil, "false")
	})
	t.Run("1 not-equals empty is true", func(t *testing.T) {
		check(`{%if 1 != empty %}true{%else%}false{% endif %}`, nil, "true")
	})

	// empty does not equal itself (special Liquid semantic)
	t.Run("empty does not equal empty", func(t *testing.T) {
		// LiquidJS: 'empty != empty' (empty == empty → false)
		check(`{%if empty == empty %}true{%else%}false{% endif %}`, nil, "false")
	})

	// empty does not equal nil
	t.Run("empty does not equal nil", func(t *testing.T) {
		// LiquidJS: 'empty != nil'
		check(`{%if empty == nil %}true{%else%}false{% endif %}`, nil, "false")
	})
}

// ── 4.4 blank literal ────────────────────────────────────────────────────────
// Sources: LiquidJS test/integration/drop/blank-drop.spec.ts
//          Ruby   test/unit/condition_unit_test.rb (blank tests)

func TestPortedLiterals_Blank(t *testing.T) {
	eng := liquid.NewEngine()

	check := func(src string, bindings map[string]any, want string) {
		t.Helper()
		require.Equal(t, want, renderExpr(t, eng, src, bindings))
	}

	// LiquidJS: render blank drop as blank string
	t.Run("blank renders as empty string", func(t *testing.T) {
		check(`{{blank}}`, nil, "")
	})

	// blank equals nil (JS + Ruby)
	t.Run("blank equals nil", func(t *testing.T) {
		// LiquidJS: 'blank equals nil'
		check(`{%if blank == nil %}blank == nil{%else%}blank != nil{% endif %}`, nil, "blank == nil")
	})

	// nil == blank (symmetric)
	t.Run("nil equals blank", func(t *testing.T) {
		// Ruby condition_unit_test: test_blank_with_nil
		check(`{% if x == blank %}yes{% endif %}`, map[string]any{"x": nil}, "yes")
	})

	// false is blank
	t.Run("false is blank", func(t *testing.T) {
		// LiquidJS: 'false is blank' ; Ruby: test_blank_with_false
		check(`{%if false == blank %}false == blank{%else%}false != blank{% endif %}`, nil, "false == blank")
	})

	// "" is blank
	t.Run("empty string is blank", func(t *testing.T) {
		// LiquidJS: '"" is blank' ; Ruby: test_blank_with_empty_string
		check(`{%if "" == blank %}"" == blank{%else%}"" != blank{% endif %}`, nil, `"" == blank`)
	})

	// "  " is blank (whitespace only)
	t.Run("whitespace-only string is blank", func(t *testing.T) {
		// LiquidJS: '"  " is blank' ; Ruby: test_blank_with_whitespace_string
		check(`{%if "  " == blank %}"  " == blank{%else%}"  " != blank{% endif %}`, nil, `"  " == blank`)
	})

	// {} is blank (empty map)
	t.Run("empty map is blank", func(t *testing.T) {
		// LiquidJS: '{} is blank' ; Ruby: test_blank_with_empty_hash
		check(`{%if obj == blank %}{} == blank{%else%}{} != blank{% endif %}`, map[string]any{"obj": map[string]any{}}, "{} == blank")
	})

	// {foo:1} is NOT blank
	t.Run("non-empty map is not blank", func(t *testing.T) {
		// LiquidJS: '{foo: 1} is not blank'
		check(`{%if obj == blank %}{foo: 1} == blank{%else%}{foo: 1} != blank{% endif %}`, map[string]any{"obj": map[string]any{"foo": 1}}, "{foo: 1} != blank")
	})

	// [] is blank
	t.Run("empty array is blank", func(t *testing.T) {
		// LiquidJS: '[] is blank' ; Ruby: test_blank_with_empty_array
		check(`{%if arr == blank %}[] == blank{%else%}[] != blank{% endif %}`, map[string]any{"arr": []any{}}, "[] == blank")
	})

	// [1] is NOT blank
	t.Run("non-empty array is not blank", func(t *testing.T) {
		// LiquidJS: '[1] is not blank'
		check(`{%if arr == blank %}[1] == blank{%else%}[1] != blank{% endif %}`, map[string]any{"arr": []any{1}}, "[1] != blank")
	})

	// true is not blank
	t.Run("true is not blank", func(t *testing.T) {
		// Ruby: test_not_blank_with_true
		check(`{% if x == blank %}yes{% else %}no{% endif %}`, map[string]any{"x": true}, "no")
	})

	// 42 is not blank (numbers are never blank)
	t.Run("number is not blank", func(t *testing.T) {
		// Ruby: test_not_blank_with_number
		check(`{% if x == blank %}yes{% else %}no{% endif %}`, map[string]any{"x": 42}, "no")
	})

	// "hello" is not blank
	t.Run("non-empty string is not blank", func(t *testing.T) {
		// Ruby: test_not_blank_with_string_content
		check(`{% if x == blank %}yes{% else %}no{% endif %}`, map[string]any{"x": "hello"}, "no")
	})

	// [1, 2, 3] is not blank
	t.Run("non-empty array var is not blank", func(t *testing.T) {
		// Ruby: test_not_blank_with_non_empty_array
		check(`{% if arr == blank %}yes{% else %}no{% endif %}`, map[string]any{"arr": []any{1, 2, 3}}, "no")
	})
}

// ── 4.5 `<>` Operator (alias for !=) ─────────────────────────────────────────
// Source: Ruby test/unit/condition_unit_test.rb (test_default_operators_evalute_true)

func TestPortedLiterals_DiamondOperator(t *testing.T) {
	eng := liquid.NewEngine()

	check := func(src string, want string) {
		t.Helper()
		require.Equal(t, want, renderExpr(t, eng, src, nil))
	}

	// Ruby: assert_evaluates_true(1, '<>', 2)
	t.Run("1 <> 2 is true", func(t *testing.T) {
		check(`{% if 1 <> 2 %}true{% else %}false{% endif %}`, "true")
	})

	// Ruby: assert_evaluates_false(1, '<>', 1)
	t.Run("1 <> 1 is false", func(t *testing.T) {
		check(`{% if 1 <> 1 %}true{% else %}false{% endif %}`, "false")
	})

	t.Run("string <> different string is true", func(t *testing.T) {
		check(`{% if "a" <> "b" %}true{% else %}false{% endif %}`, "true")
	})

	t.Run("string <> same string is false", func(t *testing.T) {
		check(`{% if "a" <> "a" %}true{% else %}false{% endif %}`, "false")
	})

	t.Run("1.0 <> 2.0 is true", func(t *testing.T) {
		check(`{% if 1.0 <> 2.0 %}true{% else %}false{% endif %}`, "true")
	})

	t.Run("1 <> 1.0 is false (cross-type)", func(t *testing.T) {
		check(`{% if 1 <> 1.0 %}true{% else %}false{% endif %}`, "false")
	})
}

// ── 4.6 `not` Unary Operator ─────────────────────────────────────────────────
// Source: LiquidJS src/render/expression.spec.ts

func TestPortedLiterals_NotOperator(t *testing.T) {
	eng := liquid.NewEngine()

	check := func(src string, want string) {
		t.Helper()
		require.Equal(t, want, renderExpr(t, eng, src, nil))
	}

	// LiquidJS: 'should support not'
	t.Run("not true is false", func(t *testing.T) {
		check(`{% if not true %}yes{% else %}no{% endif %}`, "no")
	})

	t.Run("not false is true", func(t *testing.T) {
		check(`{% if not false %}yes{% else %}no{% endif %}`, "yes")
	})

	t.Run("not nil is true", func(t *testing.T) {
		check(`{% if not nil %}yes{% else %}no{% endif %}`, "yes")
	})

	// LiquidJS: 'not should have higher precedence than and/or'
	// not 1 < 2 or not 1 > 2
	// = (not (1 < 2)) or (not (1 > 2))
	// = (not true) or (not false)
	// = false or true
	// = true
	t.Run("not has higher precedence than or", func(t *testing.T) {
		check(`{% if not 1 < 2 or not 1 > 2 %}true{% else %}false{% endif %}`, "true")
	})

	// not 1 < 2 and not 1 > 2
	// = (not true) and (not false)
	// = false and true
	// = false
	t.Run("not has higher precedence than and", func(t *testing.T) {
		check(`{% if not 1 < 2 and not 1 > 2 %}true{% else %}false{% endif %}`, "false")
	})

	// not applied to comparison expression
	t.Run("not applied to comparison", func(t *testing.T) {
		// LiquidJS: 'should support not' — not 1 < 2 → false
		check(`{% if not 1 < 2 %}true{% else %}false{% endif %}`, "false")
	})

	// not with and
	t.Run("not false and true is true", func(t *testing.T) {
		check(`{% if not false and true %}yes{% endif %}`, "yes")
	})
}

// ── 4.7 String Escape Sequences ──────────────────────────────────────────────
// Sources: LiquidJS src/render/string.spec.ts
//          Ruby   (implicit: strings with escapes evaluated correctly)

func TestPortedLiterals_StringEscapes(t *testing.T) {
	eng := liquid.NewEngine()

	check := func(src, want string) {
		t.Helper()
		require.Equal(t, want, renderExpr(t, eng, src, nil))
	}

	// LiquidJS: 'should parse \\n, \\t, \\r'
	t.Run("escape newline in double-quoted", func(t *testing.T) {
		check(`{{ "fo\no" }}`, "fo\no")
	})

	t.Run("escape tab in single-quoted", func(t *testing.T) {
		check(`{{ 'fo\to' }}`, "fo\to")
	})

	t.Run("escape carriage-return in single-quoted", func(t *testing.T) {
		check(`{{ 'fo\ro' }}`, "fo\ro")
	})

	// LiquidJS: 'should parse quote escape'
	t.Run("escape single quote in single-quoted string", func(t *testing.T) {
		check(`{{ 'fo\'o' }}`, "fo'o")
	})

	t.Run("escape double quote in double-quoted string", func(t *testing.T) {
		check(`{{ "fo\"o" }}`, `fo"o`)
	})

	// LiquidJS: 'should parse slash escape'
	t.Run("escape backslash", func(t *testing.T) {
		check(`{{ 'fo\\o' }}`, `fo\o`)
	})

	// String comparison after escaping — x must hold the unescaped value "a\b"
	t.Run("backslash escape in comparison", func(t *testing.T) {
		out := renderExpr(t, eng, `{% if "a\\b" == x %}yes{% endif %}`, map[string]any{"x": `a\b`})
		require.Equal(t, "yes", out)
	})
}

// ── 4.8 nil / null Comparisons ───────────────────────────────────────────────
// Source: Ruby test/integration/tags/statements_test.rb

func TestPortedLiterals_NilComparisons(t *testing.T) {
	eng := liquid.NewEngine()

	check := func(src string, bindings map[string]any, want string) {
		t.Helper()
		require.Equal(t, want, renderExpr(t, eng, src, bindings))
	}

	// Ruby: test_nil — var == nil when var is nil → true
	t.Run("nil variable equals nil", func(t *testing.T) {
		check(` {% if var == nil %} true {% else %} false {% endif %} `, map[string]any{"var": nil}, "  true  ")
	})

	// Ruby: test_nil — var == null (null is alias for nil) → true
	t.Run("nil variable equals null", func(t *testing.T) {
		check(` {% if var == null %} true {% else %} false {% endif %} `, map[string]any{"var": nil}, "  true  ")
	})

	// Ruby: test_not_nil — var != nil when var = 1 → true
	t.Run("non-nil variable not equals nil", func(t *testing.T) {
		check(` {% if var != nil %} true {% else %} false {% endif %} `, map[string]any{"var": 1}, "  true  ")
	})

	// Ruby: test_not_nil — var != null → true
	t.Run("non-nil variable not equals null", func(t *testing.T) {
		check(` {% if var != null %} true {% else %} false {% endif %} `, map[string]any{"var": 1}, "  true  ")
	})
}

// ── 4.9 Operator Precedence and Logic ────────────────────────────────────────
// Source: LiquidJS src/render/expression.spec.ts

func TestPortedLiterals_OperatorExpressions(t *testing.T) {
	eng := liquid.NewEngine()

	check := func(src string, bindings map[string]any, want string) {
		t.Helper()
		require.Equal(t, want, renderExpr(t, eng, src, bindings))
	}

	// JS: statements_test.rb — test_strings
	t.Run("string equality", func(t *testing.T) {
		check(` {% if 'test' == 'test' %} true {% else %} false {% endif %} `, nil, "  true  ")
	})

	t.Run("string inequality", func(t *testing.T) {
		check(` {% if 'test' != 'test' %} true {% else %} false {% endif %} `, nil, "  false  ")
	})

	// Ruby: test_var_strings_equal
	t.Run("variable string equality", func(t *testing.T) {
		check(` {% if var == "hello there!" %} true {% else %} false {% endif %} `, map[string]any{"var": "hello there!"}, "  true  ")
	})

	// LiquidJS: 'should recognize quoted value' — a string containing > is valid
	t.Run("quoted value with operator-like character", func(t *testing.T) {
		check(`{% if x == ">=" %}yes{% endif %}`, map[string]any{"x": ">="}, "yes")
	})

	// LiquidJS: 'should support value or value'
	t.Run("false or true is true", func(t *testing.T) {
		check(`{% if false or true %}yes{% endif %}`, nil, "yes")
	})

	// LiquidJS: 'should evaluate from right to left' (right-assoc)
	// true or false and false → true or (false and false) → true or false → true
	t.Run("right-to-left evaluation: or before and", func(t *testing.T) {
		check(`{% if true or false and false %}yes{% else %}no{% endif %}`, nil, "yes")
	})

	// true and false and false or true
	// Right-assoc: true and (false and (false or true)) = true and (false and true) = true and false = false
	t.Run("right-to-left evaluation: complex chain", func(t *testing.T) {
		check(`{% if true and false and false or true %}yes{% else %}no{% endif %}`, nil, "no")
	})

	// LiquidJS: 'should allow space in quoted value'
	t.Run("space in quoted value", func(t *testing.T) {
		check(`{% if " " == x %}yes{% endif %}`, map[string]any{"x": " "}, "yes")
	})

	// Ruby: test_is_collection_empty
	t.Run("empty array equals empty", func(t *testing.T) {
		check(` {% if array == empty %} true {% else %} false {% endif %} `, map[string]any{"array": []any{}}, "  true  ")
	})

	t.Run("non-empty array not equal empty", func(t *testing.T) {
		check(` {% if array == empty %} true {% else %} false {% endif %} `, map[string]any{"array": []any{1, 2, 3}}, "  false  ")
	})
}

// ── 4.10 Range `contains` operator ───────────────────────────────────────────
// Sources: LiquidJS src/render/expression.spec.ts ('should return true for "(1..5) contains 3"' etc.)
//          Ruby     test/unit/condition_unit_test.rb (contains checks, implicit range membership)

func TestPortedLiterals_RangeContains(t *testing.T) {
	eng := liquid.NewEngine()

	check := func(src string, bindings map[string]any, want string) {
		t.Helper()
		require.Equal(t, want, renderExpr(t, eng, src, bindings))
	}

	// LiquidJS: 'should return true for "(1..5) contains 3"'
	t.Run("range contains included value", func(t *testing.T) {
		check(`{% if (1..5) contains 3 %}yes{% else %}no{% endif %}`, nil, "yes")
	})

	// LiquidJS: 'should return false for "(1..5) contains 6"'
	t.Run("range does not contain out-of-range value", func(t *testing.T) {
		check(`{% if (1..5) contains 6 %}yes{% else %}no{% endif %}`, nil, "no")
	})

	// Range contains lower bound (inclusive)
	t.Run("range contains lower bound", func(t *testing.T) {
		check(`{% if (1..5) contains 1 %}yes{% else %}no{% endif %}`, nil, "yes")
	})

	// Range contains upper bound (inclusive)
	t.Run("range contains upper bound", func(t *testing.T) {
		check(`{% if (1..5) contains 5 %}yes{% else %}no{% endif %}`, nil, "yes")
	})

	// Range does not contain value just below lower bound
	t.Run("range does not contain value below lower bound", func(t *testing.T) {
		check(`{% if (1..5) contains 0 %}yes{% else %}no{% endif %}`, nil, "no")
	})

	// Range with variable end
	t.Run("range contains with variable bound", func(t *testing.T) {
		check(`{% if (1..n) contains 4 %}yes{% else %}no{% endif %}`, map[string]any{"n": 5}, "yes")
	})

	// Range in for loop with contains-like filtering — a for loop over range
	t.Run("range in for loop iterates all values", func(t *testing.T) {
		check(`{% for i in (1..3) %}{{ i }}{% endfor %}`, nil, "123")
	})
}

// ── 4.11 nil/null with ordering operators ─────────────────────────────────────
// Source: Ruby test/integration/tags/statements_test.rb (test_zero_lq_or_equal_one_involving_nil)

func TestPortedLiterals_NilOrdering(t *testing.T) {
	eng := liquid.NewEngine()

	check := func(src string, want string) {
		t.Helper()
		require.Equal(t, want, renderExpr(t, eng, src, nil))
	}

	// Ruby: test_zero_lq_or_equal_one_involving_nil
	// "{% if null <= 0 %} true {% else %} false {% endif %}" => "  false  "
	t.Run("null <= 0 is false", func(t *testing.T) {
		check(` {% if null <= 0 %} true {% else %} false {% endif %} `, "  false  ")
	})

	// Ruby: "{% if 0 <= null %} true {% else %} false {% endif %}" => "  false  "
	t.Run("0 <= null is false", func(t *testing.T) {
		check(` {% if 0 <= null %} true {% else %} false {% endif %} `, "  false  ")
	})

	// Nil with less-than ordering
	t.Run("null < 0 is false", func(t *testing.T) {
		check(` {% if null < 0 %} true {% else %} false {% endif %} `, "  false  ")
	})

	t.Run("0 < null is false", func(t *testing.T) {
		check(` {% if 0 < null %} true {% else %} false {% endif %} `, "  false  ")
	})

	// Nil with greater-than ordering
	t.Run("null > 0 is false", func(t *testing.T) {
		check(` {% if null > 0 %} true {% else %} false {% endif %} `, "  false  ")
	})

	t.Run("0 > null is false", func(t *testing.T) {
		check(` {% if 0 > null %} true {% else %} false {% endif %} `, "  false  ")
	})

	t.Run("null >= 0 is false", func(t *testing.T) {
		check(` {% if null >= 0 %} true {% else %} false {% endif %} `, "  false  ")
	})

	t.Run("0 >= null is false", func(t *testing.T) {
		check(` {% if 0 >= null %} true {% else %} false {% endif %} `, "  false  ")
	})

	// nil (Go keyword, same as null in Liquid)
	t.Run("nil <= 0 is false", func(t *testing.T) {
		check(` {% if nil <= 0 %} true {% else %} false {% endif %} `, "  false  ")
	})

	t.Run("0 <= nil is false", func(t *testing.T) {
		check(` {% if 0 <= nil %} true {% else %} false {% endif %} `, "  false  ")
	})
}
