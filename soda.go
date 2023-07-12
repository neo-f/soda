package soda

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Soda is the main class of the package.
// It contains the spec and the fiber app.
type Soda struct {
	generator *generator
	Fiber     *fiber.App
	validator *validator.Validate
}

// New creates a Soda instance.
func New(app *fiber.App) *Soda {
	return &Soda{
		generator: NewGenerator(),
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

// AddUI adds a UI to the given path, rendering the OpenAPI spec.
func (s *Soda) SetValidator(v *validator.Validate) *Soda {
	s.validator = v
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

	builder := &OperationBuilder{
		operation: openapi3.NewOperation(),
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
func (s *Soda) Group(prefix string, handlers ...fiber.Handler) *group {
	return &group{
		soda:       s,
		prefix:     prefix,
		tags:       []string{},
		handlers:   handlers,
		securities: make(map[string]*openapi3.SecurityScheme, 0),
	}
}

type group struct {
	soda       *Soda
	securities map[string]*openapi3.SecurityScheme
	prefix     string
	tags       []string
	handlers   []fiber.Handler
}

// AddTags add tags to the operation.
func (g *group) Group(prefix string, handlers ...fiber.Handler) *group {
	prefix = fmt.Sprintf("/%s/%s", strings.Trim(g.prefix, "/"), strings.Trim(prefix, "/"))
	return &group{
		soda:       g.soda,
		prefix:     prefix,
		tags:       g.tags,
		handlers:   append(g.handlers, handlers...),
		securities: g.securities,
	}
}

// AddTags add tags to the operation.
func (g *group) AddTags(tags ...string) *group {
	g.tags = append(g.tags, tags...)
	return g
}

func (g *group) AddSecurity(name string, scheme *openapi3.SecurityScheme) *group {
	g.securities[name] = scheme
	return g
}

// Get adds a GET operation.
func (g *group) Get(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, "GET", handlers...)
}

// Post adds a POST operation.
func (g *group) Post(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, "POST", handlers...)
}

// Put adds a PUT operation.
func (g *group) Put(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, "PUT", handlers...)
}

// Patch adds a PATCH operation.
func (g *group) Patch(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, "PATCH", handlers...)
}

// Delete adds a DELETE operation.
func (g *group) Delete(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, "DELETE", handlers...)
}

// Operation adds an operation.
func (g *group) Operation(path, method string, handlers ...fiber.Handler) *OperationBuilder {
	path = fmt.Sprintf("/%s/%s", strings.Trim(g.prefix, "/"), strings.Trim(path, "/"))
	op := g.soda.Operation(path, method, handlers...)
	op.AddTags(g.tags...)
	for name, scheme := range g.securities {
		op.AddSecurity(name, scheme)
	}
	return op
}
