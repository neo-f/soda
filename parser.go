package soda

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gorilla/schema"
)

type parserFunc func(*fiber.Ctx, interface{}) error

var parameterParsers = map[string]parserFunc{
	"query":  parseQuery,
	"header": parseHeader,
	"path":   parsePath,
	"cookie": parseCookie,
}

// parseQuery parses query parameters into a struct.
func parseQuery(c *fiber.Ctx, out interface{}) error {
	data := make(map[string][]string)
	c.Request().URI().QueryArgs().VisitAll(func(key, val []byte) {
		k := utils.UnsafeString(key)
		v := utils.UnsafeString(val)
		if strings.Contains(v, ",") {
			values := strings.Split(v, ",")
			for i := 0; i < len(values); i++ {
				data[k] = append(data[k], values[i])
			}
		} else {
			data[k] = append(data[k], v)
		}
	})
	return mapToStruct("query", out, data)
}

// parseHeader parses header parameters into a struct.
func parseHeader(c *fiber.Ctx, out interface{}) error {
	data := make(map[string][]string)
	c.Request().Header.VisitAll(func(key, val []byte) {
		k := utils.UnsafeString(key)
		v := utils.UnsafeString(val)
		if strings.Contains(v, ",") {
			values := strings.Split(v, ",")
			for i := 0; i < len(values); i++ {
				data[k] = append(data[k], values[i])
			}
		} else {
			data[k] = append(data[k], v)
		}
	})
	return mapToStruct("header", out, data)
}

// parsePath parses path parameters into a struct.
func parsePath(c *fiber.Ctx, out interface{}) error {
	data := make(map[string][]string)
	for _, k := range c.Route().Params {
		data[k] = []string{c.Params(k)}
	}
	return mapToStruct("path", out, data)
}

// parseCookie parses cookie parameters into a struct.
func parseCookie(c *fiber.Ctx, out interface{}) error {
	data := make(map[string][]string)
	c.Request().Header.VisitAllCookie(func(key, val []byte) {
		k := utils.UnsafeString(key)
		v := utils.UnsafeString(val)
		if strings.Contains(v, ",") {
			values := strings.Split(v, ",")
			for i := 0; i < len(values); i++ {
				data[k] = append(data[k], values[i])
			}
		} else {
			data[k] = append(data[k], v)
		}
	})
	return mapToStruct("cookie", out, data)
}

// mapToStruct converts a map[string][]string to a struct.
func mapToStruct(aliasTag string, out interface{}, data map[string][]string) error {
	// Get decoder from pool
	decoder := schema.NewDecoder()
	decoder.SetAliasTag(aliasTag)
	decoder.IgnoreUnknownKeys(true)
	// Set alias tag
	return decoder.Decode(out, data)
}
