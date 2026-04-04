package filters

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"strconv"
	"strings"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/values"
)

// getPropertyValue retrieves a property from an item using the values package.
// Returns the raw interface value of the property.
func getPropertyValue(item any, property string) any {
	return values.ValueOf(item).PropertyValue(values.ValueOf(property)).Interface()
}

// isFloatType checks if a value is a float type.
func isFloatType(v any) bool {
	switch v.(type) {
	case float32, float64:
		return true
	default:
		return false
	}
}

// whereFilter filters an array, keeping items where item[property] == value.
// If value is nil, keeps items where item[property] is truthy.
func whereFilter(array []any, property string, targetValue func(any) any) []any {
	tv := targetValue(nil)
	result := make([]any, 0)

	for _, item := range array {
		pv := getPropertyValue(item, property)
		if tv == nil {
			if values.ValueOf(pv).Test() {
				result = append(result, item)
			}
		} else {
			if values.Equal(pv, tv) {
				result = append(result, item)
			}
		}
	}

	return result
}

// rejectFilter filters an array, keeping items where item[property] != value.
// If value is nil, keeps items where item[property] is falsy.
func rejectFilter(array []any, property string, targetValue func(any) any) []any {
	tv := targetValue(nil)
	result := make([]any, 0)

	for _, item := range array {
		pv := getPropertyValue(item, property)
		if tv == nil {
			if !values.ValueOf(pv).Test() {
				result = append(result, item)
			}
		} else {
			if !values.Equal(pv, tv) {
				result = append(result, item)
			}
		}
	}

	return result
}

// groupByFilter groups items by the value of a property.
// Returns an array of maps with "name" and "items" keys.
func groupByFilter(array []any, property string) []any {
	type group struct {
		name  any
		items []any
	}

	var groups []group
	index := map[any]int{}

	for _, item := range array {
		pv := getPropertyValue(item, property)
		key := pv

		// Use a string representation for map keys that aren't comparable
		if key != nil && !reflect.TypeOf(key).Comparable() {
			key = fmt.Sprint(key)
		}

		if idx, ok := index[key]; ok {
			groups[idx].items = append(groups[idx].items, item)
		} else {
			index[key] = len(groups)
			groups = append(groups, group{name: pv, items: []any{item}})
		}
	}

	result := make([]any, len(groups))
	for i, g := range groups {
		result[i] = map[string]any{
			"name":  g.name,
			"items": g.items,
		}
	}

	return result
}

// findFilter returns the first item where item[property] == value.
// If value is nil, returns the first item where item[property] is truthy.
func findFilter(array []any, property string, targetValue func(any) any) any {
	tv := targetValue(nil)

	for _, item := range array {
		pv := getPropertyValue(item, property)
		if tv == nil {
			if values.ValueOf(pv).Test() {
				return item
			}
		} else {
			if values.Equal(pv, tv) {
				return item
			}
		}
	}

	return nil
}

// findIndexFilter returns the 0-based index of the first item where
// item[property] == value. Returns nil if not found.
func findIndexFilter(array []any, property string, targetValue func(any) any) any {
	tv := targetValue(nil)

	for i, item := range array {
		pv := getPropertyValue(item, property)
		if tv == nil {
			if values.ValueOf(pv).Test() {
				return i
			}
		} else {
			if values.Equal(pv, tv) {
				return i
			}
		}
	}

	return nil
}

// hasFilter returns true if any item in the array satisfies item[property] == value.
// If value is nil, checks if any item has a truthy property.
func hasFilter(array []any, property string, targetValue func(any) any) bool {
	tv := targetValue(nil)

	for _, item := range array {
		pv := getPropertyValue(item, property)
		if tv == nil {
			if values.ValueOf(pv).Test() {
				return true
			}
		} else {
			if values.Equal(pv, tv) {
				return true
			}
		}
	}

	return false
}

