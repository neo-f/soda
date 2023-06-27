package soda

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	timeType       = reflect.TypeOf(time.Time{})       // date-time RFC section 7.3.1
	ipType         = reflect.TypeOf(net.IP{})          // ipv4 and ipv6 RFC section 7.3.4, 7.3.5
	uriType        = reflect.TypeOf(url.URL{})         // uri RFC section 7.3.6
	byteSliceType  = reflect.TypeOf([]byte(nil))       // Byte slices will be encoded as base64
	rawMessageType = reflect.TypeOf(json.RawMessage{}) // Except for json.RawMessage
)

type jsonSchema interface {
	JSONSchema(*openapi3.T) *openapi3.Schema
}

var jsonSchemaFunc = reflect.TypeOf((*jsonSchema)(nil)).Elem()

type generator struct {
	spec *openapi3.T
}

func NewGenerator() *generator {
	return &generator{
		spec: &openapi3.T{
			OpenAPI: "3.0.3",
			Components: &openapi3.Components{
				Schemas:       make(openapi3.Schemas),
				Responses:     make(openapi3.Responses),
				RequestBodies: make(openapi3.RequestBodies),
			},
			Info: &openapi3.Info{},
		},
	}
}

func (g *generator) generateParameters(parameters *openapi3.Parameters, t reflect.Type) {
	if t.Kind() != reflect.Struct {
		return
	}

	handleField := func(f *reflect.StructField) {
		var in string
		for _, position := range []string{openapi3.ParameterInPath, openapi3.ParameterInQuery, openapi3.ParameterInHeader, openapi3.ParameterInCookie} {
			if name := f.Tag.Get(position); name != "" {
				in = position
				break
			}
		}
		field := newFieldResolver(f)
		if field.shouldEmbed() {
			g.generateParameters(parameters, f.Type)
			return
		}
		if in == "" || field.ignored {
			return
		}
		fieldSchema, _ := g.genSchema(nil, f.Type, in)
		field.injectOAITags(fieldSchema.Value)
		param := &openapi3.Parameter{
			In:          in,
			Name:        field.name(in),
			Required:    field.required(),
			Description: fieldSchema.Value.Description,
			Example:     fieldSchema.Value.Example,
			Deprecated:  fieldSchema.Value.Deprecated,
			Schema:      fieldSchema.Value.NewRef(),
		}

		if v, ok := field.tagPairs[propExplode]; ok {
			param.Explode = openapi3.BoolPtr(toBool(v))
		}
		if v, ok := field.tagPairs[propStyle]; ok {
			param.Style = v
		}
		if err := param.Validate(context.TODO()); err != nil {
			panic(err)
		}
		*parameters = append(*parameters, &openapi3.ParameterRef{Value: param})
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		handleField(&f)
	}
}

func (g *generator) GenerateResponse(operationID string, status int, model reflect.Type, typ string) *openapi3.ResponseRef {
	ref := g.getSchemaRef(model, typ, "")
	response := openapi3.NewResponse().WithJSONSchemaRef(ref).WithDescription(http.StatusText(status))
	responseName := fmt.Sprintf("%s%s", operationIDToCamelCase(operationID), strings.ReplaceAll(http.StatusText(status), " ", ""))

	if _, ok := g.spec.Components.Responses[responseName]; ok {
		i := 1
		for {
			newName := fmt.Sprintf("%s-%d", responseName, i)
			if _, ok := g.spec.Components.Responses[newName]; !ok {
				responseName = newName
				break
			}
		}
	}

	g.spec.Components.Responses[responseName] = &openapi3.ResponseRef{Value: response}
	return &openapi3.ResponseRef{Ref: fmt.Sprintf("#/components/responses/%s", responseName), Value: response}
}

func (g *generator) GenerateParameters(model reflect.Type) openapi3.Parameters {
	parameters := openapi3.NewParameters()
	g.generateParameters(&parameters, model)
	return parameters
}

