package soda

import (
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type groupResponse struct {
	code        int
	description string
	model       any
}

type Route struct {
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

	cachedSpecYAML []byte
	cachedSpecJSON []byte
}

func (rt *Route) HttpHandler() chi.Router {
	return rt.router
}

func (rt *Route) OpenAPI() *v3.Document {
	return rt.gen.doc
}

func (r *Route) Method(method string, pattern string, handler http.HandlerFunc) *OperationBuilder {
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

func (r *Route) Delete(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodDelete, pattern, handler)
}

func (r *Route) Get(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodGet, pattern, handler)
}

func (r *Route) Head(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodHead, pattern, handler)
}

func (r *Route) Options(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodOptions, pattern, handler)
}

func (r *Route) Patch(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodPatch, pattern, handler)
}

func (r *Route) Post(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodPost, pattern, handler)
}

func (r *Route) Put(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodPut, pattern, handler)
}

func (r *Route) Trace(pattern string, handler http.HandlerFunc) *OperationBuilder {
	return r.Method(http.MethodTrace, pattern, handler)
}

func (r *Route) Mount(pattern string, sub *Route) {
	// Merge sub.gen into r.gen
	for oldPath, operations := range sub.gen.doc.Paths.PathItems {
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

	appendUniqBy(sameTag, r.gen.doc.Tags, sub.gen.doc.Tags...)
	appendUniqBy(sameSecurityRequirements, r.gen.doc.Security, sub.gen.doc.Security...)

	for name, schema := range sub.gen.doc.Components.Schemas {
		r.gen.doc.Components.Schemas[name] = schema
	}

	// Merge sub.router into r.router
	r.router.Mount(pattern, sub.router)
}

func (r *Route) Group(fn func(*Route)) *Route {
	if fn != nil {
		fn(r)
	}
	return r
}

func (r *Route) With(middlewares ...func(http.Handler) http.Handler) *Route {
	return &Route{
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

func (r *Route) Route(pattern string, fn func(sub *Route)) *Route {
	route := &Route{
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

func (r *Route) Use(middlewares ...func(http.Handler) http.Handler) {
	r.commonMiddlewares = append(r.commonMiddlewares, middlewares...)
}

func (r *Route) AddDocUI(pattern string, ui UIRender) *Route {
	r.router.Get(pattern, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(ui.Render(r.gen.doc)))
	})
	return r
}

// AddJSONSpec adds the OpenAPI spec at the given path in JSON format.
func (r *Route) AddJSONSpec(pattern string) *Route {
	r.router.Get(pattern, func(w http.ResponseWriter, _ *http.Request) {
		if r.cachedSpecJSON == nil {
			r.cachedSpecJSON = r.gen.doc.RenderJSON("  ")
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(r.cachedSpecJSON)
	})
	return r
}

// AddYAMLSpec adds the OpenAPI spec at the given path in YAML format.
func (r *Route) AddYAMLSpec(pattern string) *Route {
	r.router.Get(pattern, func(w http.ResponseWriter, _ *http.Request) {
		if r.cachedSpecYAML == nil {
			spec, err := r.gen.doc.Render()
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
			r.cachedSpecYAML = spec
		}

		w.Header().Set("Content-Type", "test/yaml; charset=utf-8")
		w.Write(r.cachedSpecYAML)
	})
	return r
}

// AddTags adds tags to the operation.
func (r *Route) AddTags(tags ...string) *Route {
	appendUniqBy(sameVal, r.commonTags, tags...)

	ts := make([]*base.Tag, 0, len(tags))
	for _, tag := range tags {
		ts = append(ts, &base.Tag{Name: tag})
	}
	appendUniqBy(sameTag, r.gen.doc.Tags, ts...)
	return r
}

// SetDeprecated sets the deprecated flag for the group.
func (r *Route) SetDeprecated(deprecated bool) *Route {
	r.commonDeprecated = deprecated
	return r
}

func (r *Route) AddSecurity(securityName string, scheme *v3.SecurityScheme) *Route {
	if r.commonSecurities == nil {
		r.commonSecurities = make(map[string]*v3.SecurityScheme, 1)
	}
	r.commonSecurities[securityName] = scheme
	return r
}

// OnAfterBind adds a hook to be executed after the operation is bound.
func (r *Route) OnAfterBind(hook HookAfterBind) *Route {
	r.commonHooksAfterBind = append(r.commonHooksAfterBind, hook)
	return r
}

// OnBeforeBind adds a hook to be executed after the operation is bound.
func (r *Route) OnBeforeBind(hook HookBeforeBind) *Route {
	r.commonHooksBeforeBind = append(r.commonHooksBeforeBind, hook)
	return r
}

func (r *Route) AddJSONResponse(code int, model any, description ...string) *Route {
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

func New() *Route {
	return &Route{
		gen:    NewGenerator(),
		router: chi.NewRouter(),
	}
}

func NewWith(router chi.Router) *Route {
	return &Route{
		gen:    NewGenerator(),
		router: router,
	}
}
