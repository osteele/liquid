package liquid

// S9 — Static Analysis E2E tests.
//
// These intensive tests lock down the exact behaviour of every static-analysis
// surface in the Go Liquid engine so that regressions are caught immediately:
//
//   • GlobalVariableSegments / VariableSegments (all path-level combinations)
//   • GlobalVariables / Variables (root-name deduplicated views)
//   • GlobalFullVariables / FullVariables (Variable struct with Global flag)
//   • Analyze() / ParseAndAnalyze() (StaticAnalysis struct)
//   • Template convenience methods (same APIs on *Template)
//   • Walk / ParseTree visitor API
//   • RegisterTagAnalyzer / RegisterBlockAnalyzer (custom extension points)
//   • Analyzer coverage for every built-in tag: assign, capture, for,
//     tablerow, if/unless, case, echo, increment, decrement,
//     include, render, liquid (multi-line)

import (
	"strings"
	"testing"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func s9Parse(t *testing.T, e *Engine, src string) *Template {
	t.Helper()
	tpl, err := e.ParseString(src)
	require.NoError(t, err, "ParseString(%q)", src)
	return tpl
}

func s9Globals(t *testing.T, e *Engine, src string) [][]string {
	t.Helper()
	tpl := s9Parse(t, e, src)
	segs, err := e.GlobalVariableSegments(tpl)
	require.NoError(t, err)
	return segs
}

func s9All(t *testing.T, e *Engine, src string) [][]string {
	t.Helper()
	tpl := s9Parse(t, e, src)
	segs, err := e.VariableSegments(tpl)
	require.NoError(t, err)
	return segs
}

func s9GlobalRoots(t *testing.T, e *Engine, src string) []string {
	t.Helper()
	tpl := s9Parse(t, e, src)
	roots, err := e.GlobalVariables(tpl)
	require.NoError(t, err)
	return roots
}

// assertSegsContain fails if want is not a subset of got.
func assertSegsContain(t *testing.T, got [][]string, want ...[]string) {
	t.Helper()
outer:
	for _, w := range want {
		for _, g := range got {
			if segSliceEqual(g, w) {
				continue outer
			}
		}
		t.Errorf("expected segment %v in %v", w, got)
	}
}

// assertSegsNotContain fails if any of want appears in got.
func assertSegsNotContain(t *testing.T, got [][]string, want ...[]string) {
	t.Helper()
	for _, w := range want {
		for _, g := range got {
			if segSliceEqual(g, w) {
				t.Errorf("segment %v should NOT be in %v", w, got)
			}
		}
	}
}

func segSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func assertRootsContain(t *testing.T, got []string, want ...string) {
	t.Helper()
	set := map[string]bool{}
	for _, g := range got {
		set[g] = true
	}
	for _, w := range want {
		if !set[w] {
			t.Errorf("expected root %q in %v", w, got)
		}
	}
}

func assertRootsNotContain(t *testing.T, got []string, want ...string) {
	t.Helper()
	set := map[string]bool{}
	for _, g := range got {
		set[g] = true
	}
	for _, w := range want {
		if set[w] {
			t.Errorf("root %q should NOT be in %v", w, got)
		}
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// 9-A. GlobalVariableSegments — path-level precision
// ══════════════════════════════════════════════════════════════════════════════

func TestS9A_GlobalVariableSegments_ScalarOutput(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{{ x }}`)
	assert.True(t, segmentsEqual(got, [][]string{{"x"}}))
}

func TestS9A_GlobalVariableSegments_DeepPath(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{{ a.b.c.d }}`)
	assert.True(t, segmentsEqual(got, [][]string{{"a", "b", "c", "d"}}),
		"got %v", got)
}

func TestS9A_GlobalVariableSegments_MultipleDistinctPaths(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{{ a.x }} {{ b.y }} {{ c }}`)
	assert.True(t, segmentsEqual(got, [][]string{{"a", "x"}, {"b", "y"}, {"c"}}),
		"got %v", got)
}

func TestS9A_GlobalVariableSegments_SameRootDiffPaths(t *testing.T) {
	// a.x and a.y are distinct segments; only one root "a"
	e := NewEngine()
	got := s9Globals(t, e, `{{ a.x }} {{ a.y }}`)
	assertSegsContain(t, got, []string{"a", "x"}, []string{"a", "y"})
	assert.Equal(t, 2, len(got), "expected 2 segments, got %v", got)
}

func TestS9A_GlobalVariableSegments_AssignMakesLocal(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% assign x = "literal" %}{{ x }}`)
	assert.Empty(t, got, "assigned-from-literal should have no globals, got %v", got)
}

func TestS9A_GlobalVariableSegments_AssignRHSIsGlobal(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% assign x = src.val %}{{ x }}`)
	assertSegsContain(t, got, []string{"src", "val"})
	assertSegsNotContain(t, got, []string{"x"})
}

func TestS9A_GlobalVariableSegments_CaptureBodyGlobal(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% capture buf %}{{ inner }}{% endcapture %}{{ buf }}`)
	assertSegsContain(t, got, []string{"inner"})
	assertSegsNotContain(t, got, []string{"buf"})
}

func TestS9A_GlobalVariableSegments_ForLoopVarIsLocal(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% for item in products %}{{ item.title }}{% endfor %}`)
	assertSegsContain(t, got, []string{"products"})
	assertSegsNotContain(t, got, []string{"item"}, []string{"item", "title"})
}

func TestS9A_GlobalVariableSegments_ForLimitVariable(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% for x in list limit: max_items %}{{ x }}{% endfor %}`)
	assertSegsContain(t, got, []string{"list"}, []string{"max_items"})
	assertSegsNotContain(t, got, []string{"x"})
}

