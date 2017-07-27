package values

import (
	yaml "gopkg.in/yaml.v2"
)

type mapSliceValue struct {
	slice yaml.MapSlice
	valueEmbed
}

// func (v mapSliceValue) Equal(o Value) bool     { return v.slice == o.Interface() }
func (v mapSliceValue) Interface() interface{} { return v.slice }

func (v mapSliceValue) Contains(elem Value) bool {
	e := elem.Interface()
	for _, item := range v.slice {
		if e == item.Key {
			return true
		}
	}
	return false
}

func (v mapSliceValue) IndexValue(index Value) Value {
	e := index.Interface()
	for _, item := range v.slice {
		if e == item.Key {
			return ValueOf(item.Value)
		}
	}
	return nilValue
}

func (v mapSliceValue) PropertyValue(index Value) Value {
	result := v.IndexValue(index)
	if result == nilValue && index.Interface() == sizeKey {
		result = ValueOf(len(v.slice))
	}
	return result
}
