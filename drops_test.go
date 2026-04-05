package liquid

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

type dropTest struct{}

func (d dropTest) ToLiquid() any { return "drop" }

func TestDrops(t *testing.T) {
	require.Equal(t, "drop", FromDrop(dropTest{}))

	require.Equal(t, "not a drop", FromDrop("not a drop"))
}

type redConvertible struct{}

func (c redConvertible) ToLiquid() any {
	return map[string]any{
		"color": "red",
	}
}

func ExampleDrop_map() {
	// type redConvertible struct{}
	//
	// func (c redConvertible) ToLiquid() any {
	// 	return map[string]any{
	// 		"color": "red",
	// 	}
	// }
	engine := NewEngine()
	bindings := map[string]any{
		"car": redConvertible{},
	}
	template := `{{ car.color }}`

	out, err := engine.ParseAndRenderString(template, bindings)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(out)
	// Output: red
}

type car struct{ color, model string }

func (c car) ToLiquid() any {
	return carDrop{c.model, c.color}
}

type carDrop struct {
	Model string
	Color string `liquid:"color"`
}

func (c carDrop) Drive() string {
	return "AWD"
}

func ExampleDrop_struct() {
	// type car struct{ color, model string }
	//
	// func (c car) ToLiquid() any {
	// 	return carDrop{c.model, c.color}
	// }
	//
	// type carDrop struct {
	// 	Model string
	// 	Color string `liquid:"color"`
	// }
	//
	// func (c carDrop) Drive() string {
	// 	return "AWD"
	// }
	engine := NewEngine()
	bindings := map[string]any{
		"car": car{"blue", "S85"},
	}
	template := `{{ car.color }} {{ car.Drive }} Model {{ car.Model }}`

	out, err := engine.ParseAndRenderString(template, bindings)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(out)
	// Output: blue AWD Model S85
}

// ---------------------------------------------------------------------------
// DropMethodMissing — integration tests
// Ported from:
//   - Ruby Liquid: test/unit/drop_test.rb (liquid_method_missing)
//   - LiquidJS:    test/integration/drop/liquidMethodMissing.spec.ts
// ---------------------------------------------------------------------------

// dynamicDrop exposes a fixed field and delegates unknown lookups to a map.
type dynamicDrop struct {
	Name    string
	dynamic map[string]any
}

func (d dynamicDrop) MissingMethod(key string) any {
	return d.dynamic[key]
}

