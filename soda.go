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

func (rt *Engine) OpenAPI() *v3.Document {
	return rt.gen.doc
}

func (r *Engine) ServeDocUI(pattern string, ui UIRender) *Engine {
	r.router.Get(pattern, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(ui.Render(r.gen.doc)))
	})
	return r
}

func (r *Engine) ServeSpecJSON(pattern string) *Engine {
	r.router.Get(pattern, func(w http.ResponseWriter, _ *http.Request) {
		if r.cachedSpecJSON == nil {
			r.cachedSpecJSON = r.gen.doc.RenderJSON("  ")
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(r.cachedSpecJSON)
	})
	return r
}

func (r *Engine) ServeSpecYAML(pattern string) *Engine {
	r.router.Get(pattern, func(w http.ResponseWriter, _ *http.Request) {
		if r.cachedSpecYAML == nil {
			spec, err := r.gen.doc.Render()
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
			r.cachedSpecYAML = spec
		}

		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		w.Write(r.cachedSpecYAML)
	})
	return r
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
