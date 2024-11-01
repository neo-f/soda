package soda

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gorilla/schema"
)

var (
	decoderQuery  = newDecoder("query")
	decoderHeader = newDecoder("header")
	decoderPath   = newDecoder("path")
	decoderCookie = newDecoder("cookie")
)

func newDecoder(tag string) *schema.Decoder {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag(tag)
	decoder.IgnoreUnknownKeys(true)
	decoder.ZeroEmpty(true)
	return decoder
}

// parseQuery parses query parameters into a struct.
func parseQuery(c *fiber.Ctx, out interface{}, types map[[2]string]string) error {
	data := make(map[string][]string)
	c.Request().URI().QueryArgs().VisitAll(func(key, val []byte) {
		k := utils.UnsafeString(key)
		v := utils.UnsafeString(val)
		if types[[2]string{"query", k}] == "array" && strings.Contains(v, ",") {
			data[k] = append(data[k], strings.Split(v, ",")...)
		} else {
			data[k] = append(data[k], v)
		}
	})
	return decoderQuery.Decode(out, data)
}

// parseHeader parses header parameters into a struct.
func parseHeader(c *fiber.Ctx, out interface{}, types map[[2]string]string) error {
	data := make(map[string][]string)
	c.Request().Header.VisitAll(func(key, val []byte) {
		k := utils.UnsafeString(key)
		v := utils.UnsafeString(val)
		if types[[2]string{"header", k}] == "array" && strings.Contains(v, ",") {
			data[k] = append(data[k], strings.Split(v, ",")...)
		} else {
			data[k] = append(data[k], v)
		}
	})
	return decoderHeader.Decode(out, data)
}

// parsePath parses path parameters into a struct.
func parsePath(c *fiber.Ctx, out interface{}, _ map[[2]string]string) error {
	data := make(map[string][]string)
	for _, k := range c.Route().Params {
		data[k] = []string{c.Params(k)}
	}
	return decoderPath.Decode(out, data)
}

// parseCookie parses cookie parameters into a struct.
func parseCookie(c *fiber.Ctx, out interface{}, types map[[2]string]string) error {
	data := make(map[string][]string)
	c.Request().Header.VisitAllCookie(func(key, val []byte) {
		k := utils.UnsafeString(key)
		v := utils.UnsafeString(val)
		if types[[2]string{"cookie", k}] == "array" && strings.Contains(v, ",") {
			data[k] = append(data[k], strings.Split(v, ",")...)
		} else {
			data[k] = append(data[k], v)
		}
	})
	return decoderCookie.Decode(out, data)
}
