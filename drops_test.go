package liquid

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

type dropTest struct{}

func (d dropTest) ToLiquid() interface{} { return "drop" }

func TestDrops(t *testing.T) {
	require.Equal(t, "drop", FromDrop(dropTest{}))

	require.Equal(t, "not a drop", FromDrop("not a drop"))
}

type redConvertible struct{}

func (c redConvertible) ToLiquid() interface{} {
	return map[string]interface{}{
		"color": "red",
	}
}

func ExampleDrop_map() {
	// type redConvertible struct{}
	//
	// func (c redConvertible) ToLiquid() interface{} {
	// 	return map[string]interface{}{
	// 		"color": "red",
	// 	}
	// }
	engine := NewEngine()
	bindings := map[string]interface{}{
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

func (c car) ToLiquid() interface{} {
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
	// func (c car) ToLiquid() interface{} {
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
	bindings := map[string]interface{}{
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