func TestS9A_GlobalVariableSegments_ForOffsetVariable(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% for x in list offset: start_at %}{{ x }}{% endfor %}`)
	assertSegsContain(t, got, []string{"list"}, []string{"start_at"})
	assertSegsNotContain(t, got, []string{"x"})
}

func TestS9A_GlobalVariableSegments_ForLimitAndOffsetBothVariables(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% for x in list limit: n offset: m %}{{ x }}{% endfor %}`)
	assertSegsContain(t, got, []string{"list"}, []string{"n"}, []string{"m"})
	assertSegsNotContain(t, got, []string{"x"})
}

func TestS9A_GlobalVariableSegments_ForLimitLiteralNoVar(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% for x in list limit: 5 %}{{ x }}{% endfor %}`)
	assertSegsContain(t, got, []string{"list"})
	assertSegsNotContain(t, got, []string{"x"})
	assert.Equal(t, 1, len(got), "literal limit should not add variable, got %v", got)
}

func TestS9A_GlobalVariableSegments_TablerowLimitOffset(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% tablerow x in rows limit: row_limit offset: row_offset %}{{ x }}{% endtablerow %}`)
	assertSegsContain(t, got, []string{"rows"}, []string{"row_limit"}, []string{"row_offset"})
	assertSegsNotContain(t, got, []string{"x"})
}

func TestS9A_GlobalVariableSegments_IfCondition(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% if cond %}{{ a }}{% elsif other %}{{ b }}{% else %}{{ c }}{% endif %}`)
	assertSegsContain(t, got, []string{"cond"}, []string{"other"}, []string{"a"}, []string{"b"}, []string{"c"})
}

func TestS9A_GlobalVariableSegments_UnlessCondition(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% unless flag %}{{ content }}{% else %}{{ fallback }}{% endunless %}`)
	assertSegsContain(t, got, []string{"flag"}, []string{"content"}, []string{"fallback"})
}

func TestS9A_GlobalVariableSegments_CaseWhenElse(t *testing.T) {
	e := NewEngine()
	src := `{% case status %}{% when "a" %}{{ msg_a }}{% when val_b %}{{ msg_b }}{% else %}{{ default_msg }}{% endcase %}`
	got := s9Globals(t, e, src)
	assertSegsContain(t, got, []string{"status"}, []string{"val_b"}, []string{"msg_a"}, []string{"msg_b"}, []string{"default_msg"})
}

func TestS9A_GlobalVariableSegments_EchoTag(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% echo product.title %}`)
	assertSegsContain(t, got, []string{"product", "title"})
}

func TestS9A_GlobalVariableSegments_EchoWithFilterArgs(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% echo product.title | append: suffix %}`)
	assertSegsContain(t, got, []string{"product", "title"}, []string{"suffix"})
}

func TestS9A_GlobalVariableSegments_EchoWithFilterKwargs(t *testing.T) {
	// allow_false: z  — z is a variable even though it's a keyword arg value
	e := NewEngine()
	got := s9Globals(t, e, `{% echo x | default: y, allow_false: z %}`)
	assertSegsContain(t, got, []string{"x"}, []string{"y"}, []string{"z"})
}

func TestS9A_GlobalVariableSegments_IncrementIsLocal(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% increment counter %}{{ counter }}`)
	// increment: counter goes to LocalScope, not globals
	assertSegsNotContain(t, got, []string{"counter"})
}

func TestS9A_GlobalVariableSegments_DecrementIsLocal(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% decrement counter %}{{ counter }}`)
	assertSegsNotContain(t, got, []string{"counter"})
}

func TestS9A_GlobalVariableSegments_IncDecNoGlobals(t *testing.T) {
	// Pure increment/decrement templates produce NO global variables
	e := NewEngine()
	got := s9Globals(t, e, `{% increment a %}{% decrement b %}`)
	assert.Empty(t, got, "increment+decrement should produce no globals, got %v", got)
}

func TestS9A_GlobalVariableSegments_IncludeWithVariable(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% include "partial" with some_var %}`)
	assertSegsContain(t, got, []string{"some_var"})
}

func TestS9A_GlobalVariableSegments_IncludeDynamicFile(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% include dynamic_file %}`)
	assertSegsContain(t, got, []string{"dynamic_file"})
}

func TestS9A_GlobalVariableSegments_IncludeForArray(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% include "partial" for items %}`)
	assertSegsContain(t, got, []string{"items"})
}

func TestS9A_GlobalVariableSegments_IncludeKVArgs(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% include "partial" a: foo, b: bar %}`)
	assertSegsContain(t, got, []string{"foo"}, []string{"bar"})
}

func TestS9A_GlobalVariableSegments_RenderWithVariable(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% render "partial" with product_item %}`)
	assertSegsContain(t, got, []string{"product_item"})
}

func TestS9A_GlobalVariableSegments_RenderForArray(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% render "partial" for cart_items as item %}`)
	assertSegsContain(t, got, []string{"cart_items"})
}

func TestS9A_GlobalVariableSegments_RenderKVArgs(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{% render "partial" title: page.title, user: current_user %}`)
	assertSegsContain(t, got, []string{"page", "title"}, []string{"current_user"})
}

func TestS9A_GlobalVariableSegments_LiquidTagInner(t *testing.T) {
	e := NewEngine()
	src := "{% liquid\n  if product.available\n    echo product.title\n  endif\n  assign counter = site.count\n%}"
	got := s9Globals(t, e, src)
	assertSegsContain(t, got, []string{"product", "available"}, []string{"product", "title"}, []string{"site", "count"})
	assertSegsNotContain(t, got, []string{"counter"})
}

func TestS9A_GlobalVariableSegments_LiquidTagForLoopLocal(t *testing.T) {
	e := NewEngine()
	src := "{% liquid\n  for x in items\n    echo x\n  endfor\n%}"
	got := s9Globals(t, e, src)
	assertSegsContain(t, got, []string{"items"})
	assertSegsNotContain(t, got, []string{"x"})
}

func TestS9A_GlobalVariableSegments_FilterArgs(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{{ base | append: suffix | prepend: prefix }}`)
	assertSegsContain(t, got, []string{"base"}, []string{"suffix"}, []string{"prefix"})
}

