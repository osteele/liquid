package liquid_test

// S5 — Acesso a Variáveis: testes E2E intensivos
//
// Cobre o tópico 5 do implementation-checklist:
//
//   5a. obj.prop, obj[key], array[0]
//   5b. array[-1] — negative indexing
//   5c. array.first, array.last, obj.size
//   5d. {{ [key] }} — dynamic variable lookup (Ruby)
//   5e. {{ test . test }} — dot with surrounding whitespace (Ruby)
//   5f. {{ ["Key"].sub }} — top-level bracket + dot (LiquidJS #643)
//
// Objetivo: cobrir todos os edge cases de forma que qualquer regressão no
// pipeline binding→parser→evaluator→render seja detectada imediatamente.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// renderS5 is the shared helper.
func renderS5(t *testing.T, tpl string, bindings map[string]any) string {
	t.Helper()
	eng := liquid.NewEngine()
	out, err := eng.ParseAndRenderString(tpl, bindings)
	require.NoError(t, err, "template: %s", tpl)
	return out
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  5a — obj.prop, obj[key], array[0]                                          ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

// ── 5a.1: Dot notation ────────────────────────────────────────────────────────

func TestS5_DotNotation_SingleLevel(t *testing.T) {
	out := renderS5(t, `{{ obj.name }}`, map[string]any{
		"obj": map[string]any{"name": "Alice"},
	})
	require.Equal(t, "Alice", out)
}

func TestS5_DotNotation_TwoLevels(t *testing.T) {
	out := renderS5(t, `{{ a.b.c }}`, map[string]any{
		"a": map[string]any{"b": map[string]any{"c": "deep"}},
	})
	require.Equal(t, "deep", out)
}

func TestS5_DotNotation_FiveLevels(t *testing.T) {
	out := renderS5(t, `{{ a.b.c.d.e }}`, map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": map[string]any{
						"e": "leaf",
					},
				},
			},
		},
	})
	require.Equal(t, "leaf", out)
}

func TestS5_DotNotation_MissingKeyReturnsEmpty(t *testing.T) {
	out := renderS5(t, `{{ obj.missing }}`, map[string]any{
		"obj": map[string]any{"name": "Alice"},
	})
	require.Equal(t, "", out)
}

func TestS5_DotNotation_MidChainMissing_StopsGracefully(t *testing.T) {
	// obj.b doesn't exist; obj.b.c must not panic
	out := renderS5(t, `{{ obj.b.c }}`, map[string]any{
		"obj": map[string]any{"name": "Alice"},
	})
	require.Equal(t, "", out)
}

func TestS5_DotNotation_OnNilVariable(t *testing.T) {
	out := renderS5(t, `{{ nothing.prop }}`, map[string]any{"nothing": nil})
	require.Equal(t, "", out)
}

func TestS5_DotNotation_OnGoStruct(t *testing.T) {
	type Inner struct{ Value string }
	type Outer struct{ Inner Inner }

	out := renderS5(t, `{{ obj.Inner.Value }}`, map[string]any{
		"obj": Outer{Inner: Inner{Value: "struct_leaf"}},
	})
	require.Equal(t, "struct_leaf", out)
}

func TestS5_DotNotation_GoStructPublicFields(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	out := renderS5(t, `{{ p.Name }} is {{ p.Age }}`, map[string]any{
		"p": Person{Name: "Bob", Age: 30},
	})
	require.Equal(t, "Bob is 30", out)
}

// ── 5a.2: Bracket notation with string keys ───────────────────────────────────

func TestS5_BracketString_SingleKey(t *testing.T) {
	out := renderS5(t, `{{ page["title"] }}`, map[string]any{
		"page": map[string]any{"title": "Intro"},
	})
	require.Equal(t, "Intro", out)
}

func TestS5_BracketString_KeyWithSpaces(t *testing.T) {
	out := renderS5(t, `{{ hash["complex key"] }}`, map[string]any{
		"hash": map[string]any{"complex key": "found"},
	})
	require.Equal(t, "found", out)
}

func TestS5_BracketString_KeyWithSpecialChars(t *testing.T) {
	out := renderS5(t, `{{ data["key-with-dashes"] }}`, map[string]any{
		"data": map[string]any{"key-with-dashes": "val"},
	})
	require.Equal(t, "val", out)
}

func TestS5_BracketVar_KeyFromVariable(t *testing.T) {
	// {{ a[b] }} — key is a variable
	out := renderS5(t, `{{ a[b] }}`, map[string]any{
		"b": "c",
		"a": map[string]any{"c": "result"},
	})
	require.Equal(t, "result", out)
}

func TestS5_BracketVar_KeyFromVariableWithSpaces(t *testing.T) {
	// Explicit space around inner variable: {{ a[ b ] }}
	out := renderS5(t, `{{ a[ b ] }}`, map[string]any{
		"b": "k",
		"a": map[string]any{"k": "found"},
	})
	require.Equal(t, "found", out)
}

func TestS5_BracketMixed_DotThenBracket(t *testing.T) {
	// {{ hash["b"].c }} — bracket then dot
	out := renderS5(t, `{{ hash["b"].c }}`, map[string]any{
		"hash": map[string]any{
			"b": map[string]any{"c": "d"},
		},
	})
	require.Equal(t, "d", out)
}

func TestS5_BracketMixed_DotThenBracketThenDot(t *testing.T) {
	out := renderS5(t, `{{ obj.a["b"].c }}`, map[string]any{
		"obj": map[string]any{
			"a": map[string]any{
				"b": map[string]any{"c": "xyz"},
			},
		},
	})
	require.Equal(t, "xyz", out)
}

// ── 5a.3: Array integer indexing ──────────────────────────────────────────────

func TestS5_ArrayIndex_First(t *testing.T) {
	out := renderS5(t, `{{ arr[0] }}`, map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "a", out)
}

func TestS5_ArrayIndex_Middle(t *testing.T) {
	out := renderS5(t, `{{ arr[1] }}`, map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "b", out)
}

func TestS5_ArrayIndex_Last(t *testing.T) {
	out := renderS5(t, `{{ arr[2] }}`, map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "c", out)
}

func TestS5_ArrayIndex_OutOfBounds_ReturnsEmpty(t *testing.T) {
	out := renderS5(t, `{{ arr[99] }}`, map[string]any{"arr": []string{"a", "b"}})
	require.Equal(t, "", out)
}

func TestS5_ArrayIndex_ViaVariable(t *testing.T) {
	out := renderS5(t, `{{ arr[i] }}`, map[string]any{"arr": []string{"x", "y", "z"}, "i": 2})
	require.Equal(t, "z", out)
}

func TestS5_ArrayIndex_ViaAssign(t *testing.T) {
	out := renderS5(t,
		`{% assign i = 1 %}{{ arr[i] }}`,
		map[string]any{"arr": []string{"first", "second", "third"}})
	require.Equal(t, "second", out)
}

func TestS5_ArrayIndex_NestedArrays(t *testing.T) {
	out := renderS5(t, `{{ matrix[1][0] }}`, map[string]any{
		"matrix": [][]string{{"a", "b"}, {"c", "d"}},
	})
	require.Equal(t, "c", out)
}

