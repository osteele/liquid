package liquid

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

var emptyBindings = map[string]interface{}{}

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

func TestEngine_ParseAndRenderString(t *testing.T) {
	engine := NewEngine()
	for i, test := range liquidTests {
		t.Run(fmt.Sprint(i+1), func(t *testing.T) {
			out, err := engine.ParseAndRenderString(test.in, testBindings)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, out, test.in)
		})
	}
}

func Example() {
	engine := NewEngine()
	source := `<h1>{{ page.title }}</h1>`
	bindings := map[string]interface{}{
		"page": map[string]string{
			"title": "Introduction",
		},
	}
	out, err := engine.ParseAndRenderString(source, bindings)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: <h1>Introduction</h1>
}

func ExampleEngine_ParseAndRenderString() {
	engine := NewEngine()
	source := `{{ hello | capitalize | append: " Mundo" }}`
	bindings := map[string]interface{}{"hello": "hola"}
	out, err := engine.ParseAndRenderString(source, bindings)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: Hola Mundo
}
func ExampleEngine_ParseTemplate() {
	engine := NewEngine()
	source := `{{ hello | capitalize | append: " Mundo" }}`
	bindings := map[string]interface{}{"hello": "hola"}
	tpl, err := engine.ParseTemplate([]byte(source))
	if err != nil {
		log.Fatalln(err)
	}
	out, err := tpl.RenderString(bindings)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: Hola Mundo
}
func ExampleEngine_RegisterFilter() {
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
func ExampleEngine_RegisterFilter_optional_argument() {
	engine := NewEngine()
	// func(a, b int) int) would default the second argument to zero.
	// Then we can't tell the difference between {{ n | inc }} and
	// {{ n | inc: 0 }}. A function in the parameter list has a special
	// meaning as a default parameter.
	engine.RegisterFilter("inc", func(a int, b func(int) int) int {
		return a + b(1)
	})
	template := `10 + 1 = {{ m | inc }}; 20 + 5 = {{ n | inc: 5 }}`
	bindings := map[string]interface{}{
		"m": 10,
		"n": "20",
	}
	out, err := engine.ParseAndRenderString(template, bindings)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: 10 + 1 = 11; 20 + 5 = 25
}

func ExampleEngine_RegisterTag() {
	engine := NewEngine()
	engine.RegisterTag("echo", func(c render.Context) (string, error) {
		return c.TagArgs(), nil
	})
	template := `{% echo hello world %}`
	out, err := engine.ParseAndRenderString(template, emptyBindings)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: hello world
}

func ExampleEngine_RegisterBlock() {
	engine := NewEngine()
	engine.RegisterBlock("length", func(c render.Context) (string, error) {
		s, err := c.InnerString()
		if err != nil {
			return "", err
		}
		n := len(s)
		return fmt.Sprint(n), nil
	})

	template := `{% length %}abc{% endlength %}`
	out, err := engine.ParseAndRenderString(template, emptyBindings)
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
