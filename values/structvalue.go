package values

import (
	"reflect"
)

type structValue struct{ wrapperValue }

func (v structValue) IndexValue(index Value) Value {
	return v.PropertyValue(index)
}

func (v structValue) Contains(elem Value) bool {
	name, ok := elem.Interface().(string)
	if !ok {
		return false
	}
	rt := reflect.TypeOf(v.value)
	if rt.Kind() == reflect.Ptr {
		if _, found := rt.MethodByName(name); found {
			return true
		}
		rt = rt.Elem()
	}
	if _, found := rt.MethodByName(name); found {
		return true
	}
	if _, found := v.findField(name); found {
		return true
	}
	return false
}

func (v structValue) PropertyValue(index Value) Value {
	name, ok := index.Interface().(string)
	if !ok {
		return nilValue
	}
	rv := reflect.ValueOf(v.value)
	rt := reflect.TypeOf(v.value)
	if rt.Kind() == reflect.Ptr {
		if _, found := rt.MethodByName(name); found {
			m := rv.MethodByName(name)
			return v.invoke(m)
		}
		rt = rt.Elem()
		rv = rv.Elem()
		if !rv.IsValid() {
			return nilValue
		}
	}
	if _, found := rt.MethodByName(name); found {
		m := rv.MethodByName(name)
		return v.invoke(m)
	}
	if field, found := v.findField(name); found {
		fv := rv.FieldByName(field.Name)
		if fv.Kind() == reflect.Func {
			return v.invoke(fv)
		}
		return ValueOf(fv.Interface())
	}
	return nilValue
}

const tagKey = "liquid"

// like FieldByName, but obeys `liquid:"name"` tags
func (v structValue) findField(name string) (*reflect.StructField, bool) {
	rt := reflect.TypeOf(v.value)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if field, found := rt.FieldByName(name); found {
		if _, ok := field.Tag.Lookup(tagKey); !ok {
			return &field, true
		}
	}
	for i, n := 0, rt.NumField(); i < n; i++ {
		field := rt.Field(i)
		if field.Tag.Get(tagKey) == name {
			return &field, true
		}
	}
	return nil, false
}

func (v structValue) invoke(fv reflect.Value) Value {
	if fv.IsNil() {
		return nilValue
	}
	mt := fv.Type()
	if mt.NumIn() > 0 || mt.NumOut() > 2 {
		return nilValue
	}
	results := fv.Call([]reflect.Value{})
	if len(results) > 1 && !results[1].IsNil() {
		panic(results[1].Interface())
	}
	return ValueOf(results[0].Interface())
}
