package soda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type (
	HookBeforeBind func(w http.ResponseWriter, r *http.Request) (doNext bool)
	HookAfterBind  func(w http.ResponseWriter, r *http.Request, input interface{}) (doNext bool)
)

// OperationBuilder is a builder for a single operation.
type OperationBuilder struct {
	route     *Route
	operation *v3.Operation

	method  string
	pattern string

	input              reflect.Type
	inputBody          reflect.Type
	inputBodyField     string
	inputBodyMediaType string

	middlewares []func(http.Handler) http.Handler
	handler     http.Handler

	// hooks
	hooksBeforeBind []HookBeforeBind
	hooksAfterBind  []HookAfterBind
}

// SetSummary sets the operation-id.
func (op *OperationBuilder) SetOperationID(id string) *OperationBuilder {
	op.operation.OperationId = id
	return op
}

// SetSummary sets the operation summary.
func (op *OperationBuilder) SetSummary(summary string) *OperationBuilder {
	op.operation.Summary = summary
	return op
}

// SetDescription sets the operation description.
func (op *OperationBuilder) SetDescription(desc string) *OperationBuilder {
	op.operation.Description = desc
	return op
}

// AddTags add tags to the operation.
func (op *OperationBuilder) AddTags(tags ...string) *OperationBuilder {
	appendUniqBy(sameVal, op.operation.Tags, tags...)

	ts := make([]*base.Tag, 0, len(tags))
	for _, tag := range tags {
		ts = append(ts, &base.Tag{Name: tag})
	}
	appendUniqBy(sameTag, op.route.gen.doc.Tags, ts...)
	return op
}

// SetDeprecated marks the operation as deprecated.
func (op *OperationBuilder) SetDeprecated(deprecated bool) *OperationBuilder {
	op.operation.Deprecated = ptr(deprecated)
	return op
}

// SetInput sets the input for this operation.
// The input must be a pointer to a struct.
// If the struct has a field with the `body:"<media type>"` tag, that field is used for the request body.
// Otherwise, the struct is used for parameters.
func (op *OperationBuilder) SetInput(input interface{}) *OperationBuilder {
	inputType := reflect.TypeOf(input)
	// the input type should be a struct or pointer to a struct
	for inputType.Kind() == reflect.Ptr {
		inputType = inputType.Elem()
	}
	if inputType.Kind() != reflect.Struct {
		panic("input must be a pointer to a struct")
	}

	op.input = inputType

	for i := 0; i < inputType.NumField(); i++ {
		if body := inputType.Field(i); body.Tag.Get("body") != "" {
			op.inputBody = body.Type
			op.inputBodyMediaType = body.Tag.Get("body")
			op.inputBodyField = body.Name
			break
		}
	}

	op.operation.Parameters = op.route.gen.GenerateParameters(inputType)
	if op.inputBodyField != "" {
		op.operation.RequestBody = op.route.gen.GenerateRequestBody(op.operation.OperationId, op.inputBodyMediaType, op.inputBody)
	}
	return op
}

// AddSecurity adds Security to this operation.
func (op *OperationBuilder) AddSecurity(scheme *v3.SecurityScheme, securityName string) *OperationBuilder {
	// add the security scheme to the spec if it doesn't already exist
	if op.route.gen.doc.Components.SecuritySchemes == nil {
		op.route.gen.doc.Components.SecuritySchemes = make(map[string]*v3.SecurityScheme)
	}

	op.route.gen.doc.Components.SecuritySchemes[securityName] = scheme

	opSecurity := &base.SecurityRequirement{
		Requirements: map[string][]string{
			securityName: nil,
		},
	}

	// add the security scheme to the operation
	appendUniqBy(sameSecurityRequirements, op.operation.Security, opSecurity)
	return op
}

// AddJSONResponse adds a JSON response to the operation's responses.
// If model is not nil, a JSON response is generated for the model type.
// If model is nil, a JSON response is generated with no schema.
func (op *OperationBuilder) AddJSONResponse(code int, model any, description ...string) *OperationBuilder {
	if op.operation.Responses == nil {
		op.operation.Responses = &v3.Responses{
			Codes: map[string]*v3.Response{},
		}
	}
	ref := op.route.gen.GenerateResponse(op.operation.OperationId, code, reflect.TypeOf(model), "application/json", description...)
	op.operation.Responses.Codes[strconv.Itoa(code)] = ref
	return op
}

func (op *OperationBuilder) OnAfterBind(hook HookAfterBind) *OperationBuilder {
	op.hooksAfterBind = append(op.hooksAfterBind, hook)
	return op
}

func (op *OperationBuilder) OnBeforeBind(hook HookBeforeBind) *OperationBuilder {
	op.hooksBeforeBind = append(op.hooksBeforeBind, hook)
	return op
}

func (op *OperationBuilder) OK() {
	// Add default response if not exists
	if op.operation.Responses == nil {
		op.operation.Responses = &v3.Responses{
			Default: &v3.Response{},
		}
	}

	// Add operation to the spec
	if op.route.gen.doc.Paths == nil {
		op.route.gen.doc.Paths = &v3.Paths{
			PathItems: map[string]*v3.PathItem{},
		}
	}
	// TODO: clean the chi pattern, remove the regex etc from the parameters..
	path := op.pattern
	if op.route.gen.doc.Paths.PathItems[path] == nil {
		op.route.gen.doc.Paths.PathItems[path] = &v3.PathItem{}
	}
	pathItem := op.route.gen.doc.Paths.PathItems[path]

	switch strings.ToUpper(op.method) {
	case http.MethodGet:
		pathItem.Get = op.operation
	case http.MethodHead:
		pathItem.Head = op.operation
	case http.MethodPost:
		pathItem.Post = op.operation
	case http.MethodPut:
		pathItem.Put = op.operation
	case http.MethodPatch:
		pathItem.Patch = op.operation
	case http.MethodDelete:
		pathItem.Delete = op.operation
	case http.MethodOptions:
		pathItem.Options = op.operation
	case http.MethodTrace:
		pathItem.Trace = op.operation
	default:
		panic(fmt.Errorf("unsupported HTTP method %q", op.method))
	}

	// Add handler
	op.middlewares = append(op.middlewares, op.bindInput)
	op.route.router.With(op.middlewares...).Method(op.method, op.pattern, op.handler)
}

// bindInput binds the request body to the input struct.
func (op *OperationBuilder) bindInput(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if op.input == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Execute Hooks: BeforeBind
		for _, hook := range op.hooksBeforeBind {
			if !hook(w, r) {
				return
			}
		}

		// Bind input
		input := reflect.New(op.input).Interface()

		// parse the request parameters
		for _, parser := range parameterParsers {
			if err := parser(r, input); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}

		// parse the request body
		// TODO: support other media types
		if op.inputBodyField != "" && op.inputBodyMediaType == "json" {
			body := reflect.New(op.inputBody).Interface()
			if err := json.NewDecoder(r.Body).Decode(body); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			reflect.ValueOf(input).Elem().FieldByName(op.inputBodyField).Set(reflect.ValueOf(body).Elem())
		}

		// Execute Hooks: AfterBind
		for _, hook := range op.hooksAfterBind {
			if !hook(w, r, input) {
				return
			}
		}

		ctx := context.WithValue(r.Context(), KeyInput, input)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
