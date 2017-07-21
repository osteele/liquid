package evaluator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValue(t *testing.T) {
	var (
		nv = ValueOf(nil)
		iv = ValueOf(123)
	)
	require.Equal(t, nil, nv.Interface())
	require.Equal(t, true, ValueOf(true).Interface())
	require.Equal(t, false, ValueOf(false).Interface())
	require.Equal(t, 123, iv.Interface())
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

	require.True(t, ValueOf("seaf00d").Contains(ValueOf(0)))
}

func TestValue_hash(t *testing.T) {
	h := ValueOf(map[string]interface{}{"key": "value"})
	require.True(t, h.Contains(ValueOf("key")))
	require.False(t, h.Contains(ValueOf("missing_key")))
	require.Equal(t, "value", h.IndexValue(ValueOf("key")).Interface())
	require.Equal(t, nil, h.IndexValue(ValueOf("missing_key")).Interface())
	require.Equal(t, 1, h.PropertyValue(ValueOf("size")).Interface())

	withSizeKey := ValueOf(map[string]interface{}{"size": "value"})
	require.Equal(t, "value", withSizeKey.IndexValue(ValueOf("size")).Interface())

	hashPtr := ValueOf(&map[string]interface{}{"key": "value"})
	require.Equal(t, "value", hashPtr.IndexValue(ValueOf("key")).Interface())
	require.Equal(t, nil, hashPtr.IndexValue(ValueOf("missing_key")).Interface())
	require.Equal(t, 1, hashPtr.PropertyValue(ValueOf("size")).Interface())
}

type testValueStruct struct {
	F   int
	F1  func() int
	F2  func() (int, error)
	F2e func() (int, error)
}

func (tv testValueStruct) M1() int           { return 3 }
func (tv testValueStruct) M2() (int, error)  { return 4, nil }
func (tv testValueStruct) M2e() (int, error) { return 4, fmt.Errorf("expected error") }

func (tv *testValueStruct) PM1() int           { return 3 }
func (tv *testValueStruct) PM2() (int, error)  { return 4, nil }
func (tv *testValueStruct) PM2e() (int, error) { return 4, fmt.Errorf("expected error") }

func TestValue_struct(t *testing.T) {
	s := ValueOf(testValueStruct{
		F:   -1,
		F1:  func() int { return 1 },
		F2:  func() (int, error) { return 2, nil },
		F2e: func() (int, error) { return 0, fmt.Errorf("expected error") },
	})
	require.True(t, s.Contains(ValueOf("F")))
	require.True(t, s.Contains(ValueOf("F1")))
	require.Equal(t, -1, s.PropertyValue(ValueOf("F")).Interface())
	require.Equal(t, 1, s.PropertyValue(ValueOf("F1")).Interface())
	require.Equal(t, 2, s.PropertyValue(ValueOf("F2")).Interface())
	require.Panics(t, func() { s.PropertyValue(ValueOf("F2e")) })
	require.Equal(t, 3, s.PropertyValue(ValueOf("M1")).Interface())
	require.Equal(t, 4, s.PropertyValue(ValueOf("M2")).Interface())
	require.Panics(t, func() { s.PropertyValue(ValueOf("M2e")) })
	require.Equal(t, -1, s.IndexValue(ValueOf("F")).Interface())

	p := ValueOf(&testValueStruct{
		F:  -1,
		F1: func() int { return 1 },
	})
	require.True(t, p.Contains(ValueOf("F")))
	require.True(t, p.Contains(ValueOf("F1")))
	require.Equal(t, -1, p.PropertyValue(ValueOf("F")).Interface())
	require.Equal(t, 1, p.PropertyValue(ValueOf("F1")).Interface())
	require.Equal(t, 3, p.PropertyValue(ValueOf("M1")).Interface())
	require.Equal(t, 4, p.PropertyValue(ValueOf("M2")).Interface())
	require.Panics(t, func() { p.PropertyValue(ValueOf("M2e")) })
	require.Equal(t, 3, p.PropertyValue(ValueOf("PM1")).Interface())
	require.Equal(t, 4, p.PropertyValue(ValueOf("PM2")).Interface())
	require.Panics(t, func() { p.PropertyValue(ValueOf("PM2e")) })
}
