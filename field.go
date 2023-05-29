package soda

import (
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type fieldResolver struct {
	t        *reflect.StructField
	tagPairs map[string]string
	ignored  bool
}

// newFieldResolver creates a new fieldResolver from a reflect.StructField.
// The fieldResolver will be used to determine the name of the field in the
// OpenAPI schema.
func newFieldResolver(t *reflect.StructField) *fieldResolver {
	resolver := &fieldResolver{
		t:        t,
		ignored:  false,
		tagPairs: nil,
	}
	if oaiTags, oaiOK := t.Tag.Lookup(OpenAPITag); oaiOK {
		tags := strings.Split(oaiTags, SeparatorProp)
		if tags[0] == "-" {
			resolver.ignored = true
			return resolver
		}
		resolver.tagPairs = make(map[string]string)
		for _, tag := range tags {
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
func (f *fieldResolver) injectOAITags(schema *openapi3.Schema) {
	f.injectOAIGeneric(schema)
	switch schema.Type {
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

func (f fieldResolver) shouldEmbed() bool {
	return f.t.Anonymous && !f.ignored
}

func (f *fieldResolver) injectOAIGeneric(schema *openapi3.Schema) {
	for tag, val := range f.tagPairs {
		switch tag {
		case propTitle:
			schema.Title = val
		case propDescription:
			schema.Description = val
		case propType:
			schema.Type = val
		case propDeprecated:
			schema.Deprecated = toBool(val)
		case propAllowEmptyValue:
			schema.AllowEmptyValue = toBool(val)
		case propNullable:
			schema.Nullable = toBool(val)
		case propWriteOnly:
			schema.WriteOnly = toBool(val)
		case propReadOnly:
			schema.ReadOnly = toBool(val)
		}
	}
}

// read struct tags for string type keywords.
func (f *fieldResolver) injectOAIString(schema *openapi3.Schema) {
	for tag, val := range f.tagPairs {
		switch tag {
		case propMinLength:
			schema.MinLength = toUint(val)
		case propMaxLength:
			schema.MaxLength = ptr(toUint(val))
		case propPattern:
			schema.Pattern = val
		case propFormat:
			switch val {
			case "date-time", "date", "email", "hostname", "ipv4", "ipv6", "uri":
				schema.Format = val
			}
		case propEnum:
			for _, item := range strings.Split(val, SeparatorPropItem) {
				schema.Enum = append(schema.Enum, item)
			}
		case propDefault:
			schema.Default = val
		case propExample:
			schema.Example = val
		}
	}
}

// read struct tags for numeric type keywords.
func (f *fieldResolver) injectOAINumeric(schema *openapi3.Schema) { //nolint
	for tag, val := range f.tagPairs {
		switch tag {
		case propMultipleOf:
			schema.MultipleOf = ptr(toFloat(val))
		case propMinimum:
			schema.Min = ptr(toFloat(val))
		case propMaximum:
			schema.Max = ptr(toFloat(val))
		case PrppExclusiveMaximum:
			schema.ExclusiveMax = toBool(val)
		case PrppExclusiveMinimum:
			schema.ExclusiveMin = toBool(val)
		case propDefault:
			switch schema.Type {
			case typeInteger:
				schema.Default = toInt(val)
			case typeNumber:
				schema.Default = toFloat(val)
			}
		case propExample:
			switch schema.Type {
			case typeInteger:
				schema.Example = toInt(val)
			case typeNumber:
				schema.Example = toFloat(val)
			case typeBoolean:
				schema.Example = toBool(val)
			}
		case propEnum:
			items := strings.Split(val, SeparatorPropItem)
			switch schema.Type {
			case typeInteger:
				for _, item := range items {
					schema.Enum = append(schema.Enum, toInt(item))
				}
			case typeNumber:
				for _, item := range items {
					schema.Enum = append(schema.Enum, toFloat(item))
				}
			}
		}
	}
}

// read struct tags for array type keywords.
func (f *fieldResolver) injectOAIArray(schema *openapi3.Schema) {
	for tag, val := range f.tagPairs {
		switch tag {
		case propMinItems:
			schema.MinItems = toUint(val)
		case propMaxItems:
			schema.MaxItems = ptr(toUint(val))
		case propUniqueItems:
			schema.UniqueItems = toBool(val)
		case propDefault, propEnum, propExample:
			items := toSlice(val, tag)
			switch tag {
			case propDefault:
				schema.Default = items
			case propExample:
				schema.Example = items
			case propEnum:
				schema.Enum = []interface{}{items}
			}
		}
	}
}

// read struct tags for bool type keywords.
func (f *fieldResolver) injectOAIBoolean(schema *openapi3.Schema) {
	if val, ok := f.tagPairs[propDefault]; ok {
		schema.Default = toBool(val)
	}
	if val, ok := f.tagPairs[propExample]; ok {
		schema.Example = toBool(val)
	}
}
