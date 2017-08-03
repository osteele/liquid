package values

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testDrop struct{ proxy interface{} }

func (d testDrop) ToLiquid() interface{} { return d.proxy }

func TestToLiquid(t *testing.T) {
	require.Equal(t, 2, ToLiquid(2))
	require.Equal(t, 3, ToLiquid(testDrop{3}))
}

func TestValue_drop(t *testing.T) {
	dv := ValueOf(testDrop{"seafood"})
	require.Equal(t, "seafood", dv.Interface())
	require.Equal(t, true, dv.Contains(ValueOf("foo")))
	require.Equal(t, true, dv.Contains(ValueOf(testDrop{"foo"})))
	require.Equal(t, 7, dv.PropertyValue(ValueOf("size")).Interface())
}

func TestDrop_Resolve_race(t *testing.T) {
	d := ValueOf(testDrop{1})
	values := make(chan int, 2)
	for i := 0; i < 2; i++ {
		go func() { values <- d.Int() }()
	}
	for i := 0; i < 2; i++ {
		require.Equal(t, 1, <-values)
	}
}

func BenchmarkDrop_Resolve_1(b *testing.B) {
	d := ValueOf(testDrop{1})
	for n := 0; n < b.N; n++ {
		_ = d.Int()
	}
}

func BenchmarkDrop_Resolve_2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		d := ValueOf(testDrop{1})
		_ = d.Int()
	}
}

func BenchmarkDrop_Resolve_3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		d := ValueOf(testDrop{1})
		values := make(chan int, 10)
		for i := cap(values); i > 0; i-- {
			values <- d.Int()
		}
		for i := cap(values); i > 0; i-- {
			_ = <-values
		}
	}
}
