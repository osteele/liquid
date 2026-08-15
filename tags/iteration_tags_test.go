package tags

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/stretchr/testify/require"

	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
)

var iterationTests = []struct{ in, expected string }{
	{`{% for a in array %}{{ a }} {% endfor %}`, "first second third "},
	{`{% for a in array %}{{ a }} {% else %}else{% endfor %}`, "first second third "},
	{`{% for a in nil %}{{ a }}.{% endfor %}`, ""},
	{`{% for a in false %}{{ a }}.{% endfor %}`, ""},
	{`{% for a in 2 %}{{ a }}.{% endfor %}`, ""},
	{`{% for a in "str" %}{{ a }}.{% endfor %}`, ""},
	{`{% for a in map %}{{ a[0] }}={{ a[1] }}.{% endfor %}`, "a=1."},
	{`{% for a in map_slice %}{{ a[0] }}={{ a[1] }}.{% endfor %}`, "a=1.b=2."},
	{`{% for k in keyed_map %}{{ k }}={{ keyed_map[k] }}.{% endfor %}`, "a=1.b=2."},

	// loop modifiers
	{`{% for a in array reversed %}{{ a }}.{% endfor %}`, "third.second.first."},
	{`{% for a in array limit: 2 %}{{ a }}.{% endfor %}`, "first.second."},
	{`{% for a in array limit: limit %}{{ a }}.{% endfor %}`, "first.second."},
	{`{% for a in array limit: loopmods.limit %}{{ a }}.{% endfor %}`, "first.second."},
	{`{% for a in array limit: loopmods["limit"] %}{{ a }}.{% endfor %}`, "first.second."},
	{`{% for a in array offset: 1 %}{{ a }}.{% endfor %}`, "second.third."},
	{`{% for a in array offset: offset %}{{ a }}.{% endfor %}`, "second.third."},
	{`{% for a in array offset: loopmods.offset %}{{ a }}.{% endfor %}`, "second.third."},
	{`{% for a in array offset: loopmods["offset"] %}{{ a }}.{% endfor %}`, "second.third."},
	{`{% for a in array reversed limit: 1 %}{{ a }}.{% endfor %}`, "third."},
	{`{% for a in array limit: 0 %}{{ a }}.{% endfor %}`, ""},
	{`{% for a in array limit: 0 %}{{ a }}.{% else %}ELSE{% endfor %}`, "ELSE"},
	{`{% for a in array offset: 3 %}{{ a }}.{% endfor %}`, ""},
	{`{% for a in array offset: 10 %}{{ a }}.{% endfor %}`, ""},
	// Combining multiple modifiers (issue #6)
	// Note: In this implementation, modifiers are always applied in the order: reversed -> offset -> limit
	// The order they appear in the template syntax does not matter.
	// This differs from Ruby Shopify Liquid where syntax order matters and reversed only works when placed first.
	{`{% for a in array reversed offset:1 %}{{ a }}.{% endfor %}`, "second.first."},
	{`{% for a in array offset:1 reversed %}{{ a }}.{% endfor %}`, "second.first."}, // same result - syntax order doesn't matter
	{`{% for a in array limit:1 offset:1 %}{{ a }}.{% endfor %}`, "second."},
	{`{% for a in array offset:1 limit:1 %}{{ a }}.{% endfor %}`, "second."}, // same result
	{`{% for a in array reversed limit:1 offset:1 %}{{ a }}.{% endfor %}`, "second."},
	{`{% for a in array reversed offset:1 limit:1 %}{{ a }}.{% endfor %}`, "second."}, // same result
	{`{% for a in array limit:1 offset:1 reversed %}{{ a }}.{% endfor %}`, "second."}, // same result
	{`{% for a in array offset:1 limit:1 reversed %}{{ a }}.{% endfor %}`, "second."}, // same result

	// loop variables
	{`{% for a in array %}{{ forloop.first }}.{% endfor %}`, "true.false.false."},
	{`{% for a in array %}{{ forloop.last }}.{% endfor %}`, "false.false.true."},
	{`{% for a in array %}{{ forloop.index }}.{% endfor %}`, "1.2.3."},
	{`{% for a in array %}{{ forloop.index0 }}.{% endfor %}`, "0.1.2."},
	{`{% for a in array %}{{ forloop.rindex }}.{% endfor %}`, "3.2.1."},
	{`{% for a in array %}{{ forloop.rindex0 }}.{% endfor %}`, "2.1.0."},
	{`{% for a in array %}{{ forloop.length }}.{% endfor %}`, "3.3.3."},

	{
		`{% for i in array %}{{ forloop.index }}[{% for j in array %}{{ forloop.index }}{% endfor %}]{{ forloop.index }}{% endfor %}`,
		"1[123]12[123]23[123]3",
	},

	{`{% for a in array reversed %}{{ forloop.first }}.{% endfor %}`, "true.false.false."},
	{`{% for a in array reversed %}{{ forloop.last }}.{% endfor %}`, "false.false.true."},
	{`{% for a in array reversed %}{{ forloop.index }}.{% endfor %}`, "1.2.3."},
	{`{% for a in array reversed %}{{ forloop.rindex }}.{% endfor %}`, "3.2.1."},
	{`{% for a in array reversed %}{{ forloop.length }}.{% endfor %}`, "3.3.3."},

	{`{% for a in array limit:2 %}{{ forloop.index }}.{% endfor %}`, "1.2."},
	{`{% for a in array limit:2 %}{{ forloop.rindex }}.{% endfor %}`, "2.1."},
	{`{% for a in array limit:2 %}{{ forloop.first }}.{% endfor %}`, "true.false."},
	{`{% for a in array limit:2 %}{{ forloop.last }}.{% endfor %}`, "false.true."},
	{`{% for a in array limit:2 %}{{ forloop.length }}.{% endfor %}`, "2.2."},

	{`{% for a in array offset:1 %}{{ forloop.index }}.{% endfor %}`, "1.2."},
	{`{% for a in array offset:1 %}{{ forloop.rindex }}.{% endfor %}`, "2.1."},
	{`{% for a in array offset:1 %}{{ forloop.first }}.{% endfor %}`, "true.false."},
	{`{% for a in array offset:1 %}{{ forloop.last }}.{% endfor %}`, "false.true."},
	{`{% for a in array offset:1 %}{{ forloop.length }}.{% endfor %}`, "2.2."},

	{`{% for a in array %}{% if a == 'second' %}{% break %}{% endif %}{{ a }}{% endfor %}`, "first"},
	{`{% for a in array %}{% if a == 'second' %}{% continue %}{% endif %}{{ a }}.{% endfor %}`, "first.third."},

	// cycle
	{`{% for a in array %}{% cycle 'even', 'odd' %}.{% endfor %}`, "even.odd.even."},
	{`{% for a in array %}{% cycle '0', '1' %},{% cycle '0', '1' %}.{% endfor %}`, "0,1.0,1.0,1."},
	// {`{% for a in array %}{% cycle group: 'a', '0', '1' %},{% cycle '0', '1' %}.{% endfor %}`, "0,1.0,1.0,1."},

	// range
	{`{% for i in (3 .. 5) %}{{i}}.{% endfor %}`, "3.4.5."},
	{`{% for i in (3..5) %}{{i}}.{% endfor %}`, "3.4.5."},
	{`{% assign l = (3..5) %}{% for i in l %}{{i}}.{% endfor %}`, "3.4.5."},

	// tablerow
	{
		`{% tablerow product in products %}{{ product }}{% endtablerow %}`,
		`<tr class="row1"><td class="col1">Cool Shirt</td>
	<td class="col2">Alien Poster</td>
	<td class="col3">Batman Poster</td>
	<td class="col4">Bullseye Shirt</td>
	<td class="col5">Another Classic Vinyl</td>
	<td class="col6">Awesome Jeans</td></tr>`,
	},

	{
		`{% tablerow product in products cols:2 %}{{ product }}{% endtablerow %}`,
		`<tr class="row1"><td class="col1">Cool Shirt</td><td class="col2">Alien Poster</td></tr>
		 <tr class="row2"><td class="col1">Batman Poster</td><td class="col2">Bullseye Shirt</td></tr>
	  	 <tr class="row3"><td class="col1">Another Classic Vinyl</td><td class="col2">Awesome Jeans</td></tr>`,
	},
	{
		`{% tablerow product in products cols: cols %}{{ product }}{% endtablerow %}`,
		`<tr class="row1"><td class="col1">Cool Shirt</td><td class="col2">Alien Poster</td></tr>
		 <tr class="row2"><td class="col1">Batman Poster</td><td class="col2">Bullseye Shirt</td></tr>
	  	 <tr class="row3"><td class="col1">Another Classic Vinyl</td><td class="col2">Awesome Jeans</td></tr>`,
	},
	{
		`{% tablerow product in products cols: loopmods.cols %}{{ product }}{% endtablerow %}`,
		`<tr class="row1"><td class="col1">Cool Shirt</td><td class="col2">Alien Poster</td></tr>
		 <tr class="row2"><td class="col1">Batman Poster</td><td class="col2">Bullseye Shirt</td></tr>
		 <tr class="row3"><td class="col1">Another Classic Vinyl</td><td class="col2">Awesome Jeans</td></tr>`,
	},
	{
		`{% tablerow product in products cols: loopmods.cols %}{{ product }}{% endtablerow %}`,
		`<tr class="row1"><td class="col1">Cool Shirt</td><td class="col2">Alien Poster</td></tr>
		 <tr class="row2"><td class="col1">Batman Poster</td><td class="col2">Bullseye Shirt</td></tr>
		 <tr class="row3"><td class="col1">Another Classic Vinyl</td><td class="col2">Awesome Jeans</td></tr>`,
	},
}

