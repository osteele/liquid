package evaluator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValue(t *testing.T) {
	var (
		nv = ValueOf(nil)
		// fv = ValueOf(false)
		// tv = ValueOf(true)
		iv = ValueOf(123)
		// fv = ValueOf(123.0)
	)
	require.Equal(t, 123, iv.Int())
	require.Panics(t, func() { nv.Int() })
	require.True(t, iv.Equal(ValueOf(123)))
	require.True(t, iv.Equal(ValueOf(123.0)))
	require.True(t, iv.Less(ValueOf(200)))
}

func TestValue_array(t *testing.T) {
	av := ValueOf([]string{"first", "second", "third"})
	require.True(t, av.Contains(ValueOf("first")))
	require.False(t, av.Contains(ValueOf("missing")))
	require.Equal(t, "first", av.IndexValue(ValueOf(0)).Interface())
	require.Equal(t, "third", av.IndexValue(ValueOf(-1)).Interface())
	require.Equal(t, "first", av.PropertyValue(ValueOf("first")).Interface())
	require.Equal(t, "third", av.PropertyValue(ValueOf("last")).Interface())
	require.Equal(t, 3, av.PropertyValue(ValueOf("size")).Interface())

	empty := ValueOf([]string{})
	require.Equal(t, nil, empty.IndexValue(ValueOf(0)).Interface())
	require.Equal(t, nil, empty.IndexValue(ValueOf(-1)).Interface())
	require.Equal(t, nil, empty.PropertyValue(ValueOf("first")).Interface())
	require.Equal(t, nil, empty.PropertyValue(ValueOf("last")).Interface())
	require.Equal(t, nil, empty.PropertyValue(ValueOf("length")).Interface())
}

func TestValue_string(t *testing.T) {
	av := ValueOf("seafood")
	require.True(t, av.Contains(ValueOf("foo")))
	require.False(t, av.Contains(ValueOf("bar")))
	require.Equal(t, 7, av.PropertyValue(ValueOf("size")).Interface())
}

func TestValue_hash(t *testing.T) {
	h := ValueOf(map[string]interface{}{"key": "value"})
	require.Equal(t, "value", h.IndexValue(ValueOf("key")).Interface())
	require.Equal(t, nil, h.IndexValue(ValueOf("missing_key")).Interface())
	require.Equal(t, 1, h.PropertyValue(ValueOf("size")).Interface())

	withSizeKey := ValueOf(map[string]interface{}{"size": "value"})
	require.Equal(t, "value", withSizeKey.IndexValue(ValueOf("size")).Interface())
}
