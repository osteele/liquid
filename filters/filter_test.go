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
	// value filters
	{`4.99 | default: 2.99`, 4.99},
	{`undefined | default: 2.99`, 2.99},
	{`false | default: 2.99`, 2.99},
	{`empty_list | default: 2.99`, 2.99},

	// array filters
	// TODO sort_natural, uniq
	{`pages | map: 'category' | join`, "business, celebrities, <nil>, lifestyle, sports, <nil>, technology"},
	{`pages | map: 'category' | compact | join`, "business, celebrities, lifestyle, sports, technology"},
	{`"John, Paul, George, Ringo" | split: ", " | join: " and "`, "John and Paul and George and Ringo"},
	{`animals | sort | join: ", "`, "Sally Snake, giraffe, octopus, zebra"},
	{`sort_prop | sort: "weight" | inspect`, `[{"weight":null},{"weight":1},{"weight":3},{"weight":5}]`},
	{`fruits | reverse | join: ", "`, "plums, peaches, oranges, apples"},
	{`fruits | first`, "apples"},
	{`fruits | last`, "plums"},
	{`empty_list | first`, nil},
	{`empty_list | last`, nil},

	// date filters
	// {`article.published_at | date`, "Fri, Jul 17, 15"},
	// {`article.published_at | date: "%a, %b %d, %y"`, "Fri, Jul 17, 15"},
	{`article.published_at | date: "%Y"`, "2015"},
	// {`"2017-02-08 19:00:00 -05:00" | date`, "Wed, Feb 08, 17"},
	{`"March 14, 2016" | date: "%b %d, %y"`, "Mar 14, 16"},
	{`"2017-02-08 09:00:00" | date: "%H:%M"`, "09:00"},
	{`"2017-05-04 08:00:00 -04:00" | date: "%b %d, %Y"`, "May 04, 2017"},
	// {`"2017-02-08 09:00:00" | date: "%-H:%M"`, "9:00"},
	// {`"now" | date: "%Y-%m-%d %H:%M"`, "2017-06-28 13:27"},

	// sequence (array or string) filters
	{`"Ground control to Major Tom." | size`, 28},
	{`"apples, oranges, peaches, plums" | split: ", " | size`, 4},

	// string filters
	// TODO escape, truncatewords, url_decode, url_encode
	{`"Take my protein pills and put my helmet on" | replace: "my", "your"`, "Take your protein pills and put your helmet on"},
	{`"Take my protein pills and put my helmet on" | replace_first: "my", "your"`, "Take your protein pills and put my helmet on"},
	{`"/my/fancy/url" | append: ".html"`, "/my/fancy/url.html"},
	{`"website.com" | append: "/index.html"`, "website.com/index.html"},
	{`"title" | capitalize`, "Title"},
	{`"my great title" | capitalize`, "My great title"},
	{`"Parker Moore" | downcase`, "parker moore"},
	{`"Have you read 'James & the Giant Peach'?" | escape`, "Have you read &#39;James &amp; the Giant Peach&#39;?"},
	{`"1 < 2 & 3" | escape_once`, "1 &lt; 2 &amp; 3"},
	{`"1 &lt; 2 &amp; 3" | escape_once`, "1 &lt; 2 &amp; 3"},
	{`"apples, oranges, and bananas" | prepend: "Some fruit: "`, "Some fruit: apples, oranges, and bananas"},
	{`"I strained to see the train through the rain" | remove: "rain"`, "I sted to see the t through the "},
	{`"I strained to see the train through the rain" | remove_first: "rain"`, "I sted to see the train through the rain"},
	{`"Liquid" | slice: 0`, "L"},
	{`"Liquid" | slice: 2`, "q"},
	{`"Liquid" | slice: 2, 5`, "quid"},
	{`"Liquid" | slice: -3, 2`, "ui"},
	{`"Have <em>you</em> read <strong>Ulysses</strong>?" | strip_html`, "Have you read Ulysses?"},
	{`"Ground control to Major Tom." | truncate: 20`, "Ground control to..."},
	{`"Ground control to Major Tom." | truncate: 25, ", and so on"`, "Ground control, and so on"},
	{`"Ground control to Major Tom." | truncate: 20, ""`, "Ground control to Ma"},
	{`"Parker Moore" | upcase`, "PARKER MOORE"},
	{`"          So much room for activities!          " | strip`, "So much room for activities!"},
	{`"          So much room for activities!          " | lstrip`, "So much room for activities!          "},
	{`"          So much room for activities!          " | rstrip`, "          So much room for activities!"},

	// number filters
	{`-17 | abs`, 17},
	{`4 | abs`, 4},
	{`"-19.86" | abs`, 19.86},

	{`1.2 | ceil`, 2},
	{`2.0 | ceil`, 2},
	{`183.357 | ceil`, 184},
	{`"3.5" | ceil`, 4},

	{`1.2 | floor`, 1},
	{`2.0 | floor`, 2},
	{`183.357 | floor`, 183},

	{`4 | plus: 2`, 6},
	{`183.357 | plus: 12`, 195.357},

	{`4 | minus: 2`, 2},
	{`16 | minus: 4`, 12},
	{`183.357 | minus: 12`, 171.357},

	{`3 | times: 2`, 6},
	{`24 | times: 7`, 168},
	{`183.357 | times: 12`, 2200.284},

	{`3 | modulo: 2`, 1},
	{`24 | modulo: 7`, 3},
	// {`183.357 | modulo: 12`, 3.357}, // TODO test suit use inexact

	{`16 | divided_by: 4`, 4},
	{`5 | divided_by: 3`, 1},
	{`20 | divided_by: 7`, 2},
	{`20 | divided_by: 7.0`, 2.857142857142857},

	{`1.2 | round`, 1},
	{`2.7 | round`, 3},
	{`183.357 | round: 2`, 183.36},

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
	"empty_list": []interface{}{},
	"fruits":     []string{"apples", "oranges", "peaches", "plums"},
	"obj": map[string]interface{}{
		"a": 1,
	},
	"pages": []map[string]interface{}{
		{"name": "page 1", "category": "business"},
		{"name": "page 2", "category": "celebrities"},
		{"name": "page 3"},
		{"name": "page 4", "category": "lifestyle"},
		{"name": "page 5", "category": "sports"},
		{"name": "page 6"},
		{"name": "page 7", "category": "technology"},
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
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			value, err := expressions.EvaluateString(test.in, filterTestContext)
			require.NoErrorf(t, err, test.in)
			expected := test.expected
			switch v := value.(type) {
			case int:
				value = float64(v)
			}
			switch ex := expected.(type) {
			case int:
				expected = float64(ex)
			}
			require.Equalf(t, expected, value, test.in)
		})
	}
}
