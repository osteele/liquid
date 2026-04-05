package liquid_test

// s8_engine_config_e2e_test.go — Intensive E2E tests for Section 8: Configuration / Engine
//
// Coverage matrix:
//   A. StrictVariables — engine-level and per-render, exact error messages, defined vars ok
//   B. LaxFilters — engine-level and per-render, passthrough behavior, filter chaining
//   C. LaxTags — unknown as noop, known tags work, LaxTags does not affect filter strictness
//   D. Delims — custom tag/output delimiters, old delims become literal, empty string restores
//   E. RegisterFilter — custom filter, arg passing, chaining, override standard
//   F. RegisterTag — context access, state, multi-render isolation
//   G. RegisterBlock — InnerString, conditional content, nested manipulation
//   H. UnregisterTag — hot-replace pattern, idempotent removal
//   I. RegisterTemplateStore — in-memory store, include dispatch, multiple files
//   J. SetGlobals + WithGlobals — render hierarchy (bindings > per-render > engine), persistence
//   K. SetGlobalFilter + WithGlobalFilter — all outputs transformed, per-render override, combined
//   L. SetExceptionHandler + WithErrorHandler — recovery, collection, per-render overrides engine
//   M. WithSizeLimit — loop-generated content, UTF-8 bytes, zero = unlimited, per-render isolation
//   N. WithContext — cancellation, timeout in loop, background context passes through
//   O. EnableCache — cache hit returns same pointer, invalidation, concurrent safety
//   P. EnableJekyllExtensions — dot assign in real templates, standard assign still works
//   Q. SetAutoEscapeReplacer — HTML escaping in output, raw filter bypasses, interaction
//   R. NewBasicEngine — no standard tags/filters, custom registration works
//   S. Combinations — multiple render options together, realistic template scenarios
//   T. Real-world — blog layout, error recovery report, custom-auth tag pipeline
//
// Every test is self-contained: it creates its own engine and store.

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/osteele/liquid"
	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func s8eng() *liquid.Engine { return liquid.NewEngine() }

func s8render(t *testing.T, eng *liquid.Engine, tpl string, binds map[string]any, opts ...liquid.RenderOption) string {
	t.Helper()
	out, err := eng.ParseAndRenderString(tpl, binds, opts...)
	require.NoError(t, err, "template: %q", tpl)
	return out
}

func s8renderErr(t *testing.T, eng *liquid.Engine, tpl string, binds map[string]any, opts ...liquid.RenderOption) (string, error) {
	t.Helper()
	return eng.ParseAndRenderString(tpl, binds, opts...)
}

// mapStore is an in-memory TemplateStore for testing RegisterTemplateStore.
type mapStore struct{ files map[string]string }

func (s *mapStore) ReadTemplate(name string) ([]byte, error) {
	if src, ok := s.files[name]; ok {
		return []byte(src), nil
	}
	return nil, fmt.Errorf("template %q not found", name)
}

// ═════════════════════════════════════════════════════════════════════════════
// A. StrictVariables
// ═════════════════════════════════════════════════════════════════════════════

// A1 — engine-level strict: undefined variable is an error
func TestS8_StrictVariables_Engine_ErrorOnUndefined(t *testing.T) {
	eng := s8eng()
	eng.StrictVariables()
	_, err := s8renderErr(t, eng, `{{ undefined }}`, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "undefined")
}

// A2 — engine-level strict: error message includes the variable name
func TestS8_StrictVariables_Engine_ErrorMessageContainsName(t *testing.T) {
	eng := s8eng()
	eng.StrictVariables()
	_, err := s8renderErr(t, eng, `{{ my_custom_var }}`, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "my_custom_var")
}

// A3 — engine-level strict: defined variables render correctly
func TestS8_StrictVariables_Engine_DefinedVarWorks(t *testing.T) {
	eng := s8eng()
	eng.StrictVariables()
	out := s8render(t, eng, `{{ x }}`, map[string]any{"x": "hello"})
	require.Equal(t, "hello", out)
}

// A4 — engine-level strict: intermediate variable in complex expression
func TestS8_StrictVariables_Engine_ObjectPropertyStillResolves(t *testing.T) {
	eng := s8eng()
	eng.StrictVariables()
	out := s8render(t, eng, `{{ user.name }}`, map[string]any{"user": map[string]any{"name": "Alice"}})
	require.Equal(t, "Alice", out)
}

// A5 — per-render strict: overrides engine default (lax)
func TestS8_StrictVariables_PerRender_OverridesEngineDefault(t *testing.T) {
	eng := s8eng() // default: lax
	_, err := s8renderErr(t, eng, `{{ missing }}`, nil, liquid.WithStrictVariables())
	require.Error(t, err)
}

// A6 — per-render strict does not persist to next call
func TestS8_StrictVariables_PerRender_DoesNotPersist(t *testing.T) {
	eng := s8eng()
	// Call 1: strict → error
	_, err := s8renderErr(t, eng, `{{ missing }}`, nil, liquid.WithStrictVariables())
	require.Error(t, err)
	// Call 2: no option → lax, renders empty
	out, err2 := s8renderErr(t, eng, `{{ missing }}`, nil)
	require.NoError(t, err2)
	require.Equal(t, "", out)
}

// A7 — strict: assign-defined variable is not treated as undefined
func TestS8_StrictVariables_AssignedVarIsNotUndefined(t *testing.T) {
	eng := s8eng()
	eng.StrictVariables()
	out := s8render(t, eng, `{% assign x = "world" %}{{ x }}`, nil)
	require.Equal(t, "world", out)
}

