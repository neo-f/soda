package soda

import "github.com/getkin/kin-openapi/openapi3"

var (
	OpenAPITag        = "oai"
	SeparatorProp     = ";"
	SeparatorPropItem = ","
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
	propMaximum          = "maximum"
	PrppExclusiveMaximum = "exclusiveMaximum"
	PrppExclusiveMinimum = "exclusiveMinimum"
	// array specified properties.
	propMinItems    = "minItems"
	propMaxItems    = "maxItems"
	propUniqueItems = "uniqueItems"
)

const (
	typeBoolean = openapi3.TypeBoolean
	typeNumber  = openapi3.TypeNumber
	typeString  = openapi3.TypeString
	typeInteger = openapi3.TypeInteger
	typeArray   = openapi3.TypeArray
)

const (
	KeyInput = "soda::input"
)
