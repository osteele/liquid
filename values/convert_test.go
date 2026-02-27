package values

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/stretchr/testify/require"
)

type redConvertible struct{}

func (c redConvertible) ToLiquid() any {
	return "red"
}

var convertTests = []struct {
	value, expected any
}{
	{nil, false},
	{false, 0},
	{false, int(0)},
	{false, int8(0)},
	{false, int16(0)},
	{false, int32(0)},
	{false, int64(0)},
	{false, uint(0)},
	{false, uint8(0)},
	{false, uint16(0)},
	{false, uint32(0)},
	{false, uint64(0)},
	{true, 1},
	{true, int(1)},
	{true, int8(1)},
	{true, int16(1)},
	{true, int32(1)},
	{true, int64(1)},
	{true, uint(1)},
	{true, uint8(1)},
	{true, uint16(1)},
	{true, uint32(1)},
	{true, uint64(1)},
	{false, false},
	{true, true},
	{true, "true"},
	{false, "false"},
	{0, true},
	{2, 2},
	{2, "2"},
	{2, 2.0},
	{"", true},
	{"2", int(2)},
	{"2", int8(2)},
	{"2", int16(2)},
	{"2", int32(2)},
	{"2", int64(2)},
	{"2", uint(2)},
	{"2", uint8(2)},
	{"2", uint16(2)},
	{"2", uint32(2)},
	{"2", uint64(2)},
	{"2", 2},
	{"2", 2.0},
	{"2.0", 2.0},
	{"2.1", 2.1},
	{"2.1", float32(2.1)},
	{"2.1", float64(2.1)},
	{"string", "string"},
	{[]any{1, 2}, []any{1, 2}},
	{[]int{1, 2}, []int{1, 2}},
	{[]int{1, 2}, []any{1, 2}},
	{[]any{1, 2}, []int{1, 2}},
	{[]int{1, 2}, []string{"1", "2"}},
	{yaml.MapSlice{{Key: 1, Value: 1}}, []any{1}},
	{yaml.MapSlice{{Key: 1, Value: 1}}, []string{"1"}},
	{yaml.MapSlice{{Key: 1, Value: "a"}}, []string{"a"}},
	{yaml.MapSlice{{Key: 1, Value: "a"}}, map[any]any{1: "a"}},
	{yaml.MapSlice{{Key: 1, Value: "a"}}, map[int]string{1: "a"}},
	{yaml.MapSlice{{Key: 1, Value: "a"}}, map[string]string{"1": "a"}},
	{yaml.MapSlice{{Key: "a", Value: 1}}, map[string]string{"a": "1"}},
	{yaml.MapSlice{{Key: "a", Value: nil}}, map[string]any{"a": nil}},
	{yaml.MapSlice{{Key: nil, Value: 1}}, map[any]string{nil: "1"}},
	{Range{1, 5}, []any{1, 2, 3, 4, 5}},
	{Range{0, 0}, []any{0}},
	// {"March 14, 2016", time.Now(), timeMustParse("2016-03-14T00:00:00Z")},
	{redConvertible{}, "red"},
}

var convertErrorTests = []struct {
	value, proto any
	expected     []string
}{
	{map[string]bool{"k": true}, map[int]bool{}, []string{"map key"}},
	{map[string]string{"k": "v"}, map[string]int{}, []string{"map element"}},
	{map[any]any{"k": "v"}, map[string]int{}, []string{"map element"}},
	{"notanumber", int(0), []string{"can't convert string", "to type int"}},
	{"notanumber", uint(0), []string{"can't convert string", "to type uint"}},
	{"notanumber", float64(0), []string{"can't convert string", "to type float64"}},
}

func TestConvert(t *testing.T) {
	for i, test := range convertTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			typ := reflect.TypeOf(test.expected)
			name := fmt.Sprintf("Convert %#v -> %v", test.value, typ)
			value, err := Convert(test.value, typ)
			require.NoErrorf(t, err, name)
			require.Equalf(t, test.expected, value, name)
		})
	}
}

func TestConvert_errors(t *testing.T) {
	for i, test := range convertErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			typ := reflect.TypeOf(test.proto)
			name := fmt.Sprintf("Convert %#v -> %v", test.value, typ)
			_, err := Convert(test.value, typ)
			require.Errorf(t, err, name)

			for _, expected := range test.expected {
				require.Containsf(t, err.Error(), expected, name)
			}
		})
	}
}

func TestConvert_map(t *testing.T) {
	typ := reflect.TypeOf(map[string]string{})
	v, err := Convert(map[any]any{"key": "value"}, typ)
	require.NoError(t, err)

	m, ok := v.(map[string]string)
	require.True(t, ok)
	require.Equal(t, "value", m["key"])
}

func TestConvert_map_synonym(t *testing.T) {
	type VariableMap map[any]any

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

func TestMustConvert(t *testing.T) {
	typ := reflect.TypeOf("")
	v := MustConvert(2, typ)
	require.Equal(t, "2", v)

	typ = reflect.TypeOf(2)

	require.Panics(t, func() { MustConvert("x", typ) })
}

func TestConvert_intToTime(t *testing.T) {
	typ := reflect.TypeOf(time.Time{})

	// int -> time.Time
	v, err := Convert(int(1152098955), typ)
	require.NoError(t, err)
	require.Equal(t, time.Unix(1152098955, 0), v)

	// int64 -> time.Time
	v, err = Convert(int64(1152098955), typ)
	require.NoError(t, err)
	require.Equal(t, time.Unix(1152098955, 0), v)

	// float64 -> time.Time
	v, err = Convert(float64(1152098955), typ)
	require.NoError(t, err)
	require.Equal(t, time.Unix(1152098955, 0), v)

	// string timestamp -> time.Time
	v, err = Convert("1152098955", typ)
	require.NoError(t, err)
	require.Equal(t, time.Unix(1152098955, 0), v)
}

func TestConvert_stringToFloat(t *testing.T) {
	typ := reflect.TypeOf(float64(0))
	v, err := Convert("3.14", typ)
	require.NoError(t, err)
	require.Equal(t, 3.14, v)
}

func TestConvert_jsonNumberToFloat(t *testing.T) {
	typ := reflect.TypeOf(float64(0))
	v, err := Convert(json.Number("2.718"), typ)
	require.NoError(t, err)
	require.Equal(t, 2.718, v)
}

func TestConvert_jsonNumberToInt(t *testing.T) {
	typ := reflect.TypeOf(int(0))
	v, err := Convert(json.Number("42"), typ)
	require.NoError(t, err)
	require.Equal(t, int(42), v)
}

func TestMustConvertItem(t *testing.T) {
	v := MustConvertItem(2, []string{})
	require.Equal(t, "2", v)

	require.Panics(t, func() { MustConvertItem("x", []int{}) })
}

func timeMustParse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	return t
}
