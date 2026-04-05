package liquid_test

// B7 — Context / Scope — Intensive E2E Tests
//
// Covers all four sub-items of section 7:
//   7.1 Stack de escopos, get/set variáveis
//   7.2 Sub-contexto isolado (SpawnIsolated / RenderFileIsolated)
//   7.3 Registers (estado interno de tags)
//   7.4 Variáveis globais separadas do escopo
//
// These are original E2E specs — Ruby/JS don't expose the same Go-level API,
// so these tests exercise the full pipeline from binding to output.
// Intent: regression-guard to catch silent behaviour changes.

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/osteele/liquid"
	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── helper ────────────────────────────────────────────────────────────────────

func renderB7(t *testing.T, eng *liquid.Engine, tpl string, bindings map[string]any) string {
	t.Helper()
	out, err := eng.ParseAndRenderString(tpl, bindings)
	require.NoError(t, err, "template: %s", tpl)
	return out
}

func mustEngine(t *testing.T) *liquid.Engine {
	t.Helper()
	return liquid.NewEngine()
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  7.1 — STACK DE ESCOPOS: variáveis, acesso, atribuição                      ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

// ── 7.1.a — Tipos primitivos Go nas bindings ──────────────────────────────────

func TestB7_Scope_GoInt(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "42", renderB7(t, eng, `{{ n }}`, map[string]any{"n": 42}))
}

func TestB7_Scope_GoInt64(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "99", renderB7(t, eng, `{{ n }}`, map[string]any{"n": int64(99)}))
}

func TestB7_Scope_GoUint(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "7", renderB7(t, eng, `{{ n }}`, map[string]any{"n": uint(7)}))
}

func TestB7_Scope_GoFloat(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "3.14", renderB7(t, eng, `{{ n }}`, map[string]any{"n": 3.14}))
}

func TestB7_Scope_GoBoolTrue(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "yes", renderB7(t, eng,
		`{% if v %}yes{% else %}no{% endif %}`, map[string]any{"v": true}))
}

func TestB7_Scope_GoBoolFalse(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "no", renderB7(t, eng,
		`{% if v %}yes{% else %}no{% endif %}`, map[string]any{"v": false}))
}

func TestB7_Scope_GoNil(t *testing.T) {
	eng := mustEngine(t)
	// nil renders as empty string
	require.Equal(t, "", renderB7(t, eng, `{{ v }}`, map[string]any{"v": nil}))
	// nil is falsy in if
	require.Equal(t, "no", renderB7(t, eng,
		`{% if v %}yes{% else %}no{% endif %}`, map[string]any{"v": nil}))
}

func TestB7_Scope_GoString(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "hello", renderB7(t, eng, `{{ s }}`, map[string]any{"s": "hello"}))
}

func TestB7_Scope_GoSlice(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "a b c",
		renderB7(t, eng, `{{ arr | join: " " }}`, map[string]any{"arr": []string{"a", "b", "c"}}))
}

func TestB7_Scope_GoMap(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "tobi",
		renderB7(t, eng, `{{ user.name }}`, map[string]any{"user": map[string]any{"name": "tobi"}}))
}

// ── 7.1.b — Notação de acesso a propriedades ─────────────────────────────────

func TestB7_Scope_DotNotation_Nested(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": "deep",
				},
			},
		},
	}
	require.Equal(t, "deep", renderB7(t, eng, `{{ a.b.c.d }}`, bindings))
}

func TestB7_Scope_BracketStringKey(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{
		"user": map[string]any{"first-name": "Jane"},
	}
	require.Equal(t, "Jane", renderB7(t, eng, `{{ user["first-name"] }}`, bindings))
}

func TestB7_Scope_BracketVariableKey(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{
		"key":  "color",
		"item": map[string]any{"color": "blue", "size": "M"},
	}
	require.Equal(t, "blue", renderB7(t, eng, `{{ item[key] }}`, bindings))
}

func TestB7_Scope_BracketNestedVariableKey(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{
		"cfg":  map[string]any{"field": "title"},
		"item": map[string]any{"title": "Widget"},
	}
	require.Equal(t, "Widget", renderB7(t, eng, `{{ item[cfg.field] }}`, bindings))
}

