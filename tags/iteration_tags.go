package tags

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/render"
)

// offsetContinueRE matches "offset: continue" (with optional whitespace) in a loop arg string.
var offsetContinueRE = regexp.MustCompile(`\boffset\s*:\s*continue\b`)

// toLoopInt converts any Go numeric type to int for use as a loop limit or
// offset. Returns (n, true) on success, (0, false) if the value is not numeric.
func toLoopInt(v any) (int, bool) {
	if v == nil {
		return 0, false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int(rv.Uint()), true //nolint:gosec // G115: loop bounds are never near MaxUint64
	case reflect.Float32, reflect.Float64:
		return int(rv.Float()), true
	default:
		return 0, false
	}
}

// An IterationKeyedMap is a map that yields its keys, instead of (key, value) pairs, when iterated.
type IterationKeyedMap map[string]any

const forloopVarName = "forloop"

var (
	errLoopContinueLoop = errors.New("continue outside a loop")
	errLoopBreak        = errors.New("break outside a loop")
)

type iterable interface {
	Len() int
	Index(int) any
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
		// The next few lines could panic if the user spoofs us by creating their own loop object.
		// “C++ protects against accident, not against fraud.” – Bjarne Stroustrup
		loopRec := loopVar.(map[string]any)
		cycleMap := loopRec[".cycles"].(map[string]int)
		group, values := cycle.Group, cycle.Values
		n := cycleMap[group]
		cycleMap[group] = n + 1
		// The parser guarantees that there will be at least one item.
		_, err = io.WriteString(w, values[n%len(values)])

		return err
	}, nil
}

