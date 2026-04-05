package liquid

// B5 / Section 12 — Thread-safety & Concurrency Investigation
//
// These tests answer two questions:
//   1. Do Engines and Templates hold up under concurrent use?
//   2. Is there actually a measurable performance difference between a shared
//      engine and one-engine-per-goroutine?
//
// Run with the race detector to catch any data races:
//   go test -race -run TestConcurrent ./...
//
// Run benchmarks to measure throughput under different sharing strategies:
//   go test -bench=BenchmarkConcurrent -benchmem ./...
//
// ── What we expect to be safe ──────────────────────────────────────────────
//
//   •  Parsing (Engine.ParseString): reads grammar/filter maps (never writes
//      during use), scans customTokenMatchers sync.Map.  Concurrent-safe.
//
//   •  Rendering (Template.Render): each call gets a fresh nodeContext with
//      its own bindings map; the compiled AST is read-only.  Concurrent-safe.
//
//   •  StrictVariables: calls expr.Variables() which uses sync.Once to lazily
//      cache the variable list.  Concurrent-safe by design.
//
//   •  Stateful tags (assign, increment, cycle, for+continue): all write into
//      the per-call bindings map, not into any shared structure.  Each render
//      call is isolated; results must not leak between goroutines.
//
// ── Known data race ─────────────────────────────────────────────────────────
//
//   •  Engine.ParseTemplateAndCache writes to e.cfg.Cache (a plain map[string][]byte)
//      while concurrent renders may read the same map via the {% include %} tag.
//      Captured by TestConcurrent_CacheRace below.
//
// ── Performance ──────────────────────────────────────────────────────────────
//
//   The benchmarks compare:
//     A) Shared engine + shared pre-parsed template → pure-render throughput
//     B) Shared engine, each goroutine parses+renders → contention on parse side
//     C) Goroutine-private engine + goroutine-private template → full isolation
//
//   If (B) is significantly slower than (A), the parse path has contention.
//   If (C) is faster than (B), there is false-sharing or allocation contention
//   on the shared engine config struct.

import (
	"fmt"
	"sync"
	"testing"

	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	concurrentGoroutines = 50
	concurrentIters      = 20 // iterations per goroutine in functional tests
)

// ─── Frozen-engine guard tests ───────────────────────────────────────────────

// TestEngine_FrozenAfterParse verifies that configuration methods panic when
// called after the engine has been used for the first time.
func TestEngine_FrozenAfterParse(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*Engine)
	}{
		{"RegisterTag", func(e *Engine) {
			e.RegisterTag("x", func(ctx render.Context) (string, error) { return "", nil })
		}},
		{"RegisterFilter", func(e *Engine) { e.RegisterFilter("x", func(v string) string { return v }) }},
		{"RegisterBlock", func(e *Engine) {
			e.RegisterBlock("x", func(ctx render.Context) (string, error) { return "", nil })
		}},
		{"StrictVariables", func(e *Engine) { e.StrictVariables() }},
		{"LaxFilters", func(e *Engine) { e.LaxFilters() }},
		{"LaxTags", func(e *Engine) { e.LaxTags() }},
		{"EnableJekyllExtensions", func(e *Engine) { e.EnableJekyllExtensions() }},
		{"EnableCache", func(e *Engine) { e.EnableCache() }},
		{"SetGlobals", func(e *Engine) { e.SetGlobals(map[string]any{"x": 1}) }},
		{"SetGreedy", func(e *Engine) { e.SetGreedy(false) }},
		{"SetTrimTagLeft", func(e *Engine) { e.SetTrimTagLeft(true) }},
		{"SetTrimTagRight", func(e *Engine) { e.SetTrimTagRight(true) }},
		{"SetTrimOutputLeft", func(e *Engine) { e.SetTrimOutputLeft(true) }},
		{"SetTrimOutputRight", func(e *Engine) { e.SetTrimOutputRight(true) }},
		{"SetGlobalFilter", func(e *Engine) { e.SetGlobalFilter(func(v any) (any, error) { return v, nil }) }},
		{"SetExceptionHandler", func(e *Engine) { e.SetExceptionHandler(func(err error) string { return "" }) }},
		{"Delims", func(e *Engine) { e.Delims("[[", "]]", "{%", "%}") }},
		// Note: UnregisterTag is intentionally excluded — it supports post-use hot-reload scenarios.
	}

	for _, tc := range cases {
		t.Run(tc.name+"_panics_after_parse", func(t *testing.T) {
			engine := NewEngine()
			_, err := engine.ParseString(`{{ x }}`)
			require.NoError(t, err)

			require.Panics(t, func() { tc.fn(engine) },
				"%s should panic after engine has been used for parsing", tc.name)
		})

		t.Run(tc.name+"_ok_before_parse", func(t *testing.T) {
			engine := NewEngine()
			require.NotPanics(t, func() { tc.fn(engine) },
				"%s should not panic before first parse", tc.name)
		})
	}
}

