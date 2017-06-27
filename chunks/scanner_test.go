package chunks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
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
	require.Equal(t, "tag", tokens[0].Tag)
	require.Equal(t, "args", tokens[0].Args)

	tokens = Scan("{% tag args %}", "")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, TagChunkType, tokens[0].Type)
	require.Equal(t, "tag", tokens[0].Tag)
	require.Equal(t, "args", tokens[0].Args)
}
