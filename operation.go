package soda

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sv-tools/openapi/spec"
)

// OperationBuilder is a builder for a single operation.
type OperationBuilder struct {
	input     reflect.Type
	inputBody reflect.Type

	soda      *Soda
	operation *spec.Extendable[spec.Operation]

	path               string
	method             string
	inputBodyMediaType string
	inputBodyField     string

	handlers []fiber.Handler
}

// SetSummary sets the operation-id.
func (op *OperationBuilder) SetOperationID(id string) *OperationBuilder {
	op.operation.Spec.OperationID = id
	return op
}

// SetSummary sets the operation summary.
func (op *OperationBuilder) SetSummary(summary string) *OperationBuilder {
	op.operation.Spec.Summary = summary
	return op
}

// SetDescription sets the operation description.
func (op *OperationBuilder) SetDescription(desc string) *OperationBuilder {
	op.operation.Spec.Description = desc
	return op
}

// AddTags add tags to the operation.
func (op *OperationBuilder) AddTags(tags ...string) *OperationBuilder {
	op.operation.Spec.Tags = append(op.operation.Spec.Tags, tags...)
	for _, tag := range tags {
		found := false
		for _, t := range op.soda.generator.spec.Tags {
			if t.Spec.Name == tag {
				found = true
				break
			}
		}
		if !found {
			newTag := spec.NewTag()
			newTag.Spec.Name = tag
			op.soda.generator.spec.Tags = append(op.soda.generator.spec.Tags, newTag)
		}
	}
	return op
}

// SetDeprecated marks the operation as deprecated.
func (op *OperationBuilder) SetDeprecated(deprecated bool) *OperationBuilder {
	op.operation.Spec.Deprecated = deprecated
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
	op.operation.Spec.Parameters = op.soda.generator.GenerateParameters(inputType)
	if op.inputBodyField != "" {
		op.operation.Spec.RequestBody = op.soda.generator.GenerateRequestBody(op.operation.Spec.OperationID, op.inputBodyMediaType, op.inputBody)
	}
	return op
}

// AddJWTSecurity adds JWT authentication to this operation with the given validators.
func (op *OperationBuilder) AddSecurity(name string, scheme *spec.SecurityScheme) *OperationBuilder {
	// add the JWT security scheme to the spec if it doesn't already exist
	if op.soda.generator.spec.Components.Spec.SecuritySchemes == nil {
		op.soda.generator.spec.Components.Spec.SecuritySchemes = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.SecurityScheme]])
	}

	securityScheme := spec.NewSecuritySchemeSpec()
	securityScheme.Spec.Spec = scheme
	op.soda.generator.spec.Components.Spec.WithRefOrSpec(name, securityScheme)

	// add the security scheme to the operation
	found := false
	for _, security := range op.operation.Spec.Security {
		if _, ok := security[name]; ok {
			found = true
			break
		}
	}
	if !found {
		newSecurity := spec.NewSecurityRequirement()
		newSecurity[name] = nil
		op.operation.Spec.Security = append(op.operation.Spec.Security, newSecurity)
	}
	return op
}

// AddJSONResponse adds a JSON response to the operation's responses.
// If model is not nil, a JSON response is generated for the model type.
// If model is nil, a JSON response is generated with no schema.
func (op *OperationBuilder) AddJSONResponse(status int, model interface{}) *OperationBuilder {
	if op.operation.Spec.Responses == nil {
		op.operation.Spec.Responses = spec.NewResponses()
		op.operation.Spec.Responses.Spec.Response = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.Response]])
	}
	code := strconv.FormatInt(int64(status), 10)
	if model == nil {
		newResponse := spec.NewResponseSpec()
		newResponse.Spec.Spec.Description = http.StatusText(status)
		op.operation.Spec.Responses.Spec.Response[code] = newResponse
		return op
	}
	ref := op.soda.generator.GenerateResponse(op.operation.Spec.OperationID, status, reflect.TypeOf(model), "json")
	op.operation.Spec.Responses.Spec.Response[code] = ref
	return op
}

func (op *OperationBuilder) OK() *OperationBuilder {
	// Add default response if not exists
	if op.operation.Spec.Responses == nil {
		if op.operation.Spec.Responses == nil {
			op.operation.Spec.Responses = spec.NewResponses()
			op.operation.Spec.Responses.Spec.Response = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.Response]])
		}
		op.operation.Spec.Responses.Spec.Response["default"] = spec.NewResponseSpec()
	}

	// Add operation to the spec
	if op.soda.generator.spec.Paths == nil {
		op.soda.generator.spec.Paths = spec.NewPaths()
		op.soda.generator.spec.Paths.Spec.Paths = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.PathItem]])
	}
	path := fixPath(op.path)
	if op.soda.generator.spec.Paths.Spec.Paths[path] == nil {
		op.soda.generator.spec.Paths.Spec.Paths[path] = spec.NewPathItemSpec()
	}
	pathItem := op.soda.generator.spec.Paths.Spec.Paths[path]

	switch strings.ToUpper(op.method) {
	case fiber.MethodGet:
		pathItem.Spec.Spec.Get = op.operation
	case fiber.MethodHead:
		pathItem.Spec.Spec.Head = op.operation
	case fiber.MethodPost:
		pathItem.Spec.Spec.Post = op.operation
	case fiber.MethodPut:
		pathItem.Spec.Spec.Put = op.operation
	case fiber.MethodPatch:
		pathItem.Spec.Spec.Patch = op.operation
	case fiber.MethodDelete:
		pathItem.Spec.Spec.Delete = op.operation
	case fiber.MethodOptions:
		pathItem.Spec.Spec.Options = op.operation
	case fiber.MethodTrace:
		pathItem.Spec.Spec.Trace = op.operation
	default:
		panic(fmt.Errorf("unsupported HTTP method %q", op.method))
	}

	// Add handler
	op.handlers = append([]fiber.Handler{op.bindInput()}, op.handlers...)

	// Add route to the fiber app
	op.soda.Fiber.Add(op.method, op.path, op.handlers...)

	return op
}

// bindInput binds the request body to the input struct.
func (op *OperationBuilder) bindInput() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if op.input == nil {
			return c.Next()
		}

		// create a new instance of the input struct
		input := reflect.New(op.input).Interface()

		// parse the request parameters
		for _, parser := range parameterParsers {
			if err := parser(c, input); err != nil {
				return err
			}
		}

		// parse the request body
		if op.inputBodyField != "" {
			body := reflect.New(op.inputBody).Interface()
			if err := c.BodyParser(body); err != nil {
				return err
			}
			reflect.ValueOf(input).Elem().FieldByName(op.inputBodyField).Set(reflect.ValueOf(body).Elem())
		}

		// if the validator is not nil then validate the input struct
		if op.soda.validator != nil {
			if err := op.soda.validator.Struct(input); err != nil {
				return err
			}
		}

		// if the input implements the CustomizeValidate interface then call the Validate function
		if v, ok := input.(customizeValidate); ok {
			if err := v.Validate(); err != nil {
				return err
			}
		}
		// if the input implements the CustomizeValidateCtx interface then call the Validate function
		if v, ok := input.(customizeValidateCtx); ok {
			if err := v.Validate(c.Context()); err != nil {
				return err
			}
		}

		// add the input struct to the context
		c.Locals(KeyInput, input)
		return c.Next()
	}
}
