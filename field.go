package soda

import (
	"reflect"
	"strings"

	"github.com/sv-tools/openapi/spec"
)

type fieldResolver struct {
	t        *reflect.StructField
	tagPairs map[string]string
}

// newFieldResolver creates a new fieldResolver from a reflect.StructField.
// The fieldResolver will be used to determine the name of the field in the
// OpenAPI schema.
func newFieldResolver(t *reflect.StructField) *fieldResolver {
	resolver := &fieldResolver{t: t, tagPairs: nil}
	if oaiTags, oaiOK := t.Tag.Lookup(OpenAPITag); oaiOK {
		if oaiTags == "-" {
			return resolver
		}
		resolver.tagPairs = make(map[string]string)
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
func (f *fieldResolver) injectOAITags(schema *spec.Schema) {
	f.injectOAIGeneric(schema)
	if len(schema.Type) == 0 {
		return
	}
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

func (f fieldResolver) required() bool {
	required := f.t.Type.Kind() != reflect.Ptr
	if v, ok := f.tagPairs[propRequired]; ok {
		required = toBool(v)
	}
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

func (f *fieldResolver) injectOAIGeneric(schema *spec.Schema) {
	for tag, val := range f.tagPairs {
		switch tag {
		case propTitle:
			schema.Title = val
		case propDescription:
			schema.Description = val
		case propType:
			schema.Type = spec.NewSingleOrArray(val)
		case propDeprecated:
			schema.Deprecated = toBool(val)
		case propWriteOnly:
			schema.WriteOnly = toBool(val)
		case propReadOnly:
			schema.ReadOnly = toBool(val)
		}
	}
}

// read struct tags for string type keywords.
func (f *fieldResolver) injectOAIString(schema *spec.Schema) {
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
		case propExamples:
			schema.Examples = toSlice(val, typeString)
		}
	}
}

// read struct tags for numeric type keywords.
func (f *fieldResolver) injectOAINumeric(schema *spec.Schema) { //nolint
	for tag, val := range f.tagPairs {
		switch tag {
		case propMultipleOf:
			schema.MultipleOf = ptr(toInt(val))
		case propMinimum:
			schema.Minimum = ptr(toInt(val))
		case propMaximum:
			schema.Maximum = ptr(toInt(val))
		case propExclusiveMaximum:
			schema.ExclusiveMaximum = ptr(toInt(val))
		case propExclusiveMinimum:
			schema.ExclusiveMinimum = ptr(toInt(val))
		case propDefault:
			switch schema.Type[0] {
			case typeInteger:
				schema.Default = toInt(val)
			case typeNumber:
				schema.Default = toFloat(val)
			}
		case propEnum:
			schema.Enum = toSlice(val, schema.Type[0])
		case propExamples:
			switch schema.Type[0] {
			case typeInteger:
				schema.Examples = toSlice(val, typeInteger)
			case typeNumber:
				schema.Examples = toSlice(val, typeNumber)
			}
		}
	}
}

// read struct tags for bool type keywords.
func (f *fieldResolver) injectOAIBoolean(schema *spec.Schema) {
	if val, ok := f.tagPairs[propDefault]; ok {
		schema.Default = toBool(val)
	}
	for tag, val := range f.tagPairs {
		switch tag {
		case propDefault:
			schema.Default = toBool(val)
		case propExamples:
			schema.Examples = toSlice(val, typeBoolean)
		}
	}
}

// read struct tags for array type keywords.
func (f *fieldResolver) injectOAIArray(schema *spec.Schema) {
	for tag, val := range f.tagPairs {
		switch tag {
		case propMinItems:
			schema.MinItems = ptr(toInt(val))
		case propMaxItems:
			schema.MaxItems = ptr(toInt(val))
		case propUniqueItems:
			schema.UniqueItems = ptr(toBool(val))
		case propDefault:
			schema.Default = toSlice(val, schema.Items.Schema.Spec.Type[0])
		case propEnum:
			schema.Enum = toSlice(val, schema.Items.Schema.Spec.Type[0])
		}
	}
}
