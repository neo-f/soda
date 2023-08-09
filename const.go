package soda

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
	propExclusiveMaximum = "exclusiveMaximum"
	propExclusiveMinimum = "exclusiveMinimum"
	// array specified properties.
	propMinItems    = "minItems"
	propMaxItems    = "maxItems"
	propUniqueItems = "uniqueItems"
)

const (
	KeyInput = "soda::input"
)

const (
	typeArray   = "array"
	typeBoolean = "boolean"
	typeInteger = "integer"
	typeNumber  = "number"
	typeObject  = "object"
	typeString  = "string"
)