func TestB7_Scope_MixedDotBracketChain(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{
		"store": map[string]any{
			"products": []map[string]any{
				{"name": "Alpha"},
				{"name": "Beta"},
			},
		},
	}
	require.Equal(t, "Alpha", renderB7(t, eng, `{{ store.products[0].name }}`, bindings))
	require.Equal(t, "Beta", renderB7(t, eng, `{{ store.products[1].name }}`, bindings))
}

func TestB7_Scope_NegativeIndex(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{"arr": []string{"x", "y", "z"}}
	require.Equal(t, "z", renderB7(t, eng, `{{ arr[-1] }}`, bindings))
	require.Equal(t, "y", renderB7(t, eng, `{{ arr[-2] }}`, bindings))
	require.Equal(t, "x", renderB7(t, eng, `{{ arr[-3] }}`, bindings))
}

func TestB7_Scope_FirstLastOnArray(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{"arr": []int{10, 20, 30}}
	require.Equal(t, "10", renderB7(t, eng, `{{ arr.first }}`, bindings))
	require.Equal(t, "30", renderB7(t, eng, `{{ arr.last }}`, bindings))
}

func TestB7_Scope_FirstLastInMiddleOfChain(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{
		"products": map[string]any{
			"variants": []map[string]any{
				{"title": "Small"},
				{"title": "Large"},
			},
		},
	}
	require.Equal(t, "Small", renderB7(t, eng, `{{ products.variants.first.title }}`, bindings))
	require.Equal(t, "Large", renderB7(t, eng, `{{ products.variants.last.title }}`, bindings))
}

func TestB7_Scope_SizeOnString(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "5", renderB7(t, eng, `{{ s.size }}`, map[string]any{"s": "hello"}))
}

func TestB7_Scope_SizeOnArray(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "4", renderB7(t, eng, `{{ a.size }}`, map[string]any{"a": []int{1, 2, 3, 4}}))
}

func TestB7_Scope_SizeOnMap(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{"m": map[string]int{"a": 1, "b": 2, "c": 3}}
	require.Equal(t, "3", renderB7(t, eng, `{{ m.size }}`, bindings))
}

func TestB7_Scope_ExplicitSizeKeyInMapWins(t *testing.T) {
	// Hash with explicit "size" key: that value should be returned, not len(map).
	eng := mustEngine(t)
	bindings := map[string]any{"m": map[string]any{"x": 1, "y": 2, "size": 999}}
	require.Equal(t, "999", renderB7(t, eng, `{{ m.size }}`, bindings))
}

func TestB7_Scope_HyphenatedKey(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "godz", renderB7(t, eng, `{{ oh-my }}`, map[string]any{"oh-my": "godz"}))
}

func TestB7_Scope_ArrayBracketStringReturnsNil(t *testing.T) {
	// array["first"] should be nil — bracket string access on an array is not the same as .first
	eng := mustEngine(t)
	bindings := map[string]any{"arr": []int{1, 2, 3}}
	require.Equal(t, "", renderB7(t, eng, `{{ arr["first"] }}`, bindings))
}

// ── 7.1.c — Assign semantics ─────────────────────────────────────────────────

func TestB7_Scope_AssignTopLevel(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "42", renderB7(t, eng, `{% assign x = 42 %}{{ x }}`, nil))
}

func TestB7_Scope_AssignOverwriteExisting(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "new", renderB7(t, eng,
		`{% assign v = "new" %}{{ v }}`,
		map[string]any{"v": "original"}))
}

func TestB7_Scope_AssignInsideIfPersistsAfter(t *testing.T) {
	eng := mustEngine(t)
	// assign inside an if block still lives at top scope
	require.Equal(t, "inner", renderB7(t, eng,
		`{% if true %}{% assign x = "inner" %}{% endif %}{{ x }}`, nil))
}

func TestB7_Scope_AssignInsideForPersistsAfter(t *testing.T) {
	eng := mustEngine(t)
	// assign inside for loop: last written value survives the loop
	require.Equal(t, "3", renderB7(t, eng,
		`{% for i in (1..3) %}{% assign last = i %}{% endfor %}{{ last }}`, nil))
}

