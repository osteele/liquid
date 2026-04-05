package liquid_test

// Tests for Section 7 — Context / Escopo
//
// Ported from / inspired by:
//   Ruby: liquid/test/integration/context_test.rb
//   JS:   liquidjs/src/context/context.spec.ts

import (
	"fmt"
	"io"
	"maps"
	"testing"

	"github.com/osteele/liquid"
	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 7.1 — Stack de escopos, get/set variáveis
// ---------------------------------------------------------------------------

// scopeTests covers variable access patterns ported from Ruby context_test.rb.
var scopeTests = []struct {
	in, expected string
}{
	// Non-existing variable compares equal to nil
	// Ruby: test_variables_not_existing
	{`{% if does_not_exist == nil %}true{% endif %}`, "true"},

	// Basic types
	{`{{ str }}`, "hello"},
	{`{{ num }}`, "42"},
	{`{% if bool_true == true %}pass{% endif %}`, "pass"},
	{`{% if bool_false == false %}pass{% endif %}`, "pass"},

	// Hierarchical data — dot and bracket notation
	// Ruby: test_hierachical_data
	{`{{ hash.name }}`, "tobi"},
	{`{{ hash["name"] }}`, "tobi"},

	// Keyword comparisons (bool literals)
	// Ruby: test_keywords
	{`{% if true == expect_true %}pass{% endif %}`, "pass"},
	{`{% if false == expect_false %}pass{% endif %}`, "pass"},

	// Numeric comparisons
	// Ruby: test_digits
	{`{% if 100 == expect_100 %}pass{% endif %}`, "pass"},
	{`{% if 100.00 == expect_float %}pass{% endif %}`, "pass"},

	// Array notation
	// Ruby: test_array_notation
	{`{{ test[0] }}`, "a"},
	{`{{ test[1] }}`, "b"},
	{`{% if test[2] == nil %}pass{% endif %}`, "pass"},

	// Recursive array / object notation
	// Ruby: test_recoursive_array_notation
	{`{{ nested.test[0] }}`, "1"},
	{`{{ arr_of_obj[0].test }}`, "worked"},

	// array.first / array.last / array.size
	// Ruby: test_try_first
	{`{{ numbers.first }}`, "1"},
	{`{% if numbers.last == 5 %}pass{% endif %}`, "pass"},
	{`{{ numbers.size }}`, "5"},

	// Hashes with bracket notation
	// Ruby: test_access_hashes_with_hash_notation
	{`{{ products["count"] }}`, "5"},
	{`{{ products["tags"][0] }}`, "deepsnow"},
	{`{{ products["tags"].first }}`, "deepsnow"},

	// full bracket chain access on product
	// Ruby: test_access_hashes_with_hash_notation (extended)
	{`{{ product["variants"][0]["title"] }}`, "draft151cm"},
	{`{{ product["variants"][1]["title"] }}`, "element151cm"},
	{`{{ product["variants"].first["title"] }}`, "draft151cm"},
	{`{{ product["variants"].last["title"] }}`, "element151cm"},

	// first in middle of call chain
	// Ruby: test_first_can_appear_in_middle_of_callchain
	{`{{ product.variants[0].title }}`, "draft151cm"},
	{`{{ product.variants.first.title }}`, "draft151cm"},
	{`{{ product.variants.last.title }}`, "element151cm"},

	// assign persists across statements
	{`{% assign x = 1 %}{{ x }}{% assign x = 2 %}{{ x }}`, "12"},

	// Outer-scope variable visible inside for loop body
	{`{% assign x = "outer" %}{% for i in (1..2) %}{{ x }}{% endfor %}`, "outerouter"},

	// For-loop variable is restored to nil after the loop ends
	// (the loop variable `i` was nil before, so it is nil again after)
	{`{% for i in (1..3) %}{{ i }}{% endfor %}{{ i }}`, "123"},

	// assign inside for loop persists after loop (Liquid semantics: assign is top-scope)
	{`{% for i in (1..3) %}{% assign last = i %}{% endfor %}{{ last }}`, "3"},

	// Strings: single and double quoted
	// Ruby: test_strings
	{`{{ "hello!" }}`, "hello!"},
	{`{{ 'hello!' }}`, "hello!"},

	// Hash-to-array transition: hash whose value is a sub-array
	// Ruby: test_hash_to_array_transition
	{`{{ colors.Blue[0] }}`, "003366"},
	{`{{ colors.Red[3] }}`, "FF9999"},

	// Array size (#length_query for array)
	// Ruby: test_length_query
	{`{% if seq4.size == 4 %}true{% endif %}`, "true"},

	// Map size (#length_query for hash)
	{`{% if map4.size == 4 %}true{% endif %}`, "true"},

	// Explicit 'size' key in a hash overrides the computed size
	// Ruby: test_length_query (third case)
	{`{% if explicit_size.size == 1000 %}true{% endif %}`, "true"},

	// Hyphenated variable name
	// Ruby: test_hyphenated_variable
	{`{{ oh-my }}`, "godz"},

	// Hash-notation is array-index for arrays, but NOT for bracket-string lookup
	// Array: array.first works, array["first"] is nil
	// Ruby: test_hash_notation_only_for_hash_access
	{`{{ numbers_arr.first }}`, "1"},
	{`{% if numbers_arr["first"] == nil %}pass{% endif %}`, "pass"},
	{`{{ hash_first["first"] }}`, "Hello"},

	// Dynamic property access via variable key
	// Ruby: test_access_hashes_with_hash_access_variables
	{`{{ products[var].first }}`, "deepsnow"},
	{`{{ products[nested.var].last }}`, "freestyle"},

	// String size (JS: should return string length)
	// LiquidJS: ctx.get(['foo', 'size']) → 3
	{`{{ str.size }}`, "5"},

	// Array size via .size on named variable
	{`{{ test.size }}`, "2"},
}

// scopeTestBindingsExtra holds the extra bindings needed by the new scope tests.
var scopeTestBindingsExtra = map[string]any{
	"colors": map[string]any{
		"Blue": []string{"003366", "336699", "6699CC", "99CCFF"},
		"Red":  []string{"660000", "993333", "CC6666", "FF9999"},
	},
	"seq4":          []int{1, 2, 3, 4},
	"map4":          map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
	"explicit_size": map[string]any{"a": 1, "size": 1000},
	"oh-my":         "godz",
	"numbers_arr":   []int{1, 2, 3, 4, 5},
	"hash_first":    map[string]any{"first": "Hello"},
	"var":           "tags",
	// NOTE: "nested" is intentionally omitted here; the base scopeTestBindings already
	// defines "nested" with a "test" key; we add the "var" key to the merged map in the test.
}

var scopeTestBindings = map[string]any{
	"str":        "hello",
	"num":        42,
	"bool_true":  true,
	"bool_false": false,
	"hash":       map[string]any{"name": "tobi"},
	"test":       []string{"a", "b"},
	"numbers":    []int{1, 2, 3, 4, 5},
	"nested": map[string]any{
		"test": []int{1, 2, 3, 4, 5},
		// "var" is needed for: {{ products[nested.var].last }}
		"var": "tags",
	},
	"arr_of_obj": []map[string]any{
		{"test": "worked"},
	},
	"products": map[string]any{
		"count": 5,
		"tags":  []string{"deepsnow", "freestyle"},
	},
	"product": map[string]any{
		"variants": []map[string]any{
			{"title": "draft151cm"},
			{"title": "element151cm"},
		},
	},
	"expect_true":  true,
	"expect_false": false,
	"expect_100":   100,
	"expect_float": 100.00,
}

// scopeTestAllBindings merges scopeTestBindings with scopeTestBindingsExtra.
// The merged map is used by TestScopeStack_GetSet.
func scopeTestAllBindings() map[string]any {
	merged := make(map[string]any, len(scopeTestBindings)+len(scopeTestBindingsExtra))
	maps.Copy(merged, scopeTestBindings)
	maps.Copy(merged, scopeTestBindingsExtra)
	return merged
}

func TestScopeStack_GetSet(t *testing.T) {
	engine := liquid.NewEngine()
	bindings := scopeTestAllBindings()

	for i, test := range scopeTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			out, err := engine.ParseAndRenderString(test.in, bindings)
			require.NoErrorf(t, err, "template: %s", test.in)
			require.Equalf(t, test.expected, out, "template: %s", test.in)
		})
	}
}

