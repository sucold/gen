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
	if t.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < t.NumField(); i++ {
		var (
			name   string // 查询字段bind的优先级更高，其次是json
			method string
			ok     bool
			item   = t.Field(i)
			v      any
			v2     any
			cond   Condition
		)
		if name = item.Tag.Get("bind"); name == "" {
			if name = item.Tag.Get("json"); name == "" {
				continue
			}
		}
		if method = item.Tag.Get("cond"); method == "" {
			method = "eq"
		}

		if v, ok = get.GetField(name); !ok {
			continue
		}
		if vl.Field(i).IsZero() {
			continue
		}
		if v2 = vl.Field(i).Interface(); v2 == nil {
			continue
		}
		if cond = field.Any(v, method, v2); cond != nil {
			ret = append(ret, cond)
		}
	}
	return ret
}