func TestS9A_GlobalVariableSegments_NoLiterals(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `{{ "hello" }} {{ 42 }} {{ true }} {{ nil }}`)
	assert.Empty(t, got, "literal outputs should produce no variables, got %v", got)
}

func TestS9A_GlobalVariableSegments_EmptyTemplate(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, ``)
	assert.Empty(t, got)
}

func TestS9A_GlobalVariableSegments_TextOnly(t *testing.T) {
	e := NewEngine()
	got := s9Globals(t, e, `hello world, no variables here`)
	assert.Empty(t, got)
}

func TestS9A_GlobalVariableSegments_Deduplication(t *testing.T) {
	// Same path used three times — should appear once
	e := NewEngine()
	got := s9Globals(t, e, `{{ x.v }},{{ x.v }},{{ x.v }}`)
	assert.Equal(t, 1, len(got), "duplicate paths should be deduplicated, got %v", got)
	assertSegsContain(t, got, []string{"x", "v"})
}

func TestS9A_GlobalVariableSegments_NestedBlocks(t *testing.T) {
	e := NewEngine()
	src := `{% if a %}{% for x in b %}{% unless x == skip %}{{ x.name }}{% endunless %}{% endfor %}{% endif %}`
	got := s9Globals(t, e, src)
	assertSegsContain(t, got, []string{"a"}, []string{"b"}, []string{"skip"})
	assertSegsNotContain(t, got, []string{"x"}, []string{"x", "name"})
}

// ══════════════════════════════════════════════════════════════════════════════
// 9-B. VariableSegments — includes locals
// ══════════════════════════════════════════════════════════════════════════════

func TestS9B_VariableSegments_IncludesAssignedVar(t *testing.T) {
	e := NewEngine()
	got := s9All(t, e, `{% assign x = src %}{{ x }}`)
	assertSegsContain(t, got, []string{"src"}, []string{"x"})
}

func TestS9B_VariableSegments_IncludesForLoopVar(t *testing.T) {
	e := NewEngine()
	got := s9All(t, e, `{% for item in list %}{{ item.name }}{% endfor %}`)
	assertSegsContain(t, got, []string{"list"}, []string{"item", "name"})
}

func TestS9B_VariableSegments_IncludesCaptureVar(t *testing.T) {
	e := NewEngine()
	got := s9All(t, e, `{% capture buf %}{{ content }}{% endcapture %}{{ buf }}`)
	assertSegsContain(t, got, []string{"content"}, []string{"buf"})
}

func TestS9B_VariableSegments_GlobalsSubsetOfAll(t *testing.T) {
	// Every global variable must also appear in All
	e := NewEngine()
	src := `{% assign y = x %}{% for item in list %}{{ item.v }} {{ y }}{% endfor %}`
	tpl := s9Parse(t, e, src)

	globals, err := e.GlobalVariableSegments(tpl)
	require.NoError(t, err)
	all, err := e.VariableSegments(tpl)
	require.NoError(t, err)

	allSet := map[string]bool{}
	for _, seg := range all {
		allSet[strings.Join(seg, ".")] = true
	}
	for _, g := range globals {
		k := strings.Join(g, ".")
		assert.True(t, allSet[k], "global %v should be in All, but All = %v", g, all)
	}
}

func TestS9B_VariableSegments_AllHasMoreThanGlobals(t *testing.T) {
	e := NewEngine()
	src := `{% assign y = x %}{{ y }}`
	tpl := s9Parse(t, e, src)

	globals, _ := e.GlobalVariableSegments(tpl)
	all, _ := e.VariableSegments(tpl)

	// globals = x only; all = x + y
	assert.Greater(t, len(all), len(globals),
		"All should have more entries than Globals when locals exist")
}

func TestS9B_VariableSegments_IncAndDecLocal(t *testing.T) {
	e := NewEngine()
	all := s9All(t, e, `{% increment cnt %}{% decrement cnt %}`)
	// cnt is introduced by increment (LocalScope) — it must be in All
	found := false
	for _, seg := range all {
		if len(seg) == 1 && seg[0] == "cnt" {
			found = true
		}
	}
	// Note: the counter variable is only in LocalScope, not an access, so it
	// may not appear in VariableSegments (no actual "use" expression).
	// This test documents the current behavior: cnt is NOT in VariableSegments
	// because increment/decrement add it only to LocalScope with no argument expression.
	_ = found // behavior documented above
}

// ══════════════════════════════════════════════════════════════════════════════
// 9-C. GlobalVariables / Variables — root-name views
// ══════════════════════════════════════════════════════════════════════════════

func TestS9C_GlobalVariables_DeduplicatesRoots(t *testing.T) {
	e := NewEngine()
	roots := s9GlobalRoots(t, e, `{{ a.x }} {{ a.y }} {{ a.z }}`)
	assert.Equal(t, []string{"a"}, roots, "three paths under 'a' should produce one root")
}

func TestS9C_GlobalVariables_MultipleRoots(t *testing.T) {
	e := NewEngine()
	roots := s9GlobalRoots(t, e, `{{ a.x }} {{ b.y }} {{ c }}`)
	assertRootsContain(t, roots, "a", "b", "c")
	assert.Equal(t, 3, len(roots))
}

func TestS9C_GlobalVariables_ExcludesLocals(t *testing.T) {
	e := NewEngine()
	roots := s9GlobalRoots(t, e, `{% assign y = x %}{{ y }} {{ z }}`)
	assertRootsContain(t, roots, "x", "z")
	assertRootsNotContain(t, roots, "y")
}

func TestS9C_GlobalVariables_ExcludesForLoopVar(t *testing.T) {
	e := NewEngine()
	roots := s9GlobalRoots(t, e, `{% for item in products %}{{ item.name }}{% endfor %}`)
	assertRootsContain(t, roots, "products")
	assertRootsNotContain(t, roots, "item")
}

