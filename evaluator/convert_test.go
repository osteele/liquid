package evaluator

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type redConvertible struct{}

func (c redConvertible) ToLiquid() interface{} {
	return "red"
}

func timeMustParse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

var convertTests = []struct {
	value, proto, expected interface{}
}{
	{1, 1.0, float64(1)},
	{"2", 1, int(2)},
	{"1.2", 1.0, float64(1.2)},
	{true, 1, 1},
	{false, 1, 0},
	{nil, true, false},
	{0, true, true},
	{"", true, true},
	{1, "", "1"},
	{false, "", "false"},
	{true, "", "true"},
	{"string", "", "string"},
	{[]int{1, 2}, []string{}, []string{"1", "2"}},
	{"March 14, 2016", time.Now(), timeMustParse("2016-03-14T00:00:00Z")},
	{redConvertible{}, "", "red"},
}

func TestConvert(t *testing.T) {
	for i, test := range convertTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			typ := reflect.TypeOf(test.proto)
			name := fmt.Sprintf("Convert %#v -> %v", test.value, typ)
			value, err := Convert(test.value, typ)
			require.NoErrorf(t, err, name)
			require.Equalf(t, test.expected, value, name)
		})
	}
}
func TestConvert_map(t *testing.T) {
	typ := reflect.TypeOf(map[string]string{})
	v, err := Convert(map[interface{}]interface{}{"key": "value"}, typ)
	require.NoError(t, err)
	m, ok := v.(map[string]string)
	require.True(t, ok)
	require.Equal(t, "value", m["key"])
}

func TestConvert_map_synonym(t *testing.T) {
	type VariableMap map[interface{}]interface{}
	typ := reflect.TypeOf(map[string]string{})
	v, err := Convert(VariableMap{"key": "value"}, typ)
	require.NoError(t, err)
	m, ok := v.(map[string]string)
	require.True(t, ok)
	require.Equal(t, "value", m["key"])
}

func TestConvert_map_to_array(t *testing.T) {
	typ := reflect.TypeOf([]string{})
	v, err := Convert(map[int]string{1: "b", 2: "a"}, typ)
	require.NoError(t, err)
	array, ok := v.([]string)
	require.True(t, ok)
	sort.Strings(array)
	require.Equal(t, []string{"a", "b"}, array)
}

// func TestConvert_ptr(t *testing.T) {
// 	typ := reflect.PtrTo(reflect.TypeOf(""))
// 	v, err := Convert("a", typ)
// 	require.NoError(t, err)
// 	ptr, ok := v.(*string)
// 	fmt.Printf("%#v %T\n", v, v)
// 	require.True(t, ok)
// 	require.NotNil(t, ptr)
// 	require.Equal(t, "ab", *ptr)
// }
