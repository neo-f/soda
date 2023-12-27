package soda

import (
	"mime"
	"net/http"
	"reflect"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// GetInput gets the input value from the http request.
func GetInput[T any](c *http.Request) *T {
	return c.Context().Value(KeyInput).(*T)
}

// GenerateSchema generates an OpenAPI schema for a given model using the given name tag.
// It takes in the model to generate a schema for and a name tag to use for naming properties.
// It returns a *spec.Schema that represents the generated schema.
func GenerateSchema(model any, tag string) *base.Schema {
	// Create a new generator.
	generator := NewGenerator()

	// Generate a schema for the model.
	ref := generator.generateSchema(nil, reflect.TypeOf(model), tag)

	// Return the generated schema.
	return derefSchema(generator.doc, ref)
}

// mediaTypeAliases is a map of media type aliases to media types.
var mediaTypeAliases = map[string]string{
	"json": "application/json; charset=utf-8",
	"text": "text/plain; charset=utf-8",
	"form": "multipart/form-data; charset=utf-8",
	"pdf":  "application/pdf",
}

func resolveMediaType(mediaType string) string {
	result := mediaType
	if mt, ok := mediaTypeAliases[mediaType]; ok {
		result = mt
	}
	result, _, _ = mime.ParseMediaType(result)
	return result
}

func RegisterMediaTypeAlias(alias, mediaType string) error {
	if _, _, err := mime.ParseMediaType(mediaType); err != nil {
		return err
	}
	mediaTypeAliases[alias] = mediaType
	return nil
}
