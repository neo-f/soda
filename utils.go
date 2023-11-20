package soda

import (
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ptr creates a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// unptr gets the value from the pointer. If the pointer is nil, it returns the zero value of that type.
func unptr[T any](v *T) T {
	if v == nil {
		return reflect.Zero(reflect.TypeOf(v)).Interface().(T)
	}
	return *v
}

// toSlice converts a string to a slice, the type of conversion is determined by the typ parameter.
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

// appendUniq adds elements to the slice, but only if the element does not already exist in the slice.
func appendUniq[T comparable](slice []T, elems ...T) []T {
	seen := make(map[T]bool)
	for _, v := range slice {
		seen[v] = true
	}

	for _, elem := range elems {
		if !seen[elem] {
			slice = append(slice, elem)
			seen[elem] = true
		}
	}
	return slice
}

// cleanPath cleans the path pattern, removing the regular expression constraint strings within the chi parameters.
func cleanPath(pattern string) string {
	re := regexp.MustCompile(`\{(.*?):.*?\}`)
	return re.ReplaceAllString(pattern, "{$1}")
}

// GetInput gets the input value from the http request.
func GetInput[T any](c *http.Request) *T {
	return c.Context().Value(KeyInput).(*T)
}
