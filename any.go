package gen

import (
	"gorm.io/gen/field"
	"reflect"
)

func Where(get field.GetField, value any) []Condition {
	var (
		t   = reflect.TypeOf(value)
		vl  = reflect.ValueOf(value)
		ret = make([]Condition, 0)
	)
	if vl.Kind() == reflect.Ptr || vl.Kind() == reflect.Interface {
		if vl.IsNil() {
			return nil
		}
		return Where(get, vl.Elem().Interface())
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	fields := getAllFields(vl.Type())
	for _, item := range fields {
		var (
			name   string // 查询字段bind的优先级更高，其次是json
			method string
			ok     bool
			v      any
			cond   Condition
			f      = vl.FieldByIndex(item.Index)
		)
		if item.Type.Kind() == reflect.Struct || item.Type.Kind() == reflect.Ptr || item.Type.Kind() == reflect.Interface {
			if f.Kind() == reflect.Ptr || f.Kind() == reflect.Interface {
				if f.IsNil() {
					continue
				}
				f = f.Elem()
			}
			if ret1 := Where(get, f.Interface()); ret1 != nil {
				ret = append(ret, ret1...)
			}
			continue
		}
		if name = item.Tag.Get("bind"); name == "" {
			if name = item.Tag.Get("json"); name == "" {
				continue
			}
		}
		if method = item.Tag.Get("cond"); method == "" {
			method = "eq"
		}
		if f.IsZero() {
			continue
		}
		if v, ok = get.GetField(name); !ok {
			continue
		}
		if cond = field.Any(v, method, f.Interface()); cond != nil {
			ret = append(ret, cond)
		}
	}

	return ret
}
func getAllFields(t reflect.Type) []reflect.StructField {
	var fields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		fields = append(fields, t.Field(i))
	}
	return fields
}

type Kv struct {
	Key string
	Val string
}
type Edges struct {
	Name   string
	Unique bool
	Fields []*Kv
}
