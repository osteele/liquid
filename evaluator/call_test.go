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
	value, err = Call(reflect.ValueOf(fn), []interface{}{5, 10, 20})
	require.NoError(t, err)
	require.Equal(t, "5,10.", value)

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
