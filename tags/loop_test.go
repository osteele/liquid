package tags

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

var loopTests = []struct{ in, expected string }{
	{`{% for a in array %}{{ a }} {% endfor %}`, "first second third "},

	// loop modifiers
	{`{% for a in array reversed %}{{ a }}.{% endfor %}`, "third.second.first."},
	{`{% for a in array limit:2 %}{{ a }}.{% endfor %}`, "first.second."},
	{`{% for a in array offset:1 %}{{ a }}.{% endfor %}`, "second.third."},
	{`{% for a in array reversed limit:1 %}{{ a }}.{% endfor %}`, "third."},
	// TODO investigate how these combine; does it depend on the order
	// {`{% for a in array reversed offset:1 %}{{ a }}.{% endfor %}`, "second.first."},
	// {`{% for a in array limit:1 offset:1 %}{{ a }}.{% endfor %}`, "second."},
	// {`{% for a in array reversed limit:1 offset:1 %}{{ a }}.{% endfor %}`, "second."},

	// loop variables
	{`{% for a in array %}{{ forloop.first }}.{% endfor %}`, "true.false.false."},
	{`{% for a in array %}{{ forloop.last }}.{% endfor %}`, "false.false.true."},
	{`{% for a in array %}{{ forloop.index }}.{% endfor %}`, "1.2.3."},
	{`{% for a in array %}{{ forloop.index0 }}.{% endfor %}`, "0.1.2."},
	{`{% for a in array %}{{ forloop.rindex }}.{% endfor %}`, "3.2.1."},
	{`{% for a in array %}{{ forloop.rindex0 }}.{% endfor %}`, "2.1.0."},
	{`{% for a in array %}{{ forloop.length }}.{% endfor %}`, "3.3.3."},

	{`{% for i in array %}{{ forloop.index }}[{% for j in array %}{{ forloop.index }}{% endfor %}]{{ forloop.index }}{% endfor %}`,
		"1[123]12[123]23[123]3"},

	{`{% for a in array reversed %}{{ forloop.first }}.{% endfor %}`, "true.false.false."},
	{`{% for a in array reversed %}{{ forloop.last }}.{% endfor %}`, "false.false.true."},
	{`{% for a in array reversed %}{{ forloop.index }}.{% endfor %}`, "1.2.3."},
	{`{% for a in array reversed %}{{ forloop.rindex }}.{% endfor %}`, "3.2.1."},
	{`{% for a in array reversed %}{{ forloop.length }}.{% endfor %}`, "3.3.3."},

	{`{% for a in array limit:2 %}{{ forloop.index }}.{% endfor %}`, "1.2."},
	{`{% for a in array limit:2 %}{{ forloop.rindex }}.{% endfor %}`, "2.1."},
	{`{% for a in array limit:2 %}{{ forloop.first }}.{% endfor %}`, "true.false."},
	{`{% for a in array limit:2 %}{{ forloop.last }}.{% endfor %}`, "false.true."},
	{`{% for a in array limit:2 %}{{ forloop.length }}.{% endfor %}`, "2.2."},

	{`{% for a in array offset:1 %}{{ forloop.index }}.{% endfor %}`, "1.2."},
	{`{% for a in array offset:1 %}{{ forloop.rindex }}.{% endfor %}`, "2.1."},
	{`{% for a in array offset:1 %}{{ forloop.first }}.{% endfor %}`, "true.false."},
	{`{% for a in array offset:1 %}{{ forloop.last }}.{% endfor %}`, "false.true."},
	{`{% for a in array offset:1 %}{{ forloop.length }}.{% endfor %}`, "2.2."},

	{`{% for a in array %}{% if a == 'second' %}{% break %}{% endif %}{{ a }}{% endfor %}`, "first"},
	{`{% for a in array %}{% if a == 'second' %}{% continue %}{% endif %}{{ a }}.{% endfor %}`, "first.third."},

	{`{% for a in hash %}{{ a }}{% endfor %}`, "a"},
}

var loopTestBindings = map[string]interface{}{
	"array": []string{"first", "second", "third"},
	"hash":  map[string]interface{}{"a": 1},
}

func TestLoopTag(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(config)
	for i, test := range loopTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			ast, err := config.Parse(test.in)
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = render.Render(ast, buf, loopTestBindings, config)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}
