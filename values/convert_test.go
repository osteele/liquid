package values

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/stretchr/testify/require"
)

type redConvertible struct{}

func (c redConvertible) ToLiquid() interface{} {
	return "red"
}

var convertTests = []struct {
	value, expected interface{}
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
	{[]interface{}{1, 2}, []interface{}{1, 2}},
	{[]int{1, 2}, []int{1, 2}},
	{[]int{1, 2}, []interface{}{1, 2}},
	{[]interface{}{1, 2}, []int{1, 2}},
	{[]int{1, 2}, []string{"1", "2"}},
	{yaml.MapSlice{{Key: 1, Value: 1}}, []interface{}{1}},
	{yaml.MapSlice{{Key: 1, Value: 1}}, []string{"1"}},
	{yaml.MapSlice{{Key: 1, Value: "a"}}, []string{"a"}},
	{yaml.MapSlice{{Key: 1, Value: "a"}}, map[interface{}]interface{}{1: "a"}},
	{yaml.MapSlice{{Key: 1, Value: "a"}}, map[int]string{1: "a"}},
	{yaml.MapSlice{{Key: 1, Value: "a"}}, map[string]string{"1": "a"}},
	{yaml.MapSlice{{Key: "a", Value: 1}}, map[string]string{"a": "1"}},
	{yaml.MapSlice{{Key: "a", Value: nil}}, map[string]interface{}{"a": nil}},
	{yaml.MapSlice{{Key: nil, Value: 1}}, map[interface{}]string{nil: "1"}},
	// {"March 14, 2016", time.Now(), timeMustParse("2016-03-14T00:00:00Z")},
	{redConvertible{}, "red"},
	{int32(1540918800), func() time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05 -0700", "2018-10-31 00:00:00 +0700")
		return t
	}()},
	{int64(1540918800), func() time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05 -0700", "2018-10-31 00:00:00 +0700")
		return t
	}()},
	{float64(1540918800), func() time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05 -0700", "2018-10-31 00:00:00 +0700")
		return t
	}()},
	{"1540918800", func() time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05 -0700", "2018-10-31 00:00:00 +0700")
		return t
	}()},
	{"2018-10-31 00:00:00 +0000", func() time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05 -0700", "2018-10-31 00:00:00 +0000")
		return t
	}()},
}

var convertErrorTests = []struct {
	value, proto interface{}
	expected     []string
}{
	{map[string]bool{"k": true}, map[int]bool{}, []string{"map key"}},
	{map[string]string{"k": "v"}, map[string]int{}, []string{"map element"}},
	{map[interface{}]interface{}{"k": "v"}, map[string]int{}, []string{"map element"}},
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

func TestMustConvert(t *testing.T) {
	typ := reflect.TypeOf("")
	v := MustConvert(2, typ)
	require.Equal(t, "2", v)

	typ = reflect.TypeOf(2)
	require.Panics(t, func() { MustConvert("x", typ) })
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
