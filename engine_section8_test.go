package liquid

// Section 8 — Configuration / Engine
// Ported tests from:
//   - LiquidJS: test/integration/liquid/strict.spec.ts
//   - LiquidJS: test/integration/liquid/liquid.spec.ts (globals, strictVariables render option)
//   - Ruby Liquid: test/integration/template_test.rb (strict_variables, strict_filters, exception_renderer per render)

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ── Per-render WithStrictVariables ────────────────────────────────────────────

// Source: LiquidJS test/integration/liquid/liquid.spec.ts
// "should support `strictVariables` render option"
func TestTemplate_Render_WithStrictVariables_errors_on_undefined(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ foo }}`)
	require.NoError(t, err)

	_, renderErr := tpl.Render(Bindings{}, WithStrictVariables())
	require.Error(t, renderErr)
	require.Contains(t, renderErr.Error(), "undefined variable")
}

// Source: LiquidJS test/integration/liquid/liquid.spec.ts
// "should support `strictVariables` render option" (via ParseAndRenderString)
func TestEngine_ParseAndRenderString_WithStrictVariables_errors_on_undefined(t *testing.T) {
	engine := NewEngine()
	_, err := engine.ParseAndRenderString(`{{ foo }}`, Bindings{}, WithStrictVariables())
	require.Error(t, err)
	require.Contains(t, err.Error(), "undefined variable")
}

// Source: LiquidJS test/integration/liquid/strict.spec.ts
// "should not throw when strictVariables false (default)"
func TestEngine_ParseAndRenderString_Default_NoErrorOnUndefined(t *testing.T) {
	engine := NewEngine()
	out, err := engine.ParseAndRenderString(`before{{ notdefined }}after`, Bindings{})
	require.NoError(t, err)
	require.Equal(t, "beforeafter", out)
}

// Engine-level strict is overridden per-render (same behavior): both error
func TestTemplate_Render_WithStrictVariables_also_errors_on_engine_strict(t *testing.T) {
	engine := NewEngine()
	engine.StrictVariables()
	tpl, parseErr := engine.ParseString(`{{ undefined_var }}`)
	require.NoError(t, parseErr)

	_, renderErr := tpl.Render(Bindings{})
	require.Error(t, renderErr, "engine-level strict should also error")
}

// Source: Ruby template_test.rb — test_undefined_variables
// Defined variables still render correctly with WithStrictVariables
func TestTemplate_Render_WithStrictVariables_defined_vars_work(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ x }}`)
	require.NoError(t, err)

	out, renderErr := tpl.Render(Bindings{"x": 42}, WithStrictVariables())
	require.NoError(t, renderErr)
	require.Equal(t, "42", string(out))
}

// FRender also works with WithStrictVariables
func TestTemplate_FRender_WithStrictVariables(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ undefined_var }}`)
	require.NoError(t, err)

	var buf bytes.Buffer
	renderErr := tpl.FRender(&buf, Bindings{}, WithStrictVariables())
	require.Error(t, renderErr)
	require.Contains(t, renderErr.Error(), "undefined variable")
}

// RenderString also works with WithStrictVariables
func TestTemplate_RenderString_WithStrictVariables(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ undefined_var }}`)
	require.NoError(t, err)

	_, renderErr := tpl.RenderString(Bindings{}, WithStrictVariables())
	require.Error(t, renderErr)
	require.Contains(t, renderErr.Error(), "undefined variable")
}

// Per-render option does not persist across calls on the same template
func TestTemplate_Render_WithStrictVariables_does_not_persist(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ undefined_var }}`)
	require.NoError(t, err)

	// First render with option → error
	_, renderErr := tpl.Render(Bindings{}, WithStrictVariables())
	require.Error(t, renderErr)

	// Second render without option → no error (engine level has no strict)
	out, renderErr2 := tpl.Render(Bindings{})
	require.NoError(t, renderErr2)
	require.Equal(t, "", string(out))
}

// ParseAndFRender also supports WithStrictVariables
func TestEngine_ParseAndFRender_WithStrictVariables(t *testing.T) {
	engine := NewEngine()
	var buf bytes.Buffer
	err := engine.ParseAndFRender(&buf, []byte(`{{ undefined_var }}`), Bindings{}, WithStrictVariables())
	require.Error(t, err)
	require.Contains(t, err.Error(), "undefined variable")
}

// The error message includes the variable name (from Ruby test_undefined_variables)
func TestTemplate_Render_WithStrictVariables_error_contains_name(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ my_special_var }}`)
	require.NoError(t, err)

	_, renderErr := tpl.Render(Bindings{}, WithStrictVariables())
	require.Error(t, renderErr)
	require.Contains(t, renderErr.Error(), "my_special_var")
}

