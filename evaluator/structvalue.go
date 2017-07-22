package evaluator

import "reflect"

func (v structValue) IndexValue(index Value) Value {
	return v.PropertyValue(index)
}

func (v structValue) Contains(elem Value) bool {
	name, ok := elem.Interface().(string)
	if !ok {
		return false
	}
	rt := reflect.TypeOf(v.basis)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if _, found := rt.FieldByName(name); found {
		return true
	}
	if _, found := rt.MethodByName(name); found {
		return true
	}
	return false
}

func (v structValue) PropertyValue(index Value) Value {
	name, ok := index.Interface().(string)
	if !ok {
		return nilValue
	}
	rv := reflect.ValueOf(v.basis)
	rt := reflect.TypeOf(v.basis)
	if _, found := rt.MethodByName(name); found {
		m := rv.MethodByName(name)
		return v.invoke(m)
	}
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}
	if _, found := rt.FieldByName(name); found {
		fv := rv.FieldByName(name)
		if fv.Kind() == reflect.Func {
			return v.invoke(fv)
		}
		return ValueOf(fv.Interface())
	}
	if _, found := rt.MethodByName(name); found {
		m := rv.MethodByName(name)
		return v.invoke(m)
	}
	return nilValue
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
