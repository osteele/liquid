package filters

import (
	"fmt"
	"strings"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/stretchr/testify/require"

	"github.com/osteele/liquid/expressions"
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

	// string filters
	{`"I strained to see the train through the rain" | remove_last: "rain"`, "I strained to see the train through the "},
	{`"hello world" | remove_last: "l"`, "hello word"},
	{`"no match" | remove_last: "xyz"`, "no match"},
	{`"Take my protein pills and put my helmet on" | replace_last: "my", "your"`, "Take my protein pills and put your helmet on"},
	{`"hello world" | replace_last: "l", "L"`, "hello worLd"},
	{`"no match" | replace_last: "xyz", "abc"`, "no match"},
	{`"  hello   world  " | normalize_whitespace`, " hello world "},
	{"\"hello\nworld\ttab\" | normalize_whitespace", "hello world tab"},
	{`"one two three" | number_of_words`, 3},
	{`"" | number_of_words`, 0},
	{`"   " | number_of_words`, 0},
	{`"Hello world!" | number_of_words`, 2},
	{`"你好hello世界world" | number_of_words`, 1},
	{`"   Hello    world!    " | number_of_words`, 2},
	{`"hello world" | number_of_words: "cjk"`, 2},
	{`"你好hello世界world" | number_of_words: "cjk"`, 6},
	{`"" | number_of_words: "cjk"`, 0},
	{`"你好こんにちは안녕하세요" | number_of_words: "cjk"`, 12},
	{`"hello 日本語 world" | number_of_words: "auto"`, 5},
	{`"hello world" | number_of_words: "auto"`, 2},
	{`"你好hello世界world" | number_of_words: "auto"`, 6},
	{`"你好世界" | number_of_words: "auto"`, 4},
	{`fruits | array_to_sentence_string`, "apples, oranges, peaches, and plums"},
	{`"a,b" | split: "," | array_to_sentence_string`, "a and b"},
	{`"a" | split: "," | array_to_sentence_string`, "a"},
	{`"a,b,c" | split: "," | array_to_sentence_string: "or"`, "a, b, or c"},

	// where filter [ruby: standard_filter_test.rb]
	{`where_array | where: "ok" | map: "handle" | join: " "`, "alpha delta"},
	{`where_array | where: "ok", true | map: "handle" | join: " "`, "alpha delta"},
	{`where_array | where: "ok", false | map: "handle" | join: " "`, "beta gamma"},
	{`where_messages | where: "language", "French" | map: "message" | join`, "Bonjour!"},
	{`where_messages | where: "language", "German" | map: "message" | join`, "Hallo!"},
	{`where_messages | where: "language", "English" | map: "message" | join`, "Hello!"},
	{`where_truthy | where: "foo" | map: "foo" | join: " "`, "true for sure"},

	// reject filter [ruby: standard_filter_test.rb]
	{`where_array | reject: "ok" | map: "handle" | join: " "`, "beta gamma"},
	{`where_array | reject: "ok", true | map: "handle" | join: " "`, "beta gamma"},
	{`where_array | reject: "ok", false | map: "handle" | join: " "`, "alpha delta"},

	// group_by filter [liquidjs: test/integration/filters/array.spec.ts]
	{`group_members | group_by: "graduation_year" | map: "name" | join: ", "`, "2003, 2014, 2004"},

	// find filter [ruby: standard_filter_test.rb]
	{`find_products | find: "price", 3999 | inspect`, `{"price":3999,"title":"Alpine jacket"}`},
	// find filter [liquidjs: test/integration/filters/array.spec.ts]
	{`group_members | find: "graduation_year", 2014 | inspect`, `{"graduation_year":2014,"name":"John"}`},
	{`find_products | find: "price", 9999`, nil},

	// find_index filter [ruby: standard_filter_test.rb]
	{`find_products | find_index: "price", 3999`, 2},
	// find_index filter [liquidjs: test/integration/filters/array.spec.ts]
	{`group_members | find_index: "graduation_year", 2014`, 2},
	{`group_members | find_index: "graduation_year", 2018`, nil},

	// has filter [ruby: standard_filter_test.rb]
	{`has_array_truthy | has: "ok"`, true},
	{`has_array_truthy | has: "ok", true`, true},
	{`has_array_falsy | has: "ok"`, false},
	{`has_array_truthy | has: "ok", false`, true},
	{`has_array_all_true | has: "ok", false`, false},
	// has filter [liquidjs: test/integration/filters/array.spec.ts]
	{`group_members | has: "graduation_year", 2014`, true},
	{`group_members | has: "graduation_year", 2018`, false},

	// sum filter [ruby: standard_filter_test.rb]
	{`sum_ints | sum`, int64(3)},
	{`sum_mixed | sum`, int64(10)},
	{`sum_objects | sum: "quantity"`, int64(3)},
	{`sum_objects | sum: "weight"`, int64(7)},
	{`sum_objects | sum: "subtotal"`, int64(0)},
	{`sum_floats | sum`, 0.6000000000000001},
	{`sum_neg_floats | sum`, -0.4},

	// push filter [liquidjs: test/integration/filters/array.spec.ts]
	{`fruits | push: "grapes" | join: ", "`, "apples, oranges, peaches, plums, grapes"},
	{`fruits | push: "grapes" | size`, 5},

	// unshift filter [liquidjs: test/integration/filters/array.spec.ts]
	{`fruits | unshift: "grapes" | join: ", "`, "grapes, apples, oranges, peaches, plums"},
	{`fruits | unshift: "grapes" | size`, 5},

	// pop filter [liquidjs: test/integration/filters/array.spec.ts]
	{`fruits | pop | join: ", "`, "apples, oranges, peaches"},
	{`empty_array | pop | size`, 0},

	// shift filter [liquidjs: test/integration/filters/array.spec.ts]
	{`fruits | shift | join: ", "`, "oranges, peaches, plums"},
	{`empty_array | shift | size`, 0},

	// math filters
	{`4 | at_least: 5`, 5.0},
	{`4 | at_least: 3`, 4.0},
	{`4 | at_most: 5`, 4.0},
	{`4 | at_most: 3`, 3.0},

	// html/url filters
	{`"Have you read 'James & the Giant Peach'?" | xml_escape`, "Have you read &#39;James &amp; the Giant Peach&#39;?"},
	{`'<script>"alert"</script>' | xml_escape`, "&lt;script&gt;&#34;alert&#34;&lt;/script&gt;"},
	{`"john@liquid.com" | cgi_escape`, "john%40liquid.com"},
	{`"hello world" | cgi_escape`, "hello+world"},
	{`"foo, bar; baz?" | cgi_escape`, "foo%2C+bar%3B+baz%3F"},
	{`"hello world" | uri_escape`, "hello%20world"},
	{`"http://example.com/?q=foo, \bar?" | uri_escape`, "http://example.com/?q=foo,%20%5Cbar?"},
	{`"!#$&'()*+,/:;=?@[]" | uri_escape`, "!#$&'()*+,/:;=?@[]"},
	{`"Hello World" | slugify`, "hello-world"},
	{`"The _config.yml file" | slugify`, "the-config-yml-file"},
	{`"The _config.yml file" | slugify: "pretty"`, "the-_config.yml-file"},
	{`"The _cönfig.yml file" | slugify: "ascii"`, "the-c-nfig-yml-file"},
	{`"The cönfig.yml file" | slugify: "latin"`, "the-config-yml-file"},
	{`"The _config.yml file" | slugify: "none"`, "the _config.yml file"},
	{`"The _config.yml file" | slugify: "raw"`, "the _config.yml file"},
	{`"Hello World" | slugify: "invalid_mode"`, "hello world"},

	// base64 filters
	{`"hello" | base64_encode`, "aGVsbG8="},
	{`"aGVsbG8=" | base64_decode`, "hello"},

	// type conversion filters
	{`"3.5" | to_integer`, 3},
	{`3.9 | to_integer`, 3},
	{`"42" | to_integer`, 42},
	{`true | to_integer`, 1},
	{`false | to_integer`, 0},

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

	// Test additional int/uint/float type coercion paths
	{`int8_val | plus: 1`, int64(11)},
	{`int16_val | plus: 1`, int64(101)},
	{`int32_val | plus: 1`, int64(1001)},
	{`int64_val | plus: 1`, int64(10001)},
	{`uint8_val | plus: 1`, int64(11)},
	{`uint16_val | plus: 1`, int64(101)},
	{`uint32_val | plus: 1`, int64(1001)},
	{`float32_val | plus: 1`, 11.5},
	{`str_int | plus: 1`, 11.0},
	{`str_float | plus: 1.0`, 4.5},

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
	{`"not-base64!!!" | base64_decode`, `error applying filter "base64_decode" ("illegal base64 data at input byte 3")`},
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

	// Additional int/uint/float types for coercion testing
	"int8_val":    int8(10),
	"int16_val":   int16(100),
	"int32_val":   int32(1000),
	"int64_val":   int64(10000),
	"uint8_val":   uint8(10),
	"uint16_val":  uint16(100),
	"uint32_val":  uint32(1000),
	"float32_val": float32(10.5),
	"str_int":     "10",
	"str_float":   "3.5",

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
	// where filter test data
	"where_array": []any{
		map[string]any{"handle": "alpha", "ok": true},
		map[string]any{"handle": "beta", "ok": false},
		map[string]any{"handle": "gamma", "ok": false},
		map[string]any{"handle": "delta", "ok": true},
	},
	"where_messages": []any{
		map[string]any{"message": "Bonjour!", "language": "French"},
		map[string]any{"message": "Hello!", "language": "English"},
		map[string]any{"message": "Hallo!", "language": "German"},
	},
	"where_truthy": []any{
		map[string]any{"foo": false},
		map[string]any{"foo": true},
		map[string]any{"foo": "for sure"},
		map[string]any{"bar": true},
	},
	// has filter test data
	"has_array_truthy": []any{
		map[string]any{"handle": "alpha", "ok": true},
		map[string]any{"handle": "beta", "ok": false},
		map[string]any{"handle": "gamma", "ok": false},
		map[string]any{"handle": "delta", "ok": false},
	},
	"has_array_falsy": []any{
		map[string]any{"handle": "alpha", "ok": false},
		map[string]any{"handle": "beta", "ok": false},
		map[string]any{"handle": "gamma", "ok": false},
		map[string]any{"handle": "delta", "ok": false},
	},
	"has_array_all_true": []any{
		map[string]any{"handle": "alpha", "ok": true},
		map[string]any{"handle": "beta", "ok": true},
		map[string]any{"handle": "gamma", "ok": true},
		map[string]any{"handle": "delta", "ok": true},
	},
	// group_by / find filter test data
	"group_members": []any{
		map[string]any{"graduation_year": 2003, "name": "Jay"},
		map[string]any{"graduation_year": 2003, "name": "John"},
		map[string]any{"graduation_year": 2014, "name": "John"},
		map[string]any{"graduation_year": 2004, "name": "Jack"},
	},
	// find filter test data
	"find_products": []any{
		map[string]any{"title": "Pro goggles", "price": 1299},
		map[string]any{"title": "Thermal gloves", "price": 1499},
		map[string]any{"title": "Alpine jacket", "price": 3999},
		map[string]any{"title": "Mountain boots", "price": 3899},
		map[string]any{"title": "Safety helmet", "price": 1999},
	},
	// sum filter test data
	"sum_ints":       []any{1, 2},
	"sum_mixed":      []any{1, 2, "3", "4"},
	"sum_floats":     []any{0.1, 0.2, 0.3},
	"sum_neg_floats": []any{0.1, -0.2, -0.3},
	"sum_objects": []any{
		map[string]any{"quantity": 1},
		map[string]any{"quantity": 2, "weight": 3},
		map[string]any{"weight": 4},
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

// TestSampleFilter tests the sample filter. [liquidjs: test/integration/filters/array.spec.ts]
func TestSampleFilter(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	bindings := map[string]any{
		"fruits": []any{"apples", "oranges", "peaches", "plums"},
		"empty":  []any{},
	}
	context := expressions.NewContext(bindings, cfg)

	// sample returns a single element from the array
	t.Run("single_element", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`fruits | sample`, context)
		require.NoError(t, err)
		require.Contains(t, []any{"apples", "oranges", "peaches", "plums"}, actual)
	})

	// sample with count returns array of that size
	t.Run("with_count", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`fruits | sample: 2`, context)
		require.NoError(t, err)
		arr, ok := actual.([]any)
		require.True(t, ok)
		require.Len(t, arr, 2)
	})

	// sample with count > len returns entire array
	t.Run("count_exceeds_length", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`fruits | sample: 10`, context)
		require.NoError(t, err)
		arr, ok := actual.([]any)
		require.True(t, ok)
		require.Len(t, arr, 4)
	})

	// empty array: nil input returns empty [liquidjs: `{{ nil | sample: 2 }}`]
	t.Run("empty_array", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`empty | sample`, context)
		require.NoError(t, err)
		require.Nil(t, actual)
	})

	t.Run("empty_array_with_count", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`empty | sample: 2`, context)
		require.NoError(t, err)
		require.Equal(t, []any{}, actual)
	})
}