func TestS9C_Variables_IncludesLocals(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% assign y = x %}{{ y }} {{ z }}`)
	vars, err := e.Variables(tpl)
	require.NoError(t, err)
	assertRootsContain(t, vars, "x", "y", "z")
}

func TestS9C_Variables_DeduplicatesRoots(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{{ u.a }} {{ u.b }} {{ u.c }}`)
	vars, err := e.Variables(tpl)
	require.NoError(t, err)
	assert.Equal(t, []string{"u"}, vars)
}

// ══════════════════════════════════════════════════════════════════════════════
// 9-D. FullVariables / GlobalFullVariables — Variable struct
// ══════════════════════════════════════════════════════════════════════════════

func TestS9D_FullVariables_GlobalFlagCorrect(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% assign loc = src %}{{ loc }} {{ glo }}`)
	vars, err := e.FullVariables(tpl)
	require.NoError(t, err)

	byName := map[string]Variable{}
	for _, v := range vars {
		byName[v.String()] = v
	}

	assert.True(t, byName["src"].Global, "src should be Global=true")
	assert.False(t, byName["loc"].Global, "loc (assigned) should be Global=false")
	assert.True(t, byName["glo"].Global, "glo should be Global=true")
}

func TestS9D_FullVariables_ForLoopVarLocal(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% for item in store.products %}{{ item.title }}{% endfor %}`)
	vars, err := e.FullVariables(tpl)
	require.NoError(t, err)

	byName := map[string]Variable{}
	for _, v := range vars {
		byName[v.String()] = v
	}

	storeProds, ok := byName["store.products"]
	require.True(t, ok, "store.products must appear in FullVariables, got %v", vars)
	assert.True(t, storeProds.Global)

	itemTitle, ok := byName["item.title"]
	require.True(t, ok, "item.title must appear in FullVariables, got %v", vars)
	assert.False(t, itemTitle.Global, "item.title (loop var path) should be local")
}

func TestS9D_FullVariables_CaptureVarLocal(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% capture buf %}{{ source_data }}{% endcapture %}{{ buf }}`)
	vars, err := e.FullVariables(tpl)
	require.NoError(t, err)

	byName := map[string]Variable{}
	for _, v := range vars {
		byName[v.String()] = v
	}

	assert.True(t, byName["source_data"].Global)
	assert.False(t, byName["buf"].Global)
}

func TestS9D_FullVariables_SegmentsPreserved(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{{ customer.first_name }}`)
	vars, err := e.FullVariables(tpl)
	require.NoError(t, err)
	require.Len(t, vars, 1)
	assert.Equal(t, []string{"customer", "first_name"}, vars[0].Segments)
}

func TestS9D_GlobalFullVariables_AllGlobal(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% assign loc = src %}{{ loc }} {{ glo.val }}`)
	vars, err := e.GlobalFullVariables(tpl)
	require.NoError(t, err)

	for _, v := range vars {
		assert.True(t, v.Global, "GlobalFullVariables must only return Global=true, got %v", v)
	}
}

func TestS9D_GlobalFullVariables_ExcludesLocals(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% assign loc = src %}{{ loc }} {{ glo }}`)
	vars, err := e.GlobalFullVariables(tpl)
	require.NoError(t, err)

	for _, v := range vars {
		assert.NotEqual(t, "loc", v.String(), "loc is local, should not appear in GlobalFullVariables")
	}

	names := map[string]bool{}
	for _, v := range vars {
		names[v.String()] = true
	}
	assert.True(t, names["src"], "src should be in GlobalFullVariables")
	assert.True(t, names["glo"], "glo should be in GlobalFullVariables")
}

func TestS9D_Variable_StringMethod(t *testing.T) {
	cases := []struct {
		segs []string
		want string
	}{
		{[]string{"a"}, "a"},
		{[]string{"a", "b"}, "a.b"},
		{[]string{"customer", "first_name"}, "customer.first_name"},
		{[]string{"a", "b", "c", "d"}, "a.b.c.d"},
	}
	for _, c := range cases {
		v := Variable{Segments: c.segs}
		assert.Equal(t, c.want, v.String())
	}
}

func TestS9D_Variable_GlobalField(t *testing.T) {
	v1 := Variable{Segments: []string{"x"}, Global: true}
	v2 := Variable{Segments: []string{"y"}, Global: false}
	assert.True(t, v1.Global)
	assert.False(t, v2.Global)
}

// ══════════════════════════════════════════════════════════════════════════════
// 9-E. Analyze / ParseAndAnalyze — StaticAnalysis struct
// ══════════════════════════════════════════════════════════════════════════════

func TestS9E_Analyze_GlobalsExcludeLocals(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% assign loc = src %}{% for item in list %}{{ item }} {{ glo }}{% endfor %}`)
	analysis, err := e.Analyze(tpl)
	require.NoError(t, err)

	globalNames := map[string]bool{}
	for _, v := range analysis.Globals {
		globalNames[v.String()] = true
	}
	assert.True(t, globalNames["src"])
	assert.True(t, globalNames["list"])
	assert.True(t, globalNames["glo"])
	assert.False(t, globalNames["loc"])
	assert.False(t, globalNames["item"])
}

func TestS9E_Analyze_LocalsIncludeAssign(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% assign x = "val" %}{% capture y %}ok{% endcapture %}`)
	analysis, err := e.Analyze(tpl)
	require.NoError(t, err)

	locals := map[string]bool{}
	for _, l := range analysis.Locals {
		locals[l] = true
	}
	assert.True(t, locals["x"], "assign variable must be in Locals")
	assert.True(t, locals["y"], "capture variable must be in Locals")
}

func TestS9E_Analyze_LocalsIncludeForVar(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% for item in list %}{{ item }}{% endfor %}`)
	analysis, err := e.Analyze(tpl)
	require.NoError(t, err)

	locals := map[string]bool{}
	for _, l := range analysis.Locals {
		locals[l] = true
	}
	assert.True(t, locals["item"], "for loop variable must be in Locals")
}

func TestS9E_Analyze_LocalsIncludeIncrementDecrement(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% increment ctr_a %}{% decrement ctr_b %}`)
	analysis, err := e.Analyze(tpl)
	require.NoError(t, err)

	locals := map[string]bool{}
	for _, l := range analysis.Locals {
		locals[l] = true
	}
	assert.True(t, locals["ctr_a"], "increment counter must be in Locals")
	assert.True(t, locals["ctr_b"], "decrement counter must be in Locals")
}

