package soda

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/sv-tools/openapi/spec"
)

// Define some commonly used types.
var (
	timeType       = reflect.TypeOf(time.Time{})       // date-time RFC section 7.3.1
	ipType         = reflect.TypeOf(net.IP{})          // ipv4 and ipv6 RFC section 7.3.4, 7.3.5
	byteSliceType  = reflect.TypeOf([]byte(nil))       // Byte slices will be encoded as base64
	rawMessageType = reflect.TypeOf(json.RawMessage{}) // Except for json.RawMessage
)

// Define an interface for JSON schema generation.
type jsonSchema interface {
	JSONSchema(*spec.OpenAPI) *spec.Schema
}

// Get the type of the jsonSchema interface.
var jsonSchemaFunc = reflect.TypeOf((*jsonSchema)(nil)).Elem()

// Define the generator struct.
type generator struct {
	spec *spec.OpenAPI
}

// Create a new generator.
func NewGenerator() *generator {
	return &generator{
		spec: &spec.OpenAPI{
			OpenAPI:    "3.1.0",
			Components: spec.NewComponents(),
			Info:       spec.NewInfo(),
		},
	}
}

// Generate parameters for a given type.
func (g *generator) generateParameters(parameters *[]*spec.RefOrSpec[spec.Extendable[spec.Parameter]], t reflect.Type) {
	if t.Kind() != reflect.Struct {
		return
	}
	// Define a function to handle a field.
	handleField := func(f *reflect.StructField) {
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
		fieldSchema, err := fieldSchemaRef.GetSpec(g.spec.Components)
		if err != nil {
			panic(err)
		}
		field := newFieldResolver(f)
		field.injectOAITags(fieldSchema)

		parameter := spec.NewParameterSpec()
		parameter.Spec.Spec.In = in
		parameter.Spec.Spec.Name = field.name(in)
		parameter.Spec.Spec.Required = field.required()
		parameter.Spec.Spec.Description = fieldSchema.Description
		parameter.Spec.Spec.Deprecated = fieldSchema.Deprecated
		parameter.Spec.Spec.Schema = spec.NewRefOrSpec(nil, fieldSchema)

		if v, ok := field.tagPairs[propExplode]; ok {
			parameter.Spec.Spec.Explode = toBool(v)
		}
		if v, ok := field.tagPairs[propStyle]; ok {
			parameter.Spec.Spec.Style = v
		}
		*parameters = append(*parameters, parameter)
	}
	// Loop through the fields of the type.
	for i := 0; i < t.NumField(); i++ {
		handleField(ptr(t.Field(i)))
	}
}

// GenerateParameters generates OpenAPI parameters for a given model.
func (g *generator) GenerateParameters(model reflect.Type) []*spec.RefOrSpec[spec.Extendable[spec.Parameter]] {
	parameters := make([]*spec.RefOrSpec[spec.Extendable[spec.Parameter]], 0)
	g.generateParameters(&parameters, model)
	return parameters
}

// GenerateRequestBody generates an OpenAPI request body for a given model using the given operation ID and name tag.
// It takes in the operation ID to use for naming the request body, the name tag to use for naming properties,
// and the model to generate a request body for.
// It returns a *spec.RequestBody that represents the generated request body.
func (g *generator) GenerateRequestBody(operationID, nameTag string, model reflect.Type) *spec.RefOrSpec[spec.Extendable[spec.RequestBody]] {
	schema := g.generateSchema(nil, model, nameTag, operationID+"-body")

	media := spec.NewMediaType()
	media.Spec.Schema = schema

	requestBody := spec.NewRequestBodySpec()
	requestBody.Spec.Spec.Required = true
	if requestBody.Spec.Spec.Content == nil {
		requestBody.Spec.Spec.Content = make(map[string]*spec.Extendable[spec.MediaType])
	}
	requestBody.Spec.Spec.Content["application/json"] = media
	return requestBody
}