func loopTagCompiler(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
	// Detect and strip "offset: continue" before passing to the expression parser,
	// because "continue" would otherwise be interpreted as a variable lookup.
	rawArgs := node.Args
	isOffsetContinue := false

	if offsetContinueRE.MatchString(rawArgs) {
		rawArgs = strings.TrimSpace(offsetContinueRE.ReplaceAllString(rawArgs, ""))
		isOffsetContinue = true
	}

	stmt, err := expressions.ParseStatement(expressions.LoopStatementSelector, rawArgs)
	if err != nil {
		return nil, err
	}

	return func(w io.Writer, ctx render.Context) error {
		// loop modifiers
		val, err := ctx.Evaluate(stmt.Expr)
		if err != nil {
			return err
		}

		iter := makeIterator(val)
		if iter == nil {
			// Collection is nil or non-iterable: render else branch if present.
			// Emit a not-iterable warning when the value is non-nil (nil is
			// intentional; non-nil means the template author made a mistake).
			if val != nil && val != false {
				if hooks := ctx.AuditHooks(); hooks != nil {
					coll := collectionName(node.Args)
					hooks.EmitWarning(node.SourceLoc, node.EndLoc, node.Source, "not-iterable",
						fmt.Sprintf("%q is %T; for loop iterates zero times", coll, val))
				}
			}
			if len(node.Clauses) == 1 && node.Clauses[0].Name == "else" {
				return ctx.RenderBlock(w, node.Clauses[0])
			}
			return nil
		}

		// continueKey is used to remember where a loop over a given collection
		// ended, so a subsequent {% for ... offset:continue %} can resume.
		continueKey := "\x00for_continue_" + loopName(node.Args, stmt.Loop.Variable)

		// effectiveStart tracks the absolute start index into the (possibly
		// reversed) collection; used to advance the continue cursor.
		effectiveStart := 0

		if isOffsetContinue {
			// Resume from the position where the previous loop left off.
			continueOffset, _ := ctx.Get(continueKey).(int)

			// Ruby behavior: apply continue-offset first, then limit, then reversed.
			// The continue offset is always an absolute index into the original
			// (non-reversed) collection.
			if continueOffset >= iter.Len() {
				ctx.Set(continueKey, continueOffset) // cursor stays at end
				return nil                           // collection exhausted
			}

			effectiveStart = continueOffset
			if continueOffset > 0 {
				iter = offsetWrapper{iter, continueOffset}
			}

			// Apply limit if present.
			if stmt.Loop.Limit != nil {
				lval, err := ctx.Evaluate(stmt.Loop.Limit)
				if err != nil {
					return err
				}

				limit, ok := toLoopInt(lval)
				if !ok {
					return ctx.Errorf("loop limit must be an integer")
				}

				if limit >= 0 && limit < iter.Len() {
					iter = limitWrapper{iter, limit}
				}
			}

			// Apply reversed last (Ruby behavior: offset → limit → reversed).
			if stmt.Loop.Reversed {
				iter = reverseWrapper{iter}
			}
		} else {
			// Normal path: Ruby behavior is always offset → limit → reversed,
			// regardless of the order the modifiers appear in the template.
			if stmt.Loop.Offset != nil {
				ov, err := ctx.Evaluate(stmt.Loop.Offset)
				if err != nil {
					return err
				}

				offset, ok := toLoopInt(ov)
				if !ok {
					return ctx.Errorf("loop offset must be an integer")
				}

				if offset > 0 {
					effectiveStart = offset
					iter = offsetWrapper{iter, offset}
				}
			}

			if stmt.Loop.Limit != nil {
				lval, err := ctx.Evaluate(stmt.Loop.Limit)
				if err != nil {
					return err
				}

				limit, ok := toLoopInt(lval)
				if !ok {
					return ctx.Errorf("loop limit must be an integer")
				}

				if limit >= 0 {
					iter = limitWrapper{iter, limit}
				}
			}

			// Apply reversed last (Ruby behavior: offset → limit → reversed).
			if stmt.Loop.Reversed {
				iter = reverseWrapper{iter}
			}
		}

		// Always record the next position so a later offset:continue loop can
		// resume correctly. We store before rendering so that a {% break %} still
		// advances the cursor (matches LiquidJS behaviour).
		ctx.Set(continueKey, effectiveStart+iter.Len())

		if len(node.Clauses) > 1 {
			return errors.New("for loops accept at most one else clause")
		}

		if iter.Len() == 0 && len(node.Clauses) == 1 && node.Clauses[0].Name == "else" {
			return ctx.RenderBlock(w, node.Clauses[0])
		}

		lr := loopRenderer{stmt.Loop, node.Name, loopName(node.Args, stmt.Loop.Variable)}

		// Save audit state for nested-loop correctness, then reset for this loop.
		hooks := ctx.AuditHooks()
		savedIterCount := 0
		savedSuppressInner := false
		if hooks != nil {
			savedIterCount = hooks.IterCount()
			savedSuppressInner = hooks.SuppressInner()
			hooks.ResetIterState()
		}

		renderErr := lr.render(iter, w, ctx)

		// Emit audit event AFTER rendering so TracedCount reflects reality.
		if hooks != nil && hooks.OnIteration != nil {
			iterInfo := buildAuditIterInfo(stmt.Loop, node.Args, iter, isOffsetContinue, effectiveStart, hooks)
			hooks.OnIteration(node.SourceLoc, node.EndLoc, node.Source, iterInfo, hooks.Depth())
		}

		// Restore outer loop's iteration tracking state.
		if hooks != nil {
			hooks.RestoreIterState(savedIterCount, savedSuppressInner)
		}

		return renderErr
	}, nil
}

// loopName returns the "variable-collection" string for forloop.name.
// Args is the raw tag argument string, e.g. "a in array limit:2".
func loopName(args, variable string) string {
	const inKw = " in "
	idx := strings.Index(args, inKw)
	if idx < 0 {
		return variable + "-"
	}
	rest := strings.TrimSpace(args[idx+len(inKw):])
	// Extract the collection token: everything up to the first space (for
	// simple identifiers and range literals like "(1..5)"), which is
	// sufficient for the common Shopify use-case.
	if i := strings.IndexByte(rest, ' '); i > 0 {
		return variable + "-" + rest[:i]
	}
	return variable + "-" + rest
}

// collectionName extracts the collection name from a loop arg string.
// e.g. "item in products limit:5" → "products"
func collectionName(args string) string {
	const inKw = " in "
	idx := strings.Index(args, inKw)
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(args[idx+len(inKw):])
	if i := strings.IndexByte(rest, ' '); i > 0 {
		return rest[:i]
	}
	return rest
}

