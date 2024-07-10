package soda

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
)

// ptr creates a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// toSlice converts a string to a slice, the type of conversion is determined by the typ parameter.
func toSlice(val string, typ string) []any {
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
func toIntE(v string) (int, error) {
	return strconv.Atoi(v)
}

func toUint64E(v string) (uint64, error) {
	return strconv.ParseUint(v, 10, 64)
}

// toFloatE converts a string to float64 type, if the conversion fails, it returns an error.
func toFloatE(v string) (float64, error) {
	return strconv.ParseFloat(v, 64)
}

// genDefaultOperationID generates a default operation ID based on the method and path.
func genDefaultOperationID(method, path string) string {
	// Remove non-alphanumeric characters from the path
	cleanPath := regexOperationID.ReplaceAllString(path, "-")

	// Add the HTTP method to the front of the path
	operationID := strings.ToLower(method) + "-" + cleanPath

	return operationID
}

// cleanPath cleans the path pattern, removing the regular expression constraint strings within the chi TestCase.
func cleanPath(pattern string) string {
	re := regexp.MustCompile(`\{(.*?):.*?\}`)
	return re.ReplaceAllString(pattern, "{$1}")
}

func derefSchema(doc *openapi3.T, schemaRef *openapi3.SchemaRef) *openapi3.Schema {
	// return schemaRef.Value
	if schemaRef.Value != nil {
		return schemaRef.Value
	}
	if schemaRef.Ref != "" {
		full := schemaRef.Ref
		name := path.Base(full)
		schema, ok := doc.Components.Schemas[name]
		if !ok {
			panic(fmt.Sprintf("deref schema failed: schema %s not found in document", name))
		}
		return derefSchema(doc, schema)
	}
	panic("deref schema failed")
}

// GetInput gets the input value from the http request.
func GetInput[T any](c *fiber.Ctx) *T {
	return c.Locals(KeyInput).(*T)
}
