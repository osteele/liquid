package liquid_test

import (
	"fmt"
	"testing"

	"github.com/osteele/liquid"
)

func TestProbeSection4(t *testing.T) {
	eng := liquid.NewEngine()

	probe := func(tpl string) {
		out, err := eng.ParseAndRenderString(tpl, nil)
		fmt.Printf("tpl=%-60s got=%q err=%v\n", tpl, out, err)
	}

	probe("{% if (1..5) contains 3 %}yes{% else %}no{% endif %}")
	probe("{% if (1..5) contains 6 %}yes{% else %}no{% endif %}")
	probe("{% if null <= 0 %} true {% else %} false {% endif %}")
	probe("{% if 0 <= null %} true {% else %} false {% endif %}")
}
