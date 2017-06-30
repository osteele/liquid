package chunks

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var renderTests = []struct{ in, expected string }{
	// {"{%if syntax error%}{%endif%}", "parse error"},
	{`{{ 12 }}`, "12"},
	{`{{ x }}`, "123"},
	{`{{ page.title }}`, "Introduction"},
	{`{{ ar[1] }}`, "second"},
}

var renderTestBindings = map[string]interface{}{
	"x": 123,
	"obj": map[string]interface{}{
		"a": 1,
	},
	"animals": []string{"zebra", "octopus", "giraffe", "Sally Snake"},
	"pages": []map[string]interface{}{
		{"category": "business"},
		{"category": "celebrities"},
		{},
		{"category": "lifestyle"},
		{"category": "sports"},
		{},
		{"category": "technology"},
	},
	"sort_prop": []map[string]interface{}{
		{"weight": 1},
		{"weight": 5},
		{"weight": 3},
		{"weight": nil},
	},
	"ar": []string{"first", "second", "third"},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
}

func TestRender(t *testing.T) {
	settings := NewSettings()
	context := NewContext(renderTestBindings, settings)
	for i, test := range renderTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			ast, err := settings.Parse(test.in)
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = ast.Render(buf, context)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}
