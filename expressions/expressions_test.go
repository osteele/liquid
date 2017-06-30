package expressions

import (
	"fmt"
	"testing"

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
	{`obj.a`, "first"},
	{`obj.b.c`, "d"},
	{`obj.x`, nil},
	{`fruits.first`, "apples"},
	{`fruits.last`, "plums"},
	{`empty_list.first`, nil},
	{`empty_list.last`, nil},
	{`"abc".size`, 3},
	{`fruits.size`, 4},

	// Indices
	{`array[1]`, "second"},
	{`array[-1]`, "third"}, // undocumented
	{`array[100]`, nil},
	{`obj[1]`, nil},
	{`obj.c[0]`, "r"},

	// Expressions
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
}

var evaluatorTestContext = NewContext(map[string]interface{}{
	"n":          123,
	"array":      []string{"first", "second", "third"},
	"empty_list": []interface{}{},
	"fruits":     []string{"apples", "oranges", "peaches", "plums"},
	"obj": map[string]interface{}{
		"a": "first",
		"b": map[string]interface{}{"c": "d"},
		"c": []string{"r", "g", "b"},
	},
}, NewSettings())

func TestEvaluator(t *testing.T) {
	for i, test := range evaluatorTests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			val, err := EvaluateString(test.in, evaluatorTestContext)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, val, test.in)
		})
	}
}