// ---------------------------------------------------------------------------
// 7.2 — Sub-contexto isolado (SpawnIsolated / RenderFileIsolated)
// ---------------------------------------------------------------------------

// TestIsolatedSubcontext_DoesNotInheritParentBindings verifies that
// RenderFileIsolated passes an empty parent scope — the sub-template cannot
// read variables from the calling context.
// Ported from Ruby: test_new_isolated_subcontext_does_not_inherit_variables
func TestIsolatedSubcontext_DoesNotInheritParentBindings(t *testing.T) {
	engine := liquid.NewEngine()

	// Register the tag first — before the engine is frozen by the first parse.
	engine.RegisterTag("render_isolated_partial", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("partial_isolation.html", map[string]any{
			"explicit_var": "passed",
		})
	})

	// Cache a partial that tries to read `parent_var` (which lives only in the
	// calling context) and `explicit_var` (which will be passed explicitly).
	_, err := engine.ParseTemplateAndCache(
		[]byte(`parent={{ parent_var }} explicit={{ explicit_var }}`),
		"partial_isolation.html",
		1,
	)
	require.NoError(t, err)

	out, err := engine.ParseAndRenderString(
		`{% render_isolated_partial %}`,
		map[string]any{
			"parent_var": "should_not_be_visible",
		},
	)
	require.NoError(t, err)
	// parent_var must be empty (not inherited); explicit_var must be visible.
	require.Equal(t, "parent= explicit=passed", out)
}

