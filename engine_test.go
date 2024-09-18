package soda_test

import (
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/neo-f/soda/v3"
	. "github.com/smartystreets/goconvey/convey"
)

type mockUIRender struct{}

func (m *mockUIRender) Render(doc *openapi3.T) string {
	return "Rendered"
}

func TestEngine(t *testing.T) {
	Convey("Given a new soda Engine", t, func() {
		engine := soda.New()

		Convey("The engine should not be nil", func() {
			So(engine, ShouldNotBeNil)
		})

		Convey("The OpenAPI should not be nil", func() {
			So(engine.OpenAPI(), ShouldNotBeNil)
		})

		Convey("The App should not be nil", func() {
			So(engine.App(), ShouldNotBeNil)
		})

		Convey("When serving the documentation UI", func() {
			engine.ServeDocUI("/doc", &mockUIRender{})
			engine.ServeDocUI("/elements", soda.UIStoplightElement)

			Convey("The response should have status code 200", func() {
				req := httptest.NewRequest("GET", "/doc", nil)
				w := httptest.NewRecorder()
				engine.App().ServeHTTP(w, req)
				So(w.Code, ShouldEqual, 200)

				req = httptest.NewRequest("GET", "/elements", nil)
				w = httptest.NewRecorder()
				engine.App().ServeHTTP(w, req)
				So(w.Code, ShouldEqual, 200)
			})
		})

		Convey("When serving the specification JSON", func() {
			engine.ServeSpecJSON("/spec.json")

			req := httptest.NewRequest("GET", "/spec.json", nil)
			w := httptest.NewRecorder()
			engine.App().ServeHTTP(w, req)

			Convey("The response should have status code 200", func() {
				So(w.Code, ShouldEqual, 200)
			})
		})

		Convey("When serving the specification YAML", func() {
			engine.ServeSpecYAML("/spec.yaml")

			req := httptest.NewRequest("GET", "/spec.yaml", nil)
			w := httptest.NewRecorder()
			engine.App().ServeHTTP(w, req)

			Convey("The response should have status code 200", func() {
				So(w.Code, ShouldEqual, 200)
			})
		})

		Convey("When creating a new engine with a custom fiber App", func() {
			app := gin.New()
			newEngine := soda.NewWith(app)

			Convey("The new engine should not be nil", func() {
				So(newEngine, ShouldNotBeNil)
			})

			Convey("The new engine's app should equal the custom app", func() {
				So(newEngine.App(), ShouldEqual, app)
			})
		})
	})
}
