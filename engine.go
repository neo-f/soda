package soda

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

type Engine struct {
	*Router
	app            *gin.Engine
	cachedSpecYAML []byte
	cachedSpecJSON []byte
}

func (e *Engine) OpenAPI() *openapi3.T {
	return e.gen.doc
}

func (e *Engine) App() *gin.Engine {
	return e.app
}

func (e *Engine) ServeDocUI(pattern string, ui UIRender) *Engine {
	e.app.GET(pattern, func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", []byte(ui.Render(e.gen.doc)))
	})
	return e
}

func (e *Engine) ServeSpecJSON(pattern string) *Engine {
	if e.cachedSpecJSON == nil {
		e.cachedSpecJSON, _ = e.gen.doc.MarshalJSON()
	}
	e.app.GET(pattern, func(c *gin.Context) {
		c.Data(200, "application/json; charset=utf-8", e.cachedSpecJSON)
	})
	return e
}

func (e *Engine) ServeSpecYAML(pattern string) *Engine {
	e.app.GET(pattern, func(c *gin.Context) {
		c.Data(200, "text/yaml; charset=utf-8", e.cachedSpecYAML)
	})
	return e
}

func New() *Engine {
	return NewWith(gin.New())
}

func NewWith(app *gin.Engine) *Engine {
	return &Engine{
		app: app,
		Router: &Router{
			gen: NewGenerator(),
			Raw: app,
		},
	}
}
