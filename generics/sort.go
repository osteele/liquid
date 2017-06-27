package generics

import (
	"reflect"
	"sort"
)

type genericSortable []interface{}

// Sort any []interface{} value.
func Sort(data []interface{}) {
	sort.Sort(genericSortable(data))
}

// Len is part of sort.Interface.
func (s genericSortable) Len() int {
	return len(s)
}

// Swap is part of sort.Interface.
func (s genericSortable) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is part of sort.Interface.
func (s genericSortable) Less(i, j int) bool {
	return genericSameTypeCompare(s[i], s[j]) < 0
}

// SortByProperty sorts maps on their key indices.
func SortByProperty(data []interface{}, key string) {
	sort.Sort(sortableByProperty{data, key})
}

type sortableByProperty struct {
	data []interface{}
	key  string
}

// Len is part of sort.Interface.
func (s sortableByProperty) Len() int {
	return len(s.data)
}

// Swap is part of sort.Interface.
func (s sortableByProperty) Swap(i, j int) {
	data := s.data
	data[i], data[j] = data[j], data[i]
}

// Less is part of sort.Interface.
func (s sortableByProperty) Less(i, j int) bool {
	// index returns the value at the s.key, if in is a map that contains this key
	index := func(in interface{}) interface{} {
		rt := reflect.ValueOf(in)
		if rt.Kind() == reflect.Map && rt.Type().Key().Kind() == reflect.String {
			return rt.MapIndex(reflect.ValueOf(s.key)).Interface()
		}
		return nil
	}
	a, b := index(s.data[i]), index(s.data[j])
	// TODO implement nil-first vs. nil last
	switch {
	case a == nil:
		return true
	case b == nil:
		return false
	default:
		// TODO relax same type requirement
		return genericSameTypeCompare(a, b) < 0
	}
}