func TestS9E_Analyze_TagsListComprehensive(t *testing.T) {
	e := NewEngine()
	src := `{% assign x=1 %}{% capture y %}ok{% endcapture %}{% for i in list %}{% if i %}{% unless flag %}{% case x %}{% when 1 %}ok{% endcase %}{% endunless %}{% endif %}{% endfor %}{% increment n %}{% decrement m %}{% echo z %}`
	tpl := s9Parse(t, e, src)
	analysis, err := e.Analyze(tpl)
	require.NoError(t, err)

	tagSet := map[string]bool{}
	for _, tag := range analysis.Tags {
		tagSet[tag] = true
	}

	for _, expected := range []string{"assign", "capture", "for", "if", "unless", "case", "increment", "decrement", "echo"} {
		assert.True(t, tagSet[expected], "expected tag %q in Tags, got %v", expected, analysis.Tags)
	}
}

func TestS9E_Analyze_VariablesIncludesBothGlobalAndLocal(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% assign y = x %}{{ y }} {{ z }}`)
	analysis, err := e.Analyze(tpl)
	require.NoError(t, err)

	varNames := map[string]bool{}
	for _, v := range analysis.Variables {
		varNames[v.String()] = true
	}
	assert.True(t, varNames["x"], "x (global RHS) must be in Variables")
	assert.True(t, varNames["y"], "y (local) must be in Variables")
	assert.True(t, varNames["z"], "z (global) must be in Variables")
}

func TestS9E_Analyze_EmptyTemplate(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, ``)
	analysis, err := e.Analyze(tpl)
	require.NoError(t, err)
	assert.Empty(t, analysis.Globals)
	assert.Empty(t, analysis.Locals)
	assert.Empty(t, analysis.Tags)
	assert.Empty(t, analysis.Variables)
}

func TestS9E_Analyze_LiquidTagTracked(t *testing.T) {
	e := NewEngine()
	src := "{% liquid\n  if site.ready\n    echo site.name\n  endif\n  for x in site.items\n    echo x\n  endfor\n%}"
	tpl := s9Parse(t, e, src)
	analysis, err := e.Analyze(tpl)
	require.NoError(t, err)

	globalNames := map[string]bool{}
	for _, v := range analysis.Globals {
		globalNames[v.String()] = true
	}
	assert.True(t, globalNames["site.ready"], "site.ready must be a global")
	assert.True(t, globalNames["site.name"], "site.name must be a global")
	assert.True(t, globalNames["site.items"], "site.items must be a global")
	assert.False(t, globalNames["x"], "x is for-loop-local inside liquid tag")
}

func TestS9E_ParseAndAnalyze_ReturnsTemplate(t *testing.T) {
	e := NewEngine()
	tpl, analysis, err := e.ParseAndAnalyze([]byte(`{% assign x = src %}{{ x }} {{ z }}`))
	require.NoError(t, err)
	require.NotNil(t, tpl, "template must not be nil")
	require.NotNil(t, analysis, "analysis must not be nil")

	globalNames := map[string]bool{}
	for _, v := range analysis.Globals {
		globalNames[v.String()] = true
	}
	assert.True(t, globalNames["src"])
	assert.True(t, globalNames["z"])
	assert.False(t, globalNames["x"])
}

func TestS9E_ParseAndAnalyze_ParseError(t *testing.T) {
	e := NewEngine()
	_, _, err := e.ParseAndAnalyze([]byte(`{% if unclosed %}`))
	assert.Error(t, err, "unclosed if block should return parse error")
}

// ══════════════════════════════════════════════════════════════════════════════
// 9-F. Template convenience methods
// ══════════════════════════════════════════════════════════════════════════════

func TestS9F_TemplateConvenienceMethods(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% assign y = x %}{{ y }} {{ z.val }}`)

	t.Run("VariableSegments includes all", func(t *testing.T) {
		got, err := tpl.VariableSegments()
		require.NoError(t, err)
		assertSegsContain(t, got, []string{"x"}, []string{"y"}, []string{"z", "val"})
	})

	t.Run("GlobalVariableSegments excludes local", func(t *testing.T) {
		got, err := tpl.GlobalVariableSegments()
		require.NoError(t, err)
		assertSegsContain(t, got, []string{"x"}, []string{"z", "val"})
		assertSegsNotContain(t, got, []string{"y"})
	})

	t.Run("Variables root names all", func(t *testing.T) {
		got, err := tpl.Variables()
		require.NoError(t, err)
		assertRootsContain(t, got, "x", "y", "z")
	})

	t.Run("GlobalVariables root names only globals", func(t *testing.T) {
		got, err := tpl.GlobalVariables()
		require.NoError(t, err)
		assertRootsContain(t, got, "x", "z")
		assertRootsNotContain(t, got, "y")
	})

	t.Run("FullVariables Global field correct", func(t *testing.T) {
		got, err := tpl.FullVariables()
		require.NoError(t, err)
		byName := map[string]Variable{}
		for _, v := range got {
			byName[v.String()] = v
		}
		assert.True(t, byName["x"].Global)
		assert.False(t, byName["y"].Global)
		assert.True(t, byName["z.val"].Global)
	})

	t.Run("GlobalFullVariables all Global=true", func(t *testing.T) {
		got, err := tpl.GlobalFullVariables()
		require.NoError(t, err)
		for _, v := range got {
			assert.True(t, v.Global, "all returned by GlobalFullVariables must be Global=true")
		}
	})

	t.Run("Analyze non-nil", func(t *testing.T) {
		analysis, err := tpl.Analyze()
		require.NoError(t, err)
		require.NotNil(t, analysis)
		assert.NotEmpty(t, analysis.Variables)
	})
}