var iterationSyntaxErrorTests = []struct{ in, expected string }{
	{`{% for a b c %}{% endfor %}`, "syntax error"},
	{`{% for a in array offset %}{% endfor %}`, "undefined loop modifier"},
	{`{% cycle %}`, "syntax error"},
}

var iterationErrorTests = []struct{ in, expected string }{
	{`{% break %}`, "break outside a loop"},
	{`{% continue %}`, "continue outside a loop"},
	{`{% cycle 'a', 'b' %}`, "cycle must be within a forloop"},
	{`{% for a in array | undefined_filter %}{% endfor %}`, "undefined filter"},
	{`{% for a in array %}{{ a | undefined_filter }}{% endfor %}`, "undefined filter"},
	{`{% for a in array %}{% else %}{% else %}{% endfor %}`, "for loops accept at most one else clause"},
}

var iterationTestBindings = map[string]any{
	"array": []string{"first", "second", "third"},
	// hash has only one element, since iteration order is non-deterministic
	"map":       map[string]any{"a": 1},
	"keyed_map": IterationKeyedMap(map[string]any{"a": 1, "b": 2}),
	"map_slice": yaml.MapSlice{{Key: "a", Value: 1}, {Key: "b", Value: 2}},
	"products": []string{
		"Cool Shirt", "Alien Poster", "Batman Poster", "Bullseye Shirt", "Another Classic Vinyl", "Awesome Jeans",
	},
	"offset":   1,
	"limit":    2,
	"cols":     2,
	"loopmods": map[string]any{"limit": 2, "offset": 1, "cols": 2},
}

