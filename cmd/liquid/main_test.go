package main

import (
	"bytes"
	"testing"

	"github.com/autopilot3/liquid"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	exit = func(n int) { t.Fatalf("exit called") }

	// stdin
	src := `{{ "Hello World" | downcase | split: " " | first | append: "!"}}`
	buf := new(bytes.Buffer)
	stdin = bytes.NewBufferString(src)
	stdout = buf
	require.NoError(t, run([]string{}))
	require.Equal(t, "hello!", buf.String())

	// filename
	buf = new(bytes.Buffer)
	stdin = bytes.NewBufferString("")
	stdout = buf
	require.NoError(t, run([]string{"testdata/source.txt"}))
	require.Contains(t, buf.String(), "file system")

	// missing file
	require.Error(t, run([]string{"testdata/missing_file"}))

	// --help
	buf = new(bytes.Buffer)
	stdout = buf
	require.NoError(t, run([]string{"--help"}))
	require.Contains(t, buf.String(), "usage:")

	// --undefined-flag
	exitCode := 0
	exit = func(n int) { exitCode = n }
	require.NoError(t, run([]string{"--undefined-flag"}))
	require.Equal(t, 1, exitCode)

	// multiple args
	exitCode = 0
	exit = func(n int) { exitCode = n }
	require.NoError(t, run([]string{"file1", "file2"}))
	require.Equal(t, 1, exitCode)
}

func TestRenderAllowedTags(t *testing.T) {

	bindings := map[string]interface{}{
		"people": map[string]interface{}{
			"name":  "bob",
			"email": "bob@example.com",
		},
	}
	engine := liquid.NewEngine()
	engine.SetAllowedTags(map[string]struct{}{
		"people.name": {},
	})
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{
			"Allow name only",
			"Hello {{ people.name | default: 'there' }}, your email is {{ people.email }}!",
			"Hello bob, your email is {{ people.email }}!",
		},
		{
			"Allow name only",
			"Hello {{ people.name | default: 'there' }}, your email is {{ people.email | default: 'unknown' }}!",
			"Hello bob, your email is {{ people.email | default: 'unknown' }}!",
		},
	}
	for _, tt := range tests {
		tmpl, err := engine.ParseString(tt.src)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(tt.name, func(t *testing.T) {
			out, err := tmpl.RenderString(bindings)
			if err != nil {
				t.Fatal(err)
			}
			if out != tt.expected {
				t.Errorf("TestRenderAllowedTags() = %v, want %v", out, tt.expected)
			}
		})
	}
}