func TestB7_Scope_AssignDoesNotBleedBetweenRenders(t *testing.T) {
	eng := mustEngine(t)
	// First render sets x
	out1, err := eng.ParseAndRenderString(`{% assign x = "set" %}{{ x }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "set", out1)

	// Second render must NOT see the x from the first render
	out2, err := eng.ParseAndRenderString(`{{ x }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "", out2)
}

func TestB7_Scope_ForLoopVarRestoredAfterLoop(t *testing.T) {
	eng := mustEngine(t)
	// `i` was nil before. After the loop it should be nil again (empty output).
	require.Equal(t, "123", renderB7(t, eng,
		`{% for i in (1..3) %}{{ i }}{% endfor %}{{ i }}`, nil))
}

func TestB7_Scope_OuterVarVisibleInsideLoop(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "outerouter", renderB7(t, eng,
		`{% assign x = "outer" %}{% for i in (1..2) %}{{ x }}{% endfor %}`, nil))
}

func TestB7_Scope_CaptureBlock(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "hello world", renderB7(t, eng,
		`{% capture msg %}hello world{% endcapture %}{{ msg }}`, nil))
}

func TestB7_Scope_CapturePreservesWhitespace(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "  two spaces  ", renderB7(t, eng,
		`{% capture s %}  two spaces  {% endcapture %}{{ s }}`, nil))
}

func TestB7_Scope_CaptureCanUseVariables(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{"name": "Alice"}
	require.Equal(t, "hello Alice", renderB7(t, eng,
		`{% capture greeting %}hello {{ name }}{% endcapture %}{{ greeting }}`, bindings))
}

func TestB7_Scope_MissingVariableIsEmpty(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "", renderB7(t, eng, `{{ missing }}`, nil))
}

func TestB7_Scope_MissingVariableEqualsNil(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "yes", renderB7(t, eng,
		`{% if missing == nil %}yes{% endif %}`, nil))
}

func TestB7_Scope_DeepMissingIsEmpty(t *testing.T) {
	eng := mustEngine(t)
	// a exists but a.x.y.z doesn't — should silently be empty
	bindings := map[string]any{"a": map[string]any{"x": nil}}
	require.Equal(t, "", renderB7(t, eng, `{{ a.x.y.z }}`, bindings))
}

// ── 7.1.d — Go struct bindings ───────────────────────────────────────────────

type b7Product struct {
	Name  string
	Price float64
	Tags  []string
}

func (p b7Product) ToLiquid() any {
	return map[string]any{
		"name":  p.Name,
		"price": p.Price,
		"tags":  p.Tags,
	}
}

func TestB7_Scope_StructViaDrop(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{
		"product": b7Product{Name: "Widget", Price: 9.99, Tags: []string{"sale", "new"}},
	}
	require.Equal(t, "Widget 9.99 sale",
		renderB7(t, eng, `{{ product.name }} {{ product.price }} {{ product.tags.first }}`, bindings))
}

type b7UserDrop struct {
	FirstName string `liquid:"first_name"`
	LastName  string `liquid:"last_name"`
}

func TestB7_Scope_StructWithLiquidTags(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{
		"user": b7UserDrop{FirstName: "Jane", LastName: "Doe"},
	}
	require.Equal(t, "Jane Doe",
		renderB7(t, eng, `{{ user.first_name }} {{ user.last_name }}`, bindings))
}

// ── 7.1.e — Scope visibility through filters ─────────────────────────────────

func TestB7_Scope_FilterChainOnScopedVar(t *testing.T) {
	eng := mustEngine(t)
	bindings := map[string]any{"items": []string{"banana", "apple", "cherry"}}
	require.Equal(t, "apple", renderB7(t, eng, `{{ items | sort | first }}`, bindings))
}

func TestB7_Scope_AssignThenFilter(t *testing.T) {
	eng := mustEngine(t)
	require.Equal(t, "HELLO WORLD", renderB7(t, eng,
		`{% assign s = "hello world" %}{{ s | upcase }}`, nil))
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  7.2 — SUB-CONTEXTO ISOLADO                                                 ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

// cachePartial is a helper that compiles and caches a partial in the engine.
func cachePartial(t *testing.T, eng *liquid.Engine, filename, src string) {
	t.Helper()
	_, err := eng.ParseTemplateAndCache([]byte(src), filename, 1)
	require.NoError(t, err)
}

// ── 7.2.a — Parent scope NOT visible in isolated sub-context ─────────────────

func TestB7_Isolated_ParentVarNotVisible(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("render_isolated", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("partial.html", map[string]any{})
	})
	cachePartial(t, eng, "partial.html", `{{ secret }}`)

	out, err := eng.ParseAndRenderString(
		`{% render_isolated %}`,
		map[string]any{"secret": "top-secret"},
	)
	require.NoError(t, err)
	require.Equal(t, "", out, "parent binding must not leak into isolated context")
}

