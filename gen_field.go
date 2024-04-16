package soda

import (
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// fieldResolver is a structure that contains a reflect.StructField and a map for tag pairs.
// It is used to resolve the tags of a struct field.
type fieldResolver struct {
	t        reflect.StructField
	tagPairs map[string]string
}

// newFieldResolver creates a new fieldResolver from a reflect.StructField.
// The fieldResolver will be used to determine the name of the field in the
// OpenAPI schema.
func newFieldResolver(t reflect.StructField) *fieldResolver {
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
			k, v, _ := strings.Cut(tag, "=")
			resolver.tagPairs[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return resolver
}

// injectOAITags injects OAI tags into a schema.
func (f *fieldResolver) injectOAITags(schema *openapi3.Schema) {
	// Inject generic OAI tags
	f.injectOAIGeneric(schema)
	if schema.Type == nil || len(schema.Type.Slice()) == 0 {
		return
	}

	// Inject specific OAI tags based on the schema type
	switch {
	case schema.Type.Is(typeString):
		f.injectOAIString(schema)
	case schema.Type.Is(typeNumber), schema.Type.Is(typeInteger):
		f.injectOAINumeric(schema)
	case schema.Type.Is(typeArray):
		f.injectOAIArray(schema)
	case schema.Type.Is(typeBoolean):
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
func (f *fieldResolver) injectOAIGeneric(schema *openapi3.Schema) {
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.tagPairs {
		switch tag {
		case propTitle:
			schema.Title = val
		case propDescription:
			schema.Description = val
		case propDeprecated:
			schema.Deprecated = toBool(val)
		case propWriteOnly:
			schema.WriteOnly = toBool(val)
		case propReadOnly:
			schema.ReadOnly = toBool(val)
		}
	}
}

// injectOAIString injects OAI tags for string type into a schema.
func (f *fieldResolver) injectOAIString(schema *openapi3.Schema) {
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.tagPairs {
		switch tag {
		case propMinLength:
			if num, err := toUint64E(val); err == nil {
				schema.MinLength = num
			}
		case propMaxLength:
			if num, err := toUint64E(val); err == nil {
				schema.MaxLength = &num
			}
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
func (f *fieldResolver) injectOAINumeric(schema *openapi3.Schema) { //nolint
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.tagPairs {
		switch tag {
		case propMultipleOf:
			if num, err := toFloatE(val); err == nil {
				schema.MultipleOf = &num
			}
		case propMinimum:
			if num, err := toFloatE(val); err == nil {
				schema.Min = &num
			}
		case propMaximum:
			if num, err := toFloatE(val); err == nil {
				schema.Max = &num
			}
		case propExclusiveMaximum:
			if num, err := toFloatE(val); err == nil {
				schema.Max = ptr(num)
				schema.ExclusiveMax = true
			}
		case propExclusiveMinimum:
			if num, err := toFloatE(val); err == nil {
				schema.Min = ptr(num)
				schema.ExclusiveMin = true
			}
		case propEnum:
			schema.Enum = toSlice(val, schema.Type.Slice()[0])
		case propDefault:
			switch {
			case schema.Type.Is(typeInteger):
				if num, err := toIntE(val); err == nil {
					schema.Default = num
				}
			case schema.Type.Is(typeNumber):
				if num, err := toFloatE(val); err == nil {
					schema.Default = num
				}
			}
		case propExample:
			switch {
			case schema.Type.Is(typeInteger):
				if num, err := toIntE(val); err == nil {
					schema.Example = num
				}
			case schema.Type.Is(typeNumber):
				if num, err := toFloatE(val); err == nil {
					schema.Example = num
				}
			}
		}
	}
}

// injectOAIBoolean injects OAI tags for boolean type into a schema.
func (f *fieldResolver) injectOAIBoolean(schema *openapi3.Schema) {
	for tag, val := range f.tagPairs {
		switch tag {
		case propDefault:
			schema.Default = toBool(val)
		case propExample:
			schema.Example = toBool(val)
		}
	}
}

// injectOAIArray injects OAI tags for array type into a schema.
func (f *fieldResolver) injectOAIArray(schema *openapi3.Schema) {
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.tagPairs {
		switch tag {
		case propMinItems:
			if num, err := toUint64E(val); err == nil {
				schema.MinItems = num
			}
		case propMaxItems:
			if num, err := toUint64E(val); err == nil {
				schema.MaxItems = &num
			}
		case propUniqueItems:
			schema.UniqueItems = toBool(val)
		}
	}
}
