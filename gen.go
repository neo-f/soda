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

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
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
	JSONSchema(*v3.Document) *base.SchemaProxy
}

// Get the type of the jsonSchema interface.
var jsonSchemaFunc = reflect.TypeOf((*jsonSchema)(nil)).Elem()

// Define the generator struct.
type generator struct {
	doc *v3.Document
}

// Create a new generator.
func NewGenerator() *generator {
	return &generator{
		doc: &v3.Document{
			Version: "3.1.0",
			Paths: &v3.Paths{
				PathItems: map[string]*v3.PathItem{},
			},
			Components: &v3.Components{
				Schemas:         map[string]*base.SchemaProxy{},
				Responses:       map[string]*v3.Response{},
				Parameters:      map[string]*v3.Parameter{},
				Examples:        map[string]*base.Example{},
				RequestBodies:   map[string]*v3.RequestBody{},
				Headers:         map[string]*v3.Header{},
				SecuritySchemes: map[string]*v3.SecurityScheme{},
				Links:           map[string]*v3.Link{},
				Callbacks:       map[string]*v3.Callback{},
				Extensions:      map[string]any{},
			},
			Info: &base.Info{},
		},
	}
}

func derefSchema(doc *v3.Document, ref *base.SchemaProxy) *base.Schema {
	if ref.IsReference() {
		full := ref.GetReference()
		name := full[strings.LastIndex(full, "/")+1:]
		schema := doc.Components.Schemas[name]
		if schema == nil {
			panic(fmt.Sprintf("schema %s not found", name))
		}
		return derefSchema(doc, schema)
	}
	return ref.Schema()
}

// Generate parameters for a given type.
func (g *generator) generateParameters(parameters *[]*v3.Parameter, t reflect.Type) {
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
		field := newFieldResolver(f)
		schema := derefSchema(g.doc, fieldSchemaRef)
		field.injectOAITags(schema)

		parameter := v3.Parameter{
			In:          in,
			Name:        field.name(in),
			Required:    field.required(),
			Description: schema.Description,
			Deprecated:  unptr(schema.Deprecated),
			Schema:      fieldSchemaRef,
		}

		if v, ok := field.tagPairs[propExplode]; ok {
			parameter.Explode = ptr(toBool(v))
		}
		if v, ok := field.tagPairs[propStyle]; ok {
			parameter.Style = v
		}
		*parameters = append(*parameters, &parameter)
	}
	// Loop through the fields of the type.
	for i := 0; i < t.NumField(); i++ {
		handleField(ptr(t.Field(i)))
	}
}

// GenerateParameters generates OpenAPI parameters for a given model.
func (g *generator) GenerateParameters(model reflect.Type) []*v3.Parameter {
	parameters := make([]*v3.Parameter, 0)
	g.generateParameters(&parameters, model)
	return parameters
}

// GenerateRequestBody generates an OpenAPI request body for a given model using the given operation ID and name tag.
// It takes in the operation ID to use for naming the request body, the name tag to use for naming properties,
// and the model to generate a request body for.
// It returns a *spec.RequestBody that represents the generated request body.
func (g *generator) GenerateRequestBody(operationID, nameTag string, model reflect.Type) *v3.RequestBody {
	schema := g.generateSchema(nil, model, nameTag, operationID+"-body")
	return &v3.RequestBody{
		Required: ptr(true),
		Content: map[string]*v3.MediaType{
			"application/json": {
				Schema: schema,
			},
		},
	}
}

func (g *generator) GenerateResponse(code int, model reflect.Type, mt string, description ...string) *v3.Response {
	desc := http.StatusText(code)
	if len(description) != 0 {
		desc = description[0]
	}

	if model == nil {
		return &v3.Response{Description: desc}
	}

	if mt == "application/json" {
		schema := g.generateSchema(nil, model, "json")
		return &v3.Response{
			Description: desc,
			Content: map[string]*v3.MediaType{
				mt: {Schema: schema},
			},
		}
	}
	panic("unsupported media type " + mt)
}

func newIntSchema(min, max *float64) func() *base.Schema {
	return func() *base.Schema {
		return &base.Schema{
			Type:    []string{typeInteger},
			Minimum: min,
			Maximum: max,
		}
	}
}

