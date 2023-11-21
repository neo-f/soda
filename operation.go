package soda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type (
	// HookBeforeBind is a function type that is called before binding the request.
	// It returns a boolean indicating whether to continue the process.
	HookBeforeBind func(w http.ResponseWriter, r *http.Request) (doNext bool)

	// HookAfterBind is a function type that is called after binding the request.
	// It returns a boolean indicating whether to continue the process.
	HookAfterBind func(w http.ResponseWriter, r *http.Request, input any) (doNext bool)
)

// OperationBuilder is a struct that helps in building an operation.
type OperationBuilder struct {
	route     *route
	operation *v3.Operation

	method  string
	pattern string

	input              reflect.Type
	inputBody          reflect.Type
	inputBodyField     string
	inputBodyMediaType string

	handler http.Handler

	ignoreAPIDoc bool

	// hooks
	hooksBeforeBind []HookBeforeBind
	hooksAfterBind  []HookAfterBind
}

// SetOperationID sets the operation ID of the operation.
func (op *OperationBuilder) SetOperationID(id string) *OperationBuilder {
	op.operation.OperationId = id
	return op
}

// SetSummary sets the summary of the operation.
func (op *OperationBuilder) SetSummary(summary string) *OperationBuilder {
	op.operation.Summary = summary
	return op
}

// SetDescription sets the description of the operation.
func (op *OperationBuilder) SetDescription(desc string) *OperationBuilder {
	op.operation.Description = desc
	return op
}

// AddTags adds tags to the operation.
func (op *OperationBuilder) AddTags(tags ...string) *OperationBuilder {
	op.operation.Tags = appendUniq(op.operation.Tags, tags...)

	ts := make([]*base.Tag, 0, len(tags))
	for _, tag := range tags {
		ts = append(ts, &base.Tag{Name: tag})
	}
	op.route.gen.doc.Tags = appendUniq(op.route.gen.doc.Tags, ts...)
	return op
}

// SetDeprecated marks the operation as deprecated or not.
func (op *OperationBuilder) SetDeprecated(deprecated bool) *OperationBuilder {
	op.operation.Deprecated = ptr(deprecated)
	return op
}

// SetInput sets the input type for the operation.
func (op *OperationBuilder) SetInput(input any) *OperationBuilder {
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

// AddSecurity adds a security scheme to the operation.
func (op *OperationBuilder) AddSecurity(scheme *v3.SecurityScheme, securityName string) *OperationBuilder {
	if op.route.gen.doc.Components.SecuritySchemes == nil {
		op.route.gen.doc.Components.SecuritySchemes = make(map[string]*v3.SecurityScheme)
	}

	op.route.gen.doc.Components.SecuritySchemes[securityName] = scheme

	opSecurity := &base.SecurityRequirement{
		Requirements: map[string][]string{
			securityName: nil,
		},
	}

	op.operation.Security = appendUniq(op.operation.Security, opSecurity)
	return op
}

// AddJSONResponse adds a JSON response to the operation.
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

// SetIgnoreAPIDoc sets whether to ignore the operation when generating the API doc.
func (op *OperationBuilder) SetIgnoreAPIDoc(ignore bool) *OperationBuilder {
	op.ignoreAPIDoc = ignore
	return op
}

// OnBeforeBind adds a hook that is called before binding the request.
func (op *OperationBuilder) OnBeforeBind(hook HookBeforeBind) *OperationBuilder {
	op.hooksBeforeBind = append(op.hooksBeforeBind, hook)
	return op
}

// OnAfterBind adds a hook that is called after binding the request.
func (op *OperationBuilder) OnAfterBind(hook HookAfterBind) *OperationBuilder {
	op.hooksAfterBind = append(op.hooksAfterBind, hook)
	return op
}

// OK finalizes the operation building process.
func (op *OperationBuilder) OK() {
	if !op.ignoreAPIDoc {
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
		// clean the chi pattern, remove the regex etc from the parameters..
		path := cleanPath(op.pattern)
		if op.route.gen.doc.Paths.PathItems[path] == nil {
			op.route.gen.doc.Paths.PathItems[path] = &v3.PathItem{}
		}
		pathItem := op.route.gen.doc.Paths.PathItems[path]

		switch op.method {
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
	}

	// Add handler
	op.route.router.With(op.bindInput).Method(op.method, op.pattern, op.handler)
}

// bindInput binds the request body to the input struct.
func (op *OperationBuilder) bindInput(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if op.input == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Execute Hooks: BeforeBind
		if !op.executeHooksBeforeBind(w, r) {
			return
		}

		// Bind input
		input := reflect.New(op.input).Interface()

		// parse the request parameters
		if err := op.parseRequestParameters(r, input); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// parse the request body
		if err := op.parseRequestBody(r, input); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Execute Hooks: AfterBind
		if !op.executeHooksAfterBind(w, r, input) {
			return
		}

		ctx := context.WithValue(r.Context(), KeyInput, input)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (op *OperationBuilder) executeHooksBeforeBind(w http.ResponseWriter, r *http.Request) bool {
	for _, hook := range op.hooksBeforeBind {
		if !hook(w, r) {
			return false
		}
	}
	return true
}

func (op *OperationBuilder) parseRequestParameters(r *http.Request, input any) error {
	for _, parser := range parameterParsers {
		if err := parser(r, input); err != nil {
			return err
		}
	}
	return nil
}

func (op *OperationBuilder) parseRequestBody(r *http.Request, input any) error {
	if op.inputBodyField != "" && op.inputBodyMediaType == "json" {
		body := reflect.New(op.inputBody).Interface()
		if err := json.NewDecoder(r.Body).Decode(body); err != nil {
			return err
		}
		reflect.ValueOf(input).Elem().FieldByName(op.inputBodyField).Set(reflect.ValueOf(body).Elem())
	}
	return nil
}

func (op *OperationBuilder) executeHooksAfterBind(w http.ResponseWriter, r *http.Request, input any) bool {
	for _, hook := range op.hooksAfterBind {
		if !hook(w, r, input) {
			return false
		}
	}
	return true
}