func TestB7_Isolated_MultipleParentVarsNotVisible(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("render_multi_isolated", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("multi_partial.html", map[string]any{})
	})
	cachePartial(t, eng, "multi_partial.html", `{{ a }}|{{ b }}|{{ c }}`)

	out, err := eng.ParseAndRenderString(
		`{% render_multi_isolated %}`,
		map[string]any{"a": "A", "b": "B", "c": "C"},
	)
	require.NoError(t, err)
	require.Equal(t, "||", out)
}

// ── 7.2.b — Explicitly passed bindings ARE visible ───────────────────────────

func TestB7_Isolated_ExplicitBindingsVisible(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("render_with_title", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("passed.html", map[string]any{
			"title": "Explicit Title",
		})
	})
	cachePartial(t, eng, "passed.html", `{{ title }}`)

	out, err := eng.ParseAndRenderString(`{% render_with_title %}`, nil)
	require.NoError(t, err)
	require.Equal(t, "Explicit Title", out)
}

func TestB7_Isolated_ExplicitWinsOverGlobal(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"color": "global-blue"})
	eng.RegisterTag("render_color", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("color_partial.html", map[string]any{
			"color": "local-red",
		})
	})
	cachePartial(t, eng, "color_partial.html", `{{ color }}`)

	out, err := eng.ParseAndRenderString(`{% render_color %}`, nil)
	require.NoError(t, err)
	require.Equal(t, "local-red", out)
}

// ── 7.2.c — Globals DO propagate to isolated sub-context ─────────────────────

func TestB7_Isolated_GlobalsPropagate(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{
		"site_name": "MySite",
		"version":   "2.0",
	})
	eng.RegisterTag("render_with_globals", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("globals_partial.html", map[string]any{})
	})
	cachePartial(t, eng, "globals_partial.html", `{{ site_name }} v{{ version }}`)

	out, err := eng.ParseAndRenderString(`{% render_with_globals %}`, nil)
	require.NoError(t, err)
	require.Equal(t, "MySite v2.0", out)
}

func TestB7_Isolated_GlobalsVisibleButParentScopeNot(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"env": "production"})
	eng.RegisterTag("render_env", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("env_partial.html", map[string]any{})
	})
	cachePartial(t, eng, "env_partial.html", `env={{ env }} user={{ user }}`)

	out, err := eng.ParseAndRenderString(`{% render_env %}`,
		map[string]any{"user": "admin"}) // parent-only, must NOT leak
	require.NoError(t, err)
	require.Equal(t, "env=production user=", out)
}

// ── 7.2.d — Assign inside isolated sub-context doesn't leak to parent ─────────

func TestB7_Isolated_AssignDoesNotLeakToParent(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("render_assign_isolated", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("assign_partial.html", map[string]any{})
	})
	cachePartial(t, eng, "assign_partial.html", `{% assign leaked = "leaking" %}inner={{ leaked }}`)

	// After rendering the isolated partial, "leaked" must not be visible in the parent.
	out, err := eng.ParseAndRenderString(
		`{% render_assign_isolated %} parent={{ leaked }}`,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "inner=leaking parent=", out)
}

// ── 7.2.e — Isolated context works with complex templates ─────────────────────

func TestB7_Isolated_PartialWithForLoop(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("render_list", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("loop_partial.html", map[string]any{
			"items": []string{"a", "b", "c"},
		})
	})
	cachePartial(t, eng, "loop_partial.html",
		`{% for item in items %}{{ item }},{% endfor %}`)

	out, err := eng.ParseAndRenderString(`{% render_list %}`, nil)
	require.NoError(t, err)
	require.Equal(t, "a,b,c,", out)
}

