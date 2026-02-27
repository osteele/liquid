package values

import (
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/stretchr/testify/require"
)

func TestValue_Interface(t *testing.T) {
	nv := ValueOf(nil)
	iv := ValueOf(123)

	var ni *int = nil

	require.Nil(t, nv.Interface())
	require.Equal(t, true, ValueOf(true).Interface())
	require.Equal(t, false, ValueOf(false).Interface())
	require.Equal(t, 123, iv.Interface())
	require.Equal(t, ValueOf(ni), nilValue)
}

func TestValue_Equal(t *testing.T) {
	iv := ValueOf(123)
	require.True(t, iv.Equal(ValueOf(123)))
	require.True(t, iv.Equal(ValueOf(123.0)))
}

func TestValue_Less(t *testing.T) {
	iv := ValueOf(123)
	require.False(t, iv.Less(ValueOf(100)))
	require.True(t, iv.Less(ValueOf(200)))
	require.False(t, iv.Less(ValueOf(100.5)))
	require.True(t, iv.Less(ValueOf(200.5)))

	sv := ValueOf("b")
	require.False(t, sv.Less(ValueOf("a")))
	require.True(t, sv.Less(ValueOf("c")))
}

func TestValue_Int(t *testing.T) {
	nv := ValueOf(nil)
	iv := ValueOf(123)
	require.Equal(t, 123, iv.Int())
	require.Panics(t, func() { nv.Int() })
}

func TestValue_IndexValue(t *testing.T) {
	require.Nil(t, ValueOf(nil).PropertyValue(ValueOf("first")).Interface())
	require.Nil(t, ValueOf(false).PropertyValue(ValueOf("first")).Interface())
	require.Nil(t, ValueOf(12).PropertyValue(ValueOf("first")).Interface())

	// empty array
	empty := ValueOf([]string{})
	require.Nil(t, empty.IndexValue(ValueOf(0)).Interface())
	require.Nil(t, empty.IndexValue(ValueOf(-1)).Interface())

	// array
	lv := ValueOf([]string{"first", "second", "third"})
	require.Equal(t, "first", lv.IndexValue(ValueOf(0)).Interface())
	require.Equal(t, "third", lv.IndexValue(ValueOf(-1)).Interface())
	require.Equal(t, "second", lv.IndexValue(ValueOf(1.0)).Interface())
	require.Equal(t, "second", lv.IndexValue(ValueOf(1.1)).Interface())
	require.Nil(t, lv.IndexValue(ValueOf(nil)).Interface())

	// string map
	hv := ValueOf(map[string]any{"key": "value"})
	require.Equal(t, "value", hv.IndexValue(ValueOf("key")).Interface())
	require.Nil(t, hv.IndexValue(ValueOf("missing_key")).Interface())
	require.Nil(t, hv.IndexValue(ValueOf(nil)).Interface())

	// interface map
	hv = ValueOf(map[any]any{"key": "value"})
	require.Equal(t, "value", hv.IndexValue(ValueOf("key")).Interface())
	require.Nil(t, hv.IndexValue(ValueOf(nil)).Interface())
	require.Nil(t, hv.IndexValue(ValueOf([]string{})).Interface())
	require.Nil(t, hv.IndexValue(ValueOf(struct{}{})).Interface())

	// ptr to map
	hashPtr := ValueOf(&map[string]any{"key": "value"})
	require.Equal(t, "value", hashPtr.IndexValue(ValueOf("key")).Interface())
	require.Nil(t, hashPtr.IndexValue(ValueOf("missing_key")).Interface())
	require.Nil(t, hashPtr.IndexValue(ValueOf(nil)).Interface())

	// MapSlice
	msv := ValueOf(yaml.MapSlice{{Key: "key", Value: "value"}})
	require.Equal(t, "value", msv.IndexValue(ValueOf("key")).Interface())
	require.Nil(t, msv.IndexValue(ValueOf("missing_key")).Interface())
	require.Nil(t, msv.IndexValue(ValueOf(nil)).Interface())
}

func TestValue_PropertyValue(t *testing.T) {
	// empty array
	empty := ValueOf([]string{})
	require.Nil(t, empty.PropertyValue(ValueOf("first")).Interface())
	require.Nil(t, empty.PropertyValue(ValueOf("last")).Interface())

	// array
	lv := ValueOf([]string{"first", "second", "third"})
	require.Equal(t, "first", lv.PropertyValue(ValueOf("first")).Interface())
	require.Equal(t, "third", lv.PropertyValue(ValueOf("last")).Interface())
	require.Nil(t, lv.PropertyValue(ValueOf(nil)).Interface())

	// string map
	hv := ValueOf(map[string]any{"key": "value"})
	require.Equal(t, "value", hv.PropertyValue(ValueOf("key")).Interface())
	require.Nil(t, hv.PropertyValue(ValueOf("missing_key")).Interface())
	require.Nil(t, hv.PropertyValue(ValueOf(nil)).Interface())

	// interface map
	hv = ValueOf(map[any]any{"key": "value"})
	require.Equal(t, "value", hv.PropertyValue(ValueOf("key")).Interface())

	// ptr to map
	hashPtr := ValueOf(&map[string]any{"key": "value"})
	require.Equal(t, "value", hashPtr.PropertyValue(ValueOf("key")).Interface())
	require.Nil(t, hashPtr.PropertyValue(ValueOf("missing_key")).Interface())

	// MapSlice
	msv := ValueOf(yaml.MapSlice{{Key: "key", Value: "value"}})
	require.Equal(t, "value", msv.PropertyValue(ValueOf("key")).Interface())
	require.Nil(t, msv.PropertyValue(ValueOf("missing_key")).Interface())
	require.Nil(t, msv.PropertyValue(ValueOf(nil)).Interface())
}

