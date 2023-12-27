package soda

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"gopkg.in/yaml.v3"
)

// ptr creates a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// unptr gets the value from the pointer. If the pointer is nil, it returns the zero value of that type.
func unptr[T any](v *T) T {
	if v == nil {
		return reflect.Zero(reflect.TypeOf(v).Elem()).Interface().(T)
	}
	return *v
}

// toSlice converts a string to a slice, the type of conversion is determined by the typ parameter.
func toSlice(val, typ string) []*yaml.Node {
	ss := strings.Split(val, SeparatorPropItem)
	result := make([]*yaml.Node, 0, len(ss))
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
			result = append(result, toNode(v))
		}
	}
	return result
}

func toNode(v any) (n *yaml.Node) {
	b, _ := yaml.Marshal(v)
	_ = yaml.Unmarshal(b, &n)
	return
}

// toBool converts a string to a boolean value. If the string is empty, it returns true.
func toBool(v string) bool {
	if v == "" {
		return true
	}
	b, _ := strconv.ParseBool(v)
	return b
}

// toIntE converts a string to int64 type, if the conversion fails, it returns an error.
func toIntE(v string) (int64, error) {
	return strconv.ParseInt(v, 10, 64)
}

// toInt converts a string to int64 type, if the conversion fails, it ignores the error.
func toInt(v string) int64 {
	i, _ := toIntE(v)
	return i
}

// toFloatE converts a string to float64 type, if the conversion fails, it returns an error.
func toFloatE(v string) (float64, error) {
	return strconv.ParseFloat(v, 64)
}

// toFloat converts a string to float64 type, if the conversion fails, it ignores the error.
func toFloat(v string) float64 {
	f, _ := toFloatE(v)
	return f
}

// genDefaultOperationID generates a default operation ID based on the method and path.
func genDefaultOperationID(method, path string) string {
	// Remove non-alphanumeric characters from the path
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	cleanPath := reg.ReplaceAllString(path, "-")

	// Add the HTTP method to the front of the path
	operationID := strings.ToLower(method) + "-" + cleanPath

	return operationID
}

// cleanPath cleans the path pattern, removing the regular expression constraint strings within the chi parameters.
func cleanPath(pattern string) string {
	re := regexp.MustCompile(`\{(.*?):.*?\}`)
	return re.ReplaceAllString(pattern, "{$1}")
}

// UniqBy returns a duplicate-free version of an array, in which only the first occurrence of each element is kept.
// The order of result values is determined by the order they occur in the array. It accepts `iteratee` which is
// invoked for each element in array to generate the criterion by which uniqueness is computed.
func uniqBy[T any, U comparable](collection []T, iteratee func(item T) U) []T {
	result := make([]T, 0, len(collection))
	seen := make(map[U]struct{}, len(collection))

	for _, item := range collection {
		key := iteratee(item)

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		result = append(result, item)
	}

	return result
}

func sameSecurityRequirement(item *base.SecurityRequirement) string {
	var items []string
	for pair := item.Requirements.First(); pair != nil; pair = pair.Next() {
		val := pair.Value()
		sort.Strings(val)
		items = append(items, fmt.Sprintf("%s%s", pair.Key(), strings.Join(val, "")))
	}
	sort.Strings(items)
	return strings.Join(items, "")
}

func mergeMap[K comparable, V any](dst, src *orderedmap.Map[K, V]) {
	for pair := src.First(); pair != nil; pair = pair.Next() {
		dst.Set(pair.Key(), pair.Value())
	}
}
