package render

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var scannerCountTests = []struct {
	in  string
	len int
}{
	{`{% tag arg %}`, 1},
	{`{% tag arg %}{% tag %}`, 2},
	{`{% tag arg %}{% tag arg %}{% tag %}`, 3},
	{`{% tag %}{% tag %}`, 2},
	{`{% tag arg %}{% tag arg %}{% tag %}{% tag %}`, 4},
	{`{{ expr }}`, 1},
	{`{{ expr arg }}`, 1},
	{`{{ expr }}{{ expr }}`, 2},
	{`{{ expr arg }}{{ expr arg }}`, 2},
}

func TestChunkScanner(t *testing.T) {
	tokens := Scan("12", "")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, TextChunkType, tokens[0].Type)
	require.Equal(t, "12", tokens[0].Source)

	tokens = Scan("{{obj}}", "")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, ObjChunkType, tokens[0].Type)
	require.Equal(t, "obj", tokens[0].Args)

	tokens = Scan("{{ obj }}", "")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, ObjChunkType, tokens[0].Type)
	require.Equal(t, "obj", tokens[0].Args)

	tokens = Scan("{%tag args%}", "")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, TagChunkType, tokens[0].Type)
	require.Equal(t, "tag", tokens[0].Name)
	require.Equal(t, "args", tokens[0].Args)

	tokens = Scan("{% tag args %}", "")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, TagChunkType, tokens[0].Type)
	require.Equal(t, "tag", tokens[0].Name)
	require.Equal(t, "args", tokens[0].Args)

	tokens = Scan("pre{% tag args %}mid{{ object }}post", "")
	require.Equal(t, `[TextChunkType{"pre"} TagChunkType{Tag:"tag", Args:"args"} TextChunkType{"mid"} ObjChunkType{"object"} TextChunkType{"post"}]`, fmt.Sprint(tokens))
	require.Equal(t, "- text: pre\n- args: args\n  tag: tag\n- text: mid\n- obj: object\n- text: post\n", MustYAML(tokens))

	for i, test := range scannerCountTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := Scan(test.in, "")
			require.Len(t, tokens, test.len)
		})
	}
}
