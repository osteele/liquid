package liquid_test

// S1 — Section 1 (Tags) intensive E2E tests.
//
// Exercises ALL Section 1 tag behaviours with Go-typed bindings and
// complex template constructs. The intent is to serve as a regression
// barrier: any unintended change to Section 1 behaviour should break
// at least one test here.
//
// Sections covered:
//   1.1  Output / Expression  — {{ }}, echo
//   1.2  Variables            — assign, capture
//   1.3  Conditionals         — if/elsif/else, unless, case/when
//   1.4  Iteration            — for/else, modifiers, forloop vars,
//                               break/continue, offset:continue, cycle, tablerow
//   1.6  Structure            — raw, comment

import (
	"fmt"
	"strings"
	"testing"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/require"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func renderS1(t *testing.T, tpl string, binds map[string]any) string {
	t.Helper()
	eng := liquid.NewEngine()
	out, err := eng.ParseAndRenderString(tpl, binds)
	require.NoError(t, err, "template: %s", tpl)
	return out
}

func renderS1T(t *testing.T, tpl string) string {
	t.Helper()
	return renderS1(t, tpl, nil)
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.1  Output / Expression — {{ variable }}
// ═════════════════════════════════════════════════════════════════════════════

// ── type rendering ────────────────────────────────────────────────────────────

func TestS11_Output_String(t *testing.T) {
	require.Equal(t, "hello", renderS1(t, "{{ v }}", map[string]any{"v": "hello"}))
}

func TestS11_Output_Int(t *testing.T) {
	require.Equal(t, "42", renderS1(t, "{{ v }}", map[string]any{"v": 42}))
}

func TestS11_Output_NegativeInt(t *testing.T) {
	require.Equal(t, "-7", renderS1(t, "{{ v }}", map[string]any{"v": -7}))
}

func TestS11_Output_Float(t *testing.T) {
	require.Equal(t, "3.14", renderS1(t, "{{ v }}", map[string]any{"v": 3.14}))
}

func TestS11_Output_BoolTrue(t *testing.T) {
	require.Equal(t, "true", renderS1(t, "{{ v }}", map[string]any{"v": true}))
}

func TestS11_Output_BoolFalse(t *testing.T) {
	require.Equal(t, "false", renderS1(t, "{{ v }}", map[string]any{"v": false}))
}

func TestS11_Output_NilRendersEmpty(t *testing.T) {
	require.Equal(t, "", renderS1(t, "{{ v }}", map[string]any{"v": nil}))
}

func TestS11_Output_MissingVariableRendersEmpty(t *testing.T) {
	// unset variables are nil → render as empty string without error
	require.Equal(t, "", renderS1T(t, "{{ totally_missing }}"))
}

// ── property traversal ────────────────────────────────────────────────────────

func TestS11_Output_NestedMap(t *testing.T) {
	out := renderS1(t, "{{ user.name }}", map[string]any{
		"user": map[string]any{"name": "Alice"},
	})
	require.Equal(t, "Alice", out)
}

func TestS11_Output_DeeplyNestedMap(t *testing.T) {
	out := renderS1(t, "{{ a.b.c.d }}", map[string]any{
		"a": map[string]any{"b": map[string]any{"c": map[string]any{"d": "deep"}}},
	})
	require.Equal(t, "deep", out)
}

func TestS11_Output_GoStruct(t *testing.T) {
	type Product struct {
		Name  string
		Price float64
	}
	out := renderS1(t, "{{ p.Name }}: {{ p.Price }}", map[string]any{
		"p": Product{Name: "Widget", Price: 9.99},
	})
	require.Equal(t, "Widget: 9.99", out)
}

func TestS11_Output_NestedStruct(t *testing.T) {
	type Address struct{ City string }
	type Person struct {
		Name    string
		Address Address
	}
	out := renderS1(t, "{{ p.Name }} from {{ p.Address.City }}", map[string]any{
		"p": Person{Name: "Bob", Address: Address{City: "Paris"}},
	})
	require.Equal(t, "Bob from Paris", out)
}

func TestS11_Output_MapInStruct(t *testing.T) {
	type Wrapper struct{ Data map[string]any }
	out := renderS1(t, "{{ w.Data.key }}", map[string]any{
		"w": Wrapper{Data: map[string]any{"key": "found"}},
	})
	require.Equal(t, "found", out)
}

func TestS11_Output_NilPropertyAccess_NoError(t *testing.T) {
	// accessing a key on a nil value renders empty string, not a panic
	out := renderS1(t, "{{ x.missing }}", map[string]any{"x": nil})
	require.Equal(t, "", out)
}

func TestS11_Output_MissingNestedKey_NoError(t *testing.T) {
	out := renderS1(t, "{{ user.address.zip }}", map[string]any{
		"user": map[string]any{"name": "Alice"},
	})
	require.Equal(t, "", out)
}

// ── array access ──────────────────────────────────────────────────────────────

func TestS11_Output_ArrayIndex(t *testing.T) {
	out := renderS1(t, "{{ arr[1] }}", map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "b", out)
}

func TestS11_Output_ArrayIndex_Zero(t *testing.T) {
	out := renderS1(t, "{{ arr[0] }}", map[string]any{"arr": []int{10, 20, 30}})
	require.Equal(t, "10", out)
}

func TestS11_Output_ArrayFirst(t *testing.T) {
	out := renderS1(t, "{{ arr.first }}", map[string]any{"arr": []int{11, 22, 33}})
	require.Equal(t, "11", out)
}

func TestS11_Output_ArrayLast(t *testing.T) {
	out := renderS1(t, "{{ arr.last }}", map[string]any{"arr": []int{11, 22, 33}})
	require.Equal(t, "33", out)
}

func TestS11_Output_ArraySize(t *testing.T) {
	out := renderS1(t, "{{ arr.size }}", map[string]any{"arr": []string{"x", "y", "z"}})
	require.Equal(t, "3", out)
}

// ── filters ───────────────────────────────────────────────────────────────────

func TestS11_Output_SingleFilter(t *testing.T) {
	out := renderS1(t, "{{ name | upcase }}", map[string]any{"name": "alice"})
	require.Equal(t, "ALICE", out)
}

func TestS11_Output_FilterChain(t *testing.T) {
	out := renderS1(t, "{{ s | downcase | capitalize }}", map[string]any{"s": "HELLO WORLD"})
	require.Equal(t, "Hello world", out)
}

func TestS11_Output_FilterWithArg(t *testing.T) {
	out := renderS1(t, `{{ s | prepend: "Mr. " }}`, map[string]any{"s": "Smith"})
	require.Equal(t, "Mr. Smith", out)
}

func TestS11_Output_FilterOnNil_NoError(t *testing.T) {
	// applying a filter to nil should not panic; renders empty
	out := renderS1(t, "{{ v | upcase }}", map[string]any{"v": nil})
	require.Equal(t, "", out)
}

// ── multiple outputs ──────────────────────────────────────────────────────────

func TestS11_Output_Multiple_InTemplate(t *testing.T) {
	out := renderS1(t, "{{ a }} + {{ b }} = {{ c }}", map[string]any{"a": 1, "b": 2, "c": 3})
	require.Equal(t, "1 + 2 = 3", out)
}

func TestS11_Output_InterlevedTextAndTags(t *testing.T) {
	out := renderS1(t, "Hello, {{ name }}! You are {{ age }} years old.",
		map[string]any{"name": "Ana", "age": 28})
	require.Equal(t, "Hello, Ana! You are 28 years old.", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.1  Output / Expression — echo tag
// ═════════════════════════════════════════════════════════════════════════════

func TestS11_Echo_StringLiteral(t *testing.T) {
	require.Equal(t, "hello", renderS1T(t, `{% echo "hello" %}`))
}

func TestS11_Echo_IntLiteral(t *testing.T) {
	require.Equal(t, "42", renderS1T(t, `{% echo 42 %}`))
}

func TestS11_Echo_FloatLiteral(t *testing.T) {
	require.Equal(t, "3.14", renderS1T(t, `{% echo 3.14 %}`))
}

func TestS11_Echo_Variable(t *testing.T) {
	require.Equal(t, "world", renderS1(t, `{% echo v %}`, map[string]any{"v": "world"}))
}

func TestS11_Echo_NilVariable(t *testing.T) {
	require.Equal(t, "", renderS1(t, `{% echo v %}`, map[string]any{"v": nil}))
}

func TestS11_Echo_WithFilter(t *testing.T) {
	require.Equal(t, "HELLO", renderS1(t, `{% echo v | upcase %}`, map[string]any{"v": "hello"}))
}

func TestS11_Echo_WithFilterChain(t *testing.T) {
	require.Equal(t, "WORLD", renderS1(t, `{% echo v | downcase | upcase %}`, map[string]any{"v": "World"}))
}

func TestS11_Echo_InsideLiquidTag(t *testing.T) {
	// echo is specifically designed to work inside {% liquid %}
	src := "{% liquid\necho greeting\necho name\n%}"
	out := renderS1(t, src, map[string]any{"greeting": "Hi", "name": "there"})
	require.Equal(t, "Hithere", out)
}

func TestS11_Echo_EqualToObjectSyntax(t *testing.T) {
	// {% echo expr %} should produce the same output as {{ expr }}
	binds := map[string]any{"n": 7}
	require.Equal(t,
		renderS1(t, `{{ n }}`, binds),
		renderS1(t, `{% echo n %}`, binds))
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.2  Variables — assign
// ═════════════════════════════════════════════════════════════════════════════

func TestS12_Assign_String(t *testing.T) {
	require.Equal(t, "hello", renderS1T(t, `{% assign x = "hello" %}{{ x }}`))
}

func TestS12_Assign_Integer(t *testing.T) {
	require.Equal(t, "42", renderS1T(t, `{% assign n = 42 %}{{ n }}`))
}

func TestS12_Assign_Float(t *testing.T) {
	require.Equal(t, "3.14", renderS1T(t, `{% assign f = 3.14 %}{{ f }}`))
}

func TestS12_Assign_BoolTrue(t *testing.T) {
	require.Equal(t, "true", renderS1T(t, `{% assign b = true %}{{ b }}`))
}

func TestS12_Assign_BoolFalse(t *testing.T) {
	require.Equal(t, "false", renderS1T(t, `{% assign b = false %}{{ b }}`))
}

func TestS12_Assign_OverwritesExistingBinding(t *testing.T) {
	// assign overrides whatever was in the binding
	out := renderS1(t, `{% assign x = "new" %}{{ x }}`, map[string]any{"x": "old"})
	require.Equal(t, "new", out)
}

func TestS12_Assign_FromFilter(t *testing.T) {
	out := renderS1(t, `{% assign up = name | upcase %}{{ up }}`, map[string]any{"name": "alice"})
	require.Equal(t, "ALICE", out)
}

func TestS12_Assign_FromFilterChain(t *testing.T) {
	out := renderS1(t, `{% assign parts = s | downcase | split: " " %}{{ parts[0] }}-{{ parts[1] }}`,
		map[string]any{"s": "HELLO WORLD"})
	require.Equal(t, "hello-world", out)
}

func TestS12_Assign_Chained(t *testing.T) {
	// assigning x from a variable y that was also assigned in this template
	out := renderS1T(t, `{% assign a = "x" %}{% assign b = a %}{% assign a = "y" %}{{ a }}-{{ b }}`)
	// b captured the value of a at assignment time, not a live reference
	require.Equal(t, "y-x", out)
}

func TestS12_Assign_FromExistingBinding(t *testing.T) {
	out := renderS1(t, `{% assign y = x %}{{ y }}`, map[string]any{"x": "value"})
	require.Equal(t, "value", out)
}

func TestS12_Assign_Nil(t *testing.T) {
	out := renderS1(t, `{% assign v = nil_var %}[{{ v }}]`, map[string]any{"nil_var": nil})
	require.Equal(t, "[]", out)
}

func TestS12_Assign_UsableInConditional(t *testing.T) {
	out := renderS1T(t, `{% assign flag = true %}{% if flag %}yes{% endif %}`)
	require.Equal(t, "yes", out)
}

func TestS12_Assign_UsableInLoop(t *testing.T) {
	// assign a string, split it, iterate the parts
	out := renderS1T(t, `{% assign words = "a,b,c" | split: "," %}{% for w in words %}{{ w }}{% endfor %}`)
	require.Equal(t, "abc", out)
}

func TestS12_Assign_DoesNotLeakAcrossRenders(t *testing.T) {
	// assign in one render should not affect a separate render
	eng := liquid.NewEngine()
	out1, err := eng.ParseAndRenderString(`{% assign secret = "ok" %}{{ secret }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "ok", out1)
	out2, err := eng.ParseAndRenderString(`{{ secret }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "", out2) // no bleed-over
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.2  Variables — capture
// ═════════════════════════════════════════════════════════════════════════════

func TestS12_Capture_Basic(t *testing.T) {
	out := renderS1T(t, `{% capture msg %}hello world{% endcapture %}{{ msg }}`)
	require.Equal(t, "hello world", out)
}

func TestS12_Capture_EmptyBlock(t *testing.T) {
	out := renderS1T(t, `{% capture v %}{% endcapture %}[{{ v }}]`)
	require.Equal(t, "[]", out)
}

func TestS12_Capture_PreservesWhitespace(t *testing.T) {
	out := renderS1T(t, "{% capture v %}  spaces  {% endcapture %}[{{ v }}]")
	require.Equal(t, "[  spaces  ]", out)
}

func TestS12_Capture_MultilineContent(t *testing.T) {
	src := "{% capture block %}\nline1\nline2\n{% endcapture %}[{{ block }}]"
	out := renderS1T(t, src)
	require.Equal(t, "[\nline1\nline2\n]", out)
}

func TestS12_Capture_WithExpressions(t *testing.T) {
	out := renderS1(t, `{% capture greeting %}Hello, {{ name }}!{% endcapture %}{{ greeting }}`,
		map[string]any{"name": "Alice"})
	require.Equal(t, "Hello, Alice!", out)
}

func TestS12_Capture_WithFilters(t *testing.T) {
	out := renderS1(t, `{% capture loud %}{{ name | upcase }}{% endcapture %}{{ loud }}`,
		map[string]any{"name": "alice"})
	require.Equal(t, "ALICE", out)
}

func TestS12_Capture_OverwritesPriorValue(t *testing.T) {
	out := renderS1T(t, `{% capture v %}first{% endcapture %}{% capture v %}second{% endcapture %}{{ v }}`)
	require.Equal(t, "second", out)
}

func TestS12_Capture_UsedInConditional(t *testing.T) {
	src := `{% capture flag %}yes{% endcapture %}{% if flag == "yes" %}match{% endif %}`
	require.Equal(t, "match", renderS1T(t, src))
}

func TestS12_Capture_WithLoop(t *testing.T) {
	src := `{% capture all %}{% for i in arr %}{{ i }}{% endfor %}{% endcapture %}[{{ all }}]`
	out := renderS1(t, src, map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "[123]", out)
}

func TestS12_Capture_QuotedVarName_SingleQuote(t *testing.T) {
	// Bug fix: {% capture 'var' %} should strip quotes from the variable name
	out := renderS1T(t, `{% capture 'msg' %}quoted{% endcapture %}{{ msg }}`)
	require.Equal(t, "quoted", out)
}

func TestS12_Capture_QuotedVarName_DoubleQuote(t *testing.T) {
	out := renderS1T(t, `{% capture "msg" %}double{% endcapture %}{{ msg }}`)
	require.Equal(t, "double", out)
}

func TestS12_Capture_QuotedVarName_AccessibleLikePlain(t *testing.T) {
	// Quoted and unquoted captures should produce identical results
	plain := renderS1T(t, `{% capture x %}value{% endcapture %}{{ x }}`)
	quoted := renderS1T(t, `{% capture 'x' %}value{% endcapture %}{{ x }}`)
	require.Equal(t, plain, quoted)
}

func TestS12_Capture_DoesNotLeakAcrossRenders(t *testing.T) {
	eng := liquid.NewEngine()
	_, err := eng.ParseAndRenderString(`{% capture x %}captured{% endcapture %}`, nil)
	require.NoError(t, err)
	out, err := eng.ParseAndRenderString(`{{ x }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.3  Conditionals — if / elsif / else
// ═════════════════════════════════════════════════════════════════════════════

func TestS13_If_TrueCondition(t *testing.T) {
	out := renderS1(t, `{% if v %}yes{% endif %}`, map[string]any{"v": true})
	require.Equal(t, "yes", out)
}

func TestS13_If_FalseCondition_RendersNothing(t *testing.T) {
	out := renderS1(t, `{% if v %}yes{% endif %}`, map[string]any{"v": false})
	require.Equal(t, "", out)
}

func TestS13_If_NilCondition_RendersElse(t *testing.T) {
	out := renderS1(t, `{% if v %}yes{% else %}no{% endif %}`, map[string]any{"v": nil})
	require.Equal(t, "no", out)
}

func TestS13_If_Else_TrueTakesIf(t *testing.T) {
	out := renderS1(t, `{% if v %}yes{% else %}no{% endif %}`, map[string]any{"v": true})
	require.Equal(t, "yes", out)
}

func TestS13_If_Else_FalseTakesElse(t *testing.T) {
	out := renderS1(t, `{% if v %}yes{% else %}no{% endif %}`, map[string]any{"v": false})
	require.Equal(t, "no", out)
}

func TestS13_If_Elsif_AllBranches(t *testing.T) {
	tpl := `{% if n == 1 %}one{% elsif n == 2 %}two{% elsif n == 3 %}three{% else %}other{% endif %}`
	for _, tc := range []struct {
		n    int
		want string
	}{
		{1, "one"}, {2, "two"}, {3, "three"}, {4, "other"},
	} {
		t.Run(fmt.Sprintf("n=%d", tc.n), func(t *testing.T) {
			require.Equal(t, tc.want, renderS1(t, tpl, map[string]any{"n": tc.n}))
		})
	}
}

func TestS13_If_ManyElsif(t *testing.T) {
	// Ensures all elsif branches are checked in order
	tpl := `{% if n == 1 %}a{% elsif n == 2 %}b{% elsif n == 3 %}c{% elsif n == 4 %}d{% elsif n == 5 %}e{% else %}f{% endif %}`
	for n := 1; n <= 6; n++ {
		n := n
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			want := string(rune('a' + n - 1))
			require.Equal(t, want, renderS1(t, tpl, map[string]any{"n": n}))
		})
	}
}

func TestS13_If_And_BothTrue(t *testing.T) {
	out := renderS1(t, `{% if a and b %}yes{% else %}no{% endif %}`, map[string]any{"a": true, "b": true})
	require.Equal(t, "yes", out)
}

func TestS13_If_And_OneFalse(t *testing.T) {
	out := renderS1(t, `{% if a and b %}yes{% else %}no{% endif %}`, map[string]any{"a": true, "b": false})
	require.Equal(t, "no", out)
}

func TestS13_If_Or_OneTrue(t *testing.T) {
	out := renderS1(t, `{% if a or b %}yes{% else %}no{% endif %}`, map[string]any{"a": false, "b": true})
	require.Equal(t, "yes", out)
}

func TestS13_If_Or_BothFalse(t *testing.T) {
	out := renderS1(t, `{% if a or b %}yes{% else %}no{% endif %}`, map[string]any{"a": false, "b": false})
	require.Equal(t, "no", out)
}

func TestS13_If_ComparisonOperators(t *testing.T) {
	cases := []struct {
		tpl  string
		want string
	}{
		{`{% if 5 == 5 %}ok{% endif %}`, "ok"},
		{`{% if 5 == 4 %}ok{% else %}no{% endif %}`, "no"},
		{`{% if 5 != 4 %}ok{% endif %}`, "ok"},
		{`{% if 5 != 5 %}ok{% else %}no{% endif %}`, "no"},
		{`{% if 5 > 4 %}ok{% endif %}`, "ok"},
		{`{% if 4 > 5 %}ok{% else %}no{% endif %}`, "no"},
		{`{% if 4 < 5 %}ok{% endif %}`, "ok"},
		{`{% if 5 < 4 %}ok{% else %}no{% endif %}`, "no"},
		{`{% if 5 >= 5 %}ok{% endif %}`, "ok"},
		{`{% if 5 >= 6 %}ok{% else %}no{% endif %}`, "no"},
		{`{% if 4 <= 4 %}ok{% endif %}`, "ok"},
		{`{% if 5 <= 4 %}ok{% else %}no{% endif %}`, "no"},
	}
	for _, tc := range cases {
		t.Run(tc.tpl, func(t *testing.T) {
			require.Equal(t, tc.want, renderS1T(t, tc.tpl))
		})
	}
}

func TestS13_If_Contains_String(t *testing.T) {
	out := renderS1T(t, `{% if "foobar" contains "oba" %}yes{% else %}no{% endif %}`)
	require.Equal(t, "yes", out)
}

func TestS13_If_Contains_String_NoMatch(t *testing.T) {
	out := renderS1T(t, `{% if "foobar" contains "xyz" %}yes{% else %}no{% endif %}`)
	require.Equal(t, "no", out)
}

func TestS13_If_Contains_Array(t *testing.T) {
	out := renderS1(t, `{% if arr contains "b" %}yes{% else %}no{% endif %}`,
		map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "yes", out)
}

func TestS13_If_Contains_Array_NoMatch(t *testing.T) {
	out := renderS1(t, `{% if arr contains "z" %}yes{% else %}no{% endif %}`,
		map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "no", out)
}

func TestS13_If_Nested(t *testing.T) {
	tpl := `{% if a %}{% if b %}both{% else %}only_a{% endif %}{% else %}none{% endif %}`
	require.Equal(t, "both", renderS1(t, tpl, map[string]any{"a": true, "b": true}))
	require.Equal(t, "only_a", renderS1(t, tpl, map[string]any{"a": true, "b": false}))
	require.Equal(t, "none", renderS1(t, tpl, map[string]any{"a": false, "b": true}))
}

func TestS13_If_NestedThreeLevels(t *testing.T) {
	tpl := `{% if a %}{% if b %}{% if c %}abc{% else %}ab{% endif %}{% else %}a{% endif %}{% else %}none{% endif %}`
	require.Equal(t, "abc", renderS1(t, tpl, map[string]any{"a": true, "b": true, "c": true}))
	require.Equal(t, "ab", renderS1(t, tpl, map[string]any{"a": true, "b": true, "c": false}))
	require.Equal(t, "a", renderS1(t, tpl, map[string]any{"a": true, "b": false, "c": false}))
	require.Equal(t, "none", renderS1(t, tpl, map[string]any{"a": false, "b": true, "c": true}))
}

func TestS13_If_WithGoTypedInt(t *testing.T) {
	// all int-like types should compare correctly against integer literals
	for _, v := range []any{int8(5), int16(5), int32(5), int64(5), uint(5), uint32(5), uint64(5)} {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			out := renderS1(t, `{% if n == 5 %}yes{% else %}no{% endif %}`, map[string]any{"n": v})
			require.Equal(t, "yes", out)
		})
	}
}

func TestS13_If_WithGoTypedFloat(t *testing.T) {
	for _, v := range []any{float32(5.0), float64(5.0)} {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			out := renderS1(t, `{% if n == 5 %}yes{% else %}no{% endif %}`, map[string]any{"n": v})
			require.Equal(t, "yes", out)
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.3  Conditionals — unless
// ═════════════════════════════════════════════════════════════════════════════

func TestS13_Unless_RendersWhenFalse(t *testing.T) {
	out := renderS1(t, `{% unless v %}rendered{% endunless %}`, map[string]any{"v": false})
	require.Equal(t, "rendered", out)
}

func TestS13_Unless_SkipsWhenTrue(t *testing.T) {
	out := renderS1(t, `{% unless v %}rendered{% endunless %}`, map[string]any{"v": true})
	require.Equal(t, "", out)
}

func TestS13_Unless_RendersWhenNil(t *testing.T) {
	out := renderS1(t, `{% unless v %}rendered{% endunless %}`, map[string]any{"v": nil})
	require.Equal(t, "rendered", out)
}

func TestS13_Unless_WithElse_FalseTakesBody(t *testing.T) {
	src := `{% unless v %}body{% else %}elsebranch{% endunless %}`
	require.Equal(t, "body", renderS1(t, src, map[string]any{"v": false}))
}

func TestS13_Unless_WithElse_TrueTakesElse(t *testing.T) {
	src := `{% unless v %}body{% else %}elsebranch{% endunless %}`
	require.Equal(t, "elsebranch", renderS1(t, src, map[string]any{"v": true}))
}

func TestS13_Unless_ComplexCondition(t *testing.T) {
	// unless a == b evaluates as: not (a == b)
	out := renderS1(t, `{% unless a == b %}different{% else %}same{% endunless %}`,
		map[string]any{"a": 1, "b": 2})
	require.Equal(t, "different", out)
}

func TestS13_Unless_ComplexCondition_Equal(t *testing.T) {
	out := renderS1(t, `{% unless a == b %}different{% else %}same{% endunless %}`,
		map[string]any{"a": 5, "b": 5})
	require.Equal(t, "same", out)
}

func TestS13_Unless_Nested(t *testing.T) {
	src := `{% unless skip %}{% unless also_skip %}shown{% endunless %}{% endunless %}`
	require.Equal(t, "shown", renderS1(t, src, map[string]any{"skip": false, "also_skip": false}))
	require.Equal(t, "", renderS1(t, src, map[string]any{"skip": true, "also_skip": false}))
	require.Equal(t, "", renderS1(t, src, map[string]any{"skip": false, "also_skip": true}))
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.3  Conditionals — case / when
// ═════════════════════════════════════════════════════════════════════════════

func TestS13_Case_BasicStringMatch(t *testing.T) {
	out := renderS1(t, `{% case x %}{% when "hello" %}hi{% endcase %}`, map[string]any{"x": "hello"})
	require.Equal(t, "hi", out)
}

func TestS13_Case_NoMatchRendersEmpty(t *testing.T) {
	out := renderS1(t, `{% case x %}{% when "hello" %}hi{% endcase %}`, map[string]any{"x": "bye"})
	require.Equal(t, "", out)
}

func TestS13_Case_BasicIntMatch(t *testing.T) {
	out := renderS1(t, `{% case n %}{% when 1 %}one{% when 2 %}two{% endcase %}`, map[string]any{"n": 2})
	require.Equal(t, "two", out)
}

func TestS13_Case_ElseBranch(t *testing.T) {
	out := renderS1(t, `{% case n %}{% when 1 %}one{% else %}other{% endcase %}`, map[string]any{"n": 99})
	require.Equal(t, "other", out)
}

func TestS13_Case_OrSyntax(t *testing.T) {
	// when "a" or "b" should match either value
	tpl := `{% case x %}{% when "a" or "b" %}match{% else %}nope{% endcase %}`
	require.Equal(t, "match", renderS1(t, tpl, map[string]any{"x": "a"}))
	require.Equal(t, "match", renderS1(t, tpl, map[string]any{"x": "b"}))
	require.Equal(t, "nope", renderS1(t, tpl, map[string]any{"x": "c"}))
}

func TestS13_Case_MultipleWhenClauses(t *testing.T) {
	tpl := `{% case x %}{% when "a" %}A{% when "b" %}B{% when "c" %}C{% else %}?{% endcase %}`
	for _, tc := range []struct{ x, want string }{
		{"a", "A"}, {"b", "B"}, {"c", "C"}, {"d", "?"},
	} {
		t.Run(tc.x, func(t *testing.T) {
			require.Equal(t, tc.want, renderS1(t, tpl, map[string]any{"x": tc.x}))
		})
	}
}

func TestS13_Case_NilInputFallsToElse(t *testing.T) {
	out := renderS1(t, `{% case x %}{% when "something" %}hit{% else %}miss{% endcase %}`,
		map[string]any{"x": nil})
	require.Equal(t, "miss", out)
}

func TestS13_Case_BooleanMatch(t *testing.T) {
	tpl := `{% case b %}{% when true %}yes{% when false %}no{% endcase %}`
	require.Equal(t, "yes", renderS1(t, tpl, map[string]any{"b": true}))
	require.Equal(t, "no", renderS1(t, tpl, map[string]any{"b": false}))
}

func TestS13_Case_WithGoTypedInt(t *testing.T) {
	// Go int types should match integer literals
	tpl := `{% case n %}{% when 7 %}seven{% else %}other{% endcase %}`
	for _, v := range []any{int(7), int32(7), int64(7), uint(7), uint64(7)} {
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			require.Equal(t, "seven", renderS1(t, tpl, map[string]any{"n": v}))
		})
	}
}

func TestS13_Case_VariableSubjectAndWhen(t *testing.T) {
	// Both subject and when-value can be variables
	out := renderS1(t, `{% case x %}{% when a %}match{% else %}no{% endcase %}`,
		map[string]any{"x": "hello", "a": "hello"})
	require.Equal(t, "match", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.4  Iteration — for / else / endfor (basic)
// ═════════════════════════════════════════════════════════════════════════════

func TestS14_For_BasicStringArray(t *testing.T) {
	out := renderS1(t, `{% for s in words %}[{{ s }}]{% endfor %}`,
		map[string]any{"words": []string{"a", "b", "c"}})
	require.Equal(t, "[a][b][c]", out)
}

func TestS14_For_BasicIntArray(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{{ i }} {% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "1 2 3 ", out)
}

func TestS14_For_IntRange(t *testing.T) {
	out := renderS1T(t, `{% for i in (1..5) %}{{ i }}{% endfor %}`)
	require.Equal(t, "12345", out)
}

func TestS14_For_RangeViaVariables(t *testing.T) {
	out := renderS1(t, `{% for i in (start..stop) %}{{ i }} {% endfor %}`,
		map[string]any{"start": 3, "stop": 6})
	require.Equal(t, "3 4 5 6 ", out)
}

func TestS14_For_Else_EmptyArrayRendersElse(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{{ i }}{% else %}empty{% endfor %}`,
		map[string]any{"arr": []int{}})
	require.Equal(t, "empty", out)
}

func TestS14_For_Else_NilCollectionRendersElse(t *testing.T) {
	// Bug fix: nil collection should render else branch, not just empty string
	out := renderS1(t, `{% for i in arr %}{{ i }}{% else %}nil_else{% endfor %}`,
		map[string]any{"arr": nil})
	require.Equal(t, "nil_else", out)
}

func TestS14_For_Else_NonEmptySkipsElse(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{{ i }}{% else %}empty{% endfor %}`,
		map[string]any{"arr": []int{1, 2}})
	require.Equal(t, "12", out)
}

func TestS14_For_OverMap(t *testing.T) {
	// Iterating a map with a single known key
	out := renderS1(t, `{% for pair in m %}{{ pair[0] }}={{ pair[1] }}{% endfor %}`,
		map[string]any{"m": map[string]any{"k": "v"}})
	require.Equal(t, "k=v", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.4  Iteration — for modifiers
// ═════════════════════════════════════════════════════════════════════════════

func TestS14_For_Limit(t *testing.T) {
	out := renderS1(t, `{% for i in arr limit:2 %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{10, 20, 30, 40}})
	require.Equal(t, "1020", out)
}

func TestS14_For_Limit_Zero_RendersElse(t *testing.T) {
	out := renderS1(t, `{% for i in arr limit:0 %}{{ i }}{% else %}none{% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "none", out)
}

func TestS14_For_Offset(t *testing.T) {
	out := renderS1(t, `{% for i in arr offset:2 %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{10, 20, 30, 40}})
	require.Equal(t, "3040", out)
}

func TestS14_For_Offset_PastEnd_RendersElse(t *testing.T) {
	out := renderS1(t, `{% for i in arr offset:10 %}{{ i }}{% else %}none{% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "none", out)
}

func TestS14_For_Reversed(t *testing.T) {
	out := renderS1(t, `{% for i in arr reversed %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "321", out)
}

func TestS14_For_Reversed_SingleElement(t *testing.T) {
	out := renderS1(t, `{% for i in arr reversed %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{42}})
	require.Equal(t, "42", out)
}

func TestS14_For_Limit_And_Offset(t *testing.T) {
	// offset:1 limit:2 → skip 1 → take 2 → [20, 30]
	out := renderS1(t, `{% for i in arr limit:2 offset:1 %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{10, 20, 30, 40}})
	require.Equal(t, "2030", out)
}

func TestS14_For_AllModifiers_OffsetLimitReversed(t *testing.T) {
	// Ruby order: ALWAYS offset → limit → reversed, regardless of syntax order.
	// arr=[1,2,3,4,5]: offset:1=[2,3,4,5]; limit:3=[2,3,4]; reversed=[4,3,2]
	arr := map[string]any{"arr": []int{1, 2, 3, 4, 5}}
	want := "432"
	cases := []string{
		`{% for i in arr offset:1 limit:3 reversed %}{{ i }}{% endfor %}`,
		`{% for i in arr reversed offset:1 limit:3 %}{{ i }}{% endfor %}`,
		`{% for i in arr limit:3 reversed offset:1 %}{{ i }}{% endfor %}`,
		`{% for i in arr reversed limit:3 offset:1 %}{{ i }}{% endfor %}`,
	}
	for _, tpl := range cases {
		t.Run(tpl, func(t *testing.T) {
			require.Equal(t, want, renderS1(t, tpl, arr))
		})
	}
}

func TestS14_For_Modifier_ReversedLimitOne(t *testing.T) {
	// arr=[first,second,third]; offset:0; limit:1=[first]; reversed=[first]
	out := renderS1(t, `{% for a in array reversed limit:1 %}{{ a }}{% endfor %}`,
		map[string]any{"array": []string{"first", "second", "third"}})
	require.Equal(t, "first", out)
}

func TestS14_For_Modifier_ReversedOffsetOne(t *testing.T) {
	// arr=[first,second,third]; offset:1=[second,third]; reversed=[third,second]
	out := renderS1(t, `{% for a in array reversed offset:1 %}{{ a }}.{% endfor %}`,
		map[string]any{"array": []string{"first", "second", "third"}})
	require.Equal(t, "third.second.", out)
}

func TestS14_For_LimitFromVariable(t *testing.T) {
	out := renderS1(t, `{% for i in arr limit:n %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3, 4}, "n": 2})
	require.Equal(t, "12", out)
}

func TestS14_For_OffsetFromVariable(t *testing.T) {
	out := renderS1(t, `{% for i in arr offset:n %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3, 4}, "n": 2})
	require.Equal(t, "34", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.4  Iteration — forloop variables
// ═════════════════════════════════════════════════════════════════════════════

func TestS14_Forloop_Index(t *testing.T) {
	// forloop.index is 1-based
	out := renderS1(t, `{% for i in arr %}{{ forloop.index }}{% endfor %}`,
		map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "123", out)
}

func TestS14_Forloop_Index0(t *testing.T) {
	// forloop.index0 is 0-based
	out := renderS1(t, `{% for i in arr %}{{ forloop.index0 }}{% endfor %}`,
		map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "012", out)
}

func TestS14_Forloop_First(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{% if forloop.first %}F{% endif %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "Fabc", out)
}

func TestS14_Forloop_Last(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{{ i }}{% if forloop.last %}L{% endif %}{% endfor %}`,
		map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "abcL", out)
}

func TestS14_Forloop_SingleElement_FirstAndLast(t *testing.T) {
	// With a single element, both first and last should be true
	out := renderS1(t, `{% for i in arr %}{% if forloop.first %}F{% endif %}{% if forloop.last %}L{% endif %}{% endfor %}`,
		map[string]any{"arr": []string{"only"}})
	require.Equal(t, "FL", out)
}

func TestS14_Forloop_Length(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{{ forloop.length }} {% endfor %}`,
		map[string]any{"arr": []int{10, 20, 30}})
	// length stays constant throughout all iterations
	require.Equal(t, "3 3 3 ", out)
}

func TestS14_Forloop_Rindex(t *testing.T) {
	// rindex: items remaining including current (last item = 1)
	out := renderS1(t, `{% for i in arr %}{{ forloop.rindex }}{% endfor %}`,
		map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "321", out)
}

func TestS14_Forloop_Rindex0(t *testing.T) {
	// rindex0: items remaining after current (last item = 0)
	out := renderS1(t, `{% for i in arr %}{{ forloop.rindex0 }}{% endfor %}`,
		map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "210", out)
}

func TestS14_Forloop_Name(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{{ forloop.name }}{% endfor %}`,
		map[string]any{"arr": []string{"x"}})
	// forloop.name is "variable-collection" format
	require.Equal(t, "i-arr", out)
}

func TestS14_Forloop_Nested_IndependentVars(t *testing.T) {
	// Each nested for-loop has its own forloop variables that reset
	src := `{% for i in outer %}{% for j in inner %}{{ forloop.index }}{% endfor %}|{% endfor %}`
	out := renderS1(t, src, map[string]any{
		"outer": []string{"a", "b"},
		"inner": []string{"x", "y", "z"},
	})
	require.Equal(t, "123|123|", out)
}

func TestS14_Forloop_Nested_Length(t *testing.T) {
	// Inner length reflects inner array, outer length reflects outer array
	src := `{% for i in outer %}O{{ forloop.length }}{% for j in inner %}I{{ forloop.length }}{% endfor %}{% endfor %}`
	out := renderS1(t, src, map[string]any{
		"outer": []int{1, 2},
		"inner": []int{10, 20, 30},
	})
	require.Equal(t, "O2I3I3I3O2I3I3I3", out)
}

func TestS14_Forloop_ParentLoop(t *testing.T) {
	// forloop.parentloop gives access to the outer loop's forloop map
	src := `{% for i in outer %}{% for j in inner %}{{ forloop.parentloop.index }}-{{ forloop.index }} {% endfor %}{% endfor %}`
	out := renderS1(t, src, map[string]any{
		"outer": []string{"a", "b"},
		"inner": []string{"x", "y"},
	})
	require.Equal(t, "1-1 1-2 2-1 2-2 ", out)
}

func TestS14_Forloop_AllFieldsPresent(t *testing.T) {
	// Verify all standard forloop fields are accessible without error
	src := `{% for i in arr %}{{ forloop.index }},{{ forloop.index0 }},{{ forloop.rindex }},{{ forloop.rindex0 }},{{ forloop.first }},{{ forloop.last }},{{ forloop.length }}{% endfor %}`
	out := renderS1(t, src, map[string]any{"arr": []int{1}})
	require.Equal(t, "1,0,1,0,true,true,1", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.4  Iteration — break / continue
// ═════════════════════════════════════════════════════════════════════════════

func TestS14_Break_StopsLoop(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{% if i == 3 %}{% break %}{% endif %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3, 4, 5}})
	require.Equal(t, "12", out)
}

func TestS14_Break_OnFirstIteration(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{% break %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "", out)
}

func TestS14_Break_OnLastIteration(t *testing.T) {
	// break at the last item — everything before it is still rendered
	out := renderS1(t, `{% for i in arr %}{% if forloop.last %}{% break %}{% endif %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "12", out)
}

func TestS14_Break_OnlyExitsInnerLoop(t *testing.T) {
	// break in inner loop should not affect the outer loop
	src := `{% for i in outer %}{{ i }}{% for j in inner %}{% if j == 2 %}{% break %}{% endif %}{{ j }}{% endfor %}{% endfor %}`
	out := renderS1(t, src, map[string]any{
		"outer": []int{1, 2},
		"inner": []int{1, 2, 3},
	})
	// i=1→"1", inner j=1→"1" j=2→break;  i=2→"2", inner j=1→"1" j=2→break → "1121"
	require.Equal(t, "1121", out)
}

func TestS14_Continue_SkipsCurrentIteration(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{% if i == 2 %}{% continue %}{% endif %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3, 4}})
	require.Equal(t, "134", out)
}

func TestS14_Continue_SkipsRestOfIterationBody(t *testing.T) {
	// everything after continue in the same iteration should be skipped
	out := renderS1(t, `{% for i in arr %}{% if i == 2 %}{% continue %}{% endif %}{{ i }}-{% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "1-3-", out)
}

func TestS14_Continue_AllSkipped(t *testing.T) {
	// if every iteration hits continue, result is empty
	out := renderS1(t, `{% for i in arr %}{% continue %}{{ i }}{% endfor %}`,
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "", out)
}

func TestS14_Continue_OnlyAffectsInnerLoop(t *testing.T) {
	// continue in inner loop should not affect the outer loop
	src := `{% for i in outer %}|{% for j in inner %}{% if j == 2 %}{% continue %}{% endif %}{{ j }}{% endfor %}{% endfor %}`
	out := renderS1(t, src, map[string]any{
		"outer": []int{1, 2},
		"inner": []int{1, 2, 3},
	})
	require.Equal(t, "|13|13", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.4  Iteration — offset:continue
// ═════════════════════════════════════════════════════════════════════════════

func TestS14_OffsetContinue_Basic(t *testing.T) {
	// First loop takes items 0-1; second continues from item 2
	arr := map[string]any{"arr": []int{1, 2, 3, 4, 5, 6}}
	src := `{% for i in arr limit:2 %}{{ i }}{% endfor %};{% for i in arr limit:2 offset:continue %}{{ i }}{% endfor %}`
	out := renderS1(t, src, arr)
	require.Equal(t, "12;34", out)
}

func TestS14_OffsetContinue_ThreeChunks(t *testing.T) {
	arr := map[string]any{"arr": []int{1, 2, 3, 4, 5, 6}}
	src := `{% for i in arr limit:2 %}{{ i }}{% endfor %};` +
		`{% for i in arr limit:2 offset:continue %}{{ i }}{% endfor %};` +
		`{% for i in arr limit:2 offset:continue %}{{ i }}{% endfor %}`
	out := renderS1(t, src, arr)
	require.Equal(t, "12;34;56", out)
}

func TestS14_OffsetContinue_ExhaustedCollectionRendersEmpty(t *testing.T) {
	// When offset:continue resumes past the end of the collection, the loop
	// body and else branch are both skipped — the tag simply emits nothing.
	arr := map[string]any{"arr": []int{1, 2}}
	src := `{% for i in arr limit:10 %}{{ i }}{% endfor %};{% for i in arr offset:continue %}{{ i }}{% else %}done{% endfor %}`
	out := renderS1(t, src, arr)
	require.Equal(t, "12;", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.4  Iteration — cycle
// ═════════════════════════════════════════════════════════════════════════════

func TestS14_Cycle_TwoValues(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{% cycle "even", "odd" %}{% endfor %}`,
		map[string]any{"arr": make([]int, 4)})
	require.Equal(t, "evenoddevenodd", out)
}

func TestS14_Cycle_ThreeValues(t *testing.T) {
	out := renderS1(t, `{% for i in arr %}{% cycle "a", "b", "c" %}{% endfor %}`,
		map[string]any{"arr": make([]int, 5)})
	require.Equal(t, "abcab", out)
}

func TestS14_Cycle_WrapsAround(t *testing.T) {
	// 6 iterations with 3-value cycle → exactly 2 complete cycles
	out := renderS1(t, `{% for i in arr %}{% cycle "x", "y", "z" %}{% endfor %}`,
		map[string]any{"arr": make([]int, 6)})
	require.Equal(t, "xyzxyz", out)
}

func TestS14_Cycle_NamedGroups_Independent(t *testing.T) {
	// Two cycle groups with different names cycle independently
	src := `{% for i in arr %}{% cycle "g1": "a", "b" %}-{% cycle "g2": "x", "y", "z" %}|{% endfor %}`
	out := renderS1(t, src, map[string]any{"arr": make([]int, 3)})
	require.Equal(t, "a-x|b-y|a-z|", out)
}

func TestS14_Cycle_NamedGroups_SameValuesStillIndependent(t *testing.T) {
	// Even with same values, two groups cycle independently
	src := `{% for i in arr %}{% cycle "first": "1", "2" %} {% cycle "second": "1", "2" %}|{% endfor %}`
	out := renderS1(t, src, map[string]any{"arr": make([]int, 3)})
	require.Equal(t, "1 1|2 2|1 1|", out)
}

func TestS14_Cycle_ResetsOnNewRender(t *testing.T) {
	// Each new render starts the cycle from the beginning
	eng := liquid.NewEngine()
	renderCycle := func() string {
		out, err := eng.ParseAndRenderString(
			`{% for i in arr %}{% cycle "a","b","c" %}{% endfor %}`,
			map[string]any{"arr": make([]int, 3)})
		require.NoError(t, err)
		return out
	}
	require.Equal(t, "abc", renderCycle())
	require.Equal(t, "abc", renderCycle()) // must reset each time
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.4  Iteration — tablerow
// ═════════════════════════════════════════════════════════════════════════════

func TestS14_Tablerow_ProducesValidHTMLStructure(t *testing.T) {
	out := renderS1(t, `{% tablerow i in arr %}{{ i }}{% endtablerow %}`,
		map[string]any{"arr": []int{1, 2, 3}})
	require.Contains(t, out, `<tr class="row1">`)
	require.Contains(t, out, `<td class="col1">`)
	require.Contains(t, out, "</td>")
	require.Contains(t, out, "</tr>")
}

func TestS14_Tablerow_NoColsAllOnOneRow(t *testing.T) {
	// Without cols, all items go in a single row
	out := renderS1(t, `{% tablerow i in arr %}{{ i }}{% endtablerow %}`,
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, 1, strings.Count(out, "<tr"), "expected exactly 1 <tr>")
	require.Equal(t, 3, strings.Count(out, "<td"), "expected 3 <td> elements")
}

func TestS14_Tablerow_WithCols_TwoRows(t *testing.T) {
	out := renderS1(t, `{% tablerow i in arr cols:2 %}{{ i }}{% endtablerow %}`,
		map[string]any{"arr": []int{1, 2, 3, 4}})
	require.Equal(t, 2, strings.Count(out, "<tr"), "expected 2 rows")
	require.Equal(t, 4, strings.Count(out, "<td"), "expected 4 cells")
}

func TestS14_Tablerow_WithCols_OddCount(t *testing.T) {
	// 3 items with cols:2 → 2 rows (row1: 2 items, row2: 1 item)
	out := renderS1(t, `{% tablerow i in arr cols:2 %}{{ i }}{% endtablerow %}`,
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, 2, strings.Count(out, "<tr"), "expected 2 rows for 3-item/2-col tablerow")
	require.Equal(t, 3, strings.Count(out, "<td"), "expected 3 cells")
}

func TestS14_Tablerow_WithCols_RowClassNumbers(t *testing.T) {
	out := renderS1(t, `{% tablerow i in arr cols:2 %}{{ i }}{% endtablerow %}`,
		map[string]any{"arr": []int{1, 2, 3, 4}})
	require.Contains(t, out, `<tr class="row1">`)
	require.Contains(t, out, `<tr class="row2">`)
}

func TestS14_Tablerow_WithCols_ColClassNumbers(t *testing.T) {
	out := renderS1(t, `{% tablerow i in arr cols:2 %}{{ i }}{% endtablerow %}`,
		map[string]any{"arr": []int{1, 2, 3, 4}})
	require.Contains(t, out, `<td class="col1">`)
	require.Contains(t, out, `<td class="col2">`)
}

func TestS14_Tablerow_ForloopIndex(t *testing.T) {
	out := renderS1(t, `{% tablerow i in arr %}{{ forloop.index }} {% endtablerow %}`,
		map[string]any{"arr": []string{"a", "b", "c"}})
	require.Contains(t, out, "1 ")
	require.Contains(t, out, "2 ")
	require.Contains(t, out, "3 ")
}

func TestS14_Tablerow_ForloopFirst_Last(t *testing.T) {
	out := renderS1(t, `{% tablerow i in arr %}{% if forloop.first %}F{% endif %}{{ i }}{% if forloop.last %}L{% endif %}{% endtablerow %}`,
		map[string]any{"arr": []string{"x", "y", "z"}})
	require.Contains(t, out, "Fx")
	require.Contains(t, out, "zL")
}

func TestS14_Tablerow_ColVariables(t *testing.T) {
	// forloop.col is 1-based column index; col_first and col_last for boundary detection
	out := renderS1(t, `{% tablerow i in arr cols:2 %}{{ forloop.col }}{% endtablerow %}`,
		map[string]any{"arr": []int{1, 2, 3, 4}})
	// pattern: col1,col2,col1,col2 embedded in td content
	require.Equal(t, 2, strings.Count(out, ">1<"), "expected 2 col-1 cells")
	require.Equal(t, 2, strings.Count(out, ">2<"), "expected 2 col-2 cells")
}

func TestS14_Tablerow_WithLimit(t *testing.T) {
	out := renderS1(t, `{% tablerow i in arr limit:2 %}{{ i }}{% endtablerow %}`,
		map[string]any{"arr": []int{10, 20, 30, 40}})
	require.Contains(t, out, "10")
	require.Contains(t, out, "20")
	require.NotContains(t, out, "30")
	require.NotContains(t, out, "40")
}

func TestS14_Tablerow_WithOffset(t *testing.T) {
	out := renderS1(t, `{% tablerow i in arr offset:2 %}{{ i }}{% endtablerow %}`,
		map[string]any{"arr": []int{10, 20, 30, 40}})
	require.NotContains(t, out, "10")
	require.NotContains(t, out, "20")
	require.Contains(t, out, "30")
	require.Contains(t, out, "40")
}

func TestS14_Tablerow_Range(t *testing.T) {
	out := renderS1T(t, `{% tablerow i in (1..3) %}{{ i }}{% endtablerow %}`)
	require.Contains(t, out, "1")
	require.Contains(t, out, "2")
	require.Contains(t, out, "3")
	require.Equal(t, 3, strings.Count(out, "<td"), "expected 3 cells for range 1..3")
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.6  Structure — raw
// ═════════════════════════════════════════════════════════════════════════════

func TestS16_Raw_PreservesObjectSyntax(t *testing.T) {
	out := renderS1T(t, `{% raw %}{{ variable }}{% endraw %}`)
	require.Equal(t, "{{ variable }}", out)
}

func TestS16_Raw_PreservesTagSyntax(t *testing.T) {
	out := renderS1T(t, `{% raw %}{% if true %}yes{% endif %}{% endraw %}`)
	require.Equal(t, `{% if true %}yes{% endif %}`, out)
}

func TestS16_Raw_PreservesWhitespaceAndNewlines(t *testing.T) {
	out := renderS1T(t, "{% raw %}\n  hello\n  world\n{% endraw %}")
	require.Equal(t, "\n  hello\n  world\n", out)
}

func TestS16_Raw_EmptyBlock(t *testing.T) {
	out := renderS1T(t, `{% raw %}{% endraw %}`)
	require.Equal(t, "", out)
}

func TestS16_Raw_WithSurroundingText(t *testing.T) {
	out := renderS1T(t, `before {% raw %}{{ x }}{% endraw %} after`)
	require.Equal(t, "before {{ x }} after", out)
}

func TestS16_Raw_TripleBraces(t *testing.T) {
	out := renderS1T(t, `{% raw %}{{{ triple }}}{% endraw %}`)
	require.Equal(t, "{{{ triple }}}", out)
}

func TestS16_Raw_PercentBraceTag(t *testing.T) {
	out := renderS1T(t, `{% raw %}{% assign x = 1 %}{% endraw %}`)
	require.Equal(t, `{% assign x = 1 %}`, out)
}

func TestS16_Raw_DoesNotEvalBinding(t *testing.T) {
	// Even with a matching binding, raw should not interpolate
	out := renderS1(t, `{% raw %}{{ name }}{% endraw %}`, map[string]any{"name": "Alice"})
	require.Equal(t, "{{ name }}", out)
}

func TestS16_Raw_MultipleBlocks(t *testing.T) {
	// Two raw blocks in the same template
	out := renderS1T(t, `{% raw %}{{ a }}{% endraw %} and {% raw %}{{ b }}{% endraw %}`)
	require.Equal(t, "{{ a }} and {{ b }}", out)
}

func TestS16_Raw_AdjacentToLiquid(t *testing.T) {
	// Part rendered, part raw
	out := renderS1(t, `{{ name }} {% raw %}{{ name }}{% endraw %}`, map[string]any{"name": "Alice"})
	require.Equal(t, "Alice {{ name }}", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// 1.6  Structure — comment
// ═════════════════════════════════════════════════════════════════════════════

func TestS16_Comment_BasicSuppressed(t *testing.T) {
	out := renderS1T(t, `{% comment %}this is suppressed{% endcomment %}`)
	require.Equal(t, "", out)
}

func TestS16_Comment_DoesNotRenderExpression(t *testing.T) {
	out := renderS1(t, `{% comment %}{{ secret }}{% endcomment %}`,
		map[string]any{"secret": "password"})
	require.Equal(t, "", out)
}

func TestS16_Comment_DoesNotExecuteAssign(t *testing.T) {
	// assign inside comment should NOT execute
	src := `{% comment %}{% assign x = "set" %}{% endcomment %}[{{ x }}]`
	out := renderS1T(t, src)
	require.Equal(t, "[]", out)
}

func TestS16_Comment_WithSurroundingContent(t *testing.T) {
	out := renderS1T(t, `before{% comment %} hidden {% endcomment %}after`)
	require.Equal(t, "beforeafter", out)
}

func TestS16_Comment_Multiline(t *testing.T) {
	src := "start\n{% comment %}\nThis is\na multiline comment\n{% endcomment %}\nend"
	out := renderS1T(t, src)
	require.Equal(t, "start\n\nend", out)
}

func TestS16_Comment_WithForTagInside(t *testing.T) {
	// Tags inside comment are completely ignored — the for loop must NOT run
	src := `{% comment %}{% for i in arr %}{{ i }}{% endfor %}{% endcomment %}`
	out := renderS1(t, src, map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "", out)
}

func TestS16_Comment_FirstEndcommentClosesBlock(t *testing.T) {
	// Go's comment parser treats the FIRST {% endcomment %} as the closing
	// delimiter. Content between {% comment %} and the first {% endcomment %}
	// (including any inner {% comment %} tags) is fully discarded.
	// A second {% endcomment %} without a matching opener is a parse error.
	// This test validates the non-nested use (one comment, properly closed).
	src := `{% comment %}anything goes here, even {% if %} or tags{% endcomment %}after`
	require.Equal(t, "after", renderS1T(t, src))
}

func TestS16_Comment_MultipleCommentsInTemplate(t *testing.T) {
	// Multiple comment blocks in a single template
	src := `a{% comment %}X{% endcomment %}b{% comment %}Y{% endcomment %}c`
	require.Equal(t, "abc", renderS1T(t, src))
}

func TestS16_Comment_PreservesBindingsThatFollow(t *testing.T) {
	// Content after a comment is still rendered correctly
	out := renderS1(t, `{% comment %}secret{% endcomment %}{{ v }}`,
		map[string]any{"v": "visible"})
	require.Equal(t, "visible", out)
}
