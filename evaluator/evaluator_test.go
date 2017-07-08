package evaluator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestContains(t *testing.T) {
	require.True(t, Contains([]int{1, 2}, 2))
	require.False(t, Contains([]int{1, 2}, 3))
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

func TestSort(t *testing.T) {
	array := []interface{}{2, 1}
	Sort(array)
	require.Equal(t, []interface{}{1, 2}, array)

	array = []interface{}{"b", "a"}
	Sort(array)
	require.Equal(t, []interface{}{"a", "b"}, array)

	array = []interface{}{
		map[string]interface{}{"key": 20},
		map[string]interface{}{"key": 10},
		map[string]interface{}{},
	}
	SortByProperty(array, "key", true)
	require.Equal(t, nil, array[0].(map[string]interface{})["key"])
	require.Equal(t, 10, array[1].(map[string]interface{})["key"])
	require.Equal(t, 20, array[2].(map[string]interface{})["key"])
}
