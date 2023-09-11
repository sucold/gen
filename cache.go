package gen

import (
	"strings"

	"github.com/gogf/gf/v2/util/gconv"
)

func CacheKey(values ...any) string {
	var builder strings.Builder
	builder.WriteString("k:")
	for _, value := range values {
		builder.WriteString(gconv.String(value))
		builder.WriteString(":")
	}
	return builder.String()
}

type CacheWhere struct {
	Code  string
	Where []Condition
}