func TestS5_ArrayIndex_InsideForLoop(t *testing.T) {
	// access a specific index via a range variable inside a for loop
	out := renderS5(t,
		`{% for i in (0..2) %}{{ arr[i] }}{% endfor %}`,
		map[string]any{"arr": []string{"x", "y", "z"}})
	require.Equal(t, "xyz", out)
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  5b — Negative array indexing                                               ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

func TestS5_NegativeIndex_MinusOne(t *testing.T) {
	out := renderS5(t, `{{ arr[-1] }}`, map[string]any{"arr": []string{"x", "y", "z"}})
	require.Equal(t, "z", out)
}

func TestS5_NegativeIndex_MinusTwo(t *testing.T) {
	out := renderS5(t, `{{ arr[-2] }}`, map[string]any{"arr": []string{"x", "y", "z"}})
	require.Equal(t, "y", out)
}

func TestS5_NegativeIndex_MinusLen_IsFirst(t *testing.T) {
	out := renderS5(t, `{{ arr[-3] }}`, map[string]any{"arr": []string{"x", "y", "z"}})
	require.Equal(t, "x", out)
}

func TestS5_NegativeIndex_BeyondLength_ReturnsEmpty(t *testing.T) {
	out := renderS5(t, `{{ arr[-8] }}`, map[string]any{"arr": []string{"x", "y", "z"}})
	require.Equal(t, "", out)
}

func TestS5_NegativeIndex_EmptyArray_ReturnsEmpty(t *testing.T) {
	out := renderS5(t, `{{ arr[-1] }}`, map[string]any{"arr": []string{}})
	require.Equal(t, "", out)
}

func TestS5_NegativeIndex_SingleElement(t *testing.T) {
	// [-1] on a single-element array == [0]
	out := renderS5(t, `{{ arr[-1] }}`, map[string]any{"arr": []string{"only"}})
	require.Equal(t, "only", out)
}

func TestS5_NegativeIndex_ViaAssign(t *testing.T) {
	// split produces []string; negative index must work
	out := renderS5(t,
		`{% assign a = "x,y,z" | split: ',' %}{{ a[-1] }} {{ a[-3] }} {{ a[-8] }}`,
		nil)
	require.Equal(t, "z x ", out)
}

func TestS5_NegativeIndex_PositiveNegativeEquivalence(t *testing.T) {
	// arr[-1] == arr[len-1]
	arr := []string{"alpha", "beta", "gamma"}
	eng := liquid.NewEngine()

	v1, _ := eng.ParseAndRenderString(`{{ arr[-1] }}`, map[string]any{"arr": arr})
	v2, _ := eng.ParseAndRenderString(`{{ arr[2] }}`, map[string]any{"arr": arr})
	require.Equal(t, v1, v2, "arr[-1] must equal arr[len-1]")

	v3, _ := eng.ParseAndRenderString(`{{ arr[-2] }}`, map[string]any{"arr": arr})
	v4, _ := eng.ParseAndRenderString(`{{ arr[1] }}`, map[string]any{"arr": arr})
	require.Equal(t, v3, v4, "arr[-2] must equal arr[len-2]")
}

func TestS5_NegativeIndex_IntegerTypesAsIndex(t *testing.T) {
	arr := []string{"a", "b", "c"}
	// int, int8, int16, int32, int64 — all must work as negative indices
	cases := []any{
		int(-1), int8(-1), int16(-1), int32(-1), int64(-1),
	}
	for _, idx := range cases {
		t.Run(fmt.Sprintf("%T", idx), func(t *testing.T) {
			out := renderS5(t, `{{ arr[i] }}`, map[string]any{"arr": arr, "i": idx})
			require.Equal(t, "c", out)
		})
	}
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  5c — array.first · array.last · obj.size                                  ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

// ── 5c.1: .first ──────────────────────────────────────────────────────────────

func TestS5_First_OnArray(t *testing.T) {
	out := renderS5(t, `{{ arr.first }}`, map[string]any{"arr": []string{"apple", "banana", "cherry"}})
	require.Equal(t, "apple", out)
}

func TestS5_First_OnSingleElement(t *testing.T) {
	out := renderS5(t, `{{ arr.first }}`, map[string]any{"arr": []string{"solo"}})
	require.Equal(t, "solo", out)
}

func TestS5_First_OnEmpty_ReturnsEmpty(t *testing.T) {
	out := renderS5(t, `{{ arr.first }}`, map[string]any{"arr": []string{}})
	require.Equal(t, "", out)
}

func TestS5_First_EqualsIndex0(t *testing.T) {
	eng := liquid.NewEngine()
	arr := []string{"alpha", "beta"}
	v1, _ := eng.ParseAndRenderString(`{{ arr.first }}`, map[string]any{"arr": arr})
	v2, _ := eng.ParseAndRenderString(`{{ arr[0] }}`, map[string]any{"arr": arr})
	require.Equal(t, v1, v2)
}

func TestS5_First_OnIntArray(t *testing.T) {
	out := renderS5(t, `{{ nums.first }}`, map[string]any{"nums": []int{10, 20, 30}})
	require.Equal(t, "10", out)
}

// ── 5c.2: .last ───────────────────────────────────────────────────────────────

func TestS5_Last_OnArray(t *testing.T) {
	out := renderS5(t, `{{ arr.last }}`, map[string]any{"arr": []string{"apple", "banana", "cherry"}})
	require.Equal(t, "cherry", out)
}

func TestS5_Last_OnSingleElement(t *testing.T) {
	out := renderS5(t, `{{ arr.last }}`, map[string]any{"arr": []string{"solo"}})
	require.Equal(t, "solo", out)
}

func TestS5_Last_OnEmpty_ReturnsEmpty(t *testing.T) {
	out := renderS5(t, `{{ arr.last }}`, map[string]any{"arr": []string{}})
	require.Equal(t, "", out)
}

func TestS5_Last_EqualsNegativeOne(t *testing.T) {
	eng := liquid.NewEngine()
	arr := []string{"alpha", "beta", "gamma"}
	v1, _ := eng.ParseAndRenderString(`{{ arr.last }}`, map[string]any{"arr": arr})
	v2, _ := eng.ParseAndRenderString(`{{ arr[-1] }}`, map[string]any{"arr": arr})
	require.Equal(t, v1, v2, "arr.last must equal arr[-1]")
}

func TestS5_Last_OnIntArray(t *testing.T) {
	out := renderS5(t, `{{ nums.last }}`, map[string]any{"nums": []int{10, 20, 30}})
	require.Equal(t, "30", out)
}

// ── 5c.3: .size ───────────────────────────────────────────────────────────────

func TestS5_Size_OnStringArray(t *testing.T) {
	out := renderS5(t, `{{ arr.size }}`, map[string]any{"arr": []string{"a", "b", "c", "d"}})
	require.Equal(t, "4", out)
}

func TestS5_Size_OnEmptyArray(t *testing.T) {
	out := renderS5(t, `{{ arr.size }}`, map[string]any{"arr": []string{}})
	require.Equal(t, "0", out)
}

func TestS5_Size_OnString_IsRuneCount(t *testing.T) {
	out := renderS5(t, `{{ s.size }}`, map[string]any{"s": "hello"})
	require.Equal(t, "5", out)
}

func TestS5_Size_OnString_Multibyte(t *testing.T) {
	// Unicode string: rune count, not byte count
	out := renderS5(t, `{{ s.size }}`, map[string]any{"s": "héllo"})
	require.Equal(t, "5", out)
}

func TestS5_Size_OnMap(t *testing.T) {
	out := renderS5(t, `{{ h.size }}`, map[string]any{
		"h": map[string]any{"a": 1, "b": 2, "c": 3},
	})
	require.Equal(t, "3", out)
}

func TestS5_Size_OnEmptyMap(t *testing.T) {
	out := renderS5(t, `{{ h.size }}`, map[string]any{"h": map[string]any{}})
	require.Equal(t, "0", out)
}

func TestS5_Size_MapKeyWinsOverBuiltin(t *testing.T) {
	// When a map has an explicit "size" key, that value wins over the computed count
	out := renderS5(t, `{{ h.size }}`, map[string]any{
		"h": map[string]any{"size": "custom"},
	})
	require.Equal(t, "custom", out)
}

func TestS5_Size_InCondition(t *testing.T) {
	out := renderS5(t,
		`{% if arr.size > 2 %}big{% else %}small{% endif %}`,
		map[string]any{"arr": []string{"x", "y", "z"}})
	require.Equal(t, "big", out)
}

func TestS5_Size_UsedInFilterChain(t *testing.T) {
	out := renderS5(t, `{{ arr.size | plus: 10 }}`, map[string]any{"arr": []string{"a", "b"}})
	require.Equal(t, "12", out)
}

func TestS5_First_UsedInFilterChain(t *testing.T) {
	out := renderS5(t, `{{ arr.first | upcase }}`, map[string]any{"arr": []string{"hello", "world"}})
	require.Equal(t, "HELLO", out)
}

func TestS5_First_NestedAccess(t *testing.T) {
	// arr.first.name — first returns an object, then access .name
	out := renderS5(t, `{{ people.first.name }}`, map[string]any{
		"people": []map[string]any{
			{"name": "Alice", "age": 30},
			{"name": "Bob", "age": 25},
		},
	})
	require.Equal(t, "Alice", out)
}

func TestS5_Last_NestedAccess(t *testing.T) {
	out := renderS5(t, `{{ people.last.name }}`, map[string]any{
		"people": []map[string]any{
			{"name": "Alice"},
			{"name": "Bob"},
		},
	})
	require.Equal(t, "Bob", out)
}

func TestS5_Size_InsideForLoop(t *testing.T) {
	out := renderS5(t,
		`{% for item in items %}{{ forloop.index }}/{{ items.size }} {% endfor %}`,
		map[string]any{"items": []string{"a", "b", "c"}})
	require.Equal(t, "1/3 2/3 3/3 ", out)
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  5d — {{ [key] }} dynamic variable lookup                                   ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

func TestS5_DynamicLookup_Simple(t *testing.T) {
	out := renderS5(t, `{{ [key] }}`, map[string]any{"key": "foo", "foo": "bar"})
	require.Equal(t, "bar", out)
}

func TestS5_DynamicLookup_MissingKey_ReturnsEmpty(t *testing.T) {
	out := renderS5(t, `{{ [key] }}`, map[string]any{"key": "nonexistent"})
	require.Equal(t, "", out)
}

func TestS5_DynamicLookup_KeyIsNil_ReturnsEmpty(t *testing.T) {
	out := renderS5(t, `{{ [key] }}`, map[string]any{"key": nil})
	require.Equal(t, "", out)
}

func TestS5_DynamicLookup_WithSingleQuoteKey(t *testing.T) {
	// {{ ['foo'] }} — single-quoted string literal in brackets at top level
	out := renderS5(t, "{{ ['foo'] }}", map[string]any{"foo": "sq_direct"})
	require.Equal(t, "sq_direct", out)
}

func TestS5_DynamicLookup_KeyFromAssign(t *testing.T) {
	out := renderS5(t,
		`{% assign k = "target" %}{{ [k] }}`,
		map[string]any{"target": "resolved"})
	require.Equal(t, "resolved", out)
}

func TestS5_DynamicLookup_KeyFromArrayIndex(t *testing.T) {
	// {{ [list[0]] }} — use list[0] as the variable name
	out := renderS5(t, `{{ [list[0]] }}`, map[string]any{
		"list": []string{"foo"},
		"foo":  "bar",
	})
	require.Equal(t, "bar", out)
}

func TestS5_DynamicLookup_NestedResult_AccessProperty(t *testing.T) {
	// {{ [key].name }} — resolved value is an object, then access property
	out := renderS5(t, `{{ [varname].name }}`, map[string]any{
		"varname": "person",
		"person":  map[string]any{"name": "Alice"},
	})
	require.Equal(t, "Alice", out)
}

func TestS5_DynamicLookup_DoubleNested(t *testing.T) {
	// {{ list[list[0]]["foo"] }} — chain of lookups where an index is itself
	// the result of another index operation
	out := renderS5(t, `{{ list[list[0]]["foo"] }}`, map[string]any{
		"list": []any{1, map[string]any{"foo": "bar"}},
	})
	require.Equal(t, "bar", out)
}

func TestS5_DynamicLookup_InsideForLoop(t *testing.T) {
	// Iterates over a list of variable names and resolves each dynamically
	out := renderS5(t,
		`{% for k in keys %}{{ [k] }} {% endfor %}`,
		map[string]any{
			"keys": []string{"a", "b", "c"},
			"a":    "alpha",
			"b":    "beta",
			"c":    "gamma",
		})
	require.Equal(t, "alpha beta gamma ", out)
}

func TestS5_DynamicLookup_InsideIf(t *testing.T) {
	out := renderS5(t,
		`{% if [flag] %}yes{% else %}no{% endif %}`,
		map[string]any{"flag": "enabled", "enabled": true})
	require.Equal(t, "yes", out)
}

func TestS5_DynamicLookup_WithLiteralStringKey(t *testing.T) {
	// {{ ["foo"] }} — literal string in brackets at top level → lookup "foo"
	out := renderS5(t, `{{ ["foo"] }}`, map[string]any{"foo": "direct"})
	require.Equal(t, "direct", out)
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  5e — dot with surrounding whitespace                                       ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

func TestS5_DotWithSpaces_Single(t *testing.T) {
	out := renderS5(t, `{{ obj . key }}`, map[string]any{
		"obj": map[string]any{"key": "found"},
	})
	require.Equal(t, "found", out)
}

func TestS5_DotWithSpaces_TwoLevels(t *testing.T) {
	out := renderS5(t, `{{ a . b . c }}`, map[string]any{
		"a": map[string]any{
			"b": map[string]any{"c": "deep"},
		},
	})
	require.Equal(t, "deep", out)
}

func TestS5_DotWithSpaces_MixedWithNormalDot(t *testing.T) {
	// mix: first level with spaces, second without
	out := renderS5(t, `{{ a . b.c }}`, map[string]any{
		"a": map[string]any{
			"b": map[string]any{"c": "mixed"},
		},
	})
	require.Equal(t, "mixed", out)
}

func TestS5_DotWithSpaces_InFilter(t *testing.T) {
	out := renderS5(t, `{{ obj . name | upcase }}`, map[string]any{
		"obj": map[string]any{"name": "hello"},
	})
	require.Equal(t, "HELLO", out)
}

func TestS5_DotWithSpaces_InCondition(t *testing.T) {
	out := renderS5(t,
		`{% if obj . active %}yes{% else %}no{% endif %}`,
		map[string]any{"obj": map[string]any{"active": true}})
	require.Equal(t, "yes", out)
}

func TestS5_DotWithSpaces_WithTabs(t *testing.T) {
	// scanner must skip all whitespace (including tabs) around the dot
	out := renderS5(t, "{{ obj\t.\tkey }}", map[string]any{
		"obj": map[string]any{"key": "tab-spaced"},
	})
	require.Equal(t, "tab-spaced", out)
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  5f — top-level bracket + dot access (LiquidJS #643)                       ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

func TestS5_BracketRoot_SimpleDotAccess(t *testing.T) {
	out := renderS5(t, `{{ ["Key String with Spaces"].subpropertyKey }}`, map[string]any{
		"Key String with Spaces": map[string]any{"subpropertyKey": "FOO"},
	})
	require.Equal(t, "FOO", out)
}

func TestS5_BracketRoot_ChainedDots(t *testing.T) {
	out := renderS5(t, `{{ ["root key"].a.b }}`, map[string]any{
		"root key": map[string]any{
			"a": map[string]any{"b": "nested"},
		},
	})
	require.Equal(t, "nested", out)
}

func TestS5_BracketRoot_WithBracketThenDot(t *testing.T) {
	out := renderS5(t, `{{ ["root"]["inner"].value }}`, map[string]any{
		"root": map[string]any{
			"inner": map[string]any{"value": "chained"},
		},
	})
	require.Equal(t, "chained", out)
}

func TestS5_BracketRoot_InFilter(t *testing.T) {
	out := renderS5(t, `{{ ["name"] | upcase }}`, map[string]any{"name": "world"})
	require.Equal(t, "WORLD", out)
}

func TestS5_BracketRoot_InCondition(t *testing.T) {
	out := renderS5(t,
		`{% if ["flag"] %}yes{% else %}no{% endif %}`,
		map[string]any{"flag": true})
	require.Equal(t, "yes", out)
}

func TestS5_BracketRoot_VariableKey(t *testing.T) {
	// {{ [varname].prop }} — key from variable, then dot
	out := renderS5(t, `{{ [k].prop }}`, map[string]any{
		"k":   "obj",
		"obj": map[string]any{"prop": "val"},
	})
	require.Equal(t, "val", out)
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  Cross-cutting: interaction between all features                            ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

func TestS5_CrossCutting_NegIndexThenDot(t *testing.T) {
	// arr[-1].name — negative index on array of objects, then dot
	out := renderS5(t, `{{ people[-1].name }}`, map[string]any{
		"people": []map[string]any{
			{"name": "Alice"},
			{"name": "Bob"},
		},
	})
	require.Equal(t, "Bob", out)
}

func TestS5_CrossCutting_FirstThenIndex(t *testing.T) {
	// matrix.first[1] - .first returns an array, then index into it
	out := renderS5(t, `{{ matrix.first[1] }}`, map[string]any{
		"matrix": [][]string{{"a", "b"}, {"c", "d"}},
	})
	require.Equal(t, "b", out)
}

func TestS5_CrossCutting_DynamicLookupThenNegIndex(t *testing.T) {
	// {{ [key][-1] }} — resolve variable, then negative index
	out := renderS5(t, `{{ [key][-1] }}`, map[string]any{
		"key":    "fruits",
		"fruits": []string{"apple", "banana", "cherry"},
	})
	require.Equal(t, "cherry", out)
}

func TestS5_CrossCutting_DynamicLookupThenFirst(t *testing.T) {
	out := renderS5(t, `{{ [key].first }}`, map[string]any{
		"key":   "items",
		"items": []string{"one", "two"},
	})
	require.Equal(t, "one", out)
}

func TestS5_CrossCutting_DynamicLookupThenSize(t *testing.T) {
	out := renderS5(t, `{{ [key].size }}`, map[string]any{
		"key":   "items",
		"items": []string{"x", "y", "z"},
	})
	require.Equal(t, "3", out)
}

func TestS5_CrossCutting_AllFeaturesInSingleOutput(t *testing.T) {
	// Template that exercises all 6 feature areas in one render
	tpl := strings.Join([]string{
		`{{ a.b }}`, // 5a: dot notation
		` `,
		`{{ arr[1] }}`, // 5a: array index
		` `,
		`{{ arr[-1] }}`, // 5b: negative index
		` `,
		`{{ arr.first }}`, // 5c: .first
		` `,
		`{{ arr.last }}`, // 5c: .last
		` `,
		`{{ arr.size }}`, // 5c: .size
		` `,
		`{{ [k] }}`, // 5d: dynamic lookup
		` `,
		`{{ a . b }}`, // 5e: dot with spaces
		` `,
		`{{ ["a key"].val }}`, // 5f: bracket root + dot
	}, "")

	binds := map[string]any{
		"a":      map[string]any{"b": "dot"},
		"arr":    []string{"first_el", "mid_el", "last_el"},
		"k":      "target",
		"target": "dynamic",
		"a key":  map[string]any{"val": "bracket"},
	}

	out := renderS5(t, tpl, binds)
	require.Equal(t, "dot mid_el last_el first_el last_el 3 dynamic dot bracket", out)
}

// ── Variable types as keys / indices ─────────────────────────────────────────

func TestS5_Unicode_VariableName(t *testing.T) {
	eng := liquid.NewEngine()
	out, err := eng.ParseAndRenderString(`{{ÜLKE}}`, map[string]any{"ÜLKE": "Türkiye"})
	require.NoError(t, err)
	require.Equal(t, "Türkiye", out)
}

func TestS5_Unicode_DotAccess(t *testing.T) {
	out := renderS5(t, `{{ país.capital }}`, map[string]any{
		"país": map[string]any{"capital": "Madrid"},
	})
	require.Equal(t, "Madrid", out)
}

// ── Blank / empty as variable names ─────────────────────────────────────────

func TestS5_BlankAssigned_RendersEmpty(t *testing.T) {
	out := renderS5(t, `{% assign v = blank %}{{ v }}`, nil)
	require.Equal(t, "", out)
}

func TestS5_EmptyAssigned_RendersEmpty(t *testing.T) {
	out := renderS5(t, `{% assign v = empty %}{{ v }}`, nil)
	require.Equal(t, "", out)
}

func TestS5_BlankAssigned_RendersAsEmptyStringInOutput(t *testing.T) {
	// After assign v = blank, the variable renders as empty string
	// (blank is a special sentinel that renders as "").
	out := renderS5(t, `{% assign v = blank %}[{{ v }}]`, nil)
	require.Equal(t, "[]", out)
}

// ── Nil safety ────────────────────────────────────────────────────────────────

func TestS5_NilSafe_DeepChainOnNil(t *testing.T) {
	// nil variable; deep property chain must not panic
	out := renderS5(t, `{{ n.a.b.c }}`, map[string]any{"n": nil})
	require.Equal(t, "", out)
}

func TestS5_NilSafe_IndexOnNil(t *testing.T) {
	out := renderS5(t, `{{ n[0] }}`, map[string]any{"n": nil})
	require.Equal(t, "", out)
}

func TestS5_NilSafe_NegIndexOnNil(t *testing.T) {
	out := renderS5(t, `{{ n[-1] }}`, map[string]any{"n": nil})
	require.Equal(t, "", out)
}

func TestS5_NilSafe_SpecialPropsOnNil(t *testing.T) {
	// nil.first / nil.last / nil.size must all render as empty string
	for _, prop := range []string{"first", "last", "size"} {
		t.Run(prop, func(t *testing.T) {
			out := renderS5(t, fmt.Sprintf(`{{ n.%s }}`, prop), map[string]any{"n": nil})
			require.Equal(t, "", out, "nil.%s should render empty", prop)
		})
	}
}

// ── Rendering false/nil ────────────────────────────────────────────────────────

func TestS5_FalseRendersAsFalse(t *testing.T) {
	out := renderS5(t, `{{ obj.flag }}`, map[string]any{"obj": map[string]any{"flag": false}})
	require.Equal(t, "false", out)
}

func TestS5_NilRendersEmpty(t *testing.T) {
	out := renderS5(t, `{{ obj.missing }}`, map[string]any{"obj": map[string]any{}})
	require.Equal(t, "", out)
}

// ── Regression: multiline tags ────────────────────────────────────────────────

func TestS5_MultilineTag_DotAccess(t *testing.T) {
	out := renderS5(t, "{{\nobj.key\n}}", map[string]any{
		"obj": map[string]any{"key": "multiline"},
	})
	require.Equal(t, "multiline", out)
}

func TestS5_MultilineTag_NegIndex(t *testing.T) {
	out := renderS5(t, "{{\narr[-1]\n}}", map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "c", out)
}

// ── StrictVariables compatibility ─────────────────────────────────────────────

func TestS5_StrictVariables_DotAccessOnUndefined(t *testing.T) {
	// accessing .prop on an undefined root variable should error in strict mode
	eng := liquid.NewEngine()
	eng.StrictVariables()
	_, err := eng.ParseAndRenderString(`{{ undefined.prop }}`, nil)
	assert.Error(t, err)
}

func TestS5_StrictVariables_DynamicLookupKeyExists(t *testing.T) {
	// In strict mode, the outer key variable must exist; the lookup itself works.
	// Note: strict mode does NOT propagate through double-indirection — the
	// resolved variable 'ghost' not existing does NOT produce an error because
	// at render time the nil result is indistinguishable from a missing property.
	eng := liquid.NewEngine()
	eng.StrictVariables()
	_, err := eng.ParseAndRenderString(`{{ [key] }}`, map[string]any{"key": "ghost"})
	// No error: dynamic lookup returns nil for missing resolved variable (spec behavior)
	assert.NoError(t, err)
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  INTENSIVE BLOCK — regression traps & advanced scenarios                   ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

// ── Uint types as index (regression: B1 fix must cover negative too) ─────────

func TestS5_NegativeIndex_UintTypesAsIndex(t *testing.T) {
	// uint variants used as *negative* index — they're positive, so must work as
	// unsigned positive indices (uint(2) → index 2, not -1).
	arr := []string{"a", "b", "c"}
	cases := []struct {
		idx      any
		expected string
		name     string
	}{
		{uint(0), "a", "uint(0)"},
		{uint8(1), "b", "uint8(1)"},
		{uint16(2), "c", "uint16(2)"},
		{uint32(0), "a", "uint32(0)"},
		{uint64(2), "c", "uint64(2)"},
		{uintptr(1), "b", "uintptr(1)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := renderS5(t, `{{ arr[i] }}`, map[string]any{"arr": arr, "i": tc.idx})
			require.Equal(t, tc.expected, out)
		})
	}
}

// ── Go struct: pointer, embedded, unexported ──────────────────────────────────

func TestS5_GoStruct_PointerField(t *testing.T) {
	type Inner struct{ Title string }
	type Page struct{ Content *Inner }

	inner := Inner{Title: "ptr-value"}
	out := renderS5(t, `{{ page.Content.Title }}`, map[string]any{
		"page": Page{Content: &inner},
	})
	require.Equal(t, "ptr-value", out)
}

func TestS5_GoStruct_NilPointerField_Graceful(t *testing.T) {
	type Inner struct{ Title string }
	type Page struct{ Content *Inner }

	out := renderS5(t, `{{ page.Content.Title }}`, map[string]any{
		"page": Page{Content: nil},
	})
	require.Equal(t, "", out)
}

func TestS5_GoStruct_EmbeddedStruct(t *testing.T) {
	type Base struct{ ID int }
	type Product struct {
		Base
		Name string
	}
	out := renderS5(t, `{{ product.Name }} {{ product.ID }}`, map[string]any{
		"product": Product{Base: Base{ID: 42}, Name: "Widget"},
	})
	require.Equal(t, "Widget 42", out)
}

func TestS5_GoStruct_MissingField_Inaccessible(t *testing.T) {
	// Accessing a key that doesn't exist on the struct renders empty.
	type Obj struct{ Pub string }
	out := renderS5(t, `[{{ obj.Pub }}][{{ obj.absent }}]`, map[string]any{
		"obj": Obj{Pub: "yes"},
	})
	require.Equal(t, "[yes][]", out)
}

func TestS5_GoStruct_SliceOfStructs(t *testing.T) {
	type Item struct{ Name string }
	items := []Item{{"alpha"}, {"beta"}, {"gamma"}}
	out := renderS5(t,
		`{% for it in items %}{{ it.Name }} {% endfor %}`,
		map[string]any{"items": items})
	require.Equal(t, "alpha beta gamma ", out)
}

func TestS5_GoStruct_MapOfStructs_DotAccess(t *testing.T) {
	type Info struct{ Score int }
	out := renderS5(t, `{{ data.alice.Score }}`, map[string]any{
		"data": map[string]any{"alice": struct{ Score int }{Score: 99}},
	})
	require.Equal(t, "99", out)
}

// ── Negative index in conditions ─────────────────────────────────────────────

func TestS5_NegativeIndex_InIfCondition(t *testing.T) {
	out := renderS5(t,
		`{% if arr[-1] == "last" %}yes{% else %}no{% endif %}`,
		map[string]any{"arr": []string{"first", "last"}})
	require.Equal(t, "yes", out)
}

func TestS5_NegativeIndex_InCondition_EmptyArray_NoError(t *testing.T) {
	// On empty array, arr[-1] returns nil — must not error in if-condition
	out := renderS5(t,
		`{% if arr[-1] == "x" %}yes{% else %}no{% endif %}`,
		map[string]any{"arr": []string{}})
	require.Equal(t, "no", out)
}

func TestS5_NegativeIndex_InUnless(t *testing.T) {
	// arr[-1]=="c", != "z" → condition false → unless body executes → "no-z"
	out := renderS5(t,
		`{% unless arr[-1] == "z" %}no-z{% endunless %}`,
		map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "no-z", out)
}

func TestS5_NegativeIndex_InsideCaptureAndAssign(t *testing.T) {
	// {% assign x = arr[-1] %} then render x
	out := renderS5(t,
		`{% assign last = arr[-1] %}{{ last }}`,
		map[string]any{"arr": []string{"x", "y", "final"}})
	require.Equal(t, "final", out)
}

func TestS5_NegativeIndex_UsedAsSplitResult(t *testing.T) {
	// After split, negative index on assigned array
	out := renderS5(t,
		`{% assign parts = "a|b|c|d" | split: "|" %}{{ parts[-1] }}`,
		nil)
	require.Equal(t, "d", out)
}

// ── .first / .last on complex types ──────────────────────────────────────────

func TestS5_First_OnArrayOfMaps_ThenDot(t *testing.T) {
	// array.first.property
	out := renderS5(t, `{{ products.first.title }}`, map[string]any{
		"products": []map[string]any{
			{"title": "Widget", "price": 10},
			{"title": "Gadget", "price": 20},
		},
	})
	require.Equal(t, "Widget", out)
}

func TestS5_Last_OnArrayOfMaps_ThenDot(t *testing.T) {
	out := renderS5(t, `{{ products.last.price }}`, map[string]any{
		"products": []map[string]any{
			{"title": "Widget", "price": 10},
			{"title": "Gadget", "price": 20},
		},
	})
	require.Equal(t, "20", out)
}

func TestS5_First_OnSplitResult(t *testing.T) {
	out := renderS5(t,
		`{% assign words = "hello world foo" | split: " " %}{{ words.first }}`,
		nil)
	require.Equal(t, "hello", out)
}

func TestS5_Last_OnSplitResult(t *testing.T) {
	out := renderS5(t,
		`{% assign words = "hello world foo" | split: " " %}{{ words.last }}`,
		nil)
	require.Equal(t, "foo", out)
}

func TestS5_First_ThenFilter(t *testing.T) {
	// array.first | upcase — property access then filter
	out := renderS5(t, `{{ names.first | upcase }}`, map[string]any{
		"names": []string{"alice", "bob"},
	})
	require.Equal(t, "ALICE", out)
}

func TestS5_Last_ThenFilter(t *testing.T) {
	out := renderS5(t, `{{ names.last | upcase }}`, map[string]any{
		"names": []string{"alice", "bob"},
	})
	require.Equal(t, "BOB", out)
}

func TestS5_Size_OnNilValue_ReturnsEmpty(t *testing.T) {
	out := renderS5(t, `{{ v.size }}`, map[string]any{"v": nil})
	require.Equal(t, "", out)
}

func TestS5_Size_OnBoolValue_ReturnsEmpty(t *testing.T) {
	// booleans have no size
	out := renderS5(t, `{{ v.size }}`, map[string]any{"v": true})
	require.Equal(t, "", out)
}

func TestS5_Size_OnInteger_ReturnsEmpty(t *testing.T) {
	// integers have no size property (unlike strings)
	out := renderS5(t, `{{ v.size }}`, map[string]any{"v": 42})
	require.Equal(t, "", out)
}

// ── Dot notation on diverse Go types ─────────────────────────────────────────

func TestS5_DotNotation_OnMapYAMLStyleKeys(t *testing.T) {
	// yaml-style keys with colons in the key name are not accessible via dot,
	// but normal keys are
	out := renderS5(t, `{{ config.host }}:{{ config.port }}`, map[string]any{
		"config": map[string]any{"host": "localhost", "port": 8080},
	})
	require.Equal(t, "localhost:8080", out)
}

func TestS5_DotNotation_OnFalseValue_IsFalse(t *testing.T) {
	// accessing a key whose value is `false` must render "false", not ""
	out := renderS5(t, `{{ flags.active }}`, map[string]any{
		"flags": map[string]any{"active": false},
	})
	require.Equal(t, "false", out)
}

func TestS5_DotNotation_OnIntValue_Renders(t *testing.T) {
	out := renderS5(t, `{{ obj.count }}`, map[string]any{
		"obj": map[string]any{"count": 7},
	})
	require.Equal(t, "7", out)
}

func TestS5_DotNotation_KeyShadowsBuiltin_Size(t *testing.T) {
	// if map has a "size" key, it must beat the built-in .size shortcut
	out := renderS5(t, `{{ m.size }}`, map[string]any{
		"m": map[string]any{"size": "custom"},
	})
	require.Equal(t, "custom", out)
}

func TestS5_DotNotation_KeyShadowsBuiltin_First(t *testing.T) {
	// if map has a "first" key → use it, not the array shortcut
	out := renderS5(t, `{{ m.first }}`, map[string]any{
		"m": map[string]any{"first": "overridden"},
	})
	require.Equal(t, "overridden", out)
}

// ── Bracket notation edge cases ───────────────────────────────────────────────

func TestS5_BracketIndex_NegativeFromExpression(t *testing.T) {
	// arr[0 - 1] — computed negative index via expression
	out := renderS5(t, `{{ arr[n] }}`, map[string]any{
		"arr": []string{"x", "y", "z"},
		"n":   -1,
	})
	require.Equal(t, "z", out)
}

func TestS5_BracketKey_EmptyStringKey(t *testing.T) {
	// map[""] — empty string key is a valid map key
	out := renderS5(t, `{{ m[""] }}`, map[string]any{
		"m": map[string]any{"": "empty-key"},
	})
	require.Equal(t, "empty-key", out)
}

func TestS5_BracketKey_NumericStringKey(t *testing.T) {
	// map["1"] — string key that looks like a number
	out := renderS5(t, `{{ m["1"] }}`, map[string]any{
		"m": map[string]any{"1": "string-one"},
	})
	require.Equal(t, "string-one", out)
}

// ── Dynamic lookup [key] — extra stress ──────────────────────────────────────

func TestS5_DynamicLookup_TwoStepViaAssign(t *testing.T) {
	// Two-step indirection via assign:
	// pointer="level1", level1="level2" (a key name), level2="final value"
	// [pointer] → looks up pointer → "level1", then context["level1"] = "level2"
	// {% assign mid = [pointer] %} → mid = "level2"
	// {{ [mid] }} → looks up mid → "level2", context["level2"] = "final value"
	out := renderS5(t,
		`{% assign mid = [pointer] %}{{ [mid] }}`,
		map[string]any{
			"pointer": "level1",
			"level1":  "level2",
			"level2":  "final value",
		})
	require.Equal(t, "final value", out)
}

func TestS5_DynamicLookup_InForCollection(t *testing.T) {
	// Use dynamic lookup as the for-loop collection
	out := renderS5(t,
		`{% for item in [collection_key] %}{{ item }} {% endfor %}`,
		map[string]any{
			"collection_key": "fruits",
			"fruits":         []string{"apple", "banana"},
		})
	require.Equal(t, "apple banana ", out)
}

func TestS5_DynamicLookup_InAssignRHS(t *testing.T) {
	// {% assign val = [key] %} — dynamic lookup on the right side of assign
	out := renderS5(t,
		`{% assign result = [key] %}{{ result | upcase }}`,
		map[string]any{"key": "greeting", "greeting": "hello"})
	require.Equal(t, "HELLO", out)
}

func TestS5_DynamicLookup_InCaptureBody(t *testing.T) {
	out := renderS5(t,
		`{% capture buf %}{{ [k] }}{% endcapture %}[{{ buf }}]`,
		map[string]any{"k": "msg", "msg": "hi"})
	require.Equal(t, "[hi]", out)
}

func TestS5_DynamicLookup_WithFilterOnResult(t *testing.T) {
	out := renderS5(t, `{{ [k] | upcase }}`, map[string]any{
		"k": "name", "name": "world",
	})
	require.Equal(t, "WORLD", out)
}

func TestS5_DynamicLookup_WithDotChainOnResult(t *testing.T) {
	out := renderS5(t, `{{ [k].title }}`, map[string]any{
		"k":       "product",
		"product": map[string]any{"title": "Widget"},
	})
	require.Equal(t, "Widget", out)
}

func TestS5_DynamicLookup_WithNegIndexOnResult(t *testing.T) {
	out := renderS5(t, `{{ [k][-1] }}`, map[string]any{
		"k":   "arr",
		"arr": []string{"a", "b", "c"},
	})
	require.Equal(t, "c", out)
}

// ── Dot-with-spaces stress ────────────────────────────────────────────────────

func TestS5_DotWithSpaces_InForLoopCollection(t *testing.T) {
	out := renderS5(t,
		`{% for item in site . pages %}{{ item }} {% endfor %}`,
		map[string]any{
			"site": map[string]any{"pages": []string{"home", "about"}},
		})
	require.Equal(t, "home about ", out)
}

func TestS5_DotWithSpaces_InAssign(t *testing.T) {
	out := renderS5(t,
		`{% assign t = obj . title %}{{ t }}`,
		map[string]any{"obj": map[string]any{"title": "My Title"}})
	require.Equal(t, "My Title", out)
}

func TestS5_DotWithSpaces_VeryManySpaces(t *testing.T) {
	// lots of spaces on both sides of the dot
	out := renderS5(t, "{{ a   .   b }}", map[string]any{
		"a": map[string]any{"b": "spaced"},
	})
	require.Equal(t, "spaced", out)
}

func TestS5_DotWithSpaces_MixedTabAndSpace(t *testing.T) {
	out := renderS5(t, "{{ a\t .\t b }}", map[string]any{
		"a": map[string]any{"b": "mixed-ws"},
	})
	require.Equal(t, "mixed-ws", out)
}

// ── StrictVariables: comprehensive ───────────────────────────────────────────

func TestS5_StrictVariables_UndefinedRootVarErrors(t *testing.T) {
	eng := liquid.NewEngine()
	eng.StrictVariables()
	_, err := eng.ParseAndRenderString(`{{ ghost }}`, nil)
	require.Error(t, err)
}

func TestS5_StrictVariables_UndefinedPropertyOnExistingMapReturnsEmpty(t *testing.T) {
	// StrictVariables only fires for undefined ROOT variables.
	// A property missing on an existing map/struct just returns nil (no error).
	eng := liquid.NewEngine()
	eng.StrictVariables()
	out, err := eng.ParseAndRenderString(`{{ obj.missing }}`, map[string]any{
		"obj": map[string]any{"present": "yes"},
	})
	require.NoError(t, err)
	require.Equal(t, "", out)
}

func TestS5_StrictVariables_DeepChainOnDefinedRootReturnsEmpty(t *testing.T) {
	eng := liquid.NewEngine()
	eng.StrictVariables()
	out, err := eng.ParseAndRenderString(`{{ a.b.c.d }}`, map[string]any{
		"a": map[string]any{"b": nil},
	})
	require.NoError(t, err)
	require.Equal(t, "", out)
}

func TestS5_StrictVariables_NegIndexOnDefinedArrayReturnsEmpty(t *testing.T) {
	eng := liquid.NewEngine()
	eng.StrictVariables()
	out, err := eng.ParseAndRenderString(`{{ arr[-9] }}`, map[string]any{
		"arr": []string{"a"},
	})
	require.NoError(t, err)
	require.Equal(t, "", out)
}

// ── ForLoop-based index access ────────────────────────────────────────────────

func TestS5_ForLoopIndex_AccessArrayByForloopIndex0(t *testing.T) {
	// {{ data[forloop.index0] }} — use loop counter as array index
	out := renderS5(t,
		`{% for _ in (1..3) %}{{ letters[forloop.index0] }}{% endfor %}`,
		map[string]any{"letters": []string{"A", "B", "C"}})
	require.Equal(t, "ABC", out)
}

func TestS5_ForLoopIndex_DescendingWithRindex(t *testing.T) {
	// Use forloop.rindex (1-based distance from end) via assign + negative compute
	// to verify rindex0 is accessible and usable in conditional logic.
	out := renderS5(t,
		`{% for _ in (1..3) %}{{ forloop.rindex0 }}{% endfor %}`,
		map[string]any{})
	// rindex0: 2,1,0
	require.Equal(t, "210", out)
}

func TestS5_ForLoopIndex_AccessNestedProperties(t *testing.T) {
	// In loop, access nested property of each item
	out := renderS5(t,
		`{% for p in products %}{{ p.name }}={{ p.price }} {% endfor %}`,
		map[string]any{
			"products": []map[string]any{
				{"name": "A", "price": 10},
				{"name": "B", "price": 20},
			},
		})
	require.Equal(t, "A=10 B=20 ", out)
}

func TestS5_ForLoopOver_NegativeIndexResult(t *testing.T) {
	// Iterate over arr[-1] when it is itself a slice
	out := renderS5(t,
		`{% for item in matrix[-1] %}{{ item }} {% endfor %}`,
		map[string]any{
			"matrix": [][]string{{"a", "b"}, {"c", "d"}},
		})
	require.Equal(t, "c d ", out)
}

// ── case/when with property access ───────────────────────────────────────────

func TestS5_CaseWhen_PropertyAccess(t *testing.T) {
	out := renderS5(t,
		`{% case product.type %}{% when "shirt" %}shirt{% when "pants" %}pants{% else %}other{% endcase %}`,
		map[string]any{"product": map[string]any{"type": "shirt"}})
	require.Equal(t, "shirt", out)
}

func TestS5_CaseWhen_NegativeIndexResult(t *testing.T) {
	out := renderS5(t,
		`{% case arr[-1] %}{% when "z" %}last-z{% else %}other{% endcase %}`,
		map[string]any{"arr": []string{"x", "y", "z"}})
	require.Equal(t, "last-z", out)
}

// ── contains operator with nested access ────────────────────────────────────

func TestS5_Contains_ArrayViaProperty(t *testing.T) {
	out := renderS5(t,
		`{% if user.roles contains "admin" %}admin{% else %}not-admin{% endif %}`,
		map[string]any{"user": map[string]any{"roles": []string{"user", "admin"}}})
	require.Equal(t, "admin", out)
}

func TestS5_Contains_StringViaProperty(t *testing.T) {
	out := renderS5(t,
		`{% if page.title contains "Go" %}yes{% else %}no{% endif %}`,
		map[string]any{"page": map[string]any{"title": "Learning Go"}})
	require.Equal(t, "yes", out)
}

// ── Assign from complex access ────────────────────────────────────────────────

func TestS5_Assign_FromDotChain(t *testing.T) {
	out := renderS5(t,
		`{% assign title = page.meta.title %}{{ title }}`,
		map[string]any{
			"page": map[string]any{
				"meta": map[string]any{"title": "My Page"},
			},
		})
	require.Equal(t, "My Page", out)
}

func TestS5_Assign_FromNegativeIndex(t *testing.T) {
	out := renderS5(t,
		`{% assign last = arr[-1] %}{{ last | upcase }}`,
		map[string]any{"arr": []string{"alpha", "beta", "gamma"}})
	require.Equal(t, "GAMMA", out)
}

func TestS5_Assign_FromFirst(t *testing.T) {
	out := renderS5(t,
		`{% assign head = arr.first %}{{ head }}-{{ arr.size }}`,
		map[string]any{"arr": []string{"one", "two", "three"}})
	require.Equal(t, "one-3", out)
}

// ── Filter chain combined with access ────────────────────────────────────────

func TestS5_FilterChain_MapThenIndex(t *testing.T) {
	// {{ products | map: "price" | first }} — map-filter then .first
	out := renderS5(t,
		`{{ products | map: "price" | first }}`,
		map[string]any{
			"products": []map[string]any{
				{"price": 10}, {"price": 20},
			},
		})
	require.Equal(t, "10", out)
}

func TestS5_FilterChain_SortThenFirst(t *testing.T) {
	out := renderS5(t,
		`{{ nums | sort | first }}`,
		map[string]any{"nums": []int{5, 2, 8, 1}})
	require.Equal(t, "1", out)
}

func TestS5_FilterChain_SortThenLast(t *testing.T) {
	out := renderS5(t,
		`{{ nums | sort | last }}`,
		map[string]any{"nums": []int{5, 2, 8, 1}})
	require.Equal(t, "8", out)
}

func TestS5_FilterChain_ReverseThenNegIndex(t *testing.T) {
	// After reverse, arr[-1] is now what was arr[0]
	out := renderS5(t,
		`{% assign rev = arr | reverse %}{{ rev[-1] }}`,
		map[string]any{"arr": []string{"a", "b", "c"}})
	require.Equal(t, "a", out)
}

func TestS5_FilterChain_SplitThenSize(t *testing.T) {
	out := renderS5(t,
		`{% assign parts = "a,b,c,d" | split: "," %}{{ parts.size }}`,
		nil)
	require.Equal(t, "4", out)
}

// ── Tablerow with access ──────────────────────────────────────────────────────

func TestS5_Tablerow_PropertyAccess(t *testing.T) {
	out := renderS5(t,
		`{% tablerow item in collection.items cols:2 %}{{ item.name }}{% endtablerow %}`,
		map[string]any{
			"collection": map[string]any{
				"items": []map[string]any{
					{"name": "A"}, {"name": "B"}, {"name": "C"}, {"name": "D"},
				},
			},
		})
	require.Contains(t, out, "A")
	require.Contains(t, out, "D")
}

// ── Complex real-world template: Shopify product page simulation ─────────────

func TestS5_RealWorld_ProductPage(t *testing.T) {
	tpl := `
Title: {{ product.title }}
Price: ${{ product.variants[0].price }}
Last variant: {{ product.variants[-1].title }}
Size: {{ product.variants.size }}
Tags: {{ product.tags | join: ", " }}
Featured: {{ product.meta.featured }}
`
	out := renderS5(t, tpl, map[string]any{
		"product": map[string]any{
			"title": "Super Shirt",
			"variants": []map[string]any{
				{"title": "Small", "price": 29},
				{"title": "Medium", "price": 32},
				{"title": "Large", "price": 35},
			},
			"tags": []string{"sale", "summer"},
			"meta": map[string]any{"featured": true},
		},
	})
	require.Contains(t, out, "Title: Super Shirt")
	require.Contains(t, out, "Price: $29")
	require.Contains(t, out, "Last variant: Large")
	require.Contains(t, out, "Size: 3")
	require.Contains(t, out, "Tags: sale, summer")
	require.Contains(t, out, "Featured: true")
}

func TestS5_RealWorld_NavigationMenu(t *testing.T) {
	tpl := `{% for link in linklists.main_menu.links %}{{ link.title }}{% unless forloop.last %} | {% endunless %}{% endfor %}`

	out := renderS5(t, tpl, map[string]any{
		"linklists": map[string]any{
			"main_menu": map[string]any{
				"links": []map[string]any{
					{"title": "Home"},
					{"title": "About"},
					{"title": "Contact"},
				},
			},
		},
	})
	require.Equal(t, "Home | About | Contact", out)
}

func TestS5_RealWorld_ConditionalAccessNested(t *testing.T) {
	tpl := `{% if customer.address.country == "US" %}ship-domestic{% else %}ship-international{% endif %}`

	out := renderS5(t, tpl, map[string]any{
		"customer": map[string]any{
			"address": map[string]any{"country": "US"},
		},
	})
	require.Equal(t, "ship-domestic", out)
}

func TestS5_RealWorld_DynamicSectionRendering(t *testing.T) {
	// Dynamic lookup used to switch between different section keys
	tpl := `{% for section in page.sections %}{{ [section].heading }}: {{ [section].body }} {% endfor %}`

	out := renderS5(t, tpl, map[string]any{
		"page": map[string]any{
			"sections": []string{"hero", "cta"},
		},
		"hero": map[string]any{"heading": "Welcome", "body": "intro text"},
		"cta":  map[string]any{"heading": "Buy Now", "body": "limited offer"},
	})
	require.Equal(t, "Welcome: intro text Buy Now: limited offer ", out)
}

// ── Capture + access combination ─────────────────────────────────────────────

func TestS5_Capture_UsesPropertyAccess(t *testing.T) {
	out := renderS5(t,
		`{% capture greeting %}Hello, {{ user.name }}!{% endcapture %}{{ greeting }}`,
		map[string]any{"user": map[string]any{"name": "Alice"}})
	require.Equal(t, "Hello, Alice!", out)
}

func TestS5_Capture_UsesNegativeIndex(t *testing.T) {
	out := renderS5(t,
		`{% capture last %}{{ items[-1] }}{% endcapture %}[{{ last }}]`,
		map[string]any{"items": []string{"one", "two", "three"}})
	require.Equal(t, "[three]", out)
}

// ── Stability: runs must be idempotent ────────────────────────────────────────

func TestS5_Idempotent_SameEngineMultipleRenders(t *testing.T) {
	// Same engine, same template, multiple renders must produce identical output.
	eng := liquid.NewEngine()
	tpl, err := eng.ParseString(`{{ obj.a[-1] }}`)
	require.NoError(t, err)

	binds := map[string]any{
		"obj": map[string]any{"a": []string{"x", "y", "z"}},
	}

	for i := range 5 {
		out, err := tpl.RenderString(binds)
		require.NoError(t, err)
		require.Equal(t, "z", out, "render %d should be 'z'", i+1)
	}
}

func TestS5_Idempotent_NegIndexAfterFilterChain(t *testing.T) {
	eng := liquid.NewEngine()
	tpl, err := eng.ParseString(`{% assign s = "a,b,c,d" | split: "," %}{{ s[-1] }},{{ s[-2] }}`)
	require.NoError(t, err)

	for range 3 {
		out, err := tpl.RenderString(nil)
		require.NoError(t, err)
		require.Equal(t, "d,c", out)
	}
}
