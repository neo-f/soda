package soda

import (
	"context"
	"encoding/json"
	"math"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// Define some well-known types.
var (
	wnTime         = reflect.TypeOf(time.Time{})       // date-time RFC section 8.3.1
	wnIP           = reflect.TypeOf(net.IP{})          // ipv4 and ipv6 RFC section 7.3.4, 7.3.5
	wnByteSlice    = reflect.TypeOf([]byte(nil))       // Byte slices will be encoded as base64
	wnJSON         = reflect.TypeOf(json.RawMessage{}) // Except for json.RawMessage
	wnMapStringAny = reflect.TypeOf(map[string]any{})  // Except for map[string]any
)

// Define an interface for JSON schema generation.
type jsonSchema interface {
	JSONSchema(*openapi3.T) *openapi3.SchemaRef
}

// Get the type of the jsonSchema interface.
var jsonSchemaFunc = reflect.TypeOf((*jsonSchema)(nil)).Elem()

// Generator Define the Generator struct.
type Generator struct {
	doc *openapi3.T
}

// NewGenerator Create a new generator.
func NewGenerator() *Generator {
	return &Generator{
		doc: &openapi3.T{
			OpenAPI: "3.0.3",
			Paths:   openapi3.NewPaths(),
			Components: &openapi3.Components{
				Schemas:         openapi3.Schemas{},
				Parameters:      openapi3.ParametersMap{},
				Headers:         openapi3.Headers{},
				RequestBodies:   openapi3.RequestBodies{},
				Responses:       openapi3.ResponseBodies{},
				SecuritySchemes: openapi3.SecuritySchemes{},
				Examples:        openapi3.Examples{},
				Links:           openapi3.Links{},
				Callbacks:       openapi3.Callbacks{},
			},
			Info: &openapi3.Info{},
		},
	}
}

// Generate TestCase for a given type.
func (g *Generator) generateParameters(parameters *openapi3.Parameters, t reflect.Type) {
	if t.Kind() != reflect.Struct {
		return
	}

	// Loop through the fields of the type and handle each field.
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get(OpenAPITag) == "-" || f.Anonymous {
			if f.Anonymous {
				g.generateParameters(parameters, f.Type)
			}
			continue
		}

		in := g.determineParameterLocation(f)
		if in == "" {
			continue
		}

		fieldSchemaRef := g.generateSchemaRef(nil, f.Type, in)
		field := newTagsResolver(f)
		schema := derefSchema(g.doc, fieldSchemaRef)
		field.injectOAITags(schema)

		parameter := g.createParameter(field, schema, in, fieldSchemaRef)
		g.setAdditionalProperties(&parameter, field)
		*parameters = append(*parameters, &openapi3.ParameterRef{Value: &parameter})
	}
}

func (g *Generator) determineParameterLocation(f reflect.StructField) string {
	for _, position := range []string{"path", "query", "header", "cookie"} {
		if name := f.Tag.Get(position); name != "" {
			return position
		}
	}
	return ""
}

func (g *Generator) createParameter(field *tagsResolver, schema *openapi3.Schema, in string, schemaRef *openapi3.SchemaRef) openapi3.Parameter {
	return openapi3.Parameter{
		In:          in,
		Name:        field.name(in),
		Required:    field.required() || in == "path", // path parameters are always required
		Description: schema.Description,
		Deprecated:  schema.Deprecated,
		Schema:      schemaRef,
	}
}

func (g *Generator) setAdditionalProperties(parameter *openapi3.Parameter, field *tagsResolver) {
	if v, ok := field.pairs[propExplode]; ok {
		parameter.Explode = ptr(toBool(v))
	}
	if v, ok := field.pairs[propStyle]; ok {
		parameter.Style = v
	}
}

// GenerateParameters generates OpenAPI TestCase for a given model.
func (g *Generator) GenerateParameters(model reflect.Type) openapi3.Parameters {
	parameters := make(openapi3.Parameters, 0)
	g.generateParameters(&parameters, model)
	if err := parameters.Validate(context.Background()); err != nil {
		panic(err)
	}
	return parameters
}