func TestDropMethodMissing_knownFieldNotIntercepted(t *testing.T) {
	// Defined struct fields/methods take priority — MissingMethod is NOT called.
	engine := NewEngine()
	bindings := map[string]any{
		"obj": dynamicDrop{Name: "Alice", dynamic: map[string]any{"Name": "Shadow"}},
	}
	out, err := engine.ParseAndRenderString(`{{ obj.Name }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "Alice", out)
}

func TestDropMethodMissing_dynamicProperty(t *testing.T) {
	// Unknown keys are dispatched to MissingMethod.
	engine := NewEngine()
	bindings := map[string]any{
		"obj": dynamicDrop{dynamic: map[string]any{"color": "red", "count": 3}},
	}

	out, err := engine.ParseAndRenderString(`{{ obj.color }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "red", out)

	out, err = engine.ParseAndRenderString(`{{ obj.count }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "3", out)
}

func TestDropMethodMissing_missingKeyReturnsEmpty(t *testing.T) {
	// MissingMethod returning nil renders as empty string (not an error).
	engine := NewEngine()
	bindings := map[string]any{
		"obj": dynamicDrop{dynamic: map[string]any{}},
	}
	out, err := engine.ParseAndRenderString(`{{ obj.whatever }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "", out)
}

func TestDropMethodMissing_usableInCondition(t *testing.T) {
	engine := NewEngine()
	bindings := map[string]any{
		"obj": dynamicDrop{dynamic: map[string]any{"active": true}},
	}
	out, err := engine.ParseAndRenderString(`{% if obj.active %}yes{% endif %}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "yes", out)
}

func ExampleDropMethodMissing() {
	// type dynamicDrop struct {
	//     Name    string
	//     dynamic map[string]any
	// }
	//
	// func (d dynamicDrop) MissingMethod(key string) any {
	//     return d.dynamic[key]
	// }

	engine := NewEngine()
	bindings := map[string]any{
		"product": dynamicDrop{
			Name:    "Widget",
			dynamic: map[string]any{"price": 9.99, "sku": "W-001"},
		},
	}

	out, err := engine.ParseAndRenderString(`{{ product.Name }} — SKU: {{ product.sku }}, price: {{ product.price }}`, bindings)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(out)
	// Output: Widget — SKU: W-001, price: 9.99
}

// ---------------------------------------------------------------------------
// ContextDrop — context injection tests
// Ported from:
//   - Ruby Liquid: lib/liquid/context.rb (context= setter)
//     test/integration/standard_filter_test.rb: test_map_calls_context=
//   - LiquidJS: test/e2e/drop.spec.ts (expose context in liquidMethodMissing)
// ---------------------------------------------------------------------------

// scopeDrop reads another variable from the rendering scope via ContextDrop.
// This mirrors the Ruby pattern where a Drop accesses @context to read registers
// or scope variables set by the calling template.
// Note: scopeDrop does NOT implement ToLiquid — it exposes fields directly as
// a struct drop. ContextSetter drops that return self from ToLiquid would cause
// infinite recursion in dropWrapper.Resolve.
type scopeDrop struct {
	watchKey string
	ctx      DropRenderContext
}

func (d *scopeDrop) SetContext(ctx DropRenderContext) { d.ctx = ctx }

func (d *scopeDrop) Observed() any {
	if d.ctx == nil {
		return nil
	}
	return d.ctx.Get(d.watchKey)
}

func TestContextDrop_basic(t *testing.T) {
	// Drop receives the current scope and can read other variables.
	// Analogous to Ruby's context= setter being called on variable lookup.
	engine := NewEngine()
	drop := &scopeDrop{watchKey: "other"}
	bindings := map[string]any{
		"probe": drop,
		"other": "hello",
	}
	out, err := engine.ParseAndRenderString(`{{ probe.Observed }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "hello", out)
}

func TestContextDrop_canWriteToScope(t *testing.T) {
	// Confirm SetContext is called when a plain struct drop is looked up.
	// The scopeDrop.ctx will be populated by the time Observed() is called.
	engine := NewEngine()
	bindings := map[string]any{
		"probe": &scopeDrop{watchKey: "x"},
		"x":     99,
	}
	out, err := engine.ParseAndRenderString(`{{ probe.Observed }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "99", out)
}

func TestContextDrop_injectedBeforePropertyAccess(t *testing.T) {
	// SetContext is called before any property/method is invoked on the drop,
	// so Observed() can rely on ctx being populated.
	engine := NewEngine()
	bindings := map[string]any{
		"sensor": &scopeDrop{watchKey: "value"},
		"value":  "world",
	}
	out, err := engine.ParseAndRenderString(`{{ sensor.Observed }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "world", out)
}

func TestContextDrop_seesAssignedVariables(t *testing.T) {
	// Drop can read variables set by {% assign %} earlier in the template.
	engine := NewEngine()
	bindings := map[string]any{
		"probe": &scopeDrop{watchKey: "dynamic"},
	}
	out, err := engine.ParseAndRenderString(
		`{% assign dynamic = "assigned" %}{{ probe.Observed }}`,
		bindings,
	)
	require.NoError(t, err)
	require.Equal(t, "assigned", out)
}

func ExampleContextDrop() {
	// type scopeDrop struct {
	//     watchKey string
	//     ctx      liquid.DropRenderContext
	// }
	//
	// func (d *scopeDrop) SetContext(ctx liquid.DropRenderContext) { d.ctx = ctx }
	// func (d *scopeDrop) Observed() any { return d.ctx.Get(d.watchKey) }

	engine := NewEngine()
	bindings := map[string]any{
		"probe": &scopeDrop{watchKey: "user"},
		"user":  "Alice",
	}
	out, err := engine.ParseAndRenderString(`{{ probe.Observed }}`, bindings)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(out)
	// Output: Alice
}