func TestIterationTags(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	for i, test := range iterationTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)

			buf := new(bytes.Buffer)
			err = render.Render(root, buf, iterationTestBindings, config)
			require.NoErrorf(t, err, test.in)

			actual := buf.String()

			if strings.Contains(test.in, "{% tablerow") {
				replaceWS := regexp.MustCompile(`\n\s*`).ReplaceAllString
				actual = replaceWS(actual, "")
				test.expected = replaceWS(test.expected, "")
			}

			require.Equalf(t, test.expected, actual, test.in)
		})
	}
}

func TestIterationTags_errors(t *testing.T) {
	cfg := render.NewConfig()
	AddStandardTags(&cfg)

	for i, test := range iterationSyntaxErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			_, err := cfg.Compile(test.in, parser.SourceLoc{})
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}

	for i, test := range iterationErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1+len(iterationSyntaxErrorTests)), func(t *testing.T) {
			root, err := cfg.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			err = render.Render(root, io.Discard, iterationTestBindings, cfg)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}

// loopControlStateTests models the for-loop control state machine:
// (current iteration, control action) -> rendered output.
var loopControlStateTests = []struct {
	name     string
	template string
	expected string
}{
	{"break on first iteration", `{% for a in array %}{% break %}{{ a }}.{% endfor %}`, ""},
	{"break in middle", `{% for a in array %}{% if a == 'second' %}{% break %}{% endif %}{{ a }}.{% endfor %}`, "first."},
	{"break after last item", `{% for a in array %}{{ a }}.{% if a == 'third' %}{% break %}{% endif %}{% endfor %}`, "first.second.third."},
	{"continue on first iteration", `{% for a in array %}{% continue %}{{ a }}.{% endfor %}`, ""},
	{"continue in middle", `{% for a in array %}{% if a == 'second' %}{% continue %}{% endif %}{{ a }}.{% endfor %}`, "first.third."},
	{"continue on last item", `{% for a in array %}{{ a }}.{% if a == 'third' %}{% continue %}{% endif %}{% endfor %}`, "first.second.third."},
	{"break only affects inner loop", `{% for a in array %}{% for b in array %}{% if b == 'second' %}{% break %}{% endif %}{{ b }}{% endfor %}:{{ a }}.{% endfor %}`, "first:first.first:second.first:third."},
	{"else on empty array", `{% for a in empty %}{{ a }}.{% else %}none{% endfor %}`, "none"},
	{"else not used", `{% for a in array %}{{ a }}.{% else %}none{% endfor %}`, "first.second.third."},
}

