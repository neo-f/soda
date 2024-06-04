package soda_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v3"
	"github.com/neo-f/soda/v3"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRouter(t *testing.T) {
	Convey("Given a new fiber app and a new generator", t, func() {
		engineFiber := fiber.New()
		engine := soda.NewWith(engineFiber)

		handler := func(c fiber.Ctx) error {
			return c.SendString("Hello, World!")
		}

		Convey("When adding a GET route", func() {
			engine.Get("/hello", handler).SetOperationID("get-hello").OK()

			Convey("The route should exist in the app", func() {
				request := httptest.NewRequest("GET", "/hello", nil)
				response, err := engine.App().Test(request)
				So(err, ShouldBeNil)
				So(response.StatusCode, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When adding tags", func() {
			engine.AddTags("testTag")

			Convey("The tag should be added to the generator", func() {
				So(engine.OpenAPI().Tags, ShouldContain, &openapi3.Tag{Name: "testTag"})
			})
		})

		Convey("When setting the router as deprecated", func() {
			engine.SetDeprecated(true)
			engine.Get("/deprecated", handler).OK()

			Convey("The commonDeprecated field should be true", func() {
				So(engine.OpenAPI().Paths.Find("/deprecated").Get.Deprecated, ShouldBeTrue)
			})
		})

		Convey("When adding a security scheme", func() {
			scheme := &openapi3.SecurityScheme{
				Type: "apiKey",
			}
			engine.AddSecurity("apiKey", scheme)

			Convey("The security scheme should be added", func() {
				So(engine.OpenAPI().Components.SecuritySchemes, ShouldContainKey, "apiKey")
			})
		})

		Convey("When adding a JSON response", func() {
			engine.AddJSONResponse(http.StatusOK, map[string]string{"message": "ok"})
			engine.Get("/json", handler).OK()

			Convey("The response for status code 200 should be added", func() {
				operation := engine.OpenAPI().Paths.Find("/json").Get
				jsonSchema := openapi3.NewObjectSchema().WithAdditionalProperties(openapi3.NewStringSchema())
				response := openapi3.NewResponse().
					WithJSONSchema(jsonSchema).
					WithDescription("OK")
				So(operation.Responses.Status(200).Value, ShouldEqual, response)
			})
		})

		Convey("When setting the router to ignore API documentation", func() {
			engine.SetIgnoreAPIDoc(true)
			engine.Get("/json", handler).OK()

			Convey("The ignoreAPIDoc field should be true", func() {
				So(engine.OpenAPI().Paths.Find("/json"), ShouldBeNil)
			})
		})

		Convey("When adding a hook before bind", func() {
			var hookedValue string
			hook := func(c fiber.Ctx) error {
				hookedValue = "hooked"
				return nil
			}
			engine.OnBeforeBind(hook)
			engine.Get("/json", handler).OK()

			Convey("The hook should be executed", func() {
				request := httptest.NewRequest("GET", "/json", nil)
				_, _ = engine.App().Test(request)
				So(hookedValue, ShouldEqual, "hooked")
			})
		})

		Convey("When adding a hook after bind", func() {
			var hookedValue string
			hook := func(c fiber.Ctx, in any) error {
				hookedValue = "hooked"
				return nil
			}
			type dummyInput struct{}
			engine.OnAfterBind(hook)
			engine.Get("/json", handler).SetInput(dummyInput{}).OK()

			Convey("The hook should be executed", func() {
				request := httptest.NewRequest("GET", "/json", nil)
				_, _ = engine.App().Test(request)
				So(hookedValue, ShouldEqual, "hooked")
			})
		})

		Convey("When creating a group", func() {
			group := engine.Group("/api")
			group.AddJSONResponse(200, map[string]string{})
			group.AddJSONResponse(400, map[string]string{}, "BadRequest")
			group.AddJSONResponse(500, nil, "BadRequest")
			group.Get("/get", handler).OK()
			group.Head("/head", handler).OK()
			group.Post("/post", handler).OK()
			group.Delete("/delete", handler).OK()
			group.Put("/put", handler).OK()
			group.Patch("/patch", handler).OK()
			group.Options("/options", handler).OK()
			group.Trace("/trace", handler).OK()

			Convey("The handler should work", func() {
				methods := []string{"get", "head", "post", "delete", "put", "patch", "options", "trace"}

				for _, method := range methods {
					request := httptest.NewRequest(strings.ToUpper(method), "/api/"+method, nil)
					response, err := engine.App().Test(request)
					So(err, ShouldBeNil)
					So(response.StatusCode, ShouldEqual, http.StatusOK)

					operation := engine.OpenAPI().Paths.Find("/api/" + method).GetOperation(strings.ToUpper(method))
					So(operation, ShouldNotBeNil)
					So(operation.Responses.Value("200"), ShouldNotBeNil)
					So(operation.Responses.Value("400"), ShouldNotBeNil)
					So(*operation.Responses.Value("400").Value.Description, ShouldEqual, "BadRequest")
					So(operation.Responses.Value("500"), ShouldNotBeNil)
				}
			})

			Convey("The Operation should Be Added", func() {
				operation := engine.OpenAPI().Paths.Find("/api/get").Get
				So(operation, ShouldNotBeNil)
				So(operation.Responses.Value("200"), ShouldNotBeNil)
				So(operation.Responses.Value("400"), ShouldNotBeNil)
				So(*operation.Responses.Value("400").Value.Description, ShouldEqual, "BadRequest")
				So(operation.Responses.Value("500"), ShouldNotBeNil)
			})
		})
	})
}
