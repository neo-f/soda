package soda

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func ptr[T any](v T) *T {
	return &v
}

func toSlice(val, typ string) []interface{} {
	ss := strings.Split(val, SeparatorPropItem)
	result := make([]interface{}, 0, len(ss))
	var transform func(string) (interface{}, error)
	switch typ {
	case typeString:
		transform = func(s string) (interface{}, error) { return s, nil }
	case typeInteger:
		transform = func(s string) (interface{}, error) { return toIntE(s) }
	case typeNumber:
		transform = func(s string) (interface{}, error) { return toFloatE(s) }
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

func toUint(v string) uint64 {
	u, _ := strconv.ParseUint(v, 10, 64)
	return u
}

func toIntE(v string) (int, error) {
	return strconv.Atoi(v)
}

func toInt(v string) int {
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

func fixPath(path string) string {
	return regexFiberPath.ReplaceAllString(path, "/{${1}}")
}

func operationIDToCamelCase(operationID string) string {
	words := strings.Split(operationID, "-")
	for i := range words {
		words[i] = cases.Title(language.English).String(words[i])
	}
	return strings.Join(words, "")
}

func GetInput[T any](c *fiber.Ctx) *T {
	return c.Locals(KeyInput).(*T)
}