func TestB7_Isolated_SequentialIsolatedCallsAreIndependent(t *testing.T) {
	eng := mustEngine(t)
	callN := 0
	eng.RegisterTag("render_seq", func(c render.Context) (string, error) {
		callN++
		n := callN
		return c.RenderFileIsolated("seq_partial.html", map[string]any{
			"val": fmt.Sprintf("call%d", n),
		})
	})
	cachePartial(t, eng, "seq_partial.html", `{{ val }}`)

	out, err := eng.ParseAndRenderString(
		`{% render_seq %}|{% render_seq %}|{% render_seq %}`,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "call1|call2|call3", out)
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  7.3 — REGISTERS: estado interno de tags                                    ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

// ── 7.3.a — ctx.Set / ctx.Get persist within a single render ─────────────────

func TestB7_Registers_SetGetWithinRender(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("set_val", func(c render.Context) (string, error) {
		c.Set(".reg_val", 42)
		return "", nil
	})
	eng.RegisterTag("get_val", func(c render.Context) (string, error) {
		v := c.Get(".reg_val")
		return fmt.Sprintf("%v", v), nil
	})

	out, err := eng.ParseAndRenderString(`{% set_val %}{% get_val %}`, nil)
	require.NoError(t, err)
	require.Equal(t, "42", out)
}

func TestB7_Registers_AccumulatingStateAcrossTagCalls(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("inc_counter", func(c render.Context) (string, error) {
		const key = ".b7_counter"
		n := 0
		if v := c.Get(key); v != nil {
			n = v.(int)
		}
		n++
		c.Set(key, n)
		return fmt.Sprintf("%d", n), nil
	})

	out, err := eng.ParseAndRenderString(
		`{% inc_counter %}{% inc_counter %}{% inc_counter %}{% inc_counter %}{% inc_counter %}`,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "12345", out)
}

func TestB7_Registers_StateVisibleInsideLoop(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("touch", func(c render.Context) (string, error) {
		const key = ".b7_touch"
		n := 0
		if v := c.Get(key); v != nil {
			n = v.(int)
		}
		n++
		c.Set(key, n)
		return "", nil
	})
	eng.RegisterTag("read_touch", func(c render.Context) (string, error) {
		v := c.Get(".b7_touch")
		return fmt.Sprintf("%v", v), nil
	})

	out, err := eng.ParseAndRenderString(
		`{% for i in (1..3) %}{% touch %}{% endfor %}{% read_touch %}`,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "3", out)
}

// ── 7.3.b — State resets between render calls ─────────────────────────────────

func TestB7_Registers_ResetBetweenRenders(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("count2", func(c render.Context) (string, error) {
		const key = ".b7_count2"
		n := 0
		if v := c.Get(key); v != nil {
			n = v.(int)
		}
		n++
		c.Set(key, n)
		return fmt.Sprintf("%d", n), nil
	})

	tpl := `{% count2 %}{% count2 %}`
	for i := range 5 {
		out, err := eng.ParseAndRenderString(tpl, nil)
		require.NoError(t, err, "render %d", i+1)
		require.Equal(t, "12", out, "render %d must reset state", i+1)
	}
}

func TestB7_Registers_ResetBetweenTemplateInstances(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("cnt3", func(c render.Context) (string, error) {
		const key = ".b7_cnt3"
		n := 0
		if v := c.Get(key); v != nil {
			n = v.(int)
		}
		n++
		c.Set(key, n)
		return fmt.Sprintf("%d", n), nil
	})

	t1, _ := eng.ParseString(`{% cnt3 %}`)
	t2, _ := eng.ParseString(`{% cnt3 %}{% cnt3 %}`)

	out1, err := t1.RenderString(nil)
	require.NoError(t, err)
	require.Equal(t, "1", out1)

	out2, err := t2.RenderString(nil)
	require.NoError(t, err)
	require.Equal(t, "12", out2)

	// Running t1 again must still give "1", not "2"
	out1b, err := t1.RenderString(nil)
	require.NoError(t, err)
	require.Equal(t, "1", out1b)
}

// ── 7.3.c — ctx.Set value visible via {{ variable }} in same render ───────────

func TestB7_Registers_SetVisibleViaTemplateOutput(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("inject_x", func(c render.Context) (string, error) {
		c.Set("x", "injected")
		return "", nil
	})

	out, err := eng.ParseAndRenderString(`{% inject_x %}{{ x }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "injected", out)
}

func TestB7_Registers_SetOverwritesBinding(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("override_color", func(c render.Context) (string, error) {
		c.Set("color", "green")
		return "", nil
	})

	out, err := eng.ParseAndRenderString(
		`before={{ color }} {% override_color %}after={{ color }}`,
		map[string]any{"color": "red"},
	)
	require.NoError(t, err)
	require.Equal(t, "before=red after=green", out)
}

// ── 7.3.d — Built-in stateful tags use registers correctly ───────────────────

func TestB7_Registers_CycleStateIsPerGroup(t *testing.T) {
	eng := mustEngine(t)
	out, err := eng.ParseAndRenderString(
		`{% for i in (1..4) %}{% cycle "a","b" %}{% endfor %}`,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "abab", out)
}

func TestB7_Registers_TwoCycleGroupsAreIndependent(t *testing.T) {
	eng := mustEngine(t)
	out, err := eng.ParseAndRenderString(
		`{% for i in (1..3) %}{% cycle "g1": "A","B" %}{% cycle "g2": "x","y","z" %}{% endfor %}`,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "AxByAz", out)
}

func TestB7_Registers_IncrementIsIsolatedFromAssign(t *testing.T) {
	eng := mustEngine(t)
	// assign sets `counter` in the scope; increment has its OWN counter named `counter`
	// They must not interfere with each other.
	out, err := eng.ParseAndRenderString(
		`{% assign counter = 100 %}{% increment counter %}{% increment counter %}{{ counter }}`,
		nil,
	)
	require.NoError(t, err)
	// increment counter: 0 then 1; assign counter: 100
	require.Equal(t, "01100", out)
}

func TestB7_Registers_DecrementIsIsolatedFromIncrementAndAssign(t *testing.T) {
	eng := mustEngine(t)
	out, err := eng.ParseAndRenderString(
		`{% assign n = 50 %}{% increment n %}{% decrement n %}{{ n }}`,
		nil,
	)
	require.NoError(t, err)
	// increment n → 0, decrement n → -1, assign n → 50
	require.Equal(t, "0-150", out)
}

// ── 7.3.e — Concurrent renders share no state ─────────────────────────────────

func TestB7_Registers_ConcurrentRendersIsolated(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("concurrent_cnt", func(c render.Context) (string, error) {
		const key = ".b7_cc"
		n := 0
		if v := c.Get(key); v != nil {
			n = v.(int)
		}
		n++
		c.Set(key, n)
		return fmt.Sprintf("%d", n), nil
	})

	tpl, err := eng.ParseString(`{% concurrent_cnt %}{% concurrent_cnt %}{% concurrent_cnt %}`)
	require.NoError(t, err)

	const goroutines = 50
	results := make([]string, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			out, e := tpl.RenderString(nil)
			assert.NoError(t, e)
			results[idx] = out
		}(i)
	}
	wg.Wait()

	for i, r := range results {
		assert.Equal(t, "123", r, "goroutine %d got wrong result", i)
	}
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  7.4 — VARIÁVEIS GLOBAIS SEPARADAS DO ESCOPO                                ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

// ── 7.4.a — Globals são acessíveis em todos os templates ─────────────────────

func TestB7_Globals_BasicAccess(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"env": "production"})
	require.Equal(t, "production", renderB7(t, eng, `{{ env }}`, nil))
}

func TestB7_Globals_MultipleGlobalsAllAccessible(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"a": "AA", "b": "BB", "c": "CC"})
	require.Equal(t, "AA-BB-CC", renderB7(t, eng, `{{ a }}-{{ b }}-{{ c }}`, nil))
}

func TestB7_Globals_NestedGlobal(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{
		"site": map[string]any{
			"name":    "MySite",
			"version": "3.0",
		},
	})
	require.Equal(t, "MySite 3.0",
		renderB7(t, eng, `{{ site.name }} {{ site.version }}`, nil))
}

func TestB7_Globals_GlobalWithNilValue(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"nothing": nil})
	require.Equal(t, "", renderB7(t, eng, `{{ nothing }}`, nil))
	require.Equal(t, "yes", renderB7(t, eng,
		`{% if nothing == nil %}yes{% endif %}`, nil))
}

func TestB7_Globals_AccessibleWithNilBindings(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"greeting": "hello"})
	require.Equal(t, "hello", renderB7(t, eng, `{{ greeting }}`, nil))
}

func TestB7_Globals_AccessibleWithEmptyBindings(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"greeting": "hello"})
	require.Equal(t, "hello", renderB7(t, eng, `{{ greeting }}`, map[string]any{}))
}

// ── 7.4.b — Bindings têm prioridade sobre globals ────────────────────────────

func TestB7_Globals_BindingShadowsGlobal(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"color": "blue"})
	require.Equal(t, "red",
		renderB7(t, eng, `{{ color }}`, map[string]any{"color": "red"}))
}

func TestB7_Globals_PartialShadow(t *testing.T) {
	// Only one key is shadowed; the other global remains visible.
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"a": "global-a", "b": "global-b"})
	require.Equal(t, "local-a global-b",
		renderB7(t, eng, `{{ a }} {{ b }}`, map[string]any{"a": "local-a"}))
}

// ── 7.4.c — assign não muta os globals para renders futuros ───────────────────

func TestB7_Globals_AssignDoesNotMutateGlobalForNextRender(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"x": "original"})

	// First render: assign shadows global within the call
	out1, err := eng.ParseAndRenderString(`{% assign x = "mutated" %}{{ x }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "mutated", out1)

	// Second render: global must still be "original"
	out2, err := eng.ParseAndRenderString(`{{ x }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "original", out2)
}

func TestB7_Globals_AssignDoesNotMutateForParallelRenders(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"counter": 0})

	tpl, err := eng.ParseString(`{% assign counter = counter %}{{ counter }}`)
	require.NoError(t, err)

	const n = 100
	errs := make([]error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = tpl.RenderString(map[string]any{"counter": idx})
		}(i)
	}
	wg.Wait()
	for _, e := range errs {
		assert.NoError(t, e)
	}

	// After N parallel renders, the global must still be 0
	out, err := eng.ParseAndRenderString(`{{ counter }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "0", out)
}

// ── 7.4.d — GetGlobals ───────────────────────────────────────────────────────

func TestB7_Globals_GetGlobalsBeforeSet(t *testing.T) {
	eng := mustEngine(t)
	require.Nil(t, eng.GetGlobals())
}

func TestB7_Globals_GetGlobalsAfterSet(t *testing.T) {
	eng := mustEngine(t)
	globals := map[string]any{"k": "v"}
	eng.SetGlobals(globals)
	require.Equal(t, globals, eng.GetGlobals())
}

// ── 7.4.e — WithGlobals per-render overrides ──────────────────────────────────

func TestB7_Globals_WithGlobalsMergesOnTopOfEngine(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"base": "engine", "override": "engine"})

	tpl, err := eng.ParseString(`{{ base }}|{{ override }}`)
	require.NoError(t, err)

	// Per-render WithGlobals should override "override" but leave "base"
	out, err := tpl.RenderString(nil, liquid.WithGlobals(map[string]any{
		"override": "per-render",
	}))
	require.NoError(t, err)
	require.Equal(t, "engine|per-render", out)
}

func TestB7_Globals_WithGlobalsDoesNotAffectSubsequentRenders(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"k": "engine"})

	tpl, err := eng.ParseString(`{{ k }}`)
	require.NoError(t, err)

	// Render once with per-render override
	out1, err := tpl.RenderString(nil, liquid.WithGlobals(map[string]any{"k": "per-render"}))
	require.NoError(t, err)
	require.Equal(t, "per-render", out1)

	// Next render must see engine global
	out2, err := tpl.RenderString(nil)
	require.NoError(t, err)
	require.Equal(t, "engine", out2)
}

