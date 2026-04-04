package liquid

// Ported analysis tests from:
//   - Ruby Liquid: test/unit/parse_tree_visitor_test.rb
//   - LiquidJS:    test/integration/static_analysis/variables.spec.ts
//   - LiquidJS:    test/e2e/parse-and-analyze.spec.ts
//   - LiquidJS:    src/template/analysis.spec.ts

import (
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
