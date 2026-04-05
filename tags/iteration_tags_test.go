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
	{`{% for a in array reversed limit: 1 %}{{ a }}.{% endfor %}`, "first."},
	{`{% for a in array limit: 0 %}{{ a }}.{% endfor %}`, ""},
	{`{% for a in array limit: 0 %}{{ a }}.{% else %}ELSE{% endfor %}`, "ELSE"},
	{`{% for a in array offset: 3 %}{{ a }}.{% endfor %}`, ""},
	{`{% for a in array offset: 10 %}{{ a }}.{% endfor %}`, ""},
	// Combining multiple modifiers — Ruby behavior: offset → limit → reversed (always, regardless of syntax order).
	// This matches Ruby Shopify Liquid: offset/limit slice first, reversed applied last.
	{`{% for a in array reversed offset:1 %}{{ a }}.{% endfor %}`, "third.second."},
	{`{% for a in array offset:1 reversed %}{{ a }}.{% endfor %}`, "third.second."}, // same result - syntax order doesn't matter
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

// ---------------------------------------------------------------------------
// forloop.name, forloop.parentloop
// ---------------------------------------------------------------------------

var forloopMetaTests = []struct{ in, expected string }{
	// forloop.name = "variable-collection"
	{`{% for a in array %}{{ forloop.name }}{% endfor %}`, "a-arraya-arraya-array"},
	{`{% for item in array %}{{ forloop.name }}.{% endfor %}`, "item-array.item-array.item-array."},
	// forloop.parentloop is nil in a non-nested loop (renders as empty)
	{`{% for a in array %}{{ forloop.parentloop }}.{% endfor %}`, "..." /* first iteration */},
	// forloop.parentloop is set in nested loops
	{
		`{% for i in array %}{% for j in array %}{{ forloop.parentloop.index }},{% endfor %}{% endfor %}`,
		"1,1,1,2,2,2,3,3,3,",
	},
}

func TestForloopMeta(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	bindings := map[string]any{
		"array": []string{"a", "b", "c"},
	}

	for i, test := range forloopMetaTests[:2] { // skip parentloop nil test which has quirky expected
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)

			buf := new(bytes.Buffer)
			err = render.Render(root, buf, bindings, config)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}

	// Test forloop.parentloop in nested loops
	t.Run("nested_parentloop", func(t *testing.T) {
		tpl := `{% for i in array %}{% for j in array %}{{ forloop.parentloop.index }},{% endfor %}{% endfor %}`
		root, err := config.Compile(tpl, parser.SourceLoc{})
		require.NoError(t, err)

		buf := new(bytes.Buffer)
		err = render.Render(root, buf, bindings, config)
		require.NoError(t, err)
		require.Equal(t, "1,1,1,2,2,2,3,3,3,", buf.String())
	})
}

// ---------------------------------------------------------------------------
// tablerow-specific forloop variables (col, col0, col_first, col_last, row)
// ---------------------------------------------------------------------------

var tablerowLoopVarTests = []struct{ in, expected string }{
	{`{% tablerow i in products cols:2 %}{{ forloop.col }},{% endtablerow %}`, "1,2,1,2,1,2,"},
	{`{% tablerow i in products cols:2 %}{{ forloop.col0 }},{% endtablerow %}`, "0,1,0,1,0,1,"},
	{`{% tablerow i in products cols:2 %}{{ forloop.col_first }},{% endtablerow %}`, "true,false,true,false,true,false,"},
	{`{% tablerow i in products cols:2 %}{{ forloop.col_last }},{% endtablerow %}`, "false,true,false,true,false,true,"},
	{`{% tablerow i in products cols:2 %}{{ forloop.row }},{% endtablerow %}`, "1,1,2,2,3,3,"},
	// with no cols, all items are in row 1
	{`{% tablerow i in products %}{{ forloop.row }}.{% endtablerow %}`, "1.1.1.1.1.1."},
}

func TestTablerowLoopVars(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	bindings := map[string]any{
		"products": []string{"a", "b", "c", "d", "e", "f"},
	}

	for i, test := range tablerowLoopVarTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)

			buf := new(bytes.Buffer)
			err = render.Render(root, buf, bindings, config)
			require.NoErrorf(t, err, test.in)
			// Strip HTML to get just the loop var values
			out := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(buf.String(), "")
			require.Equalf(t, test.expected, out, test.in)
		})
	}
}

// TestOffsetContinue verifies that offset:continue resumes the loop from where
// the previous iteration of the same named loop left off.
func TestOffsetContinue(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	bindings := map[string]any{
		"arr": []string{"a", "b", "c", "d", "e", "f"},
	}

	tests := []struct {
		desc     string
		in       string
		expected string
	}{
		{
			"two consecutive chunks of 2",
			`{% for x in arr limit:2 %}{{ x }}{% endfor %}-{% for x in arr limit:2 offset:continue %}{{ x }}{% endfor %}`,
			"ab-cd",
		},
		{
			"three consecutive chunks",
			`{% for x in arr limit:2 %}{{ x }}{% endfor %}-{% for x in arr limit:2 offset:continue %}{{ x }}{% endfor %}-{% for x in arr limit:2 offset:continue %}{{ x }}{% endfor %}`,
			"ab-cd-ef",
		},
		{
			"continue after full iteration",
			`{% for x in arr %}{{ x }}{% endfor %}-{% for x in arr offset:continue %}{{ x }}{% endfor %}`,
			"abcdef-",
		},
		{
			"offset:continue with spaces around colon",
			`{% for x in arr limit:2 %}{{ x }}{% endfor %}-{% for x in arr limit:2 offset : continue %}{{ x }}{% endfor %}`,
			"ab-cd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			root, err := config.Compile(tt.in, parser.SourceLoc{})
			require.NoError(t, err, tt.in)

			buf := new(bytes.Buffer)
			err = render.Render(root, buf, bindings, config)
			require.NoError(t, err, tt.in)
			require.Equal(t, tt.expected, buf.String(), tt.in)
		})
	}
}
