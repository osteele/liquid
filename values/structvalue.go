package values

import (
	"reflect"
	"sync"
)

type structValue struct{ wrapperValue }

type structPropertyKind uint8

const (
	missingStructProperty structPropertyKind = iota
	structMethod
	structField
)

type structPropertyKey struct {
	typ  reflect.Type
	name string
}

type structProperty struct {
	kind  structPropertyKind
	index []int
}

var structPropertyCache sync.Map

func (sv structValue) IndexValue(index Value) Value {
	return sv.PropertyValue(index)
}

func (sv structValue) Contains(elem Value) bool {
	name, ok := elem.Interface().(string)
	if !ok {
		return false
	}

	property := lookupStructProperty(reflect.TypeOf(sv.value), name)
	return property.kind != missingStructProperty
}

func (sv structValue) PropertyValue(index Value) Value {
	name, ok := index.Interface().(string)
	if !ok {
		return undefinedValue
	}

	st := reflect.TypeOf(sv.value)
	property := lookupStructProperty(st, name)
	sr := reflect.ValueOf(sv.value)

	switch property.kind {
	case structMethod:
		return sv.invoke(sr.Method(property.index[0]))
	case structField:
		if st.Kind() == reflect.Ptr {
			sr = sr.Elem()
			if !sr.IsValid() {
				return undefinedValue
			}
		}

		fv := sr.FieldByIndex(property.index)
		if !fv.CanInterface() {
			return undefinedValue
		}
		if fv.Kind() == reflect.Func {
			return sv.invoke(fv)
		}

		return ValueOf(fv.Interface())
	default:
		return undefinedValue
	}
}

const tagKey = "liquid"

func lookupStructProperty(typ reflect.Type, name string) structProperty {
	key := structPropertyKey{typ: typ, name: name}
	if cached, ok := structPropertyCache.Load(key); ok {
		return cached.(structProperty)
	}

	property := inspectStructProperty(typ, name)
	if property.kind == missingStructProperty {
		return property
	}
	actual, _ := structPropertyCache.LoadOrStore(key, property)
	return actual.(structProperty)
}

func inspectStructProperty(typ reflect.Type, name string) structProperty {
	if method, ok := typ.MethodByName(name); ok {
		return structProperty{kind: structMethod, index: []int{method.Index}}
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if field, ok := typ.FieldByName(name); ok {
		if field.PkgPath == "" {
			if _, ok := field.Tag.Lookup(tagKey); !ok {
				return structProperty{kind: structField, index: field.Index}
			}
		}
	}

	for i := range typ.NumField() {
		field := typ.Field(i)
		if field.PkgPath == "" && field.Tag.Get(tagKey) == name {
			return structProperty{kind: structField, index: field.Index}
		}
	}

	return structProperty{}
}

func (sv structValue) invoke(fv reflect.Value) Value {
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