func (g *generator) GenerateRequestBody(operationID, nameTag string, model reflect.Type) *openapi3.RequestBodyRef {
	schema := g.getSchemaRef(model, nameTag, operationID+"Body")
	requestBody := openapi3.NewRequestBody().WithJSONSchemaRef(schema).WithRequired(true)
	requestName := operationIDToCamelCase(operationID)

	if _, ok := g.spec.Components.RequestBodies[requestName]; ok {
		i := 1
		for {
			newName := fmt.Sprintf("%s-%d", requestName, i)
			if _, ok := g.spec.Components.RequestBodies[newName]; !ok {
				requestName = newName
				break
			}
		}
	}

	g.spec.Components.RequestBodies[requestName] = &openapi3.RequestBodyRef{Value: requestBody}
	return &openapi3.RequestBodyRef{Ref: fmt.Sprintf("#/components/requestBodies/%s", requestName), Value: requestBody}
}

func (g *generator) getSchemaRef(rf reflect.Type, nameTag, schemaName string) *openapi3.SchemaRef {
	ref, _ := g.genSchema(nil, rf, nameTag)
	if schemaName == "" {
		schemaName = g.genSchemaName(rf)
	}
	g.spec.Components.Schemas[schemaName] = ref
	return openapi3.NewSchemaRef("#/components/schemas/"+schemaName, ref.Value)
}

func (g *generator) generateCycleSchemaRef(t reflect.Type, schema *openapi3.Schema) *openapi3.SchemaRef {
	switch t.Kind() {
	case reflect.Ptr:
		return g.generateCycleSchemaRef(t.Elem(), schema)
	case reflect.Slice:
		ref := g.generateCycleSchemaRef(t.Elem(), schema)
		g.spec.Components.Schemas[g.genSchemaName(t.Elem())] = openapi3.NewSchemaRef("", ref.Value)
		sliceSchema := openapi3.NewArraySchema()
		sliceSchema.Items = ref
		return openapi3.NewSchemaRef("", sliceSchema)
	case reflect.Map:
		ref := g.generateCycleSchemaRef(t.Elem(), schema)
		g.spec.Components.Schemas[g.genSchemaName(t.Elem())] = openapi3.NewSchemaRef("", ref.Value)
		mapSchema := openapi3.NewObjectSchema()
		mapSchema.AdditionalProperties.Schema = ref
		return openapi3.NewSchemaRef("", mapSchema)
	}

	return openapi3.NewSchemaRef("#/components/schemas/"+g.genSchemaName(t), schema)
}

