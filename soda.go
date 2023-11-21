package soda

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type groupResponse struct {
	code        int
	description string
	model       any
}

type Engine struct {
	*route

	cachedSpecYAML []byte
	cachedSpecJSON []byte
}

func (e *Engine) OpenAPI() *v3.Document {
	return e.gen.doc
}

func (e *Engine) ServeDocUI(pattern string, ui UIRender) *Engine {
	e.router.Get(pattern, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(ui.Render(e.gen.doc)))
	})
	return e
}

func (e *Engine) ServeSpecJSON(pattern string) *Engine {
	if e.cachedSpecJSON == nil {
		e.cachedSpecJSON = e.gen.doc.RenderJSON("  ")
	}
	e.router.Get(pattern, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(e.cachedSpecJSON)
	})
	return e
}

func (e *Engine) ServeSpecYAML(pattern string) *Engine {
	if e.cachedSpecYAML == nil {
		spec, _ := e.gen.doc.Render()
		e.cachedSpecYAML = spec
	}
	e.router.Get(pattern, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		_, _ = w.Write(e.cachedSpecYAML)
	})
	return e
}

func New() *Engine {
	return &Engine{
		route: &route{
			gen:    NewGenerator(),
			router: chi.NewRouter(),
		},
	}
}

func NewWith(router chi.Router) *Engine {
	return &Engine{
		route: &route{
			gen:    NewGenerator(),
			router: router,
		},
	}
}
