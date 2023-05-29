package soda

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
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
		generator: newGenerator(),
		Fiber:     app,
	}
}

// OpenAPI returns the OpenAPI spec.
func (s *Soda) OpenAPI() *openapi3.T {
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

// Get adds a GET operation.
func (s *Soda) Get(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, "GET", handlers...)
}

// Post adds a POST operation.
func (s *Soda) Post(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, "POST", handlers...)
}

// Put adds a PUT operation.
func (s *Soda) Put(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, "PUT", handlers...)
}

// Patch adds a PATCH operation.
func (s *Soda) Patch(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, "PATCH", handlers...)
}

// Delete adds a DELETE operation.
func (s *Soda) Delete(path string, handlers ...fiber.Handler) *OperationBuilder {
	return s.Operation(path, "DELETE", handlers...)
}

// Operation adds an operation.
func (s *Soda) Operation(path, method string, handlers ...fiber.Handler) *OperationBuilder {
	defaultSummary := method + " " + path
	defaultOperationID := genDefaultOperationID(method, path)

	return s.operation(path, method, handlers...).
		SetSummary(defaultSummary).
		SetOperationID(defaultOperationID)
}

func (s *Soda) operation(path, method string, handlers ...fiber.Handler) *OperationBuilder {
	operation := openapi3.NewOperation()
	return &OperationBuilder{
		operation: operation,
		path:      path,
		method:    method,
		tInput:    nil,
		soda:      s,
		handlers:  handlers,
	}
}
