package render

import (
	"bytes"
	"testing"

	"github.com/osteele/liquid/parser"
	"github.com/stretchr/testify/require"
)

func TestRenderEscapeFilter(t *testing.T) {
	cfg := NewConfig()
	cfg.SetAutoEscapeReplacer(HtmlEscaper)
	buf := new(bytes.Buffer)

	f := func(t *testing.T, tmpl string, bindings map[string]interface{}, out string) {
		t.Helper()
		buf.Reset()
		root, err := cfg.Compile(tmpl, parser.SourceLoc{})
		require.NoError(t, err)
		err = Render(root, buf, bindings, cfg)
		require.NoError(t, err)
		require.Equal(t, out, buf.String())
	}

	t.Run("unsafe", func(t *testing.T) {
		f(t,
			`{{ input }}`,
			map[string]interface{}{
				"input": "<script>doEvilStuff()</script>",
			},
			"&lt;script&gt;doEvilStuff()&lt;/script&gt;",
		)
	})

	t.Run("safe", func(t *testing.T) {
		f(t,
			`{{ input|safe }}`,
			map[string]interface{}{
				"input": "<script>doGoodStuff()</script>",
			},
			"<script>doGoodStuff()</script>",
		)
	})

	t.Run("double safe", func(t *testing.T) {
		f(t,
			`{{ input|safe|safe }}`,
			map[string]interface{}{
				"input": "<script>doGoodStuff()</script>",
			},
			"<script>doGoodStuff()</script>",
		)
	})

	t.Run("unsafe slice result", func(t *testing.T) {
		f(t,
			`{{ input }}`,
			map[string]interface{}{
				"input": []interface{}{"<a>", "<b>"},
			},
			"&lt;a&gt;&lt;b&gt;",
		)
	})

	t.Run("safe slice result", func(t *testing.T) {
		f(t,
			`{{ input|safe }}`,
			map[string]interface{}{
				"input": []interface{}{"<a>", "<b>"},
			},
			"<a><b>",
		)
	})
}
