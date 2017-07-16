package evaluator

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCall(t *testing.T) {
	fn := func(a, b string) string {
		return a + "," + b + "."
	}
	value, err := Call(reflect.ValueOf(fn), []interface{}{5, 10})
	require.NoError(t, err)
	require.Equal(t, "5,10.", value)

	// extra arguments
	_, err = Call(reflect.ValueOf(fn), []interface{}{5, 10, 20})
	require.Error(t, err)
	require.Contains(t, err.Error(), "wrong number of arguments")
	require.Contains(t, err.Error(), "given 3")
	require.Contains(t, err.Error(), "expected 2")

}

func TestCall_optional(t *testing.T) {
	fn := func(a string, b func(string) string) string {
		return a + "," + b("default") + "."
	}
	value, err := Call(reflect.ValueOf(fn), []interface{}{5})
	require.NoError(t, err)
	require.Equal(t, "5,default.", value)

	value, err = Call(reflect.ValueOf(fn), []interface{}{5, 10})
	require.NoError(t, err)
	require.Equal(t, "5,10.", value)
}