// ── Per-render WithLaxFilters ─────────────────────────────────────────────────

// Source: LiquidJS register-filters.spec.ts (adapted), Ruby template_test.rb test_undefined_filters
// Default: undefined filter causes an error
func TestTemplate_Render_Default_StrictFilters(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ "hello" | nosuchfilter }}`)
	require.NoError(t, err)

	_, renderErr := tpl.Render(Bindings{})
	require.Error(t, renderErr)
}

// WithLaxFilters: undefined filter passes through the value
func TestTemplate_Render_WithLaxFilters_passes_through(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ "hello" | nosuchfilter }}`)
	require.NoError(t, err)

	out, renderErr := tpl.Render(Bindings{}, WithLaxFilters())
	require.NoError(t, renderErr)
	require.Equal(t, "hello", string(out))
}

// WithLaxFilters via ParseAndRenderString
func TestEngine_ParseAndRenderString_WithLaxFilters_passes_through(t *testing.T) {
	engine := NewEngine()
	out, err := engine.ParseAndRenderString(`{{ "hello" | nosuchfilter }}`, Bindings{}, WithLaxFilters())
	require.NoError(t, err)
	require.Equal(t, "hello", out)
}

// WithLaxFilters: defined filters still work
func TestTemplate_Render_WithLaxFilters_defined_filters_work(t *testing.T) {
	engine := NewEngine()
	out, err := engine.ParseAndRenderString(`{{ "hello" | upcase }}`, Bindings{}, WithLaxFilters())
	require.NoError(t, err)
	require.Equal(t, "HELLO", out)
}

// Per-render WithLaxFilters does not persist across calls on the same template
func TestTemplate_Render_WithLaxFilters_does_not_persist(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ "hello" | nosuchfilter }}`)
	require.NoError(t, err)

	// First call with lax → no error
	out, renderErr := tpl.Render(Bindings{}, WithLaxFilters())
	require.NoError(t, renderErr)
	require.Equal(t, "hello", string(out))

	// Second call without lax → error (default is strict)
	_, renderErr2 := tpl.Render(Bindings{})
	require.Error(t, renderErr2)
}

// ── Engine-level globals (SetGlobals / GetGlobals) ───────────────────────────

// Source: LiquidJS liquid.spec.ts "should support `globals` render option" (engine-level equivalent)
// SetGlobals makes variables accessible in every rendering context
func TestEngine_SetGlobals_accessible_in_render(t *testing.T) {
	engine := NewEngine()
	engine.SetGlobals(map[string]any{"site_name": "Acme"})

	out, err := engine.ParseAndRenderString(`{{ site_name }}`, Bindings{})
	require.NoError(t, err)
	require.Equal(t, "Acme", out)
}

// Bindings take priority over globals when keys conflict
func TestEngine_SetGlobals_bindings_override_globals(t *testing.T) {
	engine := NewEngine()
	engine.SetGlobals(map[string]any{"x": "global"})

	out, err := engine.ParseAndRenderString(`{{ x }}`, Bindings{"x": "local"})
	require.NoError(t, err)
	require.Equal(t, "local", out)
}

// GetGlobals returns the globals that were set
func TestEngine_GetGlobals_returns_set_value(t *testing.T) {
	engine := NewEngine()
	globals := map[string]any{"foo": 42}
	engine.SetGlobals(globals)
	require.Equal(t, globals, engine.GetGlobals())
}

// Globals persist across multiple renders
func TestEngine_SetGlobals_persist_across_renders(t *testing.T) {
	engine := NewEngine()
	engine.SetGlobals(map[string]any{"version": "1.0"})

	for i := range 3 {
		out, err := engine.ParseAndRenderString(`{{ version }}`, Bindings{})
		require.NoError(t, err, "render %d", i)
		require.Equal(t, "1.0", out, "render %d", i)
	}
}

// Multiple globals are all accessible
func TestEngine_SetGlobals_multiple_variables(t *testing.T) {
	engine := NewEngine()
	engine.SetGlobals(map[string]any{
		"author": "Alice",
		"year":   2024,
	})

	out, err := engine.ParseAndRenderString(`{{ author }} {{ year }}`, Bindings{})
	require.NoError(t, err)
	require.Equal(t, "Alice 2024", out)
}

// Globals are absent if SetGlobals was never called
func TestEngine_Globals_empty_by_default(t *testing.T) {
	engine := NewEngine()
	out, err := engine.ParseAndRenderString(`{{ site_name }}`, Bindings{})
	require.NoError(t, err)
	require.Equal(t, "", out) // undefined → empty string
}

// ── Combining per-render options ──────────────────────────────────────────────

// Multiple per-render options can be combined
func TestTemplate_Render_CombinedOptions(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ "hello" | nosuchfilter }}`)
	require.NoError(t, err)

	// WithStrictVariables has no effect here (filter is the issue),
	// but WithLaxFilters makes it pass through
	out, renderErr := tpl.Render(Bindings{}, WithStrictVariables(), WithLaxFilters())
	require.NoError(t, renderErr)
	require.Equal(t, "hello", string(out))
}