// TestWhereFilterEdgeCases tests where filter edge cases. [liquidjs: test/integration/filters/array.spec.ts]
func TestWhereFilterEdgeCases(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	bindings := map[string]any{
		"products": []any{
			map[string]any{"title": "Vacuum", "type": "living room"},
			map[string]any{"title": "Spatula", "type": "kitchen"},
			map[string]any{"title": "Television", "type": "living room"},
			map[string]any{"title": "Garlic press", "type": "kitchen"},
			map[string]any{"title": "Coffee mug", "available": true},
			map[string]any{"title": "Sneakers", "available": false},
			map[string]any{"title": "Boring sneakers", "available": true},
		},
		"empty_array": []any{},
	}
	context := expressions.NewContext(bindings, cfg)

	// where with property and value [liquidjs: `products | where: "type", "kitchen"`]
	t.Run("with_value", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`products | where: "type", "kitchen" | map: "title" | join: ", "`, context)
		require.NoError(t, err)
		require.Equal(t, "Spatula, Garlic press", actual)
	})

	// where with truthy (no target value) [liquidjs: `products | where: "available"`]
	t.Run("truthy", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`products | where: "available" | map: "title" | join: ", "`, context)
		require.NoError(t, err)
		require.Equal(t, "Coffee mug, Boring sneakers", actual)
	})

	// where on empty array
	t.Run("empty_array", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`empty_array | where: "type", "x" | size`, context)
		require.NoError(t, err)
		require.Equal(t, 0, actual)
	})
}