// buildAuditIterInfo assembles an AuditIterInfo from loop metadata.
// hooks may be nil.
func buildAuditIterInfo(loop expressions.Loop, args string, iter iterable, isOffsetContinue bool, effectiveStart int, hooks *render.AuditHooks) render.AuditIterInfo {
	origLen := iter.Len()

	var limitPtr *int
	var offsetPtr *int

	if loop.Offset != nil || isOffsetContinue {
		off := effectiveStart
		offsetPtr = &off
	}

	effectiveLen := origLen
	truncated := loop.Limit != nil

	if truncated {
		v := effectiveLen
		limitPtr = &v
	}

	// TracedCount = how many iterations had their inner hooks actually called.
	tracedCount := origLen
	if hooks != nil {
		tracedCount = hooks.IterCount()
		if hooks.MaxIterItems > 0 && origLen > hooks.MaxIterItems {
			truncated = true
		}
	}

	return render.AuditIterInfo{
		Variable:    loop.Variable,
		Collection:  collectionName(args),
		Length:      effectiveLen,
		Limit:       limitPtr,
		Offset:      offsetPtr,
		Reversed:    loop.Reversed,
		Truncated:   truncated,
		TracedCount: tracedCount,
	}
}

type loopRenderer struct {
	expressions.Loop

	tagName     string
	forloopName string
}

