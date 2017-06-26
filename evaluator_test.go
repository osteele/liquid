package main

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
	{"12", 12},
	{"n", 123},
	{"obj.a", "first"},
	{"obj.b.c", "d"},
	{"obj.x", nil},
	{"ar[1]", "second"},
	{"ar[-1]", nil},
	{"ar[100]", nil},
	{"obj[1]", nil},
	{"obj.c[0]", "r"},
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
