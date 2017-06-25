package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChunkParser(t *testing.T) {
	ctx := Context{map[string]interface{}{
		"x": 123,
	},
	}

	for _, test := range chunkTests {
		tokens := ScanChunks(test.in, "")
		// fmt.Println(tokens)
		ast, err := Parse(tokens)
		require.NoError(t, err, test.in)
		// fmt.Println("%#v", ast)
		buf := new(bytes.Buffer)
		err = ast.Render(buf, ctx)
		require.NoError(t, err, test.in)
		require.Equal(t, test.expected, buf.String(), test.in)
	}
}

var chunkTests = []struct{ in, expected string }{
	{"{{12}}", "12"},
	{"{{x}}", "123"},
	{"{%if 1%}true{%endif%}", "true"},
	{"{%if x%}true{%endif%}", "true"},
	{"{%if y%}false{%endif%}", ""},
	{"{%if 1%}true{%else%}false{%endif%}", "true"},
	{"{%if y%}false{%else%}true{%endif%}", "true"},
	{"{%if 1%}0{%elsif 1%}1{%else%}2{%endif%}", "0"},
	{"{%if y%}0{%elsif 1%}1{%else%}2{%endif%}", "1"},
	{"{%if y%}0{%elsif z%}1{%else%}2{%endif%}", "2"},
}