func TestB7_Globals_BindingWinsOverWithGlobals(t *testing.T) {
	eng := mustEngine(t)

	tpl, err := eng.ParseString(`{{ k }}`)
	require.NoError(t, err)

	out, err := tpl.RenderString(
		map[string]any{"k": "binding"},
		liquid.WithGlobals(map[string]any{"k": "global"}),
	)
	require.NoError(t, err)
	require.Equal(t, "binding", out)
}

// ── 7.4.f — StrictVariables: globals count as "defined" ──────────────────────

func TestB7_Globals_StrictVariables_GlobalIsDefined(t *testing.T) {
	eng := mustEngine(t)
	eng.StrictVariables()
	eng.SetGlobals(map[string]any{"env": "production"})

	out, err := eng.ParseAndRenderString(`{{ env }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "production", out)
}

func TestB7_Globals_StrictVariables_UndefinedStillErrors(t *testing.T) {
	eng := mustEngine(t)
	eng.StrictVariables()
	eng.SetGlobals(map[string]any{"defined": "yes"})

	_, err := eng.ParseAndRenderString(`{{ undefined_var }}`, nil)
	require.Error(t, err)
}

func TestB7_Globals_StrictVariables_BindingDefinedToo(t *testing.T) {
	eng := mustEngine(t)
	eng.StrictVariables()

	out, err := eng.ParseAndRenderString(`{{ x }}`, map[string]any{"x": "bound"})
	require.NoError(t, err)
	require.Equal(t, "bound", out)
}

// ── 7.4.g — Globals acessíveis via custom tag through ctx.Get ─────────────────

func TestB7_Globals_AccessViaCtxGet(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"global_key": "global_value"})

	eng.RegisterTag("dump_global", func(c render.Context) (string, error) {
		v := c.Get("global_key")
		if v == nil {
			return "<nil>", nil
		}
		return fmt.Sprintf("%v", v), nil
	})

	out, err := eng.ParseAndRenderString(`{% dump_global %}`, nil)
	require.NoError(t, err)
	require.Equal(t, "global_value", out)
}

func TestB7_Globals_AccessViaCtxGetWithBindingShadow(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"key": "global"})

	eng.RegisterTag("read_key", func(c render.Context) (string, error) {
		return fmt.Sprintf("%v", c.Get("key")), nil
	})

	// Binding "key = local" should shadow the global in ctx.Get
	out, err := eng.ParseAndRenderString(`{% read_key %}`,
		map[string]any{"key": "local"})
	require.NoError(t, err)
	require.Equal(t, "local", out)
}

// ── 7.4.h — Globals acessíveis em sub-contextos isolados ──────────────────────

func TestB7_Globals_VisibleInIsolatedSubcontext(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"shared": "yep"})
	eng.RegisterTag("rfi_globals", func(c render.Context) (string, error) {
		return c.RenderFileIsolated("global_in_isolated.html", map[string]any{})
	})
	cachePartial(t, eng, "global_in_isolated.html", `{{ shared }}`)

	out, err := eng.ParseAndRenderString(`{% rfi_globals %}`, nil)
	require.NoError(t, err)
	require.Equal(t, "yep", out)
}

// ── 7.4.i — Bindings() method inclui globals ──────────────────────────────────

func TestB7_Globals_AppearsInBindings(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"g": "global"})

	var captured map[string]any
	eng.RegisterTag("capture_b", func(c render.Context) (string, error) {
		captured = c.Bindings()
		return "", nil
	})

	_, err := eng.ParseAndRenderString(`{% capture_b %}`,
		map[string]any{"local": "l"})
	require.NoError(t, err)

	require.Equal(t, "global", captured["g"], "global must be in Bindings()")
	require.Equal(t, "l", captured["local"], "local binding must be in Bindings()")
}

// ── 7.4.j — Globals em filtros ────────────────────────────────────────────────

func TestB7_Globals_UsableInFilterArg(t *testing.T) {
	eng := mustEngine(t)
	eng.SetGlobals(map[string]any{"separator": "-"})
	// separator is a global; use it in a filter via assignment
	out, err := eng.ParseAndRenderString(
		`{% assign sep = separator %}{{ "a,b,c" | split: "," | join: sep }}`,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "a-b-c", out)
}

// ── 7.4.k — WriteValue respect autoescape independently of globals ────────────

type b7SafeWriter struct{ buf strings.Builder }

func (w *b7SafeWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }

func TestB7_Globals_WriteValueNilIsEmpty(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("wv_nil", func(c render.Context) (string, error) {
		var buf b7SafeWriter
		err := c.WriteValue(&buf, nil)
		return buf.buf.String(), err
	})

	out, err := eng.ParseAndRenderString(`{% wv_nil %}`, nil)
	require.NoError(t, err)
	require.Equal(t, "", out)
}

func TestB7_Globals_WriteValueArrayIsJoined(t *testing.T) {
	eng := mustEngine(t)
	eng.RegisterTag("wv_arr", func(c render.Context) (string, error) {
		var buf b7SafeWriter
		err := c.WriteValue(&buf, c.Get("v"))
		return buf.buf.String(), err
	})

	out, err := eng.ParseAndRenderString(`{% wv_arr %}`,
		map[string]any{"v": []string{"x", "y", "z"}})
	require.NoError(t, err)
	require.Equal(t, "xyz", out)
}

// ensure b7SafeWriter satisfies io.Writer at compile time.
var _ io.Writer = (*b7SafeWriter)(nil)
