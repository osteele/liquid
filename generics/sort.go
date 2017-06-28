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
	return Less(s[i], s[j])
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
	index := func(i int) interface{} {
		value := s.data[i]
		rt := reflect.ValueOf(value)
		if rt.Kind() == reflect.Map && rt.Type().Key().Kind() == reflect.String {
			elem := rt.MapIndex(reflect.ValueOf(s.key))
			if elem.IsValid() {
				return elem.Interface()
			}
		}
		return nil
	}
	a, b := index(i), index(j)
	nilFirst := true
	switch {
	case a == nil && b == nil:
		return false
	case a == nil:
		return nilFirst
	case b == nil:
		return !nilFirst
	}
	return Less(a, b)
}
