package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	exit = func(n int) { t.Fatalf("exit called") }

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

	buf = new(bytes.Buffer)
	stdout = buf
	require.NoError(t, run([]string{"--help"}))
	require.Contains(t, buf.String(), "usage:")

	exitCode := 0
	exit = func(n int) { exitCode = n }
	require.NoError(t, run([]string{"--unknown-flag"}))
	require.Equal(t, 1, exitCode)
}
