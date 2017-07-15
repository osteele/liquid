package tags

import (
	"fmt"
	"io"
	"reflect"
	"regexp"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/render"
)

const forloopVarName = "forloop"

var errLoopContinueLoop = fmt.Errorf("continue outside a loop")
var errLoopBreak = fmt.Errorf("break outside a loop")

type iterable interface {
	Len() int
	Index(int) interface{}
}

func breakTag(string) (func(io.Writer, render.Context) error, error) {
	return func(_ io.Writer, ctx render.Context) error {
		return ctx.WrapError(errLoopBreak)
	}, nil
}

func continueTag(string) (func(io.Writer, render.Context) error, error) {
	return func(_ io.Writer, ctx render.Context) error {
		return ctx.WrapError(errLoopContinueLoop)
	}, nil
}

func cycleTag(args string) (func(io.Writer, render.Context) error, error) {
	stmt, err := expressions.ParseStatement(expressions.CycleStatementSelector, args)
	if err != nil {
		return nil, err
	}
	cycle := stmt.Cycle
	return func(w io.Writer, ctx render.Context) error {
		loopVar := ctx.Get(forloopVarName)
		if loopVar == nil {
			return ctx.Errorf("cycle must be within a forloop")
		}
		// the next few lines could panic if the user spoofs us by creating their own loop object
		// “C++ protects against accident, not against fraud.” – Bjarne Stroustrup
		loopRec := loopVar.(map[string]interface{})
		cycleMap := loopRec[".cycles"].(map[string]int)
		group, values := cycle.Group, cycle.Values
		n := cycleMap[group]
		cycleMap[group] = n + 1
		// The parser guarantees that there will be at least one item.
		_, err = w.Write([]byte(values[n%len(values)]))
		return err
	}, nil
}

// TODO is the Liquid syntax compatible with a context-free lexer instead?
var loopRepairMatcher = regexp.MustCompile(`^(.+\s+in\s+\(.+)\.\.(.+\).*)$`)

func loopTagCompiler(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
	src := node.Args
	if m := loopRepairMatcher.FindStringSubmatch(src); m != nil {
		src = m[1] + " .. " + m[2]
	}
	stmt, err := expressions.ParseStatement(expressions.LoopStatementSelector, src)
	if err != nil {
		return nil, err
	}
	loop := stmt.Loop
	dec := makeLoopDecorator(node.Name, loop)
	return func(w io.Writer, ctx render.Context) error {
		val, err := ctx.Evaluate(loop.Expr)
		if err != nil {
			return err
		}
		iter := makeIterator(val)
		if iter == nil {
			return nil
		}
		iter = applyLoopModifiers(loop, iter)
		// shallow-bind the loop variables; restore on exit
		defer func(index, forloop interface{}) {
			ctx.Set(forloopVarName, index)
			ctx.Set(loop.Variable, forloop)
		}(ctx.Get(forloopVarName), ctx.Get(loop.Variable))
		cycleMap := map[string]int{}
	loop:
		for i, len := 0, iter.Len(); i < len; i++ {
			ctx.Set(loop.Variable, iter.Index(i))
			ctx.Set(forloopVarName, map[string]interface{}{
				"first":   i == 0,
				"last":    i == len-1,
				"index":   i + 1,
				"index0":  i,
				"rindex":  len - i,
				"rindex0": len - i - 1,
				"length":  len,
				".cycles": cycleMap,
			})
			dec.before(w, i)
			err := ctx.RenderChildren(w)
			dec.after(w, i, len)
			switch {
			case err == nil:
			// fall through
			case err.Cause() == errLoopBreak:
				break loop
			case err.Cause() == errLoopContinueLoop:
				continue loop
			default:
				return err
			}
		}
		return nil
	}, nil
}

func makeLoopDecorator(tagName string, loop expressions.Loop) loopDecorator {
	if tagName == "tablerow" {
		return tableRowDecorator(loop.Cols)
	}
	return nullLoopDecorator{}
}

type loopDecorator interface {
	before(io.Writer, int)
	after(io.Writer, int, int)
}

type nullLoopDecorator struct{}

func (d nullLoopDecorator) before(io.Writer, int)     {}
func (d nullLoopDecorator) after(io.Writer, int, int) {}

type tableRowDecorator int

func (c tableRowDecorator) before(w io.Writer, i int) {
	cols := int(c)
	row, col := i/cols, i%cols
	if col == 0 {
		if _, err := fmt.Fprintf(w, `<tr class="row%d">`, row+1); err != nil {
			panic(err)
		}
	}
	if _, err := fmt.Fprintf(w, `<td class="col%d">`, col+1); err != nil {
		panic(err)
	}
}

func (c tableRowDecorator) after(w io.Writer, i, len int) {
	cols := int(c)
	if _, err := io.WriteString(w, `</td>`); err != nil {
		panic(err)
	}
	if (i+1)%cols == 0 || i+1 == len {
		if _, err := io.WriteString(w, `</tr>`); err != nil {
			panic(err)
		}
	}
}

func applyLoopModifiers(loop expressions.Loop, iter iterable) iterable {
	if loop.Reversed {
		iter = reverseWrapper{iter}
	}
	if loop.Offset > 0 {
		iter = offsetWrapper{iter, loop.Offset}
	}
	if loop.Limit != nil {
		iter = limitWrapper{iter, *loop.Limit}
	}
	return iter
}
func makeIterator(value interface{}) iterable {
	if iter, ok := value.(iterable); ok {
		return iter
	}
	if value == nil {
		return nil
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Array, reflect.Slice:
		return sliceWrapper(reflect.ValueOf(value))
	case reflect.Map:
		rt := reflect.ValueOf(value)
		array := make([]interface{}, 0, rt.Len())
		for _, k := range rt.MapKeys() {
			array = append(array, k.Interface())
		}
		return sliceWrapper(reflect.ValueOf(array))
	default:
		return nil
	}
}

type sliceWrapper reflect.Value

func (w sliceWrapper) Len() int                { return reflect.Value(w).Len() }
func (w sliceWrapper) Index(i int) interface{} { return reflect.Value(w).Index(i).Interface() }

type limitWrapper struct {
	i iterable
	n int
}

func (w limitWrapper) Len() int                { return intMin(w.n, w.i.Len()) }
func (w limitWrapper) Index(i int) interface{} { return w.i.Index(i) }

type offsetWrapper struct {
	i iterable
	n int
}

func (w offsetWrapper) Len() int                { return intMax(0, w.i.Len()-w.n) }
func (w offsetWrapper) Index(i int) interface{} { return w.i.Index(i + w.n) }

type reverseWrapper struct {
	i iterable
}

func (w reverseWrapper) Len() int                { return w.i.Len() }
func (w reverseWrapper) Index(i int) interface{} { return w.i.Index(w.i.Len() - 1 - i) }

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
