package soda

import (
	"maps"
	"net/http"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v3"
)

// Router is an interface that represents a HTTP router.
type Router interface {
	// Router returns the underlying fiber.Router.
	Router() fiber.Router

	// Add registers a handler function for the specified HTTP method and pattern.
	Add(method string, pattern string, handler fiber.Handler) *OperationBuilder

	// Delete registers a handler function for the DELETE HTTP method and pattern.
	Delete(pattern string, handler fiber.Handler) *OperationBuilder

	// Get registers a handler function for the GET HTTP method and pattern.
	Get(pattern string, handler fiber.Handler) *OperationBuilder

	// Head registers a handler function for the HEAD HTTP method and pattern.
	Head(pattern string, handler fiber.Handler) *OperationBuilder

	// Options registers a handler function for the OPTIONS HTTP method and pattern.
	Options(pattern string, handler fiber.Handler) *OperationBuilder

	// Patch registers a handler function for the PATCH HTTP method and pattern.
	Patch(pattern string, handler fiber.Handler) *OperationBuilder

	// Post registers a handler function for the POST HTTP method and pattern.
	Post(pattern string, handler fiber.Handler) *OperationBuilder

	// Put registers a handler function for the PUT HTTP method and pattern.
	Put(pattern string, handler fiber.Handler) *OperationBuilder

	// Trace registers a handler function for the TRACE HTTP method and pattern.
	Trace(pattern string, handler fiber.Handler) *OperationBuilder

	// AddTags adds the specified tags to the router.
	AddTags(tags ...string) Router

	// AddSecurity adds the specified security scheme to the router.
	AddSecurity(securityName string, scheme *openapi3.SecurityScheme) Router

	// AddJSONResponse adds a JSON response definition to the router.
	AddJSONResponse(code int, model any, description ...string) Router

	// SetDeprecated sets whether the router is deprecated or not.
	SetDeprecated(deprecated bool) Router

	// SetIgnoreAPIDoc sets whether the router should be ignored in the API documentation.
	SetIgnoreAPIDoc(ignore bool) Router

	// OnAfterBind sets a hook function to be called after binding the request.
	OnAfterBind(hook HookAfterBind) Router

	// OnBeforeBind sets a hook function to be called before binding the request.
	OnBeforeBind(hook HookBeforeBind) Router
}

var _ Router = (*route)(nil)

type route struct {
	gen    *generator
	router fiber.Router

	commonTags       []string
	commonDeprecated bool
	commonResponses  map[int]*openapi3.Response
	commonSecurities openapi3.SecurityRequirements

	commonHooksBeforeBind []HookBeforeBind
	commonHooksAfterBind  []HookAfterBind

	ignoreAPIDoc bool
}

// Router implements Router.
func (r *route) Router() fiber.Router {
	return r.router
}

func (r *route) Add(method string, pattern string, handler fiber.Handler) *OperationBuilder {
	builder := &OperationBuilder{
		route: r,
		operation: &openapi3.Operation{
			Summary:     method + " " + pattern,
			OperationID: genDefaultOperationID(method, pattern),
			Security:    &r.commonSecurities,
		},
		method:  method,
		pattern: pattern,
		handler: handler,

		hooksBeforeBind: r.commonHooksBeforeBind,
		hooksAfterBind:  r.commonHooksAfterBind,
		ignoreAPIDoc:    r.ignoreAPIDoc,
	}

	for code, resp := range r.commonResponses {
		builder.operation.AddResponse(code, resp)
	}

	builder.AddTags(r.commonTags...)
	builder.SetDeprecated(r.commonDeprecated)
	return builder
}

func (r *route) Delete(pattern string, handler fiber.Handler) *OperationBuilder {
	return r.Add(http.MethodDelete, pattern, handler)
}

func (r *route) Get(pattern string, handler fiber.Handler) *OperationBuilder {
	return r.Add(http.MethodGet, pattern, handler)
}

func (r *route) Head(pattern string, handler fiber.Handler) *OperationBuilder {
	return r.Add(http.MethodHead, pattern, handler)
}

func (r *route) Options(pattern string, handler fiber.Handler) *OperationBuilder {
	return r.Add(http.MethodOptions, pattern, handler)
}

func (r *route) Patch(pattern string, handler fiber.Handler) *OperationBuilder {
	return r.Add(http.MethodPatch, pattern, handler)
}

func (r *route) Post(pattern string, handler fiber.Handler) *OperationBuilder {
	return r.Add(http.MethodPost, pattern, handler)
}

func (r *route) Put(pattern string, handler fiber.Handler) *OperationBuilder {
	return r.Add(http.MethodPut, pattern, handler)
}

func (r *route) Trace(pattern string, handler fiber.Handler) *OperationBuilder {
	return r.Add(http.MethodTrace, pattern, handler)
}

func (r *route) AddTags(tags ...string) Router {
	r.commonTags = append(r.commonTags, tags...)

	for _, tag := range tags {
		r.gen.doc.Tags = append(r.gen.doc.Tags, &openapi3.Tag{
			Name: tag,
		})
	}
	return r
}

func (r *route) SetDeprecated(deprecated bool) Router {
	r.commonDeprecated = deprecated
	return r
}

func (r *route) AddSecurity(securityName string, scheme *openapi3.SecurityScheme) Router {
	r.gen.doc.Components.SecuritySchemes[securityName] = &openapi3.SecuritySchemeRef{Value: scheme}
	r.commonSecurities = append(
		r.commonSecurities,
		openapi3.SecurityRequirement{securityName: nil},
	)
	return r
}

// SetIgnoreAPIDoc implements Router.
func (r *route) SetIgnoreAPIDoc(ignore bool) Router {
	r.ignoreAPIDoc = ignore
	return r
}

func (r *route) OnAfterBind(hook HookAfterBind) Router {
	r.commonHooksAfterBind = append(r.commonHooksAfterBind, hook)
	return r
}

func (r *route) OnBeforeBind(hook HookBeforeBind) Router {
	r.commonHooksBeforeBind = append(r.commonHooksBeforeBind, hook)
	return r
}

func (r *route) AddJSONResponse(code int, model any, description ...string) Router {
	if r.commonResponses == nil {
		r.commonResponses = make(map[int]*openapi3.Response)
	}
	resp := r.gen.GenerateResponse(code, reflect.TypeOf(model), "application/json", description...)
	r.commonResponses[code] = resp
	return r
}

func (r *route) Group(prefix string, handlers ...fiber.Handler) Router {
	return &route{
		gen:                   r.gen,
		router:                r.router.Group(prefix, handlers...),
		commonTags:            r.commonTags,
		commonDeprecated:      r.commonDeprecated,
		commonResponses:       maps.Clone(r.commonResponses),
		commonSecurities:      r.commonSecurities,
		commonHooksBeforeBind: r.commonHooksBeforeBind,
		commonHooksAfterBind:  r.commonHooksAfterBind,
		ignoreAPIDoc:          r.ignoreAPIDoc,
	}
}