// TestRejectFilterEdgeCases tests reject filter edge cases. [liquidjs: test/integration/filters/array.spec.ts]
func TestRejectFilterEdgeCases(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	bindings := map[string]any{
		"products": []any{
			map[string]any{"title": "Vacuum", "type": "living room"},
			map[string]any{"title": "Spatula", "type": "kitchen"},
			map[string]any{"title": "Television", "type": "living room"},
			map[string]any{"title": "Garlic press", "type": "kitchen"},
			map[string]any{"title": "Coffee mug", "available": true},
			map[string]any{"title": "Sneakers", "available": false},
			map[string]any{"title": "Boring sneakers", "available": true},
		},
	}
	context := expressions.NewContext(bindings, cfg)

	// reject by value [liquidjs: `products | reject: "type", "kitchen"`]
	t.Run("with_value", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`products | reject: "type", "kitchen" | map: "title" | join: ", "`, context)
		require.NoError(t, err)
		require.Equal(t, "Vacuum, Television, Coffee mug, Sneakers, Boring sneakers", actual)
	})

	// reject truthy (no target value) [liquidjs: `products | reject: "available"`]
	t.Run("truthy", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`products | reject: "available" | map: "title" | join: ", "`, context)
		require.NoError(t, err)
		require.Equal(t, "Vacuum, Spatula, Television, Garlic press, Sneakers", actual)
	})

	// reject by property existence [liquidjs: `products | reject: "type"`]
	t.Run("by_property", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`products | reject: "type" | map: "title" | join: ", "`, context)
		require.NoError(t, err)
		require.Equal(t, "Coffee mug, Sneakers, Boring sneakers", actual)
	})
}

