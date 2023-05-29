# Soda

soda := [OpenAPI3.0](https://swagger.io/specification) + [fiber](https://github.com/gofiber/fiber)

> inspired on [kin-openapi3](https://github.com/getkin/kin-openapi)

## Example

```go
package main
import (
 "fmt"

 "github.com/gofiber/fiber/v2"
 "github.com/gofiber/fiber/v2/middleware/logger"
 "github.com/gofiber/fiber/v2/middleware/requestid"
 "github.com/neo-f/soda"
)

type Auth struct {
 Token string `header:"Authorization" oai:"description=some JWT Token"`
}

type ExParameters struct {
 Auth            // the embed struct will be parsed as well
 ID     int      `path:"id"      oai:"description=some id"                           `
 Q      []string `query:"q"      oai:"description=support list parameters"            `
 Limit  int      `query:"limit"  oai:"description=limit params"                      `
 Offset int8     `query:"offset" oai:"description=offset params"                     `
}

type ExBody struct {
 Body struct {
  ID     int      `json:"id"     validate:"required" oai:"description=this is the id"`
  Q      []string `json:"q"      validate:"required" oai:"description=the query to search"`
  Limit  int      `json:"limit"  validate:"required" oai:"description=limit;maximum=20"`
  Offset int      `json:"offset" validate:"required" oai:"description=offset"`
 } `body:"json"`
}

type ExampleResponse struct {
 Parameters  *ExParameters `json:"parameters"`
 RequestBody *ExBody       `json:"request_body"`
}
type ErrorResponse struct {
 Error string `json:"error"`
}

func exampleGet(c *fiber.Ctx) error {
 params := soda.GetInput[ExParameters](c)
 fmt.Println(params.Token, params.Limit, params.Offset, params.Q)
 return c.Status(200).JSON(ExampleResponse{
  Parameters: params,
 })
}

func examplePost(c *fiber.Ctx) error {
 input := soda.GetInput[ExBody](c)
 fmt.Println(input.Body.Limit, input.Body.Offset, input.Body.Q)
 return c.Status(200).JSON(ExampleResponse{
  RequestBody: input,
 })
}

func main() {
 app := soda.New(fiber.New()).
  AddJSONSpec("/openapi.json").
  AddUI("/", soda.UIStoplightElement).
  AddUI("/swagger", soda.UISwaggerUI)

 app.Fiber.Use(logger.New(), requestid.New())

 app.Get("/test/:id", exampleGet).
  SetInput(ExParameters{}).
  AddJSONResponse(200, ExampleResponse{}).
  AddJSONResponse(400, ErrorResponse{}).OK()

 app.Post("/test", examplePost).
  SetInput(ExBody{}).
  AddJSONResponse(200, ExampleResponse{}).
  AddJSONResponse(400, ErrorResponse{}).OK()
 _ = app.Fiber.Listen(":8080")
}

```

check your openapi3 spec file at <http://localhost:8080/openapi.json>

and embed openapi3 renderer

- stoplight elements: <http://localhost:8080/>
- redocly: <http://localhost:8080/redoc>
- swagger: <http://localhost:8080/swagger>
- rapidoc: <http://localhost:8080/rapidoc>

### TODO

- [ ] need add more examples to cover all the features
- [ ] support app.Group() or app.Use() maybe? need design
