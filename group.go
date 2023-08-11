package soda

import (
	"fmt"
	"path"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sv-tools/openapi/spec"
)

type Group struct {
	soda       *Soda
	securities map[string]*spec.SecurityScheme
	prefix     string
	tags       []string
	handlers   []fiber.Handler
	deprecated bool

	hooksAfterBind []HookAfterBind
}

// Group creates a new sub-group with optional prefix and middleware.
func (g *Group) Group(prefix string, handlers ...fiber.Handler) *Group {
	prefix = "/" + path.Join(strings.Trim(g.prefix, "/"), strings.Trim(prefix, "/"))
	return &Group{
		soda:       g.soda,
		securities: g.securities,
		prefix:     prefix,
		tags:       g.tags,
		handlers:   append(g.handlers, handlers...),
		deprecated: g.deprecated,

		hooksAfterBind: g.hooksAfterBind,
	}
}

// AddTags adds tags to the operation.
func (g *Group) AddTags(tags ...string) *Group {
	g.tags = append(g.tags, tags...)
	return g
}

// SetDeprecated sets the deprecated flag for the group.
func (g *Group) SetDeprecated(deprecated bool) *Group {
	g.deprecated = deprecated
	return g
}

// AddSecurity adds a security scheme to the group.
func (g *Group) AddSecurity(name string, scheme *spec.SecurityScheme) *Group {
	g.securities[name] = scheme
	return g
}

// OnAfterBind adds a hook to be executed after the operation is bound.
func (g *Group) OnAfterBind(hook HookAfterBind) *Group {
	g.hooksAfterBind = append(g.hooksAfterBind, hook)
	return g
}

// Fiber returns the underlying fiber app.
func (g *Group) Fiber() *fiber.App {
	return g.soda.Fiber
}

// Get adds a GET operation.
func (g *Group) Get(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, fiber.MethodGet, handlers...)
}

// Head adds a HEAD operation.
func (g *Group) Head(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, fiber.MethodHead, handlers...)
}

// Post adds a POST operation.
func (g *Group) Post(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, fiber.MethodPost, handlers...)
}

// Put adds a PUT operation.
func (g *Group) Put(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, fiber.MethodPut, handlers...)
}

// Delete adds a DELETE operation.
func (g *Group) Delete(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, fiber.MethodDelete, handlers...)
}

// Options adds an OPTIONS operation.
func (g *Group) Options(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, fiber.MethodOptions, handlers...)
}

// Trace adds a TRACE operation.
func (g *Group) Trace(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, fiber.MethodTrace, handlers...)
}

// Patch adds a PATCH operation.
func (g *Group) Patch(path string, handlers ...fiber.Handler) *OperationBuilder {
	return g.Operation(path, fiber.MethodPatch, handlers...)
}

// Operation adds an operation.
func (g *Group) Operation(path, method string, handlers ...fiber.Handler) *OperationBuilder {
	path = fmt.Sprintf("/%s/%s", strings.Trim(g.prefix, "/"), strings.Trim(path, "/"))
	path = strings.TrimSuffix(path, "/")
	handlers = append(g.handlers, handlers...)
	op := g.soda.Operation(path, method, handlers...)
	op.AddTags(g.tags...)
	for _, hook := range g.hooksAfterBind {
		op.OnAfterBind(hook)
	}
	op.SetDeprecated(g.deprecated)
	for name, scheme := range g.securities {
		op.AddSecurity(name, scheme)
	}
	return op
}
