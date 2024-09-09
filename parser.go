package soda

import (
	"reflect"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gorilla/schema"
)

var decoderPools = map[string]*sync.Pool{
	"path":   {New: func() any { return buildDecoder("path") }},
	"header": {New: func() any { return buildDecoder("header") }},
}

func buildDecoder(tag string) *schema.Decoder {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag(tag)
	decoder.IgnoreUnknownKeys(true)
	decoder.ZeroEmpty(true)
	return decoder
}

func bindPath(c *fiber.Ctx) func(any) error {
	return func(out any) error {
		params := c.Route().Params
		data := make(map[string][]string, len(params))
		for _, param := range params {
			data[param] = append(data[param], c.Params(param))
		}

		pathDecoder := decoderPools["path"].Get().(*schema.Decoder)
		defer decoderPools["path"].Put(pathDecoder)
		return pathDecoder.Decode(out, data)
	}
}

func bindHeader(c *fiber.Ctx) func(any) error {
	return func(out any) error {
		data := make(map[string][]string)
		c.Request().Header.VisitAll(func(key, val []byte) {
			k := string(key)
			v := string(val)

			if c.App().Config().EnableSplittingOnParsers && strings.Contains(v, ",") && equalFieldType(out, reflect.Slice, k, "header") {
				values := strings.Split(v, ",")
				for i := 0; i < len(values); i++ {
					data[k] = append(data[k], values[i])
				}
			} else {
				data[k] = append(data[k], v)
			}
		})

		headerDecoder := decoderPools["header"].Get().(*schema.Decoder)
		defer decoderPools["header"].Put(headerDecoder)
		return headerDecoder.Decode(out, data)
	}
}

// steal from fiber ;)
func equalFieldType(out interface{}, kind reflect.Kind, key, tag string) bool {
	// Get type of interface
	outTyp := reflect.TypeOf(out).Elem()
	key = strings.ToLower(key)
	// Must be a struct to match a field
	if outTyp.Kind() != reflect.Struct {
		return false
	}
	// Copy interface to an value to be used
	outVal := reflect.ValueOf(out).Elem()
	// Loop over each field
	for i := 0; i < outTyp.NumField(); i++ {
		// Get field value data
		structField := outVal.Field(i)
		// Can this field be changed?
		if !structField.CanSet() {
			continue
		}
		// Get field key data
		typeField := outTyp.Field(i)
		// Get type of field key
		structFieldKind := structField.Kind()
		// Does the field type equals input?
		if structFieldKind != kind {
			continue
		}
		// Get tag from field if exist
		inputFieldName := typeField.Tag.Get(tag)
		if inputFieldName == "" {
			inputFieldName = typeField.Name
		} else {
			inputFieldName = strings.Split(inputFieldName, ",")[0]
		}
		// Compare field/tag with provided key
		if strings.ToLower(inputFieldName) == key {
			return true
		}
	}
	return false
}
