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
