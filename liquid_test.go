package liquid

import (
	"fmt"
	"log"
	"testing"

	"github.com/osteele/liquid/tags"
	"github.com/stretchr/testify/require"
)

func init() {
	tags.DefineStandardTags()
}

var liquidTests = []struct{ in, expected string }{
	{"{{page.title}}", "Introduction"},
	{"{%if x%}true{%endif%}", "true"},
}

var liquidTestScope = map[string]interface{}{
	"x":  123,
	"ar": []string{"first", "second", "third"},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
}

func TestChunkParser(t *testing.T) {
	engine := NewEngine()
	for i, test := range liquidTests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			out, err := engine.ParseAndRenderString(test.in, liquidTestScope)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, out, test.in)
		})
	}
}

func Example() {
	engine := NewEngine()
	template := `<h1>{{page.title}}</h1>`
	scope := map[string]interface{}{
		"page": map[string]interface{}{
			"title": "Introduction",
		},
	}
	out, err := engine.ParseAndRenderString(template, scope)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: <h1>Introduction</h1>
}
