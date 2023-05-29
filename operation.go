package soda

import (
	"context"
	"log"
	"net/http"
	"reflect"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
)

// OperationBuilder is a builder for a single operation.
type OperationBuilder struct {
	operation *openapi3.Operation
	path      string
	method    string
	tInput    reflect.Type
	soda      *Soda

	handlers []fiber.Handler

	requestBody          reflect.Type
	requestBodyMediaType string
	requestBodyField     string
}

// SetSummary sets the operation-id.
func (op *OperationBuilder) SetOperationID(id string) *OperationBuilder {
	op.operation.OperationID = id
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
	op.operation.Tags = append(op.operation.Tags, tags...)
	for _, tag := range tags {
		if t := op.soda.generator.spec.Tags.Get(tag); t == nil {
			op.soda.generator.spec.Tags = append(op.soda.generator.spec.Tags, &openapi3.Tag{Name: tag})
		}
	}
	return op
}

// SetDeprecated marks the operation as deprecated.
func (op *OperationBuilder) SetDeprecated(deprecated bool) *OperationBuilder {
	op.operation.Deprecated = deprecated
	return op
}

// SetInput sets the input struct of the operation.
// The input must be a pointer to a struct. The struct will be used to generate the parameters of the operation.
// If the struct has a field with the `body` tag, then that field will be used to generate the request body of the operation.
// The body tag should be in the format `body:"<mediaType>"`, where `<mediaType>` is the media type of the request body.
func (op *OperationBuilder) SetInput(input interface{}) *OperationBuilder {
	inputType := reflect.TypeOf(input)
	// the input type should be a struct or pointer to a struct
	for inputType.Kind() == reflect.Ptr {
		inputType = inputType.Elem()
	}
	if inputType.Kind() != reflect.Struct {
		panic("input must be a pointer to a struct")
	}

	op.tInput = inputType
	for i := 0; i < inputType.NumField(); i++ {
		if body := inputType.Field(i); body.Tag.Get("body") != "" {
			op.requestBody = body.Type
			op.requestBodyMediaType = body.Tag.Get("body")
			op.requestBodyField = body.Name
			break
		}
	}
	op.operation.Parameters = op.soda.generator.GenerateParameters(inputType)
	if op.requestBodyField != "" {
		op.operation.RequestBody = op.soda.generator.GenerateRequestBody(op.operation.OperationID, op.requestBodyMediaType, op.requestBody)
	}
	return op
}

// AddJWTSecurity adds JWT authentication to this operation with the given validators.
func (op *OperationBuilder) AddJWTSecurity(validators ...fiber.Handler) *OperationBuilder {
	// add the validators to the beginning of the list of handlers
	op.handlers = append(validators, op.handlers...)

	// add the JWT security scheme to the spec if it doesn't already exist
	if len(op.soda.generator.spec.Components.SecuritySchemes) == 0 {
		op.soda.generator.spec.Components.SecuritySchemes = make(map[string]*openapi3.SecuritySchemeRef, 1)
	}
	op.soda.generator.spec.Components.SecuritySchemes["JWTAuth"] = &openapi3.SecuritySchemeRef{Value: openapi3.NewJWTSecurityScheme()}

	// add the security scheme to the operation
	if op.operation.Security == nil {
		op.operation.Security = openapi3.NewSecurityRequirements()
	}
	require := openapi3.NewSecurityRequirement().Authenticate("JWTAuth")
	op.operation.Security.With(require)
	return op
}

// AddJSONResponse adds a JSON response to the operation's responses.
// If model is not nil, a JSON response is generated for the model type.
// If model is nil, a JSON response is generated with no schema.
func (op *OperationBuilder) AddJSONResponse(status int, model interface{}) *OperationBuilder {
	if len(op.operation.Responses) == 0 {
		op.operation.Responses = make(openapi3.Responses)
	}
	if model == nil {
		op.operation.AddResponse(status, openapi3.NewResponse().WithDescription(http.StatusText(status)))
		return op
	}
	ref := op.soda.generator.GenerateResponse(op.operation.OperationID, status, reflect.TypeOf(model), "json")
	op.operation.Responses[strconv.Itoa(status)] = ref
	return op
}

func (op *OperationBuilder) OK() *OperationBuilder {
	// Add default response if not exists
	if op.operation.Responses == nil {
		op.operation.AddResponse(0, openapi3.NewResponse().WithDescription("OK"))
	}

	// Validate the operation
	if err := op.operation.Validate(context.TODO()); err != nil {
		log.Fatalln(err)
	}

	// Add operation to the spec
	op.soda.generator.spec.AddOperation(fixPath(op.path), op.method, op.operation)

	// Validate the spec
	if err := op.soda.generator.spec.Validate(context.TODO()); err != nil {
		log.Fatalln(err)
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
		if op.tInput == nil {
			return nil
		}

		// create a new instance of the input struct
		input := reflect.New(op.tInput).Interface()

		// parse the request parameters
		for _, parser := range parameterParsers {
			if err := parser(c, input); err != nil {
				return err
			}
		}

		// parse the request body
		if op.requestBodyField != "" {
			body := reflect.New(op.requestBody).Interface()
			if err := c.BodyParser(body); err != nil {
				return err
			}
			reflect.ValueOf(input).Elem().FieldByName(op.requestBodyField).Set(reflect.ValueOf(body).Elem())
		}

		// if the validator is not nil then validate the input struct
		if op.soda.validator != nil {
			if err := op.soda.validator.Struct(input); err != nil {
				return err
			}
		}

		// if the input implements the CustomizeValidate interface then call the Validate function
		if v, ok := input.(CustomizeValidate); ok {
			if err := v.Validate(c.Context()); err != nil {
				return err
			}
		}

		// add the input struct to the context
		c.Locals(KeyInput, input)
		return nil
	}
}
