package tags

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

var cfTagTests = []struct{ in, expected string }{
	// case
	{`{% case 1 %}{% when 1 %}a{% when 2 %}b{% endcase %}`, "a"},
	{`{% case 2 %}{% when 1 %}a{% when 2 %}b{% endcase %}`, "b"},
	{`{% case 3 %}{% when 1 %}a{% when 2 %}b{% endcase %}`, ""},
	// else
	{`{% case 1 %}{% when 1 %}a{% else %}b{% endcase %}`, "a"},
	{`{% case 2 %}{% when 1 %}a{% else %}b{% endcase %}`, "b"},
	// disjunction
	{`{% case 1 %}{% when 1,2 %}a{% else %}b{% endcase %}`, "a"},
	{`{% case 2 %}{% when 1,2 %}a{% else %}b{% endcase %}`, "a"},
	{`{% case 3 %}{% when 1,2 %}a{% else %}b{% endcase %}`, "b"},

	// if
	{`{% if true %}true{% endif %}`, "true"},
	{`{% if false %}false{% endif %}`, ""},
	{`{% if 0 %}true{% endif %}`, "true"},
	{`{% if 1 %}true{% endif %}`, "true"},
	{`{% if x %}true{% endif %}`, "true"},
	{`{% if y %}true{% endif %}`, ""},
	{`{% if true %}true{% endif %}`, "true"},
	{`{% if false %}false{% endif %}`, ""},
	{`{% if true %}true{% else %}false{% endif %}`, "true"},
	{`{% if false %}false{% else %}true{% endif %}`, "true"},
	{`{% if true %}0{% elsif true %}1{% else %}2{% endif %}`, "0"},
	{`{% if false %}0{% elsif true %}1{% else %}2{% endif %}`, "1"},
	{`{% if false %}0{% elsif false %}1{% else %}2{% endif %}`, "2"},

	// unless
	{`{% unless true %}false{% endunless %}`, ""},
	{`{% unless false %}true{% endunless %}`, "true"},
}

var cfTagCompilationErrorTests = []struct{ in, expected string }{
	{`{% if syntax error %}{% endif %}`, "syntax error"},
	{`{% if true %}{% elsif syntax error %}{% endif %}`, "syntax error"},
	{`{% case syntax error %}{% when 1 %}{% endcase %}`, "syntax error"},
}

var cfTagErrorTests = []struct{ in, expected string }{
	{`{% if a | undefined_filter %}{% endif %}`, "undefined filter"},
	{`{% if false %}{% elsif a | undefined_filter %}{% endif %}`, "undefined filter"},
	{`{% case 1 %}{% when 1 %}{% error %}{% endcase %}`, "tag render error"},
	{`{% case a | undefined_filter %}{% when 1 %}{% endcase %}`, "undefined filter"},
}

var cfTagWhitespaceTests = []struct{ in, expected string }{
	{`  {%- if true %}	trims outside	{% endif -%}  `, "	trims outside	"},
	{`  ({% if true -%}	trims inside	{%- endif %})  `, "  (trims inside)  "},
	{`(  {%- if true -%}	trims both	{%- endif -%}  )`, "(trims both)"},
	{`removes
{%- if true -%}
block
{%- endif -%}
lines`,
		`removes
block
lines`},
	{`removes
{%- if true -%}
block
{%- else -%}
not rendered
{%- endif -%}
lines`,
		`removes
block
lines`},
	{`removes
{%- case 1 -%}
{%- when 1 -%}
block
{%- when 2 -%}
not rendered
{%- endcase -%}
lines`,
		`removes
block
lines`},
}

func TestControlFlowTags(t *testing.T) {
	cfg := render.NewConfig()
	AddStandardTags(cfg)
	for i, test := range cfTagTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := cfg.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = render.Render(root, buf, tagTestBindings, cfg)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}

func TestControlFlowTagsWithWhitespace(t *testing.T) {
	cfg := render.NewConfig()
	AddStandardTags(cfg)
	for i, test := range cfTagWhitespaceTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := cfg.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = render.Render(root, buf, tagTestBindings, cfg)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}

func TestControlFlowTags_errors(t *testing.T) {
	cfg := render.NewConfig()
	AddStandardTags(cfg)
	cfg.AddTag("error", func(string) (func(io.Writer, render.Context) error, error) {
		return func(io.Writer, render.Context) error {
			return fmt.Errorf("tag render error")
		}, nil
	})

	for i, test := range cfTagCompilationErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			_, err := cfg.Compile(test.in, parser.SourceLoc{})
			require.Errorf(t, err, test.in)
			require.Contains(t, err.Error(), test.expected, test.in)
		})
	}
	for i, test := range cfTagErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := cfg.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			err = render.Render(root, ioutil.Discard, tagTestBindings, cfg)
			require.Errorf(t, err, test.in)
			require.Contains(t, err.Error(), test.expected, test.in)
		})
	}
}
