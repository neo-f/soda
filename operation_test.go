package soda_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/neo-f/soda/v3"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOperations(t *testing.T) {
	Convey("Given a soda engine", t, func() {
		engine := soda.New()

		Convey("When setting up a GET operation", func() {
			type input struct {
				Authorization string `header:"authorization"`
				Page          int    `query:"page"`
			}
			type output struct {
				Authorization string `json:"authorization"`
				Page          int    `json:"page"`
			}

			builder := engine.Get("/get", func(c fiber.Ctx) error {
				o := soda.GetInput[input](c)
				return c.JSON(output{
					Authorization: o.Authorization,
					Page:          o.Page,
				})
			})
			builder.
				SetOperationID("get-demo").
				SetSummary("testing").
				SetDescription("testing").
				AddTags("hey", "jude").
				SetInput(&input{}).
				SetDeprecated(true).
				AddJSONResponse(200, &output{}).
				OK()

			Convey("Then the OpenAPI documentation should be correct", func() {
				expect := engine.OpenAPI().Paths.Find("/get").Get
				So(expect.OperationID, ShouldEqual, "get-demo")
				So(expect.Summary, ShouldEqual, "testing")
				So(expect.Description, ShouldEqual, "testing")
				So(expect.Tags, ShouldResemble, []string{"hey", "jude"})
				So(expect.Deprecated, ShouldBeTrue)
			})

			Convey("And a GET request should return the expected response", func() {
				request, _ := http.NewRequest("GET", "/get?page=1", nil)
				request.Header.Add("Authorization", "Bearer XXX")
				response, err := engine.App().Test(request)
				So(err, ShouldBeNil)
				So(response.StatusCode, ShouldEqual, 200)
				body, _ := io.ReadAll(response.Body)
				expectedBody, _ := json.Marshal(output{
					Authorization: "Bearer XXX",
					Page:          1,
				})
				So(body, ShouldResemble, expectedBody)
			})
		})

		jwt := soda.NewJWTSecurityScheme("JWT")
		apiKey := soda.NewAPIKeySecurityScheme("header", "apiKey", "apiKey")

		Convey("When setting up a POST operation", func() {
			type input struct {
				Authorization string `header:"authorization"`
				Page          int    `query:"page"`
				Body          struct {
					A string `json:"a"`
				} `body:"json"`
			}
			type output struct {
				Authorization string `json:"authorization"`
				Page          int    `json:"page"`
				A             string `json:"a"`
			}

			builder := engine.Post("/post", func(c fiber.Ctx) error {
				o := soda.GetInput[input](c)
				return c.JSON(output{
					Authorization: o.Authorization,
					Page:          o.Page,
					A:             o.Body.A,
				})
			})
			builder.
				SetOperationID("post-demo").
				SetSummary("testing").
				SetDescription("testing").
				AddTags("hey", "jude").
				SetInput(&input{}).
				SetDeprecated(true).
				AddJSONResponse(200, &output{}, "testing").
				AddSecurity("jwt", jwt).
				AddSecurity("apiKey", apiKey).
				OK()

			Convey("Then the OpenAPI documentation should be correct", func() {
				expect := engine.OpenAPI().Paths.Find("/post").Post
				So(expect.OperationID, ShouldEqual, "post-demo")
				So(expect.Summary, ShouldEqual, "testing")
				So(expect.Description, ShouldEqual, "testing")
				So(expect.Tags, ShouldResemble, []string{"hey", "jude"})
				So(expect.Deprecated, ShouldBeTrue)
				So((*expect.Security)[0], ShouldContainKey, "jwt")
				So((*expect.Security)[1], ShouldContainKey, "apiKey")
			})

			Convey("And a POST request should return the expected response", func() {
				request, _ := http.NewRequest("POST", "/post?page=1", strings.NewReader(`{"a": "test"}`))
				request.Header.Add("Content-Type", "application/json")
				request.Header.Add("Authorization", "Bearer XXX")
				response, err := engine.App().Test(request)
				So(err, ShouldBeNil)
				So(response.StatusCode, ShouldEqual, 200)
				body, _ := io.ReadAll(response.Body)
				expectedBody, _ := json.Marshal(output{
					Authorization: "Bearer XXX",
					Page:          1,
					A:             "test",
				})
				So(body, ShouldResemble, expectedBody)
			})
		})

		Convey("When setting up an operation with empty input or output", func() {
			builder := engine.Get("/action", func(c fiber.Ctx) error {
				return nil
			})
			builder.
				SetOperationID("get-demo").
				SetSummary("testing").
				SetDescription("testing").
				AddTags("hey", "jude").
				SetDeprecated(true).
				AddJSONResponse(200, nil).
				OK()

			Convey("Then the OpenAPI documentation should be correct", func() {
				generatedOperation := engine.OpenAPI().Paths.Find("/action").Get
				So(generatedOperation.OperationID, ShouldEqual, "get-demo")
				So(generatedOperation.Summary, ShouldEqual, "testing")
				So(generatedOperation.Description, ShouldEqual, "testing")
				So(generatedOperation.Tags, ShouldResemble, []string{"hey", "jude"})
				So(generatedOperation.Deprecated, ShouldBeTrue)
			})

			Convey("And a GET request should return an empty response", func() {
				request, _ := http.NewRequest("GET", "/action", nil)
				request.Header.Add("Authorization", "Bearer XXX")
				response, err := engine.App().Test(request)
				So(err, ShouldBeNil)
				So(response.StatusCode, ShouldEqual, 200)
				body, _ := io.ReadAll(response.Body)
				So(body, ShouldBeEmpty)
			})
		})

		Convey("When setting up an ignored operation", func() {
			builder := engine.Get("/action", func(c fiber.Ctx) error {
				return nil
			})
			builder.
				SetOperationID("get-demo").
				SetSummary("testing").
				SetDescription("testing").
				AddTags("hey", "jude").
				SetDeprecated(true).
				AddJSONResponse(200, nil).
				IgnoreAPIDoc(true).
				OK()

			Convey("Then the operation should not be in the OpenAPI documentation", func() {
				So(engine.OpenAPI().Paths.Find("/action"), ShouldBeNil)
			})
		})

		Convey("When setting up an operation with non-struct input", func() {
			builder := engine.Get("/action", func(c fiber.Ctx) error {
				return nil
			})

			Convey("Then it should panic", func() {
				So(func() {
					builder.SetInput("0").OK()
				}, ShouldPanic)
			})
		})

		Convey("When providing before/after hooks", func() {
			emptyHandler := func(c fiber.Ctx) error {
				return nil
			}

			type testInput struct{}

			Convey("And executing hooks before and after bind", func() {
				var before, after string
				engine := soda.New()
				engine.
					Get("/action", emptyHandler).
					SetInput(testInput{}).
					OnBeforeBind(func(ctx fiber.Ctx) error {
						before = "executed"
						return nil
					}).
					OnAfterBind(func(ctx fiber.Ctx, input any) error {
						after = "executed"
						return nil
					}).
					OK()

				request, _ := http.NewRequest("GET", "/action", nil)
				_, _ = engine.App().Test(request)
				So(before, ShouldEqual, "executed")
				So(after, ShouldEqual, "executed")
			})

			Convey("And before hook returns an error", func() {
				var after string
				engine := soda.New()
				engine.
					Get("/action", emptyHandler).
					SetInput(testInput{}).
					OnBeforeBind(func(ctx fiber.Ctx) error {
						return fiber.NewError(400, "before error")
					}).
					OnAfterBind(func(ctx fiber.Ctx, input any) error {
						after = "executed"
						return nil
					}).
					OK()

				request, _ := http.NewRequest("GET", "/action", nil)
				response, _ := engine.App().Test(request)
				So(response.StatusCode, ShouldEqual, 400)
				body, _ := io.ReadAll(response.Body)
				So(string(body), ShouldEqual, "before error")
				So(after, ShouldEqual, "")
			})

			Convey("And after hook returns an error", func() {
				var before string
				engine := soda.New()
				engine.
					Get("/action", emptyHandler).
					SetInput(testInput{}).
					OnBeforeBind(func(ctx fiber.Ctx) error {
						before = "executed"
						return nil
					}).
					OnAfterBind(func(ctx fiber.Ctx, input any) error {
						return fiber.NewError(400, "before error")
					}).
					OK()

				request, _ := http.NewRequest("GET", "/action", nil)
				response, _ := engine.App().Test(request)
				So(response.StatusCode, ShouldEqual, 400)
				body, _ := io.ReadAll(response.Body)
				So(string(body), ShouldEqual, "before error")
				So(before, ShouldEqual, "executed")
			})
		})

		Convey("When bind error occurs", func() {
			type testInput struct {
				A int `query:"a"`
			}
			engine := soda.New()
			engine.
				Get("/action", func(c fiber.Ctx) error {
					return nil
				}).
				SetInput(testInput{}).
				OK()

			Convey("Then a bind error should result in a 500 status code", func() {
				request, _ := http.NewRequest("GET", "/action?a=a", nil)
				response, _ := engine.App().Test(request)
				So(response.StatusCode, ShouldEqual, 500)
			})

			Convey("And a bind error in POST request should also result in a 500 status code", func() {
				type testInput2 struct {
					Body struct {
						A int `json:"a"`
					} `body:"json"`
				}
				engine.
					Post("/action", func(c fiber.Ctx) error {
						return nil
					}).
					SetInput(testInput2{}).
					OK()

				request, _ := http.NewRequest("POST", "/action", strings.NewReader(`{"a": "a"}`))
				request.Header.Add("Content-Type", "application/json")
				response, _ := engine.App().Test(request)
				So(response.StatusCode, ShouldEqual, 500)
			})
		})
	})
}