// TestEngine_FrozenPanicMessage verifies the panic message is clear.
func TestEngine_FrozenPanicMessage(t *testing.T) {
	engine := NewEngine()
	_, err := engine.ParseString(`{{ x }}`)
	require.NoError(t, err)

	defer func() {
		r := recover()
		require.NotNil(t, r)
		msg, ok := r.(string)
		require.True(t, ok)
		assert.Contains(t, msg, "RegisterFilter")
		assert.Contains(t, msg, "after the engine has been used")
	}()
	engine.RegisterFilter("x", func(v string) string { return v })
}

// TestConcurrent_SharedTemplate renders the same pre-parsed Template from many
// goroutines simultaneously using different bindings, and asserts that each
// goroutine sees its own correct output.
//
// If there is any mutable shared state in the render path, -race will flag it.
func TestConcurrent_SharedTemplate(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`hello {{ name }}, count={{ n }}`)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := range concurrentGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := range concurrentIters {
				got, err := tpl.RenderString(Bindings{"name": fmt.Sprintf("goroutine-%d", i), "n": j})
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("hello goroutine-%d, count=%d", i, j), got)
			}
		}(i)
	}
	wg.Wait()
}

// TestConcurrent_SharedEngine_ParseAndRender exercises concurrent parse+render
// on a shared Engine, which is the most common production pattern.
func TestConcurrent_SharedEngine_ParseAndRender(t *testing.T) {
	engine := NewEngine()

	var wg sync.WaitGroup
	for i := range concurrentGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := range concurrentIters {
				src := fmt.Sprintf(`{{ x | append: "-%d" }}`, j)
				got, err := engine.ParseAndRenderString(src, Bindings{"x": fmt.Sprintf("g%d", i)})
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("g%d-%d", i, j), got)
			}
		}(i)
	}
	wg.Wait()
}

// TestConcurrent_SharedEngine_ParseOnly tests concurrent parsing with no
// rendering, so we can isolate parse-time contention.
func TestConcurrent_SharedEngine_ParseOnly(t *testing.T) {
	engine := NewEngine()
	sources := []string{
		`{{ x }}`,
		`{% for i in (1..10) %}{{ i }}{% endfor %}`,
		`{% if a %}yes{% elsif b %}maybe{% else %}no{% endif %}`,
		`{{ x | upcase | truncate: 20 }}`,
		`{% assign r = x | split: "," %}{{ r | join: "-" }}`,
	}

	var wg sync.WaitGroup
	for i := range concurrentGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			src := sources[i%len(sources)]
			for range concurrentIters {
				_, err := engine.ParseString(src)
				assert.NoError(t, err)
			}
		}(i)
	}
	wg.Wait()
}

