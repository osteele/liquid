package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	oldArgs := os.Args

	defer func() {
		os.Args = oldArgs
		stderr = os.Stderr
		stdout = os.Stdout
		stdin = os.Stdin
		exit = os.Exit
		env = os.Environ
		bindings = map[string]any{}
	}()

	exit = func(n int) {
		t.Fatalf("exit called")
	}

	os.Args = []string{"liquid"}

	// stdin
	src := `{{ "Hello World" | downcase | split: " " | first | append: "!"}}`
	buf := &bytes.Buffer{}
	stdin = bytes.NewBufferString(src)
	stdout = buf

	main()
	require.Equal(t, "hello!", buf.String())

	// environment binding
	var envCalled bool

	env = func() []string {
		envCalled = true
		return []string{"TARGET=World"}
	}
	src = `Hello, {{ TARGET }}!`
	// without -e
	stdin = bytes.NewBufferString(src)
	buf = &bytes.Buffer{}
	stdout = buf
	os.Args = []string{"liquid"}

	main()
	require.False(t, envCalled)
	require.Equal(t, "Hello, !", buf.String())
	// with -e
	stdin = bytes.NewBufferString(src)
	buf = &bytes.Buffer{}
	stdout = buf
	os.Args = []string{"liquid", "--env"}

	main()
	require.True(t, envCalled)
	require.Equal(t, "Hello, World!", buf.String())

	bindings = make(map[string]any)

	// filename
	stdin = os.Stdin
	buf = &bytes.Buffer{}
	stdout = buf
	os.Args = []string{"liquid", "testdata/source.liquid"}

	main()
	require.Contains(t, buf.String(), "file system")

	// following tests test the exit code
	var exitCalled bool

	exitCode := 0
	exit = func(n int) { exitCalled = true; exitCode = n }

	// strict variables
	stdin = bytes.NewBufferString(src)
	buf = &bytes.Buffer{}
	stderr = buf
	os.Args = []string{"liquid", "--strict"}

	main()
	require.True(t, exitCalled)
	require.Equal(t, 1, exitCode)
	require.Equal(t, "Liquid error: undefined variable in {{ TARGET }}\n", buf.String())

	exitCode = 0
	os.Args = []string{"liquid", "testdata/source.liquid"}

	main()
	require.Equal(t, 0, exitCode)

	exitCode = 0
	// missing file
	buf = &bytes.Buffer{}
	stderr = buf
	os.Args = []string{"liquid", "testdata/missing_file"}

	main()
	require.Equal(t, 1, exitCode)
	// Darwin/Linux, and Windows, have different error messages
	require.Regexp(t, "no such file|cannot find the file", buf.String())

	exitCalled = false
	// --help
	buf = &bytes.Buffer{}
	stderr = buf
	os.Args = []string{"liquid", "--help"}

	main()
	require.Contains(t, buf.String(), "usage:")
	require.True(t, exitCalled)
	require.Equal(t, 0, exitCode)

	// --undefined-flag
	buf = &bytes.Buffer{}
	stderr = buf
	os.Args = []string{"liquid", "--undefined-flag"}

	main()
	require.Equal(t, 1, exitCode)
	require.Contains(t, buf.String(), "defined")

	// multiple args
	os.Args = []string{"liquid", "testdata/source.liquid", "file2"}
	buf = &bytes.Buffer{}
	stderr = buf

	main()
	require.Contains(t, buf.String(), "too many")
	require.Equal(t, 1, exitCode)
}
