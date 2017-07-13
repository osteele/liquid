package tags

import (
	"bytes"
	"fmt"
	"testing"

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

func TestControlFlowTags(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(config)
	for i, test := range cfTagTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			ast, err := config.Compile(test.in)
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = render.Render(ast, buf, tagTestBindings, config)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}