var primitiveSchemaFunc = map[reflect.Kind]func() *base.Schema{
	reflect.Int:    newIntSchema(nil, nil),
	reflect.Uint:   newIntSchema(ptr(float64(0)), ptr(float64(math.MaxUint32))),
	reflect.Int8:   newIntSchema(ptr(float64(math.MinInt8)), ptr(float64(math.MaxInt8))),
	reflect.Uint8:  newIntSchema(ptr(float64(0)), ptr(float64(math.MaxUint8))),
	reflect.Int16:  newIntSchema(ptr(float64(math.MinInt16)), ptr(float64(math.MaxInt16))),
	reflect.Uint16: newIntSchema(ptr(float64(0)), ptr(float64(math.MaxUint16))),
	reflect.Int32:  newIntSchema(ptr(float64(math.MinInt32)), ptr(float64(math.MaxInt32))),
	reflect.Uint32: newIntSchema(ptr(float64(0)), ptr(float64(math.MaxUint32))),
	reflect.Int64:  newIntSchema(nil, nil),
	reflect.Uint64: newIntSchema(nil, nil),
	reflect.Float32: func() *base.Schema {
		return &base.Schema{Type: []string{typeNumber}}
	},
	reflect.Float64: func() *base.Schema {
		return &base.Schema{Type: []string{typeNumber}}
	},
	reflect.Bool: func() *base.Schema {
		return &base.Schema{Type: []string{typeBoolean}}
	},
	reflect.String: func() *base.Schema {
		return &base.Schema{Type: []string{typeString}}
	},
	reflect.Interface: func() *base.Schema {
		return &base.Schema{}
	},
}

// generateSchema generates an OpenAPI schema for a given type.
// It takes in a slice of parent types to check for circular references,
// the type to generate a schema for, a name tag to use for naming properties,
// and an optional name for the schema.
// It returns a RefOrSpec[Schema] that can be used to reference the generated schema.
func (g *generator) generateSchema(parents []reflect.Type, t reflect.Type, nameTag string, name ...string) *base.SchemaProxy { //nolint
	// Remove any pointer types from the type.
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Check for circular references.
	for _, parent := range parents {
		if parent == t {
			schemaName := g.generateSchemaName(t, name...)
			return base.CreateSchemaProxyRef("#/components/schemas/" + schemaName)
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
		return base.CreateSchemaProxy(fn())
	}
	// Handle arrays and slices.
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		if t == rawMessageType {
			return base.CreateSchemaProxy(&base.Schema{
				Type:   []string{typeString},
				Format: "json",
			})
		}
		if t.Kind() == reflect.Slice && t.Elem() == byteSliceType.Elem() {
			return base.CreateSchemaProxy(&base.Schema{
				Type:   []string{typeString},
				Format: "byte",
			})
		}
		schema := &base.Schema{
			Type: []string{typeArray},
		}
		if t.Kind() == reflect.Array {
			schema.MinItems = ptr(int64(t.Len()))
			schema.MaxItems = schema.MinItems
		}
		subSchema := g.generateSchema(parents, t.Elem(), nameTag)
		schema.Items = &base.DynamicValue[*base.SchemaProxy, bool]{A: subSchema}
		return base.CreateSchemaProxy(schema)
	}
	// Handle maps.
	if t.Kind() == reflect.Map {
		itemSchemaRef := g.generateSchema(parents, t.Elem(), nameTag)
		return base.CreateSchemaProxy(
			&base.Schema{
				Type:                 []string{typeObject},
				AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{A: itemSchemaRef},
			},
		)
	}

	// Handle basic types.
	switch t {
	case timeType:
		return base.CreateSchemaProxy(&base.Schema{
			Type:   []string{typeString},
			Format: "date-time",
		})
	case ipType:
		return base.CreateSchemaProxy(&base.Schema{
			Type:   []string{typeString},
			Format: "ipv4",
		})
	}

	// Handle structs.
	if t.Kind() == reflect.Struct {
		schema := &base.Schema{
			Type:       []string{typeObject},
			Properties: make(map[string]*base.SchemaProxy, t.NumField()),
		}

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
			field := newFieldResolver(&f)
			field.injectOAITags(derefSchema(g.doc, fieldSchema))

			// Add the field to the schema properties.
			schema.Properties[field.name(nameTag)] = fieldSchema
			if field.required() {
				schema.Required = append(schema.Required, field.name(nameTag))
			}
		}

		// Generate a name for the schema and add it to the OpenAPI components.
		schemaName := g.generateSchemaName(t, name...)
		g.doc.Components.Schemas[schemaName] = base.CreateSchemaProxy(schema)
		return base.CreateSchemaProxyRef("#/components/schemas/" + schemaName)
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
