package expressions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var evaluatorTestContext = Context{map[string]interface{}{
	"n":  123,
	"ar": []string{"first", "second", "third"},
	"obj": map[string]interface{}{
		"a": "first",
		"b": map[string]interface{}{"c": "d"},
		"c": []string{"r", "g", "b"},
	},
},
}

var evaluatorTests = []struct {
	in       string
	expected interface{}
}{
	// Constants
	{"12", 12},
	{"12.3", 12.3},
	{"true", true},
	{"false", false},

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

func TestEvaluator(t *testing.T) {
	for i, test := range evaluatorTests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			val, err := EvaluateExpr(test.in, evaluatorTestContext)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, val, test.in)
		})
	}
}
