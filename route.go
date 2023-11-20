package soda

import (
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Router is an interface that represents a HTTP router.
type Router interface {
	// HttpHandler returns the underlying chi.Router.
	HttpHandler() chi.Router

	// Method registers a handler function for the specified HTTP method and pattern.
	Method(method string, pattern string, handler http.HandlerFunc) *OperationBuilder

	// Delete registers a handler function for the DELETE HTTP method and pattern.
	Delete(pattern string, handler http.HandlerFunc) *OperationBuilder

	// Get registers a handler function for the GET HTTP method and pattern.
	Get(pattern string, handler http.HandlerFunc) *OperationBuilder

	// Head registers a handler function for the HEAD HTTP method and pattern.
	Head(pattern string, handler http.HandlerFunc) *OperationBuilder

	// Options registers a handler function for the OPTIONS HTTP method and pattern.
	Options(pattern string, handler http.HandlerFunc) *OperationBuilder

	// Patch registers a handler function for the PATCH HTTP method and pattern.
	Patch(pattern string, handler http.HandlerFunc) *OperationBuilder

	// Post registers a handler function for the POST HTTP method and pattern.
	Post(pattern string, handler http.HandlerFunc) *OperationBuilder

	// Put registers a handler function for the PUT HTTP method and pattern.
	Put(pattern string, handler http.HandlerFunc) *OperationBuilder

	// Trace registers a handler function for the TRACE HTTP method and pattern.
	Trace(pattern string, handler http.HandlerFunc) *OperationBuilder

	// Mount mounts a sub-router under the specified pattern.
	Mount(pattern string, sub Router)

	// Group creates a new sub-router and applies the provided function to it.
	Group(fn func(Router)) Router

	// With adds the specified middlewares to the router.
	With(middlewares ...func(http.Handler) http.Handler) Router

	// Route creates a new sub-router under the specified pattern and applies the provided function to it.
	Route(pattern string, fn func(sub Router)) Router

	// Use adds the specified middlewares to the router.
	Use(middlewares ...func(http.Handler) http.Handler)

	// AddTags adds the specified tags to the router.
	AddTags(tags ...string) Router

	// AddSecurity adds the specified security scheme to the router.
	AddSecurity(securityName string, scheme *v3.SecurityScheme) Router

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
	router chi.Router

	commonPrefix     string
	commonTags       []string
	commonDeprecated bool
	commonResponses  []groupResponse
	commonSecurities map[string]*v3.SecurityScheme

	commonMiddlewares     []func(http.Handler) http.Handler
	commonHooksBeforeBind []HookBeforeBind
	commonHooksAfterBind  []HookAfterBind

	ignoreAPIDoc bool
}

func (rt *route) HttpHandler() chi.Router {
	return rt.router
}

func (r *route) Method(method string, pattern string, handler http.HandlerFunc) *OperationBuilder {
	builder := &OperationBuilder{
		route: r,
		operation: &v3.Operation{
			Summary:     method + " " + pattern,
			OperationId: genDefaultOperationID(method, pattern),
		},
		method:  method,
		pattern: pattern,
		handler: handler,

		middlewares:     r.commonMiddlewares,
		hooksBeforeBind: r.commonHooksBeforeBind,
		hooksAfterBind:  r.commonHooksAfterBind,
		ignoreAPIDoc:    r.ignoreAPIDoc,
	}
	for name, scheme := range r.commonSecurities {
		builder.AddSecurity(scheme, name)
	}
	for _, response := range r.commonResponses {
		builder.AddJSONResponse(response.code, response.model, response.description)
	}
	builder.AddTags(r.commonTags...)
	builder.SetDeprecated(r.commonDeprecated)
	return builder
}

func (r *route) Delete(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodDelete, pattern, handler)
}

func (r *route) Get(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodGet, pattern, handler)
}

func (r *route) Head(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodHead, pattern, handler)
}

func (r *route) Options(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodOptions, pattern, handler)
}

func (r *route) Patch(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodPatch, pattern, handler)
}