// TestIsolatedSubcontext_GlobalsPropagateToIsolatedContext verifies that
// engine globals ARE visible in isolated sub-contexts even though regular
// bindings are not inherited.
// Ported from Ruby: test_new_isolated_subcontext_inherits_static_environment
func TestIsolatedSubcontext_GlobalsPropagateToIsolatedContext(t *testing.T) {
	engine := liquid.NewEngine()
	engine.SetGlobals(map[string]any{
		"site_name": "MySite",
	})

	engine.RegisterTag("render_isolated_globals", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("partial_globals.html", map[string]any{})
	})

	// Cache a partial that reads the global `site_name`.
	_, err := engine.ParseTemplateAndCache(
		[]byte(`site={{ site_name }}`),
		"partial_globals.html",
		1,
	)
	require.NoError(t, err)

	out, err := engine.ParseAndRenderString(
		`{% render_isolated_globals %}`,
		map[string]any{},
	)
	require.NoError(t, err)
	require.Equal(t, "site=MySite", out)
}

// TestIsolatedSubcontext_ExplicitBindingsVisibleInsideIsolated verifies that
// bindings explicitly passed to RenderFileIsolated are visible in the partial.
func TestIsolatedSubcontext_ExplicitBindingsVisible(t *testing.T) {
	engine := liquid.NewEngine()

	engine.RegisterTag("render_with_product", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("partial_product.html", map[string]any{
			"product": map[string]any{"title": "Widget"},
		})
	})

	_, err := engine.ParseTemplateAndCache(
		[]byte(`{{ product.title }}`),
		"partial_product.html",
		1,
	)
	require.NoError(t, err)

	out, err := engine.ParseAndRenderString(
		`{% render_with_product %}`,
		map[string]any{},
	)
	require.NoError(t, err)
	require.Equal(t, "Widget", out)
}

