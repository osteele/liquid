package evaluator

import (
	"fmt"
	"reflect"
	"strings"
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

	// extra arguments (variadic)
	fnVaridic := func(a string, b ...string) string {
		return a + "," + strings.Join(b, ",") + "."
	}
	value, err = Call(reflect.ValueOf(fnVaridic), []interface{}{5, 10})
	require.NoError(t, err)
	require.Equal(t, "5,10.", value)
	value, err = Call(reflect.ValueOf(fnVaridic), []interface{}{5, 10, 15, 20})
	require.NoError(t, err)
	require.Equal(t, "5,10,15,20.", value)

	// extra arguments (non variadic)
	_, err = Call(reflect.ValueOf(fn), []interface{}{5, 10, 20})
	require.Error(t, err)
	require.Contains(t, err.Error(), "wrong number of arguments")
	require.Contains(t, err.Error(), "given 3")
	require.Contains(t, err.Error(), "expected 2")

	// error return
	fn2 := func(int) (int, error) { return 0, fmt.Errorf("expected error") }
	_, err = Call(reflect.ValueOf(fn2), []interface{}{2})
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected error")
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
