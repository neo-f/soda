package soda

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func NewJWTSecurityScheme(description ...string) *v3.SecurityScheme {
	var desc string
	if len(description) != 0 {
		desc = description[0]
	}

	return &v3.SecurityScheme{
		Description:  desc,
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "bearer",
	}
}

func NewAPIKeySecurityScheme(in string, name string, description ...string) *v3.SecurityScheme {
	var desc string
	if len(description) != 0 {
		desc = description[0]
	}

	return &v3.SecurityScheme{
		Description: desc,
		Type:        "apiKey",
		Name:        name,
		In:          in,
	}
}
