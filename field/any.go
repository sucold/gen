package field

import (
	"reflect"
	"strings"
)

type GetField interface {
	GetField(fieldName string) (any, bool)
}

var (
	funMapCache = make(map[reflect.Type]map[string]string)
)

func getFunMap(field any) map[string]string {
	if data, ok := funMapCache[reflect.TypeOf(field)]; ok {
		return data
	}
	var (
		t    = reflect.TypeOf(field)
		data = make(map[string]string)
	)
	for i := 0; i < t.NumMethod(); i++ {
		data[strings.ToLower(t.Method(i).Name)] = t.Method(i).Name
	}
	funMapCache[t] = data
	return data
}

func Any(field any, method string, value any) Expr {
	var (
		funMap = getFunMap(field)
		ok     bool
		fv     reflect.Value
		fn     reflect.Value
		out    []reflect.Value
		cond   Expr
	)
	if method, ok = funMap[strings.ToLower(method)]; !ok {
		return nil
	}
	if fv = reflect.ValueOf(field); fv.Kind() != reflect.Struct {
		return nil
	}
	fn = fv.MethodByName(method)
	if !fn.IsValid() {
		return nil
	}
	if value == nil {
		return nil
	}
	var args []reflect.Value
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice, reflect.Array:
		vv := reflect.ValueOf(value)
		args = make([]reflect.Value, vv.Len())
		for i := 0; i < vv.Len(); i++ {
			args[i] = vv.Index(i)
		}
	default:
		args = []reflect.Value{reflect.ValueOf(value)}
	}
	if fn.Type().NumIn() != len(args) && !fn.Type().IsVariadic() {
		if fn.Type().NumIn() > len(args) {
			return nil
		}
		args = args[:fn.Type().NumIn()]
	}
	if fn.Type().NumOut() != 1 {
		return nil
	}
	if out = fn.Call(args); len(out) == 0 {
		return nil
	}
	if cond, ok = out[0].Interface().(Expr); !ok {
		return nil
	}
	return cond
}