// ── Per-render WithGlobals ─────────────────────────────────────────────────────

// Source: LiquidJS liquid.spec.ts "should support `globals` render option"
func TestEngine_ParseAndRenderString_WithGlobals(t *testing.T) {
	engine := NewEngine()
	out, err := engine.ParseAndRenderString(`{{ foo }}`, Bindings{}, WithGlobals(map[string]any{"foo": "FOO"}))
	require.NoError(t, err)
	require.Equal(t, "FOO", out)
}

// Per-render globals do not persist to the next call
func TestEngine_WithGlobals_does_not_persist(t *testing.T) {
	engine := NewEngine()

	// With per-render global
	out, err := engine.ParseAndRenderString(`{{ foo }}`, Bindings{}, WithGlobals(map[string]any{"foo": "FOO"}))
	require.NoError(t, err)
	require.Equal(t, "FOO", out)

	// Without per-render global → no engine-level global either
	out, err = engine.ParseAndRenderString(`{{ foo }}`, Bindings{})
	require.NoError(t, err)
	require.Equal(t, "", out)
}

// Per-render globals merge with engine-level globals
func TestEngine_WithGlobals_merges_with_engine_globals(t *testing.T) {
	engine := NewEngine()
	engine.SetGlobals(map[string]any{"site": "MySite"})

	out, err := engine.ParseAndRenderString(`{{ site }} {{ version }}`, Bindings{},
		WithGlobals(map[string]any{"version": "2.0"}))
	require.NoError(t, err)
	require.Equal(t, "MySite 2.0", out)
}

// Scope bindings still override per-render globals
func TestEngine_WithGlobals_overridden_by_bindings(t *testing.T) {
	engine := NewEngine()
	out, err := engine.ParseAndRenderString(`{{ foo }}`, Bindings{"foo": "local"},
		WithGlobals(map[string]any{"foo": "global"}))
	require.NoError(t, err)
	require.Equal(t, "local", out)
}

// Per-render globals override engine-level globals
func TestEngine_WithGlobals_overrides_engine_globals(t *testing.T) {
	engine := NewEngine()
	engine.SetGlobals(map[string]any{"foo": "engine"})

	out, err := engine.ParseAndRenderString(`{{ foo }}`, Bindings{},
		WithGlobals(map[string]any{"foo": "per-render"}))
	require.NoError(t, err)
	require.Equal(t, "per-render", out)
}

// Passing an empty WithGlobals is a no-op
func TestEngine_WithGlobals_empty_is_noop(t *testing.T) {
	engine := NewEngine()
	engine.SetGlobals(map[string]any{"foo": "engine"})

	out, err := engine.ParseAndRenderString(`{{ foo }}`, Bindings{}, WithGlobals(map[string]any{}))
	require.NoError(t, err)
	require.Equal(t, "engine", out)
}

// ── WithErrorHandler (exception_renderer) ─────────────────────────────────────

// Source: Ruby template_test.rb — test_exception_renderer_that_returns_string
// WithErrorHandler replaces failing node output with handler's return value
func TestWithErrorHandler_continues_rendering(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`before{{ 1 | divided_by: 0 }}after`)
	require.NoError(t, err)

	var caught error
	out, renderErr := tpl.RenderString(Bindings{}, WithErrorHandler(func(e error) string {
		caught = e
		return "<!-- error -->"
	}))
	require.NoError(t, renderErr)
	require.Equal(t, "before<!-- error -->after", out)
	require.Error(t, caught)
}

// Source: Ruby template_test.rb — test_exception_renderer_that_returns_string
// Handler receives the actual error
func TestWithErrorHandler_receives_the_error(t *testing.T) {
	engine := NewEngine()
	var caught error
	_, err := engine.ParseAndRenderString(`{{ 1 | divided_by: 0 }}`, Bindings{},
		WithErrorHandler(func(e error) string {
			caught = e
			return ""
		}),
	)
	require.NoError(t, err)
	require.Error(t, caught)
}

