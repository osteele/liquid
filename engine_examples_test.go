package liquid

import (
	"fmt"
	"log"
	"strings"

	"github.com/urbn8/liquid/render"
)

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
	tpl, err := engine.ParseString(source)
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