func TestValue_Contains(t *testing.T) {
	// array
	require.True(t, ValueOf([]int{1, 2}).Contains(ValueOf(2)))
	require.False(t, ValueOf([]int{1, 2}).Contains(ValueOf(3)))

	av := ValueOf([]string{"first", "second", "third"})
	require.True(t, av.Contains(ValueOf("first")))
	require.False(t, av.Contains(ValueOf("missing")))
	require.False(t, av.Contains(ValueOf(nil)))

	require.True(t, ValueOf([]any{nil}).Contains(ValueOf(nil)))

	// string
	sv := ValueOf("seafood")
	require.True(t, sv.Contains(ValueOf("foo")))
	require.False(t, sv.Contains(ValueOf("bar")))
	require.False(t, sv.Contains(ValueOf(nil)))

	// string contains stringifies its argument
	require.True(t, ValueOf("seaf00d").Contains(ValueOf(0)))

	// map
	hv := ValueOf(map[string]any{"key": "value"})
	require.True(t, hv.Contains(ValueOf("key")))
	require.False(t, hv.Contains(ValueOf("missing_key")))
	require.False(t, hv.Contains(ValueOf(nil)))

	// MapSlice
	msv := ValueOf(yaml.MapSlice{{Key: "key", Value: "value"}})
	require.True(t, msv.Contains(ValueOf("key")))
	require.False(t, msv.Contains(ValueOf("missing_key")))
	require.False(t, msv.Contains(ValueOf(nil)))
}

func TestNilValue(t *testing.T) {
	nv := ValueOf(nil)

	t.Run("Equal", func(t *testing.T) {
		require.True(t, nv.Equal(ValueOf(nil)))
		require.False(t, nv.Equal(ValueOf(1)))
	})

	t.Run("Less", func(t *testing.T) {
		require.False(t, nv.Less(ValueOf(1)))
	})

	t.Run("IndexValue", func(t *testing.T) {
		require.Nil(t, nv.IndexValue(ValueOf(0)).Interface())
	})

	t.Run("Contains", func(t *testing.T) {
		require.False(t, nv.Contains(ValueOf("x")))
	})

	t.Run("Int panics", func(t *testing.T) {
		require.Panics(t, func() { nv.Int() })
	})

	t.Run("PropertyValue", func(t *testing.T) {
		require.Nil(t, nv.PropertyValue(ValueOf("size")).Interface())
	})

	t.Run("Test", func(t *testing.T) {
		require.False(t, nv.Test())
	})

	t.Run("Interface", func(t *testing.T) {
		require.Nil(t, nv.Interface())
	})
}

func TestRange(t *testing.T) {
	r := NewRange(1, 5)
	require.Equal(t, 5, r.Len())
	require.Equal(t, 1, r.Index(0))
	require.Equal(t, 3, r.Index(2))
	require.Equal(t, 5, r.Index(4))

	r2 := NewRange(0, 0)
	require.Equal(t, 1, r2.Len())
	require.Equal(t, 0, r2.Index(0))
}

func TestMapSliceValue_Interface(t *testing.T) {
	ms := yaml.MapSlice{{Key: "a", Value: 1}, {Key: "b", Value: 2}}
	v := ValueOf(ms)
	result, ok := v.Interface().(yaml.MapSlice)
	require.True(t, ok)
	require.Equal(t, ms, result)
}

func TestValue_PropertyValue_size(t *testing.T) {
	require.Nil(t, ValueOf(nil).PropertyValue(ValueOf("size")).Interface())
	require.Nil(t, ValueOf(false).PropertyValue(ValueOf("size")).Interface())
	require.Nil(t, ValueOf(12).PropertyValue(ValueOf("size")).Interface())

	// string
	require.Equal(t, 7, ValueOf("seafood").PropertyValue(ValueOf("size")).Interface())

	// empty list
	empty := ValueOf([]string{})
	require.Equal(t, 0, empty.PropertyValue(ValueOf("size")).Interface())

	// list
	av := ValueOf([]string{"first", "second", "third"})
	require.Equal(t, 3, av.PropertyValue(ValueOf("size")).Interface())

	// hash
	hv := ValueOf(map[string]any{"key": "value"})
	require.Equal(t, 1, hv.PropertyValue(ValueOf("size")).Interface())

	// hash with "size" key
	withSizeKey := ValueOf(map[string]any{"size": "value"})
	require.Equal(t, "value", withSizeKey.IndexValue(ValueOf("size")).Interface())

	// hash pointer
	hashPtr := ValueOf(&map[string]any{"key": "value"})
	require.Equal(t, 1, hashPtr.PropertyValue(ValueOf("size")).Interface())

	// MapSlice
	msv := ValueOf(yaml.MapSlice{{Key: "key", Value: "value"}})
	require.Equal(t, 1, msv.PropertyValue(ValueOf("size")).Interface())

	// MapSlice with "size" key
	msv = ValueOf(yaml.MapSlice{{Key: "size", Value: "value"}})
	require.Equal(t, "value", msv.PropertyValue(ValueOf("size")).Interface())
}