// Multiple failing nodes: handler called for each
func TestWithErrorHandler_called_for_each_failure(t *testing.T) {
	engine := NewEngine()
	count := 0
	out, err := engine.ParseAndRenderString(
		`{{ 1 | divided_by: 0 }}{{ 2 | divided_by: 0 }}{{ 3 | divided_by: 0 }}`,
		Bindings{},
		WithErrorHandler(func(e error) string {
			count++
			return "X"
		}),
	)
	require.NoError(t, err)
	require.Equal(t, "XXX", out)
	require.Equal(t, 3, count)
}

// Without handler, first error stops rendering
func TestWithErrorHandler_absent_stops_on_first_error(t *testing.T) {
	engine := NewEngine()
	_, err := engine.ParseAndRenderString(`{{ 1 | divided_by: 0 }}`, Bindings{})
	require.Error(t, err)
}

// Engine-level SetExceptionHandler
func TestEngine_SetExceptionHandler(t *testing.T) {
	engine := NewEngine()
	var caught error
	engine.SetExceptionHandler(func(e error) string {
		caught = e
		return "ERR"
	})

	out, err := engine.ParseAndRenderString(`a{{ 1 | divided_by: 0 }}b`, Bindings{})
	require.NoError(t, err)
	require.Equal(t, "aERRb", out)
	require.Error(t, caught)
}

// Per-render handler overrides engine-level handler
func TestWithErrorHandler_overrides_engine_handler(t *testing.T) {
	engine := NewEngine()
	engine.SetExceptionHandler(func(e error) string { return "engine" })

	out, err := engine.ParseAndRenderString(`{{ 1 | divided_by: 0 }}`, Bindings{},
		WithErrorHandler(func(e error) string { return "per-render" }),
	)
	require.NoError(t, err)
	require.Equal(t, "per-render", out)
}

