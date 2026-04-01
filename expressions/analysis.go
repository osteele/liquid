package expressions

import (
	"strings"

	"github.com/osteele/liquid/values"
)

// variableCollector accumulates variable paths discovered during expression tracking.
type variableCollector struct {
	paths [][]string
	seen  map[string]bool
}

func newVariableCollector() *variableCollector {
	return &variableCollector{seen: map[string]bool{}}
}

func (c *variableCollector) record(path []string) {
	if len(path) == 0 {
		return
	}
	key := strings.Join(path, "\x00")
	if !c.seen[key] {
		c.seen[key] = true
		cp := make([]string, len(path))
		copy(cp, path)
		c.paths = append(c.paths, cp)
	}
}

// computeVariables runs an expression evaluator with a tracking context to collect
// all variable paths referenced by the expression. Panics are swallowed so that
// partially-analyzed expressions still return whatever paths were collected.
func computeVariables(evaluator func(Context) values.Value) [][]string {
	tc := &trackingContext{
		collector: newVariableCollector(),
		bindings:  map[string]any{},
	}
	func() {
		defer func() { recover() }() //nolint:errcheck
		result := evaluator(tc)
		if tv, ok := result.(*trackingValue); ok {
			tv.record()
		}
	}()
	if len(tc.collector.paths) == 0 {
		return nil
	}
	return tc.collector.paths
}

// trackingContext is an expressions.Context that records variable accesses.
// It is internal to the expressions package and used only by computeVariables.
type trackingContext struct {
	collector *variableCollector
	bindings  map[string]any
}

func (tc *trackingContext) Get(name string) any {
	return &trackingValue{path: []string{name}, collector: tc.collector}
}

func (tc *trackingContext) Set(name string, value any) {
	tc.bindings[name] = value
}

func (tc *trackingContext) Clone() Context {
	bindings := make(map[string]any, len(tc.bindings))
	for k, v := range tc.bindings {
		bindings[k] = v
	}
	return &trackingContext{collector: tc.collector, bindings: bindings}
}

// ApplyFilter evaluates the receiver and params to trigger path recording,
// then returns nil (filters are not applied during static analysis).
func (tc *trackingContext) ApplyFilter(_ string, receiver valueFn, params []valueFn) (any, error) {
	v := receiver(tc)
	if tv, ok := v.(*trackingValue); ok {
		tv.record()
	}
	for _, p := range params {
		pv := p(tc)
		if tv, ok := pv.(*trackingValue); ok {
			tv.record()
		}
	}
	return nil, nil
}

// untrackable is a sentinel trackingValue returned when a path can no longer be tracked
// (e.g., after a dynamic index access). Its nil collector means record() is a no-op.
var untrackable = &trackingValue{}

// trackingValue is a values.Value that records property access chains.
type trackingValue struct {
	path      []string          // accumulated path segments, e.g. ["customer", "first_name"]
	collector *variableCollector // nil for untrackable sentinel
}

func (tv *trackingValue) record() {
	if tv.collector != nil {
		tv.collector.record(tv.path)
	}
}

func (tv *trackingValue) pathAppend(segment string) *trackingValue {
	p := make([]string, len(tv.path)+1)
	copy(p, tv.path)
	p[len(tv.path)] = segment
	return &trackingValue{path: p, collector: tv.collector}
}

// Interface records the path and returns nil (used by filter args and output).
func (tv *trackingValue) Interface() any {
	tv.record()
	return nil
}

// Int records the path and returns 0 (used by range expressions and arithmetic).
func (tv *trackingValue) Int() int {
	tv.record()
	return 0
}

// Test records the path and returns true (used by if/unless/and/or conditions).
func (tv *trackingValue) Test() bool {
	tv.record()
	return true
}

func (tv *trackingValue) Equal(other values.Value) bool {
	tv.record()
	if otv, ok := other.(*trackingValue); ok {
		otv.record()
	}
	return false
}

func (tv *trackingValue) Less(other values.Value) bool {
	tv.record()
	if otv, ok := other.(*trackingValue); ok {
		otv.record()
	}
	return false
}

func (tv *trackingValue) Contains(other values.Value) bool {
	tv.record()
	if otv, ok := other.(*trackingValue); ok {
		otv.record()
	}
	return false
}

// PropertyValue extends the tracked path for named properties (e.g. x.a → ["x", "a"]).
func (tv *trackingValue) PropertyValue(key values.Value) values.Value {
	if tv.collector == nil {
		return untrackable
	}
	if s, ok := key.Interface().(string); ok {
		return tv.pathAppend(s)
	}
	tv.record()
	return untrackable
}

// IndexValue records the base path and the key (if it's also a variable),
// then returns untrackable since we don't track dynamic index paths.
func (tv *trackingValue) IndexValue(key values.Value) values.Value {
	tv.record()
	if ktv, ok := key.(*trackingValue); ok {
		ktv.record()
	}
	return untrackable
}
