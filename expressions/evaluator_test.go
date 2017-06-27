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
	{"12", 12},
	{"12.3", 12.3},
	{"true", true},
	{"false", false},
	{`'abc'`, "abc"},
	{`"abc"`, "abc"},

	// Variables
	{"n", 123},
	{"obj.a", "first"},
	{"obj.b.c", "d"},
	{"obj.x", nil},
	{"ar[1]", "second"},
	{"ar[-1]", nil},
	{"ar[100]", nil},
	{"obj[1]", nil},
	{"obj.c[0]", "r"},
	// {`fruits.first`, "apples"},
	// {`fruits.last`, "plums"},
	// {`empty_list.first`, ""},
	// {`empty_list.last`, ""},

	// Operators
	{"1 == 1", true},
	{"1 == 2", false},
	{"1 < 2", true},
	{"2 < 1", false},
	{"1 > 2", false},
	{"2 > 1", true},

	{"1.0 == 1.0", true},
	{"1.0 == 2.0", false},
	{"1.0 < 2.0", true},

	{"1.0 == 1", true},
	{"1 == 1.0", true},
	{"1 < 2.0", true},
	{"1.0 < 2", true},
}

var evaluatorTestContext = NewContext(map[string]interface{}{
	"n":          123,
	"ar":         []string{"first", "second", "third"},
	"empty_list": map[string]interface{}{},
	"fruits":     []string{"apples", "oranges", "peaches", "plums"},
	"obj": map[string]interface{}{
		"a": "first",
		"b": map[string]interface{}{"c": "d"},
		"c": []string{"r", "g", "b"},
	},
})

func TestEvaluator(t *testing.T) {
	for i, test := range evaluatorTests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			val, err := EvaluateExpr(test.in, evaluatorTestContext)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, val, test.in)
		})
	}
}
