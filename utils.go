package soda

import (
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

func ptr[T any](v T) *T {
	return &v
}

func unptr[T any](v *T) T {
	if v == nil {
		return reflect.Zero(reflect.TypeOf(v)).Interface().(T)
	}
	return *v
}

func toSlice(val, typ string) []any {
	ss := strings.Split(val, SeparatorPropItem)
	result := make([]any, 0, len(ss))
	var transform func(string) (any, error)
	switch typ {
	case typeString:
		transform = func(s string) (any, error) { return s, nil }
	case typeInteger:
		transform = func(s string) (any, error) { return toIntE(s) }
	case typeNumber:
		transform = func(s string) (any, error) { return toFloatE(s) }
	default:
		return nil
	}
	for _, s := range ss {
		if v, e := transform(s); e == nil {
			result = append(result, v)
		}
	}
	return result
}

func toBool(v string) bool {
	if v == "" {
		return true
	}
	b, _ := strconv.ParseBool(v)
	return b
}

func toIntE(v string) (int64, error) {
	return strconv.ParseInt(v, 10, 64)
}

func toInt(v string) int64 {
	i, _ := toIntE(v)
	return i
}

func toFloatE(v string) (float64, error) {
	return strconv.ParseFloat(v, 64)
}

func toFloat(v string) float64 {
	f, _ := toFloatE(v)
	return f
}

func genDefaultOperationID(method, path string) string {
	return regexOperationID.ReplaceAllString(method+" "+path, "-")
}

func appendUniqBy[T any](fn func(a, b T) bool, slice []T, elems ...T) {
	for _, elem := range elems {
		found := false
		for _, s := range slice {
			if fn(s, elem) {
				found = true
				break
			}
		}

		if !found {
			slice = append(slice, elem)
		}
	}
}

func sameSecurityRequirements(a, b *base.SecurityRequirement) bool {
	return reflect.DeepEqual(a.Requirements, b.Requirements)
}

func sameTag(a, b *base.Tag) bool {
	return a.Name == b.Name
}

func sameVal[T comparable](a, b T) bool {
	return a == b
}

func cleanPath(pattern string) string {
	// remove the chi parameter inner regex constaint strings
	re := regexp.MustCompile(`\{(.*?):.*?\}`)
	return re.ReplaceAllString(pattern, "{$1}")
}

func GetInput[T any](c *http.Request) *T {
	return c.Context().Value(KeyInput).(*T)
}
