package generics

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

var convertTests = []struct {
	value, proto, expected interface{}
}{
	{1, 1.0, float64(1)},
	{"2", 1, int(2)},
	{"1.2", 1.0, float64(1.2)},
	{true, 1, 1},
	{false, 1, 0},
	{1, "", "1"},
	{false, "", "false"},
	{true, "", "true"},
	{"string", "", "string"},
}

var eqTests = []struct {
	a, b     interface{}
	expected bool
}{
	{nil, nil, true},
	{nil, 1, false},
	{1, nil, false},
	{false, false, true},
	{false, true, false},
	{0, 1, false},
	{1, 1, true},
	{1.0, 1.0, true},
	{1, 1.0, true},
	{1, 2.0, false},
	{1.0, 1, true},
	{"a", "b", false},
	{"a", "a", true},
	{[]string{"a"}, []string{"a"}, true},
	{[]string{"a"}, []string{"a", "b"}, false},
	{[]string{"a", "b"}, []string{"a"}, false},
	{[]string{"a", "b"}, []string{"a", "b"}, true},
	{[]string{"a", "b"}, []string{"a", "c"}, false},
	{[]interface{}{1.0, 2}, []interface{}{1, 2.0}, true},
}

var lessTests = []struct {
	a, b     interface{}
	expected bool
}{
	{nil, nil, false},
	{false, true, true},
	{false, false, false},
	{false, nil, false},
	{nil, false, false},
	{0, 1, true},
	{1, 0, false},
	{1, 1, false},
	{1, 2.1, true},
	{1.1, 2, true},
	{2.1, 1, false},
	{"a", "b", true},
	{"b", "a", false},
	{[]string{"a"}, []string{"a"}, false},
}

func TestCall(t *testing.T) {
	fn := func(a, b string) string {
		return a + "," + b + "."
	}
	args := []interface{}{5, 10}
	value, err := Call(reflect.ValueOf(fn), args)
	require.NoError(t, err)
	require.Equal(t, "5,10.", value)
}
func TestConvert(t *testing.T) {
	for i, test := range convertTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			typ := reflect.TypeOf(test.proto)
			value, err := Convert(test.value, typ)
			name := fmt.Sprintf("Convert %#v -> %v", test.value, typ)
			require.NoErrorf(t, err, name)
			require.Equalf(t, test.expected, value, name)
		})
	}
}
func TestConvertMap(t *testing.T) {
	m := map[interface{}]interface{}{"key": "value"}
	typ := reflect.TypeOf(map[string]string{})
	a, err := Convert(m, typ)
	require.NoError(t, err)
	switch a := a.(type) {
	case map[string]string:
		require.Len(t, a, 1)
		require.Equal(t, "value", a["key"])
	default:
		require.Equal(t, typ.String(), reflect.TypeOf(a).String())
	}
}

func TestConvertMapSynonym(t *testing.T) {
	type VariableMap map[interface{}]interface{}
	m := VariableMap{"key": "value"}
	typ := reflect.TypeOf(map[string]string{})
	a, err := Convert(m, typ)
	require.NoError(t, err)
	switch a := a.(type) {
	case map[string]string:
		require.Len(t, a, 1)
		require.Equal(t, "value", a["key"])
	default:
		require.Equal(t, typ.String(), reflect.TypeOf(a).String())
	}
}

func TestEqual(t *testing.T) {
	for i, test := range eqTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			value := Equal(test.a, test.b)
			require.Equalf(t, test.expected, value, "%#v == %#v", test.a, test.b)
		})
	}
}

func TestLess(t *testing.T) {
	for i, test := range lessTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			value := Less(test.a, test.b)
			require.Equalf(t, test.expected, value, "%#v < %#v", test.a, test.b)
		})
	}
}

func TestLength(t *testing.T) {
	require.Equal(t, 3, Length([]int{1, 2, 3}))
	require.Equal(t, 3, Length("abc"))
	require.Equal(t, 0, Length(map[string]int{"a": 1}))
}
