package filters

import (
	"fmt"
	"strings"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/osteele/liquid/expressions"
	"github.com/stretchr/testify/require"
)

var filterTests = []struct {
	in       string
	expected any
}{
	// value filters
	{`undefined | default: 2.99`, 2.99},
	{`nil | default: 2.99`, 2.99},
	{`false | default: 2.99`, 2.99},
	{`"" | default: 2.99`, 2.99},
	{`empty_array | default: 2.99`, 2.99},
	{`empty_map | default: 2.99`, 2.99},
	{`empty_map_slice | default: 2.99`, 2.99},
	{`true | default: 2.99`, true},
	{`"true" | default: 2.99`, "true"},
	{`4.99 | default: 2.99`, 4.99},
	{`fruits | default: 2.99 | join`, "apples oranges peaches plums"},
	{`"string" | json`, "\"string\""},
	{`true | json`, "true"},
	{`1 | json`, "1"},

	// array filters
	{`pages | map: 'category' | join`, "business celebrities lifestyle sports technology"},
	{`pages | map: 'category' | compact | join`, "business celebrities lifestyle sports technology"},
	{`"mangos bananas persimmons" | split: " " | concat: fruits | join: ", "`, "mangos, bananas, persimmons, apples, oranges, peaches, plums"},
	{`"John, Paul, George, Ringo" | split: ", " | join: " and "`, "John and Paul and George and Ringo"},
	{`",John, Paul, George, Ringo" | split: ", " | join: " and "`, ",John and Paul and George and Ringo"},
	{`"John, Paul, George, Ringo," | split: ", " | join: " and "`, "John and Paul and George and Ringo,"},
	{`animals | sort | join: ", "`, "Sally Snake, giraffe, octopus, zebra"},
	{`sort_prop | sort: "weight" | inspect`, `[{"weight":null},{"weight":1},{"weight":3},{"weight":5}]`},
	{`fruits | reverse | join: ", "`, "plums, peaches, oranges, apples"},
	{`fruits | first`, "apples"},
	{`fruits | last`, "plums"},
	{`empty_array | first`, nil},
	{`empty_array | last`, nil},
	{`empty_array | last`, nil},
	{`dup_ints | uniq | join`, "1 2 3"},
	{`dup_strings | uniq | join`, "one two three"},
	{`dup_maps | uniq | map: "name" | join`, "m1 m2 m3"},
	{`mixed_case_array | sort_natural | join`, "a B c"},
	{`mixed_case_hash_values | sort_natural: 'key' | map: 'key' | join`, "a B c"},

	{`map_slice_has_nil | compact | join`, `a b`},
	{`map_slice_2 | first`, `b`},
	{`map_slice_2 | last`, `a`},
	{`map_slice_2 | join`, `b a`},
	{`map_slice_objs | map: "key" | join`, `a b`},
	{`map_slice_2 | reverse | join`, `a b`},
	{`map_slice_2 | sort | join`, `a b`},
	{`map_slice_dup | join`, `a a b`},
	{`map_slice_dup | uniq | join`, `a b`},

	{`struct_slice | map: "str" | join`, `a b c`},

	// date filters
	{`article.published_at | date`, "Fri, Jul 17, 15"},
	{`article.published_at | date: "%a, %b %d, %y"`, "Fri, Jul 17, 15"},
	{`article.published_at | date: "%Y"`, "2015"},
	{`"2017-02-08 19:00:00 -05:00" | date`, "Wed, Feb 08, 17"},
	{`"2017-05-04 08:00:00 -04:00" | date: "%b %d, %Y"`, "May 04, 2017"},
	{`"2017-02-08 09:00:00" | date: "%H:%M"`, "09:00"},
	{`"2017-02-08 09:00:00" | date: "%-H:%M"`, "9:00"},
	{`"2017-02-08 09:00:00" | date: "%d/%m"`, "08/02"},
	{`"2017-02-08 09:00:00" | date: "%e/%m"`, " 8/02"},
	{`"2017-02-08 09:00:00" | date: "%-d/%-m"`, "8/2"},
	{`"March 14, 2016" | date: "%b %d, %y"`, "Mar 14, 16"},
	{`"2017-07-09" | date: "%d/%m"`, "09/07"},
	{`"2017-07-09" | date: "%e/%m"`, " 9/07"},
	{`"2017-07-09" | date: "%-d/%-m"`, "9/7"},

	// sequence (array or string) filters
	{`"Ground control to Major Tom." | size`, 28},
	{`"apples, oranges, peaches, plums" | split: ", " | size`, 4},
	// count chars, not bytes
	{`"Straße" | size`, 6},

	// string filters
	{`"Take my protein pills and put my helmet on" | replace: "my", "your"`, "Take your protein pills and put your helmet on"},
	{`"Take my protein pills and put my helmet on" | replace_first: "my", "your"`, "Take your protein pills and put my helmet on"},
	{`"/my/fancy/url" | append: ".html"`, "/my/fancy/url.html"},
	{`"website.com" | append: "/index.html"`, "website.com/index.html"},
	{`"title" | capitalize`, "Title"},
	{`"my great title" | capitalize`, "My great title"},
	{`"" | capitalize`, ""},
	{`"Parker Moore" | downcase`, "parker moore"},
	{`"Have you read 'James & the Giant Peach'?" | escape`, "Have you read &#39;James &amp; the Giant Peach&#39;?"},
	{`"1 < 2 & 3" | escape_once`, "1 &lt; 2 &amp; 3"},
	{`string_with_newlines | newline_to_br`, "<br />Hello<br />there<br />"},
	{`"1 &lt; 2 &amp; 3" | escape_once`, "1 &lt; 2 &amp; 3"},
	{`"apples, oranges, and bananas" | prepend: "Some fruit: "`, "Some fruit: apples, oranges, and bananas"},
	{`"I strained to see the train through the rain" | remove: "rain"`, "I sted to see the t through the "},
	{`"I strained to see the train through the rain" | remove_first: "rain"`, "I sted to see the train through the rain"},

	{`"Liquid" | slice: 0`, "L"},
	{`"Liquid
Liquid" | slice: 0`, "L"},
	{`"Liquid" | slice: 2`, "q"},
	{`"Liquid" | slice: 2, 5`, "quid"},
	{`"Liquid
Liquid" | slice: 2, 4`, "quid"},
	{`"Liquid" | slice: -3, 2`, "ui"},
	{`"" | slice: 1`, ""},
	{`"Liquid" | slice: 2, 100`, "quid"},
	{`"Liquid" | slice: 100`, ""},
	{`"Liquid" | slice: 100, 200`, ""},
	{`"Liquid" | slice: -100`, "L"},
	{`"Liquid" | slice: -100, 200`, "Liquid"},
	{`"白鵬翔" | slice: 0`, "白"},
	{`"白鵬翔" | slice: 1`, "鵬"},
	{`"白鵬翔" | slice: 2`, "翔"},
	{`"白鵬翔" | slice: 0, 2`, "白鵬"},
	{`"白鵬翔" | slice: 1, 2`, "鵬翔"},
	{`"白鵬翔" | slice: 100, 200`, ""},
	{`"白鵬翔" | slice: -100`, "白"},
	{`"白鵬翔" | slice: -100, 200`, "白鵬翔"},
	{`">` + strings.Repeat(".", 10000) + `<" | slice: 1, 10000`, strings.Repeat(".", 10000)},
	{`"a,b,c" | split: "," | slice: -1 | join`, "c"},
	{`"a,b,c" | split: "," | slice: 1, 1 | join`, "b"},
	{`"a,b,c" | split: "," | slice: 0, 2 | join`, "a b"},
	{`"a,b,c" | split: "," | slice: 1, 2 | join`, "b c"},

	{`"a/b/c" | split: '/' | join: '-'`, "a-b-c"},
	{`"a/b/" | split: '/' | join: '-'`, "a-b"},
	{`"a//c" | split: '/' | join: '-'`, "a--c"},
	{`"a//" | split: '/' | join: '-'`, "a"},
	{`"/b/c" | split: '/' | join: '-'`, "-b-c"},
	{`"/b/" | split: '/' | join: '-'`, "-b"},
	{`"//c" | split: '/' | join: '-'`, "--c"},
	{`"//" | split: '/' | join: '-'`, ""},
	{`"/" | split: '/' | join: '-'`, ""},
	{`"a.b" | split: '.' | join: '-'`, "a-b"},
	{`"a..b" | split: '.' | join: '-'`, "a--b"},
	{"'a.\t.b' | split: '.' | join: '-'", "a-\t-b"},
	{`"a b" | split: ' ' | join: '-'`, "a-b"},
	{`"a  b" | split: ' ' | join: '-'`, "a-b"},
	{"'a \t b' | split: ' ' | join: '-'", "a-b"},

	{`"Have <em>you</em> read <strong>Ulysses</strong>?" | strip_html`, "Have you read Ulysses?"},
	{`string_with_newlines | strip_newlines`, "Hellothere"},

	{`"Ground control to Major Tom." | truncate: 20`, "Ground control to..."},
	{`"Ground control to Major Tom." | truncate: 25, ", and so on"`, "Ground control, and so on"},
	{`"Ground control to Major Tom." | truncate: 20, ""`, "Ground control to Ma"},
	{`"Ground" | truncate: 20`, "Ground"},
	{`"Ground control to Major Tom." | truncatewords: 3`, "Ground control to..."},
	{`"Ground control to Major Tom." | truncatewords: 3, "--"`, "Ground control to--"},
	{`"Ground control to Major Tom." | truncatewords: 3, ""`, "Ground control to"},
	{`"Ground control" | truncatewords: 3, ""`, "Ground control"},
	{`"Ground" | truncatewords: 3, ""`, "Ground"},
	{`"  Ground" | truncatewords: 3, ""`, "  Ground"},
	{`"" | truncatewords: 3, ""`, ""},
	{`"  " | truncatewords: 3, ""`, "  "},

	{`"Parker Moore" | upcase`, "PARKER MOORE"},
	{`"          So much room for activities!          " | strip`, "So much room for activities!"},
	{`"          So much room for activities!          " | lstrip`, "So much room for activities!          "},
	{`"          So much room for activities!          " | rstrip`, "          So much room for activities!"},

	{`"%27Stop%21%27+said+Fred" | url_decode`, "'Stop!' said Fred"},
	{`"john@liquid.com" | url_encode`, "john%40liquid.com"},
	{`"Tetsuro Takara" | url_encode`, "Tetsuro+Takara"},

	// number filters
	{`-17 | abs`, 17.0},
	{`4 | abs`, 4.0},
	{`"-19.86" | abs`, 19.86},

	{`1.2 | ceil`, 2},
	{`2.0 | ceil`, 2},
	{`183.357 | ceil`, 184},
	{`"3.5" | ceil`, 4},

	{`1.2 | floor`, 1},
	{`2.0 | floor`, 2},
	{`183.357 | floor`, 183},

	{`4 | plus: 2`, int64(6)},
	{`4 | plus: 2.0`, 6.0},
	{`4.0 | plus: 2`, 6.0},
	{`183.357 | plus: 12`, 195.357},

	{`4 | minus: 2`, int64(2)},
	{`4 | minus: 2.0`, 2.0},
	{`16 | minus: 4`, int64(12)},
	{`183.357 | minus: 12`, 171.357},

	{`3 | times: 2`, int64(6)},
	{`3 | times: 2.0`, 6.0},
	{`24 | times: 7`, int64(168)},
	{`183.357 | times: 12`, 2200.284},

	// Test large integers (issue #109 - should not use scientific notation)
	{`1743096453 | minus: 7`, int64(1743096446)},
	{`1743096453 | plus: 7`, int64(1743096460)},
	{`1000000 | times: 1000`, int64(1000000000)},

	// Test uint types - should preserve integer type when in int64 range
	{`small_uint | plus: 1`, int64(1001)},
	{`small_uint | minus: 1`, int64(999)},
	{`small_uint | times: 2`, int64(2000)},
	{`small_uint64 | plus: 100`, int64(1100)},
	{`small_uint64 | minus: 100`, int64(900)},
	{`small_uint64 | times: 3`, int64(3000)},

	{`3 | modulo: 2`, 1.0},
	{`24 | modulo: 7`, 3.0},
	// {`183.357 | modulo: 12 | `, 3.357}, // TODO test suit use inexact

	{`16 | divided_by: 4`, int64(4)},
	{`5 | divided_by: 3`, int64(1)},
	{`20 | divided_by: 7`, int64(2)},
	{`20 | divided_by: 7.0`, 2.857142857142857},

	{`1.2 | round`, 1.0},
	{`2.7 | round`, 3.0},
	{`183.357 | round: 2`, 183.36},

	// Jekyll extensions; added here for convenient testing
	// TODO add this just to the test environment
	{`map | inspect`, `{"a":1}`},
	{`1 | type`, `int`},
	{`"1" | type`, `string`},
}

