package values

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:recvcheck
type testValueStruct struct {
	F       int
	Nest    *testValueStruct
	Renamed int `liquid:"name"`
	Omitted int `liquid:"-"`
	F1      func() int
	F2      func() (int, error)
	F2e     func() (int, error)
}

func (tv testValueStruct) M1() int           { return 3 }
func (tv testValueStruct) M2() (int, error)  { return 4, nil }
func (tv testValueStruct) M2e() (int, error) { return 4, errors.New("expected error") }

func (tv *testValueStruct) PM1() int           { return 3 }
func (tv *testValueStruct) PM2() (int, error)  { return 4, nil }
func (tv *testValueStruct) PM2e() (int, error) { return 4, errors.New("expected error") }

func TestValue_struct(t *testing.T) {
	s := ValueOf(testValueStruct{
		F:       -1,
		Nest:    &testValueStruct{F: -2},
		Renamed: 100,
		Omitted: 200,
		F1:      func() int { return 1 },
		F2:      func() (int, error) { return 2, nil },
		F2e:     func() (int, error) { return 0, errors.New("expected error") },
	})

	// fields
	require.True(t, s.Contains(ValueOf("F")))
	require.True(t, s.Contains(ValueOf("F1")))
	require.Equal(t, -1, s.PropertyValue(ValueOf("F")).Interface())

	// Nesting
	require.Equal(t, -2, s.PropertyValue(ValueOf("Nest")).PropertyValue(ValueOf("F")).Interface())
	require.Nil(t, s.PropertyValue(ValueOf("Nest")).PropertyValue(ValueOf("Nest")).PropertyValue(ValueOf("F")).Interface())

	// field tags
	require.False(t, s.Contains(ValueOf("Renamed")))
	require.False(t, s.Contains(ValueOf("Omitted")))
	require.True(t, s.Contains(ValueOf("name")))
	require.Nil(t, s.PropertyValue(ValueOf("Renamed")).Interface())
	require.Nil(t, s.PropertyValue(ValueOf("Omitted")).Interface())
	require.Equal(t, 100, s.PropertyValue(ValueOf("name")).Interface())

	// func fields
	require.Equal(t, 1, s.PropertyValue(ValueOf("F1")).Interface())
	require.Equal(t, 2, s.PropertyValue(ValueOf("F2")).Interface())
	require.Panics(t, func() { s.PropertyValue(ValueOf("F2e")) })

	// methods
	require.Equal(t, 3, s.PropertyValue(ValueOf("M1")).Interface())
	require.Equal(t, 4, s.PropertyValue(ValueOf("M2")).Interface())
	require.Panics(t, func() { s.PropertyValue(ValueOf("M2e")) })
	require.Equal(t, -1, s.IndexValue(ValueOf("F")).Interface())
}

func TestValue_struct_ptr(t *testing.T) {
	p := ValueOf(&testValueStruct{
		F:  -1,
		F1: func() int { return 1 },
	})

	// fields
	require.True(t, p.Contains(ValueOf("F")))
	require.True(t, p.Contains(ValueOf("F1")))
	require.Equal(t, -1, p.PropertyValue(ValueOf("F")).Interface())

	// func fields
	require.Equal(t, 1, p.PropertyValue(ValueOf("F1")).Interface())

	// members
	require.Equal(t, 3, p.PropertyValue(ValueOf("M1")).Interface())
	require.Equal(t, 4, p.PropertyValue(ValueOf("M2")).Interface())
	require.Panics(t, func() { p.PropertyValue(ValueOf("M2e")) })

	// pointer members
	require.Equal(t, 3, p.PropertyValue(ValueOf("PM1")).Interface())
	require.Equal(t, 4, p.PropertyValue(ValueOf("PM2")).Interface())
	require.Panics(t, func() { p.PropertyValue(ValueOf("PM2e")) })
}

// ---------------------------------------------------------------------------
// dropMethodMissing (catch-all property access)
// ---------------------------------------------------------------------------

type methodMissingDrop struct {
	Known int
	data  map[string]any
}

func (d methodMissingDrop) MissingMethod(key string) any {
	return d.data[key]
}

func TestStructValue_MissingMethod_known(t *testing.T) {
	// Defined fields take priority; MissingMethod is NOT called for them.
	v := ValueOf(methodMissingDrop{Known: 99, data: map[string]any{"Known": "shadow"}})
	require.Equal(t, 99, v.PropertyValue(ValueOf("Known")).Interface())
}

func TestStructValue_MissingMethod_dynamic(t *testing.T) {
	// Undefined keys fall through to MissingMethod.
	v := ValueOf(methodMissingDrop{data: map[string]any{"foo": "bar", "num": 42}})
	require.Equal(t, "bar", v.PropertyValue(ValueOf("foo")).Interface())
	require.Equal(t, 42, v.PropertyValue(ValueOf("num")).Interface())
}

func TestStructValue_MissingMethod_nil(t *testing.T) {
	// MissingMethod returning nil produces a nil Value (not panic).
	v := ValueOf(methodMissingDrop{data: map[string]any{}})
	require.Nil(t, v.PropertyValue(ValueOf("missing")).Interface())
}

func TestStructValue_MissingMethod_noInterface(t *testing.T) {
	// Types without MissingMethod still return nil for unknown properties.
	v := ValueOf(testValueStruct{F: 1})
	require.Nil(t, v.PropertyValue(ValueOf("nonexistent")).Interface())
}
