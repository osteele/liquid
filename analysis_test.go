package liquid

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/osteele/liquid/render"
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

func TestVariables(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "simple variable",
			template: `{{ x }}`,
			want:     []string{"x"},
		},
		{
			name:     "property access deduplicates root",
			template: `{{ x.a }} {{ x.b }}`,
			want:     []string{"x"},
		},
		{
			name:     "multiple roots",
			template: `{{ a.x }} {{ b.y }}`,
			want:     []string{"a", "b"},
		},
		{
			name:     "includes locally-defined",
			template: `{% assign y = x %}{{ y }}`,
			want:     []string{"x", "y"},
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

			got, err := engine.Variables(tpl)
			if err != nil {
				t.Fatalf("Variables error: %v", err)
			}

			if !stringSliceSetEqual(got, tt.want) {
				t.Errorf("Variables(%q)\n  got  %v\n  want %v", tt.template, got, tt.want)
			}
		})
	}
}

func TestGlobalVariables(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "simple variable",
			template: `{{ x }}`,
			want:     []string{"x"},
		},
		{
			name:     "deduplicates root across paths",
			template: `{{ x.a }} {{ x.b }}`,
			want:     []string{"x"},
		},
		{
			name:     "excludes locally-defined",
			template: `{% assign y = x %}{{ y }}`,
			want:     []string{"x"},
		},
		{
			name:     "for loop variable excluded",
			template: `{% for item in list %}{{ item.name }}{% endfor %}`,
			want:     []string{"list"},
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

			got, err := engine.GlobalVariables(tpl)
			if err != nil {
				t.Fatalf("GlobalVariables error: %v", err)
			}

			if !stringSliceSetEqual(got, tt.want) {
				t.Errorf("GlobalVariables(%q)\n  got  %v\n  want %v", tt.template, got, tt.want)
			}
		})
	}
}

func TestFullVariables(t *testing.T) {
	engine := NewEngine()

	t.Run("marks globals correctly", func(t *testing.T) {
		tpl, parseErr := engine.ParseString(`{% assign y = x.val %}{{ y }} {{ z }}`)
		if parseErr != nil {
			t.Fatal(parseErr)
		}

		got, err := engine.FullVariables(tpl)
		if err != nil {
			t.Fatal(err)
		}

		byName := map[string]Variable{}
		for _, v := range got {
			byName[v.String()] = v
		}

		if v, ok := byName["x.val"]; !ok || !v.Global {
			t.Errorf("expected x.val to be global, got %v", got)
		}
		if v, ok := byName["y"]; !ok || v.Global {
			t.Errorf("expected y to be local, got %v", got)
		}
		if v, ok := byName["z"]; !ok || !v.Global {
			t.Errorf("expected z to be global, got %v", got)
		}
	})

	t.Run("variable String method", func(t *testing.T) {
		v := Variable{Segments: []string{"customer", "first_name"}}
		if got := v.String(); got != "customer.first_name" {
			t.Errorf("String() = %q, want %q", got, "customer.first_name")
		}
	})
}

func TestGlobalFullVariables(t *testing.T) {
	engine := NewEngine()

	tpl, parseErr := engine.ParseString(`{% assign y = src %}{{ y }} {{ z.val }}`)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	got, err := engine.GlobalFullVariables(tpl)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range got {
		if !v.Global {
			t.Errorf("GlobalFullVariables: expected all Global=true, got %v", v)
		}
	}

	names := map[string]bool{}
	for _, v := range got {
		names[v.String()] = true
	}

	if !names["src"] {
		t.Errorf("expected src in globals, got %v", got)
	}
	if !names["z.val"] {
		t.Errorf("expected z.val in globals, got %v", got)
	}
	if names["y"] {
		t.Errorf("y should not be in globals (it is assign-defined), got %v", got)
	}
}

