package liquid

// Ported analysis tests from:
//   - Ruby Liquid: test/unit/parse_tree_visitor_test.rb
//   - LiquidJS:    test/integration/static_analysis/variables.spec.ts
//   - LiquidJS:    test/e2e/parse-and-analyze.spec.ts
//   - LiquidJS:    src/template/analysis.spec.ts

import (
	"fmt"
	"strings"
	"testing"
)

// ── Ruby Liquid: ParseTreeVisitor tests ──────────────────────────────────────
// Source: test/unit/parse_tree_visitor_test.rb
// These tests validate that every tag/expression type correctly reports its
// referenced variables through static analysis.

func TestRubyLiquid_ParseTreeVisitor(t *testing.T) {
	engine := NewEngine()

	// helper: parse, analyze globals, return flat root names
	globalRootNames := func(t *testing.T, src string) []string {
		t.Helper()
		tpl, parseErr := engine.ParseString(src)
		if parseErr != nil {
			t.Fatalf("ParseString(%q): %v", src, parseErr)
		}
		names, analyzeErr := engine.GlobalVariables(tpl)
		if analyzeErr != nil {
			t.Fatalf("GlobalVariables: %v", analyzeErr)
		}
		return names
	}

	// helper: parse, analyze all variable segments
	allSegments := func(t *testing.T, src string) [][]string {
		t.Helper()
		tpl, parseErr := engine.ParseString(src)
		if parseErr != nil {
			t.Fatalf("ParseString(%q): %v", src, parseErr)
		}
		segs, analyzeErr := engine.VariableSegments(tpl)
		if analyzeErr != nil {
			t.Fatalf("VariableSegments: %v", analyzeErr)
		}
		return segs
	}

	// helper: parse, analyze global segments
	globalSegments := func(t *testing.T, src string) [][]string {
		t.Helper()
		tpl, parseErr := engine.ParseString(src)
		if parseErr != nil {
			t.Fatalf("ParseString(%q): %v", src, parseErr)
		}
		segs, analyzeErr := engine.GlobalVariableSegments(tpl)
		if analyzeErr != nil {
			t.Fatalf("GlobalVariableSegments: %v", analyzeErr)
		}
		return segs
	}

	t.Run("variable", func(t *testing.T) {
		got := globalSegments(t, `{{ test }}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("variable with filter", func(t *testing.T) {
		// Ruby: test_varible_with_filter — "test" and "infilter" both detected
		got := globalSegments(t, `{{ test | split: infilter }}`)
		want := [][]string{{"test"}, {"infilter"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("if condition", func(t *testing.T) {
		got := globalSegments(t, `{% if test %}{% endif %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("complex if condition", func(t *testing.T) {
		// Ruby: test_complex_if_condition — only "test" is variable; 1, 2 are literals
		got := globalSegments(t, `{% if 1 == 1 and 2 == test %}{% endif %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("if body", func(t *testing.T) {
		got := globalSegments(t, `{% if 1 == 1 %}{{ test }}{% endif %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unless condition", func(t *testing.T) {
		got := globalSegments(t, `{% unless test %}{% endunless %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("complex unless condition", func(t *testing.T) {
		got := globalSegments(t, `{% unless 1 == 1 and 2 == test %}{% endunless %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unless body", func(t *testing.T) {
		got := globalSegments(t, `{% unless 1 == 1 %}{{ test }}{% endunless %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("elsif condition", func(t *testing.T) {
		got := globalSegments(t, `{% if 1 == 1 %}{% elsif test %}{% endif %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("complex elsif condition", func(t *testing.T) {
		got := globalSegments(t, `{% if 1 == 1 %}{% elsif 1 == 1 and 2 == test %}{% endif %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("elsif body", func(t *testing.T) {
		got := globalSegments(t, `{% if 1 == 1 %}{% elsif 2 == 2 %}{{ test }}{% endif %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("else body", func(t *testing.T) {
		got := globalSegments(t, `{% if 1 == 1 %}{% else %}{{ test }}{% endif %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("case left", func(t *testing.T) {
		// Ruby: test_case_left — case expression itself is a variable
		got := globalSegments(t, `{% case test %}{% endcase %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("case condition", func(t *testing.T) {
		// Ruby: test_case_condition — when value is a variable
		got := globalSegments(t, `{% case 1 %}{% when test %}{% endcase %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("case when body", func(t *testing.T) {
		got := globalSegments(t, `{% case 1 %}{% when 2 %}{{ test }}{% endcase %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("case else body", func(t *testing.T) {
		got := globalSegments(t, `{% case 1 %}{% else %}{{ test }}{% endcase %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("for in", func(t *testing.T) {
		// Ruby: test_for_in — collection is a variable
		got := globalSegments(t, `{% for x in test %}{% endfor %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("for body", func(t *testing.T) {
		got := globalSegments(t, `{% for x in (1..5) %}{{ test }}{% endfor %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("for range variable", func(t *testing.T) {
		// Ruby: test_for_range — range endpoint is a variable
		got := globalSegments(t, `{% for x in (1..test) %}{% endfor %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("tablerow in", func(t *testing.T) {
		got := globalSegments(t, `{% tablerow x in test %}{% endtablerow %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("tablerow body", func(t *testing.T) {
		got := globalSegments(t, `{% tablerow x in (1..5) %}{{ test }}{% endtablerow %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("assign RHS", func(t *testing.T) {
		// Ruby: test_assign — RHS references "test"
		got := globalSegments(t, `{% assign x = test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("capture body", func(t *testing.T) {
		got := globalSegments(t, `{% capture x %}{{ test }}{% endcapture %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// ── Variables (all, including locals) ──

	t.Run("assign: All includes local and global", func(t *testing.T) {
		got := allSegments(t, `{% assign x = test %}{{ x }}`)
		want := [][]string{{"test"}, {"x"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("for: loop var in All", func(t *testing.T) {
		got := allSegments(t, `{% for item in list %}{{ item }}{% endfor %}`)
		want := [][]string{{"list"}, {"item"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("root names: deduplicate", func(t *testing.T) {
		got := globalRootNames(t, `{{ x.a }} {{ x.b }}`)
		want := []string{"x"}
		if !stringSliceSetEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// ── Ruby: cases not yet ported ──

	t.Run("dynamic variable", func(t *testing.T) {
		// Ruby: test_dynamic_variable — {{ test[inlookup] }} references both test and inlookup
		got := globalSegments(t, `{{ test[inlookup] }}`)
		want := [][]string{{"test"}, {"inlookup"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("echo tag", func(t *testing.T) {
		// Ruby: test_echo — {% echo test %} references test
		got := globalSegments(t, `{% echo test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("for limit variable", func(t *testing.T) {
		// Ruby: test_for_limit — limit: test is a referenced global
		got := globalSegments(t, `{% for x in (1..5) limit: test %}{% endfor %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("for offset variable", func(t *testing.T) {
		// Ruby: test_for_offset — offset: test is a referenced global
		got := globalSegments(t, `{% for x in (1..5) offset: test %}{% endfor %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("tablerow limit variable", func(t *testing.T) {
		// Ruby: test_tablerow_limit
		got := globalSegments(t, `{% tablerow x in (1..5) limit: test %}{% endtablerow %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("tablerow offset variable", func(t *testing.T) {
		// Ruby: test_tablerow_offset
		got := globalSegments(t, `{% tablerow x in (1..5) offset: test %}{% endtablerow %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("include dynamic filename", func(t *testing.T) {
		// Ruby: test_include — {% include test %} references test (dynamic filename)
		got := globalSegments(t, `{% include test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("include with variable", func(t *testing.T) {
		// Ruby: test_include_with — {% include "hai" with test %}
		got := globalSegments(t, `{% include "hai" with test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("include for variable", func(t *testing.T) {
		// Ruby: test_include_for — {% include "hai" for test %}
		got := globalSegments(t, `{% include "hai" for test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("render with variable", func(t *testing.T) {
		// Ruby: test_render_with — {% render "hai" with test %}
		got := globalSegments(t, `{% render "hai" with test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("render for variable", func(t *testing.T) {
		// Ruby: test_render_for — {% render "hai" for test %}
		got := globalSegments(t, `{% render "hai" for test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

// ── LiquidJS: static_analysis/variables.spec.ts ─────────────────────────────

func TestLiquidJS_VariableAnalysis(t *testing.T) {
	engine := NewEngine()

	t.Run("output statement", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{{ a }}`)
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"a"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("filter arguments as variables", func(t *testing.T) {
		// LiquidJS: "should report variables in filter arguments"
		tpl, _ := engine.ParseString(`{{ a | join: b }}`)
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"a"}, {"b"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("dotted properties", func(t *testing.T) {
		// LiquidJS: "should report dotted properties"
		tpl, _ := engine.ParseString(`{{ a.b }}`)
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"a", "b"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("group by root", func(t *testing.T) {
		// LiquidJS: "should group variables by their root value"
		tpl, _ := engine.ParseString(`{{ a.b }} {{ a.c }}`)
		got, _ := engine.GlobalVariableSegments(tpl)
		// Both a.b and a.c should appear
		want := [][]string{{"a", "b"}, {"a", "c"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("detect local variables via assign", func(t *testing.T) {
		// LiquidJS: "should detect local variables"
		tpl, _ := engine.ParseString(`{% assign a = "foo" %}{{ a }}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		all, _ := engine.VariableSegments(tpl)

		// a is local, so no globals
		if len(globals) != 0 {
			t.Errorf("globals: got %v, want empty", globals)
		}
		// but a appears in all (it is referenced in {{ a }})
		wantAll := [][]string{{"a"}}
		if !segmentsEqual(all, wantAll) {
			t.Errorf("all: got %v, want %v", all, wantAll)
		}
	})

	t.Run("assign RHS is global", func(t *testing.T) {
		// LiquidJS: "should report variables from assign tags"
		tpl, _ := engine.ParseString(`{% assign a = b %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"b"}}
		if !segmentsEqual(globals, want) {
			t.Errorf("got %v, want %v", globals, want)
		}
	})

	t.Run("capture inner variables", func(t *testing.T) {
		// LiquidJS: "should report variables from capture tags"
		tpl, _ := engine.ParseString(`{% capture a %}{% if b %}c{% endif %}{% endcapture %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"b"}}
		if !segmentsEqual(globals, want) {
			t.Errorf("got %v, want %v", globals, want)
		}
	})

	t.Run("if tags", func(t *testing.T) {
		// LiquidJS: "should report variables in if tags"
		tpl, _ := engine.ParseString(`{% if a %}b{% endif %}`)
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"a"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("nested blocks", func(t *testing.T) {
		// LiquidJS: "should report variables in nested blocks"
		tpl, _ := engine.ParseString(`{% if true %}{% if false %}{{ a }}{% endif %}{% endif %}`)
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"a"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("case with multiple when and else", func(t *testing.T) {
		// LiquidJS: "should report variables from case tags"
		src := "{% case x %}{% when y %}{{ a }}{% when z %}{{ b }}{% else %}{{ c }}{% endcase %}"
		tpl, _ := engine.ParseString(src)
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"x"}, {"y"}, {"a"}, {"z"}, {"b"}, {"c"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("for tags with range", func(t *testing.T) {
		// LiquidJS: for tags — range endpoint is a variable
		tpl, _ := engine.ParseString(`{% for x in (1..y) %}{{ x }}{% endfor %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		all, _ := engine.VariableSegments(tpl)

		wantGlobals := [][]string{{"y"}}
		wantAll := [][]string{{"y"}, {"x"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
		if !segmentsEqual(all, wantAll) {
			t.Errorf("all: got %v, want %v", all, wantAll)
		}
	})

	t.Run("for loop variable is local", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% for x in items %}{{ x.name }}{% endfor %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)

		wantGlobals := [][]string{{"items"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
	})

	t.Run("if with elsif and else", func(t *testing.T) {
		// LiquidJS: "should report variables from if tags" (full form)
		src := `{% if x %}{{ a }}{% elsif y %}{{ b }}{% else %}{{ c }}{% endif %}`
		tpl, _ := engine.ParseString(src)
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"x"}, {"a"}, {"y"}, {"b"}, {"c"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("tablerow", func(t *testing.T) {
		// LiquidJS: "should report variables from tablerow tags"
		tpl, _ := engine.ParseString(`{% tablerow x in y.z %}{{ x | append: a }}{% endtablerow %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		all, _ := engine.VariableSegments(tpl)

		// y.z is the collection, a is a filter arg — both global
		// x is the loop variable — local
		wantGlobals := [][]string{{"y", "z"}, {"a"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}

		wantAll := [][]string{{"y", "z"}, {"x"}, {"a"}}
		if !segmentsEqual(all, wantAll) {
			t.Errorf("all: got %v, want %v", all, wantAll)
		}
	})

	t.Run("unless with elsif and else", func(t *testing.T) {
		// LiquidJS: "should report variables from unless tags"
		src := `{% unless x %}{{ a }}{% else %}{{ c }}{% endunless %}`
		tpl, _ := engine.ParseString(src)
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"x"}, {"a"}, {"c"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("nested tags deep", func(t *testing.T) {
		// LiquidJS: "should report variables from nested tags"
		src := `{% if a %}{% for x in b %}{% unless x == y %}{% if 42 == c %}{{ a }}, {{ y }}{% endif %}{% endunless %}{% endfor %}{% endif %}`
		tpl, err := engine.ParseString(src)
		if err != nil {
			t.Fatal(err)
		}

		globals, _ := engine.GlobalVariableSegments(tpl)
		all, _ := engine.VariableSegments(tpl)

		// a, b, c, y are globals; x is loop variable (local)
		wantGlobals := [][]string{{"a"}, {"b"}, {"y"}, {"c"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}

		wantAll := [][]string{{"a"}, {"b"}, {"x"}, {"y"}, {"c"}}
		if !segmentsEqual(all, wantAll) {
			t.Errorf("all: got %v, want %v", all, wantAll)
		}
	})

	t.Run("multiple filters with variable args", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{{ a | append: b | prepend: c }}`)
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"a"}, {"b"}, {"c"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("assign chain: intermediate locals", func(t *testing.T) {
		// Chained assigns: y is local (assigned from x), z is local (assigned from y)
		src := `{% assign y = x %}{% assign z = y %}{{ z }}`
		tpl, _ := engine.ParseString(src)
		globals, _ := engine.GlobalVariableSegments(tpl)
		all, _ := engine.VariableSegments(tpl)

		// Only x is global
		wantGlobals := [][]string{{"x"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}

		// All includes x, y, z
		wantAll := [][]string{{"x"}, {"y"}, {"z"}}
		if !segmentsEqual(all, wantAll) {
			t.Errorf("all: got %v, want %v", all, wantAll)
		}
	})

	t.Run("capture makes local, inner uses global", func(t *testing.T) {
		src := `{% capture buf %}hello {{ name }}{% endcapture %}{{ buf }}`
		tpl, _ := engine.ParseString(src)
		globals, _ := engine.GlobalVariableSegments(tpl)

		// name is global; buf is local (capture-defined)
		wantGlobals := [][]string{{"name"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
	})

	t.Run("forloop is local inside for block", func(t *testing.T) {
		// LiquidJS: forloop is injected by the for tag itself, so it's local — not a global.
		// Bug was: forloop was missing from BlockScope in loopBlockAnalyzerFull.
		tpl, _ := engine.ParseString(`{% for item in order.items %}{{ forloop.index }}: {{ item.name }}{% endfor %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)

		wantGlobals := [][]string{{"order", "items"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
	})

	t.Run("bracket string key treated as property access", func(t *testing.T) {
		// LiquidJS: obj["prop"] == obj.prop — string literal key is a named property.
		// Bug was: IndexValue treated all [] as dynamic, recording only the base path.
		tpl, _ := engine.ParseString(`{{ customer["first_name"] }}`)
		got, _ := engine.GlobalVariableSegments(tpl)

		want := [][]string{{"customer", "first_name"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("bracket string key with spaces", func(t *testing.T) {
		// LiquidJS: a["b c"] — string key with spaces treated as property name
		tpl, _ := engine.ParseString(`{{ a["b c"] }}`)
		got, _ := engine.GlobalVariableSegments(tpl)

		want := [][]string{{"a", "b c"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("numeric index records base path only", func(t *testing.T) {
		// LiquidJS: a[1] — numeric index, only base path recorded
		tpl, _ := engine.ParseString(`{{ a[1] }}`)
		got, _ := engine.GlobalVariableSegments(tpl)

		// Go records base path as ["a"] — numeric index is discarded
		if len(got) == 0 {
			t.Errorf("expected at least one segment, got empty")
			return
		}
		if got[0][0] != "a" {
			t.Errorf("expected root 'a', got %v", got)
		}
	})

	t.Run("nested variable as index key", func(t *testing.T) {
		// LiquidJS: a[b.c] — both 'a' (base) and 'b.c' (key) are recorded as globals
		tpl, _ := engine.ParseString(`{{ a[b.c] }}`)
		got, _ := engine.GlobalVariableSegments(tpl)

		roots := map[string]bool{}
		for _, seg := range got {
			roots[strings.Join(seg, ".")] = true
		}
		if !roots["a"] {
			t.Errorf("expected 'a' in segments, got %v", got)
		}
		if !roots["b.c"] {
			t.Errorf("expected 'b.c' in segments, got %v", got)
		}
	})

	t.Run("deeply nested variable as index key", func(t *testing.T) {
		// LiquidJS: d[a[b.c]] — d, a, b.c all recorded as globals
		tpl, _ := engine.ParseString(`{{ d[a[b.c]] }}`)
		got, _ := engine.GlobalVariableSegments(tpl)

		roots := map[string]bool{}
		for _, seg := range got {
			roots[strings.Join(seg, ".")] = true
		}
		for _, expected := range []string{"d", "a", "b.c"} {
			if !roots[expected] {
				t.Errorf("expected %q in segments, got %v", expected, got)
			}
		}
	})

	t.Run("filter keyword argument variables", func(t *testing.T) {
		// LiquidJS: {{ a | default: b, allow_false: c }} — a, b, c all globals
		tpl, _ := engine.ParseString(`{{ a | default: b, allow_false: c }}`)
		got, _ := engine.GlobalVariableSegments(tpl)

		roots := map[string]bool{}
		for _, seg := range got {
			roots[strings.Join(seg, ".")] = true
		}
		for _, expected := range []string{"a", "b", "c"} {
			if !roots[expected] {
				t.Errorf("expected %q in segments, got %v", expected, got)
			}
		}
	})

	t.Run("decrement creates local", func(t *testing.T) {
		// LiquidJS: {% decrement a %} — 'a' is local (counter), no globals
		tpl, _ := engine.ParseString(`{% decrement a %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		all, _ := engine.VariableSegments(tpl)

		if len(globals) != 0 {
			t.Errorf("expected no globals, got %v", globals)
		}
		// 'a' counter is tracked as local — not referenced in output, so All is also empty
		_ = all
	})

	t.Run("increment creates local", func(t *testing.T) {
		// LiquidJS: {% increment a %} — 'a' is local (counter), no globals
		tpl, _ := engine.ParseString(`{% increment a %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)

		if len(globals) != 0 {
			t.Errorf("expected no globals, got %v", globals)
		}
	})

	t.Run("echo tag with filter kwargs", func(t *testing.T) {
		// LiquidJS: {% echo x | default: y, allow_false: z %} — x, y, z all globals
		tpl, _ := engine.ParseString(`{% echo x | default: y, allow_false: z %}`)
		got, _ := engine.GlobalVariableSegments(tpl)

		roots := map[string]bool{}
		for _, seg := range got {
			roots[strings.Join(seg, ".")] = true
		}
		for _, expected := range []string{"x", "y", "z"} {
			if !roots[expected] {
				t.Errorf("expected %q in segments, got %v", expected, got)
			}
		}
	})

	t.Run("for tag full — forloop is local, else and break/continue", func(t *testing.T) {
		// LiquidJS: full for block — forloop local, range var global, else body global
		src := "{% for x in (1..y) limit: a %}\n  {{ x }} {{ forloop.index }} {{ forloop.first }}\n{% break %}\n{% else %}\n  {{ z }}\n{% continue %}\n{% endfor %}"
		tpl, _ := engine.ParseString(src)
		globals, _ := engine.GlobalVariableSegments(tpl)
		all, _ := engine.VariableSegments(tpl)

		globalRoots := map[string]bool{}
		for _, seg := range globals {
			globalRoots[strings.Join(seg, ".")] = true
		}
		allRoots := map[string]bool{}
		for _, seg := range all {
			allRoots[strings.Join(seg, ".")] = true
		}

		// y (range end), a (limit), z (else body) are globals
		for _, expected := range []string{"y", "a", "z"} {
			if !globalRoots[expected] {
				t.Errorf("expected %q in globals, got %v", expected, globals)
			}
		}
		// forloop is local — not in globals
		if globalRoots["forloop"] || globalRoots["forloop.index"] || globalRoots["forloop.first"] {
			t.Errorf("forloop should not be global, got globals %v", globals)
		}
		// x is loop var — not in globals
		if globalRoots["x"] {
			t.Errorf("x should not be global, got globals %v", globals)
		}
		// x and forloop.* appear in All
		if !allRoots["x"] {
			t.Errorf("expected x in all, got %v", all)
		}
	})

	t.Run("liquid tag inner variables", func(t *testing.T) {
		// LiquidJS: variables inside {% liquid %} block are analyzed
		src := "{% liquid\n  if product.title\n    echo foo | upcase\n  else\n    echo \"product-1\" | upcase\n  endif\n  \n  for i in (0..5)\n    echo i\nendfor %}"
		tpl, _ := engine.ParseString(src)
		globals, _ := engine.GlobalVariableSegments(tpl)
		all, _ := engine.VariableSegments(tpl)

		globalRoots := map[string]bool{}
		for _, seg := range globals {
			globalRoots[strings.Join(seg, ".")] = true
		}
		allRoots := map[string]bool{}
		for _, seg := range all {
			allRoots[strings.Join(seg, ".")] = true
		}

		if !globalRoots["product.title"] {
			t.Errorf("expected product.title in globals, got %v", globals)
		}
		if !globalRoots["foo"] {
			t.Errorf("expected foo in globals, got %v", globals)
		}
		// i is loop var — local
		if globalRoots["i"] {
			t.Errorf("i should not be global, got %v", globals)
		}
		if !allRoots["i"] {
			t.Errorf("expected i in all, got %v", all)
		}
	})

	t.Run("unless tag full — with else", func(t *testing.T) {
		// LiquidJS: {% unless x %}{{ a }}{% else %}{{ c }}{% endunless %}
		src := "{% unless x %}\n  {{ a }}\n{% else %}\n  {{ c }}\n{% endunless %}"
		tpl, _ := engine.ParseString(src)
		got, _ := engine.GlobalVariableSegments(tpl)

		roots := map[string]bool{}
		for _, seg := range got {
			roots[strings.Join(seg, ".")] = true
		}
		for _, expected := range []string{"x", "a", "c"} {
			if !roots[expected] {
				t.Errorf("expected %q in globals, got %v", expected, got)
			}
		}
	})

	t.Run("deeply nested tags", func(t *testing.T) {
		// LiquidJS: nested if+for+unless — a, b, c, y are globals; x is local
		src := "{% if a %}\n  {% for x in b %}\n    {% unless x == y %}\n      {% if 42 == c %}\n        {{ a }}, {{ y }}\n      {% endif %}\n    {% endunless %}\n  {% endfor %}\n{% endif %}"
		tpl, _ := engine.ParseString(src)
		globals, _ := engine.GlobalVariableSegments(tpl)
		all, _ := engine.VariableSegments(tpl)

		globalRoots := map[string]bool{}
		for _, seg := range globals {
			globalRoots[strings.Join(seg, ".")] = true
		}
		allRoots := map[string]bool{}
		for _, seg := range all {
			allRoots[strings.Join(seg, ".")] = true
		}

		for _, expected := range []string{"a", "b", "c", "y"} {
			if !globalRoots[expected] {
				t.Errorf("expected %q in globals, got %v", expected, globals)
			}
		}
		if globalRoots["x"] {
			t.Errorf("x should not be global, got %v", globals)
		}
		if !allRoots["x"] {
			t.Errorf("expected x in all (loop var), got %v", all)
		}
	})
}

// ── LiquidJS: Analyze / StaticAnalysis tests ────────────────────────────────

func TestLiquidJS_StaticAnalysis(t *testing.T) {
	engine := NewEngine()

	t.Run("Analyze returns tags used", func(t *testing.T) {
		// Tests that Tags field in StaticAnalysis includes all tag types
		tpl, _ := engine.ParseString(`{% assign x = 1 %}{% for i in list %}{{ i }}{% endfor %}{% if cond %}ok{% endif %}{% unless flag %}no{% endunless %}`)
		analysis, err := engine.Analyze(tpl)
		if err != nil {
			t.Fatal(err)
		}

		tagSet := map[string]bool{}
		for _, tag := range analysis.Tags {
			tagSet[tag] = true
		}
		for _, expected := range []string{"assign", "for", "if", "unless"} {
			if !tagSet[expected] {
				t.Errorf("expected tag %q in Tags, got %v", expected, analysis.Tags)
			}
		}
	})

	t.Run("Analyze globals excludes locals", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% assign y = x.val %}{{ y }} {{ z }}`)
		analysis, err := engine.Analyze(tpl)
		if err != nil {
			t.Fatal(err)
		}

		globalNames := map[string]bool{}
		for _, v := range analysis.Globals {
			globalNames[v.String()] = true
		}

		if !globalNames["x.val"] {
			t.Errorf("expected x.val in Globals, got %v", analysis.Globals)
		}
		if !globalNames["z"] {
			t.Errorf("expected z in Globals, got %v", analysis.Globals)
		}
		if globalNames["y"] {
			t.Errorf("y should not be in Globals, got %v", analysis.Globals)
		}
	})

	t.Run("Analyze locals includes assigned vars", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% assign x = src %}{{ x }}`)
		analysis, err := engine.Analyze(tpl)
		if err != nil {
			t.Fatal(err)
		}

		localsSet := map[string]bool{}
		for _, l := range analysis.Locals {
			localsSet[l] = true
		}
		if !localsSet["x"] {
			t.Errorf("expected x in Locals, got %v", analysis.Locals)
		}
	})

	t.Run("Analyze locals includes capture vars", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% capture buf %}test{% endcapture %}{{ buf }}`)
		analysis, err := engine.Analyze(tpl)
		if err != nil {
			t.Fatal(err)
		}

		localsSet := map[string]bool{}
		for _, l := range analysis.Locals {
			localsSet[l] = true
		}
		if !localsSet["buf"] {
			t.Errorf("expected buf in Locals, got %v", analysis.Locals)
		}
	})

	t.Run("Analyze locals includes for loop vars", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% for item in list %}{{ item }}{% endfor %}`)
		analysis, err := engine.Analyze(tpl)
		if err != nil {
			t.Fatal(err)
		}

		localsSet := map[string]bool{}
		for _, l := range analysis.Locals {
			localsSet[l] = true
		}
		if !localsSet["item"] {
			t.Errorf("expected item in Locals, got %v", analysis.Locals)
		}
	})

	t.Run("Analyze Variables includes both local and global", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% assign y = x %}{{ y }} {{ z }}`)
		analysis, err := engine.Analyze(tpl)
		if err != nil {
			t.Fatal(err)
		}

		varNames := map[string]bool{}
		for _, v := range analysis.Variables {
			varNames[v.String()] = true
		}

		for _, expected := range []string{"x", "y", "z"} {
			if !varNames[expected] {
				t.Errorf("expected %q in Variables, got %v", expected, analysis.Variables)
			}
		}
	})

	t.Run("ParseAndAnalyze returns complete analysis", func(t *testing.T) {
		tpl, analysis, err := engine.ParseAndAnalyze([]byte(`{% assign x = src %}{{ x }} {{ z }}`))
		if err != nil {
			t.Fatal(err)
		}
		if tpl == nil {
			t.Fatal("expected non-nil template")
		}
		if analysis == nil {
			t.Fatal("expected non-nil StaticAnalysis")
		}

		// Check globals
		globalNames := map[string]bool{}
		for _, v := range analysis.Globals {
			globalNames[v.String()] = true
		}
		if !globalNames["src"] || !globalNames["z"] {
			t.Errorf("expected src and z in Globals, got %v", analysis.Globals)
		}
		if globalNames["x"] {
			t.Errorf("x should not be in Globals")
		}

		// Check locals
		localsSet := map[string]bool{}
		for _, l := range analysis.Locals {
			localsSet[l] = true
		}
		if !localsSet["x"] {
			t.Errorf("expected x in Locals, got %v", analysis.Locals)
		}

		// Check tags
		tagSet := map[string]bool{}
		for _, tag := range analysis.Tags {
			tagSet[tag] = true
		}
		if !tagSet["assign"] {
			t.Errorf("expected assign in Tags, got %v", analysis.Tags)
		}
	})
}

// ── LiquidJS: Variable struct unit tests ─────────────────────────────────────
// Source: src/template/analysis.spec.ts

func TestLiquidJS_VariableStruct(t *testing.T) {
	t.Run("String joins segments with dot", func(t *testing.T) {
		v := Variable{Segments: []string{"foo", "bar"}}
		if got := v.String(); got != "foo.bar" {
			t.Errorf("got %q, want %q", got, "foo.bar")
		}
	})

	t.Run("String single segment", func(t *testing.T) {
		v := Variable{Segments: []string{"foo"}}
		if got := v.String(); got != "foo" {
			t.Errorf("got %q, want %q", got, "foo")
		}
	})

	t.Run("Segments property", func(t *testing.T) {
		v := Variable{Segments: []string{"customer", "name"}}
		if len(v.Segments) != 2 || v.Segments[0] != "customer" || v.Segments[1] != "name" {
			t.Errorf("got segments %v", v.Segments)
		}
	})

	t.Run("Global field", func(t *testing.T) {
		v := Variable{Segments: []string{"x"}, Global: true}
		if !v.Global {
			t.Error("expected Global=true")
		}
		v2 := Variable{Segments: []string{"y"}, Global: false}
		if v2.Global {
			t.Error("expected Global=false")
		}
	})
}

// ── LiquidJS: FullVariables with Global marking ──────────────────────────────
// Source: test/e2e/parse-and-analyze.spec.ts (adapted)

func TestLiquidJS_FullVariablesGlobalMarking(t *testing.T) {
	engine := NewEngine()

	t.Run("marks assign-defined as local", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% assign y = x.val %}{{ y }} {{ z }}`)
		got, err := engine.FullVariables(tpl)
		if err != nil {
			t.Fatal(err)
		}

		byName := map[string]Variable{}
		for _, v := range got {
			byName[v.String()] = v
		}

		if v, ok := byName["x.val"]; !ok || !v.Global {
			t.Errorf("expected x.val to be global")
		}
		if v, ok := byName["y"]; !ok || v.Global {
			t.Errorf("expected y to be local")
		}
		if v, ok := byName["z"]; !ok || !v.Global {
			t.Errorf("expected z to be global")
		}
	})

	t.Run("GlobalFullVariables excludes locals", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% assign y = src %}{{ y }} {{ z.val }}`)
		got, err := engine.GlobalFullVariables(tpl)
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range got {
			if !v.Global {
				t.Errorf("GlobalFullVariables: all should be Global=true, got %v", v)
			}
		}

		names := map[string]bool{}
		for _, v := range got {
			names[v.String()] = true
		}
		if !names["src"] {
			t.Error("expected src in globals")
		}
		if !names["z.val"] {
			t.Error("expected z.val in globals")
		}
		if names["y"] {
			t.Error("y should not be in globals")
		}
	})

	t.Run("for loop var marked as local in FullVariables", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% for item in products %}{{ item.title }}{% endfor %}`)
		got, err := engine.FullVariables(tpl)
		if err != nil {
			t.Fatal(err)
		}

		byName := map[string]Variable{}
		for _, v := range got {
			byName[v.String()] = v
		}

		if v, ok := byName["products"]; !ok || !v.Global {
			t.Errorf("expected products to be global")
		}
		if v, ok := byName["item.title"]; !ok || v.Global {
			t.Errorf("expected item.title to be local, got %v", v)
		}
	})

	t.Run("capture var marked as local in FullVariables", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% capture buf %}hello{% endcapture %}{{ buf }}`)
		got, err := engine.FullVariables(tpl)
		if err != nil {
			t.Fatal(err)
		}

		byName := map[string]Variable{}
		for _, v := range got {
			byName[v.String()] = v
		}

		if v, ok := byName["buf"]; !ok || v.Global {
			t.Errorf("expected buf to be local")
		}
	})
}

// ── Template method tests (convenience API) ──────────────────────────────────
// Source: adapted from LiquidJS where template itself has analysis methods

func TestLiquidJS_TemplateAnalysisMethods(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{% assign y = x %}{{ y }} {{ z }}`)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("VariableSegments", func(t *testing.T) {
		got, err := tpl.VariableSegments()
		if err != nil {
			t.Fatal(err)
		}
		want := [][]string{{"x"}, {"y"}, {"z"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("GlobalVariableSegments", func(t *testing.T) {
		got, err := tpl.GlobalVariableSegments()
		if err != nil {
			t.Fatal(err)
		}
		want := [][]string{{"x"}, {"z"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Variables", func(t *testing.T) {
		got, err := tpl.Variables()
		if err != nil {
			t.Fatal(err)
		}
		if !stringSliceSetEqual(got, []string{"x", "y", "z"}) {
			t.Errorf("got %v", got)
		}
	})

	t.Run("GlobalVariables", func(t *testing.T) {
		got, err := tpl.GlobalVariables()
		if err != nil {
			t.Fatal(err)
		}
		if !stringSliceSetEqual(got, []string{"x", "z"}) {
			t.Errorf("got %v", got)
		}
	})

	t.Run("FullVariables", func(t *testing.T) {
		got, err := tpl.FullVariables()
		if err != nil {
			t.Fatal(err)
		}
		if len(got) == 0 {
			t.Error("expected non-empty FullVariables")
		}
	})

	t.Run("GlobalFullVariables", func(t *testing.T) {
		got, err := tpl.GlobalFullVariables()
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range got {
			if !v.Global {
				t.Errorf("expected all Global=true, got %v", v)
			}
		}
	})

	t.Run("Analyze", func(t *testing.T) {
		analysis, err := tpl.Analyze()
		if err != nil {
			t.Fatal(err)
		}
		if analysis == nil {
			t.Fatal("expected non-nil StaticAnalysis")
		}
		if len(analysis.Variables) == 0 {
			t.Error("expected non-empty Variables")
		}
		if len(analysis.Globals) == 0 {
			t.Error("expected non-empty Globals")
		}
	})
}

// ── Edge cases and combined scenarios ────────────────────────────────────────

func TestAnalysis_EdgeCases(t *testing.T) {
	engine := NewEngine()

	t.Run("empty template", func(t *testing.T) {
		tpl, _ := engine.ParseString(``)
		globals, _ := engine.GlobalVariableSegments(tpl)
		if len(globals) != 0 {
			t.Errorf("expected empty, got %v", globals)
		}
	})

	t.Run("only text, no variables", func(t *testing.T) {
		tpl, _ := engine.ParseString(`hello world`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		if len(globals) != 0 {
			t.Errorf("expected empty, got %v", globals)
		}
	})

	t.Run("only literals in output", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{{ "hello" }}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		if len(globals) != 0 {
			t.Errorf("expected empty, got %v", globals)
		}
	})

	t.Run("only numeric literal in output", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{{ 42 }}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		if len(globals) != 0 {
			t.Errorf("expected empty, got %v", globals)
		}
	})

	t.Run("assign literal does not produce globals", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% assign x = "hello" %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		if len(globals) != 0 {
			t.Errorf("expected empty, got %v", globals)
		}
	})

	t.Run("boolean literals in conditions", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% if true %}yes{% endif %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		if len(globals) != 0 {
			t.Errorf("expected empty, got %v", globals)
		}
	})

	t.Run("nested for loops", func(t *testing.T) {
		src := `{% for a in list_a %}{% for b in list_b %}{{ a.x }} {{ b.y }}{% endfor %}{% endfor %}`
		tpl, _ := engine.ParseString(src)
		globals, _ := engine.GlobalVariableSegments(tpl)

		// list_a and list_b are globals; a and b are loop variables
		wantGlobals := [][]string{{"list_a"}, {"list_b"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
	})

	t.Run("for with else", func(t *testing.T) {
		src := `{% for item in list %}{{ item.name }}{% else %}{{ fallback }}{% endfor %}`
		tpl, _ := engine.ParseString(src)
		globals, _ := engine.GlobalVariableSegments(tpl)

		wantGlobals := [][]string{{"list"}, {"fallback"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
	})

	t.Run("if inside for", func(t *testing.T) {
		src := `{% for item in products %}{% if item.active %}{{ item.name }}{% endif %}{% endfor %}`
		tpl, _ := engine.ParseString(src)
		globals, _ := engine.GlobalVariableSegments(tpl)

		// Only products is global; item is the loop variable
		wantGlobals := [][]string{{"products"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
	})

	t.Run("assign inside for loop", func(t *testing.T) {
		src := `{% for item in list %}{% assign x = item.val %}{{ x }}{% endfor %}`
		tpl, _ := engine.ParseString(src)
		globals, _ := engine.GlobalVariableSegments(tpl)

		// Only list is global; item and x are local
		wantGlobals := [][]string{{"list"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
	})

	t.Run("tablerow with property access on collection", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% tablerow item in site.products %}{{ item.title }}{% endtablerow %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)

		wantGlobals := [][]string{{"site", "products"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
	})

	t.Run("multiple assigns redefine same variable", func(t *testing.T) {
		src := `{% assign x = a %}{% assign x = b %}{{ x }}`
		tpl, _ := engine.ParseString(src)
		globals, _ := engine.GlobalVariableSegments(tpl)

		// a and b are both globals (both assigned to x)
		wantGlobals := [][]string{{"a"}, {"b"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
	})

	t.Run("filter with literal arg does not add variable", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{{ name | append: " Smith" }}`)
		globals, _ := engine.GlobalVariableSegments(tpl)

		wantGlobals := [][]string{{"name"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
	})

	t.Run("complex template with all tag types", func(t *testing.T) {
		src := `{% assign title = page.title %}` +
			`{% capture header %}<h1>{{ title }}</h1>{% endcapture %}` +
			`{{ header }}` +
			`{% for item in products %}` +
			`{% if item.active %}{{ item.name | append: suffix }}{% endif %}` +
			`{% endfor %}` +
			`{% unless hide_footer %}{{ footer_text }}{% endunless %}` +
			`{% case status %}{% when "active" %}{{ active_msg }}{% else %}{{ default_msg }}{% endcase %}`

		tpl, err := engine.ParseString(src)
		if err != nil {
			t.Fatal(err)
		}

		globals, _ := engine.GlobalVariableSegments(tpl)
		globalRoots, _ := engine.GlobalVariables(tpl)

		// Expected globals: page.title, products, suffix, hide_footer, footer_text, status, active_msg, default_msg
		// NOT globals: title (assigned), header (captured), item (loop var)
		expectedRoots := []string{"page", "products", "suffix", "hide_footer", "footer_text", "status", "active_msg", "default_msg"}
		if !stringSliceSetEqual(globalRoots, expectedRoots) {
			t.Errorf("globalRoots: got %v, want %v", globalRoots, expectedRoots)
		}

		// page.title should be in full segments
		found := false
		for _, seg := range globals {
			if len(seg) == 2 && seg[0] == "page" && seg[1] == "title" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected [page title] in globals, got %v", globals)
		}
	})
}

// ── Ruby Liquid: ParseTreeVisitor — missing tests ───────────────────────────
// Source: test/unit/parse_tree_visitor_test.rb
// Tests not covered in TestRubyLiquid_ParseTreeVisitor above.

func TestRubyLiquid_ParseTreeVisitorExtra(t *testing.T) {
	engine := NewEngine()

	globalSegments := func(t *testing.T, src string) [][]string {
		t.Helper()
		tpl, parseErr := engine.ParseString(src)
		if parseErr != nil {
			t.Fatalf("ParseString(%q): %v", src, parseErr)
		}
		segs, analyzeErr := engine.GlobalVariableSegments(tpl)
		if analyzeErr != nil {
			t.Fatalf("GlobalVariableSegments: %v", analyzeErr)
		}
		return segs
	}

	// test_dynamic_variable: {{ test[inlookup] }}
	// Ruby: IndexValue records base path AND the key variable.
	t.Run("dynamic variable bracket notation", func(t *testing.T) {
		got := globalSegments(t, `{{ test[inlookup] }}`)
		// Both "test" (base object) and "inlookup" (dynamic index) should appear.
		roots := make(map[string]bool)
		for _, seg := range got {
			if len(seg) > 0 {
				roots[seg[0]] = true
			}
		}
		if !roots["test"] {
			t.Errorf("expected 'test' in globals, got %v", got)
		}
		if !roots["inlookup"] {
			t.Errorf("expected 'inlookup' in globals, got %v", got)
		}
	})

	// test_echo: {% echo test %}
	t.Run("echo tag", func(t *testing.T) {
		got := globalSegments(t, `{% echo test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// test_for_limit: {% for x in (1..5) limit: test %}
	t.Run("for limit variable", func(t *testing.T) {
		got := globalSegments(t, `{% for x in (1..5) limit: test %}{% endfor %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// test_for_offset: {% for x in (1..5) offset: test %}
	t.Run("for offset variable", func(t *testing.T) {
		got := globalSegments(t, `{% for x in (1..5) offset: test %}{% endfor %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// test_tablerow_limit: {% tablerow x in (1..5) limit: test %}
	t.Run("tablerow limit variable", func(t *testing.T) {
		got := globalSegments(t, `{% tablerow x in (1..5) limit: test %}{% endtablerow %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// test_tablerow_offset: {% tablerow x in (1..5) offset: test %}
	t.Run("tablerow offset variable", func(t *testing.T) {
		got := globalSegments(t, `{% tablerow x in (1..5) offset: test %}{% endtablerow %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// test_include: {% include test %} — dynamic filename from variable
	t.Run("include with dynamic filename variable", func(t *testing.T) {
		got := globalSegments(t, `{% include test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// test_include_with: {% include "hai" with test %}
	t.Run("include with 'with' variable", func(t *testing.T) {
		got := globalSegments(t, `{% include "hai" with test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// test_include_for: {% include "hai" for test %}
	t.Run("include with 'for' variable", func(t *testing.T) {
		got := globalSegments(t, `{% include "hai" for test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// test_render_with: {% render "hai" with test %}
	t.Run("render with 'with' variable", func(t *testing.T) {
		got := globalSegments(t, `{% render "hai" with test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// test_render_for: {% render "hai" for test %}
	t.Run("render with 'for' variable", func(t *testing.T) {
		got := globalSegments(t, `{% render "hai" for test %}`)
		want := [][]string{{"test"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

// ── LiquidJS: variables.spec.ts — missing tests ──────────────────────────────
// Source: test/integration/static_analysis/variables.spec.ts

func TestLiquidJS_VariableAnalysisExtra(t *testing.T) {
	engine := NewEngine()

	// "should report variables in filter keyword arguments"
	// {{ a | default: b, allow_false: c }} — c is a named keyword arg value
	t.Run("filter keyword arg variables", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{{ a | default: b, allow_false: c }}`)
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"a"}, {"b"}, {"c"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// "should detect when a variable is in scope" — whole-template analysis
	// Go uses whole-template (flow-insensitive) analysis: if a variable is assigned
	// anywhere in the template, it's treated as local everywhere. This differs from
	// LiquidJS which does flow-sensitive analysis (tracking use before assign).
	// Our behavior: {{ a }}{% assign a = "foo" %}{{ a }} → a is local, not global.
	t.Run("variable scope detection - assign makes var local everywhere", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{{ a }}{% assign a = "foo" %}{{ a }}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		// In Go's whole-template analysis, once a is assigned, it is local everywhere.
		for _, seg := range globals {
			if len(seg) == 1 && seg[0] == "a" {
				t.Errorf("expected 'a' to be local (assigned later), but found in globals: %v", globals)
			}
		}
	})

	// "should report variables from decrement tags" — decrement creates a local counter
	// Per LiquidJS spec, {% decrement a %} introduces a as a locally-defined name.
	t.Run("decrement creates local", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% decrement a %}`)
		analysis, err := engine.Analyze(tpl)
		if err != nil {
			t.Fatal(err)
		}
		localsSet := map[string]bool{}
		for _, l := range analysis.Locals {
			localsSet[l] = true
		}
		if !localsSet["a"] {
			t.Errorf("expected a in Locals for decrement, got %v", analysis.Locals)
		}
		// No global variables expected
		if len(analysis.Globals) != 0 {
			t.Errorf("expected no globals for decrement, got %v", analysis.Globals)
		}
	})

	// "should report variables from increment tags" — increment creates a local counter
	t.Run("increment creates local", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% increment a %}`)
		analysis, err := engine.Analyze(tpl)
		if err != nil {
			t.Fatal(err)
		}
		localsSet := map[string]bool{}
		for _, l := range analysis.Locals {
			localsSet[l] = true
		}
		if !localsSet["a"] {
			t.Errorf("expected a in Locals for increment, got %v", analysis.Locals)
		}
		if len(analysis.Globals) != 0 {
			t.Errorf("expected no globals for increment, got %v", analysis.Globals)
		}
	})

	// "should report variables from echo tags"
	// {% echo x | default: y, allow_false: z %} — x, y, z are all variables
	t.Run("echo tag variables with filter kwargs", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% echo x | default: y, allow_false: z %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"x"}, {"y"}, {"z"}}
		if !segmentsEqual(globals, want) {
			t.Errorf("got %v, want %v", globals, want)
		}
	})

	// "should report variables from for tags" — for with limit as variable
	// {% for x in (1..y) limit: a %}
	t.Run("for tags with limit variable", func(t *testing.T) {
		src := `{% for x in (1..y) limit: a %}{{ x }}{% endfor %}`
		tpl, _ := engine.ParseString(src)
		globals, _ := engine.GlobalVariableSegments(tpl)
		all, _ := engine.VariableSegments(tpl)

		// y and a are global; x is the loop variable (local)
		wantGlobals := [][]string{{"y"}, {"a"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}
		// x appears in all variables
		foundX := false
		for _, seg := range all {
			if len(seg) == 1 && seg[0] == "x" {
				foundX = true
			}
		}
		if !foundX {
			t.Errorf("expected x in all variables, got %v", all)
		}
	})

	// "should report variables from liquid tags"
	// {% liquid
	//   if product.title
	//     echo foo | upcase
	//   else
	//     echo "product-1" | upcase
	//   endif
	//   for i in (0..5)
	//     echo i
	// endfor %}
	t.Run("liquid tag inner variables", func(t *testing.T) {
		src := "{% liquid\n  if product.title\n    echo foo | upcase\n  else\n    echo \"product-1\" | upcase\n  endif\n  for i in (0..5)\n    echo i\nendfor %}"
		tpl, parseErr := engine.ParseString(src)
		if parseErr != nil {
			t.Fatalf("ParseString: %v", parseErr)
		}
		globals, analyzeErr := engine.GlobalVariableSegments(tpl)
		if analyzeErr != nil {
			t.Fatal(analyzeErr)
		}
		all, analyzeErr2 := engine.VariableSegments(tpl)
		if analyzeErr2 != nil {
			t.Fatal(analyzeErr2)
		}

		// product.title and foo should be globals
		globalMap := map[string]bool{}
		for _, seg := range globals {
			globalMap[strings.Join(seg, ".")] = true
		}
		if !globalMap["product.title"] {
			t.Errorf("expected product.title in globals, got %v", globals)
		}
		if !globalMap["foo"] {
			t.Errorf("expected foo in globals, got %v", globals)
		}

		// i is the for loop variable — should be in all variables
		allMap := map[string]bool{}
		for _, seg := range all {
			allMap[strings.Join(seg, ".")] = true
		}
		if !allMap["i"] {
			t.Errorf("expected i in all variables, got %v", all)
		}
		// i is local (block scope), so NOT in globals
		if globalMap["i"] {
			t.Errorf("i should not be in globals, got %v", globals)
		}
	})

	// "should report variables from tablerow tags"
	// {% tablerow x in y.z cols:2 %}{{ x | append: a }}{% endtablerow %}
	t.Run("tablerow variables — globals and loop var local", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% tablerow x in y.z cols:2 %}{{ x | append: a }}{% endtablerow %}`)
		globals, _ := engine.GlobalVariableSegments(tpl)
		all, _ := engine.VariableSegments(tpl)

		// y.z is the collection and a is a filter arg — both global
		// x is the loop variable — local
		wantGlobals := [][]string{{"y", "z"}, {"a"}}
		if !segmentsEqual(globals, wantGlobals) {
			t.Errorf("globals: got %v, want %v", globals, wantGlobals)
		}

		// x appears in all variables
		foundX := false
		for _, seg := range all {
			if len(seg) == 1 && seg[0] == "x" {
				foundX = true
			}
		}
		if !foundX {
			t.Errorf("expected x in all variables, got %v", all)
		}
	})

	// "should report variables from unless tags" — unless with else
	// Note: Go Liquid does not support elsif inside unless (Ruby/JS LiquidJS do not
	// either in standard mode). Only else is supported inside unless.
	t.Run("unless with else variables", func(t *testing.T) {
		src := "{% unless x %}{{ a }}{% else %}{{ c }}{% endunless %}"
		tpl, parseErr := engine.ParseString(src)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"x"}, {"a"}, {"c"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// "should report variables from include/render with key-value arguments"
	t.Run("include key-value args", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% include "file" x: foo, y: bar %}`)
		got, _ := engine.GlobalVariableSegments(tpl)
		// foo and bar are variable references in kv values
		roots := map[string]bool{}
		for _, seg := range got {
			if len(seg) > 0 {
				roots[seg[0]] = true
			}
		}
		if !roots["foo"] {
			t.Errorf("expected foo in globals, got %v", got)
		}
		if !roots["bar"] {
			t.Errorf("expected bar in globals, got %v", got)
		}
	})

	t.Run("render key-value args", func(t *testing.T) {
		tpl, _ := engine.ParseString(`{% render "file" x: foo, y: bar %}`)
		got, _ := engine.GlobalVariableSegments(tpl)
		roots := map[string]bool{}
		for _, seg := range got {
			if len(seg) > 0 {
				roots[seg[0]] = true
			}
		}
		if !roots["foo"] {
			t.Errorf("expected foo in globals, got %v", got)
		}
		if !roots["bar"] {
			t.Errorf("expected bar in globals, got %v", got)
		}
	})
}

// ── LiquidJS: include/render with TemplateStore partial traversal ─────────────
// Source: test/e2e/parse-and-analyze.spec.ts — partial analysis tests.
// These require a TemplateStore to load partial templates at analysis time.
// The include tag (shared scope) traverses into partials; render (isolated scope) does not.

// inMemoryStore is a simple in-memory TemplateStore for testing partial analysis.
type inMemoryStore struct {
	files map[string]string
}

func (s *inMemoryStore) ReadTemplate(filename string) ([]byte, error) {
	if src, ok := s.files[filename]; ok {
		return []byte(src), nil
	}
	return nil, fmt.Errorf("template not found: %s", filename)
}

func TestLiquidJS_PartialAnalysis(t *testing.T) {
	t.Run("include static literal — partial vars reported", func(t *testing.T) {
		// JS: engine.globalVariableSegmentsSync('{% include "product" %}')
		// with "product" template = '{{ product.name }}'
		// → [['product', 'name']]
		engine := NewEngine()
		engine.RegisterTemplateStore(&inMemoryStore{
			files: map[string]string{
				"product": `{{ product.name }}`,
			},
		})
		tpl, err := engine.ParseString(`{% include "product" %}`)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"product", "name"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("include with-arg and partial vars both reported", func(t *testing.T) {
		// include "product" with outer — outer is a tag arg (global),
		// plus product.price inside the partial is also reported.
		engine := NewEngine()
		engine.RegisterTemplateStore(&inMemoryStore{
			files: map[string]string{
				"product": `{{ product.price }}`,
			},
		})
		tpl, err := engine.ParseString(`{% include "product" with outer %}`)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := engine.GlobalVariableSegments(tpl)
		roots := map[string]bool{}
		for _, seg := range got {
			roots[strings.Join(seg, ".")] = true
		}
		if !roots["outer"] {
			t.Errorf("expected 'outer' (with-arg) in globals, got %v", got)
		}
		if !roots["product.price"] {
			t.Errorf("expected 'product.price' (from partial) in globals, got %v", got)
		}
	})

	t.Run("include chain A → B → C", func(t *testing.T) {
		// Multi-level include: root → "a" → "b" → {{ deep_var }}
		engine := NewEngine()
		engine.RegisterTemplateStore(&inMemoryStore{
			files: map[string]string{
				"a": `{% include "b" %}`,
				"b": `{{ deep_var }}`,
			},
		})
		tpl, err := engine.ParseString(`{% include "a" %}`)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"deep_var"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("include cycle — no infinite loop", func(t *testing.T) {
		// A includes B includes A — cycle detection must prevent infinite recursion.
		// The result is a safe partial analysis (variables from non-cyclic portions).
		engine := NewEngine()
		engine.RegisterTemplateStore(&inMemoryStore{
			files: map[string]string{
				"a": `{{ a_var }}{% include "b" %}`,
				"b": `{{ b_var }}{% include "a" %}`,
			},
		})
		// This must not panic or loop forever.
		tpl, err := engine.ParseString(`{% include "a" %}`)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := engine.GlobalVariableSegments(tpl)
		// a_var and b_var should both be present (from the non-cyclic traversal paths)
		roots := map[string]bool{}
		for _, seg := range got {
			if len(seg) > 0 {
				roots[seg[0]] = true
			}
		}
		if !roots["a_var"] {
			t.Errorf("expected 'a_var' in globals, got %v", got)
		}
		if !roots["b_var"] {
			t.Errorf("expected 'b_var' in globals, got %v", got)
		}
	})

	t.Run("include dynamic filename — no partial traversal", func(t *testing.T) {
		// {% include template %} — filename is a variable, cannot traverse statically.
		// Only the variable reference itself (template) should be reported.
		engine := NewEngine()
		engine.RegisterTemplateStore(&inMemoryStore{
			files: map[string]string{
				"product": `{{ product.name }}`,
			},
		})
		tpl, err := engine.ParseString(`{% include template %}`)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := engine.GlobalVariableSegments(tpl)
		want := [][]string{{"template"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("include — assign inside partial is local to outer scope", func(t *testing.T) {
		// include shares scope: {% assign x = "foo" %} inside partial defines x
		// in the parent scope too. So x should appear in Locals, not Globals.
		engine := NewEngine()
		engine.RegisterTemplateStore(&inMemoryStore{
			files: map[string]string{
				"setter": `{% assign x = "foo" %}`,
			},
		})
		tpl, err := engine.ParseString(`{% include "setter" %}{{ x }}`)
		if err != nil {
			t.Fatal(err)
		}
		globals, _ := engine.GlobalVariableSegments(tpl)
		// x is defined inside the partial — shared scope means it's local to the whole template
		for _, seg := range globals {
			if len(seg) > 0 && seg[0] == "x" {
				t.Errorf("x should not be global (it is assigned in included partial), got globals %v", globals)
			}
		}
	})

	t.Run("render tag — no partial traversal (isolated scope)", func(t *testing.T) {
		// render creates isolated scope; internal vars are NOT globals of the outer template.
		// Only the render tag's explicit arguments (with/for/kv) are reported.
		engine := NewEngine()
		engine.RegisterTemplateStore(&inMemoryStore{
			files: map[string]string{
				"product": `{{ product.name }} {{ product.price }}`,
			},
		})
		tpl, err := engine.ParseString(`{% render "product" %}`)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := engine.GlobalVariableSegments(tpl)
		// render with no arguments → no globals from the outer scope needed
		// (product.name/price inside the partial come from render's isolated context)
		for _, seg := range got {
			if len(seg) > 0 && seg[0] == "product" {
				t.Errorf("product should not be global (render uses isolated scope), got %v", got)
			}
		}
	})

	t.Run("include missing template — graceful degradation", func(t *testing.T) {
		// If the partial doesn't exist in the store, analysis continues without it.
		// No error is returned; only the tag-level argument variables are reported.
		engine := NewEngine()
		engine.RegisterTemplateStore(&inMemoryStore{
			files: map[string]string{}, // empty store
		})
		tpl, err := engine.ParseString(`{% include "nonexistent" with source_var %}`)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := engine.GlobalVariableSegments(tpl)
		// source_var (the with-arg) should still be reported
		want := [][]string{{"source_var"}}
		if !segmentsEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
