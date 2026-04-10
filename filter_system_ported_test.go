package liquid_test

// Ported filter-system tests from:
//   - Ruby Liquid: test/integration/filter_test.rb
//   - Ruby Liquid: test/integration/filter_kwarg_test.rb
//   - LiquidJS:    test/integration/liquid/register-filters.spec.ts
//   - LiquidJS:    src/template/filter.spec.ts
//
// Covers checklist section 3: Sistema de Filtros
//   3.1  Filtros posicionais                            — ✅
//   3.2  Keyword args em filtros (filter: arg, key: v)  — ✅
//   3.3  global_filter                                  — ✅ (see engine_test.go)

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/osteele/liquid"
	"github.com/osteele/liquid/expressions"
	"github.com/stretchr/testify/require"
)

// helper: create a fresh engine with a single registered filter.
func newEngineWith(name string, fn any) *liquid.Engine {
	e := liquid.NewEngine()
	e.RegisterFilter(name, fn)
	return e
}

// renderStr renders src with the given engine+bindings, failing the test on error.
func renderStr(t *testing.T, eng *liquid.Engine, src string, bindings map[string]any) string {
	t.Helper()
	out, err := eng.ParseAndRenderString(src, bindings)
	require.NoError(t, err, "template: %s", src)
	return out
}

// ── 3.1 Positional filter args ────────────────────────────────────────────────

// TestPortedFilterSystem_PositionalArgs verifies that custom filters registered via
// RegisterFilter receive positional arguments correctly.
// Ruby: filter_test.rb::test_local_filter, test_underscore_in_filter_name
func TestPortedFilterSystem_PositionalArgs(t *testing.T) {
	// single-arg filter (value only)
	eng := newEngineWith("money", func(input int) string {
		return fmt.Sprintf(" %d$ ", input)
	})
	got := renderStr(t, eng, "{{var | money}}", map[string]any{"var": 1000})
	require.Equal(t, " 1000$ ", got)

	// filter with underscore in name (Ruby: test_underscore_in_filter_name)
	eng2 := newEngineWith("money_with_underscore", func(input int) string {
		return fmt.Sprintf(" %d$ ", input)
	})
	got2 := renderStr(t, eng2, "{{var | money_with_underscore}}", map[string]any{"var": 1000})
	require.Equal(t, " 1000$ ", got2)

	// two-arg filter: value + positional argument
	eng3 := newEngineWith("add", func(a, b int) int { return a + b })
	got3 := renderStr(t, eng3, "{{a | add: b}}", map[string]any{"a": 3, "b": 2})
	require.Equal(t, "5", got3)
}

// TestPortedFilterSystem_FilterOverride verifies that registering a second filter
// with the same name replaces the first (last-write-wins).
// Ruby: filter_test.rb::test_second_filter_overwrites_first
func TestPortedFilterSystem_FilterOverride(t *testing.T) {
	eng := liquid.NewEngine()
	eng.RegisterFilter("money", func(input int) string {
		return fmt.Sprintf(" %d$ ", input)
	})
	eng.RegisterFilter("money", func(input int) string {
		return fmt.Sprintf(" %d$ CAD ", input)
	})
	got := renderStr(t, eng, "{{var | money}}", map[string]any{"var": 1000})
	require.Equal(t, " 1000$ CAD ", got)
}

// ── 3.2 Keyword args in filters ───────────────────────────────────────────────

// TestPortedFilterSystem_CustomFilterWithKwargs verifies that a custom filter can
// receive keyword arguments as expressions.NamedArg values.
// Ruby: filter_test.rb::test_filter_with_keyword_arguments (SubstituteFilter pattern)
func TestPortedFilterSystem_CustomFilterWithKwargs(t *testing.T) {
	// substitute performs %{key} replacement using kwargs.
	eng := newEngineWith("substitute", func(input string, args ...any) string {
		for _, a := range args {
			if na, ok := a.(expressions.NamedArg); ok {
				input = strings.ReplaceAll(input, "%{"+na.Name+"}", fmt.Sprint(na.Value))
			}
		}
		return input
	})

	got := renderStr(t, eng,
		`{{ input | substitute: first_name: surname, last_name: 'doe' }}`,
		map[string]any{
			"surname": "john",
			"input":   "hello %{first_name}, %{last_name}",
		})
	require.Equal(t, "hello john, doe", got)
}

// TestPortedFilterSystem_HyphenatedKwargs verifies that keyword argument keys that
// contain hyphens are parsed correctly.
// Ruby: filter_kwarg_test.rb::test_can_parse_data_kwargs
func TestPortedFilterSystem_HyphenatedKwargs(t *testing.T) {
	// html_tag collects kwargs and formats them as HTML attributes.
	eng := newEngineWith("html_tag", func(tag string, args ...any) string {
		var parts []string
		for _, a := range args {
			if na, ok := a.(expressions.NamedArg); ok {
				parts = append(parts, fmt.Sprintf("%s='%v'", na.Name, na.Value))
			}
		}
		return strings.Join(parts, " ")
	})

	got := renderStr(t, eng,
		`{{ 'img' | html_tag: data-src: 'src', data-widths: '100, 200' }}`,
		nil)
	require.Equal(t, "data-src='src' data-widths='100, 200'", got)
}