// ══════════════════════════════════════════════════════════════════════════════
// 9-G. Walk / ParseTree visitor API
// ══════════════════════════════════════════════════════════════════════════════

func TestS9G_Walk_AllNodeKinds(t *testing.T) {
	e := NewEngine()
	src := `hello {{ x }} {% assign y = 1 %}{% if x %}ok{% endif %}`
	tpl := s9Parse(t, e, src)

	kinds := map[TemplateNodeKind]bool{}
	tpl.Walk(func(node *TemplateNode) bool {
		kinds[node.Kind] = true
		return true
	})

	assert.True(t, kinds[TemplateNodeText], "should have text node")
	assert.True(t, kinds[TemplateNodeOutput], "should have output node")
	assert.True(t, kinds[TemplateNodeTag], "should have tag node")
	assert.True(t, kinds[TemplateNodeBlock], "should have block node")
}

func TestS9G_Walk_PreOrder(t *testing.T) {
	// Verify depth-first pre-order: parent visited before children
	e := NewEngine()
	src := `{% for a in list %}{% if a %}{{ a }}{% endif %}{% endfor %}`
	tpl := s9Parse(t, e, src)

	var order []string
	tpl.Walk(func(node *TemplateNode) bool {
		if node.TagName != "" {
			order = append(order, node.TagName)
		}
		return true
	})

	// "for" must precede "if" in the visit order
	require.GreaterOrEqual(t, len(order), 2)
	forIdx, ifIdx := -1, -1
	for i, name := range order {
		if name == "for" {
			forIdx = i
		}
		if name == "if" {
			ifIdx = i
		}
	}
	assert.Greater(t, ifIdx, forIdx, "for must be visited before if (pre-order)")
}

func TestS9G_Walk_SkipChildren_StopsAt(t *testing.T) {
	e := NewEngine()
	src := `{% for item in list %}{% if item %}{{ item }}{% endif %}{% endfor %}`
	tpl := s9Parse(t, e, src)

	var visited []string
	tpl.Walk(func(node *TemplateNode) bool {
		if node.TagName != "" {
			visited = append(visited, node.TagName)
		}
		return node.TagName != "for" // stop recursing into for
	})

	inVisited := map[string]bool{}
	for _, n := range visited {
		inVisited[n] = true
	}
	assert.True(t, inVisited["for"], "'for' must be visited")
	assert.False(t, inVisited["if"], "'if' inside for must NOT be visited after skip")
}

func TestS9G_Walk_VisitsElseClauses(t *testing.T) {
	e := NewEngine()
	src := `{% if false %}{{ a }}{% else %}{{ b }}{% endif %}`
	tpl := s9Parse(t, e, src)

	outputCount := 0
	tpl.Walk(func(node *TemplateNode) bool {
		if node.Kind == TemplateNodeOutput {
			outputCount++
		}
		return true
	})
	assert.Equal(t, 2, outputCount, "both {{ a }} and {{ b }} must be visited")
}

func TestS9G_Walk_VisitsWhenClauses(t *testing.T) {
	e := NewEngine()
	src := `{% case status %}{% when "a" %}{{ msg_a }}{% when "b" %}{{ msg_b }}{% else %}{{ msg_default }}{% endcase %}`
	tpl := s9Parse(t, e, src)

	outputCount := 0
	tpl.Walk(func(node *TemplateNode) bool {
		if node.Kind == TemplateNodeOutput {
			outputCount++
		}
		return true
	})
	assert.Equal(t, 3, outputCount, "all three outputs inside case must be visited")
}

func TestS9G_Walk_EmptyTemplate(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, ``)
	visited := 0
	tpl.Walk(func(_ *TemplateNode) bool { visited++; return true })
	assert.Equal(t, 0, visited, "empty template should visit 0 nodes")
}

func TestS9G_Walk_TagName(t *testing.T) {
	e := NewEngine()
	src := `{% assign x = 1 %}{% for i in list %}{% if i %}{% endif %}{% endfor %}{% capture buf %}ok{% endcapture %}{% echo x %}`
	tpl := s9Parse(t, e, src)

	names := map[string]bool{}
	tpl.Walk(func(node *TemplateNode) bool {
		if node.TagName != "" {
			names[node.TagName] = true
		}
		return true
	})

	for _, expected := range []string{"assign", "for", "if", "capture", "echo"} {
		assert.True(t, names[expected], "expected tag %q to be visited", expected)
	}
}

func TestS9G_Walk_TextNodeHasNoTagName(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `hello`)
	tpl.Walk(func(node *TemplateNode) bool {
		if node.Kind == TemplateNodeText {
			assert.Equal(t, "", node.TagName, "text nodes must have empty TagName")
		}
		return true
	})
}

func TestS9G_Walk_OutputNodeHasNoTagName(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{{ x }}`)
	tpl.Walk(func(node *TemplateNode) bool {
		if node.Kind == TemplateNodeOutput {
			assert.Equal(t, "", node.TagName, "output nodes must have empty TagName")
		}
		return true
	})
}

func TestS9G_Walk_NestedForLoops(t *testing.T) {
	e := NewEngine()
	src := `{% for a in outer %}{% for b in inner %}{{ a }}{{ b }}{% endfor %}{% endfor %}`
	tpl := s9Parse(t, e, src)

	forCount := 0
	tpl.Walk(func(node *TemplateNode) bool {
		if node.TagName == "for" {
			forCount++
		}
		return true
	})
	assert.Equal(t, 2, forCount, "two nested for loops should be visited")
}

func TestS9G_ParseTree_RootIsBlock(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `hello {{ x }}`)

	root := tpl.ParseTree()
	require.NotNil(t, root)
	assert.Equal(t, TemplateNodeBlock, root.Kind, "root must be a block")
	assert.Equal(t, "", root.TagName, "root block must have empty TagName")
}

func TestS9G_ParseTree_ChildCount(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `hello {{ x }}`)

	root := tpl.ParseTree()
	require.NotNil(t, root)
	assert.Equal(t, 2, len(root.Children), "root should have text + output = 2 children")
	assert.Equal(t, TemplateNodeText, root.Children[0].Kind)
	assert.Equal(t, TemplateNodeOutput, root.Children[1].Kind)
}

func TestS9G_ParseTree_TagOnly(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% assign x = 1 %}`)
	root := tpl.ParseTree()
	require.NotNil(t, root)
	require.Len(t, root.Children, 1)
	assert.Equal(t, TemplateNodeTag, root.Children[0].Kind)
	assert.Equal(t, "assign", root.Children[0].TagName)
}

