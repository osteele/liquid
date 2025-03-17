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
			_ = <-values
		}
	}
}