func (r *route) Post(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodPost, pattern, handler)
}

func (r *route) Put(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodPut, pattern, handler)
}

func (r *route) Trace(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodTrace, pattern, handler)
}

func (r *route) Mount(pattern string, sub Router) {
	subRoute, ok := sub.(*route)
	if !ok {
		return
	}

	if !r.ignoreAPIDoc {
		// Merge sub.gen into r.gen
		for oldPath, operations := range subRoute.gen.doc.Paths.PathItems {
			path := path.Join(pattern, oldPath)
			exists, ok := r.gen.doc.Paths.PathItems[path]
			if !ok {
				r.gen.doc.Paths.PathItems[path] = operations
				continue
			}
			exists.Get = operations.Get
			exists.Post = operations.Post
			exists.Put = operations.Put
			exists.Delete = operations.Delete
			exists.Patch = operations.Patch
			exists.Head = operations.Head
			exists.Options = operations.Options
			exists.Trace = operations.Trace
		}

		r.gen.doc.Tags = appendUniq(r.gen.doc.Tags, subRoute.gen.doc.Tags...)
		r.gen.doc.Security = appendUniq(r.gen.doc.Security, subRoute.gen.doc.Security...)

		for name, schema := range subRoute.gen.doc.Components.Schemas {
			r.gen.doc.Components.Schemas[name] = schema
		}
	}

	// Merge sub.router into r.router
	r.router.Mount(pattern, subRoute.router)
}

func (r *route) Group(fn func(Router)) Router {
	if fn != nil {
		fn(r)
	}
	return r
}

func (r *route) With(middlewares ...func(http.Handler) http.Handler) Router {
	return &route{
		gen:                   r.gen,
		router:                r.router,
		commonPrefix:          r.commonPrefix,
		commonTags:            r.commonTags,
		commonDeprecated:      r.commonDeprecated,
		commonResponses:       r.commonResponses,
		commonSecurities:      r.commonSecurities,
		commonMiddlewares:     append(r.commonMiddlewares, middlewares...),
		commonHooksBeforeBind: r.commonHooksBeforeBind,
		commonHooksAfterBind:  r.commonHooksAfterBind,
	}
}

func (r *route) Route(pattern string, fn func(sub Router)) Router {
	route := &route{
		gen:          NewGenerator(),
		router:       chi.NewRouter(),
		commonPrefix: pattern,

		commonTags:            r.commonTags,
		commonDeprecated:      r.commonDeprecated,
		commonResponses:       r.commonResponses,
		commonSecurities:      r.commonSecurities,
		commonMiddlewares:     r.commonMiddlewares,
		commonHooksBeforeBind: r.commonHooksBeforeBind,
		commonHooksAfterBind:  r.commonHooksAfterBind,
	}
	fn(route)
	r.Mount(pattern, route)
	return r
}

func (r *route) Use(middlewares ...func(http.Handler) http.Handler) {
	r.commonMiddlewares = append(r.commonMiddlewares, middlewares...)
}

func (r *route) AddTags(tags ...string) Router {
	r.commonTags = appendUniq(r.commonTags, tags...)

	ts := make([]*base.Tag, 0, len(tags))
	for _, tag := range tags {
		ts = append(ts, &base.Tag{Name: tag})
	}
	r.gen.doc.Tags = appendUniq(r.gen.doc.Tags, ts...)
	return r
}

func (r *route) SetDeprecated(deprecated bool) Router {
	r.commonDeprecated = deprecated
	return r
}

func (r *route) AddSecurity(securityName string, scheme *v3.SecurityScheme) Router {
	if r.commonSecurities == nil {
		r.commonSecurities = make(map[string]*v3.SecurityScheme, 1)
	}
	r.commonSecurities[securityName] = scheme
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
	desc := http.StatusText(code)
	if len(description) != 0 {
		desc = description[0]
	}

	r.commonResponses = append(r.commonResponses, groupResponse{
		code:        code,
		description: desc,
		model:       model,
	})
	return r
}