// GenerateResponse generates an OpenAPI response for a given model using the given operation ID, status code, and name tag.
// It takes in the operation ID to use for naming the response, the status code to use for the response,
// the model to generate a response for, and the name tag to use for naming properties.
// It returns a *spec.Response that represents the generated response.
func (g *generator) GenerateResponse(operationID string, code int, model reflect.Type, nameTag string, description ...string) *spec.RefOrSpec[spec.Extendable[spec.Response]] {
	desc := http.StatusText(code)
	if len(description) != 0 {
		desc = description[0]
	}

	if model == nil {
		response := spec.NewResponseSpec()
		response.Spec.Spec.Description = desc
		return response
	}

	schema := g.generateSchema(nil, model, nameTag)

	media := spec.NewMediaType()
	media.Spec.Schema = schema

	response := spec.NewResponseSpec()
	response.Spec.Spec.Description = desc
	if response.Spec.Spec.Content == nil {
		response.Spec.Spec.Content = make(map[string]*spec.Extendable[spec.MediaType])
	}
	if nameTag == "json" {
		response.Spec.Spec.Content["application/json"] = media
	} else {
		response.Spec.Spec.Content[nameTag] = media
	}
	return response
}

func newIntSchema(min, max *int) func() *spec.Schema {
	return func() *spec.Schema {
		schema := &spec.Schema{}
		schema.Type = []string{typeInteger}
		schema.Minimum = min
		schema.Maximum = max
		return schema
	}
}

var primitiveSchemaFunc = map[reflect.Kind]func() *spec.Schema{
	reflect.Int:    newIntSchema(nil, nil),
	reflect.Uint:   newIntSchema(ptr(0), ptr(math.MaxUint32)),
	reflect.Int8:   newIntSchema(ptr(math.MinInt8), ptr(math.MaxInt8)),
	reflect.Uint8:  newIntSchema(ptr(0), ptr(math.MaxUint8)),
	reflect.Int16:  newIntSchema(ptr(math.MinInt16), ptr(math.MaxInt16)),
	reflect.Uint16: newIntSchema(ptr(0), ptr(math.MaxUint16)),
	reflect.Int32:  newIntSchema(ptr(math.MinInt32), ptr(math.MaxInt32)),
	reflect.Uint32: newIntSchema(ptr(0), ptr(math.MaxUint32)),
	reflect.Int64:  newIntSchema(nil, nil),
	reflect.Uint64: newIntSchema(nil, nil),
	reflect.Float32: func() *spec.Schema {
		schema := &spec.Schema{}
		schema.Type = []string{typeNumber}
		return schema
	},
	reflect.Float64: func() *spec.Schema {
		schema := &spec.Schema{}
		schema.Type = []string{typeNumber}
		return schema
	},
	reflect.Bool: func() *spec.Schema {
		schema := &spec.Schema{}
		schema.Type = []string{typeBoolean}
		return schema
	},
	reflect.String: func() *spec.Schema {
		schema := &spec.Schema{}
		schema.Type = []string{typeString}
		return schema
	},
	reflect.Interface: func() *spec.Schema {
		schema := &spec.Schema{}
		return schema
	},
}