// sumFilter sums numeric values in an array.
// If property is provided, sums the values of that property.
func sumFilter(array []any, property func(string) string) any {
	prop := property("")

	hasFloat := false
	var intSum int64
	var floatSum float64

	for _, item := range array {
		var v any
		if prop != "" {
			v = getPropertyValue(item, prop)
		} else {
			v = item
		}

		if v == nil {
			continue
		}

		if isIntegerType(v) {
			intSum += toInt64(v)
		} else if isFloatType(v) {
			hasFloat = true
			floatSum += toFloat64(v)
		} else if s, ok := v.(string); ok {
			f, err := strconv.ParseFloat(s, 64)
			if err == nil {
				// Check if the string represents an integer
				if f == float64(int64(f)) && !strings.Contains(s, ".") {
					intSum += int64(f)
				} else {
					hasFloat = true
					floatSum += f
				}
			}
			// non-numeric strings are skipped (contribute 0)
		}
		// other types (maps, slices, etc.) are skipped
	}

	if hasFloat {
		return floatSum + float64(intSum)
	}

	return intSum
}

// pushFilter returns a new array with the element appended.
func pushFilter(array []any, element any) []any {
	result := make([]any, len(array)+1)
	copy(result, array)
	result[len(array)] = element

	return result
}

// unshiftFilter returns a new array with the element prepended.
func unshiftFilter(array []any, element any) []any {
	result := make([]any, len(array)+1)
	result[0] = element
	copy(result[1:], array)

	return result
}

// popFilter returns a new array with the last element removed.
func popFilter(array []any) []any {
	if len(array) == 0 {
		return []any{}
	}

	result := make([]any, len(array)-1)
	copy(result, array[:len(array)-1])

	return result
}

// shiftFilter returns a new array with the first element removed.
func shiftFilter(array []any) []any {
	if len(array) == 0 {
		return []any{}
	}

	result := make([]any, len(array)-1)
	copy(result, array[1:])

	return result
}

// sampleFilter returns N random elements from an array.
// If count is 1, returns a single element. Otherwise returns an array.
func sampleFilter(array []any, count func(int) int) any {
	n := count(1)

	if len(array) == 0 {
		if n == 1 {
			return nil
		}

		return []any{}
	}

	if n == 1 {
		return array[rand.IntN(len(array))]
	}

	// Shuffle a copy and take the first n elements
	result := make([]any, len(array))
	copy(result, array)

	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	if n > len(result) {
		n = len(result)
	}

	return result[:n]
}

// toAnySlice converts any slice/array value to []any using reflection.
// Returns nil, false if value is not a slice or array.
func toAnySlice(value any) ([]any, bool) {
	if value == nil {
		return nil, false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}
	items := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		items[i] = rv.Index(i).Interface()
	}
	return items, true
}

// parseExpArgs extracts and validates the varName/exprStr pair from _exp filter args.
func parseExpArgs(filterName string, args []any) (varName, exprStr string, err error) {
	if len(args) < 2 {
		return "", "", fmt.Errorf("%s requires two arguments: variable name and expression", filterName)
	}
	varName, ok1 := args[0].(string)
	exprStr, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return "", "", fmt.Errorf("%s: arguments must be strings", filterName)
	}
	return varName, exprStr, nil
}

// whereExpFilter keeps items where the expression evaluates to truthy.
func whereExpFilter(ctx expressions.Context, value any, args []any) (any, error) {
	varName, exprStr, err := parseExpArgs("where_exp", args)
	if err != nil {
		return nil, err
	}
	items, ok := toAnySlice(value)
	if !ok {
		return []any{}, nil
	}
	expr, err := expressions.Parse(exprStr)
	if err != nil {
		return nil, err
	}
	result := make([]any, 0, len(items))
	for _, el := range items {
		child := ctx.Clone()
		child.Set(varName, el)
		v, err := expr.Evaluate(child)
		if err != nil {
			return nil, err
		}
		if values.ValueOf(v).Test() {
			result = append(result, el)
		}
	}
	return result, nil
}

