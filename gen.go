package soda

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// Define some commonly used types.
var (
	timeType       = reflect.TypeOf(time.Time{})       // date-time RFC section 8.3.1
	ipType         = reflect.TypeOf(net.IP{})          // ipv4 and ipv6 RFC section 7.3.4, 7.3.5
	byteSliceType  = reflect.TypeOf([]byte(nil))       // Byte slices will be encoded as base64
	rawMessageType = reflect.TypeOf(json.RawMessage{}) // Except for json.RawMessage
)

// Define an interface for JSON schema generation.
type jsonSchema interface {
	JSONSchema(*openapi3.T) *openapi3.SchemaRef
}

// Get the type of the jsonSchema interface.
var jsonSchemaFunc = reflect.TypeOf((*jsonSchema)(nil)).Elem()

// Define the generator struct.
type generator struct {
	doc *openapi3.T
}

// Create a new generator.
func NewGenerator() *generator {
	return &generator{
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

func derefSchema(doc *openapi3.T, schemaRef *openapi3.SchemaRef) *openapi3.Schema {
	// return schemaRef.Value
	if schemaRef.Value != nil {
		return schemaRef.Value
	}
	if schemaRef.Ref != "" {
		full := schemaRef.Ref
		name := path.Base(full)
		schema, ok := doc.Components.Schemas[name]
		if !ok {
			panic(fmt.Sprintf("schema %s not found", name))
		}
		return derefSchema(doc, schema)
	}
	panic("deref schema failed")
}

// Generate parameters for a given type.
func (g *generator) generateParameters(parameters *[]*openapi3.ParameterRef, t reflect.Type) {
	if t.Kind() != reflect.Struct {
		return
	}
	// Define a function to handle a field.
	handleField := func(f reflect.StructField) {
		if f.Tag.Get(OpenAPITag) == "-" {
			return
		}
		if f.Anonymous {
			g.generateParameters(parameters, f.Type)
			return
		}
		var in string
		for _, position := range []string{"path", "query", "header", "cookie"} {
			if name := f.Tag.Get(position); name != "" {
				in = position
				break
			}
		}
		if in == "" {
			return
		}

		fieldSchemaRef := g.generateSchema(nil, f.Type, in)
		field := newFieldResolver(f)
		schema := derefSchema(g.doc, fieldSchemaRef)
		field.injectOAITags(schema)

		parameter := openapi3.Parameter{
			In:          in,
			Name:        field.name(in),
			Required:    field.required(),
			Description: schema.Description,
			Deprecated:  schema.Deprecated,
			Schema:      fieldSchemaRef,
		}

		if v, ok := field.tagPairs[propExplode]; ok {
			parameter.Explode = ptr(toBool(v))
		}
		if v, ok := field.tagPairs[propStyle]; ok {
			parameter.Style = v
		}
		*parameters = append(*parameters, &openapi3.ParameterRef{Value: &parameter})
	}
	// Loop through the fields of the type.
	for i := 0; i < t.NumField(); i++ {
		handleField(t.Field(i))
	}
}

// GenerateParameters generates OpenAPI parameters for a given model.
func (g *generator) GenerateParameters(model reflect.Type) openapi3.Parameters {
	parameters := make([]*openapi3.ParameterRef, 0)
	g.generateParameters(&parameters, model)
	return parameters
}

// GenerateRequestBody generates an OpenAPI request body for a given model using the given operation ID and name tag.
// It takes in the operation ID to use for naming the request body, the name tag to use for naming properties,
// and the model to generate a request body for.
// It returns a *spec.RequestBody that represents the generated request body.
func (g *generator) GenerateRequestBody(operationID, nameTag string, model reflect.Type) *openapi3.RequestBody {
	schema := g.generateSchema(nil, model, nameTag, operationID+"-body")
	return &openapi3.RequestBody{
		Required: true,
		Content: map[string]*openapi3.MediaType{
			"application/json": {
				Schema: schema,
			},
		},
	}
}

func (g *generator) GenerateResponse(code int, model reflect.Type, mt string, description ...string) *openapi3.Response {
	desc := http.StatusText(code)
	if len(description) != 0 {
		desc = description[0]
	}

	if model == nil {
		return &openapi3.Response{Description: ptr(desc)}
	}

	if mt == "application/json" {
		schema := g.generateSchema(nil, model, "json")
		return &openapi3.Response{
			Description: ptr(desc),
			Content: map[string]*openapi3.MediaType{
				mt: {Schema: schema},
			},
		}
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

// generateSchema generates an OpenAPI schema for a given type.
// It takes in a slice of parent types to check for circular references,
// the type to generate a schema for, a name tag to use for naming properties,
// and an optional name for the schema.
// It returns a RefOrSpec[Schema] that can be used to reference the generated schema.
func (g *generator) generateSchema(parents []reflect.Type, t reflect.Type, nameTag string, name ...string) *openapi3.SchemaRef { //nolint
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
	if fn, ok := primitiveSchemaFunc[t.Kind()]; ok {
		return fn().NewRef()
	}
	// Handle arrays and slices.
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		if t == rawMessageType {
			return openapi3.NewStringSchema().WithFormat("json").NewRef()
		}
		if t.Kind() == reflect.Slice && t.Elem() == byteSliceType.Elem() {
			return openapi3.NewBytesSchema().NewRef()
		}
		schema := openapi3.NewArraySchema()
		if t.Kind() == reflect.Array {
			schema.MinItems = uint64(t.Len())
			schema.MaxItems = ptr(schema.MinItems)
		}
		schema.Items = g.generateSchema(parents, t.Elem(), nameTag)
		return schema.NewRef()
	}
	// Handle maps.
	if t.Kind() == reflect.Map {
		itemSchemaRef := g.generateSchema(parents, t.Elem(), nameTag)
		return openapi3.NewObjectSchema().WithAdditionalProperties(itemSchemaRef.Value).NewRef()
	}

	// Handle basic types.
	switch t {
	case timeType:
		return openapi3.NewDateTimeSchema().NewRef()
	case ipType:
		return openapi3.NewStringSchema().WithFormat("ipv4").NewRef()
	}

	// Handle structs.
	if t.Kind() == reflect.Struct {
		schema := openapi3.NewObjectSchema()

		// Iterate over the struct fields.
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			// Check for the OpenAPI tag "-" to skip the field.
			if f.Tag.Get(OpenAPITag) == "-" {
				break
			}

			// Handle embedded structs.
			if f.Anonymous {
				embedSchema := g.generateSchema(parents, f.Type, nameTag)
				for k, v := range derefSchema(g.doc, embedSchema).Properties {
					schema.Properties[k] = v
				}
				continue
			}

			// Generate a schema for the field.
			fieldSchema := g.generateSchema(parents, f.Type, nameTag)
			// Create a field resolver to handle OpenAPI tags.
			field := newFieldResolver(f)
			field.injectOAITags(derefSchema(g.doc, fieldSchema))

			// Add the field to the schema properties.
			schema.Properties[field.name(nameTag)] = fieldSchema
			if field.required() {
				schema.Required = append(schema.Required, field.name(nameTag))
			}
		}

		// Generate a name for the schema and add it to the OpenAPI components.
		schemaName := g.generateSchemaName(t, name...)
		g.doc.Components.Schemas[schemaName] = schema.NewRef()
		return openapi3.NewSchemaRef("#/components/schemas/"+schemaName, nil)
	}

	panic("unsupported type " + t.String())
}

// generateSchemaName generates a name for an OpenAPI schema based on the given type.
// It takes in the type to generate a name for and an optional name to use instead of generating one.
// It returns a string representing the generated schema name.
func (g *generator) generateSchemaName(t reflect.Type, name ...string) string {
	// Use the provided name if one was given.
	if len(name) != 0 {
		return name[0]
	}

	// Remove any pointer types from the type.
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
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

	// Generate a unique anonymous name.
	for i := 1; ; i++ {
		name := fmt.Sprintf("Anonymous%d", i)
		if _, ok := g.doc.Components.Schemas[name]; ok {
			continue
		}
		return name
	}
}

// GenerateSchema generates an OpenAPI schema for a given model using the given name tag.
// It takes in the model to generate a schema for and a name tag to use for naming properties.
// It returns a *spec.Schema that represents the generated schema.
func GenerateSchema(model any, nameTag string) *openapi3.Schema {
	// Create a new generator.
	generator := NewGenerator()

	// Generate a schema for the model.
	ref := generator.generateSchema(nil, reflect.TypeOf(model), nameTag)

	// Return the generated schema.
	return derefSchema(generator.doc, ref)
}
