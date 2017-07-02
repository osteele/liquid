package liquid

import (
	"fmt"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/osteele/liquid/chunks"
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

func TestTemplateRenderString(t *testing.T) {
	engine := NewEngine()
	template, err := engine.ParseTemplate([]byte(`{{ "hello world" | capitalize }}`))
	require.NoError(t, err)
	out, err := template.RenderString(testContext)
	require.NoError(t, err)
	require.Equal(t, "Hello world", out)
}

func Example() {
	engine := NewEngine()
	template := `<h1>{{ page.title }}</h1>`
	bindings := map[string]interface{}{
		"page": map[string]string{
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

func Example_filter() {
	engine := NewEngine()
	engine.DefineFilter("has_prefix", strings.HasPrefix)
	template := `{{ title | has_prefix: "Intro" }}`

	bindings := map[string]interface{}{
		"title": "Introduction",
	}
	out, err := engine.ParseAndRenderString(template, NewContext(bindings))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: true
}

func Example_tag() {
	engine := NewEngine()
	engine.DefineTag("echo", func(w io.Writer, c chunks.RenderContext) error {
		args := c.TagArgs()
		_, err := w.Write([]byte(args))
		return err
	})

	template := `{% echo hello world %}`
	bindings := map[string]interface{}{}
	out, err := engine.ParseAndRenderString(template, NewContext(bindings))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: hello world
}

func Example_tag_pair() {
	engine := NewEngine()
	engine.DefineStartTag("length", func(w io.Writer, c chunks.RenderContext) error {
		s, err := c.InnerString()
		if err != nil {
			return err
		}
		n := len(s)
		_, err = w.Write([]byte(fmt.Sprint(n)))
		return err
	})

	template := `{% length %}abc{% endlength %}`
	bindings := map[string]interface{}{}
	out, err := engine.ParseAndRenderString(template, NewContext(bindings))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: 3
}
