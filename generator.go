package soda

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/sv-tools/openapi/spec"
)

var (
	timeType       = reflect.TypeOf(time.Time{})       // date-time RFC section 7.3.1
	ipType         = reflect.TypeOf(net.IP{})          // ipv4 and ipv6 RFC section 7.3.4, 7.3.5
	uriType        = reflect.TypeOf(url.URL{})         // uri RFC section 7.3.6
	byteSliceType  = reflect.TypeOf([]byte(nil))       // Byte slices will be encoded as base64
	rawMessageType = reflect.TypeOf(json.RawMessage{}) // Except for json.RawMessage
)

type jsonSchema interface {
	JSONSchema(*spec.OpenAPI) *spec.Schema
}

var jsonSchemaFunc = reflect.TypeOf((*jsonSchema)(nil)).Elem()

type generator struct {
	spec *spec.OpenAPI
}

func NewGenerator() *generator {
	return &generator{
		spec: &spec.OpenAPI{
			OpenAPI:    "3.1.0",
			Components: spec.NewComponents(),
			Info:       spec.NewInfo(),
		},
	}
}

func (g *generator) generateParameters(parameters *[]*spec.RefOrSpec[spec.Extendable[spec.Parameter]], t reflect.Type) {
	if t.Kind() != reflect.Struct {
		return
	}

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
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		handleField(&f)
	}
}

func (g *generator) GenerateParameters(model reflect.Type) []*spec.RefOrSpec[spec.Extendable[spec.Parameter]] {
	parameters := make([]*spec.RefOrSpec[spec.Extendable[spec.Parameter]], 0)
	g.generateParameters(&parameters, model)
	return parameters
}

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

func (g *generator) GenerateResponse(operationID string, status int, model reflect.Type, nameTag string) *spec.RefOrSpec[spec.Extendable[spec.Response]] {
	schema := g.generateSchema(nil, model, nameTag)

	media := spec.NewMediaType()
	media.Spec.Schema = schema

	response := spec.NewResponseSpec()
	response.Spec.Spec.Description = http.StatusText(status)
	if response.Spec.Spec.Content == nil {
		response.Spec.Spec.Content = make(map[string]*spec.Extendable[spec.MediaType])
	}
	response.Spec.Spec.Content["application/json"] = media
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

func (g *generator) generateSchema(parents []reflect.Type, t reflect.Type, nameTag string, name ...string) *spec.RefOrSpec[spec.Schema] { //nolint
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for _, parent := range parents {
		if parent == t {
			schemaName := g.generateSchemaName(t, name...)
			return spec.NewSchemaRef(spec.NewRef("#/components/schemas/" + schemaName))
		}
	}
	if t.Implements(jsonSchemaFunc) {
		js := reflect.New(t).Interface().(jsonSchema).JSONSchema(g.spec)
		schema := spec.NewSchemaSpec()
		schema.Spec = js
		return schema
	}
	parents = append(parents, t)
	if fn, ok := primitiveSchemaFunc[t.Kind()]; ok {
		return &spec.RefOrSpec[spec.Schema]{Spec: fn()}
	}
	switch t.Kind() {
	case reflect.Struct:
		switch t {
		case timeType:
			schema := spec.NewSchemaSpec()
			schema.Spec.Type = spec.NewSingleOrArray(typeString)
			schema.Spec.Format = "date-time"
			return schema
		case uriType:
			schema := spec.NewSchemaSpec()
			schema.Spec.Type = spec.NewSingleOrArray(typeString)
			schema.Spec.Format = "uri"
			return schema
		case ipType:
			schema := spec.NewSchemaSpec()
			schema.Spec.Type = spec.NewSingleOrArray(typeString)
			schema.Spec.Format = "ipv4"
			return schema
		default:
			schema := spec.NewSchemaSpec()
			schema.Spec.Type = spec.NewSingleOrArray(typeObject)
			schema.Spec.Properties = make(map[string]*spec.RefOrSpec[spec.Schema])

			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				if f.Tag.Get(OpenAPITag) == "-" {
					break
				}
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
				fieldSchemaRef := g.generateSchema(parents, f.Type, nameTag)
				fieldSchema, err := fieldSchemaRef.GetSpec(g.spec.Components)
				if err != nil {
					panic(err)
				}
				field := newFieldResolver(&f)
				field.injectOAITags(fieldSchema)
				schema.Spec.Properties[field.name(nameTag)] = fieldSchemaRef
				if field.required() {
					schema.Spec.Required = append(schema.Spec.Required, field.name(nameTag))
				}
			}

			schemaName := g.generateSchemaName(t, name...)
			g.spec.Components.Spec.WithRefOrSpec(schemaName, schema)
			return spec.NewSchemaRef(spec.NewRef("#/components/schemas/" + schemaName))
		}
	case reflect.Map:
		schema := spec.NewSchemaSpec()
		schema.Spec.Type = spec.NewSingleOrArray(typeObject)
		schema.Spec.AdditionalProperties = spec.NewBoolOrSchema(false, g.generateSchema(parents, t.Elem(), nameTag))
		return schema

	case reflect.Slice, reflect.Array:
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
		if t.Kind() == reflect.Array {
			schema.Spec.MinItems = ptr(t.Len())
			schema.Spec.MaxItems = schema.Spec.MinItems
		}
		subSchema := g.generateSchema(parents, t.Elem(), nameTag)
		schema.Spec.Items = spec.NewBoolOrSchema(false, subSchema)
		return schema

	default:
		panic("unsupported type " + t.String())
	}
}

func (g *generator) generateSchemaName(t reflect.Type, name ...string) string {
	if len(name) != 0 {
		return name[0]
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
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

	for i := 1; ; i++ {
		name := fmt.Sprintf("Anonymous%d", i)
		if _, ok := g.spec.Components.Spec.Schemas[name]; ok {
			continue
		}
		return name
	}
}

func GenerateSchema(model interface{}, nameTag string) *spec.Schema {
	generator := NewGenerator()
	ref := generator.generateSchema(nil, reflect.TypeOf(model), nameTag)
	schema, err := ref.GetSpec(generator.spec.Components)
	if err != nil {
		panic(err)
	}
	return schema
}
