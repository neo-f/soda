package soda

import (
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// tagsResolver is a structure that contains a reflect.StructField and a map for tag pairs.
// It is used to resolve the tags of a struct field.
type tagsResolver struct {
	f     reflect.StructField
	pairs map[string]string
}

// newTagsResolver creates a new fieldResolver from a reflect.StructField.
// The fieldResolver will be used to determine the name of the field in the
// OpenAPI schema.
func newTagsResolver(f reflect.StructField) *tagsResolver {
	// Initialize a new fieldResolver
	resolver := &tagsResolver{f: f, pairs: nil}
	// Look up the OpenAPI tags
	if oaiTags, oaiOK := f.Tag.Lookup(OpenAPITag); oaiOK {
		// Create a map for the tag pairs
		resolver.pairs = make(map[string]string)
		// Split the tags and store them in the map
		for _, tag := range strings.Split(oaiTags, SeparatorProp) {
			tag = strings.TrimSpace(tag)
			k, v, _ := strings.Cut(tag, "=")
			resolver.pairs[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return resolver
}

// injectOAITags injects OAI tags into a schema.
func (f tagsResolver) injectOAITags(schema *openapi3.Schema) {
	// Inject generic OAI tags
	f.injectOAIGeneric(schema)
	// if schema.Type == nil || len(schema.Type.Slice()) == 0 {
	// 	return
	// }

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
func (f tagsResolver) required() bool {
	// By default, a field is required if it is not a pointer
	required := f.f.Type.Kind() != reflect.Ptr
	// Check the 'required' tag
	if v, ok := f.pairs[propRequired]; ok {
		required = toBool(v)
	}
	return required
}

// name returns the name of the field.
// If the field is tagged with the specified tag, then that tag is used instead.
// If the tag contains a comma, then only the first part of the tag is used.
func (f tagsResolver) name(tag ...string) string {
	if len(tag) > 0 {
		if name := f.f.Tag.Get(tag[0]); name != "" {
			return strings.Split(name, ",")[0]
		}
	}
	return f.f.Name
}

// injectOAIGeneric injects generic OAI tags into a schema.
func (f *tagsResolver) injectOAIGeneric(schema *openapi3.Schema) {
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.pairs {
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
		case propNullable:
			schema.Nullable = toBool(val)
		}
	}
}

// injectOAIString injects OAI tags for string type into a schema.
func (f *tagsResolver) injectOAIString(schema *openapi3.Schema) {
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.pairs {
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
func (f *tagsResolver) injectOAINumeric(schema *openapi3.Schema) { //nolint
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.pairs {
		switch tag {
		case propMultipleOf:
			if num, err := toFloatE(val); err == nil {
				schema.MultipleOf = &num
			}
		case propMinimum, propMin:
			if num, err := toFloatE(val); err == nil {
				schema.Min = &num
			}
		case propExclusiveMinimum:
			schema.ExclusiveMin = toBool(val)
		case propMaximum, propMax:
			if num, err := toFloatE(val); err == nil {
				schema.Max = &num
			}
		case propExclusiveMaximum:
			schema.ExclusiveMax = toBool(val)
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
func (f *tagsResolver) injectOAIBoolean(schema *openapi3.Schema) {
	for tag, val := range f.pairs {
		switch tag {
		case propDefault:
			schema.Default = toBool(val)
		case propExample:
			schema.Example = toBool(val)
		}
	}
}

// injectOAIArray injects OAI tags for array type into a schema.
func (f *tagsResolver) injectOAIArray(schema *openapi3.Schema) {
	// Iterate over the tag pairs and inject them into the schema
	for tag, val := range f.pairs {
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