var filterErrorTests = []struct {
	in    string
	error string
}{
	{`20 | divided_by: 's'`, `error applying filter "divided_by" ("invalid divisor: 's'")`},
	{`20 | divided_by: 0`, `error applying filter "divided_by" ("division by zero")`},
}

var filterTestBindings = map[string]any{
	"empty_array":     []any{},
	"empty_map":       map[string]any{},
	"empty_map_slice": yaml.MapSlice{},
	"map": map[string]any{
		"a": 1,
	},
	"map_slice_2":       yaml.MapSlice{{Key: 1, Value: "b"}, {Key: 2, Value: "a"}},
	"map_slice_dup":     yaml.MapSlice{{Key: 1, Value: "a"}, {Key: 2, Value: "a"}, {Key: 3, Value: "b"}},
	"map_slice_has_nil": yaml.MapSlice{{Key: 1, Value: "a"}, {Key: 2, Value: nil}, {Key: 3, Value: "b"}},
	"map_slice_objs": yaml.MapSlice{
		{Key: 1, Value: map[string]any{"key": "a"}},
		{Key: 2, Value: map[string]any{"key": "b"}},
	},
	"mixed_case_array": []string{"c", "a", "B"},
	"mixed_case_hash_values": []map[string]any{
		{"key": "c"},
		{"key": "a"},
		{"key": "B"},
	},
	"sort_prop": []map[string]any{
		{"weight": 1},
		{"weight": 5},
		{"weight": 3},
		{"weight": nil},
	},
	"string_with_newlines": "\nHello\nthere\n",
	"dup_ints":             []int{1, 2, 1, 3},
	"dup_strings":          []string{"one", "two", "one", "three"},

	// Test uint types in arithmetic operations (issue #109)
	"small_uint":   uint(1000),
	"small_uint64": uint64(1000),

	// for examples from liquid docs
	"animals": []string{"zebra", "octopus", "giraffe", "Sally Snake"},
	"fruits":  []string{"apples", "oranges", "peaches", "plums"},
	"article": map[string]any{
		"published_at": timeMustParse("2015-07-17T15:04:05Z"),
	},
	"page": map[string]any{
		"title": "Introduction",
	},
	"pages": []map[string]any{
		{"name": "page 1", "category": "business"},
		{"name": "page 2", "category": "celebrities"},
		{"name": "page 3"},
		{"name": "page 4", "category": "lifestyle"},
		{"name": "page 5", "category": "sports"},
		{"name": "page 6"},
		{"name": "page 7", "category": "technology"},
	},
	"struct_slice": []struct {
		Str string `liquid:"str"`
	}{
		{Str: "a"},
		{Str: "b"},
		{Str: "c"},
	},
}

func TestFilters(t *testing.T) {
	t.Setenv("TZ", "America/New_York")

	var (
		m1 = map[string]any{"name": "m1"}
		m2 = map[string]any{"name": "m2"}
		m3 = map[string]any{"name": "m3"}
	)

	filterTestBindings["dup_maps"] = []any{m1, m2, m1, m3}

	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	context := expressions.NewContext(filterTestBindings, cfg)

	for i, test := range filterTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			actual, err := expressions.EvaluateString(test.in, context)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, actual, test.in)
		})
	}

	for i, test := range filterErrorTests {
		t.Run(fmt.Sprintf("%02d", i+len(filterTests)+1), func(t *testing.T) {
			_, err := expressions.EvaluateString(test.in, context)
			require.EqualErrorf(t, err, test.error, test.in)
		})
	}
}

func timeMustParse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	return t
}
