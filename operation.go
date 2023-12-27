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
	"github.com/pb33f/libopenapi/orderedmap"
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

	input               reflect.Type
	inputBody           reflect.Type
	inputBodyField      string
	inputBodyResolveTag string

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
	op.operation.Tags = append(op.operation.Tags, tags...)
	for _, tag := range tags {
		op.route.gen.doc.Tags = append(op.route.gen.doc.Tags, &base.Tag{
			Name: tag,
		})
	}
	// remove duplicates
	op.operation.Tags = uniqBy(op.operation.Tags, func(item string) string { return item })
	op.route.gen.doc.Tags = uniqBy(op.route.gen.doc.Tags, func(item *base.Tag) string { return item.Name })
	return op
}

// SetDeprecated marks the operation as deprecated or not.
func (op *OperationBuilder) SetDeprecated(deprecated *bool) *OperationBuilder {
	op.operation.Deprecated = deprecated
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
			op.inputBodyResolveTag = body.Tag.Get("body")
			op.inputBodyField = body.Name
			break
		}
	}

	op.operation.Parameters = op.route.gen.GenerateParameters(inputType)
	if op.inputBodyField != "" {
		op.operation.RequestBody = op.route.gen.GenerateRequestBody(op.operation.OperationId, op.inputBodyResolveTag, op.inputBody)
	}
	return op
}

// AddSecurity adds a security scheme to the operation.
func (op *OperationBuilder) AddSecurity(securityName string, scheme *v3.SecurityScheme) *OperationBuilder {
	op.route.gen.doc.Components.SecuritySchemes.Set(securityName, scheme)
	op.operation.Security = append(op.operation.Security, &base.SecurityRequirement{
		Requirements: orderedmap.New[string, []string](),
	})
	op.operation.Security = uniqBy(op.operation.Security, sameSecurityRequirement)
	return op
}

// AddJSONResponse adds a JSON response to the operation.
func (op *OperationBuilder) AddJSONResponse(code int, model any, description ...string) *OperationBuilder {
	return op.addResponse(code, "json", model, description...)
}

func (op *OperationBuilder) AddPlainTextResponse(code int, description ...string) *OperationBuilder {
	return op.addResponse(code, "text/plain", nil, description...)
}

func (op *OperationBuilder) AddResponse(code int, mediaType string, description ...string) *OperationBuilder {
	return op.addResponse(code, mediaType, nil, description...)
}

func (op *OperationBuilder) addResponse(code int, mediaType string, model any, description ...string) *OperationBuilder {
	ref := op.route.gen.GenerateResponse(code, mediaType, reflect.TypeOf(model), description...)
	if op.operation.Responses == nil {
		op.operation.Responses = &v3.Responses{
			Codes: orderedmap.New[string, *v3.Response](),
		}
	}
	scode := strconv.Itoa(code)
	if _, exists := op.operation.Responses.Codes.Get(scode); !exists {
		op.operation.Responses.Codes.Set(scode, &v3.Response{
			Content: orderedmap.New[string, *v3.MediaType](),
		})
	}

	for pair := ref.Content.First(); pair != nil; pair = pair.Next() {
		op.operation.Responses.Codes.Value(scode).Content.Set(pair.Key(), pair.Value())
	}
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
				PathItems: orderedmap.New[string, *v3.PathItem](),
			}
		}
		// clean the chi pattern, remove the regex etc from the parameters..
		path := cleanPath(op.pattern)
		if _, exists := op.route.gen.doc.Paths.PathItems.Get(path); !exists {
			op.route.gen.doc.Paths.PathItems.Set(path, &v3.PathItem{})
		}
		pathItem := op.route.gen.doc.Paths.PathItems.Value(path)

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
	// NOTE: currently only supports json
	if op.inputBodyField != "" {
		mediaType := resolveMediaType(op.inputBodyResolveTag)
		switch mediaType {
		case "application/json":
			body := reflect.New(op.inputBody).Interface()
			if err := json.NewDecoder(r.Body).Decode(body); err != nil {
				return err
			}
			reflect.ValueOf(input).Elem().FieldByName(op.inputBodyField).Set(reflect.ValueOf(body).Elem())
		case "multipart/form-data":
		}
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
