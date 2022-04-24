package expressions

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/osteele/liquid/values"
	"github.com/stretchr/testify/require"
)

var evaluatorTests = []struct {
	in       string
	expected interface{}
}{
	// Literals
	{`12`, 12},
	{`12.3`, 12.3},
	{`true`, true},
	{`false`, false},
	{`'abc'`, "abc"},
	{`"abc"`, "abc"},

	// Variables
	{`n`, 123},

	// Attributes
	{`hash.a`, "first"},
	{`hash.b.c`, "d"},
	{`hash["b"].c`, "d"},
	{`hash.x`, nil},
	{`fruits.first`, "apples"},
	{`fruits.last`, "plums"},
	{`empty_list.first`, nil},
	{`empty_list.last`, nil},
	{`"abc".size`, 3},
	{`fruits.size`, 4},
	{`hash.size`, 3},
	{`hash_with_size_key.size`, "key_value"},

	// Indices
	{`array[1]`, "second"},
	{`array[-1]`, "third"}, // undocumented
	{`array[100]`, nil},
	{`hash[1]`, nil},
	{`hash.c[0]`, "r"},

	// Range
	{`(1..5)`, values.NewRange(1, 5)},
	{`(1..range.end)`, values.NewRange(1, 5)},
	{`(1..range["end"])`, values.NewRange(1, 5)},
	{`(range.begin..range.end)`, values.NewRange(1, 5)},

	// Expressions
	{`(1)`, 1},
	{`(n)`, 123},

	// Operators
	{`1 == 1`, true},
	{`1 == 2`, false},
	{`1.0 == 1.0`, true},
	{`1.0 == 2.0`, false},
	{`1.0 == 1`, true},
	{`1 == 1.0`, true},
	{`"a" == "a"`, true},
	{`"a" == "b"`, false},

	{`1 != 1`, false},
	{`1 != 2`, true},
	{`1.0 != 1.0`, false},
	{`1 != 1.0`, false},
	{`1 != 2.0`, true},

	{`1 < 2`, true},
	{`2 < 1`, false},
	{`1.0 < 2.0`, true},
	{`1.0 < 2`, true},
	{`1 < 2.0`, true},
	{`1.0 < 2`, true},
	{`"a" < "a"`, false},
	{`"a" < "b"`, true},
	{`"b" < "a"`, false},

	{`1 > 2`, false},
	{`2 > 1`, true},

	{`1 <= 1`, true},
	{`1 <= 2`, true},
	{`2 <= 1`, false},
	{`"a" <= "a"`, true},
	{`"a" <= "b"`, true},
	{`"b" <= "a"`, false},

	{`1 >= 1`, true},
	{`1 >= 2`, false},
	{`2 >= 1`, true},

	{`true and false`, false},
	{`true and true`, true},
	{`true and true and true`, true},
	{`false or false`, false},
	{`false or true`, true},

	{`"seafood" contains "foo"`, true},
	{`"seafood" contains "bar"`, false},
	{`array contains "first"`, true},
	{`interface_array contains "first"`, true},
	{`"foo" contains "missing"`, false},
	{`nil contains "missing"`, false},

	// filters
	{`"seafood" | length`, 8},
}

var evaluatorTestBindings = (map[string]interface{}{
	"n":               123,
	"array":           []string{"first", "second", "third"},
	"interface_array": []interface{}{"first", "second", "third"},
	"empty_list":      []interface{}{},
	"fruits":          []string{"apples", "oranges", "peaches", "plums"},
	"hash": map[string]interface{}{
		"a": "first",
		"b": map[string]interface{}{"c": "d"},
		"c": []string{"r", "g", "b"},
	},
	"hash_with_size_key": map[string]interface{}{"size": "key_value"},
	"range": map[string]interface{}{
		"begin": 1,
		"end":   5,
	},
})

func TestEvaluateString(t *testing.T) {
	cfg := NewConfig()
	cfg.AddFilter("length", strings.Count)
	ctx := NewContext(evaluatorTestBindings, cfg)
	for i, test := range evaluatorTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			val, err := EvaluateString(test.in, ctx)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, val, test.in)
		})
	}

	_, err := EvaluateString("syntax error", ctx)
	require.Error(t, err)

	_, err = EvaluateString("1 | undefined_filter", ctx)
	require.Error(t, err)

	cfg.AddFilter("error", func(input interface{}) (string, error) { return "", errors.New("test error") })
	_, err = EvaluateString("1 | error", ctx)
	require.Error(t, err)
}

func TestClosure(t *testing.T) {
	cfg := NewConfig()
	ctx := NewContext(map[string]interface{}{"x": 1}, cfg)
	expr, err := Parse("x")
	require.NoError(t, err)
	c1 := closure{expr, ctx}
	c2 := c1.Bind("x", 2)
	x1, err := c1.Evaluate()
	require.NoError(t, err)
	x2, err := c2.Evaluate()
	require.NoError(t, err)
	require.Equal(t, 1, x1)
	require.Equal(t, 2, x2)
}
