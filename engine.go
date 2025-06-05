package soda

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v3"
)

type options struct {
	excludeUnexportFields bool
}

type Option func(*options)

// WithExcludeUnexportedFields is an option to exclude unexported fields from the OpenAPI spec generation.
func WithExcludeUnexportedFields() Option {
	return func(o *options) {
		o.excludeUnexportFields = true
	}
}

type Engine struct {
	*Router
	app            *fiber.App
	cachedSpecYAML []byte
	cachedSpecJSON []byte
}

func (e *Engine) OpenAPI() *openapi3.T {
	return e.gen.doc
}

func (e *Engine) App() *fiber.App {
	return e.app
}

func (e *Engine) ServeDocUI(pattern string, ui UIRender) *Engine {
	e.app.Get(pattern, func(c *fiber.Ctx) error {
		c.Context().SetContentType("text/html; charset=utf-8")
		return c.SendString(ui.Render(e.gen.doc))
	})
	return e
}

func (e *Engine) ServeSpecJSON(pattern string) *Engine {
	if e.cachedSpecJSON == nil {
		e.cachedSpecJSON, _ = e.gen.doc.MarshalJSON()
	}
	e.app.Get(pattern, func(c *fiber.Ctx) error {
		c.Context().SetContentType("application/json; charset=utf-8")
		return c.Send(e.cachedSpecJSON)
	})
	return e
}

func (e *Engine) ServeSpecYAML(pattern string) *Engine {
	if e.cachedSpecYAML == nil {
		spec, _ := yaml.Marshal(e.gen.doc)
		e.cachedSpecYAML = spec
	}
	e.app.Get(pattern, func(c *fiber.Ctx) error {
		c.Context().SetContentType("text/yaml; charset=utf-8")
		return c.Send(e.cachedSpecYAML)
	})
	return e
}

func New(opts ...Option) *Engine {
	return NewWith(fiber.New(), opts...)
}

func NewWith(app *fiber.App, opts ...Option) *Engine {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	return &Engine{
		app: app,
		Router: &Router{
			gen: NewGenerator(opts...),
			Raw: app,
		},
	}
}
