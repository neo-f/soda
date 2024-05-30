package soda

import (
	"net/http"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/gorilla/schema"
)

type (
	// HookBeforeBind is a function type that is called before binding the request. It returns a boolean indicating whether to continue the process.
	HookBeforeBind func(ctx *fiber.Ctx) error

	// HookAfterBind is a function type that is called after binding the request. It returns a boolean indicating whether to continue the process.
	HookAfterBind func(ctx *fiber.Ctx, input any) error
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

	handlers []fiber.Handler

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
	handlers := append([]fiber.Handler{op.bindInput}, op.handlers...)
	op.route.Raw.Add(op.method, op.pattern, handlers...).Name(op.operation.OperationID)
}

// bindInput binds the request body to the input struct.
func (op *OperationBuilder) bindInput(ctx *fiber.Ctx) error {
	// Execute Hooks: BeforeBind
	for _, hook := range op.hooksBeforeBind {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	if op.input == nil {
		return ctx.Next()
	}

	// Bind input
	input := reflect.New(op.input).Interface()

	// Bind the input
	binders := []func(any) error{
		bindPath(ctx),
		bindHeader(ctx),
		ctx.QueryParser,
		ctx.CookieParser,
	}
	for _, binder := range binders {
		if err := binder(input); err != nil {
			return err
		}
	}

	// Bind the request body
	if op.inputBodyField != "" {
		body := reflect.New(op.inputBody).Interface()
		if err := ctx.BodyParser(body); err != nil {
			return err
		}
		reflect.ValueOf(input).Elem().FieldByName(op.inputBodyField).Set(reflect.ValueOf(body).Elem())
	}

	// Execute Hooks: AfterBind
	for _, hook := range op.hooksAfterBind {
		if err := hook(ctx, input); err != nil {
			return err
		}
	}

	ctx.Locals(KeyInput, input)
	return ctx.Next()
}

var decoderPools = map[string]*sync.Pool{
	PathTag:   {New: func() any { return buildDecoder(PathTag) }},
	HeaderTag: {New: func() any { return buildDecoder(HeaderTag) }},
}

func buildDecoder(tag string) *schema.Decoder {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag(tag)
	decoder.IgnoreUnknownKeys(true)
	decoder.ZeroEmpty(true)
	return decoder
}

func bindPath(c *fiber.Ctx) func(any) error {
	return func(out any) error {
		params := c.Route().Params
		data := make(map[string][]string, len(params))
		for _, param := range params {
			data[param] = append(data[param], c.Params(param))
		}

		pathDecoder := decoderPools[PathTag].Get().(*schema.Decoder)
		defer decoderPools[PathTag].Put(pathDecoder)
		return pathDecoder.Decode(out, data)
	}
}

func bindHeader(c *fiber.Ctx) func(any) error {
	return func(out any) error {
		data := make(map[string][]string)
		c.Request().Header.VisitAll(func(key, val []byte) {
			k := string(key)
			v := string(val)

			if c.App().Config().EnableSplittingOnParsers && strings.Contains(v, ",") && equalFieldType(out, reflect.Slice, k, HeaderTag) {
				values := strings.Split(v, ",")
				for i := 0; i < len(values); i++ {
					data[k] = append(data[k], values[i])
				}
			} else {
				data[k] = append(data[k], v)
			}
		})

		headerDecoder := decoderPools[HeaderTag].Get().(*schema.Decoder)
		defer decoderPools[HeaderTag].Put(headerDecoder)
		return headerDecoder.Decode(out, data)
	}
}

// steal from fiber ;)
func equalFieldType(out interface{}, kind reflect.Kind, key, tag string) bool {
	// Get type of interface
	outTyp := reflect.TypeOf(out).Elem()
	key = strings.ToLower(key)
	// Must be a struct to match a field
	if outTyp.Kind() != reflect.Struct {
		return false
	}
	// Copy interface to an value to be used
	outVal := reflect.ValueOf(out).Elem()
	// Loop over each field
	for i := 0; i < outTyp.NumField(); i++ {
		// Get field value data
		structField := outVal.Field(i)
		// Can this field be changed?
		if !structField.CanSet() {
			continue
		}
		// Get field key data
		typeField := outTyp.Field(i)
		// Get type of field key
		structFieldKind := structField.Kind()
		// Does the field type equals input?
		if structFieldKind != kind {
			continue
		}
		// Get tag from field if exist
		inputFieldName := typeField.Tag.Get(tag)
		if inputFieldName == "" {
			inputFieldName = typeField.Name
		} else {
			inputFieldName = strings.Split(inputFieldName, ",")[0]
		}
		// Compare field/tag with provided key
		if strings.ToLower(inputFieldName) == key {
			return true
		}
	}
	return false
}
