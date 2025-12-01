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

// filterParam represents a filter parameter that can be either positional or named
type filterParam struct {
	name  string  // empty string for positional parameters
	value valueFn // the parameter value expression
}

func (c *Config) ensureMapIsCreated() {
	if c.filters == nil {
		c.filters = make(map[string]interface{})
	}
}

// AddFilter adds a filter to the filter dictionary.
func (c *Config) AddFilter(name string, fn any) {
	rf := reflect.ValueOf(fn)
	switch {
	case rf.Kind() != reflect.Func:
		panic("a filter must be a function")
	case rf.Type().NumIn() < 1:
		panic("a filter function must have at least one input")
	case rf.Type().NumOut() < 1 || 2 < rf.Type().NumOut():
		panic("a filter must be have one or two outputs")
		// case rf.Type().Out(1).Implements(…):
		// 	panic(typeError("a filter's second output must be type error"))
	}
	c.ensureMapIsCreated()
	c.filters[name] = fn
}

func (c *Config) AddSafeFilter() {
	// Reading from a nil map is safe; delay allocation until we need to write.
	if c.filters["safe"] == nil {
		c.ensureMapIsCreated()
		c.filters["safe"] = func(in interface{}) interface{} {
			if in, alreadySafe := in.(values.SafeValue); alreadySafe {
				return in
			}
			return values.SafeValue{
				Value: in,
			}
		}
	}
}

var (
	closureType   = reflect.TypeOf(closure{})
	interfaceType = reflect.TypeOf([]any{}).Elem()
)

func isClosureInterfaceType(t reflect.Type) bool {
	return closureType.ConvertibleTo(t) && !interfaceType.ConvertibleTo(t)
}

func (ctx *context) ApplyFilter(name string, receiver valueFn, params []filterParam) (any, error) {
	filter, ok := ctx.filters[name]
	if !ok {
		panic(UndefinedFilter(name))
	}

	fr := reflect.ValueOf(filter)
	args := []any{receiver(ctx).Interface()}

	// Separate positional and named parameters
	var positionalParams []filterParam
	namedParams := make(map[string]any)

	for _, param := range params {
		if param.name == "" {
			positionalParams = append(positionalParams, param)
		} else {
			namedParams[param.name] = param.value(ctx).Interface()
		}
	}

	// Check if filter function accepts named arguments (last param is map[string]any or map[string]interface{})
	acceptsNamedArgs := false
	namedArgsIndex := -1
	if fr.Type().NumIn() > 1 {
		lastParamType := fr.Type().In(fr.Type().NumIn() - 1)
		if lastParamType.Kind() == reflect.Map &&
			lastParamType.Key().Kind() == reflect.String &&
			(lastParamType.Elem().Kind() == reflect.Interface || lastParamType.Elem() == reflect.TypeOf((*any)(nil)).Elem()) {
			acceptsNamedArgs = true
			namedArgsIndex = fr.Type().NumIn() - 1
		}
	}

	// Process positional parameters
	for i, param := range positionalParams {
		// Calculate the actual parameter index (1-based because receiver is first)
		paramIdx := i + 1

		// Skip the named args slot if it exists and we've reached it
		if acceptsNamedArgs && paramIdx >= namedArgsIndex {
			break
		}

		if paramIdx < fr.Type().NumIn() && isClosureInterfaceType(fr.Type().In(paramIdx)) {
			expr, err := Parse(param.value(ctx).Interface().(string))
			if err != nil {
				panic(err)
			}

			args = append(args, closure{expr, ctx})
		} else {
			args = append(args, param.value(ctx).Interface())
		}
	}

	// Add named arguments map if the filter accepts them (always pass, even if empty)
	if acceptsNamedArgs {
		args = append(args, namedParams)
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