// generateSchema generates an OpenAPI schema for a given type.
// It takes in a slice of parent types to check for circular references,
// the type to generate a schema for, a name tag to use for naming properties,
// and an optional name for the schema.
// It returns a RefOrSpec[Schema] that can be used to reference the generated schema.
func (g *generator) generateSchema(parents []reflect.Type, t reflect.Type, nameTag string, name ...string) *spec.RefOrSpec[spec.Schema] { //nolint
	// Remove any pointer types from the type.
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Check for circular references.
	for _, parent := range parents {
		if parent == t {
			schemaName := g.generateSchemaName(t, name...)
			return spec.NewSchemaRef(spec.NewRef("#/components/schemas/" + schemaName))
		}
	}
	// Check if the type implements the jsonSchema interface.
	if t.Implements(jsonSchemaFunc) {
		js := reflect.New(t).Interface().(jsonSchema).JSONSchema(g.spec)
		schema := spec.NewSchemaSpec()
		schema.Spec = js
		return schema
	}
	parents = append(parents, t)

	// Handle primitive types.
	if fn, ok := primitiveSchemaFunc[t.Kind()]; ok {
		return &spec.RefOrSpec[spec.Schema]{Spec: fn()}
	}
	// Handle arrays and slices.
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		if t == rawMessageType {
			schema := spec.NewSchemaSpec()
			schema.Spec.Type = spec.NewSingleOrArray(typeString)
			schema.Spec.Format = "byte"
			return schema
		}
		if t.Kind() == reflect.Slice && t.Elem() == byteSliceType.Elem() {
			schema := spec.NewSchemaSpec()
			schema.Spec.Type = spec.NewSingleOrArray(typeString)
			schema.Spec.Format = "byte"
			return schema
		}
		schema := spec.NewSchemaSpec()
		schema.Spec.Type = spec.NewSingleOrArray(typeArray)
		if t.Kind() == reflect.Array {
			schema.Spec.MinItems = ptr(t.Len())
			schema.Spec.MaxItems = schema.Spec.MinItems
		}
		subSchema := g.generateSchema(parents, t.Elem(), nameTag)
		schema.Spec.Items = spec.NewBoolOrSchema(false, subSchema)
		return schema
	}
	// Handle maps.
	if t.Kind() == reflect.Map {
		itemSchemaRef := g.generateSchema(parents, t.Elem(), nameTag)
		schema := spec.NewSchemaSpec()
		schema.Spec.Type = spec.NewSingleOrArray(typeObject)
		schema.Spec.AdditionalProperties = spec.NewBoolOrSchema(false, itemSchemaRef)
		return schema
	}

	// Handle basic types.
	switch t {
	case timeType:
		schema := spec.NewSchemaSpec()
		schema.Spec.Type = spec.NewSingleOrArray(typeString)
		schema.Spec.Format = "date-time"
		return schema
	case ipType:
		schema := spec.NewSchemaSpec()
		schema.Spec.Type = spec.NewSingleOrArray(typeString)
		schema.Spec.Format = "ipv4"
		return schema
	}

	// Handle structs.
	if t.Kind() == reflect.Struct {
		schema := spec.NewSchemaSpec()
		schema.Spec.Type = spec.NewSingleOrArray(typeObject)
		schema.Spec.Properties = make(map[string]*spec.RefOrSpec[spec.Schema])

		// Iterate over the struct fields.
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			// Check for the OpenAPI tag "-" to skip the field.
			if f.Tag.Get(OpenAPITag) == "-" {
				break
			}

			// Handle embedded structs.
			if f.Anonymous {
				embedSchemaRef := g.generateSchema(parents, f.Type, nameTag)
				embedSchema, err := embedSchemaRef.GetSpec(g.spec.Components)
				if err != nil {
					panic(err)
				}
				for k, v := range embedSchema.Properties {
					schema.Spec.Properties[k] = v
				}
				continue
			}

			// Generate a schema for the field.
			fieldSchemaRef := g.generateSchema(parents, f.Type, nameTag)
			fieldSchema, err := fieldSchemaRef.GetSpec(g.spec.Components)
			if err != nil {
				panic(err)
			}

			// Create a field resolver to handle OpenAPI tags.
			field := newFieldResolver(&f)
			field.injectOAITags(fieldSchema)

			// Add the field to the schema properties.
			schema.Spec.Properties[field.name(nameTag)] = fieldSchemaRef
			if field.required() {
				schema.Spec.Required = append(schema.Spec.Required, field.name(nameTag))
			}
		}

		// Generate a name for the schema and add it to the OpenAPI components.
		schemaName := g.generateSchemaName(t, name...)
		g.spec.Components.Spec.WithRefOrSpec(schemaName, schema)
		return spec.NewSchemaRef(spec.NewRef("#/components/schemas/" + schemaName))
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
		if _, ok := g.spec.Components.Spec.Schemas[name]; ok {
			continue
		}
		return name
	}
}

// GenerateSchema generates an OpenAPI schema for a given model using the given name tag.
// It takes in the model to generate a schema for and a name tag to use for naming properties.
// It returns a *spec.Schema that represents the generated schema.
func GenerateSchema(model interface{}, nameTag string) *spec.Schema {
	// Create a new generator.
	generator := NewGenerator()

	// Generate a schema for the model.
	ref := generator.generateSchema(nil, reflect.TypeOf(model), nameTag)
	schema, err := ref.GetSpec(generator.spec.Components)
	if err != nil {
		panic(err)
	}

	// Return the generated schema.
	return schema
}
