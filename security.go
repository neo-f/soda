package soda

import (
	"github.com/getkin/kin-openapi/openapi3"
)

func NewJWTSecurityScheme(description ...string) *openapi3.SecurityScheme {
	sec := openapi3.NewJWTSecurityScheme()
	if len(description) != 0 {
		sec = sec.WithDescription(description[0])
	}
	return sec
}

func NewAPIKeySecurityScheme(in string, name string, description ...string) *openapi3.SecurityScheme {
	sec := openapi3.NewSecurityScheme().
		WithType("apiKey").
		WithIn(in).
		WithName(name)
	if len(description) != 0 {
		sec = sec.WithDescription(description[0])
	}
	return sec
}