// GenerateRequestBody generates an OpenAPI request body for a given model using the given operation ID and name tag.
// It takes in the operation ID to use for naming the request body, the name tag to use for naming properties,
// and the model to generate a request body for.
// It returns a *spec.RequestBody that represents the generated request body.
func (g *Generator) GenerateRequestBody(operationID, nameTag string, model reflect.Type) *openapi3.RequestBody {
	schema := g.generateSchemaRef(nil, model, nameTag, operationID+"-body")
	return openapi3.
		NewRequestBody().
		WithRequired(true).
		WithJSONSchemaRef(schema)
}

func (g *Generator) GenerateResponse(code int, model any, mt string, description string) *openapi3.Response {
	desc := http.StatusText(code)
	if description != "" {
		desc = description
	}
	response := openapi3.NewResponse().WithDescription(desc)
	if model == nil {
		return response
	}

	if mt == "application/json" {
		schema := g.generateSchemaRef(nil, reflect.TypeOf(model), "json")
		return response.WithJSONSchemaRef(schema)
	}
	panic("unsupported media type " + mt)
}

var primitiveSchemaFunc = map[reflect.Kind]func() *openapi3.Schema{
	reflect.Int: openapi3.NewIntegerSchema,
	reflect.Uint: func() *openapi3.Schema {
		return openapi3.NewIntegerSchema().WithMin(0)
	},
	reflect.Int8: func() *openapi3.Schema {
		return openapi3.NewIntegerSchema().WithMin(math.MinInt8).WithMax(math.MaxInt8)
	},
	reflect.Uint8: func() *openapi3.Schema {
		return openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint8)
	},
	reflect.Int16: func() *openapi3.Schema {
		return openapi3.NewIntegerSchema().WithMin(math.MinInt16).WithMax(math.MaxInt16)
	},
	reflect.Uint16: func() *openapi3.Schema {
		return openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint16)
	},
	reflect.Int32: func() *openapi3.Schema {
		return openapi3.NewInt32Schema().WithMin(math.MinInt32).WithMax(math.MaxInt32)
	},
	reflect.Uint32: func() *openapi3.Schema {
		return openapi3.NewInt32Schema().WithMin(0).WithMax(math.MaxUint32)
	},
	reflect.Int64: func() *openapi3.Schema {
		return openapi3.NewInt64Schema().WithMin(math.MinInt64).WithMax(math.MaxInt64)
	},
	reflect.Uint64: func() *openapi3.Schema {
		return openapi3.NewInt64Schema().WithMin(0).WithMax(math.MaxUint64)
	},
	reflect.Float32:   openapi3.NewFloat64Schema,
	reflect.Float64:   openapi3.NewFloat64Schema,
	reflect.Bool:      openapi3.NewBoolSchema,
	reflect.String:    openapi3.NewStringSchema,
	reflect.Interface: openapi3.NewSchema,
}