func (loop loopRenderer) render(iter iterable, w io.Writer, ctx render.Context) error {
	// loop decorator
	decorator, err := makeLoopDecorator(loop, ctx)
	if err != nil {
		return err
	}

	// shallow-bind the loop variables; restore on exit
	parentLoopVal := ctx.Get(forloopVarName)
	defer func(index, forloop any) {
		ctx.Set(forloopVarName, index)
		ctx.Set(loop.Variable, forloop)
	}(parentLoopVal, ctx.Get(loop.Variable))

	cycleMap := map[string]int{}
	// Pre-allocate the forloop map once and reuse it across iterations.
	forloopMap := map[string]any{
		"first":      false,
		"last":       false,
		"index":      0,
		"index0":     0,
		"rindex":     0,
		"rindex0":    0,
		"length":     0,
		"name":       loop.forloopName,
		"parentloop": parentLoopVal,
		".cycles":    cycleMap,
	}

	// For tablerow loops, determine effective columns for forloop metadata.
	var tablerowCols int
	if td, ok := decorator.(tableRowDecorator); ok {
		tablerowCols = int(td)
		forloopMap["col"] = 1
		forloopMap["col0"] = 0
		forloopMap["col_first"] = true
		forloopMap["col_last"] = false
		forloopMap["row"] = 1
	}

	ctx.Set(forloopVarName, forloopMap)

loop:

	for i, l := 0, iter.Len(); i < l; i++ {
		ctx.Set(loop.Variable, iter.Index(i))
		forloopMap["first"] = i == 0
		forloopMap["last"] = i == l-1
		forloopMap["index"] = i + 1
		forloopMap["index0"] = i
		forloopMap["rindex"] = l - i
		forloopMap["rindex0"] = l - i - 1
		forloopMap["length"] = l
		if tablerowCols > 0 {
			col0 := i % tablerowCols
			forloopMap["col"] = col0 + 1
			forloopMap["col0"] = col0
			forloopMap["col_first"] = col0 == 0
			forloopMap["col_last"] = col0+1 == tablerowCols || i+1 == l
			forloopMap["row"] = i/tablerowCols + 1
		}

		// Audit: count this iteration and suppress inner hooks if limit reached.
		if auditHooks := ctx.AuditHooks(); auditHooks != nil {
			if auditHooks.MaxIterItems > 0 && auditHooks.IterCount() >= auditHooks.MaxIterItems {
				auditHooks.SetSuppressInner(true)
			} else {
				// Count only iterations that are actually traced.
				auditHooks.IncrIterCount()
			}
		}

		decorator.before(w, i)
		err := ctx.RenderChildren(w)
		decorator.after(w, i, l)

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
}

func makeLoopDecorator(loop loopRenderer, ctx render.Context) (loopDecorator, error) {
	if loop.tagName == "tablerow" {
		if loop.Cols != nil {
			val, err := ctx.Evaluate(loop.Cols)
			if err != nil {
				return nil, err
			}

			cols, ok := val.(int)
			if !ok {
				return nil, ctx.Errorf("loop cols must be an integer")
			}

			if cols > 0 {
				return tableRowDecorator(cols), nil
			}
		}

		return tableRowDecorator(math.MaxInt32), nil
	}

	return forLoopDecorator{}, nil
}

type loopDecorator interface {
	before(io.Writer, int)
	after(io.Writer, int, int)
}

type forLoopDecorator struct{}

func (d forLoopDecorator) before(io.Writer, int)     {}
func (d forLoopDecorator) after(io.Writer, int, int) {}

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

func (c tableRowDecorator) after(w io.Writer, i, l int) {
	cols := int(c)

	if _, err := io.WriteString(w, `</td>`); err != nil {
		panic(err)
	}

	if (i+1)%cols == 0 || i+1 == l {
		if _, err := io.WriteString(w, `</tr>`); err != nil {
			panic(err)
		}
	}
}

func applyLoopModifiers(loop expressions.Loop, ctx render.Context, iter iterable) (iterable, error) {
	if loop.Reversed {
		iter = reverseWrapper{iter}
	}

	if loop.Offset != nil {
		val, err := ctx.Evaluate(loop.Offset)
		if err != nil {
			return nil, err
		}

		offset, ok := toLoopInt(val)
		if !ok {
			return nil, ctx.Errorf("loop offset must be an integer")
		}

		if offset > 0 {
			iter = offsetWrapper{iter, offset}
		}
	}

	if loop.Limit != nil {
		val, err := ctx.Evaluate(loop.Limit)
		if err != nil {
			return nil, err
		}

		limit, ok := toLoopInt(val)
		if !ok {
			return nil, ctx.Errorf("loop limit must be an integer")
		}

		if limit >= 0 {
			iter = limitWrapper{iter, limit}
		}
	}

	return iter, nil
}

func makeIterator(value any) iterable {
	if iter, ok := value.(iterable); ok {
		return iter
	}

	if value == nil {
		return nil
	}

	switch value := value.(type) {
	case IterationKeyedMap:
		return makeIterationKeyedMap(value)
	case yaml.MapSlice:
		return mapSliceWrapper{value}
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Array, reflect.Slice:
		return sliceWrapper(reflect.ValueOf(value))
	case reflect.Map:
		rv := reflect.ValueOf(value)

		array := make([][]any, rv.Len())
		for i, k := range rv.MapKeys() {
			v := rv.MapIndex(k)
			array[i] = []any{k.Interface(), v.Interface()}
		}

		return sliceWrapper(reflect.ValueOf(array))
	default:
		return nil
	}
}

func makeIterationKeyedMap(m map[string]any) iterable {
	// Iteration chooses a random start, so we need a copy of the keys to iterate through them.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Sorting isn't necessary to match Shopify liquid, but it simplifies debugging.
	sort.Strings(keys)

	return sliceWrapper(reflect.ValueOf(keys))
}

type sliceWrapper reflect.Value

func (w sliceWrapper) Len() int        { return reflect.Value(w).Len() }
func (w sliceWrapper) Index(i int) any { return reflect.Value(w).Index(i).Interface() }

type mapSliceWrapper struct{ ms yaml.MapSlice }

func (w mapSliceWrapper) Len() int { return len(w.ms) }
func (w mapSliceWrapper) Index(i int) any {
	item := w.ms[i]
	return []any{item.Key, item.Value}
}

type limitWrapper struct {
	i iterable
	n int
}

func (w limitWrapper) Len() int        { return min(w.n, w.i.Len()) }
func (w limitWrapper) Index(i int) any { return w.i.Index(i) }

type offsetWrapper struct {
	i iterable
	n int
}

func (w offsetWrapper) Len() int        { return max(0, w.i.Len()-w.n) }
func (w offsetWrapper) Index(i int) any { return w.i.Index(i + w.n) }

type reverseWrapper struct {
	i iterable
}

func (w reverseWrapper) Len() int        { return w.i.Len() }
func (w reverseWrapper) Index(i int) any { return w.i.Index(w.i.Len() - 1 - i) }
