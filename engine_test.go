package liquid

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

// There's a lot more tests in the filters and tags sub-packages.
// This collects a minimal set for testing end-to-end.
var liquidTests = []struct{ in, expected string }{
	{`{{ page.title }}`, "Introduction"},
	{`{% if x %}true{% endif %}`, "true"},
	{`{{ "upper" | upcase }}`, "UPPER"},
}

var testBindings = map[string]interface{}{
	"x":  123,
	"ar": []string{"first", "second", "third"},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
}

func TestLiquid(t *testing.T) {
	engine := NewEngine()
	for i, test := range liquidTests {
		t.Run(fmt.Sprint(i+1), func(t *testing.T) {
			out, err := engine.ParseAndRenderString(test.in, testBindings)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, out, test.in)
		})
	}
}

func TestTemplateRenderString(t *testing.T) {
	engine := NewEngine()
	template, err := engine.ParseTemplate([]byte(`{{ "hello world" | capitalize }}`))
	require.NoError(t, err)
	out, err := template.RenderString(testBindings)
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
	out, err := engine.ParseAndRenderString(template, bindings)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: <h1>Introduction</h1>
}

func Example_register_filter() {
	engine := NewEngine()
	engine.RegisterFilter("has_prefix", strings.HasPrefix)
	template := `{{ title | has_prefix: "Intro" }}`

	bindings := map[string]interface{}{
		"title": "Introduction",
	}
	out, err := engine.ParseAndRenderString(template, bindings)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: true
}

func Example_register_tag() {
	engine := NewEngine()
	engine.RegisterTag("echo", func(c render.RenderContext) (string, error) {
		return c.TagArgs(), nil
	})

	template := `{% echo hello world %}`
	bindings := map[string]interface{}{}
	out, err := engine.ParseAndRenderString(template, bindings)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: hello world
}

func Example_register_block() {
	engine := NewEngine()
	engine.RegisterBlock("length", func(c render.RenderContext) (string, error) {
		s, err := c.InnerString()
		if err != nil {
			return "", err
		}
		n := len(s)
		return fmt.Sprint(n), nil
	})

	template := `{% length %}abc{% endlength %}`
	bindings := map[string]interface{}{}
	out, err := engine.ParseAndRenderString(template, bindings)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: 3
}

type redConvertible struct{}

func (c redConvertible) ToLiquid() interface{} {
	return "red"
}

func ExampleDrop() {
	// type redConvertible struct{}
	//
	// func (c redConvertible) ToLiquid() interface{} {
	// 	return "red"
	// }
	engine := NewEngine()
	bindings := map[string]interface{}{
		"drop": redConvertible{},
	}
	template := `{{ drop }}`
	out, err := engine.ParseAndRenderString(template, bindings)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: red
}