// TestGroupByFilter tests the group_by filter. [liquidjs: test/integration/filters/array.spec.ts]
func TestGroupByFilter(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	bindings := map[string]any{
		"members": []any{
			map[string]any{"graduation_year": 2003, "name": "Jay"},
			map[string]any{"graduation_year": 2003, "name": "John"},
			map[string]any{"graduation_year": 2004, "name": "Jack"},
		},
	}
	context := expressions.NewContext(bindings, cfg)

	t.Run("basic", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`members | group_by: "graduation_year" | inspect`, context)
		require.NoError(t, err)
		require.Equal(t, `[{"items":[{"graduation_year":2003,"name":"Jay"},{"graduation_year":2003,"name":"John"}],"name":2003},{"items":[{"graduation_year":2004,"name":"Jack"}],"name":2004}]`, actual)
	})
}

// TestSumFilterEdgeCases tests sum filter edge cases. [ruby: standard_filter_test.rb]
func TestSumFilterEdgeCases(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	bindings := map[string]any{
		"with_nil":    []any{1, nil, 2},
		"with_true":   []any{1, true, nil},
		"with_string": []any{1, "foo", map[string]any{"quantity": 3}},
	}
	context := expressions.NewContext(bindings, cfg)

	// nil values are skipped [ruby: sum([1, nil, ...])]
	t.Run("with_nil", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`with_nil | sum`, context)
		require.NoError(t, err)
		require.Equal(t, int64(3), actual)
	})

	// non-numeric values (strings, maps) are skipped [ruby: sum([1, [2], "foo", { "quantity" => 3 }]) = 3]
	t.Run("with_string", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`with_string | sum`, context)
		require.NoError(t, err)
		require.Equal(t, int64(1), actual)
	})
}

