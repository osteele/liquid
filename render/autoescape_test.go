package render

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osteele/liquid/parser"
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

// TestReplacerWriterIOContract verifies that replacerWriter.Write correctly
// implements the io.Writer contract by returning len(p)
func TestReplacerWriterIOContract(t *testing.T) {
	buf := new(bytes.Buffer)
	rw := &replacerWriter{
		replacer: HtmlEscaper,
		w:        buf,
	}

	// Test with input that gets escaped (different output length)
	input := []byte("<script>")
	n, err := rw.Write(input)
	require.NoError(t, err)
	require.Equal(t, len(input), n, "Write must return len(p) per io.Writer contract")
	require.Equal(t, "&lt;script&gt;", buf.String(), "Content should be escaped")

	// Test with normal input (same output length)
	buf.Reset()
	input2 := []byte("hello world")
	n2, err2 := rw.Write(input2)
	require.NoError(t, err2)
	require.Equal(t, len(input2), n2, "Write must return len(p) for normal input")
	require.Equal(t, "hello world", buf.String())

	// Test with empty input
	buf.Reset()
	input3 := []byte("")
	n3, err3 := rw.Write(input3)
	require.NoError(t, err3)
	require.Equal(t, 0, n3, "Write must return 0 for empty input")
	require.Equal(t, "", buf.String())
}
