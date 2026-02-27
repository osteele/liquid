package values

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testDrop struct{ proxy any }

func (d testDrop) ToLiquid() any { return d.proxy }

func TestToLiquid(t *testing.T) {
	require.Equal(t, 2, ToLiquid(2))
	require.Equal(t, 3, ToLiquid(testDrop{3}))
}

func TestValue_drop(t *testing.T) {
	dv := ValueOf(testDrop{"seafood"})
	require.Equal(t, "seafood", dv.Interface())
	require.True(t, dv.Contains(ValueOf("foo")))
	require.True(t, dv.Contains(ValueOf(testDrop{"foo"})))
	require.Equal(t, 7, dv.PropertyValue(ValueOf("size")).Interface())
}

func TestDrop_Equal(t *testing.T) {
	dv := ValueOf(testDrop{42})
	require.True(t, dv.Equal(ValueOf(42)))
	require.False(t, dv.Equal(ValueOf(99)))
}

func TestDrop_Less(t *testing.T) {
	dv := ValueOf(testDrop{10})
	require.True(t, dv.Less(ValueOf(20)))
	require.False(t, dv.Less(ValueOf(5)))
}

func TestDrop_IndexValue(t *testing.T) {
	dv := ValueOf(testDrop{[]string{"a", "b", "c"}})
	require.Equal(t, "a", dv.IndexValue(ValueOf(0)).Interface())
	require.Equal(t, "c", dv.IndexValue(ValueOf(2)).Interface())
}

func TestDrop_Test(t *testing.T) {
	require.True(t, ValueOf(testDrop{1}).Test())
	require.True(t, ValueOf(testDrop{"hello"}).Test())
	require.False(t, ValueOf(testDrop{nil}).Test())
}

func TestDrop_Resolve_race(t *testing.T) {
	d := ValueOf(testDrop{1})
	values := make(chan int, 2)

	for range 2 {
		go func() { values <- d.Int() }()
	}

	for range 2 {
		require.Equal(t, 1, <-values)
	}
}

func BenchmarkDrop_Resolve_1(b *testing.B) {
	d := ValueOf(testDrop{1})

	for range b.N {
		_ = d.Int()
	}
}

func BenchmarkDrop_Resolve_2(b *testing.B) {
	for range b.N {
		d := ValueOf(testDrop{1})
		_ = d.Int()
	}
}

func BenchmarkDrop_Resolve_3(b *testing.B) {
	for range b.N {
		d := ValueOf(testDrop{1})

		values := make(chan int, 10)
		for i := cap(values); i > 0; i-- {
			values <- d.Int()
		}

		for i := cap(values); i > 0; i-- {
			//lint:ignore S1005 TODO look up how else to read the values
			<-values
		}
	}
}
