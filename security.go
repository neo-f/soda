package soda

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// NewJWTSecurityScheme creates a new JWT security scheme.
func NewJWTSecurityScheme(description ...string) *openapi3.SecurityScheme {
	sec := openapi3.NewJWTSecurityScheme()
	if len(description) != 0 {
		sec = sec.WithDescription(description[0])
	}
	return sec
}

// NewAPIKeySecurityScheme creates a new API key security scheme.
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

// NewHTTPBasicSecurityScheme creates a new HTTP Basic authentication security scheme.
func NewHTTPBasicSecurityScheme(description ...string) *openapi3.SecurityScheme {
	sec := openapi3.NewSecurityScheme().
		WithType("http").
		WithScheme("basic")
	if len(description) != 0 {
		sec = sec.WithDescription(description[0])
	}
	return sec
}

// NewHTTPBearerSecurityScheme creates a new HTTP Bearer token security scheme.
// This is a generic bearer token scheme (not JWT-specific).
// For JWT, use NewJWTSecurityScheme instead.
func NewHTTPBearerSecurityScheme(description ...string) *openapi3.SecurityScheme {
	sec := openapi3.NewSecurityScheme().
		WithType("http").
		WithScheme("bearer")
	if len(description) != 0 {
		sec = sec.WithDescription(description[0])
	}
	return sec
}

// NewOAuth2SecurityScheme creates a new OAuth2 security scheme with the specified flows.
// Example:
//
//	flows := &openapi3.OAuthFlows{
//		Implicit: &openapi3.OAuthFlow{
//			AuthorizationURL: "https://example.com/oauth/authorize",
//			Scopes:           map[string]string{"read": "Read access"},
//		},
//	}
//	scheme := NewOAuth2SecurityScheme(flows, "OAuth2 authentication")
func NewOAuth2SecurityScheme(flows *openapi3.OAuthFlows, description ...string) *openapi3.SecurityScheme {
	sec := openapi3.NewSecurityScheme().
		WithType("oauth2")
	sec.Flows = flows
	if len(description) != 0 {
		sec = sec.WithDescription(description[0])
	}
	return sec
}

// NewOpenIDConnectSecurityScheme creates a new OpenID Connect security scheme.
func NewOpenIDConnectSecurityScheme(openIdConnectUrl string, description ...string) *openapi3.SecurityScheme {
	sec := openapi3.NewSecurityScheme().
		WithType("openIdConnect")
	sec.OpenIdConnectUrl = openIdConnectUrl
	if len(description) != 0 {
		sec = sec.WithDescription(description[0])
	}
	return sec
}
