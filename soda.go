package soda

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sv-tools/openapi/spec"
	"gopkg.in/yaml.v3"
)

// Soda is the main class of the package.
// It contains the spec and the fiber app.
type Soda struct {
	generator *generator
	Fiber     *fiber.App
}

// New creates a Soda instance.
func New(app *fiber.App) *Soda {
	return &Soda{
		generator: NewGenerator(),
		Fiber:     app,
	}
}

// OpenAPI returns the OpenAPI spec.
func (s *Soda) OpenAPI() *spec.OpenAPI {
	return s.generator.spec
}

// AddUI adds a UI to the given path, rendering the OpenAPI spec.
func (s *Soda) AddUI(path string, ui UIRender) *Soda {
	s.Fiber.Get(path, func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(ui.Render(s.OpenAPI()))
	})
	return s
}

// AddJSONSpec adds the OpenAPI spec at the given path in JSON format.
func (s *Soda) AddJSONSpec(path string) *Soda {
	s.Fiber.Get(path, func(c *fiber.Ctx) error {
		return c.JSON(s.OpenAPI())
	})
	return s
}

// AddYAMLSpec adds the OpenAPI spec at the given path in YAML format.
func (s *Soda) AddYAMLSpec(path string) *Soda {
	s.Fiber.Get(path, func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/yaml; charset=utf-8")
		spec, err := yaml.Marshal(s.generator.spec)
		if err != nil {
			return err
		}
		return c.Send(spec)
	})
	return s
}

// Get adds a GET operation.
func (s *Soda) Get(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, fiber.MethodGet, handlers...)
}

// Head adds a HEAD operation.
func (s *Soda) Head(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, fiber.MethodHead, handlers...)
}

// Post adds a POST operation.
func (s *Soda) Post(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, fiber.MethodPost, handlers...)
}

// Put adds a PUT operation.
func (s *Soda) Put(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, fiber.MethodPut, handlers...)
}

// Delete adds a DELETE operation.
func (s *Soda) Delete(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, fiber.MethodDelete, handlers...)
}

// Options adds an OPTIONS operation.
func (s *Soda) Options(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, fiber.MethodOptions, handlers...)
}

// Trace adds a TRACE operation.
func (s *Soda) Trace(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, fiber.MethodTrace, handlers...)
}

// Patch adds a PATCH operation.
func (s *Soda) Patch(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, fiber.MethodPatch, handlers...)
}

// Operation adds an operation.
func (s *Soda) Operation(path, method string, handlers ...fiber.Handler) *OperationBuilder {
	defaultSummary := method + " " + path
	defaultOperationID := genDefaultOperationID(method, path)

	builder := &OperationBuilder{
		operation: spec.NewOperation(),
		path:      path,
		method:    method,
		input:     nil,
		soda:      s,
		handlers:  handlers,
	}
	builder.SetSummary(defaultSummary).SetOperationID(defaultOperationID)
	return builder
}

// Group creates a new sub-group with optional prefix and middleware.
func (s *Soda) Group(prefix string, handlers ...fiber.Handler) *Group {
	return &Group{
		soda:       s,
		prefix:     prefix,
		tags:       []string{},
		handlers:   handlers,
		securities: make(map[string]*spec.SecurityScheme, 0),
	}
}
