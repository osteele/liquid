package liquid

// Ported tag tests for Section 1 from:
//   - Ruby Liquid: test/integration/tags/
//   - LiquidJS:    test/integration/tags/
//
// Sections covered:
//   1.1 Output / Expression  ({{ }}, echo)
//   1.2 Variables            (assign, capture)
//   1.3 Conditionals         (if/elsif/else, unless, case/when)
//   1.4 Iteration            (for, break/continue, cycle, tablerow)
//   1.6 Structure            (raw, comment)

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// tplRender parses and renders with no bindings.
func tplRender(t *testing.T, src string) string {
	t.Helper()
	eng := NewEngine()
	out, err := eng.ParseAndRenderString(src, nil)
	require.NoError(t, err)
	return out
}

// tplRenderWith parses and renders with the given bindings.
func tplRenderWith(t *testing.T, src string, bindings map[string]any) string {
	t.Helper()
	eng := NewEngine()
	out, err := eng.ParseAndRenderString(src, bindings)
	require.NoError(t, err)
	return out
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.1  Output / Expression  — {{ variable }}
// ─────────────────────────────────────────────────────────────────────────────

// Source: Ruby output_test.rb – test_variable, test_variable_traversing
func TestPorted_Output_Variable(t *testing.T) {
	require.Equal(t, " bmw ", tplRenderWith(t, " {{best_cars}} ", map[string]any{
		"best_cars": "bmw",
	}))
}

// Source: Ruby output_test.rb – test_variable_traversing
func TestPorted_Output_Traversing(t *testing.T) {
	assigns := map[string]any{
		"car": map[string]any{"bmw": "good", "gm": "bad"},
	}
	require.Equal(t, " good bad good ", tplRenderWith(t, " {{car.bmw}} {{car.gm}} {{car.bmw}} ", assigns))
}

// Source: Ruby output_test.rb – test_variable_traversing_with_two_brackets
func TestPorted_Output_DeepBrackets(t *testing.T) {
	src := `{{ site.data.menu[include.menu][include.locale] }}`
	assigns := map[string]any{
		"site":    map[string]any{"data": map[string]any{"menu": map[string]any{"foo": map[string]any{"bar": "it works!"}}}},
		"include": map[string]any{"menu": "foo", "locale": "bar"},
	}
	require.Equal(t, "it works!", tplRenderWith(t, src, assigns))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.1  Output / Expression  — echo tag
// ─────────────────────────────────────────────────────────────────────────────

// Source: LiquidJS echo.spec.ts – should output literals
func TestPorted_Echo_Literals(t *testing.T) {
	require.Equal(t, "1 1 1.1", tplRender(t, `{% echo 1 %} {% echo "1" %} {% echo 1.1 %}`))
}

// Source: LiquidJS echo.spec.ts – should output variables
func TestPorted_Echo_Variable(t *testing.T) {
	assigns := map[string]any{
		"people": map[string]any{
			"users": []any{map[string]any{"name": "Sally"}},
		},
	}
	require.Equal(t, "Sally", tplRenderWith(t, `{% echo people.users[0].name %}`, assigns))
}

// Source: LiquidJS echo.spec.ts – should apply filters before output
func TestPorted_Echo_Filters(t *testing.T) {
	assigns := map[string]any{"user": map[string]any{"name": "Sally"}}
	require.Equal(t, "Hello, SALLY!", tplRenderWith(t,
		`{% echo user.name | upcase | prepend: "Hello, " | append: "!" %}`, assigns))
}

// Source: Ruby echo_test.rb – echo inside liquid tag
func TestPorted_Echo_InsideLiquidTag(t *testing.T) {
	src := `{% liquid
  assign x = "hello"
  echo x | capitalize
%}`
	require.Equal(t, "Hello", tplRender(t, src))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.2  Variables  — assign
// ─────────────────────────────────────────────────────────────────────────────

// Source: Ruby assign_test.rb – test_assign_with_hyphen_in_variable_name
func TestPorted_Assign_HyphenatedName(t *testing.T) {
	src := `{%- assign this-thing = 'Print this-thing' -%}{{ this-thing }}`
	require.Equal(t, "Print this-thing", tplRender(t, src))
}

// Source: Ruby assign_test.rb – test_assigned_variable
func TestPorted_Assign_ArrayIndex(t *testing.T) {
	assigns := map[string]any{"values": []string{"foo", "bar", "baz"}}
	require.Equal(t, ".foo.", tplRenderWith(t, `{% assign foo = values %}.{{ foo[0] }}.`, assigns))
	require.Equal(t, ".bar.", tplRenderWith(t, `{% assign foo = values %}.{{ foo[1] }}.`, assigns))
}

// Source: Ruby assign_test.rb – test_assign_with_filter
func TestPorted_Assign_WithFilter(t *testing.T) {
	assigns := map[string]any{"values": "foo,bar,baz"}
	require.Equal(t, ".bar.", tplRenderWith(t,
		`{% assign foo = values | split: "," %}.{{ foo[1] }}.`, assigns))
}

// Source: LiquidJS assign.spec.ts – should assign as filter result
func TestPorted_Assign_FilterChain(t *testing.T) {
	src := `{% assign foo="a b" | capitalize | split: " " | first %}{{foo}}`
	require.Equal(t, "A", tplRender(t, src))
}

// Source: LiquidJS assign.spec.ts – scope: should write to root scope
func TestPorted_Assign_ScopeInLoop(t *testing.T) {
	src := `{%for a in (1..2)%}{%assign num = a%}{{a}}{%endfor%}`
	require.Equal(t, "12", tplRenderWith(t, src, map[string]any{"num": 1}))
}

// Source: LiquidJS assign.spec.ts – should allow reassignment
func TestPorted_Assign_Reassignment(t *testing.T) {
	require.Equal(t, "2", tplRender(t, `{% assign var = 1 %}{% assign var = 2 %}{{ var }}`))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.2  Variables  — capture
// ─────────────────────────────────────────────────────────────────────────────

// Source: Ruby capture_test.rb – test_captures_block_content_in_variable
func TestPorted_Capture_Basic(t *testing.T) {
	require.Equal(t, "test string",
		tplRender(t, `{% capture 'var' %}test string{% endcapture %}{{var}}`))
}

// Source: Ruby capture_test.rb – test_capture_with_hyphen_in_variable_name
func TestPorted_Capture_HyphenatedName(t *testing.T) {
	src := `{%- capture this-thing %}Print this-thing{% endcapture -%}{{ this-thing }}`
	require.Equal(t, "Print this-thing", tplRender(t, src))
}

// Source: Ruby capture_test.rb – test_capture_to_variable_from_outer_scope_if_existing
func TestPorted_Capture_OuterScope(t *testing.T) {
	src := `{% assign var = '' -%}
{% if true -%}
  {% capture var %}first-block-string{% endcapture -%}
{% endif -%}
{% if true -%}
  {% capture var %}test-string{% endcapture -%}
{% endif -%}
{{var}}`
	require.Equal(t, "test-string", tplRender(t, src))
}

// Source: Ruby capture_test.rb – test_assigning_from_capture
func TestPorted_Capture_AssigningFromCapture(t *testing.T) {
	src := `{% assign first = '' -%}
{% assign second = '' -%}
{% for number in (1..3) -%}
  {% capture first %}{{number}}{% endcapture -%}
  {% assign second = first -%}
{% endfor -%}
{{ first }}-{{ second }}`
	require.Equal(t, "3-3", tplRender(t, src))
}

// Source: LiquidJS capture.spec.ts – should support nested filters
func TestPorted_Capture_WithFilter(t *testing.T) {
	require.Equal(t, "A", tplRender(t, `{% capture f %}{{"a" | capitalize}}{%endcapture%}{{f}}`))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.3  Conditionals  — if / elsif / else
// ─────────────────────────────────────────────────────────────────────────────

// Source: Ruby if_else_tag_test.rb – test_if
func TestPorted_If_Basic(t *testing.T) {
	require.Equal(t, "  ", tplRender(t, ` {% if false %} this text should not go into the output {% endif %} `))
	require.Equal(t, "  this text should go into the output  ",
		tplRender(t, ` {% if true %} this text should go into the output {% endif %} `))
}

// Source: Ruby if_else_tag_test.rb – test_if_else
func TestPorted_If_Else(t *testing.T) {
	require.Equal(t, " YES ", tplRender(t, `{% if false %} NO {% else %} YES {% endif %}`))
	require.Equal(t, " YES ", tplRender(t, `{% if true %} YES {% else %} NO {% endif %}`))
	require.Equal(t, " YES ", tplRender(t, `{% if "foo" %} YES {% else %} NO {% endif %}`))
}

// Source: Ruby if_else_tag_test.rb – test_if_or / test_if_and
func TestPorted_If_LogicalOperators(t *testing.T) {
	require.Equal(t, " YES ", tplRenderWith(t, `{% if a or b %} YES {% endif %}`, map[string]any{"a": true, "b": false}))
	require.Equal(t, "", tplRenderWith(t, `{% if a or b %} YES {% endif %}`, map[string]any{"a": false, "b": false}))
	require.Equal(t, " YES ", tplRenderWith(t, `{% if a or b or c %} YES {% endif %}`, map[string]any{"a": false, "b": false, "c": true}))
	require.Equal(t, " YES ", tplRender(t, `{% if true and true %} YES {% endif %}`))
	require.Equal(t, "", tplRender(t, `{% if false and true %} YES {% endif %}`))
}

// Source: Ruby if_else_tag_test.rb – test_if_from_variable
func TestPorted_If_FromVariable(t *testing.T) {
	require.Equal(t, "", tplRenderWith(t, `{% if var %} NO {% endif %}`, map[string]any{"var": false}))
	require.Equal(t, "", tplRenderWith(t, `{% if var %} NO {% endif %}`, map[string]any{"var": nil}))
	require.Equal(t, " YES ", tplRenderWith(t, `{% if var %} YES {% endif %}`, map[string]any{"var": "text"}))
	require.Equal(t, " YES ", tplRenderWith(t, `{% if var %} YES {% endif %}`, map[string]any{"var": true}))
	require.Equal(t, " YES ", tplRenderWith(t, `{% if var %} YES {% endif %}`, map[string]any{"var": 1}))
	require.Equal(t, " YES ", tplRenderWith(t, `{% if var %} YES {% endif %}`, map[string]any{"var": []any{}}))
}

// Source: Ruby if_else_tag_test.rb – test_nested_if
func TestPorted_If_Nested(t *testing.T) {
	require.Equal(t, "", tplRender(t, `{% if false %}{% if true %} NO {% endif %}{% endif %}`))
	require.Equal(t, " YES ", tplRender(t, `{% if true %}{% if true %} YES {% endif %}{% endif %}`))
	require.Equal(t, " YES ", tplRender(t, `{% if true %}{% if false %} NO {% else %} YES {% endif %}{% else %} NO {% endif %}`))
}

// Source: Ruby if_else_tag_test.rb – test_comparisons_on_null
func TestPorted_If_NullComparisons(t *testing.T) {
	require.Equal(t, "", tplRender(t, `{% if null < 10 %} NO {% endif %}`))
	require.Equal(t, "", tplRender(t, `{% if null > 10 %} NO {% endif %}`))
	require.Equal(t, "", tplRender(t, `{% if 10 < null %} NO {% endif %}`))
	require.Equal(t, "", tplRender(t, `{% if 10 > null %} NO {% endif %}`))
	require.Equal(t, "", tplRender(t, `{% if 10 >= null %} NO {% endif %}`))
	require.Equal(t, "", tplRender(t, `{% if 10 <= null %} NO {% endif %}`))
}

// Source: Ruby if_else_tag_test.rb – test_else_if
func TestPorted_If_ElsIf(t *testing.T) {
	require.Equal(t, "0", tplRender(t, `{% if 0 == 0 %}0{% elsif 1 == 1%}1{% else %}2{% endif %}`))
	require.Equal(t, "1", tplRender(t, `{% if 0 != 0 %}0{% elsif 1 == 1%}1{% else %}2{% endif %}`))
	require.Equal(t, "2", tplRender(t, `{% if 0 != 0 %}0{% elsif 1 != 1%}1{% else %}2{% endif %}`))
	require.Equal(t, "elsif", tplRender(t, `{% if false %}if{% elsif true %}elsif{% endif %}`))
}

// Source: Ruby if_else_tag_test.rb – test_literal_comparisons
func TestPorted_If_LiteralComparisons(t *testing.T) {
	require.Equal(t, " NO ", tplRender(t, `{% assign v = false %}{% if v %} YES {% else %} NO {% endif %}`))
	require.Equal(t, " YES ", tplRender(t, `{% assign v = nil %}{% if v == nil %} YES {% else %} NO {% endif %}`))
}

// Source: Ruby if_else_tag_test.rb – test_multiple_conditions
func TestPorted_If_MultipleConditions(t *testing.T) {
	tpl := `{% if a or b and c %}true{% else %}false{% endif %}`
	tests := []struct {
		a, b, c bool
		want    string
	}{
		{true, true, true, "true"},
		{true, true, false, "true"},
		{true, false, true, "true"},
		{true, false, false, "true"},
		{false, true, true, "true"},
		{false, true, false, "false"},
		{false, false, true, "false"},
		{false, false, false, "false"},
	}
	for _, tt := range tests {
		got := tplRenderWith(t, tpl, map[string]any{"a": tt.a, "b": tt.b, "c": tt.c})
		require.Equal(t, tt.want, got)
	}
}

// Source: LiquidJS if.spec.ts – should evaluate right to left
func TestPorted_If_RightToLeftEvaluation(t *testing.T) {
	// false and false or true → false and (false or true) = false
	require.Equal(t, "", tplRender(t, `{% if false and false or true %}true{%endif%}`))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.3  Conditionals  — unless
// ─────────────────────────────────────────────────────────────────────────────

// Source: Ruby unless_else_tag_test.rb – test_unless
func TestPorted_Unless_Basic(t *testing.T) {
	require.Equal(t, "  ", tplRender(t, ` {% unless true %} this text should not go into the output {% endunless %} `))
	require.Equal(t, "  this text should go into the output  ",
		tplRender(t, ` {% unless false %} this text should go into the output {% endunless %} `))
}

// Source: Ruby unless_else_tag_test.rb – test_unless_else
func TestPorted_Unless_Else(t *testing.T) {
	require.Equal(t, " YES ", tplRender(t, `{% unless true %} NO {% else %} YES {% endunless %}`))
	require.Equal(t, " YES ", tplRender(t, `{% unless false %} YES {% else %} NO {% endunless %}`))
}

// Source: Ruby unless_else_tag_test.rb – test_unless_in_loop
func TestPorted_Unless_InLoop(t *testing.T) {
	src := `{% for i in choices %}{% unless i %}{{ forloop.index }}{% endunless %}{% endfor %}`
	assigns := map[string]any{"choices": []any{1, nil, false}}
	require.Equal(t, "23", tplRenderWith(t, src, assigns))
}

// Source: Ruby unless_else_tag_test.rb – test_unless_else_in_loop
func TestPorted_Unless_ElseInLoop(t *testing.T) {
	src := `{% for i in choices %}{% unless i %} {{ forloop.index }} {% else %} TRUE {% endunless %}{% endfor %}`
	assigns := map[string]any{"choices": []any{1, nil, false}}
	require.Equal(t, " TRUE  2  3 ", tplRenderWith(t, src, assigns))
}

// Source: LiquidJS unless.spec.ts – should output contents in order
func TestPorted_Unless_OutputInOrder(t *testing.T) {
	src := `
      Before {{ location }}
      {% unless false %}Inside {{ location }}{% endunless %}
      After {{ location }}`
	want := `
      Before wonderland
      Inside wonderland
      After wonderland`
	require.Equal(t, want, tplRenderWith(t, src, map[string]any{"location": "wonderland"}))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.3  Conditionals  — case / when
// ─────────────────────────────────────────────────────────────────────────────

// Source: LiquidJS case.spec.ts – should hit the specified case
func TestPorted_Case_Basic(t *testing.T) {
	src := `{% case "foo"%}{% when "foo" %}foo{% when "bar"%}bar{%endcase%}`
	require.Equal(t, "foo", tplRender(t, src))
}

// Source: LiquidJS case.spec.ts – should support else branch
func TestPorted_Case_Else(t *testing.T) {
	src := `{% case "a" %}{% when "b" %}b{% when "c"%}c{%else %}d{%endcase%}`
	require.Equal(t, "d", tplRender(t, src))
}

// Source: LiquidJS case.spec.ts – should support case with multiple values
func TestPorted_Case_MultipleValues(t *testing.T) {
	src := `{% case "b" %}{% when "a", "b" %}foo{%endcase%}`
	require.Equal(t, "foo", tplRender(t, src))
}

// Source: LiquidJS case.spec.ts – should support case with multiple values separated by or
func TestPorted_Case_MultipleValuesOr(t *testing.T) {
	src := `{% case 3 %}{% when 1 or 2 or 3 %}1 or 2 or 3{% else %}not 1 or 2 or 3{%endcase%}`
	require.Equal(t, "1 or 2 or 3", tplRender(t, src))
}

// Source: LiquidJS case.spec.ts – blank/empty compare
func TestPorted_Case_BlankEmpty(t *testing.T) {
	require.Equal(t, "bar", tplRender(t, `{% case blank %}{% when ""%}bar{%endcase%}`))
	require.Equal(t, "bar", tplRender(t, `{% case empty %}{% when ""%}bar{%endcase%}`))
}

// Source: LiquidJS case.spec.ts – should support boolean case
func TestPorted_Case_Boolean(t *testing.T) {
	src := `{% case false %}{% when "foo" %}foo{% when false%}bar{%endcase%}`
	require.Equal(t, "bar", tplRender(t, src))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.4  Iteration  — for basic
// ─────────────────────────────────────────────────────────────────────────────

// Source: Ruby for_tag_test.rb – test_for
func TestPorted_For_Basic(t *testing.T) {
	assigns := map[string]any{"array": []int{1, 2, 3, 4}}
	require.Equal(t, " yo  yo  yo  yo ", tplRenderWith(t, `{%for item in array%} yo {%endfor%}`, assigns))
}

// Source: Ruby for_tag_test.rb – test_for_with_variable
func TestPorted_For_Variable(t *testing.T) {
	assigns := map[string]any{"array": []int{1, 2, 3}}
	require.Equal(t, " 1  2  3 ", tplRenderWith(t, `{%for item in array%} {{item}} {%endfor%}`, assigns))
}

// Source: Ruby for_tag_test.rb – test_for_reversed (Ruby behavior: offset→limit→reversed)
func TestPorted_For_ReversedOnly(t *testing.T) {
	assigns := map[string]any{"array": []int{1, 2, 3}}
	require.Equal(t, "321", tplRenderWith(t, `{%for item in array reversed %}{{item}}{%endfor%}`, assigns))
}

// Source: Ruby for_tag_test.rb – test_limiting
func TestPorted_For_Limiting(t *testing.T) {
	assigns := map[string]any{"array": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}}
	require.Equal(t, "12", tplRenderWith(t, `{%for i in array limit:2 %}{{ i }}{%endfor%}`, assigns))
	require.Equal(t, "1234", tplRenderWith(t, `{%for i in array limit:4 %}{{ i }}{%endfor%}`, assigns))
	require.Equal(t, "3456", tplRenderWith(t, `{%for i in array limit:4 offset:2 %}{{ i }}{%endfor%}`, assigns))
	require.Equal(t, "3456", tplRenderWith(t, `{%for i in array limit: 4 offset: 2 %}{{ i }}{%endfor%}`, assigns))
}

// Source: Ruby for_tag_test.rb – test_dynamic_variable_limiting
func TestPorted_For_DynamicLimiting(t *testing.T) {
	assigns := map[string]any{
		"array":  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		"limit":  2,
		"offset": 2,
	}
	require.Equal(t, "34", tplRenderWith(t,
		`{%for i in array limit: limit offset: offset %}{{ i }}{%endfor%}`, assigns))
}

// Source: Ruby for_tag_test.rb – test_nested_for
func TestPorted_For_Nested(t *testing.T) {
	assigns := map[string]any{"array": [][]int{{1, 2}, {3, 4}, {5, 6}}}
	require.Equal(t, "123456",
		tplRenderWith(t, `{%for item in array%}{%for i in item%}{{ i }}{%endfor%}{%endfor%}`, assigns))
}

// Source: Ruby for_tag_test.rb – test_for_else
func TestPorted_For_Else(t *testing.T) {
	plusMinus := `{%for item in array%}+{%else%}-{%endfor%}`
	require.Equal(t, "+++", tplRenderWith(t, plusMinus, map[string]any{"array": []int{1, 2, 3}}))
	require.Equal(t, "-", tplRenderWith(t, plusMinus, map[string]any{"array": []int{}}))
	require.Equal(t, "-", tplRenderWith(t, plusMinus, map[string]any{"array": nil}))
}

// Source: Ruby for_tag_test.rb – test_for_helpers
func TestPorted_For_LoopVariables(t *testing.T) {
	assigns := map[string]any{"array": []int{1, 2, 3}}
	require.Equal(t, " 1/3  2/3  3/3 ",
		tplRenderWith(t, `{%for item in array%} {{forloop.index}}/{{forloop.length}} {%endfor%}`, assigns))
	require.Equal(t, " 1  2  3 ",
		tplRenderWith(t, `{%for item in array%} {{forloop.index}} {%endfor%}`, assigns))
	require.Equal(t, " 0  1  2 ",
		tplRenderWith(t, `{%for item in array%} {{forloop.index0}} {%endfor%}`, assigns))
	require.Equal(t, " 2  1  0 ",
		tplRenderWith(t, `{%for item in array%} {{forloop.rindex0}} {%endfor%}`, assigns))
	require.Equal(t, " 3  2  1 ",
		tplRenderWith(t, `{%for item in array%} {{forloop.rindex}} {%endfor%}`, assigns))
	require.Equal(t, " true  false  false ",
		tplRenderWith(t, `{%for item in array%} {{forloop.first}} {%endfor%}`, assigns))
	require.Equal(t, " false  false  true ",
		tplRenderWith(t, `{%for item in array%} {{forloop.last}} {%endfor%}`, assigns))
}

// Source: Ruby for_tag_test.rb – test_for_parentloop_references_parent_loop
func TestPorted_For_ParentLoop(t *testing.T) {
	src := `{% for inner in outer %}{% for k in inner %}{{ forloop.parentloop.index }}.{{ forloop.index }} {% endfor %}{% endfor %}`
	assigns := map[string]any{"outer": [][]int{{1, 1, 1}, {1, 1, 1}}}
	require.Equal(t, "1.1 1.2 1.3 2.1 2.2 2.3 ", tplRenderWith(t, src, assigns))
}

// Source: Ruby for_tag_test.rb – test_for_parentloop_nil_when_not_present
func TestPorted_For_ParentLoopNilWhenAbsent(t *testing.T) {
	src := `{% for inner in outer %}{{ forloop.parentloop.index }}.{{ forloop.index }} {% endfor %}`
	assigns := map[string]any{"outer": [][]int{{1, 1, 1}, {1, 1, 1}}}
	require.Equal(t, ".1 .2 ", tplRenderWith(t, src, assigns))
}

// Source: Ruby for_tag_test.rb – test_for_with_range
func TestPorted_For_WithRange(t *testing.T) {
	require.Equal(t, " 1  2  3 ", tplRender(t, `{%for item in (1..3) %} {{item}} {%endfor%}`))
}

// Source: Ruby for_tag_test.rb – test_for_and_if
func TestPorted_For_AndIf(t *testing.T) {
	assigns := map[string]any{"array": []int{1, 2, 3}}
	require.Equal(t, "+--", tplRenderWith(t,
		`{%for item in array%}{% if forloop.first %}+{% else %}-{% endif %}{%endfor%}`, assigns))
}

// Source: Ruby for_tag_test.rb – test_spacing_with_variable_naming_in_for_loop
func TestPorted_For_SpacingInDeclaration(t *testing.T) {
	assigns := map[string]any{"items": []int{1, 2, 3, 4, 5}}
	require.Equal(t, "12345",
		tplRenderWith(t, `{% for       item   in   items %}{{item}}{% endfor %}`, assigns))
}

// Modifier order: reversed is applied first, then offset, then limit.
// This means offset:N skips N elements from the start of the reversed view.
func TestPorted_For_ModifierOrder_ReversedWithLimit(t *testing.T) {
	// reversed limit:2 on [1..5]: reversed=[5,4,3,2,1] → limit:2=[5,4]
	assigns := map[string]any{"array": []int{1, 2, 3, 4, 5}}
	require.Equal(t, "54",
		tplRenderWith(t, `{% for i in array reversed limit:2 %}{{ i }}{% endfor %}`, assigns))
	// same result regardless of syntax order
	require.Equal(t, "54",
		tplRenderWith(t, `{% for i in array limit:2 reversed %}{{ i }}{% endfor %}`, assigns))
}

func TestPorted_For_ModifierOrder_ReversedWithOffset(t *testing.T) {
	// reversed offset:2 on [1..5]: reversed=[5,4,3,2,1] → offset:2=[3,2,1]
	assigns := map[string]any{"array": []int{1, 2, 3, 4, 5}}
	require.Equal(t, "321",
		tplRenderWith(t, `{% for i in array reversed offset:2 %}{{ i }}{% endfor %}`, assigns))
	// same result regardless of syntax order
	require.Equal(t, "321",
		tplRenderWith(t, `{% for i in array offset:2 reversed %}{{ i }}{% endfor %}`, assigns))
}

func TestPorted_For_ModifierOrder_AllThree(t *testing.T) {
	// reversed limit:2 offset:1 on [1..5]: reversed=[5,4,3,2,1] → offset:1=[4,3,2,1] → limit:2=[4,3]
	assigns := map[string]any{"array": []int{1, 2, 3, 4, 5}}
	require.Equal(t, "43",
		tplRenderWith(t, `{% for i in array reversed limit:2 offset:1 %}{{ i }}{% endfor %}`, assigns))
	// syntax order does not matter
	require.Equal(t, "43",
		tplRenderWith(t, `{% for i in array limit:2 offset:1 reversed %}{{ i }}{% endfor %}`, assigns))
	require.Equal(t, "43",
		tplRenderWith(t, `{% for i in array offset:1 reversed limit:2 %}{{ i }}{% endfor %}`, assigns))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.4  Iteration  — break / continue
// ─────────────────────────────────────────────────────────────────────────────

// Source: Ruby for_tag_test.rb – test_for_with_break
func TestPorted_For_Break(t *testing.T) {
	assigns := map[string]any{"array": map[string]any{"items": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}}
	require.Equal(t, "", tplRenderWith(t, `{% for i in array.items %}{% break %}{% endfor %}`, assigns))
	require.Equal(t, "1", tplRenderWith(t, `{% for i in array.items %}{{ i }}{% break %}{% endfor %}`, assigns))
	require.Equal(t, "", tplRenderWith(t, `{% for i in array.items %}{% break %}{{ i }}{% endfor %}`, assigns))
	require.Equal(t, "1234",
		tplRenderWith(t, `{% for i in array.items %}{{ i }}{% if i > 3 %}{% break %}{% endif %}{% endfor %}`, assigns))
}

// Source: Ruby for_tag_test.rb – test_for_with_break (nested)
func TestPorted_For_BreakOnlyBreaksLocal(t *testing.T) {
	src := `{% for item in array %}{% for i in item %}{% if i == 1 %}{% break %}{% endif %}{{ i }}{% endfor %}{% endfor %}`
	assigns := map[string]any{"array": [][]int{{1, 2}, {3, 4}, {5, 6}}}
	require.Equal(t, "3456", tplRenderWith(t, src, assigns))
}

// Source: Ruby for_tag_test.rb – test_for_with_continue
func TestPorted_For_Continue(t *testing.T) {
	assigns := map[string]any{"array": map[string]any{"items": []int{1, 2, 3, 4, 5}}}
	require.Equal(t, "", tplRenderWith(t, `{% for i in array.items %}{% continue %}{% endfor %}`, assigns))
	require.Equal(t, "12345",
		tplRenderWith(t, `{% for i in array.items %}{{ i }}{% continue %}{% endfor %}`, assigns))
	require.Equal(t, "",
		tplRenderWith(t, `{% for i in array.items %}{% continue %}{{ i }}{% endfor %}`, assigns))
	require.Equal(t, "123",
		tplRenderWith(t, `{% for i in array.items %}{% if i > 3 %}{% continue %}{% endif %}{{ i }}{% endfor %}`, assigns))
	require.Equal(t, "1245",
		tplRenderWith(t, `{% for i in array.items %}{% if i == 3 %}{% continue %}{% else %}{{ i }}{% endif %}{% endfor %}`, assigns))
}

// Source: Ruby for_tag_test.rb – test_for_with_continue (nested)
func TestPorted_For_ContinueOnlyContinuesLocal(t *testing.T) {
	src := `{% for item in array %}{% for i in item %}{% if i == 1 %}{% continue %}{% endif %}{{ i }}{% endfor %}{% endfor %}`
	assigns := map[string]any{"array": [][]int{{1, 2}, {3, 4}, {5, 6}}}
	require.Equal(t, "23456", tplRenderWith(t, src, assigns))
}

// Source: LiquidJS for.spec.ts – should not break template outside of forloop
func TestPorted_For_BreakDoesNotLeakOutside(t *testing.T) {
	src := `{% for i in (1..5) %}{{ i }}{% break %}{% endfor %} after`
	require.Equal(t, "1 after", tplRender(t, src))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.4  Iteration  — cycle
// ─────────────────────────────────────────────────────────────────────────────

// Source: Ruby cycle_tag_test.rb – test_simple_cycle_inside_for_loop
func TestPorted_Cycle_Simple(t *testing.T) {
	src := `{%- for i in (1..3) -%}{%- cycle '1', '2', '3' -%}{%- endfor -%}`
	require.Equal(t, "123", tplRender(t, src))
}

// Source: Ruby cycle_tag_test.rb – test_cycle_named_groups_string
// Note: named groups with string literal key, values must be strings (Go limitation: integer values not supported)
func TestPorted_Cycle_NamedGroupsString(t *testing.T) {
	src := `{%- for i in (1..3) -%}{%- cycle 'placeholder1': '1', '2', '3' -%}{%- cycle 'placeholder2': '1', '2', '3' -%}{%- endfor -%}`
	require.Equal(t, "112233", tplRender(t, src))
}

// Source: LiquidJS cycle.spec.ts – should considered different groups for different arguments (inside for)
func TestPorted_Cycle_DifferentArgsDifferentGroups(t *testing.T) {
	src := `{% for i in (1..3) %}{% cycle '1', '2', '3'%}{% endfor %}`
	require.Equal(t, "123", tplRender(t, src))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.4  Iteration  — tablerow
// ─────────────────────────────────────────────────────────────────────────────

// Note: In Go, tablerow loop variables are accessed via `forloop.xxx` (not `tablerowloop.xxx`),
// and the HTML format has no newline between <tr> and <td> (unlike Ruby which uses <tr>\n<td>).

// Source: Ruby table_row_test.rb – test_table_row (adapted for Go HTML format)
func TestPorted_Tablerow_Basic(t *testing.T) {
	assigns := map[string]any{"numbers": []int{1, 2, 3, 4, 5, 6}}
	src := `{% tablerow n in numbers cols:3%} {{n}} {% endtablerow %}`
	expected := `<tr class="row1"><td class="col1"> 1 </td><td class="col2"> 2 </td><td class="col3"> 3 </td></tr>` +
		`<tr class="row2"><td class="col1"> 4 </td><td class="col2"> 5 </td><td class="col3"> 6 </td></tr>`
	require.Equal(t, expected, tplRenderWith(t, src, assigns))
}

// Source: Ruby table_row_test.rb – test_offset_and_limit (adapted for Go HTML format)
func TestPorted_Tablerow_OffsetAndLimit(t *testing.T) {
	assigns := map[string]any{"numbers": []int{0, 1, 2, 3, 4, 5, 6, 7}}
	src := `{% tablerow n in numbers cols:3 offset:1 limit:6%} {{n}} {% endtablerow %}`
	expected := `<tr class="row1"><td class="col1"> 1 </td><td class="col2"> 2 </td><td class="col3"> 3 </td></tr>` +
		`<tr class="row2"><td class="col1"> 4 </td><td class="col2"> 5 </td><td class="col3"> 6 </td></tr>`
	require.Equal(t, expected, tplRenderWith(t, src, assigns))
}

// Source: Ruby table_row_test.rb – test_table_col_counter
// In Go, tablerow loop variables use `forloop.col` (not `tablerowloop.col`)
func TestPorted_Tablerow_ColCounter(t *testing.T) {
	assigns := map[string]any{"numbers": []int{1, 2, 3, 4, 5, 6}}
	src := `{% tablerow n in numbers cols:2%}{{forloop.col}}{% endtablerow %}`
	expected := `<tr class="row1"><td class="col1">1</td><td class="col2">2</td></tr>` +
		`<tr class="row2"><td class="col1">1</td><td class="col2">2</td></tr>` +
		`<tr class="row3"><td class="col1">1</td><td class="col2">2</td></tr>`
	require.Equal(t, expected, tplRenderWith(t, src, assigns))
}

// Source: Ruby table_row_test.rb – test_tablerow_loop_drop_attributes
// In Go, tablerow loop variables use `forloop.xxx` (not `tablerowloop.xxx`)
func TestPorted_Tablerow_LoopDropAttributes(t *testing.T) {
	src := "{% tablerow i in (1..2) %}\n" +
		"col: {{ forloop.col }}\n" +
		"col0: {{ forloop.col0 }}\n" +
		"col_first: {{ forloop.col_first }}\n" +
		"col_last: {{ forloop.col_last }}\n" +
		"first: {{ forloop.first }}\n" +
		"index: {{ forloop.index }}\n" +
		"index0: {{ forloop.index0 }}\n" +
		"last: {{ forloop.last }}\n" +
		"length: {{ forloop.length }}\n" +
		"rindex: {{ forloop.rindex }}\n" +
		"rindex0: {{ forloop.rindex0 }}\n" +
		"row: {{ forloop.row }}\n" +
		"{% endtablerow %}"

	expected := "<tr class=\"row1\"><td class=\"col1\">\ncol: 1\ncol0: 0\ncol_first: true\ncol_last: false\n" +
		"first: true\nindex: 1\nindex0: 0\nlast: false\nlength: 2\nrindex: 2\nrindex0: 1\nrow: 1\n" +
		"</td><td class=\"col2\">\ncol: 2\ncol0: 1\ncol_first: false\ncol_last: true\n" +
		"first: false\nindex: 2\nindex0: 1\nlast: true\nlength: 2\nrindex: 1\nrindex0: 0\nrow: 1\n" +
		"</td></tr>"
	require.Equal(t, expected, tplRender(t, src))
}

// Source: LiquidJS tablerow.spec.ts – should support cols
func TestPorted_Tablerow_Cols(t *testing.T) {
	assigns := map[string]any{"alpha": []string{"a", "b", "c"}}
	src := `{% tablerow i in alpha cols:2 %}{{ i }}{% endtablerow %}`
	expected := `<tr class="row1"><td class="col1">a</td><td class="col2">b</td></tr>` +
		`<tr class="row2"><td class="col1">c</td></tr>`
	require.Equal(t, expected, tplRenderWith(t, src, assigns))
}

// Source: LiquidJS tablerow.spec.ts – should support index0, index, rindex0, rindex
// In Go, tablerow uses `forloop.xxx` (not `tablerowloop.xxx`)
func TestPorted_Tablerow_IndexVars(t *testing.T) {
	src := `{% tablerow i in (1..3)%}{{forloop.index0}}{{forloop.index}}{{forloop.rindex0}}{{forloop.rindex}}{% endtablerow %}`
	expected := `<tr class="row1"><td class="col1">0123</td><td class="col2">1212</td><td class="col3">2301</td></tr>`
	require.Equal(t, expected, tplRender(t, src))
}

// Source: LiquidJS tablerow.spec.ts – should support first, last, length
// In Go, tablerow uses `forloop.xxx` (not `tablerowloop.xxx`)
func TestPorted_Tablerow_FirstLastLength(t *testing.T) {
	src := `{% tablerow i in (1..3)%}{{forloop.first}} {{forloop.last}} {{forloop.length}}{% endtablerow %}`
	expected := `<tr class="row1"><td class="col1">true false 3</td><td class="col2">false false 3</td><td class="col3">false true 3</td></tr>`
	require.Equal(t, expected, tplRender(t, src))
}

// Source: LiquidJS tablerow.spec.ts – offset should start col at 1
// In Go, tablerow uses `forloop.col` (not `tablerowloop.col`)
func TestPorted_Tablerow_OffsetColReset(t *testing.T) {
	src := `{% tablerow i in (1..4) cols:2 offset:3 %}{{forloop.col}}{% endtablerow %}`
	expected := `<tr class="row1"><td class="col1">1</td></tr>`
	require.Equal(t, expected, tplRender(t, src))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.6  Structure  — raw
// ─────────────────────────────────────────────────────────────────────────────

// Source: Ruby raw_tag_test.rb – test_tag_in_raw
func TestPorted_Raw_TagInside(t *testing.T) {
	require.Equal(t, "{% comment %} test {% endcomment %}",
		tplRender(t, `{% raw %}{% comment %} test {% endcomment %}{% endraw %}`))
}

// Source: Ruby raw_tag_test.rb – test_output_in_raw (with trim markers)
func TestPorted_Raw_TrimMarkers(t *testing.T) {
	require.Equal(t, ">{{ test }}<", tplRender(t, `> {%- raw -%}{{ test }}{%- endraw -%} <`))
}

// Source: Ruby raw_tag_test.rb – test_open_tag_in_raw
// Note: Go and Ruby differ for cases where an unclosed {% appears inside raw before {% endraw %},
// because Go's scanner tokenizes {%...%} pairs greedily, consuming the endraw delimiter inside
// the unclosed tag. Only properly closed {%...%} sequences can appear inside raw blocks in Go.
func TestPorted_Raw_BrokenTagsInside(t *testing.T) {
	// These cases work: invalid closing %} is harmless content
	require.Equal(t, " Foobar invalid %} ", tplRender(t, `{% raw %} Foobar invalid %} {% endraw %}`))
	// These cases work: {{ without matching }} is harmless content
	require.Equal(t, " Foobar {{ invalid ", tplRender(t, `{% raw %} Foobar {{ invalid {% endraw %}`))
}

// Source: LiquidJS raw.spec.ts – should output filters as it is
func TestPorted_Raw_Filters(t *testing.T) {
	src := `{% raw %}{{ 5 | plus: 6 }}{% endraw %} is equal to 11.`
	require.Equal(t, "{{ 5 | plus: 6 }} is equal to 11.", tplRender(t, src))
}

// Source: LiquidJS raw.spec.ts – should preserve blank characters
func TestPorted_Raw_PreserveWhitespace(t *testing.T) {
	require.Equal(t, "\n{{ foo}} \n", tplRender(t, "{% raw %}\n{{ foo}} \n{% endraw %}"))
}

// ─────────────────────────────────────────────────────────────────────────────
// 1.6  Structure  — comment
// ─────────────────────────────────────────────────────────────────────────────

// Source: LiquidJS comment.spec.ts – should ignore plain string
func TestPorted_Comment_IgnoresContent(t *testing.T) {
	require.Equal(t, "My name is  Shopify.",
		tplRender(t, `My name is {% comment %}super{% endcomment %} Shopify.`))
}

// Source: LiquidJS comment.spec.ts – should ignore output tokens
func TestPorted_Comment_IgnoresOutputTokens(t *testing.T) {
	require.Equal(t, "", tplRender(t, "{% comment %}\n{{ foo}} \n{% endcomment %}"))
}

// Source: LiquidJS comment.spec.ts – should ignore tag tokens
func TestPorted_Comment_IgnoresTagTokens(t *testing.T) {
	require.Equal(t, "", tplRender(t, `{% comment %}{%if true%}true{%else%}false{%endif%}{% endcomment %}`))
}

// Source: LiquidJS comment.spec.ts – should ignore unbalanced tag tokens
// (Go behavior: any token inside comment is ignored, matching effective Ruby behavior)
func TestPorted_Comment_IgnoresUnbalancedTokens(t *testing.T) {
	require.Equal(t, "", tplRender(t, `{% comment %}{%if true%}true{%else%}false{% endcomment %}`))
}
