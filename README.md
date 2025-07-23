# Soda

[![codecov](https://codecov.io/github/neo-f/soda/branch/master/graph/badge.svg?token=uYHY9DCbNe)](https://codecov.io/github/neo-f/soda)

**Soda** is a powerful Go library that seamlessly integrates [OpenAPI 3](https://swagger.io/specification) documentation with the [Fiber](https://github.com/gofiber/fiber) web framework. It automatically generates comprehensive API documentation from your Go structs and route definitions, eliminating the need for manual OpenAPI specification writing.

## 🚀 Features

- **Automatic OpenAPI 3 Generation**: Generate complete OpenAPI specifications from Go structs
- **Fiber Integration**: Built specifically for the Fiber web framework
- **Zero Configuration**: Works out of the box with sensible defaults
- **Type Safety**: Leverage Go's type system for API documentation
- **Interactive Documentation**: Built-in support for Swagger UI, ReDoc, RapiDoc, and Stoplight Elements
- **Request/Response Binding**: Automatic binding of HTTP requests to Go structs
- **Security Schemes**: Support for JWT, API Key, and custom security schemes
- **Validation**: Built-in validation using struct tags
- **Extensible**: Custom hooks and middleware support

## 📦 Installation

```bash
go get github.com/neo-f/soda/v3
```

## 🏁 Quick Start

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/neo-f/soda/v3"
)

// Define your request/response structs
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

type UserResponse struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // Create a new soda engine
    engine := soda.New()

    // Set API info
    engine.OpenAPI().Info.Title = "User API"
    engine.OpenAPI().Info.Version = "1.0.0"
    engine.OpenAPI().Info.Description = "A simple user management API"

    // Define routes with automatic OpenAPI documentation
    engine.Post("/users", createUser).
        SetSummary("Create a new user").
        SetDescription("Creates a new user account with the provided information").
        SetInput(CreateUserRequest{}).
        AddJSONResponse(201, UserResponse{}, "User created successfully").
        AddJSONResponse(400, nil, "Invalid input data").
        AddJSONResponse(500, nil, "Internal server error").
        OK()

    // Serve interactive documentation
    engine.ServeDocUI("/docs", soda.UISwaggerUI)
    engine.ServeSpecJSON("/openapi.json")
    engine.ServeSpecYAML("/openapi.yaml")

    // Start the server
    engine.App().Listen(":3000")
}

func createUser(c *fiber.Ctx) error {
    input := soda.GetInput[CreateUserRequest](c)

    // Your business logic here
    response := UserResponse{
        ID:    1,
        Name:  input.Name,
        Email: input.Email,
    }

    return c.Status(201).JSON(response)
}
```

## 📖 Documentation

### Struct Tags

Soda uses struct tags to generate OpenAPI specifications. Here are the supported tags:

#### Basic Tags

```go
type Example struct {
    // Path parameters
    ID int `path:"id"`
    // Query parameters
    Search string `query:"search"`
    // Header parameters
    Authorization string `header:"Authorization"`
    // Cookie parameters
    Session string `cookie:"session"`
    // Request body
    Profile Profile `body:"application/json"`
    // OpenAPI specific tags
    Name string `json:"name" oai:"minLength=3,maxLength=50,description=The user's full name"`
    Age  int    `json:"age" oai:"minimum=18,maximum=120,description=User's age in years"`
}
```

#### Validation Tags

```go
type ValidatedRequest struct {
    Email    string `json:"email" oai:"format=email,example=user@example.com"`
    Phone    string `json:"phone" oai:"format=phone,example=+1234567890"`
    Website  string `json:"website" oai:"format=uri,example=https://example.com"`
    // Enum values
    Role     string `json:"role" oai:"enum=admin,user,guest"`
    // String constraints
    Username string `json:"username" oai:"minLength=3,maxLength=20,pattern=^[a-zA-Z0-9]+$"`
    // Numeric constraints
    Score    int    `json:"score" oai:"minimum=0,maximum=100,multipleOf=5"`
    // Array constraints
    Tags     []string `json:"tags" oai:"minItems=1,maxItems=10,uniqueItems=true"`
    // Date/time
    Birthday time.Time `json:"birthday" oai:"format=date-time"`
}
```

### Route Definition

#### Basic Routes

```go
// GET /users/:id
engine.Get("/users/:id", getUser).
    SetSummary("Get user by ID").
    AddJSONResponse(200, UserResponse{}).
    AddJSONResponse(404, nil, "User not found").OK()

// POST /users
engine.Post("/users", createUser).
    SetInput(CreateUserRequest{}).
    AddJSONResponse(201, UserResponse{}).OK()

// PUT /users/:id
engine.Put("/users/:id", updateUser).
    SetInput(UpdateUserRequest{}).
    AddJSONResponse(200, UserResponse{}).OK()

// DELETE /users/:id
engine.Delete("/users/:id", deleteUser).
    AddJSONResponse(204, nil).OK()
```

#### Route Groups

```go
// Create a route group with common settings
api := engine.Group("/api/v1")

// Add common tags
api.AddTags("User Management")

// Add common responses
api.AddJSONResponse(400, ErrorResponse{}, "Bad Request")
api.AddJSONResponse(500, ErrorResponse{}, "Internal Server Error")

// Define routes within the group
api.Get("/users", listUsers).OK()
api.Post("/users", createUser).OK()
api.Get("/users/:id", getUser).OK()
api.Put("/users/:id", updateUser).OK()
api.Delete("/users/:id", deleteUser).OK()
```

#### Security

```go
// Add JWT security to the entire API
engine.AddSecurity("jwt", soda.NewJWTSecurityScheme("JWT authentication"))
// Add API key security
engine.AddSecurity("api-key", soda.NewAPIKeySecurityScheme("header", "X-API-Key", "API key authentication"))
// Apply security to specific routes
engine.Post("/admin/users", createAdminUser).
    AddSecurity("jwt", soda.NewJWTSecurityScheme()).
    AddJSONResponse(201, AdminUserResponse{}).OK()
```

### Advanced Features

#### Custom Hooks

```go
engine.Post("/users", createUser).
    OnBeforeBind(func(c *fiber.Ctx) error {
        // Validate request before binding
        contentType := c.Get("Content-Type")
        if contentType != "application/json" {
            return fiber.ErrBadRequest
        }
        return nil
    }).
    OnAfterBind(func(c *fiber.Ctx, input any) error {
        // Process input after binding
        userInput := input.(*CreateUserRequest)
        userInput.Email = strings.ToLower(userInput.Email)
        return nil
    })
```

#### Custom Schema Generation

```go
// Implement custom JSON schema generation
type CustomType struct {
    Value string
}

func (c CustomType) JSONSchema(doc *openapi3.T) *openapi3.SchemaRef {
    return openapi3.NewStringSchema().
        WithFormat("custom").
        WithDescription("Custom type with special validation").
        NewRef()
}
```

## 🎯 Usage Examples

### Complete RESTful API

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/neo-f/soda/v3"
)

type (
    CreateTodoRequest struct {
        Title       string `json:"title" oai:"minLength=1,maxLength=100"`
        Description string `json:"description" oai:"maxLength=500"`
        Completed   bool   `json:"completed"`
    }

    UpdateTodoRequest struct {
        Title       *string `json:"title,omitempty" oai:"minLength=1,maxLength=100"`
        Description *string `json:"description,omitempty" oai:"maxLength=500"`
        Completed   *bool   `json:"completed,omitempty"`
    }

    TodoResponse struct {
        ID          int    `json:"id"`
        Title       string `json:"title"`
        Description string `json:"description"`
        Completed   bool   `json:"completed"`
        CreatedAt   string `json:"created_at" format:"date-time"`
        UpdatedAt   string `json:"updated_at" format:"date-time"`
    }

    TodoListResponse struct {
        Todos []TodoResponse `json:"todos"`
        Total int            `json:"total"`
    }

    ErrorResponse struct {
        Error string `json:"error"`
    }
)

func main() {
    engine := soda.New()

    // Configure API
    engine.OpenAPI().Info = &openapi3.Info{
        Title:       "Todo API",
        Version:     "1.0.0",
        Description: "A comprehensive todo management API",
    }

    // Add common security
    engine.AddSecurity("bearer", soda.NewJWTSecurityScheme("JWT Bearer Token"))

    // Add common responses
    engine.AddJSONResponse(400, ErrorResponse{}).
        AddJSONResponse(401, ErrorResponse{}).
        AddJSONResponse(500, ErrorResponse{})

    // API routes
    api := engine.Group("/api/v1")
    api.AddTags("Todos")

    // List todos
    api.Get("/todos", listTodos).
        SetSummary("List all todos").
        SetDescription("Retrieve a paginated list of todos").
        AddJSONResponse(200, TodoListResponse{}).OK()

    // Create todo
    api.Post("/todos", createTodo).
        SetSummary("Create a new todo").
        SetInput(CreateTodoRequest{}).
        AddJSONResponse(201, TodoResponse{}).OK()

    // Get todo
    api.Get("/todos/:id", getTodo).
        SetSummary("Get todo by ID").
        AddJSONResponse(200, TodoResponse{}).
        AddJSONResponse(404, ErrorResponse{}).OK()

    // Update todo
    api.Put("/todos/:id", updateTodo).
        SetSummary("Update todo").
        SetInput(UpdateTodoRequest{}).
        AddJSONResponse(200, TodoResponse{}).
        AddJSONResponse(404, ErrorResponse{}).OK()

    // Delete todo
    api.Delete("/todos/:id", deleteTodo).
        SetSummary("Delete todo").
        AddJSONResponse(204, nil).
        AddJSONResponse(404, ErrorResponse{}).OK()

    // Serve documentation
    engine.ServeDocUI("/docs", soda.UISwaggerUI)
    engine.ServeSpecJSON("/openapi.json")
    engine.ServeSpecYAML("/openapi.yaml")

    engine.App().Listen(":3000")
}

// Handler implementations would go here...
```

## 📊 Available UI Options

Soda provides several built-in options for serving interactive API documentation:

```go
// Swagger UI (most popular)
engine.ServeDocUI("/swagger", soda.UISwaggerUI)

// ReDoc (clean and modern)
engine.ServeDocUI("/redoc", soda.UIRedoc)

// RapiDoc (feature-rich)
engine.ServeDocUI("/rapidoc", soda.UIRapiDoc)

// Stoplight Elements (elegant design)
engine.ServeDocUI("/elements", soda.UIStoplightElement)
```

## 🔧 Configuration

### OpenAPI Information

```go
engine.OpenAPI().Info = &openapi3.Info{
    Title:          "Your API Title",
    Version:        "1.0.0",
    Description:    "Detailed API description",
    TermsOfService: "https://example.com/terms",
    Contact: &openapi3.Contact{
        Name:  "API Support",
        Email: "support@example.com",
        URL:   "https://example.com/support",
    },
    License: &openapi3.License{
        Name: "MIT",
        URL:  "https://opensource.org/licenses/MIT",
    },
}

// Add servers
engine.OpenAPI().Servers = openapi3.Servers{
    &openapi3.Server{
        URL:         "https://api.example.com/v1",
        Description: "Production server",
    },
    &openapi3.Server{
        URL:         "https://staging-api.example.com/v1",
        Description: "Staging server",
    },
}
```

### Custom UIRender

You can implement your own documentation UI:

```go
type CustomUIRender struct{}

func (c CustomUIRender) Render(doc *openapi3.T) string {
    // Your custom HTML/JS implementation
    return `<!DOCTYPE html><html>...</html>`
}

// Use it
engine.ServeDocUI("/custom", CustomUIRender{})
```

## 🧪 Testing

Soda includes comprehensive test coverage. Run the tests with:

```bash
go test ./...
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Fiber](https://github.com/gofiber/fiber) - Fast HTTP framework for Go
- [kin-openapi](https://github.com/getkin/kin-openapi) - OpenAPI 3 specification implementation
- [Swagger UI](https://swagger.io/tools/swagger-ui/) - Interactive API documentation
- [ReDoc](https://github.com/Redocly/redoc) - OpenAPI/Swagger-generated API Reference Documentation
- [RapiDoc](https://github.com/rapi-doc/RapiDoc) - Web Component for OpenAPI spec
- [Stoplight Elements](https://github.com/stoplightio/elements) - Beautiful API documentation

## 📈 Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes.

## 📞 Support

If you have any questions or need help, please open an issue on the [GitHub repository](https://github.com/neo-f/soda/issues).

