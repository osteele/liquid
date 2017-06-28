package filters

import (
	"fmt"
	"testing"
	"time"

	"github.com/osteele/liquid/expressions"
	"github.com/stretchr/testify/require"
)

func init() {
	DefineStandardFilters()
}

var filterTests = []struct {
	in       string
	expected interface{}
}{
	// values
	{`4.99 | default: 2.99`, 4.99},
	{`undefined | default: 2.99`, 2.99},
	{`false | default: 2.99`, 2.99},
	{`empty_list | default: 2.99`, 2.99},

	// date filters
	{`article.published_at | date`, "Fri, Jul 17, 15"},
	{`article.published_at | date: "%a, %b %d, %y"`, "Fri, Jul 17, 15"},
	{`article.published_at | date: "%Y"`, "2015"},
	{`"2017-02-08 19:00:00 -05:00" | date`, "Wed, Feb 08, 17"},
	{`"March 14, 2016" | date: "%b %d, %y"`, "Mar 14, 16"},
	// {`"now" | date: "%Y-%m-%d %H:%M"`, "2017-06-28 13:27"},

	// list filters
	// site.pages | map: 'category' | compact | join "," %}
	{`"John, Paul, George, Ringo" | split: ", " | join: " and "`, "John and Paul and George and Ringo"},
	{`animals | sort | join: ", "`, "Sally Snake, giraffe, octopus, zebra"},
	{`sort_prop | sort: "weight" | inspect`, `[{"weight":null},{"weight":1},{"weight":3},{"weight":5}]`},
	{`fruits | reverse | join: ", "`, "plums, peaches, oranges, apples"},
	// map, slice, sort_natural, size, uniq

	// string filters
	{`"Take my protein pills and put my helmet on" | replace: "my", "your"`, "Take your protein pills and put your helmet on"},
	{`"Take my protein pills and put my helmet on" | replace_first: "my", "your"`, "Take your protein pills and put my helmet on"},
	{`"/my/fancy/url" | append: ".html"`, "/my/fancy/url.html"},
	{`"website.com" | append: "/index.html"`, "website.com/index.html"},
	{`"title" | capitalize`, "Title"},
	{`"my great title" | capitalize`, "My great title"},
	{`"Parker Moore" | downcase`, "parker moore"},
	{`"Parker Moore" | upcase`, "PARKER MOORE"},
	{`"          So much room for activities!          " | strip`, "So much room for activities!"},
	{`"          So much room for activities!          " | lstrip`, "So much room for activities!          "},
	{`"          So much room for activities!          " | rstrip`, "          So much room for activities!"},
	{`"apples, oranges, and bananas" | prepend: "Some fruit: "`, "Some fruit: apples, oranges, and bananas"},
	{`"I strained to see the train through the rain" | remove: "rain"`, "I sted to see the t through the "},
	{`"I strained to see the train through the rain" | remove_first: "rain"`, "I sted to see the train through the rain"},
	{`"Ground control to Major Tom." | truncate: 20`, "Ground control to..."},
	{`"Ground control to Major Tom." | truncate: 25, ", and so on"`, "Ground control, and so on"},
	{`"Ground control to Major Tom." | truncate: 20, ""`, "Ground control to Ma"},
	// {`"Have you read 'James & the Giant Peach'?" | escape`, ""},
	// {`"1 < 2 & 3" | escape_once`, ""},
	// {`"1 &lt; 2 &amp; 3" | escape_once`, ""},
	// newline_to_br,strip_html, strip_newlines, truncatewords, // url_decode, url_encode

	// number filters
	{`-17 | abs`, 17},
	{`4 | abs`, 4},
	{`"-19.86" | abs`, 19.86},

	{`1.2 | ceil`, 2},
	{`2.0 | ceil`, 2},
	{`183.357 | ceil`, 184},
	{`"3.5" | ceil`, 4},

	// {`16 | divided_by: 4`, 4},
	// {`5 | divided_by: 3`, 1},
	// {`20 | divided_by: 7.0`, 123},

	{`1.2 | floor`, 1},
	{`2.0 | floor`, 2},
	{`183.357 | floor`, 183},
	// minus, modulo, plus, round,times

	// Jekyll extensions; added here for convenient testing
	// TODO add this just to the test environment
	{`obj | inspect`, `{"a":1}`},
}

func timeMustParse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

var filterTestContext = expressions.NewContext(map[string]interface{}{
	"x":       123,
	"animals": []string{"zebra", "octopus", "giraffe", "Sally Snake"},
	"article": map[string]interface{}{
		"published_at": timeMustParse("2015-07-17T15:04:05Z"),
	},
	"empty_list": map[string]interface{}{},
	"fruits":     []string{"apples", "oranges", "peaches", "plums"},
	"obj": map[string]interface{}{
		"a": 1,
	},
	"pages": []map[string]interface{}{
		{"category": "business"},
		{"category": "celebrities"},
		{},
		{"category": "lifestyle"},
		{"category": "sports"},
		{},
		{"category": "technology"},
	},
	"sort_prop": []map[string]interface{}{
		{"weight": 1},
		{"weight": 5},
		{"weight": 3},
		{"weight": nil},
	},
	"ar": []string{"first", "second", "third"},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
})

func TestFilters(t *testing.T) {
	for i, test := range filterTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			value, err := expressions.EvaluateExpr(test.in, filterTestContext)
			require.NoErrorf(t, err, test.in)
			expected := test.expected
			switch ex := expected.(type) {
			case int:
				expected = float64(ex)
			}
			require.Equalf(t, expected, value, test.in)
		})
	}
}
