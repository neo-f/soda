package soda

import (
	"net/http"
	"reflect"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

type (
	// HookBeforeBind is a function type that is called before binding the request. It returns a boolean indicating whether to continue the process.
	HookBeforeBind func(ctx *gin.Context)

	// HookAfterBind is a function type that is called after binding the request. It returns a boolean indicating whether to continue the process.
	HookAfterBind func(ctx *gin.Context, input any)
)

// OperationBuilder is a struct that helps in building an operation.
type OperationBuilder struct {
	route     *Router
	operation *openapi3.Operation

	method      string
	patternFull string
	pattern     string

	input              reflect.Type
	inputBody          reflect.Type
	inputBodyField     string
	inputBodyMediaType string

	handlers []gin.HandlerFunc

	ignoreAPIDoc bool

	// hooks
	hooksBeforeBind []HookBeforeBind
	hooksAfterBind  []HookAfterBind
}

// SetOperationID sets the operation ID of the operation.
func (op *OperationBuilder) SetOperationID(id string) *OperationBuilder {
	op.operation.OperationID = id
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
	// op.operation.Tags = append(op.operation.Tags, tags...)
	for _, tag := range tags {
		if !slices.Contains(op.operation.Tags, tag) {
			op.operation.Tags = append(op.operation.Tags, tag)
		}
		if op.route.gen.doc.Tags.Get(tag) == nil {
			op.route.gen.doc.Tags = append(op.route.gen.doc.Tags, &openapi3.Tag{Name: tag})
		}
	}
	return op
}

// SetDeprecated marks the operation as deprecated or not.
func (op *OperationBuilder) SetDeprecated(deprecated bool) *OperationBuilder {
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
		panic("input must be a struct")
	}

	op.input = inputType
	op.setInputBody(inputType)

	op.operation.Parameters = op.route.gen.GenerateParameters(inputType)
	op.setRequestBody()
	return op
}

// setInputBody sets the input body from the input type.
func (op *OperationBuilder) setInputBody(inputType reflect.Type) {
	for i := 0; i < inputType.NumField(); i++ {
		if body := inputType.Field(i); body.Tag.Get("body") != "" {
			op.inputBody = body.Type
			op.inputBodyMediaType = body.Tag.Get("body")
			op.inputBodyField = body.Name
			break
		}
	}
}

// setRequestBody sets the request body.
func (op *OperationBuilder) setRequestBody() {
	if op.inputBodyField == "" {
		return
	}
	op.operation.RequestBody = &openapi3.RequestBodyRef{
		Value: op.route.gen.GenerateRequestBody(
			op.operation.OperationID,
			op.inputBodyMediaType,
			op.inputBody,
		),
	}
}

// AddSecurity adds a security scheme to the operation.
func (op *OperationBuilder) AddSecurity(securityName string, scheme *openapi3.SecurityScheme) *OperationBuilder {
	op.route.gen.doc.Components.SecuritySchemes[securityName] = &openapi3.SecuritySchemeRef{
		Value: scheme,
	}
	op.operation.Security.With(openapi3.NewSecurityRequirement().Authenticate(securityName))
	return op
}

// AddJSONResponse adds a JSON response to the operation.
func (op *OperationBuilder) AddJSONResponse(code int, model any, description ...string) *OperationBuilder {
	desc := http.StatusText(code)
	if len(description) > 0 {
		desc = description[0]
	}
	ref := op.route.gen.GenerateResponse(code, model, "application/json", desc)
	op.operation.AddResponse(code, ref)
	return op
}

// SetIgnoreAPIDoc sets whether to ignore the operation when generating the API doc.
func (op *OperationBuilder) IgnoreAPIDoc(ignore bool) *OperationBuilder {
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
		path := cleanPath(op.patternFull)
		op.route.gen.doc.AddOperation(path, op.method, op.operation)
	}
	handlers := append([]gin.HandlerFunc{op.bindInput}, op.handlers...)
	op.route.Raw.Handle(op.method, op.pattern, handlers...)
}

// bindInput binds the request body to the input struct.
func (op *OperationBuilder) bindInput(ctx *gin.Context) {
	// Execute Hooks: BeforeBind
	for _, hook := range op.hooksBeforeBind {
		hook(ctx)
		if ctx.IsAborted() {
			return
		}
	}

	if op.input == nil {
		ctx.Next()
		return
	}

	// Bind input
	input := reflect.New(op.input).Interface()

	// Bind the parameters
	binders := []func(*gin.Context, any) error{
		BindQuery,
		BindPath,
		BindHeader,
		BindCookie,
	}
	for _, binder := range binders {
		if err := binder(ctx, input); err != nil {
			_ = ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
			return
		}
	}

	// Bind the request body
	switch op.inputBodyMediaType {
	case "application/json", "json":
		body := reflect.New(op.inputBody).Interface()
		if err := ctx.ShouldBindJSON(body); err != nil {
			_ = ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
			return
		}
		reflect.ValueOf(input).Elem().FieldByName(op.inputBodyField).Set(reflect.ValueOf(body).Elem())
		// TODO: Add more media types support
	}

	// Execute Hooks: AfterBind
	for _, hook := range op.hooksAfterBind {
		hook(ctx, input)
	}

	ctx.Set(KeyInput, input)
	ctx.Next()
}
