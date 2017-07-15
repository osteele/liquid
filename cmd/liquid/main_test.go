package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	src := `{{ "Hello World" | downcase | split: " " | first | append: "!"}}`
	buf := new(bytes.Buffer)
	stdin = bytes.NewBufferString(src)
	stdout = buf
	require.NoError(t, run([]string{}))
	require.Equal(t, "hello!", buf.String())

	buf = new(bytes.Buffer)
	stdin = bytes.NewBufferString("")
	stdout = buf
	require.NoError(t, run([]string{"testdata/source.txt"}))
	require.Contains(t, buf.String(), "file system")
}