// rejectExpFilter keeps items where the expression evaluates to falsy.
func rejectExpFilter(ctx expressions.Context, value any, args []any) (any, error) {
	varName, exprStr, err := parseExpArgs("reject_exp", args)
	if err != nil {
		return nil, err
	}
	items, ok := toAnySlice(value)
	if !ok {
		return []any{}, nil
	}
	expr, err := expressions.Parse(exprStr)
	if err != nil {
		return nil, err
	}
	result := make([]any, 0, len(items))
	for _, el := range items {
		child := ctx.Clone()
		child.Set(varName, el)
		v, err := expr.Evaluate(child)
		if err != nil {
			return nil, err
		}
		if !values.ValueOf(v).Test() {
			result = append(result, el)
		}
	}
	return result, nil
}

// groupByExpFilter groups items by the value returned by the expression.
// Returns [{name: val, items: [...]}, ...] preserving insertion order.
func groupByExpFilter(ctx expressions.Context, value any, args []any) (any, error) {
	varName, exprStr, err := parseExpArgs("group_by_exp", args)
	if err != nil {
		return nil, err
	}
	items, ok := toAnySlice(value)
	if !ok {
		return []any{}, nil
	}
	expr, err := expressions.Parse(exprStr)
	if err != nil {
		return nil, err
	}

	type group struct {
		name  any
		items []any
	}
	var groups []group
	index := map[any]int{}

	for _, el := range items {
		child := ctx.Clone()
		child.Set(varName, el)
		v, err := expr.Evaluate(child)
		if err != nil {
			return nil, err
		}
		key := v
		if key != nil && !reflect.TypeOf(key).Comparable() {
			key = fmt.Sprint(key)
		}
		if idx, ok := index[key]; ok {
			groups[idx].items = append(groups[idx].items, el)
		} else {
			index[key] = len(groups)
			groups = append(groups, group{name: v, items: []any{el}})
		}
	}

	result := make([]any, len(groups))
	for i, g := range groups {
		result[i] = map[string]any{"name": g.name, "items": g.items}
	}
	return result, nil
}

// findExpFilter returns the first item where the expression evaluates to truthy, or nil.
func findExpFilter(ctx expressions.Context, value any, args []any) (any, error) {
	varName, exprStr, err := parseExpArgs("find_exp", args)
	if err != nil {
		return nil, err
	}
	items, ok := toAnySlice(value)
	if !ok {
		return nil, nil
	}
	expr, err := expressions.Parse(exprStr)
	if err != nil {
		return nil, err
	}
	for _, el := range items {
		child := ctx.Clone()
		child.Set(varName, el)
		v, err := expr.Evaluate(child)
		if err != nil {
			return nil, err
		}
		if values.ValueOf(v).Test() {
			return el, nil
		}
	}
	return nil, nil
}

// findIndexExpFilter returns the 0-based index of the first item where the expression is truthy.
// Returns nil if no item matches.
func findIndexExpFilter(ctx expressions.Context, value any, args []any) (any, error) {
	varName, exprStr, err := parseExpArgs("find_index_exp", args)
	if err != nil {
		return nil, err
	}
	items, ok := toAnySlice(value)
	if !ok {
		return nil, nil
	}
	expr, err := expressions.Parse(exprStr)
	if err != nil {
		return nil, err
	}
	for i, el := range items {
		child := ctx.Clone()
		child.Set(varName, el)
		v, err := expr.Evaluate(child)
		if err != nil {
			return nil, err
		}
		if values.ValueOf(v).Test() {
			return i, nil
		}
	}
	return nil, nil
}

// hasExpFilter returns true if any item in the array satisfies the expression.
func hasExpFilter(ctx expressions.Context, value any, args []any) (any, error) {
	varName, exprStr, err := parseExpArgs("has_exp", args)
	if err != nil {
		return nil, err
	}
	items, ok := toAnySlice(value)
	if !ok {
		return false, nil
	}
	expr, err := expressions.Parse(exprStr)
	if err != nil {
		return nil, err
	}
	for _, el := range items {
		child := ctx.Clone()
		child.Set(varName, el)
		v, err := expr.Evaluate(child)
		if err != nil {
			return nil, err
		}
		if values.ValueOf(v).Test() {
			return true, nil
		}
	}
	return false, nil
}