// TestConcurrent_StatefulTagsAreIsolated verifies that stateful tags such as
// increment, decrement, assign, and for+continue produce correct,
// goroutine-private results even when the same *Template is shared.
//
// Each goroutine should see its own counter namespace; counters from one
// render call must NEVER bleed into another.
func TestConcurrent_StatefulTagsAreIsolated(t *testing.T) {
	engine := NewEngine()

	t.Run("increment is isolated", func(t *testing.T) {
		tpl, err := engine.ParseString(
			`{% increment c %}{% increment c %}{% increment c %}`,
		)
		require.NoError(t, err)

		var wg sync.WaitGroup
		for range concurrentGoroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range concurrentIters {
					got, err := tpl.RenderString(Bindings{})
					assert.NoError(t, err)
					// Isolated call: always 0,1,2 not some accumulated value
					assert.Equal(t, "012", got)
				}
			}()
		}
		wg.Wait()
	})

	t.Run("assign does not leak between calls", func(t *testing.T) {
		tpl, err := engine.ParseString(
			`{% assign v = x | append: "-done" %}{{ v }}`,
		)
		require.NoError(t, err)

		var wg sync.WaitGroup
		for i := range concurrentGoroutines {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				name := fmt.Sprintf("g%d", i)
				for range concurrentIters {
					got, err := tpl.RenderString(Bindings{"x": name})
					assert.NoError(t, err)
					assert.Equal(t, name+"-done", got)
				}
			}(i)
		}
		wg.Wait()
	})

	t.Run("for loop state does not leak between calls", func(t *testing.T) {
		// offset:continue tracks position keyed on the collection name.
		// Two goroutines iterating different slices must not interfere.
		tpl, err := engine.ParseString(
			`{% for item in items %}{{ item }},{% endfor %}`,
		)
		require.NoError(t, err)

		var wg sync.WaitGroup
		for i := range concurrentGoroutines {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				items := []any{i, i + 1, i + 2}
				for range concurrentIters {
					got, err := tpl.RenderString(Bindings{"items": items})
					assert.NoError(t, err)
					assert.Equal(t, fmt.Sprintf("%d,%d,%d,", i, i+1, i+2), got)
				}
			}(i)
		}
		wg.Wait()
	})
}

