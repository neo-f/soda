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

// import (
//	"testing"

//	"github.com/gofiber/fiber/v3"
//	"github.com/neo-f/soda/v3"
//	. "github.com/smartystreets/goconvey/convey"
//)

// func TestEngine(t *testing.T) {
//	Convey("Given a soda operation builder", t, func() {
//		engine := soda.New()

//		Convey("When a GET operation is added", func() {
//			handler := func(c fiber.Ctx) error {
//				return c.SendString("get handler")
//			}
//			type input struct {
//				Size  int      `query:"size"  oai:"description=Size of the response"`
//				Limit int      `query:"limit" oai:"description=Limit of the response"`
//				Q     *string  `query:"q"     oai:"description=Query string"`
//				Names []string `query:"names" oai:"description=Names"`
//				ID    string   `uri:"id"      oai:"description=id"`
//			}
//			engine.Get("/get/:id", handler).SetInput(input{}).OK()

//			op := engine.OpenAPI().Paths.Find("/get/:id").Get

//			Convey("Then the OpenAPI spec should have the GET operation", func() {
//				So(op, ShouldNotBeNil)
//			})
//			//
//			Convey("And the OpenAPI spec should have the GET operation with the correct parameters", func() {
//				So(op.Parameters, ShouldHaveLength, 5)

//				p0 := op.Parameters[0].Value
//				So(p0.Name, ShouldEqual, "size")
//				So(p0.Description, ShouldEqual, "Size of the response")
//				So(p0.Required, ShouldBeTrue)
//				So(p0.In, ShouldEqual, "query")

//				p1 := op.Parameters[1].Value
//				So(p1.Name, ShouldEqual, "limit")
//				So(p1.Description, ShouldEqual, "Limit of the response")
//				So(p1.Required, ShouldBeTrue)
//				So(p0.In, ShouldEqual, "query")

//				p2 := op.Parameters[2].Value
//				So(p2.Name, ShouldEqual, "q")
//				So(p2.Description, ShouldEqual, "Query string")
//				So(p2.Required, ShouldBeFalse)
//				So(p0.In, ShouldEqual, "query")

//				p3 := op.Parameters[3].Value
//				So(p3.Name, ShouldEqual, "names")
//				So(p3.Description, ShouldEqual, "Names")
//				So(p3.Required, ShouldBeTrue)
//				So(p0.In, ShouldEqual, "query")

//				p4 := op.Parameters[4].Value
//				So(p4.Name, ShouldEqual, "id")
//				So(p4.Description, ShouldEqual, "id")
//				So(p4.Required, ShouldBeTrue)
//				So(p4.In, ShouldEqual, "path")
//			})
//		})

//		Convey("When a POST operation is added", func() {
//			handler := func(c fiber.Ctx) error {
//				return c.SendString("post handler")
//			}

//			type input struct {
//				Body struct {
//					A string `json:"a" oai:"description=A"`
//					B int    `json:"b" oai:"description=B"`
//				} `body:"json"`
//			}

//			engine.Post("/post", handler).SetInput(input{}).OK()
//			op := engine.OpenAPI().Paths.Find("/post").Post

//			Convey("Then the OpenAPI spec should have the POST operation", func() {
//				So(op, ShouldNotBeNil)
//			})

//			Convey("And the OpenAPI spec should have the POST operation with the correct request body", func() {
//				So(op.RequestBody, ShouldNotBeNil)
//				So(op.RequestBody.Value.Required, ShouldBeTrue)

//				So(op.RequestBody.Value.Content["application/json"].Schema.Value.Properties, ShouldHaveLength, 2)
//			})
//		})
//	})
//}