func TestLoopControlState(t *testing.T) {
	cfg := render.NewConfig()
	AddStandardTags(&cfg)

	bindings := map[string]any{
		"array": []string{"first", "second", "third"},
		"empty": []string{},
	}

	for _, test := range loopControlStateTests {
		t.Run(test.name, func(t *testing.T) {
			root, err := cfg.Compile(test.template, parser.SourceLoc{})
			require.NoErrorf(t, err, test.template)

			buf := new(bytes.Buffer)
			err = render.Render(root, buf, bindings, cfg)
			require.NoErrorf(t, err, test.template)
			require.Equalf(t, test.expected, buf.String(), test.template)
		})
	}
}

// forloopMapStateTests models the forloop variable state at each iteration.
// Each row is (loop configuration, iteration index) -> expected forloop map.
var forloopMapStateTests = []struct {
	name     string
	template string
	bindings map[string]any
	expected []map[string]any
}{
	{
		name:     "three iterations",
		template: `{% for a in array %}{% record_forloop %}{% endfor %}`,
		bindings: map[string]any{"array": []string{"x", "y", "z"}},
		expected: []map[string]any{
			{"first": true, "last": false, "index": 1, "index0": 0, "rindex": 3, "rindex0": 2, "length": 3},
			{"first": false, "last": false, "index": 2, "index0": 1, "rindex": 2, "rindex0": 1, "length": 3},
			{"first": false, "last": true, "index": 3, "index0": 2, "rindex": 1, "rindex0": 0, "length": 3},
		},
	},
	{
		name:     "limit modifier",
		template: `{% for a in array limit:2 %}{% record_forloop %}{% endfor %}`,
		bindings: map[string]any{"array": []string{"x", "y", "z"}},
		expected: []map[string]any{
			{"first": true, "last": false, "index": 1, "index0": 0, "rindex": 2, "rindex0": 1, "length": 2},
			{"first": false, "last": true, "index": 2, "index0": 1, "rindex": 1, "rindex0": 0, "length": 2},
		},
	},
	{
		name:     "offset modifier",
		template: `{% for a in array offset:1 %}{% record_forloop %}{% endfor %}`,
		bindings: map[string]any{"array": []string{"x", "y", "z"}},
		expected: []map[string]any{
			{"first": true, "last": false, "index": 1, "index0": 0, "rindex": 2, "rindex0": 1, "length": 2},
			{"first": false, "last": true, "index": 2, "index0": 1, "rindex": 1, "rindex0": 0, "length": 2},
		},
	},
	{
		name:     "reversed modifier",
		template: `{% for a in array reversed %}{% record_forloop %}{% endfor %}`,
		bindings: map[string]any{"array": []string{"x", "y", "z"}},
		expected: []map[string]any{
			{"first": true, "last": false, "index": 1, "index0": 0, "rindex": 3, "rindex0": 2, "length": 3},
			{"first": false, "last": false, "index": 2, "index0": 1, "rindex": 2, "rindex0": 1, "length": 3},
			{"first": false, "last": true, "index": 3, "index0": 2, "rindex": 1, "rindex0": 0, "length": 3},
		},
	},
	{
		name:     "single iteration",
		template: `{% for a in one %}{% record_forloop %}{% endfor %}`,
		bindings: map[string]any{"one": []string{"only"}},
		expected: []map[string]any{
			{"first": true, "last": true, "index": 1, "index0": 0, "rindex": 1, "rindex0": 0, "length": 1},
		},
	},
}

func TestForloopMapState(t *testing.T) {
	var recorded []map[string]any

	cfg := render.NewConfig()
	AddStandardTags(&cfg)
	cfg.AddTag("record_forloop", func(string) (func(io.Writer, render.Context) error, error) {
		return func(_ io.Writer, ctx render.Context) error {
			loopVar := ctx.Get("forloop")
			require.NotNil(t, loopVar, "forloop variable should be set")
			loopMap, ok := loopVar.(map[string]any)
			require.True(t, ok, "forloop variable should be a map")

			// Clone and drop the internal .cycles map so the table stays focused
			// on the documented forloop variables.
			cloned := make(map[string]any, len(loopMap))
			for k, v := range loopMap {
				if k == ".cycles" {
					continue
				}
				cloned[k] = v
			}
			recorded = append(recorded, cloned)
			return nil
		}, nil
	})

	for _, test := range forloopMapStateTests {
		t.Run(test.name, func(t *testing.T) {
			recorded = nil

			root, err := cfg.Compile(test.template, parser.SourceLoc{})
			require.NoErrorf(t, err, test.template)

			err = render.Render(root, io.Discard, test.bindings, cfg)
			require.NoErrorf(t, err, test.template)
			require.Equalf(t, test.expected, recorded, test.template)
		})
	}
}