func (g *generator) genSchema(parents []reflect.Type, t reflect.Type, nameTag string) (*openapi3.SchemaRef, bool) { //nolint
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for _, parent := range parents {
		if parent == t {
			return nil, true
		}
	}
	if t.Implements(jsonSchemaFunc) {
		return reflect.New(t).Interface().(jsonSchema).JSONSchema(g.spec).NewRef(), false
	}

	parents = append(parents, t)

	switch t.Kind() {
	case reflect.Struct:
		switch t {
		case timeType:
			return openapi3.NewDateTimeSchema().NewRef(), false
		case uriType:
			return openapi3.NewStringSchema().WithFormat("uri").NewRef(), false
		case ipType:
			return openapi3.NewStringSchema().WithFormat("ipv4").NewRef(), false
		default:
			schema := openapi3.NewObjectSchema()
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				field := newFieldResolver(&f)
				var fieldSchemaRef *openapi3.SchemaRef
				if field.ignored {
					break
				}
				if field.shouldEmbed() {
					if ref, cycle := g.genSchema(parents, f.Type, nameTag); cycle {
						fieldSchemaRef = g.generateCycleSchemaRef(f.Type, schema)
					} else {
						fieldSchemaRef = ref
					}
					for k, v := range fieldSchemaRef.Value.Properties {
						schema.Properties[k] = v
					}
					continue
				}
				if ref, cycle := g.genSchema(parents, f.Type, nameTag); cycle {
					fieldSchemaRef = g.generateCycleSchemaRef(f.Type, schema)
				} else {
					fieldSchemaRef = ref
				}

				field.injectOAITags(fieldSchemaRef.Value)
				if fieldSchemaRef.Value.Nullable {
					nullSchema := openapi3.NewSchema()
					nullSchema.Type = "null"
					fieldSchemaRef.Value = &openapi3.Schema{
						OneOf: openapi3.SchemaRefs{
							fieldSchemaRef,
							nullSchema.NewRef(),
						},
					}
				}
				schema.Properties[field.name(nameTag)] = fieldSchemaRef
				if field.required() {
					schema.Required = append(schema.Required, field.name(nameTag))
				}
			}

			return schema.NewRef(), false
		}
	case reflect.Map:
		schema := openapi3.NewObjectSchema()
		additionalProperties, cycle := g.genSchema(parents, t.Elem(), nameTag)
		if cycle {
			additionalProperties = g.generateCycleSchemaRef(t.Elem(), schema)
		}
		schema.AdditionalProperties.Schema = additionalProperties
		return schema.NewRef(), false

	case reflect.Slice, reflect.Array:
		if t == rawMessageType {
			return openapi3.NewBytesSchema().NewRef(), false
		}
		if t.Kind() == reflect.Slice && t.Elem() == byteSliceType.Elem() {
			return openapi3.NewBytesSchema().NewRef(), false
		}
		schema := openapi3.NewArraySchema()
		if t.Kind() == reflect.Array {
			schema.MinItems = uint64(t.Len())
			schema.MaxItems = &schema.MinItems
		}
		if ref, cycle := g.genSchema(parents, t.Elem(), nameTag); cycle {
			schema.Items = g.generateCycleSchemaRef(t.Elem(), schema)
		} else {
			schema.Items = ref
		}
		return schema.NewRef(), false

	case reflect.Interface:
		return openapi3.NewSchema().WithAnyAdditionalProperties().NewRef(), false
	case reflect.Int:
		return openapi3.NewIntegerSchema().NewRef(), false
	case reflect.Uint:
		return openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint).NewRef(), false
	case reflect.Int8:
		return openapi3.NewIntegerSchema().WithMin(math.MinInt8).WithMax(math.MaxInt8).NewRef(), false
	case reflect.Uint8:
		return openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint8).NewRef(), false

	case reflect.Int16:
		return openapi3.NewIntegerSchema().WithMin(math.MinInt16).WithMax(math.MaxInt16).NewRef(), false
	case reflect.Uint16:
		return openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint16).NewRef(), false

	case reflect.Int32:
		return openapi3.NewInt32Schema().WithMin(math.MinInt32).WithMax(math.MaxInt32).NewRef(), false
	case reflect.Uint32:
		return openapi3.NewInt32Schema().WithMin(0).WithMax(math.MaxUint32).NewRef(), false

	case reflect.Int64:
		return openapi3.NewInt64Schema().NewRef(), false
	case reflect.Uint64:
		return openapi3.NewInt64Schema().WithMin(0).NewRef(), false

	case reflect.Float32:
		return openapi3.NewFloat64Schema().WithFormat("float").NewRef(), false
	case reflect.Float64:
		return openapi3.NewFloat64Schema().WithFormat("double").NewRef(), false

	case reflect.Bool:
		return openapi3.NewBoolSchema().NewRef(), false
	case reflect.String:
		return openapi3.NewStringSchema().NewRef(), false
	default:
		panic("unsupported type " + t.String())
	}
}

func (g *generator) GenerateSchema(model interface{}, tag string) *openapi3.Schema {
	t := reflect.TypeOf(model)
	ref, _ := g.genSchema(nil, t, tag)
	return ref.Value
}

func (g *generator) genSchemaName(rf reflect.Type) string {
	for rf.Kind() == reflect.Ptr {
		rf = rf.Elem()
	}
	name := rf.String()
	if strings.HasPrefix(name, "[]") {
		name = strings.TrimPrefix(name, "[]")
		name += "List"
	}
	return regexSchemaName.ReplaceAllString(name, "")
}
