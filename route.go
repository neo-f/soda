package soda

import (
	"net/http"
	"path"
	"reflect"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
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
	Mount(pattern string, sub *Engine)

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
	commonDeprecated *bool
	commonResponses  *orderedmap.Map[string, *v3.Response]
	commonSecurities []*base.SecurityRequirement

	commonHooksBeforeBind []HookBeforeBind
	commonHooksAfterBind  []HookAfterBind

	ignoreAPIDoc bool
}

func (r *route) HttpHandler() chi.Router {
	return r.router
}

func (r *route) Method(method string, pattern string, handler http.HandlerFunc) *OperationBuilder {
	builder := &OperationBuilder{
		route: r,
		operation: &v3.Operation{
			Summary:     method + " " + pattern,
			OperationId: genDefaultOperationID(method, pattern),
			Security:    r.commonSecurities,
			Tags:        r.commonTags,
		},
		method:  method,
		pattern: pattern,
		handler: handler,

		hooksBeforeBind: r.commonHooksBeforeBind,
		hooksAfterBind:  r.commonHooksAfterBind,
		ignoreAPIDoc:    r.ignoreAPIDoc,
	}

	mergeMap(builder.operation.Responses.Codes, r.commonResponses)
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

func (r *route) copyOperation(src *v3.Operation) *v3.Operation {
	if src == nil {
		return nil
	}
	dst := *src
	dst.Deprecated = r.commonDeprecated

	dst.Tags = append(dst.Tags, r.commonTags...)
	dst.Tags = uniqBy(dst.Tags, func(item string) string { return item })

	dst.Security = append(dst.Security, r.commonSecurities...)
	dst.Security = uniqBy(dst.Security, sameSecurityRequirement)

	mergeMap(dst.Responses.Codes, r.commonResponses)
	return &dst
}

func (r *route) Mount(pattern string, sub *Engine) {
	if !r.ignoreAPIDoc {
		// Merge sub.gen into r.gen
		for pair := sub.gen.doc.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
			oldPath := pair.Key()
			operations := pair.Value()
			path := path.Join(pattern, oldPath)
			pathItem, ok := r.gen.doc.Paths.PathItems.Get(path)
			if !ok {
				pathItem = &v3.PathItem{}
			}
			pathItem.Get = r.copyOperation(operations.Get)
			pathItem.Post = r.copyOperation(operations.Post)
			pathItem.Put = r.copyOperation(operations.Put)
			pathItem.Delete = r.copyOperation(operations.Delete)
			pathItem.Patch = r.copyOperation(operations.Patch)
			pathItem.Head = r.copyOperation(operations.Head)
			pathItem.Options = r.copyOperation(operations.Options)
			pathItem.Trace = r.copyOperation(operations.Trace)
			r.gen.doc.Paths.PathItems.Set(path, pathItem)
		}

		r.gen.doc.Tags = append(r.gen.doc.Tags, sub.gen.doc.Tags...)
		r.gen.doc.Tags = uniqBy(r.gen.doc.Tags, func(item *base.Tag) string { return item.Name })

		r.gen.doc.Security = append(r.gen.doc.Security, sub.gen.doc.Security...)
		r.gen.doc.Security = uniqBy(r.gen.doc.Security, sameSecurityRequirement)

		mergeMap(r.gen.doc.Components.Schemas, sub.gen.doc.Components.Schemas)
		mergeMap(r.gen.doc.Components.Responses, sub.gen.doc.Components.Responses)
		mergeMap(r.gen.doc.Components.Parameters, sub.gen.doc.Components.Parameters)
		mergeMap(r.gen.doc.Components.Examples, sub.gen.doc.Components.Examples)
		mergeMap(r.gen.doc.Components.RequestBodies, sub.gen.doc.Components.RequestBodies)
		mergeMap(r.gen.doc.Components.Headers, sub.gen.doc.Components.Headers)
		mergeMap(r.gen.doc.Components.SecuritySchemes, sub.gen.doc.Components.SecuritySchemes)
		mergeMap(r.gen.doc.Components.Links, sub.gen.doc.Components.Links)
		mergeMap(r.gen.doc.Components.Callbacks, sub.gen.doc.Components.Callbacks)
		mergeMap(r.gen.doc.Components.Extensions, sub.gen.doc.Components.Extensions)
	}

	// Merge sub.router into r.router
	r.router.Mount(pattern, sub.router)
}

func (r *route) Group(fn func(Router)) Router {
	if fn != nil {
		fn(r)
	}
	return r
}

func (r *route) Route(pattern string, fn func(sub Router)) Router {
	newRouter := &Engine{
		route: &route{
			gen:          NewGenerator(),
			router:       chi.NewRouter(),
			commonPrefix: pattern,

			commonTags:            r.commonTags,
			commonDeprecated:      r.commonDeprecated,
			commonResponses:       r.commonResponses,
			commonSecurities:      r.commonSecurities,
			commonHooksBeforeBind: r.commonHooksBeforeBind,
			commonHooksAfterBind:  r.commonHooksAfterBind,
		},
	}
	fn(newRouter)
	r.Mount(pattern, newRouter)
	return r
}

func (r *route) Use(middlewares ...func(http.Handler) http.Handler) {
	r.router.Use(middlewares...)
}

func (r *route) With(middlewares ...func(http.Handler) http.Handler) Router {
	return &route{
		gen:                   r.gen,
		router:                r.router.With(middlewares...),
		commonPrefix:          r.commonPrefix,
		commonTags:            r.commonTags,
		commonDeprecated:      r.commonDeprecated,
		commonResponses:       r.commonResponses,
		commonSecurities:      r.commonSecurities,
		commonHooksBeforeBind: r.commonHooksBeforeBind,
		commonHooksAfterBind:  r.commonHooksAfterBind,
	}
}

func (r *route) AddTags(tags ...string) Router {
	r.commonTags = append(r.commonTags, tags...)
	r.commonTags = uniqBy(r.commonTags, func(item string) string { return item })

	ts := make([]*base.Tag, 0, len(tags))
	for _, tag := range tags {
		ts = append(ts, &base.Tag{Name: tag})
	}
	r.gen.doc.Tags = append(r.gen.doc.Tags, ts...)
	r.gen.doc.Tags = uniqBy(r.gen.doc.Tags, func(item *base.Tag) string { return item.Name })
	return r
}

func (r *route) SetDeprecated(deprecated bool) Router {
	r.commonDeprecated = ptr(deprecated)
	return r
}

func (r *route) AddSecurity(securityName string, scheme *v3.SecurityScheme) Router {
	r.gen.doc.Components.SecuritySchemes.Set(securityName, scheme)
	r.commonSecurities = append(r.commonSecurities, &base.SecurityRequirement{
		Requirements: orderedmap.New[string, []string](),
	})
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
		r.commonResponses = orderedmap.New[string, *v3.Response]()
	}
	resp := r.gen.GenerateResponse(code, "json", reflect.TypeOf(model), description...)
	r.commonResponses.Set(strconv.Itoa(code), resp)
	return r
}

func (r *route) AddPlainTextResponse(code int, description ...string) Router {
	if r.commonResponses == nil {
		r.commonResponses = orderedmap.New[string, *v3.Response]()
	}
	resp := r.gen.GenerateResponse(code, "text", nil, description...)
	r.commonResponses.Set(strconv.Itoa(code), resp)
	return r
}