// TestPortedFilterSystem_KVArgsPassedAsNamedArgs verifies that key-value arguments
// are passed to the filter as expressions.NamedArg values (one NamedArg per pair).
// JS: register-filters.spec.ts::key-value arguments
func TestPortedFilterSystem_KVArgsPassedAsNamedArgs(t *testing.T) {
	// obj_test serialises all received arguments as a JSON array, representing
	// named args as ["key", value] pairs — matching the LiquidJS assertion format.
	eng := newEngineWith("obj_test", func(v any, args ...any) string {
		elems := make([]any, 0, 1+len(args))
		elems = append(elems, v)
		for _, a := range args {
			if na, ok := a.(expressions.NamedArg); ok {
				elems = append(elems, []any{na.Name, na.Value})
			} else {
				elems = append(elems, a)
			}
		}
		b, _ := json.Marshal(elems)
		return string(b)
	})

	got := renderStr(t, eng,
		`{{ "a" | obj_test: k1: "v1", k2: foo }}`,
		map[string]any{"foo": "bar"})
	require.Equal(t, `["a",["k1","v1"],["k2","bar"]]`, got)
}

// TestPortedFilterSystem_MixedPositionalAndKwargs verifies that positional arguments
// and keyword arguments can be mixed in a single filter call.
// JS: register-filters.spec.ts::mixed arguments
func TestPortedFilterSystem_MixedPositionalAndKwargs(t *testing.T) {
	eng := newEngineWith("obj_test", func(v any, args ...any) string {
		elems := make([]any, 0, 1+len(args))
		elems = append(elems, v)
		for _, a := range args {
			if na, ok := a.(expressions.NamedArg); ok {
				elems = append(elems, []any{na.Name, na.Value})
			} else {
				elems = append(elems, a)
			}
		}
		b, _ := json.Marshal(elems)
		return string(b)
	})

	got := renderStr(t, eng,
		`{{ "a" | obj_test: "something", k1: "v1", k2: foo }}`,
		map[string]any{"foo": "bar"})
	require.Equal(t, `["a","something",["k1","v1"],["k2","bar"]]`, got)
}

// ── 3.1 Undefined filter in lax mode ─────────────────────────────────────────

// TestPortedFilterSystem_UndefinedFilterIgnored verifies that an undefined filter
// passes through the value unchanged when LaxFilters mode is enabled.
// Ruby: filter_test.rb::test_nonexistent_filter_is_ignored
// Note: Ruby ignores undefined filters by default; in Go this requires explicit
// engine.LaxFilters() or render-time liquid.WithLaxFilters().
func TestPortedFilterSystem_UndefinedFilterIgnored(t *testing.T) {
	eng := liquid.NewEngine()
	eng.LaxFilters() // Required in Go; Ruby uses this as default behaviour
	got := renderStr(t, eng, `{{ var | xyzzy }}`, map[string]any{"var": 1000})
	require.Equal(t, "1000", got)
}

// TestPortedFilterSystem_UndefinedFilterError verifies that an undefined filter
// returns an error when NOT in lax-filters mode (the Go default).
// This is the Go default — stricter than Ruby.
func TestPortedFilterSystem_UndefinedFilterError(t *testing.T) {
	eng := liquid.NewEngine()
	_, err := eng.ParseAndRenderString(`{{ var | xyzzy }}`, map[string]any{"var": 1000})
	require.Error(t, err, "Go default should error on undefined filter")
	require.Contains(t, err.Error(), "undefined filter")
}

// ── 3.2 Wrong argument count ──────────────────────────────────────────────────

// TestPortedFilterSystem_WrongArgCount verifies that passing too many positional
// arguments to a filter returns an error whose message includes the phrase
// "wrong number of arguments".
// Ruby: filter_test.rb::test_liquid_argument_error
func TestPortedFilterSystem_WrongArgCount(t *testing.T) {
	eng := liquid.NewEngine()
	// `size` takes no extra args; passing one triggers a CallParityError.
	_, err := eng.ParseAndRenderString(`{{ '' | size: 'too many args' }}`, nil)
	require.Error(t, err, "expected an error for wrong arg count")
	require.Contains(t, err.Error(), "wrong number of arguments",
		"error message should mention wrong number of arguments")
	// The render error always starts with "Liquid error" (matching Ruby prefix).
	require.True(t, strings.HasPrefix(err.Error(), "Liquid error"),
		"error should carry the 'Liquid error' prefix, got: %s", err.Error())
}

// TestPortedFilterSystem_WrongArgCount_StrictMode confirms the same behaviour
// holds in strict-filters mode.
func TestPortedFilterSystem_WrongArgCount_StrictMode(t *testing.T) {
	eng := liquid.NewEngine()
	_, err := eng.ParseAndRenderString(`{{ '' | size: 'extra' }}`, nil,
		liquid.WithStrictVariables())
	require.Error(t, err)
	require.Contains(t, err.Error(), "wrong number of arguments")
}
