package soda

import (
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	TagOpenAPI        = "oai"
	SeparatorProp     = ";"
	SeparatorPropItem = ","

	TagHeader = openapi3.ParameterInHeader
	TagQuery  = openapi3.ParameterInQuery
	TagCookie = openapi3.ParameterInCookie
	TagPath   = openapi3.ParameterInPath
)

// parameter props.
const (
	propExplode = "explode"
	propStyle   = "style"
)

// schema props.
const (
	// generic properties.
	propTitle           = "title"
	propDescription     = "description"
	propType            = "type"
	propDeprecated      = "deprecated"
	propAllowEmptyValue = "allowEmptyValue"
	propNullable        = "nullable"
	propReadOnly        = "readOnly"
	propWriteOnly       = "writeOnly"
	propEnum            = "enum"
	propDefault         = "default"
	propExample         = "example"
	propRequired        = "required"
	// string specified properties.
	propMinLength = "minLength"
	propMaxLength = "maxLength"
	propPattern   = "pattern"
	propFormat    = "format"
	// number specified properties.
	propMultipleOf       = "multipleOf"
	propMinimum          = "minimum"
	propMin              = "min"
	propMaximum          = "maximum"
	propMax              = "max"
	propExclusiveMaximum = "exclusiveMaximum"
	propExclusiveMinimum = "exclusiveMinimum"
	// array specified properties.
	propMinItems    = "minItems"
	propMaxItems    = "maxItems"
	propUniqueItems = "uniqueItems"
)

const KeyInput = "soda::input"

const (
	typeArray   = "array"
	typeBoolean = "boolean"
	typeInteger = "integer"
	typeNumber  = "number"
	typeObject  = "object"
	typeString  = "string"
)

var (
	regexOperationID = regexp.MustCompile("[^a-zA-Z0-9]+")
	regexSchemaName  = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
)