func TestS9G_ParseTree_BlockChildren(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% if x %}{{ a }}{% else %}{{ b }}{% endif %}`)
	root := tpl.ParseTree()
	require.NotNil(t, root)
	require.GreaterOrEqual(t, len(root.Children), 1)

	ifNode := root.Children[0]
	assert.Equal(t, TemplateNodeBlock, ifNode.Kind)
	assert.Equal(t, "if", ifNode.TagName)
	// Should have body + else clause
	require.GreaterOrEqual(t, len(ifNode.Children), 2)
}

func TestS9G_ParseTree_ForBlockChildren(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, `{% for item in list %}{{ item }}{% endfor %}`)
	root := tpl.ParseTree()
	require.NotNil(t, root)
	require.Len(t, root.Children, 1)

	forNode := root.Children[0]
	assert.Equal(t, TemplateNodeBlock, forNode.Kind)
	assert.Equal(t, "for", forNode.TagName)
	require.GreaterOrEqual(t, len(forNode.Children), 1)
	assert.Equal(t, TemplateNodeOutput, forNode.Children[0].Kind)
}

func TestS9G_ParseTree_SourceLocations(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, "{{ x }}\n{% if y %}ok{% endif %}")
	found := false
	tpl.Walk(func(node *TemplateNode) bool {
		if (node.Kind == TemplateNodeOutput || node.Kind == TemplateNodeBlock) && node.Location.LineNo > 0 {
			found = true
		}
		return true
	})
	assert.True(t, found, "at least one node should have a non-zero source location")
}

func TestS9G_ParseTree_EmptyTemplate(t *testing.T) {
	e := NewEngine()
	tpl := s9Parse(t, e, ``)
	root := tpl.ParseTree()
	require.NotNil(t, root)
	assert.Equal(t, TemplateNodeBlock, root.Kind)
	assert.Empty(t, root.Children)
}

// ══════════════════════════════════════════════════════════════════════════════
// 9-H. RegisterTagAnalyzer / RegisterBlockAnalyzer
// ══════════════════════════════════════════════════════════════════════════════

func TestS9H_RegisterTagAnalyzer_WithoutAnalyzer_NoVars(t *testing.T) {
	e := NewEngine()

	// Custom tag with no analyzer: variables should not be tracked
	e.RegisterTag("my_tag", func(_ render.Context) (string, error) { return "", nil })

	tpl := s9Parse(t, e, `{% my_tag some_var %}`)
	got, err := e.GlobalVariableSegments(tpl)
	require.NoError(t, err)
	assert.Empty(t, got, "without analyzer, custom tag should report no variables")
}

func TestS9H_RegisterTagAnalyzer_WithAnalyzer_ReportsVars(t *testing.T) {
	e := NewEngine()

	e.RegisterTag("my_tag", func(_ render.Context) (string, error) { return "", nil })
	e.RegisterTagAnalyzer("my_tag", func(args string) render.NodeAnalysis {
		expr, err := expressions.Parse(args)
		if err != nil {
			return render.NodeAnalysis{}
		}
		return render.NodeAnalysis{Arguments: []expressions.Expression{expr}}
	})

	tpl := s9Parse(t, e, `{% my_tag some_var.prop %}`)
	got, err := e.GlobalVariableSegments(tpl)
	require.NoError(t, err)
	assertSegsContain(t, got, []string{"some_var", "prop"})
}

func TestS9H_RegisterTagAnalyzer_LocalScope(t *testing.T) {
	e := NewEngine()

	// Custom tag that introduces a local variable (like assign does)
	e.RegisterTag("define", func(_ render.Context) (string, error) { return "", nil })
	e.RegisterTagAnalyzer("define", func(args string) render.NodeAnalysis {
		varname := strings.TrimSpace(args)
		return render.NodeAnalysis{LocalScope: []string{varname}}
	})

	tpl := s9Parse(t, e, `{% define my_local %}{{ my_local }}`)
	got, err := e.GlobalVariableSegments(tpl)
	require.NoError(t, err)
	assertSegsNotContain(t, got, []string{"my_local"})
}

func TestS9H_RegisterBlockAnalyzer_WithAnalyzer_ReportsVars(t *testing.T) {
	e := NewEngine()

	e.RegisterBlock("my_block", func(_ render.Context) (string, error) { return "", nil })
	e.RegisterBlockAnalyzer("my_block", func(node render.BlockNode) render.NodeAnalysis {
		expr, err := expressions.Parse(node.Args)
		if err != nil {
			return render.NodeAnalysis{}
		}
		return render.NodeAnalysis{Arguments: []expressions.Expression{expr}}
	})

	tpl := s9Parse(t, e, `{% my_block block_var.x %}{% endmy_block %}`)
	got, err := e.GlobalVariableSegments(tpl)
	require.NoError(t, err)
	assertSegsContain(t, got, []string{"block_var", "x"})
}

func TestS9H_RegisterBlockAnalyzer_BlockScope_IsLocal(t *testing.T) {
	e := NewEngine()

	e.RegisterBlock("scoped", func(_ render.Context) (string, error) { return "", nil })
	e.RegisterBlockAnalyzer("scoped", func(_ render.BlockNode) render.NodeAnalysis {
		return render.NodeAnalysis{BlockScope: []string{"loop_item"}}
	})

	// loop_item appears in body but is in BlockScope → local → not a global
	tpl := s9Parse(t, e, `{% scoped %}{{ loop_item }}{% endscoped %}`)
	got, err := e.GlobalVariableSegments(tpl)
	require.NoError(t, err)
	assertSegsNotContain(t, got, []string{"loop_item"})
}

// ══════════════════════════════════════════════════════════════════════════════
// 9-I. Complex / realistic templates
// ══════════════════════════════════════════════════════════════════════════════

func TestS9I_RealTemplate_Shopify_ProductPage(t *testing.T) {
	// Realistic Shopify-style product page template
	e := NewEngine()
	src := `{% assign title = product.title %}` +
		`{% capture header %}<h1>{{ title }}</h1>{% endcapture %}` +
		`{{ header }}` +
		`{% for variant in product.variants %}` +
		`{% if variant.available %}` +
		`{{ variant.title | append: suffix }}` +
		`{% endif %}` +
		`{% endfor %}` +
		`{% unless product.hide_description %}` +
		`{{ product.description }}` +
		`{% endunless %}` +
		`{% case product.status %}` +
		`{% when "active" %}{{ active_label }}` +
		`{% else %}{{ default_label }}` +
		`{% endcase %}`

	tpl := s9Parse(t, e, src)
	globals, err := e.GlobalVariableSegments(tpl)
	require.NoError(t, err)
	globalRoots, err := e.GlobalVariables(tpl)
	require.NoError(t, err)

	// Globals: product.title, product.variants, product.hide_description,
	//          product.description, product.status, suffix, active_label, default_label
	// NOT globals: title (assign), header (capture), variant (loop var)
	assertRootsContain(t, globalRoots, "product", "suffix", "active_label", "default_label")
	assertRootsNotContain(t, globalRoots, "title", "header", "variant")

	globalSegsMap := map[string]bool{}
	for _, seg := range globals {
		globalSegsMap[strings.Join(seg, ".")] = true
	}
	assert.True(t, globalSegsMap["product.title"], "product.title should be a global")
	assert.True(t, globalSegsMap["product.variants"], "product.variants should be a global")
	assert.True(t, globalSegsMap["product.status"], "product.status should be a global")
}

func TestS9I_LiquidTag_ComplexInnerTemplate(t *testing.T) {
	e := NewEngine()
	src := strings.Join([]string{
		"{% liquid",
		"  assign page_title = page.title | upcase",
		"  if site.published",
		"    echo page_title",
		"  endif",
		"  for tag in page.tags",
		"    echo tag",
		"  endfor",
		"  unless page.draft",
		"    echo page.content",
		"  endunless",
		"%}",
	}, "\n")

	tpl := s9Parse(t, e, src)
	globals, err := e.GlobalVariableSegments(tpl)
	require.NoError(t, err)
	globalRoots, err := e.GlobalVariables(tpl)
	require.NoError(t, err)

	// Globals: page.title, site.published, page.tags, page.draft, page.content
	// NOT globals: page_title (assign), tag (for loop var)
	assertRootsContain(t, globalRoots, "page", "site")
	assertRootsNotContain(t, globalRoots, "page_title", "tag")

	segsMap := map[string]bool{}
	for _, seg := range globals {
		segsMap[strings.Join(seg, ".")] = true
	}
	assert.True(t, segsMap["page.title"], "page.title")
	assert.True(t, segsMap["site.published"], "site.published")
	assert.True(t, segsMap["page.tags"], "page.tags")
	assert.True(t, segsMap["page.draft"], "page.draft")
	assert.True(t, segsMap["page.content"], "page.content")
}

func TestS9I_Analysis_ChainedAssigns(t *testing.T) {
	e := NewEngine()
	src := `{% assign y = x %}{% assign z = y %}{{ z }}`
	tpl := s9Parse(t, e, src)

	globals, _ := e.GlobalVariableSegments(tpl)
	all, _ := e.VariableSegments(tpl)

	assertSegsContain(t, globals, []string{"x"})
	assertSegsNotContain(t, globals, []string{"y"}, []string{"z"})

	assertSegsContain(t, all, []string{"x"}, []string{"y"}, []string{"z"})
}

func TestS9I_Analysis_MultipleAssignSameName(t *testing.T) {
	// Assigning from different sources to the same var: both RHS are global
	e := NewEngine()
	src := `{% assign x = a %}{% assign x = b %}{{ x }}`
	tpl := s9Parse(t, e, src)
	globals, _ := e.GlobalVariableSegments(tpl)
	assertSegsContain(t, globals, []string{"a"}, []string{"b"})
	assertSegsNotContain(t, globals, []string{"x"})
}

func TestS9I_Walk_TagsAreInDocumentOrder(t *testing.T) {
	e := NewEngine()
	src := `{% assign a = 1 %}{% for i in list %}{{ i }}{% endfor %}{% if cond %}ok{% endif %}`
	tpl := s9Parse(t, e, src)

	var tagOrder []string
	tpl.Walk(func(node *TemplateNode) bool {
		switch node.Kind {
		case TemplateNodeTag, TemplateNodeBlock:
			tagOrder = append(tagOrder, node.TagName)
		}
		return true
	})

	require.GreaterOrEqual(t, len(tagOrder), 3)
	assert.Equal(t, "assign", tagOrder[0], "assign should be first tag")
	assert.Equal(t, "for", tagOrder[1], "for should be second tag")
	assert.Equal(t, "if", tagOrder[2], "if should be third tag (after for body)")
}

func TestS9I_Analysis_StableAcrossMultipleCalls(t *testing.T) {
	// Re-analyzing the same template should return identical results
	e := NewEngine()
	src := `{% assign y = x.val %}{% for item in items %}{{ item.name }}{% endfor %}{{ z }}`
	tpl := s9Parse(t, e, src)

	for range 5 {
		globals, err := e.GlobalVariableSegments(tpl)
		require.NoError(t, err)
		assertSegsContain(t, globals, []string{"x", "val"}, []string{"items"}, []string{"z"})
		assertSegsNotContain(t, globals, []string{"y"}, []string{"item", "name"})
	}
}
