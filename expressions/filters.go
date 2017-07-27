package expressions

import (
	"fmt"
	"reflect"

	"github.com/osteele/liquid/values"
)

// An InterpreterError is an error during expression interpretation.
// It is used for errors in the input expression, to distinguish them
// from implementation errors in the interpreter.
type InterpreterError string

func (e InterpreterError) Error() string { return string(e) }

// UndefinedFilter is an error that the named filter is not defined.
type UndefinedFilter string

func (e UndefinedFilter) Error() string {
	return fmt.Sprintf("undefined filter %q", string(e))
}

// FilterError is the error returned by a filter when it is applied
type FilterError struct {
	FilterName string
	Err        error
}

func (e FilterError) Error() string {
	return fmt.Sprintf("error applying filter %q (%q)", e.FilterName, e.Err)
}

type valueFn func(Context) values.Value

// AddFilter adds a filter to the filter dictionary.
func (c *Config) AddFilter(name string, fn interface{}) {
	rf := reflect.ValueOf(fn)
	switch {
	case rf.Kind() != reflect.Func:
		panic(fmt.Errorf("a filter must be a function"))
	case rf.Type().NumIn() < 1:
		panic(fmt.Errorf("a filter function must have at least one input"))
	case rf.Type().NumOut() < 1 || 2 < rf.Type().NumOut():
		panic(fmt.Errorf("a filter must be have one or two outputs"))
		// case rf.Type().Out(1).Implements(…):
		// 	panic(typeError("a filter's second output must be type error"))
	}
	if len(c.filters) == 0 {
		c.filters = make(map[string]interface{})
	}
	c.filters[name] = fn
}

var closureType = reflect.TypeOf(closure{})
var interfaceType = reflect.TypeOf([]interface{}{}).Elem()

func isClosureInterfaceType(t reflect.Type) bool {
	return closureType.ConvertibleTo(t) && !interfaceType.ConvertibleTo(t)
}

func (ctx *context) ApplyFilter(name string, receiver valueFn, params []valueFn) (interface{}, error) {
	filter, ok := ctx.filters[name]
	if !ok {
		panic(UndefinedFilter(name))
	}
	fr := reflect.ValueOf(filter)
	args := []interface{}{receiver(ctx).Interface()}
	for i, param := range params {
		if i+1 < fr.Type().NumIn() && isClosureInterfaceType(fr.Type().In(i+1)) {
			expr, err := Parse(param(ctx).Interface().(string))
			if err != nil {
				panic(err)
			}
			args = append(args, closure{expr, ctx})
		} else {
			args = append(args, param(ctx).Interface())
		}
	}
	out, err := values.Call(fr, args)
	if err != nil {
		if e, ok := err.(*values.CallParityError); ok {
			err = &values.CallParityError{NumArgs: e.NumArgs - 1, NumParams: e.NumParams - 1}
		}
		return nil, err
	}
	switch out := out.(type) {
	case []byte:
		return string(out), nil
	default:
		return out, nil
	}
}
