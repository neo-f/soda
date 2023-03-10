package soda

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Options struct {
	swaggerPath          *string
	rapiDocPath          *string
	redocPath            *string
	stoplightElementPath *string
	openAPISpecJSONPath  *string
	validator            *validator.Validate
	fiberConfig          []fiber.Config
}
type Option func(o *Options)

func WithOpenAPISpec(path string) Option {
	return func(o *Options) {
		o.openAPISpecJSONPath = &path
	}
}

func WithSwagger(path string) Option {
	return func(o *Options) {
		o.swaggerPath = &path
	}
}

func WithRedoc(path string) Option {
	return func(o *Options) {
		o.redocPath = &path
	}
}

func WithStoplightElements(path string) Option {
	return func(o *Options) {
		o.stoplightElementPath = &path
	}
}

func WithRapiDoc(path string) Option {
	return func(o *Options) {
		o.rapiDocPath = &path
	}
}

func WithFiberConfig(config ...fiber.Config) Option {
	return func(o *Options) {
		o.fiberConfig = config
	}
}

func EnableValidateRequest(v ...*validator.Validate) Option {
	var validate *validator.Validate
	if len(v) == 0 {
		validate = validator.New()
	} else {
		validate = v[0]
	}
	return func(o *Options) {
		o.validator = validate
	}
}

type Soda struct {
	specOnce     sync.Once
	oaiGenerator *oaiGenerator
	Options      *Options
	*fiber.App
	spec []byte
}

func (s *Soda) GetOpenAPIJSON() []byte {
	s.specOnce.Do(func() {
		if err := s.oaiGenerator.openapi.Validate(context.TODO()); err != nil {
			log.Fatalln(err)
		}
		spec, err := s.oaiGenerator.openapi.MarshalJSON()
		if err != nil {
			log.Fatalln(err)
		}
		s.spec = spec
	})
	return s.spec
}

func (s *Soda) OpenAPI() *openapi3.T {
	return s.oaiGenerator.openapi
}

func New(title, version string, options ...Option) *Soda {
	opt := &Options{}
	for _, option := range options {
		option(opt)
	}

	s := &Soda{
		oaiGenerator: newGenerator(&openapi3.Info{Title: title, Version: version}),
		App:          fiber.New(opt.fiberConfig...),
		Options:      opt,
	}

	if opt.openAPISpecJSONPath != nil {
		s.Get(*opt.openAPISpecJSONPath, func(ctx *fiber.Ctx) error {
			ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
			return ctx.Send(s.GetOpenAPIJSON())
		}).
			AddTags("Documentation").
			SetOperationID("doc-openapi").
			SetSummary("OpenAPI Specification").
			SetDescription(`[OpenAPI3](https://swagger.io/specification) OpenAPI Specification File Download`).
			AddResponseWithContentType(200, fiber.MIMEApplicationJSONCharsetUTF8).
			OK()
	}

	if opt.redocPath != nil {
		s.Get(*opt.redocPath, func(ctx *fiber.Ctx) error {
			ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
			return ctx.SendString(s.htmlRedoc())
		}).
			AddTags("Documentation").
			SetOperationID("doc-redoc").
			SetSummary("redoc").
			SetDescription(`[Redoc](https://github.com/Redocly/redoc) OpenAPI Renderer`).
			AddResponseWithContentType(200, fiber.MIMETextHTMLCharsetUTF8).
			OK()
	}

	if opt.swaggerPath != nil {
		s.Get(*opt.swaggerPath, func(ctx *fiber.Ctx) error {
			ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
			return ctx.SendString(s.htmlSwagger())
		}).
			AddTags("Documentation").
			SetOperationID("doc-swagger").
			SetSummary("swagger").
			SetDescription(`[Swagger UI](https://swagger.io/tools/swagger-ui/) OpenAPI Renderer`).
			AddResponseWithContentType(200, fiber.MIMETextHTMLCharsetUTF8).
			OK()
	}

	if opt.rapiDocPath != nil {
		s.Get(*opt.rapiDocPath, func(ctx *fiber.Ctx) error {
			ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
			return ctx.SendString(s.htmlRapiDoc())
		}).
			AddTags("Documentation").
			SetOperationID("doc-rapidoc").
			SetSummary("rapidoc").
			SetDescription(`[RapiDoc](https://github.com/mrin9/RapiDoc) OpenAPI Renderer`).
			AddResponseWithContentType(200, fiber.MIMETextHTMLCharsetUTF8).
			OK()
	}

	if opt.stoplightElementPath != nil {
		s.Get(*opt.stoplightElementPath, func(ctx *fiber.Ctx) error {
			ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
			return ctx.SendString(s.htmlStoplightElements())
		}).
			AddTags("Documentation").
			SetOperationID("doc-elements").
			SetSummary("stoplight elements").
			SetDescription(`[Elements](https://github.com/stoplightio/elements) OpenAPI Renderer`).
			AddResponseWithContentType(200, fiber.MIMETextHTMLCharsetUTF8).
			OK()
	}
	return s
}

func (s *Soda) newOperation(path, method string, handlers ...fiber.Handler) *Operation {
	operation := openapi3.NewOperation()
	operation.AddResponse(0, openapi3.NewResponse().WithDescription("OK"))
	op := &Operation{
		Operation:    operation,
		Path:         path,
		Method:       method,
		TParameters:  nil,
		TRequestBody: nil,
		Soda:         s,
		handlers:     handlers,
	}
	return op
}

func (s *Soda) Get(path string, handlers ...fiber.Handler) *Operation {
	return s.Handle(path, "GET", handlers...)
}
func (s *Soda) Post(path string, handlers ...fiber.Handler) *Operation {
	return s.Handle(path, "POST", handlers...)
}
func (s *Soda) Put(path string, handlers ...fiber.Handler) *Operation {
	return s.Handle(path, "PUT", handlers...)
}
func (s *Soda) Patch(path string, handlers ...fiber.Handler) *Operation {
	return s.Handle(path, "PATCH", handlers...)
}
func (s *Soda) Delete(path string, handlers ...fiber.Handler) *Operation {
	return s.Handle(path, "DELETE", handlers...)
}
func (s *Soda) Handle(path, method string, handlers ...fiber.Handler) *Operation {
	summary := method + " " + path
	return s.newOperation(path, method, handlers...).SetSummary(summary).SetOperationID(genID(path, method))
}

func genID(path, method string) string {
	nt := true
	var s string
	for _, r := range path {
		switch r {
		case ':', '-', '_', '/', '.':
			nt = true
		default:
			if nt {
				s += strings.ToUpper(string(r))
			} else {
				s += string(r)
			}
			nt = false
		}
	}
	return s + method
}
