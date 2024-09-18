package soda

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gorilla/schema"
)

func BindHeader(c *gin.Context, out any) error {
	headers := c.Request.Header
	data := make(map[string][]string, len(headers))
	var err error

	for k, vs := range headers {
		if err != nil {
			break
		}

		if strings.Contains(k, "[") {
			k, err = parseParamSquareBrackets(k)
		}

		for _, v := range vs {
			if strings.Contains(v, ",") && equalFieldType(out, reflect.Slice, k, TagHeader) {
				data[k] = append(data[k], strings.Split(v, ",")...)
			} else {
				data[k] = append(data[k], v)
			}
		}
	}

	if err != nil {
		return err
	}

	if err := parse(TagHeader, out, data); err != nil {
		return err
	}
	if binding.Validator != nil {
		return binding.Validator.ValidateStruct(out)
	}
	return nil
}

func BindQuery(c *gin.Context, out any) error {
	queries := c.Request.URL.Query()

	data := make(map[string][]string, len(queries))
	var err error

	for k, vs := range queries {
		if err != nil {
			break
		}

		if strings.Contains(k, "[") {
			k, err = parseParamSquareBrackets(k)
		}

		for _, v := range vs {
			if strings.Contains(v, ",") && equalFieldType(out, reflect.Slice, k, TagQuery) {
				data[k] = append(data[k], strings.Split(v, ",")...)
			} else {
				data[k] = append(data[k], v)
			}
		}
	}
	if err != nil {
		return err
	}

	if err := parse(TagQuery, out, data); err != nil {
		return err
	}
	if binding.Validator != nil {
		return binding.Validator.ValidateStruct(out)
	}
	return nil
}

func BindPath(c *gin.Context, out any) error {
	data := make(map[string][]string, len(c.Params))
	for _, v := range c.Params {
		data[v.Key] = []string{v.Value}
	}
	if err := parse(TagPath, out, data); err != nil {
		return err
	}
	if binding.Validator != nil {
		return binding.Validator.ValidateStruct(out)
	}
	return nil
}

func BindCookie(c *gin.Context, out any) error {
	data := make(map[string][]string, len(c.Request.Cookies()))
	for _, cookie := range c.Request.Cookies() {
		if strings.Contains(cookie.Value, ",") && equalFieldType(out, reflect.Slice, cookie.Name, TagCookie) {
			data[cookie.Name] = append(data[cookie.Name], strings.Split(cookie.Value, ",")...)
		} else {
			data[cookie.Name] = append(data[cookie.Name], cookie.Value)
		}
	}

	if err := parse(TagCookie, out, data); err != nil {
		return err
	}
	if binding.Validator != nil {
		return binding.Validator.ValidateStruct(out)
	}
	return nil
}

func parseParamSquareBrackets(k string) (string, error) {
	bb := bytes.NewBuffer(nil)
	kbytes := []byte(k)
	openBracketsCount := 0

	for i, b := range kbytes {
		if b == '[' {
			openBracketsCount++
			if i+1 < len(kbytes) && kbytes[i+1] != ']' {
				if err := bb.WriteByte('.'); err != nil {
					return "", err //nolint:wrapcheck // unnecessary to wrap it
				}
			}
			continue
		}

		if b == ']' {
			openBracketsCount--
			if openBracketsCount < 0 {
				return "", errors.New("unmatched brackets")
			}
			continue
		}

		if err := bb.WriteByte(b); err != nil {
			return "", err //nolint:wrapcheck // unnecessary to wrap it
		}
	}

	if openBracketsCount > 0 {
		return "", errors.New("unmatched brackets")
	}

	return bb.String(), nil
}

func equalFieldType(out any, kind reflect.Kind, key, tag string) bool {
	// Get type of interface
	outTyp := reflect.TypeOf(out).Elem()
	key = strings.ToLower(key)

	// Support maps
	if outTyp.Kind() == reflect.Map && outTyp.Key().Kind() == reflect.String {
		return true
	}

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
			// Is the field an embedded struct?
			if structFieldKind == reflect.Struct {
				// Loop over embedded struct fields
				for j := 0; j < structField.NumField(); j++ {
					structFieldField := structField.Field(j)

					// Can this embedded field be changed?
					if !structFieldField.CanSet() {
						continue
					}

					// Is the embedded struct field type equal to the input?
					if structFieldField.Kind() == kind {
						return true
					}
				}
			}

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

// decoderPoolMap helps to improve binders
var decoderPoolMap = map[string]*sync.Pool{}

var ErrMapNotConvertable = errors.New("binder: map is not convertable to map[string]string or map[string][]string")

func init() {
	for _, tag := range []string{TagHeader, TagCookie, TagQuery, TagPath} {
		decoderPoolMap[tag] = &sync.Pool{New: func() any {
			decoder := schema.NewDecoder()
			decoder.IgnoreUnknownKeys(true)
			decoder.ZeroEmpty(true)
			return decoder
		}}
	}
}

// parse data into the map or struct
func parse(aliasTag string, out any, data map[string][]string) error {
	ptrVal := reflect.ValueOf(out)

	// Get pointer value
	if ptrVal.Kind() == reflect.Ptr {
		ptrVal = ptrVal.Elem()
	}

	// Parse into the map
	if ptrVal.Kind() == reflect.Map && ptrVal.Type().Key().Kind() == reflect.String {
		return parseToMap(ptrVal.Interface(), data)
	}

	// Parse into the struct
	return parseToStruct(aliasTag, out, data)
}

// Parse data into the struct with gorilla/schema
func parseToStruct(aliasTag string, out any, data map[string][]string) error {
	// Get decoder from pool
	schemaDecoder := decoderPoolMap[aliasTag].Get().(*schema.Decoder) //nolint:errcheck,forcetypeassert // not needed
	defer decoderPoolMap[aliasTag].Put(schemaDecoder)

	// Set alias tag
	schemaDecoder.SetAliasTag(aliasTag)

	return schemaDecoder.Decode(out, data)
}

// Parse data into the map
// thanks to https://github.com/gin-gonic/gin/blob/master/binding/binding.go
func parseToMap(ptr any, data map[string][]string) error {
	elem := reflect.TypeOf(ptr).Elem()

	// map[string][]string
	if elem.Kind() == reflect.Slice {
		newMap, ok := ptr.(map[string][]string)
		if !ok {
			return ErrMapNotConvertable
		}

		for k, v := range data {
			newMap[k] = v
		}

		return nil
	}

	// map[string]string
	newMap, ok := ptr.(map[string]string)
	if !ok {
		return ErrMapNotConvertable
	}

	for k, v := range data {
		newMap[k] = v[len(v)-1]
	}

	return nil
}
