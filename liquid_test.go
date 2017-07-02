package liquid

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

// There's a lot more tests in the filters and tags sub-packages.
// This collects a minimal set for testing end-to-end.
var liquidTests = []struct{ in, expected string }{
	{`{{ page.title }}`, "Introduction"},
	{`{% if x %}true{% endif %}`, "true"},
	{`{{ "upper" | upcase }}`, "UPPER"},
}

var testContext = NewContext(map[string]interface{}{
	"x":  123,
	"ar": []string{"first", "second", "third"},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
})

func TestLiquid(t *testing.T) {
	engine := NewEngine()
	for i, test := range liquidTests {
		t.Run(fmt.Sprint(i+1), func(t *testing.T) {
			out, err := engine.ParseAndRenderString(test.in, testContext)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, out, test.in)
		})
	}
}

func Example() {
	engine := NewEngine()
	template := `<h1>{{page.title}}</h1>`
	bindings := map[string]interface{}{
		"page": map[string]interface{}{
			"title": "Introduction",
		},
	}
	context := NewContext(bindings)
	out, err := engine.ParseAndRenderString(template, context)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: <h1>Introduction</h1>
}