// TestIsolatedSubcontext_GlobalScopeBindingWins verifies that an explicit
// binding passed to RenderFileIsolated shadows a global with the same key.
func TestIsolatedSubcontext_ExplicitBindingWinsOverGlobal(t *testing.T) {
	engine := liquid.NewEngine()
	engine.SetGlobals(map[string]any{
		"color": "global-blue",
	})

	engine.RegisterTag("render_color_partial", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("partial_color.html", map[string]any{
			"color": "local-red",
		})
	})

	_, err := engine.ParseTemplateAndCache(
		[]byte(`{{ color }}`),
		"partial_color.html",
		1,
	)
	require.NoError(t, err)

	out, err := engine.ParseAndRenderString(
		`{% render_color_partial %}`,
		map[string]any{},
	)
	require.NoError(t, err)
	require.Equal(t, "local-red", out)
}

// ---------------------------------------------------------------------------
// 7.3 — Registers (estado interno de tags)
// ---------------------------------------------------------------------------

// TestRegisters_StatePersistedWithinRender verifies that custom tags can
// store state in the rendering context (using ctx.Set / ctx.Get) and that
// this state is shared across tag invocations within the same render call.
func TestRegisters_StatePersistedWithinRender(t *testing.T) {
	engine := liquid.NewEngine()

	// A stateful counter tag: counts how many times it has been called
	// within a single render by storing state in the context bindings
	// under a reserved key.
	engine.RegisterTag("counter_tag", func(c render.Context) (string, error) {
		const key = ".counter_register"
		n := 0
		if v := c.Get(key); v != nil {
			n = v.(int)
		}
		n++
		c.Set(key, n)
		return fmt.Sprintf("%d", n), nil
	})

	// Three invocations in the same render must count 1, 2, 3.
	out, err := engine.ParseAndRenderString(
		`{% counter_tag %}{% counter_tag %}{% counter_tag %}`,
		map[string]any{},
	)
	require.NoError(t, err)
	require.Equal(t, "123", out)
}

// TestRegisters_StateResetBetweenRenders verifies that register state does
// NOT bleed from one render call into the next.
func TestRegisters_StateResetBetweenRenders(t *testing.T) {
	engine := liquid.NewEngine()

	engine.RegisterTag("counter_tag2", func(c render.Context) (string, error) {
		const key = ".counter_register2"
		n := 0
		if v := c.Get(key); v != nil {
			n = v.(int)
		}
		n++
		c.Set(key, n)
		return fmt.Sprintf("%d", n), nil
	})

	tpl := `{% counter_tag2 %}{% counter_tag2 %}`

	// First render: expect "12"
	out1, err := engine.ParseAndRenderString(tpl, map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "12", out1)

	// Second render: must also start from 1, not carry over state
	out2, err := engine.ParseAndRenderString(tpl, map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "12", out2)
}

// TestRegisters_CycleTagUsesInternalState verifies that the built-in cycle
// tag (which relies on context state) works correctly — this exercises the
// registers concept end-to-end using a built-in feature.
func TestRegisters_CycleTagState(t *testing.T) {
	engine := liquid.NewEngine()

	out, err := engine.ParseAndRenderString(
		`{% for i in (1..6) %}{% cycle "one", "two", "three" %}{% endfor %}`,
		map[string]any{},
	)
	require.NoError(t, err)
	require.Equal(t, "onetwothreeonetwothree", out)
}

// TestRegisters_CycleTagNamedGroups verifies that named cycle groups
// maintain independent counters.
func TestRegisters_CycleTagNamedGroups(t *testing.T) {
	engine := liquid.NewEngine()

	out, err := engine.ParseAndRenderString(
		`{% for i in (1..3) %}{% cycle "group1": "a", "b" %}{% cycle "group2": "x", "y" %}{% endfor %}`,
		map[string]any{},
	)
	require.NoError(t, err)
	require.Equal(t, "axbyax", out)
}