// generateSchemaRef generates an OpenAPI schema for a given type.
// It takes in a slice of parent types to check for circular references,
// the type to generate a schema for, a name tag to use for naming properties,
// and an optional name for the schema.
// It returns a RefOrSpec[Schema] that can be used to reference the generated schema.
func (g *Generator) generateSchemaRef(parents []reflect.Type, t reflect.Type, nameTag string, name ...string) *openapi3.SchemaRef { //nolint
	// Remove any pointer types from the type.
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Check for circular references.
	for _, parent := range parents {
		if parent == t {
			schemaName := g.generateSchemaName(t, name...)
			return openapi3.NewSchemaRef("#/components/schemas/"+schemaName, nil)
		}
	}
	// Check if the type implements the jsonSchema interface.
	if t.Implements(jsonSchemaFunc) {
		js := reflect.New(t).Interface().(jsonSchema).JSONSchema(g.doc)
		return js
	}
	parents = append(parents, t)

	// Handle primitive types.
	if primitiveSchema, ok := primitiveSchemaFunc[t.Kind()]; ok {
		return primitiveSchema().NewRef()
	}

	// Handle well-known types.
	switch t {
	case wnMapStringAny:
		return openapi3.NewObjectSchema().WithAnyAdditionalProperties().NewRef()
	case wnTime:
		return openapi3.NewDateTimeSchema().NewRef()
	case wnIP:
		return openapi3.NewStringSchema().WithFormat("ipv4").NewRef()
	case wnByteSlice:
		return openapi3.NewBytesSchema().NewRef()
	case wnJSON:
		return openapi3.NewStringSchema().WithFormat("json").NewRef()
	}

	// Handle arrays and slices.
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		schema := openapi3.NewArraySchema()
		if t.Kind() == reflect.Array {
			schema.MinItems = uint64(t.Len())
			schema.MaxItems = ptr(schema.MinItems)
		}
		schema.Items = g.generateSchemaRef(parents, t.Elem(), nameTag)
		return schema.NewRef()
	}
	// Handle maps.
	if t.Kind() == reflect.Map {
		itemSchemaRef := g.generateSchemaRef(parents, t.Elem(), nameTag)
		return openapi3.NewObjectSchema().WithAdditionalProperties(itemSchemaRef.Value).NewRef()
	}

	// Handle structs.
	if t.Kind() == reflect.Struct {
		schema := openapi3.NewObjectSchema()

		// Iterate over the struct fields.
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			// Check for the OpenAPI tag "-" to skip the field, skip json tag "-" as well
			if f.Tag.Get(OpenAPITag) == "-" || f.Tag.Get("json") == "-" {
				continue
			}

			// Handle embedded structs.
			if f.Anonymous {
				embedSchema := derefSchema(g.doc, g.generateSchemaRef(parents, f.Type, nameTag))
				for k, v := range embedSchema.Properties {
					schema.Properties[k] = v
				}
				schema.Required = append(schema.Required, embedSchema.Required...)
				continue
			}

			// Generate a schema for the field.
			fieldSchema := g.generateSchemaRef(parents, f.Type, nameTag)
			// Create a field resolver to handle OpenAPI tags.
			field := newTagsResolver(f)
			if fieldSchema.Value != nil {
				field.injectOAITags(derefSchema(g.doc, fieldSchema))
			}

			// Add the field to the schema properties.
			schema.Properties[field.name(nameTag)] = fieldSchema
			if field.required() {
				schema.Required = append(schema.Required, field.name(nameTag))
			}
		}

		// Generate a name for the schema and add it to the OpenAPI components.
		schemaName := g.generateSchemaName(t, name...)
		g.doc.Components.Schemas[schemaName] = schema.NewRef()
		return openapi3.NewSchemaRef("#/components/schemas/"+schemaName, schema)
	}

	panic("unsupported type " + t.String())
}

// generateSchemaName generates a name for an OpenAPI schema based on the given type.
// It takes in the type to generate a name for and an optional name to use instead of generating one.
// It returns a string representing the generated schema name.
func (g *Generator) generateSchemaName(t reflect.Type, name ...string) string {
	// Use the provided name if one was given.
	if len(name) != 0 {
		return name[0]
	}

	// Generate a name based on the type's package path.
	if t.PkgPath() != "" {
		name := t.String()
		if strings.HasPrefix(name, "[]") {
			name = strings.TrimPrefix(name, "[]")
			name += "List"
		}
		if name == "" {
			name = "Object"
		}
		return regexSchemaName.ReplaceAllString(name, "")
	}

	panic("cannot generate a name for an anonymous type")
}

// GenerateSchemaRef generates an OpenAPI schema for a given model using the given name tag.
// It takes in the model to generate a schema for and a name tag to use for naming properties.
// It returns a *spec.Schema that represents the generated schema.
func GenerateSchemaRef(model any, nameTag string, name ...string) *openapi3.SchemaRef {
	// Create a new generator.
	generator := NewGenerator()

	t := reflect.TypeOf(model)
	// Generate a schema for the model.
	ref := generator.generateSchemaRef(nil, t, nameTag, name...)

	// Return the generated schema.
	return ref
}
