package liquid

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

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
			require.Equalf(t, test.expected, string(out), test.in)
		})
	}
}