// ---------------------------------------------------------------------------
// 7.4 — Variáveis globais separadas do escopo
// ---------------------------------------------------------------------------

// TestGlobals_AccessibleInTemplate verifies that globals set on the engine
// are accessible inside a rendered template.
func TestGlobals_AccessibleInTemplate(t *testing.T) {
	engine := liquid.NewEngine()
	engine.SetGlobals(map[string]any{
		"site": map[string]any{"name": "MySite"},
	})

	out, err := engine.ParseAndRenderString(`{{ site.name }}`, map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "MySite", out)
}

// TestGlobals_ScopeBindingWinsOverGlobal verifies that a binding passed at
// render time takes precedence over a global with the same key.
// Ruby: test_static_environments_are_read_with_lower_priority_than_environments
func TestGlobals_ScopeBindingWinsOverGlobal(t *testing.T) {
	engine := liquid.NewEngine()
	engine.SetGlobals(map[string]any{
		"color":      "global-blue",
		"unshadowed": "from-global",
	})

	out, err := engine.ParseAndRenderString(
		`{{ color }} {{ unshadowed }}`,
		map[string]any{
			"color": "local-red", // shadows the global
		},
	)
	require.NoError(t, err)
	require.Equal(t, "local-red from-global", out)
}

// TestGlobals_MultipleGlobals verifies that multiple global variables are
// all accessible.
func TestGlobals_MultipleGlobals(t *testing.T) {
	engine := liquid.NewEngine()
	engine.SetGlobals(map[string]any{
		"a": "alpha",
		"b": "beta",
		"c": "gamma",
	})

	out, err := engine.ParseAndRenderString(`{{ a }}-{{ b }}-{{ c }}`, map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "alpha-beta-gamma", out)
}

// TestGlobals_AssignDoesNotPersistAcrossRenders verifies that an assign tag
// modifies the local binding for that render call only — subsequent renders
// still see the original global value.
func TestGlobals_AssignDoesNotPersistAcrossRenders(t *testing.T) {
	engine := liquid.NewEngine()
	engine.SetGlobals(map[string]any{
		"x": "global",
	})

	// Assign shadows the global within ONE render.
	out1, err := engine.ParseAndRenderString(
		`{% assign x = "local" %}{{ x }}`,
		map[string]any{},
	)
	require.NoError(t, err)
	require.Equal(t, "local", out1)

	// A fresh render must still see the global value.
	out2, err := engine.ParseAndRenderString(`{{ x }}`, map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "global", out2)
}

// TestGlobals_GetGlobals verifies that GetGlobals returns exactly what was
// set via SetGlobals.
func TestGlobals_GetGlobals(t *testing.T) {
	engine := liquid.NewEngine()

	require.Nil(t, engine.GetGlobals(), "GetGlobals should be nil before SetGlobals")

	globals := map[string]any{"key": "value", "num": 42}
	engine.SetGlobals(globals)

	got := engine.GetGlobals()
	require.Equal(t, globals, got)
}

// TestGlobals_EmptyBindingsWithGlobals verifies that globals are the only
// available variables when the bindings map provided to Render is empty.
func TestGlobals_EmptyBindingsWithGlobals(t *testing.T) {
	engine := liquid.NewEngine()
	engine.SetGlobals(map[string]any{"greeting": "hello"})

	out, err := engine.ParseAndRenderString(`{{ greeting }} world`, map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "hello world", out)
}

