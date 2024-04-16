package soda

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v3"
	"gopkg.in/yaml.v3"
)

type Engine struct {
	*route

	cachedSpecYAML []byte
	cachedSpecJSON []byte
}

func (e *Engine) OpenAPI() *openapi3.T {
	return e.gen.doc
}

func (e *Engine) ServeDocUI(pattern string, ui UIRender) *Engine {
	e.router.Get(pattern, func(c fiber.Ctx) error {
		c.Context().SetContentType("text/html; charset=utf-8")
		return c.SendString(ui.Render(e.gen.doc))
	})
	return e
}

func (e *Engine) ServeSpecJSON(pattern string) *Engine {
	if e.cachedSpecJSON == nil {
		e.cachedSpecJSON, _ = e.gen.doc.MarshalJSON()
	}
	e.router.Get(pattern, func(c fiber.Ctx) error {
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
	e.router.Get(pattern, func(c fiber.Ctx) error {
		c.Context().SetContentType("text/yaml; charset=utf-8")
		return c.Send(e.cachedSpecYAML)
	})
	return e
}

func New() *Engine {
	return NewWith(fiber.New())
}

func NewWith(f *fiber.App) *Engine {
	return &Engine{
		route: &route{
			gen:    NewGenerator(),
			router: f,
		},
	}
}