// WithErrorHandler as an error collector pattern (like template.errors in Ruby)
func TestWithErrorHandler_collect_errors(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ a }}{{ 1 | divided_by: 0 }}{{ b }}`)
	require.NoError(t, err)

	var errs []error
	out, renderErr := tpl.RenderString(Bindings{"a": "hello", "b": "world"},
		WithErrorHandler(func(e error) string {
			errs = append(errs, e)
			return ""
		}),
	)
	require.NoError(t, renderErr)
	require.Equal(t, "helloworld", out)
	require.Len(t, errs, 1)
}

// ── WithContext (time-based limits) ───────────────────────────────────────────

// Source: LiquidJS dos.spec.ts concept — context cancellation stops rendering
// WithContext stops rendering when cancelled
func TestWithContext_cancelled_stops_rendering(t *testing.T) {
	engine := NewEngine()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	tpl, err := engine.ParseString(`{% for i in (1..100) %}{{ i }}{% endfor %}`)
	require.NoError(t, err)

	_, renderErr := tpl.RenderString(Bindings{}, WithContext(ctx))
	require.Error(t, renderErr)
}

// WithContext passes through when not cancelled
func TestWithContext_not_cancelled_renders_normally(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	out, err := engine.ParseAndRenderString(`{{ x }}`, Bindings{"x": "hello"}, WithContext(ctx))
	require.NoError(t, err)
	require.Equal(t, "hello", out)
}

// WithContext: deadline exceeded returns error
func TestWithContext_deadline_exceeded(t *testing.T) {
	engine := NewEngine()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // let it expire

	tpl, err := engine.ParseString(`{% for i in (1..10000) %}{{ i }}{% endfor %}`)
	require.NoError(t, err)

	_, renderErr := tpl.RenderString(Bindings{}, WithContext(ctx))
	require.Error(t, renderErr)
}

// ── WithSizeLimit ─────────────────────────────────────────────────────────────

// Source: Ruby template_test.rb — test_resource_limits_render_length
// WithSizeLimit aborts when output exceeds the limit
func TestWithSizeLimit_exceeded_returns_error(t *testing.T) {
	engine := NewEngine()
	// Template that produces more than 5 bytes
	out, err := engine.ParseAndRenderString(`0123456789`, Bindings{}, WithSizeLimit(5))
	require.Error(t, err)
	require.Contains(t, err.Error(), "size limit")
	_ = out
}

// WithSizeLimit passes when output is within the limit
func TestWithSizeLimit_within_limit_succeeds(t *testing.T) {
	engine := NewEngine()
	out, err := engine.ParseAndRenderString(`0123456789`, Bindings{}, WithSizeLimit(10))
	require.NoError(t, err)
	require.Equal(t, "0123456789", out)
}

// Source: Ruby test_resource_limits_render_length — limit is in bytes
func TestWithSizeLimit_counts_bytes(t *testing.T) {
	engine := NewEngine()
	// "すごい" is 9 bytes in UTF-8
	out, err := engine.ParseAndRenderString(`すごい`, Bindings{}, WithSizeLimit(8))
	require.Error(t, err)
	_ = out

	out2, err2 := engine.ParseAndRenderString(`すごい`, Bindings{}, WithSizeLimit(9))
	require.NoError(t, err2)
	require.Equal(t, "すごい", out2)
}

// Zero size limit is treated as no limit
func TestWithSizeLimit_zero_means_no_limit(t *testing.T) {
	engine := NewEngine()
	out, err := engine.ParseAndRenderString(`hello world`, Bindings{}, WithSizeLimit(0))
	require.NoError(t, err)
	require.Equal(t, "hello world", out)
}

// ── Engine.LaxTags ────────────────────────────────────────────────────────────

// Source: Ruby error_mode: :lax concept
// LaxTags makes unknown tags compile as no-ops instead of errors
func TestEngine_LaxTags_unknown_tag_is_noop(t *testing.T) {
	engine := NewEngine()
	engine.LaxTags()

	out, err := engine.ParseAndRenderString(`before{% unknown_tag args %}after`, Bindings{})
	require.NoError(t, err)
	require.Equal(t, "beforeafter", out)
}

// Without LaxTags, unknown tags are parse errors
func TestEngine_LaxTags_default_errors_on_unknown(t *testing.T) {
	engine := NewEngine()
	_, err := engine.ParseString(`{% unknown_tag %}`)
	require.Error(t, err)
}

// Known tags still work with LaxTags enabled
func TestEngine_LaxTags_known_tags_still_work(t *testing.T) {
	engine := NewEngine()
	engine.LaxTags()

	out, err := engine.ParseAndRenderString(`{% if true %}yes{% endif %}`, Bindings{})
	require.NoError(t, err)
	require.Equal(t, "yes", out)
}

// Multiple unknown tags are all silently ignored
func TestEngine_LaxTags_multiple_unknown_tags(t *testing.T) {
	engine := NewEngine()
	engine.LaxTags()

	out, err := engine.ParseAndRenderString(
		`{% tag_a %}{{ x }}{% tag_b arg %}`,
		Bindings{"x": "hello"},
	)
	require.NoError(t, err)
	require.Equal(t, "hello", out)
}

// ── Template cache (EnableCache / ClearCache) ─────────────────────────────────

// ParseString returns the same object when cache is enabled
func TestEngine_EnableCache_caches_parsed_template(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache()

	tpl1, err1 := engine.ParseString(`{{ x }}`)
	require.NoError(t, err1)

	tpl2, err2 := engine.ParseString(`{{ x }}`)
	require.NoError(t, err2)

	require.Same(t, tpl1, tpl2, "cached template should be the same pointer")
}

// Different source strings produce different cached templates
func TestEngine_EnableCache_different_sources_are_different(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache()

	tpl1, _ := engine.ParseString(`{{ x }}`)
	tpl2, _ := engine.ParseString(`{{ y }}`)
	require.NotSame(t, tpl1, tpl2)
}

// Cache does not interfere with rendering correctness
func TestEngine_EnableCache_renders_correctly(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache()

	for i := range 3 {
		out, err := engine.ParseAndRenderString(`{{ x }}`, Bindings{"x": i})
		require.NoError(t, err)
		require.Equal(t, string(rune('0'+i)), out)
	}
}

// ClearCache evicts all cached templates
func TestEngine_ClearCache_evicts_entries(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache()

	tpl1, _ := engine.ParseString(`{{ x }}`)
	engine.ClearCache()
	tpl2, _ := engine.ParseString(`{{ x }}`)

	require.NotSame(t, tpl1, tpl2, "after ClearCache a new template should be created")
}

// Without EnableCache, ParseString always parses fresh
func TestEngine_NoCache_always_parses_fresh(t *testing.T) {
	engine := NewEngine()

	tpl1, _ := engine.ParseString(`{{ x }}`)
	tpl2, _ := engine.ParseString(`{{ x }}`)
	require.NotSame(t, tpl1, tpl2)
}

// Cache is concurrency-safe
func TestEngine_EnableCache_concurrencySafe(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache()

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := engine.ParseAndRenderString(`{{ x }}`, Bindings{"x": "ok"})
			require.NoError(t, err)
			require.Equal(t, "ok", out)
		}()
	}
	wg.Wait()
}