// TestGlobals_NilBindingsFallbackToGlobals verifies that passing nil
// for bindings still exposes globals.
func TestGlobals_NilBindingsFallbackToGlobals(t *testing.T) {
	engine := liquid.NewEngine()
	engine.SetGlobals(map[string]any{"greeting": "hello"})

	out, err := engine.ParseAndRenderString(`{{ greeting }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "hello", out)
}

// TestGlobals_AccessibleViaCustomTag verifies that globals are visible to
// custom tags through ctx.Get.
func TestGlobals_AccessibleViaCustomTag(t *testing.T) {
	engine := liquid.NewEngine()
	engine.SetGlobals(map[string]any{"global_key": "global_value"})

	engine.RegisterTag("read_global", func(c render.Context) (string, error) {
		v := c.Get("global_key")
		if v == nil {
			return "<nil>", nil
		}
		return fmt.Sprintf("%v", v), nil
	})

	out, err := engine.ParseAndRenderString(`{% read_global %}`, map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "global_value", out)
}

// TestGlobals_GlobalsInStrictVariablesMode verifies that accessing a global
// does NOT trigger a StrictVariables error — globals are "defined" variables.
func TestGlobals_GlobalsInStrictVariablesMode(t *testing.T) {
	engine := liquid.NewEngine()
	engine.StrictVariables()
	engine.SetGlobals(map[string]any{"env": "production"})

	out, err := engine.ParseAndRenderString(`{{ env }}`, map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "production", out)
}

// ---------------------------------------------------------------------------
// Extra: Bindings() method exposes current scope
// ---------------------------------------------------------------------------

// TestContext_BindingsMethod verifies that ctx.Bindings() returns the
// current scope, including merged globals.
func TestContext_BindingsMethod(t *testing.T) {
	engine := liquid.NewEngine()
	engine.SetGlobals(map[string]any{"global_x": "gx"})

	var captured map[string]any
	engine.RegisterTag("capture_bindings", func(c render.Context) (string, error) {
		captured = c.Bindings()
		return "", nil
	})

	_, err := engine.ParseAndRenderString(
		`{% capture_bindings %}`,
		map[string]any{"local_y": "ly"},
	)
	require.NoError(t, err)

	require.Equal(t, "gx", captured["global_x"], "global must appear in Bindings()")
	require.Equal(t, "ly", captured["local_y"], "local binding must appear in Bindings()")
}

// TestContext_SetPersistsWithinRender verifies that ctx.Set mutates the
// scope and that subsequent reads (via {{ variable }} or ctx.Get) see the
// updated value within the same render.
func TestContext_SetPersistsWithinRender(t *testing.T) {
	engine := liquid.NewEngine()

	engine.RegisterTag("set_x_to_999", func(c render.Context) (string, error) {
		c.Set("x", 999)
		return "", nil
	})

	out, err := engine.ParseAndRenderString(
		`{% set_x_to_999 %}{{ x }}`,
		map[string]any{"x": 0},
	)
	require.NoError(t, err)
	require.Equal(t, "999", out)
}

// TestContext_WriteValue verifies that ctx.WriteValue respects the same
// rendering rules as {{ expr }} — nil → "", arrays space-joined, etc.
func TestContext_WriteValue(t *testing.T) {
	engine := liquid.NewEngine()

	engine.RegisterTag("write_val", func(c render.Context) (string, error) {
		var buf writeCaptureBuffer
		if err := c.WriteValue(&buf, c.Get("v")); err != nil {
			return "", err
		}
		return buf.String(), nil
	})

	cases := []struct {
		binding any
		want    string
	}{
		{nil, ""},
		{[]string{"a", "b", "c"}, "abc"},
		{"hello", "hello"},
		{42, "42"},
	}

	for _, tc := range cases {
		out, err := engine.ParseAndRenderString(
			`{% write_val %}`,
			map[string]any{"v": tc.binding},
		)
		require.NoError(t, err)
		require.Equal(t, tc.want, out)
	}
}

// writeCaptureBuffer is a minimal io.Writer backed by a byte slice.
type writeCaptureBuffer struct{ data []byte }

func (w *writeCaptureBuffer) Write(p []byte) (int, error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *writeCaptureBuffer) String() string { return string(w.data) }

// ensure writeCaptureBuffer satisfies io.Writer at compile time.
var _ io.Writer = (*writeCaptureBuffer)(nil)
