package soda

import "github.com/getkin/kin-openapi/openapi3"

const (
	OpenAPITag        = "oai"
	SeparatorProp     = ";"
	SeparatorPropItem = ","
)

// parameter props.
const (
	PropExplode = "explode"
	PropStyle   = "style"
)

// schema props.
const (
	// generic properties.
	PropTitle           = "title"
	PropDescription     = "description"
	PropType            = "type"
	PropDeprecated      = "deprecated"
	PropAllowEmptyValue = "allowEmptyValue"
	PropNullable        = "nullable"
	PropReadOnly        = "readOnly"
	PropWriteOnly       = "writeOnly"
	PropEnum            = "enum"
	PropDefault         = "default"
	PropExample         = "example"
	PropRequired        = "required"
	// string specified properties.
	PropMinLength = "minLength"
	PropMaxLength = "maxLength"
	PropPattern   = "pattern"
	PropFormat    = "format"
	// number specified properties.
	PropMultipleOf       = "multipleOf"
	PropMinimum          = "minimum"
	PropMaximum          = "maximum"
	PropExclusiveMaximum = "exclusiveMaximum"
	PropExclusiveMinimum = "exclusiveMinimum"
	// array specified properties.
	PropMinItems    = "minItems"
	PropMaxItems    = "maxItems"
	PropUniqueItems = "uniqueItems"
)

const (
	TypeBoolean = openapi3.TypeBoolean
	TypeNumber  = openapi3.TypeNumber
	TypeString  = openapi3.TypeString
	TypeInteger = openapi3.TypeInteger
	TypeArray   = openapi3.TypeArray
	TypeObject  = openapi3.TypeObject
)

const (
	KeyParameter   = "soda::parameters"
	KeyRequestBody = "soda::request_body"
)