func TestParseAndAnalyze(t *testing.T) {
	engine := NewEngine()

	tpl, analysis, err := engine.ParseAndAnalyze([]byte(`{% assign x = src %}{{ x }} {{ z }}`))
	if err != nil {
		t.Fatalf("ParseAndAnalyze error: %v", err)
	}
	if tpl == nil {
		t.Fatal("expected non-nil template")
	}
	if analysis == nil {
		t.Fatal("expected non-nil StaticAnalysis")
	}

	gotGlobals := map[string]bool{}
	for _, v := range analysis.Globals {
		gotGlobals[v.String()] = true
	}
	if !gotGlobals["src"] || !gotGlobals["z"] {
		t.Errorf("expected src and z in Globals, got %v", analysis.Globals)
	}
	if gotGlobals["x"] {
		t.Errorf("x should not be in Globals, got %v", analysis.Globals)
	}

	localsSet := map[string]bool{}
	for _, l := range analysis.Locals {
		localsSet[l] = true
	}
	if !localsSet["x"] {
		t.Errorf("expected x in Locals, got %v", analysis.Locals)
	}
}

func TestStaticAnalysisTags(t *testing.T) {
	engine := NewEngine()

	tpl, parseErr := engine.ParseString(`{% assign x = 1 %}{% for i in list %}{{ i }}{% endfor %}{% if cond %}ok{% endif %}`)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	analysis, err := engine.Analyze(tpl)
	if err != nil {
		t.Fatal(err)
	}

	tagSet := map[string]bool{}
	for _, tag := range analysis.Tags {
		tagSet[tag] = true
	}

	for _, expected := range []string{"assign", "for", "if"} {
		if !tagSet[expected] {
			t.Errorf("expected tag %q in Tags, got %v", expected, analysis.Tags)
		}
	}
}

func TestTemplateAnalysisMethods(t *testing.T) {
	engine := NewEngine()

	tpl, err := engine.ParseString(`{% assign y = x %}{{ y }} {{ z }}`)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Variables", func(t *testing.T) {
		got, err := tpl.Variables()
		if err != nil {
			t.Fatal(err)
		}
		if !stringSliceSetEqual(got, []string{"x", "y", "z"}) {
			t.Errorf("got %v", got)
		}
	})

	t.Run("GlobalVariables", func(t *testing.T) {
		got, err := tpl.GlobalVariables()
		if err != nil {
			t.Fatal(err)
		}
		if !stringSliceSetEqual(got, []string{"x", "z"}) {
			t.Errorf("got %v", got)
		}
	})

	t.Run("FullVariables", func(t *testing.T) {
		got, err := tpl.FullVariables()
		if err != nil {
			t.Fatal(err)
		}
		if len(got) == 0 {
			t.Error("expected non-empty FullVariables")
		}
	})

	t.Run("GlobalFullVariables", func(t *testing.T) {
		got, err := tpl.GlobalFullVariables()
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range got {
			if !v.Global {
				t.Errorf("GlobalFullVariables: expected all Global=true, got %v", v)
			}
		}
	})

	t.Run("Analyze", func(t *testing.T) {
		analysis, err := tpl.Analyze()
		if err != nil {
			t.Fatal(err)
		}
		if analysis == nil {
			t.Fatal("expected non-nil StaticAnalysis")
		}
	})
}

func TestRegisterTagAnalyzer(t *testing.T) {
	engine := NewEngine()

	// Register a custom tag that outputs a variable
	engine.RegisterTag("output_var", func(_ render.Context) (string, error) {
		return "", nil
	})

	// Without analyzer, the variable won't be tracked
	tpl, parseErr := engine.ParseString(`{% output_var myvar %}`)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	got, err := engine.GlobalVariableSegments(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected no variables before analyzer, got %v", got)
	}
}

// stringSliceSetEqual compares two string slices as sets (order-independent).
func stringSliceSetEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	setA := map[string]bool{}
	for _, s := range a {
		setA[s] = true
	}
	for _, s := range b {
		if !setA[s] {
			return false
		}
	}
	return true
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
		var result strings.Builder
		for i, seg := range s {
			if i > 0 {
				result.WriteString("\x00")
			}
			result.WriteString(seg)
		}
		return result.String()
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