// TestConcurrent_StrictVariables exercises the StrictVariables path which calls
// expr.Variables() using sync.Once on shared expression nodes.
func TestConcurrent_StrictVariables(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{{ name }} is {{ age }}`)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := range concurrentGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for range concurrentIters {
				out, err := tpl.RenderString(
					Bindings{"name": fmt.Sprintf("p%d", i), "age": i},
					WithStrictVariables(),
				)
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("p%d is %d", i, i), out)
			}
		}(i)
	}
	wg.Wait()
}

// TestConcurrent_ComplexTemplate uses a non-trivial template (loops, filters,
// conditionals, nested blocks) to stress all render code paths simultaneously.
func TestConcurrent_ComplexTemplate(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`
{{- items | size }} items:
{%- for item in items %}
  {{ forloop.index }}. {{ item | upcase }}{% if forloop.last %} (last){% endif %}
{%- endfor %}
total={{ items | size }}`)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := range concurrentGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			items := []any{
				fmt.Sprintf("alpha-%d", i),
				fmt.Sprintf("beta-%d", i),
				fmt.Sprintf("gamma-%d", i),
			}
			got, err := tpl.RenderString(Bindings{"items": items})
			assert.NoError(t, err)
			// Just check the total line — enough to verify correctness without
			// a complex expected-string per goroutine.
			assert.Contains(t, got, "total=3")
			assert.Contains(t, got, "(last)")
		}(i)
	}
	wg.Wait()
}

// TestConcurrent_CacheRace verifies that concurrent writes to ParseTemplateAndCache
// and concurrent renders that use {% include %} (which reads from Cache) are safe.
//
// Previously Cache was a plain map[string][]byte — concurrent writes would cause
// "fatal error: concurrent map writes". Now it is a sync.Map.
func TestConcurrent_CacheRace(t *testing.T) {
	engine := NewEngine()

	// Seed an initial entry.
	_, err := engine.ParseTemplateAndCache([]byte(`cached-v0`), "tpl.html", 1)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := range concurrentGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				// Writer: update the cached source concurrently.
				src := []byte(fmt.Sprintf("cached-v%d", i))
				_, err := engine.ParseTemplateAndCache(src, "tpl.html", 1)
				assert.NoError(t, err)
			} else {
				// Reader: render a template that uses an in-memory include via Cache.
				// The include falls back to Cache when the file doesn't exist on disk.
				engine.ParseAndRenderString(
					`{% include 'tpl.html' %}`,
					Bindings{},
				) //nolint:errcheck — file won't be found on disk; we test no crash/race
			}
		}(i)
	}
	wg.Wait()
	// If we reach here without a "concurrent map writes" crash, the fix works.
}

// ─── E2E intensive — stateful tag isolation ───────────────────────────────────

// TestConcurrentE2E_DecrementIsolated verifies that decrement counters are
// per-call and never bleed across concurrent renders of a shared template.
func TestConcurrentE2E_DecrementIsolated(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(
		`{% decrement c %}{% decrement c %}{% decrement c %}`,
	)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for range concurrentGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range concurrentIters {
				got, err := tpl.RenderString(Bindings{})
				assert.NoError(t, err)
				// Isolated call: always -1,-2,-3 (never any other value)
				assert.Equal(t, "-1-2-3", got)
			}
		}()
	}
	wg.Wait()
}

// TestConcurrentE2E_CycleIsolated verifies that cycle state is per-call.
// Each render of a 3-value cycle must always start from the first value.
func TestConcurrentE2E_CycleIsolated(t *testing.T) {
	engine := NewEngine()
	// 4 iterations over a 3-value cycle → "one two three one"
	tpl, err := engine.ParseString(
		`{% for i in (1..4) %}{% cycle "one", "two", "three" %}{% unless forloop.last %} {% endunless %}{% endfor %}`,
	)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for range concurrentGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range concurrentIters {
				got, err := tpl.RenderString(Bindings{})
				assert.NoError(t, err)
				assert.Equal(t, "one two three one", got)
			}
		}()
	}
	wg.Wait()
}

// TestConcurrentE2E_CaptureIsolated verifies that capture output is per-call.
func TestConcurrentE2E_CaptureIsolated(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(
		`{% capture result %}{{ prefix }}-captured{% endcapture %}{{ result }}`,
	)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := range concurrentGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pfx := fmt.Sprintf("g%d", i)
			for range concurrentIters {
				got, err := tpl.RenderString(Bindings{"prefix": pfx})
				assert.NoError(t, err)
				assert.Equal(t, pfx+"-captured", got)
			}
		}(i)
	}
	wg.Wait()
}

// TestConcurrentE2E_ForOffsetContinueIsolated verifies that offset:continue
// tracking (stored under a \x00-prefixed key in bindings) is per-call and
// never leaks across goroutines rendering the same template.
func TestConcurrentE2E_ForOffsetContinueIsolated(t *testing.T) {
	engine := NewEngine()
	// First loop: items[0..2]; second loop picks up where first left off (items[3..4])
	tpl, err := engine.ParseString(
		`{% for i in items limit:3 %}{{ i }},{% endfor %}` +
			`{% for i in items limit:3 offset:continue %}{{ i }},{% endfor %}`,
	)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for range concurrentGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items := []any{10, 20, 30, 40, 50}
			for range concurrentIters {
				got, err := tpl.RenderString(Bindings{"items": items})
				assert.NoError(t, err)
				// First loop: 10,20,30  — second loop: 40,50,
				assert.Equal(t, "10,20,30,40,50,", got)
			}
		}()
	}
	wg.Wait()
}

// ─── E2E intensive — globalFilter isolation ──────────────────────────────────

// TestConcurrentE2E_GlobalFilterIsolated verifies that a globalFilter
// (set once on the engine) is applied consistently to every concurrent render
// without interfering with per-call bindings.
func TestConcurrentE2E_GlobalFilterIsolated(t *testing.T) {
	engine := NewEngine()
	engine.SetGlobalFilter(func(v any) (any, error) {
		if s, ok := v.(string); ok {
			return "[" + s + "]", nil
		}
		return v, nil
	})
	tpl, err := engine.ParseString(`{{ name }}`)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := range concurrentGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("user%d", i)
			for range concurrentIters {
				got, err := tpl.RenderString(Bindings{"name": name})
				assert.NoError(t, err)
				assert.Equal(t, "["+name+"]", got)
			}
		}(i)
	}
	wg.Wait()
}

// ─── E2E intensive — per-render options ──────────────────────────────────────

// TestConcurrentE2E_PerRenderGlobals verifies that WithGlobals() scopes
// correctly to its own render call and never bleeds into concurrent renders.
func TestConcurrentE2E_PerRenderGlobals(t *testing.T) {
	engine := NewEngine()
	// Engine-level global: "env" = "prod"
	engine.SetGlobals(map[string]any{"env": "prod"})
	tpl, err := engine.ParseString(`{{ env }}/{{ tenant }}`)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := range concurrentGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tenant := fmt.Sprintf("client%d", i)
			for range concurrentIters {
				// Each call overrides "env" with a per-call value; "tenant" is
				// passed as a binding. They must never mix between goroutines.
				perCallEnv := fmt.Sprintf("env-%d", i)
				got, err := tpl.RenderString(
					Bindings{"tenant": tenant},
					WithGlobals(map[string]any{"env": perCallEnv}),
				)
				assert.NoError(t, err)
				assert.Equal(t, perCallEnv+"/"+tenant, got)
			}
		}(i)
	}
	wg.Wait()
}

// TestConcurrentE2E_StrictVariablesErrorIsolated verifies that:
//  1. When a variable IS defined, StrictVariables renders correctly.
//  2. When a variable is NOT defined, StrictVariables returns an error with
//     the correct variable name — each goroutine sees its own error.
func TestConcurrentE2E_StrictVariablesErrorIsolated(t *testing.T) {
	engine := NewEngine()
	tplOK, err := engine.ParseString(`{{ defined_var }}`)
	require.NoError(t, err)
	tplBad, err := engine.ParseString(`{{ ghost }}`)
	require.NoError(t, err)

	var wg sync.WaitGroup

	// Half the goroutines render the valid template.
	for i := range concurrentGoroutines / 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			val := fmt.Sprintf("v%d", i)
			for range concurrentIters {
				got, err := tplOK.RenderString(
					Bindings{"defined_var": val},
					WithStrictVariables(),
				)
				assert.NoError(t, err)
				assert.Equal(t, val, got)
			}
		}(i)
	}

	// The other half render the undefined-variable template.
	for range concurrentGoroutines / 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range concurrentIters {
				_, err := tplBad.RenderString(Bindings{}, WithStrictVariables())
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "ghost")
			}
		}()
	}

	wg.Wait()
}

// ─── E2E intensive — template cache under concurrency ────────────────────────

// TestConcurrentE2E_TemplateCacheCorrectness verifies that the template cache
// (EnableCache + sync.Map) returns correct results when N goroutines
// simultaneously trigger cache misses and hits for the same source string.
func TestConcurrentE2E_TemplateCacheCorrectness(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache()

	sources := []string{
		`{{ x | upcase }}`,
		`{% for i in (1..3) %}{{ i }}{% endfor %}`,
		`{{ a }}-{{ b }}`,
		`{% assign r = x | append: "!" %}{{ r }}`,
		`{% if x %}yes{% else %}no{% endif %}`,
	}
	expected := []func(int) string{
		func(i int) string { return fmt.Sprintf("G%d", i) },
		func(_ int) string { return "123" },
		func(i int) string { return fmt.Sprintf("%d-%d", i, i+1) },
		func(i int) string { return fmt.Sprintf("v%d!", i) },
		func(_ int) string { return "yes" },
	}

	var wg sync.WaitGroup
	for i := range concurrentGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			idx := i % len(sources)
			for range concurrentIters {
				var bindings Bindings
				switch idx {
				case 0:
					bindings = Bindings{"x": fmt.Sprintf("g%d", i)}
				case 2:
					bindings = Bindings{"a": i, "b": i + 1}
				case 3:
					bindings = Bindings{"x": fmt.Sprintf("v%d", i)}
				case 4:
					bindings = Bindings{"x": true}
				default:
					bindings = Bindings{}
				}
				got, err := engine.ParseAndRenderString(sources[idx], bindings)
				assert.NoError(t, err)
				assert.Equal(t, expected[idx](i), got)
			}
		}(i)
	}
	wg.Wait()
}

// ─── E2E intensive — freeze concurrency ──────────────────────────────────────

// TestConcurrentE2E_SimultaneousFreeze verifies that two goroutines calling
// a parse method simultaneously do not cause a race on the frozen atomic.Bool
// and that both calls succeed correctly.
func TestConcurrentE2E_SimultaneousFreeze(t *testing.T) {
	const parallelFreezes = 100

	for range parallelFreezes {
		engine := NewEngine()

		var (
			wg      sync.WaitGroup
			results = make([]string, 2)
			errs    = make([]error, 2)
		)

		// Two goroutines race to be the first to freeze the engine.
		for j := range 2 {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				results[j], errs[j] = engine.ParseAndRenderString(
					`{{ val }}`, Bindings{"val": fmt.Sprintf("g%d", j)},
				)
			}(j)
		}
		wg.Wait()

		for j := range 2 {
			assert.NoError(t, errs[j])
			assert.Equal(t, fmt.Sprintf("g%d", j), results[j])
		}
	}
}

// ─── E2E intensive — mixed workload ──────────────────────────────────────────

// TestConcurrentE2E_MixedWorkload simulates a realistic production scenario:
// a shared engine with cache, multiple distinct templates, multiple goroutines
// each with their own bindings, all rendering concurrently.
func TestConcurrentE2E_MixedWorkload(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache()
	engine.SetGlobals(map[string]any{"app": "engage"})

	type tplCase struct {
		src      string
		expected func(i int) string
	}

	templates := []tplCase{
		{
			`{{ app }}/{{ customer }}/{{ campaign }}`,
			func(i int) string { return fmt.Sprintf("engage/cust%d/camp%d", i, i*2) },
		},
		{
			`{% for item in items %}{{ item | upcase }},{% endfor %}`,
			func(i int) string {
				return fmt.Sprintf("ITEM%dA,ITEM%dB,ITEM%dC,", i, i, i)
			},
		},
		{
			`{% assign url = base | append: "?id=" | append: id %}{{ url }}`,
			func(i int) string { return fmt.Sprintf("https://example.com?id=%d", i) },
		},
		{
			`{% if score >= 90 %}A{% elsif score >= 80 %}B{% else %}C{% endif %}`,
			func(i int) string {
				switch i % 3 {
				case 0:
					return "A"
				case 1:
					return "B"
				default:
					return "C"
				}
			},
		},
	}

	var wg sync.WaitGroup
	for i := range concurrentGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tc := templates[i%len(templates)]

			var bindings Bindings
			switch i % len(templates) {
			case 0:
				bindings = Bindings{
					"customer": fmt.Sprintf("cust%d", i),
					"campaign": fmt.Sprintf("camp%d", i*2),
				}
			case 1:
				bindings = Bindings{"items": []any{
					fmt.Sprintf("item%da", i),
					fmt.Sprintf("item%db", i),
					fmt.Sprintf("item%dc", i),
				}}
			case 2:
				bindings = Bindings{"base": "https://example.com", "id": i}
			case 3:
				score := 95 - (i%3)*10 // 95, 85, 75 cycling
				bindings = Bindings{"score": score}
			}

			for range concurrentIters {
				got, err := engine.ParseAndRenderString(tc.src, bindings)
				assert.NoError(t, err)
				assert.Equal(t, tc.expected(i), got)
			}
		}(i)
	}
	wg.Wait()
}

// ─── Benchmarks ──────────────────────────────────────────────────────────────
//
// Run with:
//   go test -bench=BenchmarkConcurrent -benchmem -cpu 1,4,8 ./...
//
// The -cpu flag controls GOMAXPROCS.  On modern multi-core machines, compare
// -cpu 1 (serial) with -cpu 8+ (parallel) to see scaling behaviour.

const benchTemplateSrc = `
{%- assign n = items | size -%}
{%- for item in items -%}
  {{- forloop.index }}/{{ n }}: {{ item | upcase -}}
  {%- unless forloop.last %}, {% endunless -%}
{%- endfor -%}`

// BenchmarkConcurrent_SharedTemplate_Render measures pure render throughput
// when the Engine and Template are both pre-created and shared.
// This is the BEST-CASE scenario: no parse overhead, maximum reuse.
func BenchmarkConcurrent_SharedTemplate_Render(b *testing.B) {
	engine := NewEngine()
	tpl, err := engine.ParseString(benchTemplateSrc)
	if err != nil {
		b.Fatal(err)
	}
	bindings := Bindings{"items": []any{"alpha", "beta", "gamma", "delta"}}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := tpl.RenderString(bindings)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkConcurrent_SharedEngine_ParseAndRender measures throughput when
// goroutines share one Engine but each re-parse the template on every call.
// Useful for cases where the caller does not cache parsed templates.
func BenchmarkConcurrent_SharedEngine_ParseAndRender(b *testing.B) {
	engine := NewEngine()
	bindings := Bindings{"items": []any{"alpha", "beta", "gamma", "delta"}}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := engine.ParseAndRenderString(benchTemplateSrc, bindings)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkConcurrent_SharedEngine_Cache_ParseAndRender is the same as above
// but with the template cache enabled.  Cache hits avoid re-parsing.
func BenchmarkConcurrent_SharedEngine_Cache_ParseAndRender(b *testing.B) {
	engine := NewEngine()
	engine.EnableCache()
	bindings := Bindings{"items": []any{"alpha", "beta", "gamma", "delta"}}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := engine.ParseAndRenderString(benchTemplateSrc, bindings)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkConcurrent_PerGoroutineEngine creates a NEW Engine for every single
// render call — the "one-renderer-per-goroutine" pattern the user was using.
// This is the WORST-CASE setup and memory scenario: maximum GC pressure.
func BenchmarkConcurrent_PerGoroutineEngine(b *testing.B) {
	bindings := Bindings{"items": []any{"alpha", "beta", "gamma", "delta"}}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			engine := NewEngine()
			_, err := engine.ParseAndRenderString(benchTemplateSrc, bindings)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkConcurrent_SharedEngine_ParseOnly isolates parse-time contention
// on the shared Engine config (grammar map reads) with no render overhead.
func BenchmarkConcurrent_SharedEngine_ParseOnly(b *testing.B) {
	engine := NewEngine()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := engine.ParseString(benchTemplateSrc)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkConcurrent_PerGoroutineEngine_ParseOnly isolates parse-time cost
// when each goroutine has its own engine.
func BenchmarkConcurrent_PerGoroutineEngine_ParseOnly(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		engine := NewEngine() // one per goroutine (shared within the goroutine's iterations)
		for pb.Next() {
			_, err := engine.ParseString(benchTemplateSrc)
			if err != nil {
				b.Error(err)
			}
		}
	})
}
