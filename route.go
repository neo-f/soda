package soda

import (
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type Router interface {
	HttpHandler() chi.Router
	Method(method string, pattern string, handler http.HandlerFunc) *OperationBuilder
	Delete(pattern string, handler http.HandlerFunc) *OperationBuilder
	Get(pattern string, handler http.HandlerFunc) *OperationBuilder
	Head(pattern string, handler http.HandlerFunc) *OperationBuilder
	Options(pattern string, handler http.HandlerFunc) *OperationBuilder
	Patch(pattern string, handler http.HandlerFunc) *OperationBuilder
	Post(pattern string, handler http.HandlerFunc) *OperationBuilder
	Put(pattern string, handler http.HandlerFunc) *OperationBuilder
	Trace(pattern string, handler http.HandlerFunc) *OperationBuilder

	Mount(pattern string, sub Router)
	Group(fn func(Router)) Router
	With(middlewares ...func(http.Handler) http.Handler) Router
	Route(pattern string, fn func(sub Router)) Router
	Use(middlewares ...func(http.Handler) http.Handler)

	AddTags(tags ...string) Router
	AddSecurity(securityName string, scheme *v3.SecurityScheme) Router
	AddJSONResponse(code int, model any, description ...string) Router
	SetDeprecated(deprecated bool) Router
	OnAfterBind(hook HookAfterBind) Router
	OnBeforeBind(hook HookBeforeBind) Router
}

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
	}
	for name, scheme := range r.commonSecurities {
		builder.AddSecurity(scheme, name)
	}
	for _, response := range r.commonResponses {
		builder.AddJSONResponse(response.code, response.model, response.description)
	}
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

	appendUniqBy(sameTag, r.gen.doc.Tags, subRoute.gen.doc.Tags...)
	appendUniqBy(sameSecurityRequirements, r.gen.doc.Security, subRoute.gen.doc.Security...)

	for name, schema := range subRoute.gen.doc.Components.Schemas {
		r.gen.doc.Components.Schemas[name] = schema
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

// AddTags adds tags to the operation.
func (r *route) AddTags(tags ...string) Router {
	appendUniqBy(sameVal, r.commonTags, tags...)

	ts := make([]*base.Tag, 0, len(tags))
	for _, tag := range tags {
		ts = append(ts, &base.Tag{Name: tag})
	}
	appendUniqBy(sameTag, r.gen.doc.Tags, ts...)
	return r
}

// SetDeprecated sets the deprecated flag for the group.
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

// OnAfterBind adds a hook to be executed after the operation is bound.
func (r *route) OnAfterBind(hook HookAfterBind) Router {
	r.commonHooksAfterBind = append(r.commonHooksAfterBind, hook)
	return r
}

// OnBeforeBind adds a hook to be executed after the operation is bound.
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
