package soda_test

import (
	"net/http/httptest"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v3"
	"github.com/neo-f/soda/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockUIRender struct{}

func (m *mockUIRender) Render(doc *openapi3.T) string {
	return "Rendered"
}

var _ = Describe("Engine", func() {
	var engine *soda.Engine

	BeforeEach(func() {
		engine = soda.New()
		Expect(engine).ToNot(BeNil())
	})

	Describe("OpenAPI", func() {
		It("should not be nil", func() {
			Expect(engine.OpenAPI()).ToNot(BeNil())
		})
	})

	Describe("App", func() {
		It("should not be nil", func() {
			Expect(engine.App()).ToNot(BeNil())
		})
	})

	Describe("ServeDocUI", func() {
		It("should respond with status code 200", func() {
			engine.ServeDocUI("/doc", &mockUIRender{})
			req := httptest.NewRequest("GET", "/doc", nil)
			resp, _ := engine.App().Test(req)
			Expect(resp.StatusCode).To(Equal(200))
		})
	})

	Describe("ServeSpecJSON", func() {
		It("should respond with status code 200", func() {
			engine.ServeSpecJSON("/spec.json")
			req := httptest.NewRequest("GET", "/spec.json", nil)
			resp, _ := engine.App().Test(req)
			Expect(resp.StatusCode).To(Equal(200))
		})
	})

	Describe("ServeSpecYAML", func() {
		It("should respond with status code 200", func() {
			engine.ServeSpecYAML("/spec.yaml")
			req := httptest.NewRequest("GET", "/spec.yaml", nil)
			resp, _ := engine.App().Test(req)
			Expect(resp.StatusCode).To(Equal(200))
		})
	})

	Describe("NewWith", func() {
		It("should return a new engine with a custom fiber App", func() {
			app := fiber.New()
			engine := soda.NewWith(app)
			Expect(engine).ToNot(BeNil())
			Expect(engine.App()).To(Equal(app))
		})
	})
})