// TestFindFilterEdgeCases tests find filter edge cases.
func TestFindFilterEdgeCases(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	bindings := map[string]any{
		"members": []any{
			map[string]any{"graduation_year": 2013, "name": "Jay"},
			map[string]any{"graduation_year": 2014, "name": "John"},
			map[string]any{"graduation_year": 2014, "name": "Jack", "age": 13},
		},
		"empty_array": []any{},
	}
	context := expressions.NewContext(bindings, cfg)

	// find by truthy property (no value) [liquidjs: `members | find: "age"`]
	t.Run("truthy", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`members | find: "age" | inspect`, context)
		require.NoError(t, err)
		require.Equal(t, `{"age":13,"graduation_year":2014,"name":"Jack"}`, actual)
	})

	// find not found returns nil [liquidjs: `members | find: "graduation_year", 2018`]
	t.Run("not_found", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`members | find: "graduation_year", 2018`, context)
		require.NoError(t, err)
		require.Nil(t, actual)
	})

	// find on empty array returns nil [ruby: products | find: 'title.content', 'Not found']
	t.Run("empty_array", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`empty_array | find: "price", 100`, context)
		require.NoError(t, err)
		require.Nil(t, actual)
	})
}

// TestHasFilterEdgeCases tests has filter edge cases.
func TestHasFilterEdgeCases(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	bindings := map[string]any{
		"empty_array": []any{},
		"members": []any{
			map[string]any{"graduation_year": 2013, "name": "Jay"},
			map[string]any{"graduation_year": 2014, "name": "John"},
			map[string]any{"graduation_year": 2014, "name": "Jack", "age": 13},
		},
	}
	context := expressions.NewContext(bindings, cfg)

	// has on empty array returns false [ruby: has([], 'foo', 'bar') = false]
	t.Run("empty_array", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`empty_array | has: "foo", "bar"`, context)
		require.NoError(t, err)
		require.Equal(t, false, actual)
	})

	// has truthy checks if any item has a truthy property [liquidjs: `members | has: "age"`]
	t.Run("truthy", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`members | has: "age"`, context)
		require.NoError(t, err)
		require.Equal(t, true, actual)
	})

	// has truthy not found returns false [liquidjs: `members | has: "height"`]
	t.Run("truthy_not_found", func(t *testing.T) {
		actual, err := expressions.EvaluateString(`members | has: "height"`, context)
		require.NoError(t, err)
		require.Equal(t, false, actual)
	})
}

// TestPushFilterImmutability verifies push does not mutate the original array. [liquidjs: test/integration/filters/array.spec.ts]
func TestPushFilterImmutability(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	original := []any{"hey"}
	bindings := map[string]any{
		"val": original,
	}
	context := expressions.NewContext(bindings, cfg)

	actual, err := expressions.EvaluateString(`val | push: "foo" | join: ","`, context)
	require.NoError(t, err)
	require.Equal(t, "hey,foo", actual)

	// Original should not be mutated
	require.Equal(t, []any{"hey"}, original)
}

// TestPopFilterImmutability verifies pop does not mutate the original array. [liquidjs: test/integration/filters/array.spec.ts]
func TestPopFilterImmutability(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	original := []any{"hey", "you"}
	bindings := map[string]any{
		"val": original,
	}
	context := expressions.NewContext(bindings, cfg)

	actual, err := expressions.EvaluateString(`val | pop | join: ","`, context)
	require.NoError(t, err)
	require.Equal(t, "hey", actual)

	// Original should not be mutated
	require.Equal(t, []any{"hey", "you"}, original)
}

// TestUnshiftFilterImmutability verifies unshift does not mutate the original array. [liquidjs: test/integration/filters/array.spec.ts]
func TestUnshiftFilterImmutability(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	original := []any{"you"}
	bindings := map[string]any{
		"val": original,
	}
	context := expressions.NewContext(bindings, cfg)

	actual, err := expressions.EvaluateString(`val | unshift: "hey" | join: ", "`, context)
	require.NoError(t, err)
	require.Equal(t, "hey, you", actual)

	// Original should not be mutated
	require.Equal(t, []any{"you"}, original)
}

// TestShiftFilterImmutability verifies shift does not mutate the original array. [liquidjs: test/integration/filters/array.spec.ts]
func TestShiftFilterImmutability(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	original := []any{"hey", "you"}
	bindings := map[string]any{
		"val": original,
	}
	context := expressions.NewContext(bindings, cfg)

	actual, err := expressions.EvaluateString(`val | shift | join: ","`, context)
	require.NoError(t, err)
	require.Equal(t, "you", actual)

	// Original should not be mutated
	require.Equal(t, []any{"hey", "you"}, original)
}
