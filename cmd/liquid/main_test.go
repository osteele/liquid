package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/autopilot3/ap3-types-go/types/date"
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
			"name": "bob",
		},
	}
	tests := []struct {
		name                 string
		allowTagsWithDefault bool
		src                  string
		expected             string
	}{
		{
			"Allow name only",
			false,
			"Hello {{ people.name | default: 'there' }}, your email is {{ people.email }}! {% if people.random == '123' %} you can't see me {% endif %}",
			"Hello bob, your email is {{ people.email }}! {% if people.random == '123' %} you can't see me {% endif %}",
		},
		{
			"Allow name only, others have default",
			false,
			"Hello {{ people.name | default: 'there' }}, your email is {{ people.email | default: 'unknown' }}!",
			"Hello bob, your email is {{ people.email | default: 'unknown' }}!",
		},
		{
			"Allow name and default",
			true,
			"Hello {{ people.name | default: 'there' }}, your email is {{ people.email | default: 'unknown' }}!",
			"Hello bob, your email is unknown!",
		},
		{
			"Allow name and default",
			true,
			"Hello {{ people.name | default: 'there' }}, your email is {{ people.email | default: 'unknown' }}!{% if people.random == '123' %} you can't see me.{% endif %}",
			"Hello bob, your email is unknown!{% if people.random == '123' %} you can't see me.{% endif %}",
		},
	}
	for _, tt := range tests {
		engine := liquid.NewEngine()
		engine.SetAllowedTags(map[string]struct{}{
			"people.name": {},
		})
		if tt.allowTagsWithDefault {
			engine.AllowedTagsWithDefault()
		}
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

func TestDateFormat(t *testing.T) {

	bDate, _ := date.New(1994, 4, 28, "UTC")
	bindings := map[string]interface{}{
		"people": map[string]interface{}{
			"birthday": bDate,
		},
	}
	tests := []struct {
		name     string
		v        interface{}
		src      string
		expected string
	}{
		{
			"date",
			bDate,
			"{{ people.birthday | dateFormatOrDefault: 'dmy' | default: '0001-01-01' }}",
			"28/04/1994",
		},
		{
			"time",
			time.Date(1994, time.April, 28, 0, 0, 0, 0, time.UTC),
			"{{ people.birthday | dateFormatOrDefault: 'dmy' | default: '0001-01-01' }}",
			"28/04/1994",
		},
	}
	for _, tt := range tests {
		engine := liquid.NewEngine()
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
