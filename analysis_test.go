package liquid

import (
	"reflect"
	"sort"
	"testing"
)

func TestGlobalVariableSegments(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		want     [][]string
	}{
		{
			name:     "simple variable",
			template: `{{ x }}`,
			want:     [][]string{{"x"}},
		},
		{
			name:     "property access",
			template: `{{ x.a.b }}`,
			want:     [][]string{{"x", "a", "b"}},
		},
		{
			name:     "assign makes local",
			template: `{% assign y = x.val %}{{ y }}`,
			want:     [][]string{{"x", "val"}},
		},
		{
			name:     "for loop variable is local",
			template: `{% for item in list %}{{ item.name }}{% endfor %}`,
			want:     [][]string{{"list"}},
		},
		{
			name:     "if condition",
			template: `{% if cond %}{{ a }}{% else %}{{ b }}{% endif %}`,
			want:     [][]string{{"cond"}, {"a"}, {"b"}},
		},
		{
			name:     "filter does not change path",
			template: `{{ x | upcase }}`,
			want:     [][]string{{"x"}},
		},
		{
			name:     "capture makes local",
			template: `{% capture buf %}{{ x }}{% endcapture %}`,
			want:     [][]string{{"x"}},
		},
		{
			name:     "assign of literal makes var local",
			template: `{% assign x = 1 %}{{ x }}`,
			want:     nil,
		},
		{
			name:     "case statement",
			template: `{% case status %}{% when "active" %}{{ a }}{% endcase %}`,
			want:     [][]string{{"status"}, {"a"}},
		},
		{
			name:     "multiple variables",
			template: `{{ customer.first_name }} {% assign x = "hello" %} {{ order.total }}`,
			want:     [][]string{{"customer", "first_name"}, {"order", "total"}},
		},
		{
			name:     "unless",
			template: `{% unless flag %}{{ val }}{% endunless %}`,
			want:     [][]string{{"flag"}, {"val"}},
		},
		{
			name:     "elsif clauses",
			template: `{% if a %}{{ x }}{% elsif b %}{{ y }}{% else %}{{ z }}{% endif %}`,
			want:     [][]string{{"a"}, {"x"}, {"b"}, {"y"}, {"z"}},
		},
		{
			name:     "tablerow",
			template: `{% tablerow item in products %}{{ item.title }}{% endtablerow %}`,
			want:     [][]string{{"products"}},
		},
		{
			name:     "no variables",
			template: `hello world`,
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, parseErr := engine.ParseString(tt.template)
			if parseErr != nil {
				t.Fatalf("ParseString(%q) error: %v", tt.template, parseErr)
			}

			got, analyzeErr := engine.GlobalVariableSegments(tpl)
			if analyzeErr != nil {
				t.Fatalf("GlobalVariableSegments error: %v", analyzeErr)
			}

			if !segmentsEqual(got, tt.want) {
				t.Errorf("GlobalVariableSegments(%q)\n  got  %v\n  want %v", tt.template, got, tt.want)
			}
		})
	}
}

func TestVariableSegments(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		wantAll  [][]string
	}{
		{
			name:     "assign: local var appears in All but not Globals",
			template: `{% assign x = src %}{{ x }}`,
			wantAll:  [][]string{{"src"}, {"x"}},
		},
		{
			name:     "for loop: loop var in All",
			template: `{% for item in list %}{{ item }}{% endfor %}`,
			wantAll:  [][]string{{"list"}, {"item"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, parseErr := engine.ParseString(tt.template)
			if parseErr != nil {
				t.Fatalf("ParseString(%q) error: %v", tt.template, parseErr)
			}

			got, analyzeErr := engine.VariableSegments(tpl)
			if analyzeErr != nil {
				t.Fatalf("VariableSegments error: %v", analyzeErr)
			}

			if !segmentsEqual(got, tt.wantAll) {
				t.Errorf("VariableSegments(%q)\n  got  %v\n  want %v", tt.template, got, tt.wantAll)
			}
		})
	}
}

func TestTemplateGlobalVariableSegments(t *testing.T) {
	engine := NewEngine()
	tpl, parseErr := engine.ParseString(`{{ user.name }}`)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	got, analyzeErr := tpl.GlobalVariableSegments()
	if analyzeErr != nil {
		t.Fatal(analyzeErr)
	}

	want := [][]string{{"user", "name"}}
	if !segmentsEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// segmentsEqual compares two [][]string sets ignoring order.
func segmentsEqual(a, b [][]string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	key := func(s []string) string {
		result := ""
		for i, seg := range s {
			if i > 0 {
				result += "\x00"
			}
			result += seg
		}
		return result
	}
	aKeys := make([]string, len(a))
	bKeys := make([]string, len(b))
	for i, s := range a {
		aKeys[i] = key(s)
	}
	for i, s := range b {
		bKeys[i] = key(s)
	}
	sort.Strings(aKeys)
	sort.Strings(bKeys)
	return reflect.DeepEqual(aKeys, bKeys)
}
