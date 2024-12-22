package expressions

import (
	gocontext "context"
	"fmt"
	"reflect"
	"strings"

	"github.com/autopilot3/ap3-helpers-go/logger"
	"github.com/autopilot3/liquid/values"
)

// Context is the expression evaluation context. It maps variables names to values.
type Context interface {
	ApplyFilter(string, valueFn, []valueFn) (interface{}, error)
	// Clone returns a copy with a new variable binding map
	// (so that copy.Set does effect the source context.)
	Clone() Context
	Get(string) interface{}
	Set(string, interface{})
}

type context struct {
	Config
	bindings map[string]interface{}
}

// NewContext makes a new expression evaluation context.
func NewContext(vars map[string]interface{}, cfg Config) Context {
	return &context{cfg, vars}
}

func (c *context) Clone() Context {
	bindings := map[string]interface{}{}
	for k, v := range c.bindings {
		bindings[k] = v
	}
	return &context{c.Config, bindings}
}

// Get looks up a variable value in the expression context.
func (c *context) Get(name string) interface{} {
	return values.ToLiquid(c.bindings[name])
}

// Set sets a variable value in the expression context.
func (c *context) Set(name string, value interface{}) {
	c.bindings[name] = value
}

type varsContext struct {
	Config
	variables   map[string]interface{}
	currentVars []string
}

// NewContext makes a new expression evaluation context.
func NewVariablesContext(vars map[string]interface{}, cfg Config) Context {
	return &varsContext{
		Config:    cfg,
		variables: vars,
	}
}

func (c *varsContext) BuildVar(name string) {
	c.currentVars = append(c.currentVars, name)
}

func (c *varsContext) Clone() Context {
	return c
}

const LatestVarNameKey = "$LATEST$"
const LoopVarsKey = "$LOOP_VARS$"

type VariableBind struct {
	Loop       bool
	Attributes map[string]*VariableBind
}

type LoopVar struct {
	Name   string
	Source string
}

type LoopVarsStack struct {
	LastKey int
	Vars    map[int]LoopVar
}

func (l *LoopVarsStack) Set(name string, source string) int {
	key := l.LastKey
	l.Vars[key] = LoopVar{
		Name:   name + ".",
		Source: source,
	}
	l.LastKey++
	return key
}

func (l *LoopVarsStack) Remove(key int) {
	delete(l.Vars, key)
}

func NewLoopVars() *LoopVarsStack {
	return &LoopVarsStack{
		Vars: make(map[int]LoopVar),
	}
}

// Get looks up a variable value in the expression context.
func (c *varsContext) Get(name string) interface{} {
	if len(c.currentVars) > 0 {
		for idx := len(c.currentVars) - 1; idx >= 0; idx-- {
			name += "." + c.currentVars[idx]
		}
		c.currentVars = c.currentVars[:0]
	}
	if loopVars, ok := c.variables[LoopVarsKey]; ok {
		loopVars := loopVars.(*LoopVarsStack)
		for _, loopVar := range loopVars.Vars {
			if strings.HasPrefix(name, loopVar.Name) {
				var bind *VariableBind
				attributeName := strings.TrimPrefix(name, loopVar.Name)
				// only 2 levels of array is allowed
				objArrays := strings.SplitN(loopVar.Source, "[]", 2)
				if len(objArrays) > 1 {
					if val, ok := c.variables[objArrays[0]]; ok {
						bind = val.(*VariableBind)
						attr, ok := bind.Attributes[strings.TrimPrefix(objArrays[1], ".")]
						if ok {
							if attr.Attributes == nil {
								attr.Attributes = make(map[string]*VariableBind)
							}
							attr.Loop = true
							attr.Attributes[attributeName] = &VariableBind{}
						}
					}
				} else {
					if val, ok := c.variables[loopVar.Source]; !ok {
						bind = &VariableBind{
							Loop:       true,
							Attributes: make(map[string]*VariableBind),
						}
						c.variables[loopVar.Source] = bind
					} else if val != nil {
						bind = val.(*VariableBind)
						bind.Loop = true
						if bind.Attributes == nil {
							bind.Attributes = make(map[string]*VariableBind)
						}
					} else {
						bind = &VariableBind{
							Loop:       true,
							Attributes: make(map[string]*VariableBind),
						}
						logger.Errorw(gocontext.Background(), fmt.Sprintf("Error variables is not of VariableBind type: variables: %+v, current varibale %+v, config filter %+v,", c.variables, c.currentVars, c.Config.filters), "Get", "context")
					}
					bind.Attributes[attributeName] = &VariableBind{}
				}
				c.variables[LatestVarNameKey] = loopVar.Source + "[]." + attributeName
				return values.ValueOf(nil)
			}
		}
	}
	if _, ok := c.variables[name]; !ok {
		c.variables[name] = &VariableBind{}
	}
	c.variables[LatestVarNameKey] = name
	return values.ValueOf(nil)
}

// Set sets a variable value in the expression context.
func (c *varsContext) Set(name string, value interface{}) {
}

func (ctx *varsContext) ApplyFilter(name string, receiver valueFn, params []valueFn) (interface{}, error) {
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
