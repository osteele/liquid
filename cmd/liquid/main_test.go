package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	src := `{{ "Hello World" | downcase | split: " " | first | append: "!"}}`
	stdin = bytes.NewBufferString(src)
	buf := new(bytes.Buffer)
	stdout = buf
	require.NoError(t, run([]string{}))
	require.Equal(t, "hello!", buf.String())
}
