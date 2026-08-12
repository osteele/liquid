package values

import (
	"reflect"
)

type structValue struct{ wrapperValue }

func (sv structValue) IndexValue(index Value) Value {
	return sv.PropertyValue(index)
}

func (sv structValue) Contains(elem Value) bool {
	name, ok := elem.Interface().(string)
	if !ok {
		return false
	}

	st := reflect.TypeOf(sv.value)
	if st.Kind() == reflect.Ptr {
		if _, found := st.MethodByName(name); found {
			return true
		}

		st = st.Elem()
	}

	if _, found := st.MethodByName(name); found {
		return true
	}

	if _, found := sv.findField(name); found {
		return true
	}

	return false
}

func (sv structValue) PropertyValue(index Value) Value {
	name, ok := index.Interface().(string)
	if !ok {
		return undefinedValue
	}

	sr := reflect.ValueOf(sv.value)

	st := reflect.TypeOf(sv.value)
	if st.Kind() == reflect.Ptr {
		if _, found := st.MethodByName(name); found {
			m := sr.MethodByName(name)
			return sv.invoke(m)
		}

		st = st.Elem()

		sr = sr.Elem()
		if !sr.IsValid() {
			return undefinedValue
		}
	}

	if _, ok := st.MethodByName(name); ok {
		m := sr.MethodByName(name)
		return sv.invoke(m)
	}

	if field, ok := sv.findField(name); ok {
		fv := sr.FieldByName(field.Name)
		if !fv.CanInterface() {
			return undefinedValue
		}
		if fv.Kind() == reflect.Func {
			return sv.invoke(fv)
		}

		return ValueOf(fv.Interface())
	}

	return undefinedValue
}

const tagKey = "liquid"

// like FieldByName, but obeys `liquid:"name"` tags
func (sv structValue) findField(name string) (*reflect.StructField, bool) {
	sr := reflect.TypeOf(sv.value)
	if sr.Kind() == reflect.Ptr {
		sr = sr.Elem()
	}

	if field, ok := sr.FieldByName(name); ok {
		if field.PkgPath == "" {
			if _, ok := field.Tag.Lookup(tagKey); !ok {
				return &field, true
			}
		}
	}

	for i, n := 0, sr.NumField(); i < n; i++ {
		field := sr.Field(i)
		if field.PkgPath == "" && field.Tag.Get(tagKey) == name {
			return &field, true
		}
	}

	return nil, false
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
