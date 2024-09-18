package soda

import (
	"maps"
	"net/http"
	"path"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

type Router struct {
	Raw gin.IRouter
	gen *Generator

	commonPrefix     string
	commonTags       []string
	commonDeprecated bool
	commonResponses  map[int]*openapi3.Response
	commonSecurities openapi3.SecurityRequirements

	commonHooksBeforeBind []HookBeforeBind
	commonHooksAfterBind  []HookAfterBind

	ignoreAPIDoc bool
}

func (r *Router) createOperationBuilder(method string, pattern, patternFull string, handlers ...gin.HandlerFunc) *OperationBuilder {
	return &OperationBuilder{
		route: r,
		operation: &openapi3.Operation{
			Summary:     method + " " + patternFull,
			OperationID: genDefaultOperationID(method, patternFull),
			Security:    &r.commonSecurities,
		},
		method:      method,
		patternFull: patternFull,
		pattern:     pattern,
		handlers:    handlers,

		hooksBeforeBind: r.commonHooksBeforeBind,
		hooksAfterBind:  r.commonHooksAfterBind,
		ignoreAPIDoc:    r.ignoreAPIDoc,
	}
}

func (r *Router) Add(method string, pattern string, handlers ...gin.HandlerFunc) *OperationBuilder {
	patternFull := path.Join(r.commonPrefix, pattern)
	builder := r.createOperationBuilder(method, pattern, patternFull, handlers...)
	for code, resp := range r.commonResponses {
		builder.operation.AddResponse(code, resp)
	}
	builder.AddTags(r.commonTags...)
	builder.SetDeprecated(r.commonDeprecated)
	return builder
}

func (r *Router) Delete(pattern string, handlers ...gin.HandlerFunc) *OperationBuilder {
	return r.Add(http.MethodDelete, pattern, handlers...)
}

func (r *Router) Get(pattern string, handlers ...gin.HandlerFunc) *OperationBuilder {
	return r.Add(http.MethodGet, pattern, handlers...)
}

func (r *Router) Head(pattern string, handlers ...gin.HandlerFunc) *OperationBuilder {
	return r.Add(http.MethodHead, pattern, handlers...)
}

func (r *Router) Options(pattern string, handlers ...gin.HandlerFunc) *OperationBuilder {
	return r.Add(http.MethodOptions, pattern, handlers...)
}

func (r *Router) Patch(pattern string, handlers ...gin.HandlerFunc) *OperationBuilder {
	return r.Add(http.MethodPatch, pattern, handlers...)
}

func (r *Router) Post(pattern string, handlers ...gin.HandlerFunc) *OperationBuilder {
	return r.Add(http.MethodPost, pattern, handlers...)
}

func (r *Router) Put(pattern string, handlers ...gin.HandlerFunc) *OperationBuilder {
	return r.Add(http.MethodPut, pattern, handlers...)
}

func (r *Router) Trace(pattern string, handlers ...gin.HandlerFunc) *OperationBuilder {
	return r.Add(http.MethodTrace, pattern, handlers...)
}

func (r *Router) AddTags(tags ...string) *Router {
	r.commonTags = append(r.commonTags, tags...)

	for _, tag := range tags {
		r.gen.doc.Tags = append(r.gen.doc.Tags, &openapi3.Tag{
			Name: tag,
		})
	}
	return r
}

func (r *Router) SetDeprecated(deprecated bool) *Router {
	r.commonDeprecated = deprecated
	return r
}

func (r *Router) AddSecurity(securityName string, scheme *openapi3.SecurityScheme) *Router {
	r.gen.doc.Components.SecuritySchemes[securityName] = &openapi3.SecuritySchemeRef{Value: scheme}
	r.commonSecurities = append(
		r.commonSecurities,
		openapi3.SecurityRequirement{securityName: nil},
	)
	return r
}

// SetIgnoreAPIDoc implements Router.
func (r *Router) SetIgnoreAPIDoc(ignore bool) *Router {
	r.ignoreAPIDoc = ignore
	return r
}

func (r *Router) OnAfterBind(hook HookAfterBind) *Router {
	r.commonHooksAfterBind = append(r.commonHooksAfterBind, hook)
	return r
}

func (r *Router) OnBeforeBind(hook HookBeforeBind) *Router {
	r.commonHooksBeforeBind = append(r.commonHooksBeforeBind, hook)
	return r
}

func (r *Router) AddJSONResponse(code int, model any, description ...string) *Router {
	desc := http.StatusText(code)
	if len(description) > 0 {
		desc = description[0]
	}

	if r.commonResponses == nil {
		r.commonResponses = make(map[int]*openapi3.Response)
	}
	if model == nil {
		r.commonResponses[code] = openapi3.NewResponse().WithDescription(desc)
		return r
	}
	resp := r.gen.GenerateResponse(code, model, "application/json", desc)
	r.commonResponses[code] = resp
	return r
}

func (r *Router) Group(prefix string, handlers ...gin.HandlerFunc) *Router {
	return &Router{
		gen:                   r.gen,
		Raw:                   r.Raw.Group(prefix, handlers...),
		commonPrefix:          path.Join(r.commonPrefix, prefix),
		commonTags:            r.commonTags,
		commonDeprecated:      r.commonDeprecated,
		commonResponses:       maps.Clone(r.commonResponses),
		commonSecurities:      r.commonSecurities,
		commonHooksBeforeBind: r.commonHooksBeforeBind,
		commonHooksAfterBind:  r.commonHooksAfterBind,
		ignoreAPIDoc:          r.ignoreAPIDoc,
	}
}