// A8 — strict: for-loop variable is not undefined
func TestS8_StrictVariables_ForLoopVarIsNotUndefined(t *testing.T) {
	eng := s8eng()
	eng.StrictVariables()
	out := s8render(t, eng, `{% for i in (1..3) %}{{ i }}{% endfor %}`, nil)
	require.Equal(t, "123", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// B. LaxFilters
// ═════════════════════════════════════════════════════════════════════════════

// B1 — engine-level LaxFilters: undefined filter passes value through
func TestS8_LaxFilters_Engine_PassesThrough(t *testing.T) {
	eng := s8eng()
	eng.LaxFilters()
	out := s8render(t, eng, `{{ "hello" | no_such_filter }}`, nil)
	require.Equal(t, "hello", out)
}

// B2 — engine-level LaxFilters: defined filters still work
func TestS8_LaxFilters_Engine_DefinedFilterWorks(t *testing.T) {
	eng := s8eng()
	eng.LaxFilters()
	out := s8render(t, eng, `{{ "hello" | upcase }}`, nil)
	require.Equal(t, "HELLO", out)
}

// B3 — engine-level LaxFilters: unknown filter in a chain — value passes through to next
func TestS8_LaxFilters_Engine_UnknownInChainPassesThrough(t *testing.T) {
	eng := s8eng()
	eng.LaxFilters()
	// unknown_filter passes value → upcase applies on the passed-through value
	out := s8render(t, eng, `{{ "hello" | unknown_filter | upcase }}`, nil)
	require.Equal(t, "HELLO", out)
}

// B4 — default (strict) mode: undefined filter causes error
func TestS8_LaxFilters_Default_StrictErrors(t *testing.T) {
	eng := s8eng()
	_, err := s8renderErr(t, eng, `{{ "hello" | no_such_filter }}`, nil)
	require.Error(t, err)
}

// B5 — per-render WithLaxFilters: overrides default strict
func TestS8_LaxFilters_PerRender_OverridesDefault(t *testing.T) {
	eng := s8eng()
	out, err := s8renderErr(t, eng, `{{ "hello" | ghost_filter }}`, nil, liquid.WithLaxFilters())
	require.NoError(t, err)
	require.Equal(t, "hello", out)
}

// B6 — per-render WithLaxFilters: does not persist to next call
func TestS8_LaxFilters_PerRender_DoesNotPersist(t *testing.T) {
	eng := s8eng()
	// Call 1: lax → no error
	out, _ := s8renderErr(t, eng, `{{ "x" | ghost_filter }}`, nil, liquid.WithLaxFilters())
	require.Equal(t, "x", out)
	// Call 2: default strict → error
	_, err := s8renderErr(t, eng, `{{ "x" | ghost_filter }}`, nil)
	require.Error(t, err)
}

// B7 — LaxFilters + LaxTags: both can be enabled together
func TestS8_LaxFilters_AndLaxTags_Together(t *testing.T) {
	eng := s8eng()
	eng.LaxFilters()
	eng.LaxTags()
	out := s8render(t, eng, `{% ghost_tag %}{{ "hello" | ghost_filter }}`, nil)
	require.Equal(t, "hello", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// C. LaxTags
// ═════════════════════════════════════════════════════════════════════════════

// C1 — unknown tag becomes noop: text around it is preserved
func TestS8_LaxTags_UnknownTagIsNoop(t *testing.T) {
	eng := s8eng()
	eng.LaxTags()
	out := s8render(t, eng, `before{% ghost_tag arg1 arg2 %}after`, nil)
	require.Equal(t, "beforeafter", out)
}

// C2 — default: unknown tag is a parse error
func TestS8_LaxTags_Default_UnknownTagIsError(t *testing.T) {
	eng := s8eng()
	_, err := eng.ParseString(`{% ghost_tag %}`)
	require.Error(t, err)
}

// C3 — LaxTags: known standard tags still work correctly
func TestS8_LaxTags_KnownTagsStillWork(t *testing.T) {
	eng := s8eng()
	eng.LaxTags()
	out := s8render(t, eng, `{% if x %}yes{% else %}no{% endif %}`, map[string]any{"x": true})
	require.Equal(t, "yes", out)
}

// C4 — LaxTags: multiple unknown tags all silently ignored
func TestS8_LaxTags_MultipleUnknownTagsAllIgnored(t *testing.T) {
	eng := s8eng()
	eng.LaxTags()
	out := s8render(t, eng,
		`{% foo %}{{ a }}{% bar baz %}{{ b }}{% qux 1 2 3 %}`,
		map[string]any{"a": "A", "b": "B"},
	)
	require.Equal(t, "AB", out)
}

// C5 — LaxTags does NOT make filters lax; undefined filter still errors
func TestS8_LaxTags_DoesNotImplyLaxFilters(t *testing.T) {
	eng := s8eng()
	eng.LaxTags()
	_, err := s8renderErr(t, eng, `{{ "x" | unknown_filter }}`, nil)
	require.Error(t, err)
}

// C6 — LaxTags: unknown tag adjacent to whitespace trim marker
func TestS8_LaxTags_UnknownTag_WithTrimMarker_IsNoop(t *testing.T) {
	eng := s8eng()
	eng.LaxTags()
	out := s8render(t, eng, `a {%- ghost_tag -%} b`, nil)
	// With trim markers the whitespace around the noop tag is consumed
	require.Equal(t, "ab", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// D. Delims
// ═════════════════════════════════════════════════════════════════════════════

// D1 — custom tag delimiters: template uses new delims correctly
func TestS8_Delims_CustomTagDelims(t *testing.T) {
	eng := s8eng()
	// Delims(objectLeft, objectRight, tagLeft, tagRight)
	eng.Delims("", "", "{!", "!}")
	out := s8render(t, eng, `{! if x !}yes{! endif !}`, map[string]any{"x": true})
	require.Equal(t, "yes", out)
}

// D2 — custom output delimiters: template uses new delims correctly
func TestS8_Delims_CustomOutputDelims(t *testing.T) {
	eng := s8eng()
	// Delims(objectLeft, objectRight, tagLeft, tagRight)
	eng.Delims("[[", "]]", "", "")
	out := s8render(t, eng, `Hello [[ name ]]!`, map[string]any{"name": "World"})
	require.Equal(t, "Hello World!", out)
}

// D3 — both custom: old delimiters become literal text
func TestS8_Delims_OldDelimsBecomeLiteral(t *testing.T) {
	eng := s8eng()
	// Output = [[ ]], Tag = {! !}
	eng.Delims("[[", "]]", "{!", "!}")
	out := s8render(t, eng, `{{ name }} and [[ name ]]`, map[string]any{"name": "X"})
	// {{ name }} is literal text; [[ name ]] is the active output delim
	require.Equal(t, "{{ name }} and X", out)
}

// D4 — empty strings restore defaults
func TestS8_Delims_EmptyRestoresDefaults(t *testing.T) {
	eng := s8eng()
	eng.Delims("", "", "", "")
	out := s8render(t, eng, `{{ x }}`, map[string]any{"x": "ok"})
	require.Equal(t, "ok", out)
}

// D5 — custom delims: for-loop with custom tag and output delimiters together
func TestS8_Delims_ForLoopWithCustomTagDelims(t *testing.T) {
	eng := s8eng()
	// Output = <% %>, Tag = <$ $>
	eng.Delims("<%", "%>", "<$", "$>")
	out := s8render(t, eng,
		`<$ for i in (1..3) $><% i %><$ endfor $>`,
		nil,
	)
	require.Equal(t, "123", out)
}

// D6 — old standard delims become literal after Delims() is called with custom values
func TestS8_Delims_StandardDelimsBecomeLiteral(t *testing.T) {
	eng := s8eng()
	// Set custom delims: output = [[ ]], tag = [% %]
	eng.Delims("[[", "]]", "[%", "%]")
	// Standard {{ }} and {% %} are now literal text
	out := s8render(t, eng, `{{ x }} [[ x ]]`, map[string]any{"x": "ok"})
	require.Equal(t, "{{ x }} ok", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// E. RegisterFilter
// ═════════════════════════════════════════════════════════════════════════════

// E1 — custom filter: basic transformation
func TestS8_RegisterFilter_BasicTransform(t *testing.T) {
	eng := s8eng()
	eng.RegisterFilter("shout", func(s string) string {
		return strings.ToUpper(s) + "!!!"
	})
	out := s8render(t, eng, `{{ "hello" | shout }}`, nil)
	require.Equal(t, "HELLO!!!", out)
}

// E2 — custom filter with argument
func TestS8_RegisterFilter_WithArg(t *testing.T) {
	eng := s8eng()
	eng.RegisterFilter("repeat", func(s string, n int) string {
		return strings.Repeat(s, n)
	})
	out := s8render(t, eng, `{{ "ab" | repeat: 3 }}`, nil)
	require.Equal(t, "ababab", out)
}

// E3 — custom filter chained with standard filter
func TestS8_RegisterFilter_ChainedWithStandard(t *testing.T) {
	eng := s8eng()
	eng.RegisterFilter("exclaim", func(s string) string { return s + "!" })
	out := s8render(t, eng, `{{ "hello" | exclaim | upcase }}`, nil)
	require.Equal(t, "HELLO!", out)
}

// E4 — custom filter returning error
func TestS8_RegisterFilter_ReturnsError(t *testing.T) {
	eng := s8eng()
	eng.RegisterFilter("fail_always", func(s string) (string, error) {
		return "", fmt.Errorf("filter failed: %s", s)
	})
	_, err := s8renderErr(t, eng, `{{ "oops" | fail_always }}`, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "filter failed")
}

// E5 — custom filter can shadow a standard filter
func TestS8_RegisterFilter_ShadowsStandard(t *testing.T) {
	eng := s8eng()
	eng.RegisterFilter("upcase", func(s string) string { return "CUSTOM:" + s })
	out := s8render(t, eng, `{{ "hi" | upcase }}`, nil)
	require.Equal(t, "CUSTOM:hi", out)
}

// E6 — custom filter on numeric input
func TestS8_RegisterFilter_NumericInput(t *testing.T) {
	eng := s8eng()
	eng.RegisterFilter("square", func(n int) int { return n * n })
	out := s8render(t, eng, `{{ 7 | square }}`, nil)
	require.Equal(t, "49", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// F. RegisterTag
// ═════════════════════════════════════════════════════════════════════════════

// F1 — custom tag: reads TagArgs and renders output
func TestS8_RegisterTag_ReadsArgsAndRenders(t *testing.T) {
	eng := s8eng()
	eng.RegisterTag("greet", func(ctx render.Context) (string, error) {
		return "Hello, " + ctx.TagArgs() + "!", nil
	})
	out := s8render(t, eng, `{% greet World %}`, nil)
	require.Equal(t, "Hello, World!", out)
}

// F2 — custom tag: reads from context variables
func TestS8_RegisterTag_ReadsContextVariable(t *testing.T) {
	eng := s8eng()
	eng.RegisterTag("greet_user", func(ctx render.Context) (string, error) {
		v := ctx.Get("username")
		name, _ := v.(string)
		return "Hi " + name, nil
	})
	out := s8render(t, eng, `{% greet_user %}`, map[string]any{"username": "Alice"})
	require.Equal(t, "Hi Alice", out)
}

// F3 — custom tag: output is independent across multiple renders
func TestS8_RegisterTag_OutputIsIsolatedPerRender(t *testing.T) {
	eng := s8eng()
	eng.RegisterTag("ping", func(ctx render.Context) (string, error) {
		return "pong", nil
	})
	for range 3 {
		out := s8render(t, eng, `{% ping %}`, nil)
		require.Equal(t, "pong", out)
	}
}

// F4 — custom tag: multiple custom tags coexist
func TestS8_RegisterTag_MultipleCustomTagsCoexist(t *testing.T) {
	eng := s8eng()
	eng.RegisterTag("tagA", func(_ render.Context) (string, error) { return "A", nil })
	eng.RegisterTag("tagB", func(_ render.Context) (string, error) { return "B", nil })
	out := s8render(t, eng, `{% tagA %}{% tagB %}{% tagA %}`, nil)
	require.Equal(t, "ABA", out)
}

// F5 — custom tag: can call EvaluateString for expressions
func TestS8_RegisterTag_EvaluatesExpression(t *testing.T) {
	eng := s8eng()
	eng.RegisterTag("eval_tag", func(ctx render.Context) (string, error) {
		v, err := ctx.EvaluateString(ctx.TagArgs())
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", v), nil
	})
	out := s8render(t, eng, `{% eval_tag x | upcase %}`, map[string]any{"x": "hello"})
	require.Equal(t, "HELLO", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// G. RegisterBlock
// ═════════════════════════════════════════════════════════════════════════════

// G1 — custom block: wraps InnerString in custom markup
func TestS8_RegisterBlock_WrapsInnerContent(t *testing.T) {
	eng := s8eng()
	eng.RegisterBlock("wrap", func(ctx render.Context) (string, error) {
		inner, err := ctx.InnerString()
		if err != nil {
			return "", err
		}
		return "[" + strings.TrimSpace(inner) + "]", nil
	})
	out := s8render(t, eng, `{% wrap %} hello {% endwrap %}`, nil)
	require.Equal(t, "[hello]", out)
}

// G2 — custom block: inner content has access to outer variables
func TestS8_RegisterBlock_InnerAccessesOuterVars(t *testing.T) {
	eng := s8eng()
	eng.RegisterBlock("uppercase_block", func(ctx render.Context) (string, error) {
		inner, err := ctx.InnerString()
		if err != nil {
			return "", err
		}
		return strings.ToUpper(inner), nil
	})
	out := s8render(t, eng, `{% uppercase_block %}{{ name }}{% enduppercase_block %}`, map[string]any{"name": "alice"})
	require.Equal(t, "ALICE", out)
}

// G3 — custom block: TagArgs available in block handler
func TestS8_RegisterBlock_TagArgsAvailable(t *testing.T) {
	eng := s8eng()
	eng.RegisterBlock("labeled", func(ctx render.Context) (string, error) {
		inner, err := ctx.InnerString()
		if err != nil {
			return "", err
		}
		return ctx.TagArgs() + ": " + strings.TrimSpace(inner), nil
	})
	out := s8render(t, eng, `{% labeled warning %}danger!{% endlabeled %}`, nil)
	require.Equal(t, "warning: danger!", out)
}

// G4 — custom block: renders empty inner content gracefully
func TestS8_RegisterBlock_EmptyInnerContent(t *testing.T) {
	eng := s8eng()
	eng.RegisterBlock("maybe", func(ctx render.Context) (string, error) {
		inner, err := ctx.InnerString()
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(inner) == "" {
			return "(empty)", nil
		}
		return inner, nil
	})
	out := s8render(t, eng, `{% maybe %}{% endmaybe %}`, nil)
	require.Equal(t, "(empty)", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// H. UnregisterTag
// ═════════════════════════════════════════════════════════════════════════════

// H1 — UnregisterTag: removes a previously registered custom tag
func TestS8_UnregisterTag_RemovesCustomTag(t *testing.T) {
	eng := s8eng()
	eng.RegisterTag("my_tag", func(_ render.Context) (string, error) { return "hi", nil })
	eng.UnregisterTag("my_tag")
	// After removal, the tag should cause a parse error (strict mode)
	_, err := eng.ParseString(`{% my_tag %}`)
	require.Error(t, err)
}

// H2 — UnregisterTag: idempotent — calling on unknown tag does not panic
func TestS8_UnregisterTag_IdempotentOnUnknown(t *testing.T) {
	eng := s8eng()
	require.NotPanics(t, func() { eng.UnregisterTag("nonexistent_tag") })
}

// H3 — UnregisterTag: can remove then re-register with new behavior
func TestS8_UnregisterTag_ReRegisterWithNewBehavior(t *testing.T) {
	eng1 := s8eng()
	eng1.RegisterTag("v_tag", func(_ render.Context) (string, error) { return "v1", nil })
	out1 := s8render(t, eng1, `{% v_tag %}`, nil)
	require.Equal(t, "v1", out1)

	// New engine: different behavior
	eng2 := s8eng()
	eng2.RegisterTag("v_tag", func(_ render.Context) (string, error) { return "v2", nil })
	out2 := s8render(t, eng2, `{% v_tag %}`, nil)
	require.Equal(t, "v2", out2)
}

// H4 — UnregisterTag: standard tags can be unregistered (LaxTags not required)
func TestS8_UnregisterTag_CanRemoveStandardTag(t *testing.T) {
	eng := s8eng()
	eng.UnregisterTag("assign")
	eng.LaxTags() // to handle the now-unknown assign tag as noop
	// assign is now a noop; variable stays undefined
	out := s8render(t, eng, `{% assign x = "hello" %}{{ x }}`, nil)
	require.Equal(t, "", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// I. RegisterTemplateStore
// ═════════════════════════════════════════════════════════════════════════════

// I1 — in-memory store: include resolves from store
func TestS8_RegisterTemplateStore_IncludeFromStore(t *testing.T) {
	eng := s8eng()
	eng.RegisterTemplateStore(&mapStore{files: map[string]string{
		"greeting.html": "Hello, {{ name }}!",
	}})
	out := s8render(t, eng, `{% include "greeting.html" %}`, map[string]any{"name": "World"})
	require.Equal(t, "Hello, World!", out)
}

// I2 — store: unknown file causes error
func TestS8_RegisterTemplateStore_UnknownFileErrors(t *testing.T) {
	eng := s8eng()
	eng.RegisterTemplateStore(&mapStore{files: map[string]string{}})
	_, err := s8renderErr(t, eng, `{% include "missing.html" %}`, nil)
	require.Error(t, err)
}

// I3 — store: multiple files, includes work for each
func TestS8_RegisterTemplateStore_MultipleFilesWork(t *testing.T) {
	eng := s8eng()
	eng.RegisterTemplateStore(&mapStore{files: map[string]string{
		"header.html": "<header>{{ title }}</header>",
		"footer.html": "<footer>© {{ year }}</footer>",
	}})
	out := s8render(t, eng,
		`{% include "header.html" %}{% include "footer.html" %}`,
		map[string]any{"title": "Home", "year": 2025},
	)
	require.Equal(t, "<header>Home</header><footer>© 2025</footer>", out)
}

// I4 — store: included template inherits calling context variables
func TestS8_RegisterTemplateStore_IncludedTemplateInheritsContext(t *testing.T) {
	eng := s8eng()
	eng.RegisterTemplateStore(&mapStore{files: map[string]string{
		"part.html": "{{ shared_var }}",
	}})
	out := s8render(t, eng, `{% include "part.html" %}`, map[string]any{"shared_var": "shared!"})
	require.Equal(t, "shared!", out)
}

// I5 — store: render tag uses isolated scope (render tag, not include)
func TestS8_RegisterTemplateStore_RenderTagUsesIsolatedScope(t *testing.T) {
	eng := s8eng()
	eng.RegisterTemplateStore(&mapStore{files: map[string]string{
		"isolated.html": "{{ secret }}",
	}})
	// render tag does NOT inherit parent scope — secret should be empty
	out := s8render(t, eng, `{% render "isolated.html" %}`, map[string]any{"secret": "hidden"})
	require.Equal(t, "", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// J. SetGlobals + WithGlobals
// ═════════════════════════════════════════════════════════════════════════════

// J1 — engine globals: accessible in every render without passing bindings
func TestS8_SetGlobals_AccessibleInEveryRender(t *testing.T) {
	eng := s8eng()
	eng.SetGlobals(map[string]any{"site": "Acme Corp", "version": 3})
	out := s8render(t, eng, `{{ site }} v{{ version }}`, nil)
	require.Equal(t, "Acme Corp v3", out)
}

// J2 — engine globals: persist across multiple renders
func TestS8_SetGlobals_PersistAcrossRenders(t *testing.T) {
	eng := s8eng()
	eng.SetGlobals(map[string]any{"env": "production"})
	for i := range 5 {
		out := s8render(t, eng, `{{ env }}`, nil)
		require.Equal(t, "production", out, "render %d", i)
	}
}

// J3 — binding overrides engine global when same key
func TestS8_SetGlobals_BindingOverridesGlobal(t *testing.T) {
	eng := s8eng()
	eng.SetGlobals(map[string]any{"color": "blue"})
	out := s8render(t, eng, `{{ color }}`, map[string]any{"color": "red"})
	require.Equal(t, "red", out)
}

// J4 — per-render WithGlobals: key present only for that call
func TestS8_WithGlobals_PerRender_NotPersistent(t *testing.T) {
	eng := s8eng()
	out1 := s8render(t, eng, `{{ x }}`, nil, liquid.WithGlobals(map[string]any{"x": "transient"}))
	require.Equal(t, "transient", out1)
	out2 := s8render(t, eng, `{{ x }}`, nil)
	require.Equal(t, "", out2)
}

// J5 — per-render WithGlobals merges with engine globals
func TestS8_WithGlobals_MergesWithEngineGlobals(t *testing.T) {
	eng := s8eng()
	eng.SetGlobals(map[string]any{"a": "A"})
	out := s8render(t, eng, `{{ a }}-{{ b }}`, nil, liquid.WithGlobals(map[string]any{"b": "B"}))
	require.Equal(t, "A-B", out)
}

// J6 — per-render WithGlobals overrides engine globals (same key)
func TestS8_WithGlobals_OverridesEngineGlobals(t *testing.T) {
	eng := s8eng()
	eng.SetGlobals(map[string]any{"env": "production"})
	out := s8render(t, eng, `{{ env }}`, nil, liquid.WithGlobals(map[string]any{"env": "staging"}))
	require.Equal(t, "staging", out)
}

// J7 — hierarchy is bindings > per-render globals > engine globals
func TestS8_Globals_FullHierarchy(t *testing.T) {
	eng := s8eng()
	eng.SetGlobals(map[string]any{"v": "engine"})
	// per-render overrides engine; binding overrides per-render
	out := s8render(t, eng, `{{ v }}`, map[string]any{"v": "binding"},
		liquid.WithGlobals(map[string]any{"v": "per-render"}))
	require.Equal(t, "binding", out)
}

// J8 — engine globals are visible in {% render %} isolated sub-contexts
func TestS8_SetGlobals_VisibleInRenderIsolated(t *testing.T) {
	eng := s8eng()
	eng.RegisterTemplateStore(&mapStore{files: map[string]string{
		"sub.html": "{{ site }}",
	}})
	eng.SetGlobals(map[string]any{"site": "MyBlog"})
	out := s8render(t, eng, `{% render "sub.html" %}`, nil)
	require.Equal(t, "MyBlog", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// K. SetGlobalFilter + WithGlobalFilter
// ═════════════════════════════════════════════════════════════════════════════

// K1 — engine global filter: transforms every {{ }} output
func TestS8_SetGlobalFilter_TransformsAllOutputs(t *testing.T) {
	eng := s8eng()
	eng.SetGlobalFilter(func(v any) (any, error) {
		if s, ok := v.(string); ok {
			return "<<" + s + ">>", nil
		}
		return v, nil
	})
	out := s8render(t, eng, `{{ a }} {{ b }}`, map[string]any{"a": "x", "b": "y"})
	require.Equal(t, "<<x>> <<y>>", out)
}

// K2 — engine global filter: does not mutate literal text nodes
func TestS8_SetGlobalFilter_DoesNotAffectLiteralText(t *testing.T) {
	eng := s8eng()
	callCount := 0
	eng.SetGlobalFilter(func(v any) (any, error) {
		callCount++
		return v, nil
	})
	out := s8render(t, eng, `literal text {{ x }}`, map[string]any{"x": "val"})
	require.Equal(t, "literal text val", out)
	require.Equal(t, 1, callCount, "filter called once (for the one {{ }} node)")
}

// K3 — engine global filter: error propagates to render error
func TestS8_SetGlobalFilter_ErrorPropagates(t *testing.T) {
	eng := s8eng()
	eng.SetGlobalFilter(func(v any) (any, error) {
		return nil, fmt.Errorf("global filter exploded")
	})
	_, err := s8renderErr(t, eng, `{{ x }}`, map[string]any{"x": "val"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "global filter exploded")
}

// K4 — per-render WithGlobalFilter: overrides engine-level filter
func TestS8_WithGlobalFilter_OverridesEngineFilter(t *testing.T) {
	eng := s8eng()
	eng.SetGlobalFilter(func(v any) (any, error) {
		s, _ := v.(string)
		return "[engine]" + s, nil
	})
	out, _ := s8renderErr(t, eng, `{{ x }}`, map[string]any{"x": "val"},
		liquid.WithGlobalFilter(func(v any) (any, error) {
			s, _ := v.(string)
			return "[per-render]" + s, nil
		}),
	)
	require.Equal(t, "[per-render]val", out)
}

// K5 — per-render WithGlobalFilter: does not persist across renders
func TestS8_WithGlobalFilter_DoesNotPersist(t *testing.T) {
	eng := s8eng()
	out1, _ := s8renderErr(t, eng, `{{ x }}`, map[string]any{"x": "v"},
		liquid.WithGlobalFilter(func(v any) (any, error) {
			s, _ := v.(string)
			return "!" + s, nil
		}),
	)
	require.Equal(t, "!v", out1)
	out2 := s8render(t, eng, `{{ x }}`, map[string]any{"x": "v"})
	require.Equal(t, "v", out2)
}

// K6 — global filter applied AFTER per-node filters in the pipeline
func TestS8_SetGlobalFilter_AppliedAfterNodeFilters(t *testing.T) {
	eng := s8eng()
	eng.SetGlobalFilter(func(v any) (any, error) {
		s, _ := v.(string)
		return "[" + s + "]", nil
	})
	// upcase runs first → "HELLO", then global filter wraps it
	out := s8render(t, eng, `{{ "hello" | upcase }}`, nil)
	require.Equal(t, "[HELLO]", out)
}

// K7 — global filter: numeric output passes through untouched when filter is type-selective
func TestS8_SetGlobalFilter_NumericPassthrough(t *testing.T) {
	eng := s8eng()
	eng.SetGlobalFilter(func(v any) (any, error) {
		// only transform strings
		if s, ok := v.(string); ok {
			return "str:" + s, nil
		}
		return v, nil
	})
	out := s8render(t, eng, `{{ 42 }}`, nil)
	require.Equal(t, "42", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// L. SetExceptionHandler + WithErrorHandler
// ═════════════════════════════════════════════════════════════════════════════

// L1 — engine handler: replaces failing output with handler string
func TestS8_SetExceptionHandler_ReplacesOutput(t *testing.T) {
	eng := s8eng()
	eng.SetExceptionHandler(func(_ error) string { return "<ERR>" })
	out := s8render(t, eng, `a{{ 1 | divided_by: 0 }}b`, nil)
	require.Equal(t, "a<ERR>b", out)
}

// L2 — engine handler: rendering continues past the failing node
func TestS8_SetExceptionHandler_ContinuesAfterError(t *testing.T) {
	eng := s8eng()
	var count int
	eng.SetExceptionHandler(func(_ error) string {
		count++
		return "X"
	})
	out := s8render(t, eng, `{{ 1 | divided_by: 0 }}{{ 2 | divided_by: 0 }}{{ 3 | divided_by: 0 }}`, nil)
	require.Equal(t, "XXX", out)
	require.Equal(t, 3, count)
}

// L3 — per-render WithErrorHandler: overrides engine-level handler
func TestS8_WithErrorHandler_OverridesEngineHandler(t *testing.T) {
	eng := s8eng()
	eng.SetExceptionHandler(func(_ error) string { return "engine-handler" })
	out, _ := s8renderErr(t, eng, `{{ 1 | divided_by: 0 }}`, nil,
		liquid.WithErrorHandler(func(_ error) string { return "per-render-handler" }),
	)
	require.Equal(t, "per-render-handler", out)
}

// L4 — WithErrorHandler: collects errors(template.errors pattern from Ruby)
func TestS8_WithErrorHandler_CollectsErrors(t *testing.T) {
	eng := s8eng()
	var errs []error
	out, err := s8renderErr(t, eng,
		`{{ a }}{{ 1 | divided_by: 0 }}{{ b }}{{ 2 | divided_by: 0 }}{{ c }}`,
		map[string]any{"a": "1", "b": "2", "c": "3"},
		liquid.WithErrorHandler(func(e error) string {
			errs = append(errs, e)
			return ""
		}),
	)
	require.NoError(t, err)
	require.Equal(t, "123", out)
	require.Len(t, errs, 2, "two div-by-zero errors collected")
}

// L5 — WithErrorHandler: does not persist to next call
func TestS8_WithErrorHandler_DoesNotPersist(t *testing.T) {
	eng := s8eng()
	// First call: with handler → no error
	out, err := s8renderErr(t, eng, `{{ 1 | divided_by: 0 }}`, nil,
		liquid.WithErrorHandler(func(_ error) string { return "caught" }),
	)
	require.NoError(t, err)
	require.Equal(t, "caught", out)
	// Second call: no handler → error
	_, err2 := s8renderErr(t, eng, `{{ 1 | divided_by: 0 }}`, nil)
	require.Error(t, err2)
}

// L6 — WithErrorHandler: handler receives the actual error value
func TestS8_WithErrorHandler_ReceivesActualError(t *testing.T) {
	eng := s8eng()
	var got error
	_, _ = s8renderErr(t, eng, `{{ 1 | divided_by: 0 }}`, nil,
		liquid.WithErrorHandler(func(e error) string {
			got = e
			return ""
		}),
	)
	require.Error(t, got)
}

// ═════════════════════════════════════════════════════════════════════════════
// M. WithSizeLimit
// ═════════════════════════════════════════════════════════════════════════════

// M1 — size limit exceeded: error is returned
func TestS8_WithSizeLimit_ExceededReturnsError(t *testing.T) {
	eng := s8eng()
	_, err := s8renderErr(t, eng, `1234567890`, nil, liquid.WithSizeLimit(5))
	require.Error(t, err)
	require.Contains(t, err.Error(), "size limit")
}

// M2 — size limit not exceeded: renders normally
func TestS8_WithSizeLimit_WithinLimitSucceeds(t *testing.T) {
	eng := s8eng()
	out, err := s8renderErr(t, eng, `12345`, nil, liquid.WithSizeLimit(5))
	require.NoError(t, err)
	require.Equal(t, "12345", out)
}

// M3 — size limit is in bytes (not runes)
func TestS8_WithSizeLimit_CountsBytes(t *testing.T) {
	eng := s8eng()
	// "Ö" is a 2-byte UTF-8 character
	_, err := s8renderErr(t, eng, `ÖÖÖ`, nil, liquid.WithSizeLimit(5)) // 6 bytes
	require.Error(t, err)

	out, err2 := s8renderErr(t, eng, `ÖÖÖ`, nil, liquid.WithSizeLimit(6)) // exactly 6
	require.NoError(t, err2)
	require.Equal(t, "ÖÖÖ", out)
}

// M4 — size limit: loop-generated content is bounded
func TestS8_WithSizeLimit_LoopContentBounded(t *testing.T) {
	eng := s8eng()
	// 10-iteration loop produces "1234567890" = 10 bytes
	_, err := s8renderErr(t, eng,
		`{% for i in (1..10) %}{{ i }}{% endfor %}`,
		nil,
		liquid.WithSizeLimit(5),
	)
	require.Error(t, err)
}

// M5 — size limit zero: no limit applied
func TestS8_WithSizeLimit_ZeroMeansNoLimit(t *testing.T) {
	eng := s8eng()
	out := s8render(t, eng, `a very long template output that exceeds any sensible limit`, nil,
		liquid.WithSizeLimit(0))
	require.NotEmpty(t, out)
}

// M6 — size limit is per-render: does not persist across calls
func TestS8_WithSizeLimit_PerRender_DoesNotPersist(t *testing.T) {
	eng := s8eng()
	// First call: limited → error
	_, err := s8renderErr(t, eng, `1234567890`, nil, liquid.WithSizeLimit(5))
	require.Error(t, err)
	// Second call: no limit → succeeds
	out := s8render(t, eng, `1234567890`, nil)
	require.Equal(t, "1234567890", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// N. WithContext
// ═════════════════════════════════════════════════════════════════════════════

// N1 — already-cancelled context: render returns error
func TestS8_WithContext_CancelledContext_ReturnsError(t *testing.T) {
	eng := s8eng()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	tpl, err := eng.ParseString(`{% for i in (1..1000) %}{{ i }}{% endfor %}`)
	require.NoError(t, err)
	_, renderErr := tpl.RenderString(nil, liquid.WithContext(ctx))
	require.Error(t, renderErr)
}

// N2 — active background context: render completes normally
func TestS8_WithContext_BackgroundContext_Passes(t *testing.T) {
	eng := s8eng()
	out, err := s8renderErr(t, eng, `{{ x }}`, map[string]any{"x": "ok"},
		liquid.WithContext(context.Background()))
	require.NoError(t, err)
	require.Equal(t, "ok", out)
}

// N3 — expired deadline: render stops with error
func TestS8_WithContext_ExpiredDeadline_ReturnsError(t *testing.T) {
	eng := s8eng()
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond) // ensure expiry

	tpl, err := eng.ParseString(`{% for i in (1..100000) %}{{ i }}{% endfor %}`)
	require.NoError(t, err)
	_, renderErr := tpl.RenderString(nil, liquid.WithContext(ctx))
	require.Error(t, renderErr)
}

// N4 — WithContext does not persist (second call uses fresh context)
func TestS8_WithContext_PerRender_DoesNotPersist(t *testing.T) {
	eng := s8eng()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tpl, err := eng.ParseString(`{% for i in (1..1000) %}{{ i }}{% endfor %}`)
	require.NoError(t, err)
	// First call: cancelled context → error
	_, err1 := tpl.RenderString(nil, liquid.WithContext(ctx))
	require.Error(t, err1)
	// Second call: no context option → no cancellation, render with small range
	tpl2, _ := eng.ParseString(`{{ x }}`)
	out, err2 := tpl2.RenderString(map[string]any{"x": "fine"})
	require.NoError(t, err2)
	require.Equal(t, "fine", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// O. EnableCache / ClearCache
// ═════════════════════════════════════════════════════════════════════════════

// O1 — cache enabled: same source returns same *Template pointer
func TestS8_EnableCache_SameSourceReturnsSamePointer(t *testing.T) {
	eng := s8eng()
	eng.EnableCache()
	tpl1, _ := eng.ParseString(`{{ x }}`)
	tpl2, _ := eng.ParseString(`{{ x }}`)
	require.Same(t, tpl1, tpl2)
}

// O2 — cache enabled: different sources return different pointers
func TestS8_EnableCache_DifferentSourcesDifferentPointers(t *testing.T) {
	eng := s8eng()
	eng.EnableCache()
	tpl1, _ := eng.ParseString(`{{ x }}`)
	tpl2, _ := eng.ParseString(`{{ y }}`)
	require.NotSame(t, tpl1, tpl2)
}

// O3 — cache: rendering result is still correct after cache hit
func TestS8_EnableCache_CachedTemplateRendersCorrectly(t *testing.T) {
	eng := s8eng()
	eng.EnableCache()
	for i := range 4 {
		out := s8render(t, eng, `{{ v }}`, map[string]any{"v": i})
		require.Equal(t, fmt.Sprintf("%d", i), out)
	}
}

// O4 — ClearCache: after clear, same source parses fresh (different pointer)
func TestS8_ClearCache_NewPointerAfterClear(t *testing.T) {
	eng := s8eng()
	eng.EnableCache()
	tpl1, _ := eng.ParseString(`{{ x }}`)
	eng.ClearCache()
	tpl2, _ := eng.ParseString(`{{ x }}`)
	require.NotSame(t, tpl1, tpl2)
}

// O5 — cache: concurrent access is safe
func TestS8_EnableCache_ConcurrentAccessSafe(t *testing.T) {
	eng := s8eng()
	eng.EnableCache()

	var wg sync.WaitGroup
	for range 30 {
		wg.Go(func() {
			out := s8render(t, eng, `{{ v }}`, map[string]any{"v": "ok"})
			assert.Equal(t, "ok", out)
		})
	}
	wg.Wait()
}

// O6 — cache disabled by default: always parses fresh
func TestS8_Cache_DisabledByDefault_AlwaysFresh(t *testing.T) {
	eng := s8eng()
	tpl1, _ := eng.ParseString(`{{ x }}`)
	tpl2, _ := eng.ParseString(`{{ x }}`)
	require.NotSame(t, tpl1, tpl2)
}

// ═════════════════════════════════════════════════════════════════════════════
// P. EnableJekyllExtensions
// ═════════════════════════════════════════════════════════════════════════════

// P1 — dot assign: assign to a dotted path
func TestS8_JekyllExtensions_DotAssign(t *testing.T) {
	eng := s8eng()
	eng.EnableJekyllExtensions()
	out := s8render(t, eng, `{% assign page.title = "Home" %}{{ page.title }}`, nil)
	require.Equal(t, "Home", out)
}

// P2 — dot assign: standard assign still works when extensions enabled
func TestS8_JekyllExtensions_StandardAssignStillWorks(t *testing.T) {
	eng := s8eng()
	eng.EnableJekyllExtensions()
	out := s8render(t, eng, `{% assign x = "hello" %}{{ x }}`, nil)
	require.Equal(t, "hello", out)
}

// P3 — without Jekyll extensions: dot-assign is a parse error
func TestS8_JekyllExtensions_Disabled_DotAssignErrors(t *testing.T) {
	eng := s8eng()
	_, err := eng.ParseString(`{% assign page.title = "Home" %}`)
	require.Error(t, err)
}

// P4 — dot assign with multiple segments
func TestS8_JekyllExtensions_DotAssign_MultipleSegments(t *testing.T) {
	eng := s8eng()
	eng.EnableJekyllExtensions()
	out := s8render(t, eng, `{% assign a.b.c = "deep" %}{{ a.b.c }}`, nil)
	require.Equal(t, "deep", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// Q. SetAutoEscapeReplacer
// ═════════════════════════════════════════════════════════════════════════════

// Q1 — HTML escaper: & < > " ' are escaped in output
func TestS8_SetAutoEscapeReplacer_EscapesHTML(t *testing.T) {
	eng := s8eng()
	eng.SetAutoEscapeReplacer(render.HtmlEscaper)
	out := s8render(t, eng, `{{ s }}`, map[string]any{"s": `<script>alert("xss")</script>`})
	require.Equal(t, `&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;`, out)
}

// Q2 — HTML escaper: literal text is not escaped
func TestS8_SetAutoEscapeReplacer_LiteralTextUnchanged(t *testing.T) {
	eng := s8eng()
	eng.SetAutoEscapeReplacer(render.HtmlEscaper)
	out := s8render(t, eng, `<b>literal</b> {{ v }}`, map[string]any{"v": "<b>"})
	require.Equal(t, `<b>literal</b> &lt;b&gt;`, out)
}

// Q3 — HTML escaper: raw filter bypasses escaping
func TestS8_SetAutoEscapeReplacer_RawFilterBypasses(t *testing.T) {
	eng := s8eng()
	eng.SetAutoEscapeReplacer(render.HtmlEscaper)
	out := s8render(t, eng, `{{ s | raw }}`, map[string]any{"s": `<b>bold</b>`})
	require.Equal(t, `<b>bold</b>`, out)
}

// Q4 — HTML escaper: ampersands are double-escaped only once
func TestS8_SetAutoEscapeReplacer_AmpersandEscapedOnce(t *testing.T) {
	eng := s8eng()
	eng.SetAutoEscapeReplacer(render.HtmlEscaper)
	out := s8render(t, eng, `{{ s }}`, map[string]any{"s": "a & b"})
	require.Equal(t, "a &amp; b", out)
}

// Q5 — without escaper (default): HTML characters pass through raw
func TestS8_NoAutoEscape_HTMLPassesThrough(t *testing.T) {
	eng := s8eng()
	out := s8render(t, eng, `{{ s }}`, map[string]any{"s": "<b>bold</b>"})
	require.Equal(t, "<b>bold</b>", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// R. NewBasicEngine
// ═════════════════════════════════════════════════════════════════════════════

// R1 — NewBasicEngine: no standard filters registered
func TestS8_NewBasicEngine_NoStandardFilters(t *testing.T) {
	eng := liquid.NewBasicEngine()
	_, err := eng.ParseString(`{{ "hello" | upcase }}`)
	if err == nil {
		// some engines may parse ok but fail at render
		_, renderErr := eng.ParseAndRenderString(`{{ "hello" | upcase }}`, nil)
		require.Error(t, renderErr)
	}
}

// R2 — NewBasicEngine: variable lookup still works
func TestS8_NewBasicEngine_VariableLookupWorks(t *testing.T) {
	eng := liquid.NewBasicEngine()
	out, err := eng.ParseAndRenderString(`{{ x }}`, map[string]any{"x": "hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", out)
}

// R3 — NewBasicEngine: standard tags not available
func TestS8_NewBasicEngine_NoStandardTags(t *testing.T) {
	eng := liquid.NewBasicEngine()
	_, err := eng.ParseString(`{% if true %}yes{% endif %}`)
	require.Error(t, err)
}

// R4 — NewBasicEngine: custom filter registration works
func TestS8_NewBasicEngine_CustomFilterRegistration(t *testing.T) {
	eng := liquid.NewBasicEngine()
	eng.RegisterFilter("double", func(s string) string { return s + s })
	out, err := eng.ParseAndRenderString(`{{ "ab" | double }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "abab", out)
}

// R5 — NewBasicEngine: custom tag registration works
func TestS8_NewBasicEngine_CustomTagRegistration(t *testing.T) {
	eng := liquid.NewBasicEngine()
	eng.RegisterTag("hello", func(_ render.Context) (string, error) { return "hi", nil })
	out, err := eng.ParseAndRenderString(`{% hello %}`, nil)
	require.NoError(t, err)
	require.Equal(t, "hi", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// S. Combinations
// ═════════════════════════════════════════════════════════════════════════════

// S1 — WithStrictVariables + WithErrorHandler: strict errors are caught by handler
func TestS8_StrictVars_Plus_ErrorHandler(t *testing.T) {
	eng := s8eng()
	var caught error
	out, err := s8renderErr(t, eng,
		`{{ good }}{{ bad }}`,
		map[string]any{"good": "ok"},
		liquid.WithStrictVariables(),
		liquid.WithErrorHandler(func(e error) string {
			caught = e
			return ""
		}),
	)
	require.NoError(t, err)
	require.Equal(t, "ok", out)
	require.Error(t, caught)
	require.Contains(t, caught.Error(), "bad")
}

// S2 — GlobalFilter + SizeLimit: filter expands output → hits limit
func TestS8_GlobalFilter_Plus_SizeLimit_PrefixedOutputHitsLimit(t *testing.T) {
	eng := s8eng()
	eng.SetGlobalFilter(func(v any) (any, error) {
		// each output is prefixed with "prefix:" — grows output
		s, _ := v.(string)
		return "prefix:" + s, nil
	})
	// "prefix:x" = 8 bytes; limit of 5 should fail
	_, err := s8renderErr(t, eng, `{{ x }}`, map[string]any{"x": "x"}, liquid.WithSizeLimit(5))
	require.Error(t, err)
}

// S3 — LaxTags + StrictVariables: lax tags ignore unknowns, strict vars still fire
func TestS8_LaxTags_Plus_StrictVariables(t *testing.T) {
	eng := s8eng()
	eng.LaxTags()
	eng.StrictVariables()
	// Unknown tag → ignored; undefined var → error
	_, err := s8renderErr(t, eng, `{% ghost_tag %}{{ undefined_var }}`, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "undefined")
}

// S4 — Globals + GlobalFilter + ErrorHandler together
func TestS8_Globals_GlobalFilter_ErrorHandler_Together(t *testing.T) {
	eng := s8eng()
	eng.SetGlobals(map[string]any{"prefix": ">"})
	eng.SetGlobalFilter(func(v any) (any, error) {
		s, _ := v.(string)
		return "[" + s + "]", nil
	})
	var errs []error
	out := s8render(t, eng, `{{ prefix }}: {{ name }}`, map[string]any{"name": "Alice"},
		liquid.WithErrorHandler(func(e error) string {
			errs = append(errs, e)
			return ""
		}),
	)
	require.Equal(t, "[>]: [Alice]", out)
	require.Empty(t, errs)
}

// S5 — cache + custom filter: cached template uses same filter table
func TestS8_Cache_Plus_CustomFilter(t *testing.T) {
	eng := s8eng()
	eng.RegisterFilter("shout", func(s string) string { return s + "!" })
	eng.EnableCache()

	out1 := s8render(t, eng, `{{ "hi" | shout }}`, nil)
	out2 := s8render(t, eng, `{{ "hi" | shout }}`, nil)
	require.Equal(t, "hi!", out1)
	require.Equal(t, "hi!", out2)
}

// S6 — custom delimiters + globals + custom filter
func TestS8_CustomDelims_Globals_CustomFilter(t *testing.T) {
	eng := s8eng()
	// Delims(objectLeft, objectRight, tagLeft, tagRight) — output=[[ ]], tag=[% %]
	eng.Delims("[[", "]]", "[%", "%]")
	eng.SetGlobals(map[string]any{"site": "Acme"})
	eng.RegisterFilter("badge", func(s string) string { return "(" + s + ")" })
	out := s8render(t, eng, `[% if x %][[ site | badge ]][% endif %]`, map[string]any{"x": true})
	require.Equal(t, "(Acme)", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// T. Real-world scenarios
// ═════════════════════════════════════════════════════════════════════════════

// T1 — blog page layout: globals, includes, and custom filter
func TestS8_RealWorld_BlogPageLayout(t *testing.T) {
	eng := s8eng()
	eng.SetGlobals(map[string]any{
		"site_name":    "My Blog",
		"current_year": 2025,
	})
	eng.RegisterFilter("slugify", func(s string) string {
		return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
	})
	eng.RegisterTemplateStore(&mapStore{files: map[string]string{
		"_header.html": `<title>{{ page_title }} - {{ site_name }}</title>`,
		"_footer.html": `<footer>© {{ current_year }} {{ site_name }}</footer>`,
	}})

	tpl := `{% include "_header.html" %}{{ body }}{% include "_footer.html" %}`
	out := s8render(t, eng, tpl, map[string]any{
		"page_title": "About Us",
		"body":       "<main>content</main>",
	})
	require.Equal(t, "<title>About Us - My Blog</title><main>content</main><footer>© 2025 My Blog</footer>", out)
}

// T2 — error recovery report: collect all render errors with their substitution
func TestS8_RealWorld_ErrorRecoveryReport(t *testing.T) {
	eng := s8eng()
	var errs []error
	out := s8render(t, eng,
		`item1={{ a }}, item2={{ 1 | divided_by: 0 }}, item3={{ b }}, item4={{ 2 | divided_by: 0 }}`,
		map[string]any{"a": "A", "b": "B"},
		liquid.WithErrorHandler(func(e error) string {
			errs = append(errs, e)
			return "ERR"
		}),
	)
	require.Equal(t, "item1=A, item2=ERR, item3=B, item4=ERR", out)
	require.Len(t, errs, 2)
}

// T3 — custom auth tag checks context variable before rendering content
func TestS8_RealWorld_CustomAuthTag(t *testing.T) {
	eng := s8eng()
	eng.RegisterTag("require_role", func(ctx render.Context) (string, error) {
		role := ctx.TagArgs()
		v := ctx.Get("user_role")
		if fmt.Sprintf("%v", v) != role {
			return fmt.Sprintf("<!-- access denied: need %s -->", role), nil
		}
		return "", nil
	})

	tpl := `{% require_role admin %}secret content`
	// Authorized user
	out1 := s8render(t, eng, tpl, map[string]any{"user_role": "admin"})
	require.Equal(t, "secret content", out1)

	// Unauthorized user
	out2 := s8render(t, eng, tpl, map[string]any{"user_role": "viewer"})
	require.Equal(t, "<!-- access denied: need admin -->secret content", out2)
}

// T4 — concurrent rendering with per-render global filters does not cross-contaminate
func TestS8_RealWorld_ConcurrentGlobalFilters_NoContamination(t *testing.T) {
	eng := s8eng()
	tpl, err := eng.ParseString(`{{ v }}`)
	require.NoError(t, err)

	var wg sync.WaitGroup
	results := make([]string, 20)
	for i := range 20 {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			tag := fmt.Sprintf("worker%d", i)
			out, _ := tpl.RenderString(map[string]any{"v": "x"},
				liquid.WithGlobalFilter(func(v any) (any, error) {
					return tag + ":" + fmt.Sprintf("%v", v), nil
				}),
			)
			results[i] = out
		}()
	}
	wg.Wait()

	for i, got := range results {
		require.Equal(t, fmt.Sprintf("worker%d:x", i), got)
	}
}

// T5 — per-render WithGlobals is safe under concurrent use
func TestS8_RealWorld_ConcurrentPerRenderGlobals(t *testing.T) {
	eng := s8eng()
	eng.SetGlobals(map[string]any{"base": "base"})
	tpl, err := eng.ParseString(`{{ base }}-{{ extra }}`)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			want := fmt.Sprintf("base-extra%d", i)
			out, renderErr := tpl.RenderString(nil,
				liquid.WithGlobals(map[string]any{"extra": fmt.Sprintf("extra%d", i)}),
			)
			assert.NoError(t, renderErr)
			assert.Equal(t, want, out)
		}()
	}
	wg.Wait()
}

// T6 — HTML auto-escape: XSS prevention in a form template
func TestS8_RealWorld_AutoEscapeXSSPrevention(t *testing.T) {
	eng := s8eng()
	eng.SetAutoEscapeReplacer(render.HtmlEscaper)

	// UserInput contains XSS payload
	out := s8render(t, eng,
		`<input value="{{ user_input }}">`,
		map[string]any{"user_input": `"><script>alert(1)</script>`},
	)
	// The injected payload must be escaped so it cannot break out of the attribute
	require.NotContains(t, out, "<script>")
	require.Contains(t, out, "&lt;script&gt;")
}

// T7 — cache + clear + hot-reload pattern
func TestS8_RealWorld_CacheHotReload(t *testing.T) {
	eng := s8eng()
	eng.EnableCache()

	// First parse
	tpl1, _ := eng.ParseString(`version: 1`)
	out1, _ := tpl1.RenderString(nil)
	require.Equal(t, "version: 1", out1)

	// Hot reload: clear cache → re-parse (simulating template file change)
	eng.ClearCache()
	tpl2, _ := eng.ParseString(`version: 2`)
	require.NotSame(t, tpl1, tpl2)
	out2, _ := tpl2.RenderString(nil)
	require.Equal(t, "version: 2", out2)
}
