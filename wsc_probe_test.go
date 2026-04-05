package liquid_test

import (
	"fmt"
	"testing"

	"github.com/osteele/liquid"
)

func TestWSC_Probe(t *testing.T) {
	eng := liquid.NewEngine()
	r := func(src string, binds map[string]any) string {
		out, err := eng.ParseAndRenderString(src, binds)
		if err != nil {
			return "ERR: " + err.Error()
		}
		return out
	}
	arr := map[string]any{"arr": []int{1, 2, 3}}

	cases := []struct {
		name, tpl, want string
		binds           map[string]any
	}{
		{"for trim both", "{%- for i in arr -%}{{ i }}{%- endfor -%}", "123", arr},
		{"for trim around content", "\n{%- for i in arr -%}\n  {{ i }}\n{%- endfor -%}\n", "123", arr},
		{"for trim right open only", "{% for i in arr -%}\n{{ i }}\n{% endfor %}", "1\n2\n3\n", arr},
		{"nested if in for", "{%- for i in arr -%}{%- if i > 1 -%},{{ i }}{%- endif -%}{%- endfor -%}", ",2,3", arr},
		{"for no trim keeps ws", "{% for i in arr %}\n  item: {{ i }}\n{% endfor %}", "\n  item: 1\n\n  item: 2\n", map[string]any{"arr": []int{1, 2}}},
		{"for else trim", "{%- for i in arr -%}{{ i }}{%- else -%}empty{%- endfor -%}", "empty", map[string]any{"arr": []int{}}},
		{"trim assign", "a\n{%- assign x = 1 -%}\nb", "ab", nil},
		{"trim capture", "{%- capture x -%}  hello  {%- endcapture -%}[{{ x }}]", "[hello]", nil},
		{"trim comment", "a\n{%- # comment -%}\nb", "ab", nil},
		{"for body pre-trim per iter", "{%- for i in arr %}\n  {{ i }}\n{%- endfor -%}", "\n  1\n  2\n  3", arr},
		{"deeply nested if-for trim", "{%- for i in arr -%}{%- for j in arr -%}{%- if i == j -%}{{ i }}{%- endif -%}{%- endfor -%}{%- endfor -%}", "123", arr},
		{"trim on unless", "a\n{%- unless false -%}b{%- endunless -%}\nc", "abc", nil},
		{"trim around case", "a\n{%- case v -%}{%- when 1 -%}one{%- endcase -%}\nb", "aoneb", map[string]any{"v": 1}},
		{"trim for with text body", "list:\n{%- for i in arr -%}\n- {{ i }}\n{%- endfor %}", "list:- 1- 2- 3", arr},
		{"for first/last trim mix", "{% for i in arr -%}\n{{ i }}{% unless forloop.last %},{% endunless %}\n{%- endfor %}", "1,2,3", arr},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := r(c.tpl, c.binds)
			if got != c.want {
				t.Errorf("\n  got:  %q\n  want: %q", got, c.want)
			} else {
				fmt.Printf("PASS: %s\n", c.name)
			}
		})
	}
}
