package soda

import (
	"reflect"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// fieldResolver is a structure that contains a reflect.StructField and a map for tag pairs.
// It is used to resolve the tags of a struct field.
type fieldResolver struct {
	t        *reflect.StructField
	tagPairs map[string]string
}

// newFieldResolver creates a new fieldResolver from a reflect.StructField.
// The fieldResolver will be used to determine the name of the field in the
// OpenAPI schema.
func newFieldResolver(t *reflect.StructField) *fieldResolver {
	// Initialize a new fieldResolver
	resolver := &fieldResolver{t: t, tagPairs: nil}
	// Look up the OpenAPI tags
	if oaiTags, oaiOK := t.Tag.Lookup(OpenAPITag); oaiOK {
		if oaiTags == "-" {
			return resolver
		}
		// Create a map for the tag pairs
		resolver.tagPairs = make(map[string]string)
		// Split the tags and store them in the map
		for _, tag := range strings.Split(oaiTags, SeparatorProp) {
			tag = strings.TrimSpace(tag)
			pair := strings.Split(tag, "=")
			if len(pair) == 2 {
				resolver.tagPairs[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
			} else {
				resolver.tagPairs[strings.TrimSpace(pair[0])] = ""
			}
		}
	}
	return resolver
}

// injectOAITags injects OAI tags into a schema.
func (f *fieldResolver) injectOAITags(schema *base.Schema) {
	// Inject generic OAI tags
	f.injectOAIGeneric(schema)
	if len(schema.Type) == 0 {
		return
	}
	// Inject specific OAI tags based on the schema type
	switch schema.Type[0] {
	case typeString:
		f.injectOAIString(schema)
	case typeNumber, typeInteger:
		f.injectOAINumeric(schema)
	case typeArray:
		f.injectOAIArray(schema)
	case typeBoolean:
		f.injectOAIBoolean(schema)
	}
}

// required checks if the field is required.
func (f fieldResolver) required() bool {
	// By default, a field is required if it is not a pointer
	required := f.t.Type.Kind() != reflect.Ptr
	// Check the 'required' tag
	if v, ok := f.tagPairs[propRequired]; ok {
		required = toBool(v)
	}
	// Check the 'nullable' tag
	if v, ok := f.tagPairs[propNullable]; ok {
		required = !toBool(v)
	}
	return required
}

// name returns the name of the field.
// If the field is tagged with the specified tag, then that tag is used instead.
// If the tag contains a comma, then only the first part of the tag is used.
func (f fieldResolver) name(tag ...string) string {
	if len(tag) > 0 {
		if name := f.t.Tag.Get(tag[0]); name != "" {
			return strings.Split(name, ",")[0]
		}
	}
	return f.t.Name
}

// injectOAIGeneric injects generic OAI tags into a schema.
func (f *fieldResolver) injectOAIGeneric(schema *base.Schema) {
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.tagPairs {
		switch tag {
		case propTitle:
			schema.Title = val
		case propDescription:
			schema.Description = val
		case propType:
			schema.Type = []string{val}
		case propDeprecated:
			schema.Deprecated = ptr(toBool(val))
		case propWriteOnly:
			schema.WriteOnly = toBool(val)
		case propReadOnly:
			schema.ReadOnly = toBool(val)
		}
	}
}

// injectOAIString injects OAI tags for string type into a schema.
func (f *fieldResolver) injectOAIString(schema *base.Schema) {
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.tagPairs {
		switch tag {
		case propMinLength:
			schema.MinLength = ptr(toInt(val))
		case propMaxLength:
			schema.MaxLength = ptr(toInt(val))
		case propPattern:
			schema.Pattern = val
		case propFormat:
			schema.Format = val
		case propEnum:
			schema.Enum = toSlice(val, typeString)
		case propDefault:
			schema.Default = val
		case propExample:
			schema.Example = val
		}
	}
}

// injectOAINumeric injects OAI tags for numeric type into a schema.
func (f *fieldResolver) injectOAINumeric(schema *base.Schema) { //nolint
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.tagPairs {
		switch tag {
		case propMultipleOf:
			schema.MultipleOf = ptr(toFloat(val))
		case propMinimum:
			schema.Minimum = ptr(toFloat(val))
		case propMaximum:
			schema.Maximum = ptr(toFloat(val))
		case propExclusiveMaximum:
			if num, err := toFloatE(val); err == nil {
				schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{B: num}
				continue
			}
			schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{A: toBool(val)}
		case propExclusiveMinimum:
			if num, err := toFloatE(val); err == nil {
				schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{B: num}
				continue
			}
			schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{A: toBool(val)}
		case propDefault:
			switch schema.Type[0] {
			case typeInteger:
				schema.Default = toInt(val)
			case typeNumber:
				schema.Default = toFloat(val)
			}
		case propEnum:
			schema.Enum = toSlice(val, schema.Type[0])
		case propExample:
			schema.Example = toInt(val)
		}
	}
}

// injectOAIArray injects OAI tags for array type into a schema.
func (f *fieldResolver) injectOAIArray(schema *base.Schema) {
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.tagPairs {
		switch tag {
		case propMinItems:
			schema.MinItems = ptr(toInt(val))
		case propMaxItems:
			schema.MaxItems = ptr(toInt(val))
		case propUniqueItems:
			schema.UniqueItems = ptr(toBool(val))
		}
	}
}

// injectOAIBoolean injects OAI tags for boolean type into a schema.
func (f *fieldResolver) injectOAIBoolean(schema *base.Schema) {
	if val, ok := f.tagPairs[propDefault]; ok {
		schema.Default = toBool(val)
	}

	for tag, val := range f.tagPairs {
		switch tag {
		case propDefault:
			schema.Default = toBool(val)
		case propExample:
			schema.Example = toBool(val)
		}
	}
}
